package translate_book

import (
	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/adapter/rabbitmq"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/logger"
)

type messageAboutTranslate struct {
	chapterPath       string
	author            string
	title             string
	chapterOrder      int
	totalChapterCount int
}

func (s *Service) sendMessageAboutTranslate(userId domain.Id, message *messageAboutTranslate) {
	const operation = "translate_book.sendMessageAboutTranslate"

	err := s.rabbitmq.Publish(&rabbitmq.NotificationData{
		Id:     uuid.New().String(),
		UserId: userId,
		Data: map[string]interface{}{
			"chapterPath":       message.chapterPath,
			"author":            message.author,
			"title":             message.title,
			"order":             message.chapterOrder,
			"totalChapterCount": message.totalChapterCount,
		},
	})
	if err != nil {
		logger.Error("fail publish message", err, operation)
	}

	//websocket.SendMessage(userId, &websocket.Message{
	//	Event: websocket.ChapterTranslatedEvent,
	//	Body: map[string]any{
	//		"chapterPath":       message.chapterPath,
	//		"author":            message.author,
	//		"title":             message.title,
	//		"order":             message.chapterOrder,
	//		"totalChapterCount": message.totalChapterCount,
	//	},
	//})
	//
	//fcmTokens, err := s.pg.GetFcmTokensByUserId(context.Background(), userId)
	//if err != nil {
	//	slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", userId))
	//}
	//for _, fcmToken := range fcmTokens {
	//	_, err := s.messagingFirebaseClient.Send(context.Background(), &messaging.Message{
	//		Token: fcmToken.Token,
	//		Notification: &messaging.Notification{
	//			Title: fmt.Sprintf("Переведена глава %d", message.chapterOrder),
	//			Body:  fmt.Sprintf("Книга: %s - %s.\nПозже отправим уведомление о следующих главах", message.author, message.title),
	//		},
	//	})
	//	if err != nil {
	//		slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", userId))
	//	}
	//}
}

type messageAboutError struct {
	author       string
	title        string
	errorMessage string
}

func (s *Service) sendMessageAboutError(userId domain.Id, message *messageAboutError) {
	const operation = "translate_book.sendMessageAboutError"

	err := s.rabbitmq.Publish(&rabbitmq.NotificationData{
		Id:     uuid.New().String(),
		UserId: userId,
		Data: map[string]interface{}{
			"author": message.author,
			"title":  message.title,
			"error":  message.errorMessage,
		},
	})
	if err != nil {
		logger.Error("fail publish message", err, operation)
	}
	//websocket.SendMessage(userId, &websocket.Message{
	//	Event: websocket.ErrorEvent,
	//	Body: map[string]interface{}{
	//		"author": message.author,
	//		"title":  message.title,
	//		"error":  message.errorMessage,
	//	},
	//})
	//
	//fcmTokens, err := s.pg.GetFcmTokensByUserId(context.Background(), userId)
	//if err != nil {
	//	slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", userId))
	//}
	//for _, fcmToken := range fcmTokens {
	//	_, err := s.messagingFirebaseClient.Send(context.Background(), &messaging.Message{
	//		Token: fcmToken.Token,
	//		Notification: &messaging.Notification{
	//			Title: fmt.Sprintf("Перевод книги прервался"),
	//			Body:  fmt.Sprintf("%s\nКнига: %s - %s", message.errorMessage, message.author, message.title),
	//		},
	//	})
	//	if err != nil {
	//		slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", userId))
	//	}
	//}
}

type messageAboutTranslateSuccess struct {
	book *domain.Book
}

func (s *Service) sendMessageAboutTranslateSuccess(userId domain.Id, message *messageAboutTranslateSuccess) {
	const operation = "translate_book.sendMessageAboutTranslateSuccess"

	err := s.rabbitmq.Publish(&rabbitmq.NotificationData{
		Id:     uuid.New().String(),
		UserId: userId,
		Data: map[string]interface{}{
			"book": message.book,
		},
	})
	if err != nil {
		logger.Error("fail publish message", err, operation)
	}

	//websocket.SendMessage(userId, &websocket.Message{
	//	Event: websocket.TranslateSuccessEvent,
	//	Body: map[string]interface{}{
	//		"book": message.book,
	//	},
	//})
	//
	//fcmTokens, err := s.pg.GetFcmTokensByUserId(context.Background(), userId)
	//if err != nil {
	//	slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", userId))
	//}
	//for _, fcmToken := range fcmTokens {
	//	_, err := s.messagingFirebaseClient.Send(context.Background(), &messaging.Message{
	//		Token: fcmToken.Token,
	//		Notification: &messaging.Notification{
	//			Title: fmt.Sprintf("Перевод книги завершился"),
	//			Body:  fmt.Sprintf("Книга: %s - %s была переведена, теперь ее можно скачать из библиотеки книг", message.book.Author, message.book.Title),
	//		},
	//		Data: map[string]string{
	//			"bookId": message.book.Id.String(),
	//		},
	//	})
	//	if err != nil {
	//		slog.Error(fmt.Sprintf("%s: %s", operation, err.Error()), slog.Any("userId", userId))
	//	}
	//}
}
