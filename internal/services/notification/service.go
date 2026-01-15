package notification

import (
	"context"
	"fmt"
	"log/slog"

	"firebase.google.com/go/v4/messaging"
	"github.com/nimyab/nim2book-back/internal/controller/websocket"
	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/internal/repositories"
)

type Service struct {
	fcmTokenRepo            *repositories.FcmTokenRepository
	messagingFirebaseClient *messaging.Client
}

var service *Service

func New(
	messagingFirebaseClient *messaging.Client,
	fcmTokenRepo *repositories.FcmTokenRepository,
) *Service {
	service = &Service{
		fcmTokenRepo:            fcmTokenRepo,
		messagingFirebaseClient: messagingFirebaseClient,
	}
	return service
}

func (s *Service) ProcessNotification(ctx context.Context, d models.Notification) {
	const operation = "notification.ProcessNotification"

	fcmTokens, err := s.fcmTokenRepo.GetFcmTokensByUserId(ctx, d.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId))
	}

	switch d.Type {
	case models.NotificationBookTranslated:
		data, ok := d.Data.(*models.NotificationBookTranslatedData)
		if !ok {
			slog.Error(fmt.Sprintf("%s: %s", operation, "error data mapping"), slog.Any("data", d.Data), slog.Any("type", d.Type))
			return
		}

		for _, fcmToken := range fcmTokens {
			info, err := s.messagingFirebaseClient.Send(context.Background(), &messaging.Message{
				Token: fcmToken.Token,
				Notification: &messaging.Notification{
					Title: "Перевод книги завершился",
					Body:  fmt.Sprintf("Книга: %s - %s была переведена, теперь ее можно скачать из библиотеки книг", data.Book.Author, data.Book.Title),
				},
				Data: map[string]string{
					"bookId": data.Book.ID.String(),
				},
			})
			slog.Info("test notification", slog.String("fcmToken", fcmToken.Token), slog.String("info", info))
			if err != nil {
				slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId), slog.Any("type", d.Type), slog.String("fcmToken", fcmToken.Token))
				_ = s.fcmTokenRepo.DeleteFcmToken(ctx, fcmToken.Token, d.UserId)
			}
		}

		websocket.SendMessage(d.UserId, &websocket.Message{
			Event: websocket.TranslateSucceedEvent,
			Body: map[string]interface{}{
				"book": data.Book,
			},
		})
	case models.NotificationError:
		data, ok := d.Data.(*models.NotificationErrorData)
		if !ok {
			slog.Error(fmt.Sprintf("%s: %s", operation, "error data mapping"), slog.Any("data", d.Data), slog.Any("type", d.Type))
			return
		}

		for _, fcmToken := range fcmTokens {
			info, err := s.messagingFirebaseClient.Send(context.Background(), &messaging.Message{
				Token: fcmToken.Token,
				Notification: &messaging.Notification{
					Title: "Перевод книги прервался",
					Body:  fmt.Sprintf("%s\nКнига: %s - %s", data.ErrorMessage, data.Author, data.Title),
				},
			})
			slog.Info("test notification", slog.String("fcmToken", fcmToken.Token), slog.String("info", info))
			if err != nil {
				slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId), slog.Any("type", d.Type), slog.String("fcmToken", fcmToken.Token))
				_ = s.fcmTokenRepo.DeleteFcmToken(ctx, fcmToken.Token, d.UserId)
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
	case models.NotificationChapterTranslateSucceed:
		data, ok := d.Data.(*models.NotificationChapterTranslateSucceedData)
		if !ok {
			slog.Error(fmt.Sprintf("%s: %s", operation, "error data mapping"), slog.Any("data", d.Data), slog.Any("type", d.Type))
			return
		}

		for _, fcmToken := range fcmTokens {
			info, err := s.messagingFirebaseClient.Send(context.Background(), &messaging.Message{
				Token: fcmToken.Token,
				Notification: &messaging.Notification{
					Title: fmt.Sprintf("Переведена глава %d", data.ChapterOrder),
					Body:  fmt.Sprintf("Книга: %s - %s.\nПозже отправим уведомление о следующих главах", data.Author, data.Title),
				},
			})
			slog.Info("test notification", slog.String("fcmToken", fcmToken.Token), slog.String("info", info))
			if err != nil {
				slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId), slog.Any("type", d.Type), slog.String("fcmToken", fcmToken.Token))
				_ = s.fcmTokenRepo.DeleteFcmToken(ctx, fcmToken.Token, d.UserId)
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
	case models.NotificationTest:
		data, ok := d.Data.(*models.NotificationTestData)
		if !ok {
			slog.Error(fmt.Sprintf("%s: %s", operation, "error data mapping"), slog.Any("data", d.Data), slog.Any("type", d.Type))
			return
		}

		for _, fcmToken := range fcmTokens {
			info, err := s.messagingFirebaseClient.Send(context.Background(), &messaging.Message{
				Token: fcmToken.Token,
				Notification: &messaging.Notification{
					Title: data.Title,
					Body:  data.Body,
				},
			})
			slog.Info("test notification", slog.String("fcmToken", fcmToken.Token), slog.String("info", info))
			if err != nil {
				slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", d.UserId), slog.Any("type", d.Type), slog.String("fcmToken", fcmToken.Token))
				_ = s.fcmTokenRepo.DeleteFcmToken(ctx, fcmToken.Token, d.UserId)
			}
		}
	default:
		slog.Error("unknown notification type", slog.Any("type", d.Type))
	}
}
