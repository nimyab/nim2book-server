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
	"github.com/nimyab/nim2book-back/internal/adapter/postgres_sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/services/libretranslate/translate"
	"github.com/nimyab/nim2book-back/internal/services/word_aligner/align"
	"github.com/nimyab/nim2book-back/pkg/contains_letters"
	"github.com/nimyab/nim2book-back/pkg/logger"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
	"golang.org/x/sync/errgroup"
)

type S3 interface {
	Upload(path string, data []byte) error
	Check(path string) error
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

type NotificationSender interface {
	Emit(ctx context.Context, notification *domain.Notification)
}

type Service struct {
	maxRequestCount int
	waitDuration    time.Duration

	mu                          sync.Mutex
	currentCountBookTranslating int

	s3          S3
	pg          Postgres
	wordAligner WordAligner
	translator  Translator

	notificationSignal NotificationSender
}

type translatedChapterResult struct {
	Chapter      *domain.ChapterAlignNode
	ChapterOrder int
	Path         string
	Error        error
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
	maxRequestCount int,
	waitDuration time.Duration,
	notificationSignal NotificationSender,
) *Service {
	service = &Service{
		s3:                          s3,
		pg:                          pg,
		wordAligner:                 wordAligner,
		translator:                  translator,
		maxRequestCount:             maxRequestCount,
		waitDuration:                waitDuration,
		currentCountBookTranslating: 0,
		notificationSignal:          notificationSignal,
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
	if err != nil && !errors.Is(err, postgres_sqlc.ErrBookNotFound) {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, errors.New("failed to find book")
	}

	translatedData := &translateStruct{
		Book:      parsedBook,
		Chapters:  chapters,
		CoverData: coverData,
		UserId:    userId,
		From:      input.From,
		To:        input.To,
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
	data *translateStruct,
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
	resultChan := make(chan translatedChapterResult, len(data.Chapters))

	go s.translateChapters(ctx, resultChan, data)

	paths := make([]string, 0, len(data.Chapters))
	for result := range resultChan {
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
			path, err = s.saveChapterToS3(result.Chapter, data.Book.Title)
			if err != nil {
				logger.Error(
					fmt.Sprintf("failed to save to s3 chapter %d order, title: %s, author: %s", result.Chapter.Order, data.Book.Title, data.Book.Author),
					err,
					operation,
				)
				return ErrFailedSaveToStorage
			}
		}

		paths = append(paths, path)
		slog.Info(fmt.Sprintf("paths is %v now", paths), slog.String("operation", operation))

		// отправляем уведомления о переведенной главе
		s.notificationSignal.Emit(ctx, &domain.Notification{
			UserId: data.UserId,
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

	var cover *string = nil
	if data.CoverData != nil {
		coverPath, err := s.saveCoverToS3(data.CoverData, data.Book.Title)
		if err != nil {
			slog.Error(err.Error(), slog.String("operation", operation))
		} else {
			cover = &coverPath
		}
	}

	newBook := &domain.Book{
		Author:       data.Book.Author,
		Title:        data.Book.Title,
		ChapterPaths: paths,
		Cover:        cover,
	}
	newBook, err := s.pg.CreateBook(context.Background(), newBook)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to save book to database, title: %s, author: %s", data.Book.Title, data.Book.Author),
			err,
			operation,
		)
		return ErrFailedSaveBookToDatabase
	}

	// уведомление что книга создана
	s.notificationSignal.Emit(ctx, &domain.Notification{
		UserId: data.UserId,
		Type:   domain.NotificationBookTranslated,
		Data: &domain.NotificationBookTranslatedData{
			Book: newBook,
		},
	})

	return nil
}

func (s *Service) translateChapters(
	ctx context.Context,
	resultChan chan<- translatedChapterResult,
	data *translateStruct,
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
			resultChan <- translatedChapterResult{
				Path:         path,
				ChapterOrder: i,
				Chapter:      nil,
				Error:        nil,
			}
			continue
		}

		startTime := time.Now()

		translatedChapter := make([]domain.ParagraphAlignNode, len(chapter.Paragraphs))

		g, ctxErrGroup := errgroup.WithContext(ctx)
		// set limit to prevent ddos translator and word aligner services
		g.SetLimit(s.maxRequestCount)

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
					slog.Int("chapter order", i),
					slog.Int("paragraph index", idx),
					slog.String("operation", operation),
				)

				alignedParagraph, err := s.translateAndAlignParagraph(paragraph, data.From, data.To)
				if err != nil {
					return err
				}
				translatedChapter[idx] = alignedParagraph

				slog.Info(
					"translated paragraph",
					slog.Int("chapter order", i),
					slog.Int("paragraph index", idx),
					slog.String("duration", time.Since(startTranslateParagraphTime).String()),
					slog.String("operation", operation),
				)

				return nil
			})
		}

		if err := g.Wait(); err != nil {
			slog.Error(err.Error(), slog.String("operation", operation))
			resultChan <- translatedChapterResult{
				Chapter:      nil,
				Error:        ErrFailedTranslateChapter,
				ChapterOrder: i,
			}
			return
		}

		var translatedTitle string
		if len(chapter.Title) > 0 {
			translatedTitleOutput, err := s.translator.Translate(&translate.Input{
				Q:      chapter.Title,
				Source: data.From,
				Target: data.To,
			})
			if err != nil {
				slog.Error(err.Error(), slog.String("operation", operation))
				resultChan <- translatedChapterResult{
					ChapterOrder: i,
					Chapter:      nil,
					Error:        ErrFailedTranslateChapter,
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
			ChapterOrder: i,
			Chapter:      &chapterNode,
			Error:        nil,
		}

		duration := time.Since(startTime)
		slog.Info(
			fmt.Sprintf("translated chapter %d", chapterNode.Order),
			slog.String("duration", duration.String()),
			slog.String("title", data.Book.Title),
			slog.String("author", data.Book.Author),
		)
	}

}

func (s *Service) translateAndAlignParagraph(paragraph string, from domain.SupportedLang, to domain.SupportedLang) (domain.ParagraphAlignNode, error) {
	const operation = "translate_book.Service.translateAndAlignParagraph"

	// перевод параграфа (использую libretranslate, которую развернул на серваке у себя)
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
	time.Sleep(time.Duration(s.currentCountBookTranslating) * s.waitDuration)

	// выравнивание слов (самописный выравниватель на змеинном, можно в docker-compose найти образ)
	alignOutput, err := s.wordAligner.Align(&align.Input{
		SourceText: paragraph,
		TargetText: translateOutput.TranslatedText,
	})
	time.Sleep(time.Duration(s.currentCountBookTranslating) * s.waitDuration)

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

func (s *Service) checkChapterInStorage(chapterOrder int, bookTitle string) string {
	const operation = "translate_book.Service.checkChapterPath"

	path := fmt.Sprintf("book/%s/%d.json", strings.ReplaceAll(bookTitle, " ", "_"), chapterOrder)
	slog.Info(path, slog.String("operation", operation))

	if err := s.s3.Check(path); err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return ""
	}

	return path
}
