package translate_personal_user_book

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/services/libretranslate/translate"
	"github.com/nimyab/nim2book-back/internal/services/translate/dto"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
	pb "github.com/nimyab/nim2book-back/proto/word_aligner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/timsims/pamphlet"
	"google.golang.org/grpc"
)

// MockProtoWordAligner implements pb.AlignmentServiceClient
type MockProtoWordAligner struct {
	mock.Mock
}

func (m *MockProtoWordAligner) Align(ctx context.Context, in *pb.AlignRequest, opts ...grpc.CallOption) (*pb.AlignResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AlignResponse), args.Error(1)
}

func TestStartTranslate(t *testing.T) {
	// Setup
	mockS3 := NewMockS3(t)
	mockRepo := NewMockPersonalBookRepository(t)
	mockAuthorRepo := NewMockAuthorRepository(t)
	mockTranslator := NewMockTranslator(t)
	mockWordAligner := &MockProtoWordAligner{}
	mockNotification := NewMockNotificationSender(t)

	cfg := dto.Config{
		WaitDuration:    0,
		MaxRequestCount: 1,
	}

	service := New(mockS3, mockRepo, mockAuthorRepo, mockWordAligner, mockTranslator, cfg, mockNotification)

	// Data
	userID := uuid.New()
	bookTitle := "Test Book"
	authorName := "Test Author"

	chapter1 := epub_parser.FormattedChapter{
		Content: []epub_parser.ContentUnit{
			{
				Type:     epub_parser.ContentTypeText,
				TextNode: &epub_parser.TextUnit{Text: "Hello world"},
			},
		},
	}

	chapters := []epub_parser.FormattedChapter{chapter1}

	pamphletBook := &pamphlet.Book{
		Title:  bookTitle,
		Author: authorName,
	}

	data := &dto.TranslationContext{
		Book:         pamphletBook,
		Chapters:     chapters,
		CoverData:    []byte("cover"),
		UserID:       userID,
		From:         domain.En,
		To:           domain.Ru,
		PersonalBook: nil, // New book
	}

	// Expectations

	// 1. saveCoverToS3
	mockS3.On("Upload", mock.MatchedBy(func(path string) bool {
		// cover/Test_Book/<uuid>
		return len(path) > 0
	}), []byte("cover")).Return(nil)

	// 2. AuthorRepo.GetOrCreate
	mockAuthorRepo.On("GetOrCreate", mock.Anything, authorName).Return(&domain.Author{
		Name: authorName,
	}, nil)

	// 3. BookRepo.Create (PersonalBook)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(b *domain.PersonalBook) bool {
		return b.Title == bookTitle && b.Author.Name == authorName && b.User.ID == userID
	})).Return(&domain.PersonalBook{
		ID:     uuid.New(),
		Title:  bookTitle,
		Author: &domain.Author{Name: authorName},
		User:   &domain.User{ID: userID},
	}, nil)

	// 4. GetChapterByPersonalBookIDAndOrder (in translateChapters)
	mockRepo.On("GetChapterByPersonalBookIDAndOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	// 5. checkChapterInStorage (in translateChapters)
	// Need to check how checkChapterInStorage is implemented in this service
	// likely similar to translate_book but with userID in path?
	// Let's assume path logic.
	mockS3.On("Check", mock.Anything).Return(assert.AnError) // Not found

	// 5. logic.TranslateChapter -> Translator.Translate
	mockTranslator.On("Translate", mock.MatchedBy(func(input *translate.Input) bool {
		return input.Q == "Hello world" && input.Source == "en" && input.Target == "ru"
	})).Return(&translate.Output{
		TranslatedText: "Привет мир",
	}, nil)

	// 6. logic.TranslateChapter -> WordAligner.Align
	mockWordAligner.On("Align", mock.Anything, mock.MatchedBy(func(req *pb.AlignRequest) bool {
		return req.SourceText == "Hello world" && req.TargetText == "Привет мир"
	}), mock.Anything).Return(&pb.AlignResponse{
		Alignments: []*pb.AlignmentResult{
			{SrcWord: "Hello", TargetWord: "Привет"},
			{SrcWord: "world", TargetWord: "мир"},
		},
	}, nil)

	// 7. saveChapterToS3
	mockS3.On("Upload", mock.MatchedBy(func(path string) bool {
		// book/Test_Book/<userID>/0.json (maybe? check service.go)
		return true
	}), mock.Anything).Return(nil)

	// 8. BookRepo.CreateChapter
	mockRepo.On("CreateChapter", mock.Anything, mock.MatchedBy(func(c *domain.PersonalBookChapter) bool {
		return c.Order == 0 && c.Title == ""
	})).Return(&domain.PersonalBookChapter{}, nil)

	// 9. Notification (Chapter success)
	mockNotification.On("Emit", mock.Anything, mock.MatchedBy(func(n *domain.Notification) bool {
		return n.Type == domain.NotificationChapterTranslateSucceed
	}))

	// 10. Notification (Book success)
	mockNotification.On("Emit", mock.Anything, mock.MatchedBy(func(n *domain.Notification) bool {
		return n.Type == domain.NotificationPersonalBookTranslated
	}))

	// Execute
	err := service.startTranslate(data)

	// Assert
	assert.NoError(t, err)
	mockS3.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockAuthorRepo.AssertExpectations(t)
	mockTranslator.AssertExpectations(t)
	mockWordAligner.AssertExpectations(t)
	mockNotification.AssertExpectations(t)
}
