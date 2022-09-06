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

func (s *ChatsServiceServer) DeleteChat(ctx context.Context, req *pb.DeleteChatRequest) (*pb.Response, error) {
	err := s.cht_ctrl.Delete(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}
	return &pb.Response{}, nil
}

func (s *ChatsServiceServer) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.Chat, error) {
	chat, err := s.cht_ctrl.Create(ctx, req.Chat)
	if err != nil {
		return nil, err
	}
	return chat.Chat, nil
}

func (s *ChatsServiceServer) SendChatMessage(ctx context.Context, req *pb.SendChatMessageRequest) (*pb.ChatMessage, error) {
	chat, err := s.msg_ctrl.Create(ctx, req.Message)
	if err != nil {
		return nil, err
	}
	return chat.ChatMessage, nil
}

func (s *ChatsServiceServer) GetChatMessage(ctx context.Context, req *pb.GetChatMessageRequest) (*pb.ChatMessage, error) {
	chat, err := s.msg_ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}
	return chat.ChatMessage, nil
}

func (s *ChatsServiceServer) DeleteChatMessage(ctx context.Context, req *pb.DeleteChatMessageRequest) (*pb.Response, error) {
	err := s.msg_ctrl.Delete(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}
	return &pb.Response{}, nil
}

func (s *ChatsServiceServer) UpdateChatMessage(ctx context.Context, msg *pb.ChatMessage) (*pb.ChatMessage, error) {
	err := s.msg_ctrl.Update(ctx, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
