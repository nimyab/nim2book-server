package rabbitmq

import (
	"encoding/json"
	"log/slog"
)

func CreateNotificationConsumer(rabbitmq *Rabbitmq, serviceName string, callback func(NotificationData)) error {
	const operation = "rabbitmq.CreateNotificationConsumer"

	queueName := serviceName + "_queue"

	q, err := rabbitmq.channel.QueueDeclare(
		queueName,
		true, false, false, false,
		nil,
	)
	if err != nil {
		return err
	}

	if err = rabbitmq.channel.QueueBind(
		q.Name,
		"",
		NotificationExchangeName,
		false,
		nil,
	); err != nil {
		return err
	}

	messages, err := rabbitmq.channel.Consume(
		q.Name,
		"",
		false, false, false, false, nil,
	)

	go func() {
		for message := range messages {
			var notificationData NotificationData
			if err := json.Unmarshal(message.Body, &notificationData); err != nil {
				slog.Error(
					"message processing",
					slog.String("error", err.Error()),
					slog.String("body", string(message.Body)),
					slog.String("operation", operation),
				)
			}
			callback(notificationData)
		}
	}()

	return nil
}
