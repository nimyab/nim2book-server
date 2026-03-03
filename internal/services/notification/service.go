package notification

import (
	"context"
	"fmt"
	"log/slog"

	"firebase.google.com/go/v4/messaging"
	"github.com/nimyab/nim2book-back/internal/controller/websocket"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
)

type FcmTokenRepository interface {
	ListByUserID(ctx context.Context, userID domain.ID, opts repository.QueryOptions) ([]*domain.FcmToken, error)
	DeleteByToken(ctx context.Context, token string) error
}

type MessagingClient interface {
	Send(ctx context.Context, message *messaging.Message) (string, error)
}

type WebsocketSender interface {
	SendMessage(userId domain.ID, msg *websocket.Message)
}

type Service struct {
	fcmTokenRepo    FcmTokenRepository
	messagingClient MessagingClient
	websocketSender WebsocketSender
}

func New(
	messagingClient MessagingClient,
	fcmTokenRepo FcmTokenRepository,
	websocketSender WebsocketSender,
) *Service {
	return &Service{
		fcmTokenRepo:    fcmTokenRepo,
		messagingClient: messagingClient,
		websocketSender: websocketSender,
	}
}

// Emit sends notification (implements NotificationSender interface for translate services)
func (s *Service) Emit(notification *domain.Notification) {
	go s.ProcessNotification(notification)
}

func (s *Service) ProcessNotification(d *domain.Notification) {
	const operation = "notification.ProcessNotification"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fcmTokenPtrs, err := s.fcmTokenRepo.ListByUserID(ctx, d.UserId, repository.QueryOptions{})
	if err != nil {
		slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId))
		return
	}

	// Convert []*FcmToken to []FcmToken for compatibility
	fcmTokens := make([]domain.FcmToken, len(fcmTokenPtrs))
	for i, ptr := range fcmTokenPtrs {
		fcmTokens[i] = *ptr
	}

	switch d.Type {
	case domain.NotificationBookTranslated:
		data, ok := d.Data.(*domain.NotificationBookTranslatedData)
		if !ok {
			slog.Error(fmt.Sprintf("%s: %s", operation, "error data mapping"), slog.Any("data", d.Data), slog.Any("type", d.Type))
			return
		}

		for _, fcmToken := range fcmTokens {
			info, err := s.messagingClient.Send(ctx, &messaging.Message{
				Token: fcmToken.Token,
				Notification: &messaging.Notification{
					Title: "Перевод книги завершился",
					Body:  fmt.Sprintf("Книга: %s - %s была переведена, теперь ее можно скачать из библиотеки книг", data.Book.Author.Name, data.Book.Title),
				},
				Data: map[string]string{
					"bookId": data.Book.ID.String(),
				},
			})
			slog.Info("test notification", slog.String("fcmToken", fcmToken.Token), slog.String("info", info))
			if err != nil {
				slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId), slog.Any("type", d.Type), slog.String("fcmToken", fcmToken.Token))
				_ = s.fcmTokenRepo.DeleteByToken(ctx, fcmToken.Token)
			}
		}

		s.websocketSender.SendMessage(d.UserId, &websocket.Message{
			Event: websocket.TranslateSucceedEvent,
			Body: map[string]any{
				"book": data.Book,
			},
		})
	case domain.NotificationError:
		data, ok := d.Data.(*domain.NotificationErrorData)
		if !ok {
			slog.Error(fmt.Sprintf("%s: %s", operation, "error data mapping"), slog.Any("data", d.Data), slog.Any("type", d.Type))
			return
		}

		for _, fcmToken := range fcmTokens {
			info, err := s.messagingClient.Send(ctx, &messaging.Message{
				Token: fcmToken.Token,
				Notification: &messaging.Notification{
					Title: "Перевод книги прервался",
					Body:  fmt.Sprintf("%s\nКнига: %s - %s", data.ErrorMessage, data.Author, data.Title),
				},
			})
			slog.Info("test notification", slog.String("fcmToken", fcmToken.Token), slog.String("info", info))
			if err != nil {
				slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId), slog.Any("type", d.Type), slog.String("fcmToken", fcmToken.Token))
				_ = s.fcmTokenRepo.DeleteByToken(ctx, fcmToken.Token)
			}
		}

		s.websocketSender.SendMessage(d.UserId, &websocket.Message{
			Event: websocket.ErrorEvent,
			Body: map[string]any{
				"author": data.Author,
				"title":  data.Title,
				"error":  data.ErrorMessage,
			},
		})
	case domain.NotificationChapterTranslateSucceed:
		data, ok := d.Data.(*domain.NotificationChapterTranslateSucceedData)
		if !ok {
			slog.Error(fmt.Sprintf("%s: %s", operation, "error data mapping"), slog.Any("data", d.Data), slog.Any("type", d.Type))
			return
		}

		for _, fcmToken := range fcmTokens {
			info, err := s.messagingClient.Send(ctx, &messaging.Message{
				Token: fcmToken.Token,
				Notification: &messaging.Notification{
					Title: fmt.Sprintf("Переведена глава %d", data.ChapterOrder),
					Body:  fmt.Sprintf("Книга: %s - %s.\nПозже отправим уведомление о следующих главах", data.Author, data.Title),
				},
			})
			slog.Info("test notification", slog.String("fcmToken", fcmToken.Token), slog.String("info", info))
			if err != nil {
				slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId), slog.Any("type", d.Type), slog.String("fcmToken", fcmToken.Token))
				_ = s.fcmTokenRepo.DeleteByToken(ctx, fcmToken.Token)
			}
		}

		s.websocketSender.SendMessage(d.UserId, &websocket.Message{
			Event: websocket.ChapterTranslatedEvent,
			Body: map[string]any{
				"author":       data.Author,
				"title":        data.Title,
				"chapterOrder": data.ChapterOrder,
			},
		})
	case domain.NotificationTest:
		data, ok := d.Data.(*domain.NotificationTestData)
		if !ok {
			slog.Error(fmt.Sprintf("%s: %s", operation, "error data mapping"), slog.Any("data", d.Data), slog.Any("type", d.Type))
			return
		}

		for _, fcmToken := range fcmTokens {
			info, err := s.messagingClient.Send(ctx, &messaging.Message{
				Token: fcmToken.Token,
				Notification: &messaging.Notification{
					Title: data.Title,
					Body:  data.Body,
				},
			})
			slog.Info("test notification", slog.String("fcmToken", fcmToken.Token), slog.String("info", info))
			if err != nil {
				slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId), slog.Any("type", d.Type), slog.String("fcmToken", fcmToken.Token))
				_ = s.fcmTokenRepo.DeleteByToken(ctx, fcmToken.Token)
			}
		}
	default:
		slog.Error("unknown notification type", slog.Any("type", d.Type))
	}
}
