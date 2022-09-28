package broker

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Broker struct {
	rbmq  *amqp.Connection
	log   *zap.Logger
	ch    *amqp.Channel
	title string
}

var broker Broker

const ChatExchange = "chats"

func Configure(logger *zap.Logger, rbmq *amqp.Connection) {
	log := logger.Named("ChatsExchange")
	ch, err := rbmq.Channel()
	if err != nil {
		panic("Can't create RBMQ channel")
	}

	broker = Broker{log: log, ch: ch, title: "chats", rbmq: rbmq}
	broker.ch.ExchangeDeclare(broker.title, "topic", true, false, false, false, nil)
}

type MsgPub func(msg interface{}) error

func GetPublisher(subtopic string) MsgPub {
	return func(msg interface{}) error {
		body, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		return broker.ch.Publish(broker.title, subtopic, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	}
}

func GetConsumer(ctx context.Context, uuid string) (<-chan amqp.Delivery, error) {

	payload := ctx.Value(nocloud.NoCloudAccount)
	if payload == nil {
		return nil, errors.New("empty credentials")
	}
	requestor := payload.(string)
	timestamp := time.Now().String()
	queueTitle := timestamp + requestor + uuid

	_, err := broker.ch.QueueDeclare(
		queueTitle, // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		broker.log.Warn("Failed to declare queue")
		return nil, err
	}

	err = broker.ch.QueueBind(
		queueTitle,   // queue name
		uuid,         // routing key
		broker.title, // exchange
		false,
		nil)
	if err != nil {
		broker.log.Warn("Failed to bind queue")
		return nil, err
	}

	msgs, err := broker.ch.Consume(
		queueTitle,          // queue
		timestamp+requestor, // consumer
		false,               // auto ack
		true,                // exclusive
		false,               // no local
		false,               // no wait
		nil,                 // args
	)
	if err != nil {
		broker.log.Warn("Failed to consume exchange")
		return nil, err
	}

	return msgs, nil
}
