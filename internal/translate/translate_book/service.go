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
	"sync"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/adapter/postgres"
	"github.com/nimyab/nim2book-back/internal/adapter/rabbitmq"
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

type Rabbitmq interface {
	Publish(data *rabbitmq.NotificationData) error
}

type Postgres interface {
	GetBookByAuthorAndTitle(ctx context.Context, author, title string) (*domain.Book, error)
	CreateBook(ctx context.Context, book *domain.Book) (*domain.Book, error)
	GetFcmTokensByUserId(ctx context.Context, userId domain.Id) ([]domain.FcmToken, error)
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

	mu                          sync.Mutex
	currentCountBookTranslating int

	s3          S3
	pg          Postgres
	wordAligner WordAligner
	translator  Translator
	rabbitmq    Rabbitmq

	messagingFirebaseClient *messaging.Client
}

type translatedChapterResult struct {
	Chapter *domain.ChapterAlignNode
	Error   error
}

var (
	ErrFailedTranslateChapter   = errors.New("failed translate chapter")
	ErrFailedSaveToStorage      = errors.New("failed save to storage")
	ErrFailedSaveBookToDatabase = errors.New("failed save book to database")
)

var service *Service

func New(
	s3 S3,
	pg Postgres,
	wordAligner WordAligner,
	translator Translator,
	rabbitmq Rabbitmq,
	maxRequestCount int,
	waitMilliseconds time.Duration,
	messagingFirebaseClient *messaging.Client,
) *Service {
	service = &Service{
		s3:                          s3,
		pg:                          pg,
		wordAligner:                 wordAligner,
		translator:                  translator,
		rabbitmq:                    rabbitmq,
		maxRequestCount:             maxRequestCount,
		waitMilliseconds:            waitMilliseconds,
		currentCountBookTranslating: 0,
		messagingFirebaseClient:     messagingFirebaseClient,
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

	parsedBook, chapters, coverData, err := epub_parser.Parse(data)
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

	go func() {
		err := s.startTranslate(userId, chapters, coverData, parsedBook, input)
		if err != nil {
			switch {
			case errors.Is(err, ErrFailedSaveToStorage):
				if err := s.rabbitmq.Publish(&rabbitmq.NotificationData{
					Id:     uuid.New().String(),
					UserId: userId,
					Data: map[string]interface{}{
						"book":         parsedBook,
						"errorMessage": "Произошла ошибка во время сохранения главы, попробуйте перевести книгу позже.",
					},
				}); err != nil {
					logger.Error("fail send message to rabbit", err, operation)
				}
			case errors.Is(err, ErrFailedTranslateChapter):
				if err := s.rabbitmq.Publish(&rabbitmq.NotificationData{
					Id:     uuid.New().String(),
					UserId: userId,
					Data: map[string]interface{}{
						"book":         parsedBook,
						"errorMessage": "Произошла ошибка во время перевода главы, попробуйте перевести книгу позже.",
					},
				}); err != nil {
					logger.Error("fail send message to rabbit", err, operation)
				}
			case errors.Is(err, ErrFailedSaveBookToDatabase):
				if err := s.rabbitmq.Publish(&rabbitmq.NotificationData{
					Id:     uuid.New().String(),
					UserId: userId,
					Data: map[string]interface{}{
						"book":         parsedBook,
						"errorMessage": "Произошла ошибка во время сохранения книги, попробуйте перевести книгу позже, извините за неудобства.",
					},
				}); err != nil {
					logger.Error("fail send message to rabbit", err, operation)
				}
			default:
				logger.Error("unexpected error", err, operation)
			}
		}
	}()

	return &Output{Message: "start translate"}, nil
}

func (s *Service) startTranslate(
	userId domain.Id,
	chapters []epub_parser.FormattedChapter,
	coverData []byte,
	book *pamphlet.Book,
	input *Input,
) error {
	const operation = "translate_book.startTranslate"

	// увеличиваем количество переводимых книг, нужно для задержек
	s.mu.Lock()
	s.currentCountBookTranslating++
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.currentCountBookTranslating--
		s.mu.Unlock()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// use buffer chan to saveChapterToS3 non block translateChapters goroutine
	resultChan := make(chan translatedChapterResult, len(chapters))

	go s.translateChapters(ctx, resultChan, chapters, book, input.From, input.To)

	paths := make([]string, 0, len(chapters))
	for result := range resultChan {
		if result.Error != nil {
			logger.Error(
				fmt.Sprintf("failed to translate chapter, title: %s, author: %s", book.Title, book.Author),
				result.Error,
				operation,
			)
			return ErrFailedTranslateChapter
		}
		if result.Chapter == nil {
			logger.Error(
				fmt.Sprintf("translated chapter is nil, title: %s, author: %s", book.Title, book.Author),
				result.Error,
				operation,
			)
			return ErrFailedTranslateChapter
		}
		path, err := s.saveChapterToS3(result.Chapter, book.Title)
		if err != nil {
			logger.Error(
				fmt.Sprintf("failed to save to s3 chapter %d order, title: %s, author: %s", result.Chapter.Order, book.Title, book.Author),
				err,
				operation,
			)
			return ErrFailedSaveToStorage
		}
		paths = append(paths, path)

		// отправляем уведомления о переведенной главе
		if err := s.rabbitmq.Publish(&rabbitmq.NotificationData{
			Id:     uuid.New().String(),
			UserId: userId,
			Data: map[string]interface{}{
				"chapterPath":       path,
				"author":            book.Author,
				"title":             book.Title,
				"chapterOrder":      result.Chapter.Order,
				"totalChapterCount": len(book.Chapters),
			},
		}); err != nil {
			logger.Error("fail send message to rabbit", err, operation)
		}
	}

	var cover *string = nil
	if coverData != nil {
		coverPath, err := s.saveCoverToS3(coverData, book.Title)
		if err != nil {
			slog.Error(err.Error(), slog.String("operation", operation))
		} else {
			cover = &coverPath
		}
	}

	newBook := &domain.Book{
		Author:       book.Author,
		Title:        book.Title,
		ChapterPaths: paths,
		Cover:        cover,
	}
	newBook, err := s.pg.CreateBook(context.Background(), newBook)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to save book to database, title: %s, author: %s", book.Title, book.Author),
			err,
			operation,
		)
		return ErrFailedSaveBookToDatabase
	}

	if err := s.rabbitmq.Publish(&rabbitmq.NotificationData{
		Id:     uuid.New().String(),
		UserId: userId,
		Data: map[string]interface{}{
			"book": newBook,
		},
	}); err != nil {
		logger.Error("fail send message to rabbit", err, operation)
	}

	return nil
}

func (s *Service) translateChapters(
	ctx context.Context,
	resultChan chan<- translatedChapterResult,
	chapters []epub_parser.FormattedChapter,
	book *pamphlet.Book,
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
				Error:   ErrFailedTranslateChapter,
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
					Error:   ErrFailedTranslateChapter,
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
			slog.String("title", book.Title),
			slog.String("author", book.Author),
		)
	}

}

func (s *Service) translateAndAlignParagraph(paragraph string, from domain.SupportedLang, to domain.SupportedLang) (domain.ParagraphAlignNode, error) {
	const operation = "translate_book.Service.translateAndAlignParagraph"

	// перевод параграфа
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
	time.Sleep(time.Duration(s.currentCountBookTranslating) * s.waitMilliseconds)

	// выравнивание слов
	alignOutput, err := s.wordAligner.Align(&align.Input{
		SourceText: paragraph,
		TargetText: translateOutput.TranslatedText,
	})
	time.Sleep(time.Duration(s.currentCountBookTranslating) * s.waitMilliseconds)

	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return domain.ParagraphAlignNode{}, errors.New("failed to align words")
	}

	// в массив alignWords добавляем только уникальные слова, уникальность проверяем по исходному тексту
	// пример: у нас есть тект "the boy" и переведен как "мальчик", слова "the" и "boy" будут ссылаться на слово "мальчик", нас это устраивает, но если будет наоборот
	// то есть если у нас будет одно слово на англ и два на русском, то нас это не будет устраивать, так как появляются проблемы на фронте в отображении слов, слова просто повторяются
	// p.s. можно подумать почему в русском не будут повторяться, все просто, в русском мы показываем весь пораграф, а не разбиваем его на слова, а выбранное слово выделяем по индексам
	alignWords := make([]domain.WordAlignNode, 0, len(alignOutput.Alignments))
	for _, alignment := range alignOutput.Alignments {
		alignWord := domain.WordAlignNode{
			IndexesOriginalWord:   alignment.SourceIndexes,
			IndexesTranslatedWord: alignment.TargetIndexes,
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

func (s *Service) saveChapterToS3(chapter *domain.ChapterAlignNode, bookTitle string) (string, error) {
	const operation = "translate_book.Service.saveChapterToS3"

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

func (s *Service) saveCoverToS3(coverData []byte, bookTitle string) (string, error) {
	const operation = "translate_book.Service.saveCoverToS3"

	path := fmt.Sprintf("cover/%s/%s", strings.ReplaceAll(bookTitle, " ", "_"), uuid.New().String())

	if err := s.s3.Upload(path, coverData); err != nil {
		return "", fmt.Errorf("%s: failed upload to s3: %w", operation, err)
	}

	return path, nil
}
