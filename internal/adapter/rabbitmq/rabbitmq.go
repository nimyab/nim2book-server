package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	RabbitmqUrl string
}

type Rabbitmq struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func New(cfg *Config) (*Rabbitmq, error) {
	conn, err := amqp.Dial(cfg.RabbitmqUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()

	rabbit := &Rabbitmq{
		conn:    conn,
		channel: channel,
	}
	if err = rabbit.createNotificationProducer(); err != nil {
		return nil, fmt.Errorf("failed to create notification provider: %w", err)
	}

	return rabbit, nil
}

func (r *Rabbitmq) Close() error {
	if r.channel != nil {
		return r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
