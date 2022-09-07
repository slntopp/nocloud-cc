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
	log     *zap.Logger
	db      driver.Database
	col     driver.Collection
	graph   driver.Graph
	ns2chts driver.Collection
}

type ChatMessage struct {
	*chatpb.ChatMessage
	driver.DocumentMeta
}

type ChatsMessagesController struct {
	log     *zap.Logger
	db      driver.Database
	col     driver.Collection
	cht2msg driver.Collection
	ns2msg  driver.Collection
	graph   driver.Graph
}

func NewChatsController(logger *zap.Logger, db driver.Database) ChatsController {
	ctx := context.TODO()
	log := logger.Named("ChatsController")
	log.Info("Creating ChatsController")

	graph := nograph.GraphGetEnsure(log, ctx, db, noschema.PERMISSIONS_GRAPH.Name)
	col := nograph.GraphGetVertexEnsure(log, ctx, db, graph, schema.CHATS_COL)

	nograph.GraphGetEdgeEnsure(log, ctx, graph, schema.NS2CHTS, noschema.NAMESPACES_COL, schema.CHATS_COL)

	ns2chts := nograph.GraphGetEdgeEnsure(log, ctx, graph, schema.NS2CHTS, noschema.NAMESPACES_COL, schema.CHATS_COL)

	return ChatsController{log: log, col: col, graph: graph, ns2chts: ns2chts, db: db}
}

func NewChatsMessagesController(logger *zap.Logger, db driver.Database) ChatsMessagesController {
	ctx := context.TODO()
	log := logger.Named("ChatsMessagesController")
	log.Info("Creating ChatsMessagesController")

	graph := nograph.GraphGetEnsure(log, ctx, db, noschema.PERMISSIONS_GRAPH.Name)
	col := nograph.GraphGetVertexEnsure(log, ctx, db, graph, schema.CHATS_MESSAGES_COL)

	cht2msg := nograph.GraphGetEdgeEnsure(log, ctx, graph, schema.CHT2MSG, schema.CHATS_COL, schema.CHATS_MESSAGES_COL)
	ns2msg := nograph.GraphGetEdgeEnsure(log, ctx, graph, schema.NS2MSG, noschema.NAMESPACES_COL, schema.CHATS_MESSAGES_COL)

	return ChatsMessagesController{log: log, col: col, cht2msg: cht2msg, graph: graph, ns2msg: ns2msg, db: db}
}

// Get Chat by id from the database
func (ctrl *ChatsController) Get(ctx context.Context, id string) (*Chat, error) {
	logger := ctrl.log.Named("GetChat")
	logger.Info("Getting chat", zap.String("id", id))

	chat := &proto.Chat{}
	meta, err := ctrl.col.ReadDocument(ctx, id, chat)
	if err != nil {
		return nil, err
	}
	chat.Uuid = meta.ID.Key()

	// requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	// ok := graph.HasAccess(ctx, ctrl.db, requestor, ns.ID.String(), access.ADMIN)

	return &Chat{chat, meta}, nil
}

func (ctrl *ChatsController) Delete(ctx context.Context, id string) error {
	logger := ctrl.log.Named("DeleteChat")
	logger.Info("Deleting chat", zap.String("id", id))

	_, err := ctrl.col.RemoveDocument(ctx, id)
	return err
}

func (ctrl *ChatsController) Create(ctx context.Context, chat *chatpb.Chat) (*Chat, error) {
	logger := ctrl.log.Named("CreatingChat")
	logger.Info("Creating chat", zap.String("id", chat.GetUuid()), zap.Any("chat", chat))

	meta, err := ctrl.col.CreateDocument(ctx, chat)
	if err != nil {
		return nil, err
	}
	// TODO create chat namespace

	chat.Uuid = meta.ID.Key()

	_, err = ctrl.ns2chts.CreateDocument(ctx, nograph.Access{
		From: driver.DocumentID(chat.Owner), To: driver.DocumentID(chat.Uuid),
		// Edit access for own messages
		Level: 3,
	})
	if err != nil {
		logger.Warn("Could not link namespace and chat", zap.String("chat", chat.Uuid), zap.String("namespace", chat.Owner))
	}

	return &Chat{chat, meta}, nil
}

func (ctrl *ChatsController) Update(ctx context.Context, chat *chatpb.Chat) error {
	logger := ctrl.log.Named("UpdateChat")
	logger.Info("Updating chat", zap.String("id", chat.GetUuid()), zap.Any("chat", chat))

	_, err := ctrl.col.ReplaceDocument(ctx, chat.GetUuid(), chat)

	return err
}

func (ctrl *ChatsMessagesController) Create(ctx context.Context, msg *chatpb.ChatMessage) (*ChatMessage, error) {
	logger := ctrl.log.Named("CreateChatMessage")
	logger.Info("Creating message", zap.Any("message", msg))

	meta, err := ctrl.col.CreateDocument(ctx, msg)
	if err != nil {
		return nil, err
	}
	msg.Uuid = meta.ID.Key()

	// TODO create personal namespace

	_, err = ctrl.cht2msg.CreateDocument(ctx, nograph.Access{
		From: driver.DocumentID(msg.To), To: driver.DocumentID(msg.Uuid),
		// Read only access for all chat members
		Level: 1,
	})
	if err != nil {
		logger.Warn("Could not link chat and message", zap.String("chat", msg.To), zap.String("message", msg.Uuid))
	}

	_, err = ctrl.ns2msg.CreateDocument(ctx, nograph.Access{
		From: driver.DocumentID(msg.From), To: driver.DocumentID(msg.Uuid),
		// Edit access for own messages
		Level: 3,
	})
	if err != nil {
		logger.Warn("Could not link namespace and message", zap.String("chat", msg.From), zap.String("message", msg.Uuid))
	}

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

func (ctrl *ChatsMessagesController) Delete(ctx context.Context, id string) error {
	logger := ctrl.log.Named("DeleteChatMessage")
	logger.Info("Deleting message", zap.String("id", id))

	_, err := ctrl.col.RemoveDocument(ctx, id)
	return err
}

func (ctrl *ChatsMessagesController) Update(ctx context.Context, msg *chatpb.ChatMessage) error {
	logger := ctrl.log.Named("UpdateChatMessage")
	logger.Info("Updating message", zap.String("id", msg.GetUuid()), zap.Any("message", msg))

	_, err := ctrl.col.ReplaceDocument(ctx, msg.GetUuid(), msg)

	return err
}
