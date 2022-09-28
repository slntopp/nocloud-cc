package chats

import (
	"github.com/slntopp/nocloud-cc/pkg/broker"
)

var queues map[string]broker.MsgPub = make(map[string]broker.MsgPub)

func GetChatPub(uuid string) broker.MsgPub {
	if _, ok := queues[uuid]; !ok {
		queues[uuid] = broker.CreateChatExchange(uuid)
	}

	return queues[uuid]
}
