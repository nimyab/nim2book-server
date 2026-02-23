package translate_book

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
	mockBookRepo := NewMockBookRepository(t)
	mockAuthorRepo := NewMockAuthorRepository(t)
	mockTranslator := NewMockTranslator(t)
	mockWordAligner := &MockProtoWordAligner{}

	cfg := dto.Config{
		WaitDuration:    0,
		MaxRequestCount: 1,
	}

	service := New(mockS3, mockBookRepo, mockAuthorRepo, mockWordAligner, mockTranslator, cfg)

	// Data
	userID := uuid.New()
	bookTitle := "Test Book"
	authorName := "Test Author"

	chapter1 := epub_parser.FormattedChapter{
		Paragraphs: []string{"Hello world"},
	}

	chapters := []epub_parser.FormattedChapter{chapter1}

	pamphletBook := &pamphlet.Book{
		Title:  bookTitle,
		Author: authorName,
	}

	data := &dto.TranslationContext{
		Book:      pamphletBook,
		Chapters:  chapters,
		CoverData: []byte("cover"),
		UserID:    userID,
		From:      domain.En,
		To:        domain.Ru,
	}

	// Expectations

	// 1. checkChapterInStorage
	mockS3.On("Check", "book/Test_Book/0.json").Return(assert.AnError)

	// 2. logic.TranslateChapter -> Translator.Translate
	mockTranslator.On("Translate", mock.MatchedBy(func(input *translate.Input) bool {
		return input.Q == "Hello world" && input.Source == "en" && input.Target == "ru"
	})).Return(&translate.Output{
		TranslatedText: "Привет мир",
	}, nil)

	// 3. logic.TranslateChapter -> WordAligner.Align
	mockWordAligner.On("Align", mock.Anything, mock.MatchedBy(func(req *pb.AlignRequest) bool {
		return req.SourceText == "Hello world" && req.TargetText == "Привет мир"
	}), mock.Anything).Return(&pb.AlignResponse{
		Alignments: []*pb.AlignmentResult{
			{SrcWord: "Hello", TargetWord: "Привет"},
			{SrcWord: "world", TargetWord: "мир"},
		},
	}, nil)

	// 4. saveChapterToS3
	mockS3.On("Upload", "book/Test_Book/0.json", mock.Anything).Return(nil)

	// 5. AuthorRepo.GetOrCreate
	mockAuthorRepo.On("GetOrCreate", mock.Anything, authorName).Return(&domain.Author{
		Name: authorName,
	}, nil)

	// 6. saveCoverToS3
	mockS3.On("Upload", mock.MatchedBy(func(path string) bool {
		// cover/Test_Book/<uuid>
		return len(path) > 0
	}), []byte("cover")).Return(nil)

	// 7. BookRepo.Create
	mockBookRepo.On("Create", mock.Anything, mock.MatchedBy(func(b *domain.Book) bool {
		return b.Title == bookTitle && b.Author.Name == authorName
	})).Return(&domain.Book{
		ID:     uuid.New(),
		Title:  bookTitle,
		Author: &domain.Author{Name: authorName},
	}, nil)

	// 8. BookRepo.CreateChapter
	mockBookRepo.On("CreateChapter", mock.Anything, mock.MatchedBy(func(c *domain.BookChapter) bool {
		return c.Order == 0 && c.ContentURL == "book/Test_Book/0.json"
	})).Return(&domain.BookChapter{}, nil)

	// Execute
	err := service.startTranslate(data)

	// Assert
	assert.NoError(t, err)
	mockS3.AssertExpectations(t)
	mockBookRepo.AssertExpectations(t)
	mockAuthorRepo.AssertExpectations(t)
	mockTranslator.AssertExpectations(t)
	mockWordAligner.AssertExpectations(t)
}
