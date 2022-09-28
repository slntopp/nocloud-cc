package broker

import (
	pb "github.com/slntopp/nocloud-cc/pkg/chats/proto"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Broker struct {
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

	broker = Broker{log: log, ch: ch, title: "chats"}
	broker.ch.ExchangeDeclare(broker.title, "fanout", true, false, false, false, nil)
}

type MsgPub func(msg *pb.ChatMessage) error

func GetPublisher(uuid string) MsgPub {
	return func(msg *pb.ChatMessage) error {
		body, err := proto.Marshal(msg)
		if err != nil {
			return err
		}
		return broker.ch.Publish(broker.title, uuid, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	}
}

func GetConsumer(uuid string) (<-chan amqp.Delivery, error) {
	// err := broker.ch.ExchangeDeclare(
	// 	"chats", // name
	// 	"topic", // type
	// 	true,    // durable
	// 	false,   // auto-deleted
	// 	false,   // internal
	// 	false,   // no-wait
	// 	nil,     // arguments
	// )
	// if err != nil {
	// 	return nil, err
	// }

	q, err := broker.ch.QueueDeclare(
		uuid,  // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, nil
	}

	err = broker.ch.QueueBind(
		q.Name,       // queue name
		uuid,         // routing key
		broker.title, // exchange
		false,
		nil)

	msgs, err := broker.ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}
