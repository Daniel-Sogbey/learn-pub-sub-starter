package pubsub

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	err := ch.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	valJson, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(ctx,
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        valJson,
		},
	)

	return err
}

type QueueType int

const (
	Durable QueueType = iota
	Transient
)

var queueTypes = map[QueueType]string{
	Durable:   "durable",
	Transient: "transient",
}

func (q QueueType) String() string {
	return queueTypes[q]
}

func DeclareAnBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType QueueType) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	var queue amqp.Queue

	if simpleQueueType == Durable {
		queue, err = ch.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			nil,
		)
	} else {
		queue, err = ch.QueueDeclare(
			queueName,
			false,
			true,
			true,
			false,
			nil,
		)
	}

	if err != nil {
		return nil, queue, err
	}

	err = ch.QueueBind(
		queueName,
		key,
		exchange,
		false,
		nil,
	)

	return ch, queue, err
}
