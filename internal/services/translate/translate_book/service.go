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

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
	"github.com/nimyab/nim2book-back/internal/services/libretranslate/translate"
	"github.com/nimyab/nim2book-back/internal/services/translate/dto"
	logic "github.com/nimyab/nim2book-back/internal/services/translate/translate_logic"
	"github.com/nimyab/nim2book-back/internal/services/word_aligner/align"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
	pb "github.com/nimyab/nim2book-back/proto/word_aligner"
)

type S3 interface {
	Upload(path string, data []byte) error
	Check(path string) error
}

type BookRepository interface {
	GetByAuthorAndTitle(ctx context.Context, authorName, title string) (*domain.Book, error)
	Create(ctx context.Context, book *domain.Book) (*domain.Book, error)
	CreateChapter(ctx context.Context, chapter *domain.BookChapter) (*domain.BookChapter, error)
}

type AuthorRepository interface {
	GetOrCreate(ctx context.Context, name string) (*domain.Author, error)
}

type WordAligner interface {
	Align(input *align.Input) (*align.Output, error)
}

type Translator interface {
	Translate(input *translate.Input) (*translate.Output, error)
}

type Service struct {
	config dto.Config

	mu                          sync.Mutex
	currentCountBookTranslating int

	s3          S3
	bookRepo    BookRepository
	authorRepo  AuthorRepository
	wordAligner pb.AlignmentServiceClient
	translator  Translator
	logic       *logic.Logic
}

var (
	ErrFailedTranslateChapter   = errors.New("failed translate chapter")
	ErrFailedSaveToStorage      = errors.New("failed save to storage")
	ErrFailedSaveBookToDatabase = errors.New("failed save book to database")
)

func New(
	s3 S3,
	bookRepo BookRepository,
	authorRepo AuthorRepository,
	wordAligner pb.AlignmentServiceClient,
	translator Translator,
	config dto.Config,
) *Service {
	return &Service{
		s3:                          s3,
		bookRepo:                    bookRepo,
		authorRepo:                  authorRepo,
		wordAligner:                 wordAligner,
		translator:                  translator,
		config:                      config,
		currentCountBookTranslating: 0,
		logic:                       logic.New(translator, wordAligner),
	}
}

func (s *Service) Throttle() {
	time.Sleep(time.Duration(s.currentCountBookTranslating) * s.config.WaitDuration)
}

func (s *Service) TranslateBook(ctx context.Context, input *Input, book *multipart.FileHeader, userId domain.ID) (*Output, error) {
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

	parsedData, err := epub_parser.Parse(data)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, errors.New("failed to parse book file")
	}

	slog.Info(
		"Book is parsed successfully",
		slog.String("duration", time.Since(startParse).String()),
		slog.Int("chapters count", len(parsedData.FormattedChapter)),
		slog.String("author", parsedData.Book.Author),
		slog.String("title", parsedData.Book.Title),
	)

	existedBook, err := s.bookRepo.GetByAuthorAndTitle(ctx, parsedData.Book.Author, parsedData.Book.Title)
	if existedBook != nil {
		return &Output{Book: existedBook}, nil
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, errors.New("failed to find book")
	}

	translatedData := &dto.TranslationContext{
		Book:      parsedData.Book,
		Chapters:  parsedData.FormattedChapter,
		CoverData: parsedData.Cover,
		UserID:    userId,
		From:      input.From,
		To:        input.To,
	}

	go func() {
		err := s.startTranslate(translatedData)
		if err != nil {
			slog.Error("failed translate book", slog.String("operation", operation), slog.Any("translated data", translatedData))
		}
	}()

	return &Output{Message: "start translate"}, nil
}

func (s *Service) startTranslate(
	data *dto.TranslationContext,
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
	resultChan := make(chan dto.ChapterResult, len(data.Chapters))

	go s.translateChapters(ctx, resultChan, data)

	paths := make([]string, 0, len(data.Chapters))
	for result := range resultChan {
		if result.Error != nil {
			slog.Error(
				"failed to translate chapter",
				slog.String("operation", operation),
				slog.String("title", data.Book.Title),
				slog.String("author", data.Book.Author),
				slog.Any("error", result.Error),
			)
			return ErrFailedTranslateChapter
		}
		if result.Chapter == nil && result.Path == "" {
			slog.Error(
				"translated chapter is nil",
				slog.String("operation", operation),
				slog.String("title", data.Book.Title),
				slog.String("author", data.Book.Author),
				slog.Int("chapter order", result.ChapterOrder),
			)
			return ErrFailedTranslateChapter
		}

		var path string
		if result.Path != "" {
			path = result.Path
		} else {
			var err error
			path, err = s.saveChapterToS3(result.Chapter, data.Book.Title)
			if err != nil {
				slog.Error(
					"failed to save chapter to s3",
					slog.String("operation", operation),
					slog.String("title", data.Book.Title),
					slog.String("author", data.Book.Author),
					slog.Int("chapter order", result.Chapter.Order),
					slog.Any("error", err),
				)
				return ErrFailedSaveToStorage
			}
		}

		paths = append(paths, path)
		slog.Info(fmt.Sprintf("paths is %v now", paths), slog.String("operation", operation))

		// отправляем уведомления о переведенной главе
		slog.Info(
			"chapter translate",
			slog.String("operation", operation),
			slog.String("title", data.Book.Title),
			slog.String("author", data.Book.Author),
			slog.Int("chapter order", result.ChapterOrder),
			slog.String("chapter path", path),
			slog.Int("total translated chapter count", len(paths)),
		)
	}

	// Получаем или создаём автора
	author, err := s.authorRepo.GetOrCreate(ctx, data.Book.Author)
	if err != nil {
		slog.Error("failed to get or create author", slog.String("author", data.Book.Author), slog.Any("error", err), slog.String("operation", operation))
		return ErrFailedSaveBookToDatabase
	}

	var coverURL *string
	if data.CoverData != nil {
		coverPath, err := s.saveCoverToS3(data.CoverData, data.Book.Title)
		if err != nil {
			slog.Error(err.Error(), slog.String("operation", operation))
		} else {
			coverURL = &coverPath
		}
	}

	newBook := &domain.Book{
		Author:         author,
		Title:          data.Book.Title,
		CoverURL:       coverURL,
		OriginalLang:   string(data.From),
		TranslatedLang: string(data.To),
	}
	newBook, err = s.bookRepo.Create(ctx, newBook)
	if err != nil {
		slog.Error(
			"failed to save book to database",
			slog.String("title", data.Book.Title),
			slog.String("author", data.Book.Author),
			slog.Any("error", err),
			slog.String("operation", operation),
		)
		return ErrFailedSaveBookToDatabase
	}

	// Создаём главы книги
	for i, path := range paths {
		chapter := &domain.BookChapter{
			Book:       newBook,
			Order:      i,
			ContentURL: path,
		}
		_, err := s.bookRepo.CreateChapter(ctx, chapter)
		if err != nil {
			slog.Error(
				"failed to save chapter to database",
				slog.Int("chapter order", i),
				slog.String("title", data.Book.Title),
				slog.Any("error", err),
				slog.String("operation", operation),
			)
		}
	}

	// уведомление что книга создана
	slog.Info(
		"book is translated successfully",
		slog.String("operation", operation),
		slog.Any("book", newBook),
	)

	return nil
}

func (s *Service) translateChapters(
	ctx context.Context,
	resultChan chan<- dto.ChapterResult,
	data *dto.TranslationContext,
) {
	const operation = "translate_book.Service.translateChapters"

	defer close(resultChan)

	for i, chapter := range data.Chapters {
		// chapter order будем считать как i, если брать значения из chapter.Order, то могут быть проблемы с порядком глав
		// в книге главы могут не содержать какой-то письменный контент и мы такие главы пропускаем при парсинге, а значит порядок глав будет нарушен

		select {
		case <-ctx.Done():
			slog.Debug("context cancelled", slog.String("operation", operation))
			return
		default:
		}

		path := s.checkChapterInStorage(i, data.Book.Title)
		if path != "" {
			slog.Info(fmt.Sprintf("chapter %d already exist", i), slog.String("operation", operation), slog.String("path", path))
			resultChan <- dto.ChapterResult{
				Path:         path,
				ChapterOrder: i,
				Chapter:      nil,
				Error:        nil,
			}
			continue
		}

		chapterNode, err := s.logic.TranslateChapter(
			ctx,
			chapter,
			data.From,
			data.To,
			s,
			s.config.MaxRequestCount,
			func(imgData []byte) (string, error) {
				return s.saveImageToS3(imgData, data.Book.Title)
			},
		)
		if err != nil {
			slog.Error(err.Error(), slog.String("operation", operation))
			resultChan <- dto.ChapterResult{
				Chapter:      nil,
				Error:        ErrFailedTranslateChapter,
				ChapterOrder: i,
			}
			return
		}

		chapterNode.Order = i

		resultChan <- dto.ChapterResult{
			ChapterOrder: i,
			Chapter:      chapterNode,
			Error:        nil,
		}
	}
}

func (s *Service) saveImageToS3(data []byte, bookTitle string) (string, error) {
	const operation = "translate_book.Service.saveImageToS3"

	// Generate a unique filename
	filename := uuid.New().String() + ".jpg"
	path := fmt.Sprintf("book/%s/images/%s", strings.ReplaceAll(bookTitle, " ", "_"), filename)

	if err := s.s3.Upload(path, data); err != nil {
		return "", fmt.Errorf("%s: failed upload to s3: %w", operation, err)
	}

	return path, nil
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

func (s *Service) checkChapterInStorage(chapterOrder int, bookTitle string) string {
	const operation = "translate_book.Service.checkChapterPath"

	path := fmt.Sprintf("book/%s/%d.json", strings.ReplaceAll(bookTitle, " ", "_"), chapterOrder)
	slog.Info(path, slog.String("operation", operation))

	if err := s.s3.Check(path); err != nil {
		slog.Warn(err.Error(), slog.String("operation", operation))
		return ""
	}

	return path
}
