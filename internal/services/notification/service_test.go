package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/controller/websocket"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
	"github.com/stretchr/testify/mock"
)

func TestProcessNotification(t *testing.T) {
	userID := uuid.New()
	bookID := uuid.New()
	tests := []struct {
		name         string
		notification *domain.Notification
		mockRepo     func(*MockFcmTokenRepository)
		mockClient   func(*MockMessagingClient)
		mockWs       func(*MockWebsocketSender)
	}{
		{
			name: "Success Book Translated",
			notification: &domain.Notification{
				UserId: userID,
				Type:   domain.NotificationBookTranslated,
				Data: &domain.NotificationBookTranslatedData{
					Book: &domain.Book{
						ID:     bookID,
						Title:  "Book Title",
						Author: &domain.Author{Name: "Author Name"},
					},
				},
			},
			mockRepo: func(m *MockFcmTokenRepository) {
				m.On("ListByUserID", mock.Anything, userID, repository.QueryOptions{}).Return([]*domain.FcmToken{
					{Token: "token1"},
				}, nil)
			},
			mockClient: func(m *MockMessagingClient) {
				m.On("Send", mock.Anything, mock.MatchedBy(func(msg *messaging.Message) bool {
					return msg.Token == "token1" && msg.Notification.Title == "Перевод книги завершился"
				})).Return("message-id", nil)
			},
			mockWs: func(m *MockWebsocketSender) {
				m.On("SendMessage", userID, mock.MatchedBy(func(msg *websocket.Message) bool {
					return msg.Event == websocket.TranslateSucceedEvent
				})).Return()
			},
		},
		{
			name: "Success Error Notification",
			notification: &domain.Notification{
				UserId: userID,
				Type:   domain.NotificationError,
				Data: &domain.NotificationErrorData{
					ErrorMessage: "Translation failed",
					Author:       "Author Name",
					Title:        "Book Title",
				},
			},
			mockRepo: func(m *MockFcmTokenRepository) {
				m.On("ListByUserID", mock.Anything, userID, repository.QueryOptions{}).Return([]*domain.FcmToken{
					{Token: "token1"},
				}, nil)
			},
			mockClient: func(m *MockMessagingClient) {
				m.On("Send", mock.Anything, mock.MatchedBy(func(msg *messaging.Message) bool {
					return msg.Token == "token1" && msg.Notification.Title == "Перевод книги прервался"
				})).Return("message-id", nil)
			},
			mockWs: func(m *MockWebsocketSender) {
				m.On("SendMessage", userID, mock.MatchedBy(func(msg *websocket.Message) bool {
					return msg.Event == websocket.ErrorEvent
				})).Return()
			},
		},
		{
			name: "Repo Error",
			notification: &domain.Notification{
				UserId: userID,
				Type:   domain.NotificationBookTranslated,
				Data: &domain.NotificationBookTranslatedData{
					Book: &domain.Book{ID: bookID},
				},
			},
			mockRepo: func(m *MockFcmTokenRepository) {
				m.On("ListByUserID", mock.Anything, userID, repository.QueryOptions{}).Return(nil, errors.New("db error"))
			},
			mockClient: func(m *MockMessagingClient) {
				// Should not be called
			},
			mockWs: func(m *MockWebsocketSender) {
				// Should not be called
			},
		},
		{
			name: "Send Error",
			notification: &domain.Notification{
				UserId: userID,
				Type:   domain.NotificationBookTranslated,
				Data: &domain.NotificationBookTranslatedData{
					Book: &domain.Book{
						ID:     bookID,
						Title:  "Book Title",
						Author: &domain.Author{Name: "Author Name"},
					},
				},
			},
			mockRepo: func(m *MockFcmTokenRepository) {
				m.On("ListByUserID", mock.Anything, userID, repository.QueryOptions{}).Return([]*domain.FcmToken{
					{Token: "token1"},
				}, nil)
				m.On("DeleteByToken", mock.Anything, "token1").Return(nil)
			},
			mockClient: func(m *MockMessagingClient) {
				m.On("Send", mock.Anything, mock.Anything).Return("", errors.New("send error"))
			},
			mockWs: func(m *MockWebsocketSender) {
				m.On("SendMessage", userID, mock.MatchedBy(func(msg *websocket.Message) bool {
					return msg.Event == websocket.TranslateSucceedEvent
				})).Return()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockFcmTokenRepository(t)
			mockClient := new(MockMessagingClient)
			mockWs := new(MockWebsocketSender)

			if tt.mockRepo != nil {
				tt.mockRepo(mockRepo)
			}
			if tt.mockClient != nil {
				tt.mockClient(mockClient)
			}
			if tt.mockWs != nil {
				tt.mockWs(mockWs)
			}

			service := New(mockClient, mockRepo, mockWs)

			// ProcessNotification runs in a goroutine in Emit, but here we call it directly for testing synchronously
			// or we can call Emit and wait a bit, but direct call is better for unit testing logic
			service.ProcessNotification(context.Background(), tt.notification)

			// Allow some time for async operations if any (though we are calling ProcessNotification directly)
			time.Sleep(10 * time.Millisecond)
		})
	}
}
