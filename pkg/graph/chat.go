package graph

import (
	"context"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-cc/pkg/chats/proto"
	chatpb "github.com/slntopp/nocloud-cc/pkg/chats/proto"
	"github.com/slntopp/nocloud-cc/pkg/schema"
	nograph "github.com/slntopp/nocloud/pkg/graph"
	noschema "github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type Chat struct {
	*chatpb.Chat
	driver.DocumentMeta
}
type ChatsController struct {
	log   *zap.Logger
	col   driver.Collection
	graph driver.Graph
}

type ChatMessage struct {
	*chatpb.ChatMessage
	driver.DocumentMeta
}

type ChatsMessagesController struct {
	log     *zap.Logger
	col     driver.Collection
	cht2msg driver.Collection
	graph   driver.Graph
}

func NewChatsController(logger *zap.Logger, db driver.Database) ChatsController {
	ctx := context.TODO()
	log := logger.Named("ChatsController")

	graph := nograph.GraphGetEnsure(log, ctx, db, noschema.PERMISSIONS_GRAPH.Name)
	col := nograph.GraphGetVertexEnsure(log, ctx, db, graph, schema.CHATS_COL)

	nograph.GraphGetEdgeEnsure(log, ctx, graph, schema.NS2CHTS, noschema.NAMESPACES_COL, schema.CHATS_COL)

	return ChatsController{log: log, col: col, graph: graph}
}

func NewChatsMessagesController(logger *zap.Logger, db driver.Database) ChatsMessagesController {
	ctx := context.TODO()
	log := logger.Named("ChatsMessagesController")

	graph := nograph.GraphGetEnsure(log, ctx, db, noschema.PERMISSIONS_GRAPH.Name)
	col := nograph.GraphGetVertexEnsure(log, ctx, db, graph, schema.CHATS_MESSAGES_COL)

	cht2msg := nograph.GraphGetEdgeEnsure(log, ctx, graph, schema.CHT2MSG, schema.CHATS_COL, schema.CHATS_MESSAGES_COL)

	return ChatsMessagesController{log: log, col: col, cht2msg: cht2msg, graph: graph}
}

// Get Chat by id the database
func (ctrl *ChatsController) Get(ctx context.Context, id string) (*Chat, error) {
	logger := ctrl.log.Named("GetChat")
	logger.Info("Getting chat", zap.String("id", id))

	chat := &proto.Chat{}
	meta, err := ctrl.col.ReadDocument(ctx, id, chat)
	if err != nil {
		return nil, err
	}
	chat.Uuid = meta.ID.Key()

	return &Chat{chat, meta}, nil
}

func (ctrl *ChatsController) Create(ctx context.Context, chat *chatpb.Chat) (*Chat, error) {
	meta, err := ctrl.col.CreateDocument(ctx, chat)
	if err != nil {
		return nil, err
	}
	chat.Uuid = meta.ID.Key()
	return &Chat{chat, meta}, nil
}

func (ctrl *ChatsMessagesController) Create(ctx context.Context, msg *chatpb.ChatMessage) (*ChatMessage, error) {
	meta, err := ctrl.col.CreateDocument(ctx, msg)
	if err != nil {
		return nil, err
	}
	msg.Uuid = meta.ID.Key()
	return &ChatMessage{msg, meta}, nil
}

func (ctrl *ChatsMessagesController) Get(ctx context.Context, id string) (ChatMessage, error) {
	logger := ctrl.log.Named("GetChatMessage")
	logger.Info("Getting chat message", zap.String("id", id))

	var msg chatpb.ChatMessage
	meta, err := ctrl.col.ReadDocument(ctx, id, &msg)
	if err != nil {
		return ChatMessage{}, err
	}
	msg.Uuid = meta.ID.Key()

	return ChatMessage{&msg, meta}, nil
}
