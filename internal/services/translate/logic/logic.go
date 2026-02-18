package logic

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

func New(translator Translator, wordAligner WordAligner) *Logic {
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
) (*domain.ChapterAlignNode, error) {
	const operation = "translate.logic.TranslateChapter"

	startTime := time.Now()

	translatedChapter := make([]domain.ParagraphAlignNode, len(chapter.Paragraphs))

	g, ctxErrGroup := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrency)

	for idx, paragraph := range chapter.Paragraphs {
		idx, paragraph := idx, paragraph
		g.Go(func() error {
			select {
			case <-ctxErrGroup.Done():
				return ctxErrGroup.Err()
			default:
			}

			startTranslateParagraphTime := time.Now()
			slog.Info(
				"start translate paragraph",
				slog.Int("paragraph length", len([]rune(paragraph))),
				slog.Int("paragraph index", idx),
				slog.String("operation", operation),
			)

			alignedParagraph, err := l.TranslateAndAlignParagraph(ctxErrGroup, paragraph, from, to, throttler)
			if err != nil {
				return err
			}
			translatedChapter[idx] = alignedParagraph

			slog.Info(
				"translated paragraph",
				slog.Int("paragraph index", idx),
				slog.String("duration", time.Since(startTranslateParagraphTime).String()),
				slog.String("operation", operation),
			)

			return nil
		})
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
			return nil, fmt.Errorf("%s: failed to translate title: %w", operation, err)
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
		"translated chapter",
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
	const operation = "translate.logic.TranslateAndAlignParagraph"

	// перевод параграфа (использую libretranslate, которую развернул на серваке у себя)
	translateOutput, err := l.translator.Translate(&translate.Input{
		Q:      paragraph,
		Source: from,
		Target: to,
	})
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return domain.ParagraphAlignNode{}, errors.New("failed to translate paragraph")
	}

	// case: error on alignment if paragraph has no letters.
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
		return domain.ParagraphAlignNode{}, errors.New("failed to align words")
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
