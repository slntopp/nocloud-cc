package chats

import (
	"github.com/slntopp/nocloud-cc/pkg/chats/proto"
)

type ChatsServiceServer struct {
	proto.UnimplementedChatServiceServer
}

func NewChatsServer() *ChatsServiceServer {
	return &ChatsServiceServer{}
}
