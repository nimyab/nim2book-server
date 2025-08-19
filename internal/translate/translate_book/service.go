package translate_book

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"strings"
	"time"

	"github.com/nimyab/nim2book-back/internal/adapter/postgres"
	"github.com/nimyab/nim2book-back/internal/controller/websocket"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/libretranslate/translate"
	"github.com/nimyab/nim2book-back/internal/word_aligner/align"
	"github.com/nimyab/nim2book-back/pkg/contains_letters"
	"github.com/nimyab/nim2book-back/pkg/logger"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
	"github.com/timsims/pamphlet"
	"golang.org/x/sync/errgroup"
)

type S3 interface {
	Upload(path string, data []byte) error
}

type Postgres interface {
	GetBookByAuthorAndTitle(ctx context.Context, author, title string) (*domain.Book, error)
	CreateBook(ctx context.Context, book *domain.Book) (*domain.Book, error)
}

type WordAligner interface {
	Align(input *align.Input) (*align.Output, error)
}

type Translator interface {
	Translate(input *translate.Input) (*translate.Output, error)
}

type Service struct {
	maxRequestCount  int
	waitMilliseconds time.Duration

	s3          S3
	pg          Postgres
	wordAligner WordAligner
	translator  Translator
}

type translatedChapterResult struct {
	Chapter *domain.ChapterAlignNode
	Error   error
}

var service *Service

func New(
	s3 S3,
	pg Postgres,
	wordAligner WordAligner,
	translator Translator,
	maxRequestCount int,
	waitMilliseconds time.Duration,
) *Service {
	service = &Service{
		s3:               s3,
		pg:               pg,
		wordAligner:      wordAligner,
		translator:       translator,
		maxRequestCount:  maxRequestCount,
		waitMilliseconds: waitMilliseconds,
	}
	return service
}

func (s *Service) TranslateBook(input *Input, book *multipart.FileHeader, userId domain.Id) (*Output, error) {
	const operation = "translate_book.Service.TranslateBook"

	file, err := book.Open()
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, errors.New("failed to open book file")
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, errors.New("failed to read book file")
	}

	startParse := time.Now()

	parsedBook, chapters, err := epub_parser.Parse(data)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, errors.New("failed to parse book file")
	}

	slog.Info(
		"Book is parsed successfully",
		slog.String("duration", time.Since(startParse).String()),
		slog.Int("chapters count", len(chapters)),
		slog.String("author", parsedBook.Author),
		slog.String("title", parsedBook.Title),
	)

	existedBook, err := s.pg.GetBookByAuthorAndTitle(context.Background(), parsedBook.Author, parsedBook.Title)
	if existedBook != nil {
		return &Output{Book: existedBook}, nil
	}
	if err != nil && !errors.Is(err, postgres.ErrBookNotFound) {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, errors.New("failed to find book")
	}

	go s.startTranslate(userId, chapters, parsedBook, input)

	return &Output{Message: "start translate"}, nil
}

func (s *Service) startTranslate(userId domain.Id, chapters []epub_parser.FormattedChapter, book *pamphlet.Book, input *Input) {
	const operation = "translate_book.startTranslate"

	sendErrorMessage := func(body map[string]any) {
		websocket.SendMessage(userId, &websocket.Message{
			Event: websocket.ErrorEvent,
			Body:  body,
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// use buffer chan to saveToS3 non block translateChapters goroutine
	resultChan := make(chan translatedChapterResult, len(chapters))

	go s.translateChapters(ctx, resultChan, chapters, input.From, input.To)

	paths := make([]string, 0, len(chapters))
	for result := range resultChan {
		if result.Error != nil {
			logger.Error("failed to translate chapter", result.Error, operation)
			return
		}
		if result.Chapter == nil {
			logger.Error("translated chapter is nil", result.Error, operation)
			sendErrorMessage(map[string]any{
				"error": "internal server error: translated chapter is nil",
			})
			return
		}
		path, err := s.saveToS3(result.Chapter, book.Title)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to save to s3 chapter %d order", result.Chapter.Order), err, operation)
			sendErrorMessage(map[string]any{
				"error": fmt.Sprintf("failed to save to s3 chapter %d order", result.Chapter.Order),
			})
			return
		}
		paths = append(paths, path)
		websocket.SendMessage(userId, &websocket.Message{
			Event: websocket.ChapterTranslatedEvent,
			Body: map[string]any{
				"chapterPath":       path,
				"author":            book.Author,
				"title":             book.Title,
				"order":             result.Chapter.Order,
				"totalChapterCount": len(book.Chapters),
			},
		})
	}

	newBook := &domain.Book{
		Author:       book.Author,
		Title:        book.Title,
		ChapterPaths: paths,
	}
	newBook, err := s.pg.CreateBook(context.Background(), newBook)
	if err != nil {
		logger.Error("failed to save book to database", err, operation)
		sendErrorMessage(map[string]any{
			"error": "failed to save book to database",
		})
		return
	}
}

func (s *Service) translateChapters(
	ctx context.Context,
	resultChan chan<- translatedChapterResult,
	chapters []epub_parser.FormattedChapter,
	from domain.SupportedLang,
	to domain.SupportedLang,
) {
	const operation = "translate_book.Service.translateChapters"

	defer close(resultChan)

	for i, chapter := range chapters {
		select {
		case <-ctx.Done():
			slog.Debug("context cancelled", slog.String("operation", operation))
			return
		default:
		}

		startTime := time.Now()

		translatedChapter := make([]domain.ParagraphAlignNode, len(chapter.Paragraphs))
		// set limit to prevent ddos translator and word aligner services
		g := new(errgroup.Group)
		g.SetLimit(s.maxRequestCount)

		for idx, paragraph := range chapter.Paragraphs {
			idx, paragraph := idx, paragraph
			g.Go(func() error {
				alignedParagraph, err := s.translateAndAlignParagraph(paragraph, from, to)
				if err != nil {
					return err
				}
				translatedChapter[idx] = alignedParagraph
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			slog.Error(err.Error(), slog.String("operation", operation))
			resultChan <- translatedChapterResult{
				Chapter: nil,
				Error:   errors.New("failed to startTranslate paragraphs"),
			}
			return
		}

		var translatedTitle string
		if len(chapter.Title) > 0 {
			translatedTitleOutput, err := s.translator.Translate(&translate.Input{
				Q:      chapter.Title,
				Source: from,
				Target: to,
			})
			if err != nil {
				slog.Error(err.Error(), slog.String("operation", operation))
				resultChan <- translatedChapterResult{
					Chapter: nil,
					Error:   errors.New("failed to startTranslate chapter title"),
				}
				return
			}
			translatedTitle = translatedTitleOutput.TranslatedText
		} else {
			translatedTitle = ""
		}

		chapterNode := domain.ChapterAlignNode{
			Id:              chapter.ID,
			Content:         translatedChapter,
			Order:           i,
			Title:           chapter.Title,
			TranslatedTitle: translatedTitle,
		}

		resultChan <- translatedChapterResult{
			Chapter: &chapterNode,
			Error:   nil,
		}

		duration := time.Since(startTime)
		slog.Info(
			fmt.Sprintf("translated chapter %d", chapterNode.Order),
			slog.String("duration", duration.String()),
		)
	}

}

func (s *Service) translateAndAlignParagraph(paragraph string, from domain.SupportedLang, to domain.SupportedLang) (domain.ParagraphAlignNode, error) {
	const operation = "translate_book.Service.translateAndAlignParagraph"

	translateOutput, err := s.translator.Translate(&translate.Input{
		Q:      paragraph,
		Source: from,
		Target: to,
	})
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return domain.ParagraphAlignNode{}, errors.New("failed to startTranslate paragraph")
	}

	// case: 500 error on alignment if paragraph has no letters.
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
	time.Sleep(s.waitMilliseconds)

	alignOutput, err := s.wordAligner.Align(&align.Input{
		SourceText: paragraph,
		TargetText: translateOutput.TranslatedText,
	})
	time.Sleep(s.waitMilliseconds)

	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return domain.ParagraphAlignNode{}, errors.New("failed to align words")
	}

	alignWord := make([]domain.WordAlignNode, len(alignOutput.Alignments))
	for i, alignment := range alignOutput.Alignments {
		alignWord[i] = domain.WordAlignNode{
			IndexesOriginalWord:   alignment.SourceIndexes,
			IndexesTranslatedWord: alignment.TargetIndexes,
		}
	}

	alignedParagraph := domain.ParagraphAlignNode{
		OriginalParagraph:   paragraph,
		TranslatedParagraph: translateOutput.TranslatedText,
		AlignmentWords:      alignWord,
	}

	return alignedParagraph, nil
}

func (s *Service) saveToS3(chapter *domain.ChapterAlignNode, bookTitle string) (string, error) {
	const operation = "translate_book.Service.saveToS3"

	path := fmt.Sprintf("book/%s/%d.json", strings.ReplaceAll(bookTitle, " ", "_"), chapter.Order)

	data, err := json.Marshal(chapter)
	if err != nil {
		return "", fmt.Errorf("%s: failed to marshal chapter: %w", operation, err)
	}

	if err := s.s3.Upload(path, data); err != nil {
		return "", fmt.Errorf("%s: failed upload to s3: %w", operation, err)
	}

	return path, nil
}
