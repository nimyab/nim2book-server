package rabbitmq

import (
	"encoding/json"
	"time"

	"github.com/nimyab/nim2book-back/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	NotificationExchangeName = "notification.exchange"
)

const ()

type NotificationData struct {
	Id        string                 `json:"id"`
	UserId    domain.Id              `json:"userId"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

func (r *Rabbitmq) createNotificationProducer() error {
	return r.channel.ExchangeDeclare(
		NotificationExchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
}

func (r *Rabbitmq) Publish(data *NotificationData) error {
	bytesData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err = r.channel.Publish(
		NotificationExchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         bytesData,
		},
	); err != nil {
		return err
	}
	return nil
}
