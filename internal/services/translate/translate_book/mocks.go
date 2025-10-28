package translate_book

import (
	"context"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/services/libretranslate/translate"
	"github.com/nimyab/nim2book-back/internal/services/word_aligner/align"
	"github.com/stretchr/testify/mock"
)

// MockS3 is a mock implementation of S3 interface
type MockS3 struct {
	mock.Mock
}

func (m *MockS3) Upload(path string, data []byte) error {
	args := m.Called(path, data)
	return args.Error(0)
}

func (m *MockS3) Check(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

// MockPostgres is a mock implementation of Postgres interface
type MockPostgres struct {
	mock.Mock
}

func (m *MockPostgres) GetBookByAuthorAndTitle(ctx context.Context, author, title string) (*domain.Book, error) {
	args := m.Called(ctx, author, title)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Book), args.Error(1)
}

func (m *MockPostgres) CreateBook(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	args := m.Called(ctx, book)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Book), args.Error(1)
}

func (m *MockPostgres) GetFcmTokensByUserId(ctx context.Context, userId domain.Id) ([]domain.FcmToken, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.FcmToken), args.Error(1)
}

// MockWordAligner is a mock implementation of WordAligner interface
type MockWordAligner struct {
	mock.Mock
}

func (m *MockWordAligner) Align(input *align.Input) (*align.Output, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*align.Output), args.Error(1)
}

// MockTranslator is a mock implementation of Translator interface
type MockTranslator struct {
	mock.Mock
}

func (m *MockTranslator) Translate(input *translate.Input) (*translate.Output, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*translate.Output), args.Error(1)
}

// MockNotificationSender is a mock implementation of NotificationSender interface
type MockNotificationSender struct {
	mock.Mock
}

func (m *MockNotificationSender) Emit(ctx context.Context, notification *domain.Notification) {
	m.Called(ctx, notification)
}
