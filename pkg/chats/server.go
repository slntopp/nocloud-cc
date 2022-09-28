package chats

import (
	"context"
	"encoding/json"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-cc/pkg/broker"
	pb "github.com/slntopp/nocloud-cc/pkg/chats/proto"
	"github.com/slntopp/nocloud-cc/pkg/graph"
	"github.com/slntopp/nocloud-cc/pkg/schema"
	"github.com/slntopp/nocloud/pkg/nocloud/access"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatsServiceServer struct {
	pb.UnimplementedChatServiceServer
	db       driver.Database
	cht_ctrl graph.ChatsController
	msg_ctrl graph.ChatsMessagesController
	log      *zap.Logger
}

func NewChatsServer(log *zap.Logger, db driver.Database) *ChatsServiceServer {
	logger := log.Named("ChatServer")
	chatsController := graph.NewChatsController(logger, db)
	messagesController := graph.NewChatsMessagesController(logger, db)
	return &ChatsServiceServer{
		db:       db,
		log:      logger,
		cht_ctrl: chatsController,
		msg_ctrl: messagesController,
	}
}

func (s *ChatsServiceServer) GetChat(ctx context.Context, req *pb.GetChatRequest) (*pb.Chat, error) {
	s.log.Info("Got GetChat Request", zap.Any("request", req))
	chat, err := s.cht_ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}
	return chat.Chat, nil
}

func (s *ChatsServiceServer) DeleteChat(ctx context.Context, req *pb.DeleteChatRequest) (*pb.Response, error) {
	s.log.Info("Got DeleteChat Request", zap.Any("request", req))
	err := s.cht_ctrl.Delete(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}
	return &pb.Response{}, nil
}

func (s *ChatsServiceServer) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.Chat, error) {
	s.log.Info("Got CreateChat Request", zap.Any("request", req))
	chat, err := s.cht_ctrl.Create(ctx, req.Chat)
	if err != nil {
		return nil, err
	}
	return chat.Chat, nil
}

func (s *ChatsServiceServer) Update(ctx context.Context, chat *pb.Chat) (*pb.Chat, error) {
	s.log.Info("Got UpdateChat Request", zap.Any("request", chat))
	err := s.cht_ctrl.Update(ctx, chat)
	if err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *ChatsServiceServer) SendChatMessage(ctx context.Context, req *pb.SendChatMessageRequest) (*pb.ChatMessage, error) {
	s.log.Info("Got SendChatMessage Request", zap.Any("request", req))
	msg, err := s.msg_ctrl.Create(ctx, req.GetMessage(), req.GetEntities())
	if err != nil {
		return nil, err
	}
	if err := GetChatPub(msg.To)(msg.ChatMessage); err != nil {
		s.log.Warn("Error while publishing message", zap.Error(err))
	}

	return msg.ChatMessage, nil
}

func (s *ChatsServiceServer) GetChatMessage(ctx context.Context, req *pb.GetChatMessageRequest) (*pb.ChatMessage, error) {
	s.log.Info("Got GetChatMessage Request", zap.Any("request", req))
	chat, err := s.msg_ctrl.Get(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}
	return chat.ChatMessage, nil
}

func (s *ChatsServiceServer) DeleteChatMessage(ctx context.Context, req *pb.DeleteChatMessageRequest) (*pb.Response, error) {
	s.log.Info("Got DeleteChatMessage Request", zap.Any("request", req))
	err := s.msg_ctrl.Delete(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}
	return &pb.Response{}, nil
}

func (s *ChatsServiceServer) UpdateChatMessage(ctx context.Context, msg *pb.ChatMessage) (*pb.ChatMessage, error) {
	s.log.Info("Got UpdateChatMessage Request", zap.Any("request", msg))
	err := s.msg_ctrl.Update(ctx, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *ChatsServiceServer) Invite(ctx context.Context, req *pb.InviteChatRequest) (*pb.Response, error) {
	s.log.Info("Got Invite Request", zap.Any("request", req))
	err := s.cht_ctrl.InviteUser(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.Response{}, nil
}
func (s *ChatsServiceServer) ListChatMessages(ctx context.Context, req *pb.ListChatMessagesRequest) (*pb.ListChatMessagesResponse, error) {
	s.log.Info("Got List Messages Request", zap.Any("request", req))
	messages, err := s.msg_ctrl.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.ListChatMessagesResponse{Messages: messages}, nil
}

func (s *ChatsServiceServer) Stream(req *pb.ChatMessageStreamRequest, stream pb.ChatService_StreamServer) error {
	s.log.Info("Got ChatMessageStream Request", zap.Any("request", req))
	uuid := req.GetUuid()

	if !graph.HasAccess(stream.Context(), s.db, schema.ACC2CHTS, uuid, access.READ) {
		s.log.Warn("Access check failed", zap.Any("context", stream.Context()))
		return status.Error(codes.PermissionDenied, "Not enough access to subscribe to chat")
	}

	msgs, err := broker.GetConsumer(stream.Context(), uuid)
	if err != nil {
		return err
	}

	for msg := range msgs {
		s.log.Info("Unmarshaling incoming message")
		chatMessage := &pb.ChatMessage{}
		json.Unmarshal(msg.Body, chatMessage)
		stream.Send(chatMessage)
	}

	return nil
}
