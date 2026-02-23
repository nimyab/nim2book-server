package translate_personal_user_book

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
	logic "github.com/nimyab/nim2book-back/internal/services/translate/translated_logic"
	"github.com/nimyab/nim2book-back/pkg/logger"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
	pb "github.com/nimyab/nim2book-back/proto/word_aligner"
)

type S3 interface {
	Upload(path string, data []byte) error
	Check(path string) error
}

type PersonalBookRepository interface {
	GetByUserAndAuthorAndTitle(ctx context.Context, userID domain.ID, authorName, title string) (*domain.PersonalBook, error)
	Create(ctx context.Context, book *domain.PersonalBook) (*domain.PersonalBook, error)
	CreateChapter(ctx context.Context, chapter *domain.PersonalBookChapter) (*domain.PersonalBookChapter, error)
	GetChapterByPersonalBookIDAndOrder(ctx context.Context, personalBookID domain.ID, orderChapter int) (*domain.PersonalBookChapter, error)
}

type AuthorRepository interface {
	GetOrCreate(ctx context.Context, name string) (*domain.Author, error)
}

type WordAligner interface {
	Align(ctx context.Context, req *pb.AlignRequest) (*pb.AlignResponse, error)
}

type Translator interface {
	Translate(input *translate.Input) (*translate.Output, error)
}

type NotificationSender interface {
	Emit(ctx context.Context, notification *domain.Notification)
}

type Service struct {
	config dto.Config

	mu                          sync.Mutex
	currentCountBookTranslating int

	s3               S3
	personalBookRepo PersonalBookRepository
	authorRepo       AuthorRepository
	wordAligner      pb.AlignmentServiceClient
	translator       Translator

	notificationSignal NotificationSender

	logic *logic.Logic
}

var (
	ErrFailedTranslateChapter   = errors.New("failed translate chapter")
	ErrFailedSaveToStorage      = errors.New("failed save to storage")
	ErrFailedSaveBookToDatabase = errors.New("failed save book to database")
	ErrFailedGetOrCreateAuthor  = errors.New("failed get or create author")
)

func New(
	s3 S3,
	personalBookRepo PersonalBookRepository,
	authorRepo AuthorRepository,
	wordAligner pb.AlignmentServiceClient,
	translator Translator,
	config dto.Config,
	notificationSignal NotificationSender,
) *Service {
	return &Service{
		s3:                          s3,
		personalBookRepo:            personalBookRepo,
		authorRepo:                  authorRepo,
		wordAligner:                 wordAligner,
		translator:                  translator,
		config:                      config,
		currentCountBookTranslating: 0,
		notificationSignal:          notificationSignal,
		logic:                       logic.New(translator, wordAligner),
	}
}

func (s *Service) Throttle() {
	time.Sleep(time.Duration(s.currentCountBookTranslating) * s.config.WaitDuration)
}

func (s *Service) TranslatePersonalUserBook(ctx context.Context, input *Input, book *multipart.FileHeader, userId domain.ID) (*Output, error) {
	const operation = "translate_personal_user_book.Service.TranslatePersonalUserBook"

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

	existedBook, err := s.personalBookRepo.GetByUserAndAuthorAndTitle(ctx, userId, parsedBook.Author, parsedBook.Title)
	if existedBook != nil && (existedBook.ProcessStatus == domain.ProcessStatusCompleted || existedBook.ProcessStatus == domain.ProcessStatusInProgress) {
		return &Output{Book: existedBook}, nil
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, errors.New("failed to find book")
	}

	translatedData := &dto.TranslationContext{
		Book:         parsedBook,
		Chapters:     chapters,
		CoverData:    coverData,
		UserID:       userId,
		From:         input.From,
		To:           input.To,
		PersonalBook: existedBook,
	}

	go func() {
		err := s.startTranslate(translatedData)
		if err != nil {
			switch {
			case errors.Is(err, ErrFailedSaveToStorage):
				s.notificationSignal.Emit(context.Background(), &domain.Notification{
					UserId: userId,
					Type:   domain.NotificationError,
					Data: &domain.NotificationErrorData{
						Title:        parsedBook.Title,
						Author:       parsedBook.Author,
						ErrorMessage: "Произошла ошибка во время сохранения главы, попробуйте перевести книгу позже.",
					},
				})
			case errors.Is(err, ErrFailedTranslateChapter):
				s.notificationSignal.Emit(context.Background(), &domain.Notification{
					UserId: userId,
					Type:   domain.NotificationError,
					Data: &domain.NotificationErrorData{
						Author:       parsedBook.Author,
						Title:        parsedBook.Title,
						ErrorMessage: "Произошла ошибка во время перевода главы, попробуйте перевести книгу позже.",
					},
				})
			case errors.Is(err, ErrFailedSaveBookToDatabase):
				s.notificationSignal.Emit(context.Background(), &domain.Notification{
					Type:   domain.NotificationError,
					UserId: userId,
					Data: &domain.NotificationErrorData{
						Title:        parsedBook.Title,
						Author:       parsedBook.Author,
						ErrorMessage: "Произошла ошибка во время сохранения книги, попробуйте перевести книгу позже, извините за неудобства.",
					},
				})
			default:
				logger.Error("unexpected error", err, operation)
			}
		}
	}()

	return &Output{Message: "start translate"}, nil
}

func (s *Service) startTranslate(
	data *dto.TranslationContext,
) error {
	const operation = "translate_personal_user_book.startTranslate"

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

	var coverURL *string
	if data.CoverData != nil {
		coverPath, err := s.saveCoverToS3(data.CoverData, data.Book.Title, data.UserID)
		if err != nil {
			slog.Error(err.Error(), slog.String("operation", operation))
		} else {
			coverURL = &coverPath
		}
	}

	// Получаем или создаём автора
	author, err := s.authorRepo.GetOrCreate(ctx, data.Book.Author)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to get or create author: %s", data.Book.Author),
			err,
			operation,
		)
		return ErrFailedGetOrCreateAuthor
	}

	newBook := data.PersonalBook
	// если книга не существует, то создаем её
	if newBook == nil {
		newBook = &domain.PersonalBook{
			Author:         author,
			Title:          data.Book.Title,
			CoverURL:       coverURL,
			OriginalLang:   string(data.From),
			TranslatedLang: string(data.To),
			User: &domain.User{
				ID: data.UserID,
			},
		}
		newBook, err = s.personalBookRepo.Create(ctx, newBook)
		if err != nil {
			logger.Error(
				fmt.Sprintf("failed to create personal book to database, title: %s, author: %s", data.Book.Title, data.Book.Author),
				err,
				operation,
			)
			return ErrFailedSaveBookToDatabase
		}
		data.PersonalBook = newBook
	}

	// use buffer chan to saveChapterToS3 non block translateChapters goroutine
	resultChan := make(chan dto.ChapterResult, len(data.Chapters))

	go s.translateChapters(ctx, resultChan, data)

	for result := range resultChan {
		if result.ExistChapter != nil {
			slog.Info(fmt.Sprintf("chapter %d already exist", result.ChapterOrder), slog.String("operation", operation), slog.Any("chapter", result.ExistChapter))
			continue
		}

		if result.Error != nil {
			logger.Error(
				fmt.Sprintf("failed to translate chapter, title: %s, author: %s", data.Book.Title, data.Book.Author),
				result.Error,
				operation,
			)
			return ErrFailedTranslateChapter
		}
		if result.Chapter == nil && result.Path == "" {
			logger.Error(
				fmt.Sprintf("translated chapter is nil, title: %s, author: %s", data.Book.Title, data.Book.Author),
				result.Error,
				operation,
			)
			return ErrFailedTranslateChapter
		}

		var path string
		if result.Path != "" {
			path = result.Path
		} else {
			var err error
			path, err = s.saveChapterToS3(result.Chapter, data.Book.Title, data.UserID)
			if err != nil {
				logger.Error(
					fmt.Sprintf("failed to save to s3 chapter %d order, title: %s, author: %s", result.Chapter.Order, data.Book.Title, data.Book.Author),
					err,
					operation,
				)
				return ErrFailedSaveToStorage
			}
		}

		chapter := &domain.PersonalBookChapter{
			PersonalBook:    newBook,
			Order:           result.ChapterOrder,
			ContentURL:      path,
			Title:           result.Chapter.Title,
			TranslatedTitle: result.Chapter.TranslatedTitle,
		}
		newChapter, err := s.personalBookRepo.CreateChapter(ctx, chapter)
		if err != nil {
			logger.Error(
				fmt.Sprintf("failed to create personal book chapter, title: %s, author: %s", data.Book.Title, data.Book.Author),
				err,
				operation,
			)
		}
		slog.Info("chapter created and added", slog.String("operation", operation), slog.Any("chapter", newChapter))

		// отправляем уведомления о переведенной главе
		s.notificationSignal.Emit(ctx, &domain.Notification{
			UserId: data.UserID,
			Type:   domain.NotificationChapterTranslateSucceed,
			Data: &domain.NotificationChapterTranslateSucceedData{
				Author:            data.Book.Author,
				ChapterPath:       path,
				Title:             data.Book.Title,
				ChapterOrder:      result.ChapterOrder,
				TotalChapterCount: len(data.Book.Chapters),
			},
		})
	}

	// уведомление что книга создана
	s.notificationSignal.Emit(ctx, &domain.Notification{
		UserId: data.UserID,
		Type:   domain.NotificationPersonalBookTranslated,
		Data: &domain.NotificationPersonalBookTranslatedData{
			Book: newBook,
		},
	})

	return nil
}

func (s *Service) translateChapters(
	ctx context.Context,
	resultChan chan<- dto.ChapterResult,
	data *dto.TranslationContext,
) {
	const operation = "translate_personal_user_book.Service.translateChapters"

	defer close(resultChan)

	for i, chapter := range data.Chapters {
		select {
		case <-ctx.Done():
			slog.Debug("context cancelled", slog.String("operation", operation))
			return
		default:
		}

		existChapter, err := s.personalBookRepo.GetChapterByPersonalBookIDAndOrder(ctx, data.PersonalBook.ID, i)
		if existChapter != nil {
			resultChan <- dto.ChapterResult{
				ExistChapter: existChapter,
				ChapterOrder: i,
				Error:        nil,
			}
			continue
		}
		if err != nil {
			slog.Error(err.Error(), slog.String("operation", operation))
			continue
		}

		path := s.checkChapterInStorage(i, data.Book.Title, data.UserID)
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

		chapterNode, err := s.logic.TranslateChapter(ctx, chapter, data.From, data.To, s, s.config.MaxRequestCount)
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

func (s *Service) saveChapterToS3(chapter *domain.ChapterAlignNode, bookTitle string, userId domain.ID) (string, error) {
	const operation = "translate_personal_user_book.Service.saveChapterToS3"

	path := fmt.Sprintf("user/%s/book/%s/%d.json", userId, strings.ReplaceAll(bookTitle, " ", "_"), chapter.Order)

	data, err := json.Marshal(chapter)
	if err != nil {
		return "", fmt.Errorf("%s: failed to marshal chapter: %w", operation, err)
	}

	if err := s.s3.Upload(path, data); err != nil {
		return "", fmt.Errorf("%s: failed upload to s3: %w", operation, err)
	}

	return path, nil
}

func (s *Service) saveCoverToS3(coverData []byte, bookTitle string, userId domain.ID) (string, error) {
	const operation = "translate_personal_user_book.Service.saveCoverToS3"

	path := fmt.Sprintf("user/%s/cover/%s/%s", userId, strings.ReplaceAll(bookTitle, " ", "_"), uuid.New().String())

	if err := s.s3.Upload(path, coverData); err != nil {
		return "", fmt.Errorf("%s: failed upload to s3: %w", operation, err)
	}

	return path, nil
}

func (s *Service) checkChapterInStorage(chapterOrder int, bookTitle string, userId domain.ID) string {
	const operation = "translate_personal_user_book.Service.checkChapterPath"

	path := fmt.Sprintf("user/%s/book/%s/%d.json", userId, strings.ReplaceAll(bookTitle, " ", "_"), chapterOrder)
	slog.Info(path, slog.String("operation", operation))

	if err := s.s3.Check(path); err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return ""
	}

	return path
}
