package broker

import (
	"encoding/json"

	pubsub "github.com/slntopp/nocloud/pkg/instances"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

var (
	log *zap.Logger
	ps  *pubsub.PubSub
	ch  *amqp.Channel
)

const ChatExchange = "chats"

func Configure(logger *zap.Logger, rbmq *amqp.Connection) {
	log = logger.Named("ChatsExchange")
	ps = pubsub.NewPubSub(log, nil, rbmq)
	ch = ps.Channel()
	ps.TopicExchange(ch, ChatExchange)
}

func MessageConsumer(uuid string) <-chan amqp.Delivery {
	topic := ChatExchange + "." + uuid
	q, err := ch.QueueDeclare(
		topic, false, false, true, false, nil,
	)
	if err != nil {
		log.Fatal("Failed to declare a queue", zap.Error(err))
	}

	err = ch.QueueBind(q.Name, topic, ChatExchange, false, nil)
	if err != nil {
		log.Fatal("Failed to bind a queue", zap.Error(err))
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal("Failed to register a consumer", zap.Error(err))
	}
	return msgs
}

type MsgPub func(msg interface{}) error

func MessagePublisher(ch *amqp.Channel, exchange, subtopic string) MsgPub {
	topic := exchange + "." + subtopic
	return func(msg interface{}) error {
		body, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		return ch.Publish(exchange, topic, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	}
}

func CreateChatExchange(uuid string) MsgPub {
	return MessagePublisher(ch, ChatExchange, uuid)
}
