package translate_logic

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/services/libretranslate/translate"
	"github.com/nimyab/nim2book-back/pkg/contains_letters"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
	pb "github.com/nimyab/nim2book-back/proto/word_aligner"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type WordAligner interface {
	Align(ctx context.Context, in *pb.AlignRequest, opts ...grpc.CallOption) (*pb.AlignResponse, error)
}

type Translator interface {
	Translate(input *translate.Input) (*translate.Output, error)
}

type Throttler interface {
	Throttle()
}

type ThrottlerFunc func()

func (f ThrottlerFunc) Throttle() {
	f()
}

type Logic struct {
	translator  Translator
	wordAligner WordAligner
}

func NewLogic(translator Translator, wordAligner WordAligner) *Logic {
	return &Logic{
		translator:  translator,
		wordAligner: wordAligner,
	}
}

func (l *Logic) TranslateChapter(
	ctx context.Context,
	chapter epub_parser.FormattedChapter,
	from, to domain.SupportedLang,
	throttler Throttler,
	maxConcurrency int,
	imageSaver func(data []byte) (string, error),
) (*domain.ChapterAlignNode, error) {
	const operation = "translate.flow.TranslateChapter"

	startTime := time.Now()

	translatedChapter := make([]domain.ContentNode, len(chapter.Content))

	g, ctxErrGroup := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrency)

	for idx, item := range chapter.Content {

		if item.Type == epub_parser.ContentTypeImage {
			// Извлечение данных изображения из zip файла
			data, err := item.ImageNode.File.GetRawContent()
			if err != nil {
				slog.Error("не удалось открыть файл изображения", slog.String("error", err.Error()))
				continue
			}

			// Сохранение изображения, если предоставлен saver
			var imageURL string
			if imageSaver != nil {
				url, err := imageSaver(data)
				if err != nil {
					slog.Error("не удалось сохранить изображение", slog.String("error", err.Error()))
				} else {
					imageURL = url
				}
			}

			translatedChapter[idx] = domain.ContentNode{
				Type: domain.ParagraphAlignNodeTypeImage,
				ImageNode: &domain.ImageNode{
					Path: imageURL,
					Alt:  item.ImageNode.Alt,
				},
			}
			continue
		}

		if item.Type == epub_parser.ContentTypeText {
			g.Go(func() error {
				select {
				case <-ctxErrGroup.Done():
					return ctxErrGroup.Err()
				default:
				}

				startTranslateParagraphTime := time.Now()
				slog.Info(
					"начало перевода параграфа",
					slog.Int("paragraph length", len([]rune(item.TextNode.Text))),
					slog.Int("paragraph index", idx),
					slog.String("operation", operation),
				)

				alignedParagraph, err := l.TranslateAndAlignParagraph(ctxErrGroup, item.TextNode.Text, from, to, throttler)
				if err != nil {
					return err
				}
				translatedChapter[idx] = domain.ContentNode{
					Type:               domain.ParagraphAlignNodeTypeParagraph,
					ParagraphAlignNode: &alignedParagraph,
				}

				slog.Info(
					"конец перевода параграфа",
					slog.Duration("duration", time.Since(startTranslateParagraphTime)),
					slog.Int("paragraph index", idx),
					slog.String("operation", operation),
				)

				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	var translatedTitle string
	if len(chapter.Title) > 0 {
		translatedTitleOutput, err := l.translator.Translate(&translate.Input{
			Q:      chapter.Title,
			Source: from,
			Target: to,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: не удалось перевести заголовок: %w", operation, err)
		}
		translatedTitle = translatedTitleOutput.TranslatedText
	} else {
		translatedTitle = ""
	}

	chapterNode := &domain.ChapterAlignNode{
		Id:              chapter.ID,
		Content:         translatedChapter,
		Title:           chapter.Title,
		TranslatedTitle: translatedTitle,
	}

	duration := time.Since(startTime)
	slog.Info(
		"глава переведена",
		slog.String("duration", duration.String()),
		slog.String("operation", operation),
	)

	return chapterNode, nil
}

func (l *Logic) TranslateAndAlignParagraph(
	ctx context.Context,
	paragraph string,
	from domain.SupportedLang,
	to domain.SupportedLang,
	throttler Throttler,
) (domain.ParagraphAlignNode, error) {
	const operation = "translate.flow.TranslateAndAlignParagraph"

	// перевод параграфа (использую libretranslate, которую развернул на серваке у себя)
	translateOutput, err := l.translator.Translate(&translate.Input{
		Q:      paragraph,
		Source: from,
		Target: to,
	})
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return domain.ParagraphAlignNode{}, errors.New("не удалось перевести параграф")
	}

	// случай: ошибка при выравнивании, если параграф не содержит букв.
	if !contains_letters.ContainsLetters(paragraph) || !contains_letters.ContainsLetters(translateOutput.TranslatedText) {
		alignedParagraph := domain.ParagraphAlignNode{
			OriginalParagraph:   paragraph,
			TranslatedParagraph: translateOutput.TranslatedText,
			AlignmentWords: []domain.WordAlignNode{{
				IndexesOriginalWord:   [2]int{0, len([]rune(paragraph)) - 1},
				IndexesTranslatedWord: [2]int{0, len([]rune(translateOutput.TranslatedText)) - 1},
			}},
		}
		return alignedParagraph, nil
	}

	if throttler != nil {
		throttler.Throttle()
	}

	alignOutput, err := l.wordAligner.Align(ctx, &pb.AlignRequest{
		SourceText: paragraph,
		TargetText: translateOutput.TranslatedText,
	})

	if throttler != nil {
		throttler.Throttle()
	}

	if err != nil {
		slog.Error(
			err.Error(),
			slog.String("operation", operation),
			slog.String("source text", paragraph),
			slog.String("target text", translateOutput.TranslatedText),
		)
		return domain.ParagraphAlignNode{}, errors.New("не удалось выровнять слова")
	}

	// в массив alignWords добавляем только уникальные слова, уникальность проверяем по исходному тексту
	// пример: у нас есть тект "the boy" и переведен как "мальчик", слова "the" и "boy" будут ссылаться на слово "мальчик", нас это устраивает, но если будет наоборот
	// то есть если у нас будет одно слово на англ и два на русском, то нас это не будет устраивать, так как появляются проблемы на фронте в отображении слов, слова просто повторяются
	// p.s. можно подумать почему в русском не будут повторяться, все просто, в русском мы показываем весь пораграф, а не разбиваем его на слова, а выбранное слово выделяем по индексам
	// p.p.s в новом выравнивателе таких проблем нет, но на всякий случай оставил эту проверку
	alignWords := make([]domain.WordAlignNode, 0, len(alignOutput.Alignments))
	for _, alignment := range alignOutput.Alignments {
		alignWord := domain.WordAlignNode{
			IndexesOriginalWord:   [2]int{int(alignment.SrcStart), int(alignment.SrcEnd)},
			IndexesTranslatedWord: [2]int{int(alignment.TargetStart), int(alignment.TargetEnd)},
		}
		exist := false
		for _, addedAlignWord := range alignWords {
			if alignWord.IndexesOriginalWord[0] == addedAlignWord.IndexesOriginalWord[0] &&
				alignWord.IndexesOriginalWord[1] == addedAlignWord.IndexesOriginalWord[1] {
				exist = true
				break
			}
		}
		if !exist {
			alignWords = append(alignWords, alignWord)
		}
	}

	alignedParagraph := domain.ParagraphAlignNode{
		OriginalParagraph:   paragraph,
		TranslatedParagraph: translateOutput.TranslatedText,
		AlignmentWords:      alignWords,
	}

	return alignedParagraph, nil
}
