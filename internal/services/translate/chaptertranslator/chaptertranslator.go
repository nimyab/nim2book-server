package chaptertranslator

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/services/libretranslate/translate"
	"github.com/nimyab/nim2book-back/pkg/contains_letters"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
	"github.com/nimyab/nim2book-back/proto/word_aligner"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type FileSave = func([]byte) (string, error)

type Throttle = func()

type Translator interface {
	Translate(input *translate.Input) (*translate.Output, error)
}

type Aligner interface {
	Align(context.Context, *word_aligner.AlignRequest, ...grpc.CallOption) (*word_aligner.AlignResponse, error)
}

type ChapterTranslator struct {
	fileSave   FileSave
	throttle   Throttle
	translator Translator
	aligner    Aligner
}

func New(
	fileSave FileSave,
	throttle Throttle,
	translator Translator,
	aligner Aligner,
) *ChapterTranslator {
	return &ChapterTranslator{
		fileSave:   fileSave,
		throttle:   throttle,
		translator: translator,
		aligner:    aligner,
	}
}

var (
	ErrTranslateParagraph          = errors.New("paragraph translate error")
	ErrAlignParagraph              = errors.New("paragraph align error")
	ErrNotFoundSuchContentUnitType = errors.New("not found such content unit type")
	ErrFailedToTranslateChapter    = errors.New("failed to translate chapter")
)

type Input struct {
	Chapter        epub_parser.FormattedChapter
	ChapterOrder   int
	From, To       domain.SupportedLang
	MaxConcurrency int
}

func (t *ChapterTranslator) TranslateChapter(ctx context.Context, input Input) (*domain.ChapterAlignNode, error) {
	const operation = "translate.ChapterTranslator.TranslateChapter"
	logger := slog.With(
		slog.String("operation", operation),
		slog.String("chapterTitle", input.Chapter.CapterTitle),
		slog.Int("chapterOrder", input.ChapterOrder),
	)

	startTime := time.Now()

	contentNodes := make([]domain.ContentNode, len(input.Chapter.Content))

	g, ctxErrGroup := errgroup.WithContext(ctx)
	g.SetLimit(input.MaxConcurrency)

	for idx, contentUnit := range input.Chapter.Content {
		g.Go(func() error {
			select {
			case <-ctxErrGroup.Done():
				return ctxErrGroup.Err()
			default:
			}

			if contentUnit.Type == epub_parser.ContentTypeText && contentUnit.TextNode != nil {
				startTranslateParagraph := time.Now()
				logger.Info(
					"start translating paragraph",
					slog.Int("contentUnitIndex", idx),
					slog.Int("paragraph length", len(contentUnit.TextNode.Text)),
				)

				paragraph, err := t.translateAndAlignParagraph(ctxErrGroup, contentUnit.TextNode.Text, input.From, input.To)
				if err != nil {
					logger.Error(
						"failed to translate and align paragraph",
						slog.String("err", err.Error()),
						slog.String("paragraph", contentUnit.TextNode.Text),
					)
					return err
				}
				contentNodes[idx] = domain.ContentNode{
					Type:               domain.ParagraphAlignNodeTypeParagraph,
					ParagraphAlignNode: &paragraph,
				}

				logger.Info(
					"finished translating paragraph",
					slog.Int("contentUnitIndex", idx),
					slog.Int("paragraph length", len(contentUnit.TextNode.Text)),
					slog.Duration("duration", time.Since(startTranslateParagraph)),
				)

				return nil
			}

			if contentUnit.Type == epub_parser.ContentTypeImage && contentUnit.ImageNode != nil {
				data, err := contentUnit.ImageNode.File.GetRawContent()
				if err != nil {
					logger.Error(
						"failed to get raw content of image",
						slog.String("err", err.Error()),
						slog.Int("contentUnitIndex", idx),
					)
					// Если не удалось получить контент изображения, мы логируем ошибку и продолжаем без этого изображения.
					return nil
				}
				imagePath, err := t.fileSave(data)
				if err != nil {
					logger.Error(
						"failed to save image",
						slog.String("err", err.Error()),
						slog.Int("contentUnitIndex", idx),
					)
					// Если не удалось сохранить изображение, мы логируем ошибку и продолжаем без этого изображения.
					return nil
				}
				contentNodes[idx] = domain.ContentNode{
					Type: domain.ParagraphAlignNodeTypeImage,
					ImageNode: &domain.ImageNode{
						Path: imagePath,
						Alt:  contentUnit.ImageNode.Alt,
					},
				}
				return nil
			}

			return ErrNotFoundSuchContentUnitType
		})
	}

	if err := g.Wait(); err != nil {
		slog.Error("failed to translate chapter", slog.String("err", err.Error()))
		return nil, ErrFailedToTranslateChapter
	}

	translatedTitle := ""
	if len(input.Chapter.CapterTitle) > 0 {
		translatedTitleOutput, err := t.translator.Translate(&translate.Input{
			Q:      input.Chapter.CapterTitle,
			Source: input.From,
			Target: input.To,
		})
		if err != nil {
			logger.Error(
				"failed to translate chapter title",
				slog.String("err", err.Error()),
				slog.String("chapterTitle", input.Chapter.CapterTitle),
			)
		} else {
			translatedTitle = translatedTitleOutput.TranslatedText
		}
	}

	logger.Info(
		"finished translating chapter",
		slog.String("chapterTitle", input.Chapter.CapterTitle),
		slog.Duration("duration", time.Since(startTime)),
	)

	return &domain.ChapterAlignNode{
		Content:         contentNodes,
		TranslatedTitle: translatedTitle,
		Title:           input.Chapter.CapterTitle,
		Order:           input.ChapterOrder,
		Id:              input.Chapter.PamphletChapterData.ID,
	}, nil
}

func (t *ChapterTranslator) translateAndAlignParagraph(
	ctx context.Context,
	paragraph string,
	from, to domain.SupportedLang,
) (domain.ParagraphAlignNode, error) {
	const operation = "translate.ChapterTranslator.TranslateAndAlignParagraph"
	logger := slog.With(slog.String("operation", operation))

	translateParagraphOutput, err := t.translator.Translate(&translate.Input{
		Q:      paragraph,
		Source: from,
		Target: to,
	})
	if err != nil {
		logger.Error(
			"failed to translate paragraph",
			slog.String("err", err.Error()),
			slog.String("paragraph", paragraph),
			slog.String("from", string(from)),
			slog.String("to", string(to)),
		)
		return domain.ParagraphAlignNode{}, ErrTranslateParagraph
	}
	translateParagraph := translateParagraphOutput.TranslatedText
	// Функция чтобы не дудосить себя
	t.throttle()

	// случай: ошибка при выравнивании, если параграф не содержит букв.
	if !contains_letters.ContainsLetters(paragraph) || !contains_letters.ContainsLetters(translateParagraph) {
		return domain.ParagraphAlignNode{
			OriginalParagraph:   paragraph,
			TranslatedParagraph: translateParagraph,
			AlignmentWords: []domain.WordAlignNode{{
				IndexesOriginalWord:   [2]int{0, len([]rune(paragraph)) - 1},
				IndexesTranslatedWord: [2]int{0, len([]rune(translateParagraph)) - 1},
			}},
		}, nil
	}

	alignParagraphs, err := t.aligner.Align(ctx, &word_aligner.AlignRequest{
		SourceText: paragraph,
		TargetText: translateParagraph,
	})
	if err != nil {
		logger.Error(
			"failed to align paragraphs",
			slog.String("err", err.Error()),
			slog.String("originalParagraph", paragraph),
			slog.String("translatedParagraph", translateParagraph),
		)
		return domain.ParagraphAlignNode{}, ErrAlignParagraph
	}
	// Функция чтобы не дудосить себя
	t.throttle()

	alignWords := make([]domain.WordAlignNode, 0, len(alignParagraphs.Alignments))
	for _, alignment := range alignParagraphs.Alignments {
		alignWords = append(alignWords, domain.WordAlignNode{
			IndexesOriginalWord:   [2]int{int(alignment.SrcStart), int(alignment.SrcEnd)},
			IndexesTranslatedWord: [2]int{int(alignment.TargetStart), int(alignment.TargetEnd)},
		})
	}

	return domain.ParagraphAlignNode{
		OriginalParagraph:   paragraph,
		TranslatedParagraph: translateParagraph,
		AlignmentWords:      alignWords,
	}, nil
}
