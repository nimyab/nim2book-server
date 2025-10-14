package notification

import (
	"context"
	"fmt"
	"log/slog"

	"firebase.google.com/go/v4/messaging"
	"github.com/nimyab/nim2book-back/internal/controller/websocket"
	"github.com/nimyab/nim2book-back/internal/domain"
)

type Postgres interface {
	GetFcmTokensByUserId(ctx context.Context, userId domain.Id) ([]domain.FcmToken, error)
}

type Service struct {
	pg                      Postgres
	messagingFirebaseClient *messaging.Client
}

var service *Service

func New(
	messagingFirebaseClient *messaging.Client,
	pg Postgres,
) *Service {
	service = &Service{
		pg:                      pg,
		messagingFirebaseClient: messagingFirebaseClient,
	}
	return service
}

func (s *Service) ProcessNotification(ctx context.Context, d *domain.Notification) {
	const operation = "notification.ProcessNotification"

	fcmTokens, err := s.pg.GetFcmTokensByUserId(ctx, d.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId))
	}

	switch d.Type {
	case domain.NotificationBookTranslated:
		data, ok := d.Data.(*domain.NotificationBookTranslatedData)
		if !ok {
			slog.Error(fmt.Sprintf("%s: %s", operation, "error data mapping"), slog.Any("data", d.Data), slog.Any("type", d.Type))
		}

		for _, fcmToken := range fcmTokens {
			_, err := s.messagingFirebaseClient.Send(context.Background(), &messaging.Message{
				Token: fcmToken.Token,
				Notification: &messaging.Notification{
					Title: fmt.Sprintf("Перевод книги завершился"),
					Body:  fmt.Sprintf("Книга: %s - %s была переведена, теперь ее можно скачать из библиотеки книг", data.Book.Author, data.Book.Title),
				},
				Data: map[string]string{
					"bookId": data.Book.Id.String(),
				},
			})
			if err != nil {
				slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId), slog.Any("type", d.Type))
			}
		}

		websocket.SendMessage(d.UserId, &websocket.Message{
			Event: websocket.TranslateSucceedEvent,
			Body: map[string]interface{}{
				"book": data.Book,
			},
		})
	case domain.NotificationError:
		data, ok := d.Data.(*domain.NotificationErrorData)
		if !ok {
			slog.Error(fmt.Sprintf("%s: %s", operation, "error data mapping"), slog.Any("data", d.Data), slog.Any("type", d.Type))
		}

		for _, fcmToken := range fcmTokens {
			_, err := s.messagingFirebaseClient.Send(context.Background(), &messaging.Message{
				Token: fcmToken.Token,
				Notification: &messaging.Notification{
					Title: fmt.Sprintf("Перевод книги прервался"),
					Body:  fmt.Sprintf("%s\nКнига: %s - %s", data.ErrorMessage, data.Author, data.Title),
				},
			})
			if err != nil {
				slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId), slog.Any("type", d.Type))
			}
		}

		websocket.SendMessage(d.UserId, &websocket.Message{
			Event: websocket.ErrorEvent,
			Body: map[string]interface{}{
				"author": data.Author,
				"title":  data.Title,
				"error":  data.ErrorMessage,
			},
		})
	case domain.NotificationChapterTranslateSucceed:
		data, ok := d.Data.(*domain.NotificationChapterTranslateSucceedData)
		if !ok {
			slog.Error(fmt.Sprintf("%s: %s", operation, "error data mapping"), slog.Any("data", d.Data), slog.Any("type", d.Type))
		}

		for _, fcmToken := range fcmTokens {
			_, err := s.messagingFirebaseClient.Send(context.Background(), &messaging.Message{
				Token: fcmToken.Token,
				Notification: &messaging.Notification{
					Title: fmt.Sprintf("Переведена глава %d", data.ChapterOrder),
					Body:  fmt.Sprintf("Книга: %s - %s.\nПозже отправим уведомление о следующих главах", data.Author, data.Title),
				},
			})
			if err != nil {
				slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId), slog.Any("type", d.Type))
			}
		}

		websocket.SendMessage(d.UserId, &websocket.Message{
			Event: websocket.ChapterTranslatedEvent,
			Body: map[string]interface{}{
				"chapterPath":       data.ChapterPath,
				"author":            data.Author,
				"title":             data.Title,
				"chapterOrder":      data.ChapterOrder,
				"totalChapterCount": data.TotalChapterCount,
			},
		})
	default:
		slog.Error("unknown notification type", slog.Any("type", d.Type))
	}
}
