package chats

import (
	"context"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-cc/pkg/chats/proto"
	"github.com/slntopp/nocloud-cc/pkg/graph"
	"go.uber.org/zap"
)

type ChatsServiceServer struct {
	pb.UnimplementedChatServiceServer
	cht_ctrl graph.ChatsController
	msg_ctrl graph.ChatsMessagesController
	log      *zap.Logger
}

func NewChatsServer(log *zap.Logger, db driver.Database) *ChatsServiceServer {
	logger := log.Named("ChatServer")
	chatsController := graph.NewChatsController(logger, db)
	messagesController := graph.NewChatsMessagesController(logger, db)
	return &ChatsServiceServer{
		cht_ctrl: chatsController,
		msg_ctrl: messagesController,
	}
}

func (s *ChatsServiceServer) GetChat(ctx context.Context, req *pb.GetChatRequest) (*pb.Chat, error) {
	chat, err := s.cht_ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}
	return chat.Chat, nil
}

// TODO fix return value
// func (s *ChatsServiceServer) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.Chat, error) {
// 	chat, err := s.cht_ctrl.Create(ctx, req.Chat)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return chat.Chat, nil
// }

// TODO add proto
// func (s *ChatsServiceServer) GetChatMessage(ctx context.Context, req *pb.GetChatMessageRequest) (*pb.Chat, error) {
// 	chat, err := s.ctrl.Get(ctx, req.GetUuid())
// 	if err != nil {
// 		return nil, err
// 	}
// 	return chat.Chat, nil
// }
