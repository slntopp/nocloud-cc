package graph

import (
	"context"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-cc/pkg/chats/proto"
	chatpb "github.com/slntopp/nocloud-cc/pkg/chats/proto"
	"github.com/slntopp/nocloud-cc/pkg/schema"
	nograph "github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/access"
	noschema "github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Chat struct {
	*chatpb.Chat
	driver.DocumentMeta
}
type ChatsController struct {
	log      *zap.Logger
	db       driver.Database
	col      driver.Collection
	graph    driver.Graph
	acc2chts driver.Collection
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
	acc2msg driver.Collection
	graph   driver.Graph
}

func NewChatsController(logger *zap.Logger, db driver.Database) ChatsController {
	ctx := context.TODO()
	log := logger.Named("ChatsController")
	log.Info("Creating ChatsController")

	graph := nograph.GraphGetEnsure(log, ctx, db, noschema.PERMISSIONS_GRAPH.Name)
	col := nograph.GraphGetVertexEnsure(log, ctx, db, graph, schema.CHATS_COL)

	nograph.GraphGetEdgeEnsure(log, ctx, graph, schema.ACC2CHTS, noschema.ACCOUNTS_COL, schema.CHATS_COL)

	acc2chts := nograph.GraphGetEdgeEnsure(log, ctx, graph, schema.ACC2CHTS, noschema.ACCOUNTS_COL, schema.CHATS_COL)

	return ChatsController{log: log, col: col, graph: graph, acc2chts: acc2chts, db: db}
}

func NewChatsMessagesController(logger *zap.Logger, db driver.Database) ChatsMessagesController {
	ctx := context.TODO()
	log := logger.Named("ChatsMessagesController")
	log.Info("Creating ChatsMessagesController")

	graph := nograph.GraphGetEnsure(log, ctx, db, noschema.PERMISSIONS_GRAPH.Name)
	col := nograph.GraphGetVertexEnsure(log, ctx, db, graph, schema.CHATS_MESSAGES_COL)

	cht2msg := nograph.GraphGetEdgeEnsure(log, ctx, graph, schema.CHT2MSG, schema.CHATS_COL, schema.CHATS_MESSAGES_COL)
	acc2msg := nograph.GraphGetEdgeEnsure(log, ctx, graph, schema.ACC2MSG, noschema.ACCOUNTS_COL, schema.CHATS_MESSAGES_COL)

	return ChatsMessagesController{log: log, col: col, cht2msg: cht2msg, graph: graph, acc2msg: acc2msg, db: db}
}

// Get Chat by id from the database
func (ctrl *ChatsController) Get(ctx context.Context, id string) (*Chat, error) {
	logger := ctrl.log.Named("GetChat")
	logger.Info("Getting chat", zap.String("id", id))
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	chat := &proto.Chat{}
	meta, err := ctrl.col.ReadDocument(ctx, id, chat)
	if err != nil {
		return nil, err
	}
	chat.Uuid = meta.ID.Key()

	if !HasAccess(ctx, ctrl.db, schema.ACC2CHTS, requestor, id, access.READ) {
		return nil, status.Error(codes.PermissionDenied, "Permission Denied")
	}

	return &Chat{chat, meta}, nil

}

func (ctrl *ChatsController) Delete(ctx context.Context, id string) error {
	logger := ctrl.log.Named("DeleteChat")
	logger.Info("Deleting chat", zap.String("id", id))
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if !HasAccess(ctx, ctrl.db, schema.ACC2CHTS, requestor, id, access.MGMT) {
		return status.Error(codes.PermissionDenied, "Permission Denied")
	}

	_, err := ctrl.col.RemoveDocument(ctx, id)

	return err
}

func (ctrl *ChatsController) Create(ctx context.Context, chat *chatpb.Chat) (*Chat, error) {
	logger := ctrl.log.Named("CreatingChat")
	logger.Info("Creating chat", zap.String("id", chat.GetUuid()), zap.Any("chat", chat))
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	chat.Owner = requestor

	meta, err := ctrl.col.CreateDocument(ctx, chat)
	if err != nil {
		return nil, err
	}

	chat.Uuid = meta.ID.Key()

	_, err = ctrl.acc2chts.CreateDocument(ctx, nograph.Access{
		From:  driver.NewDocumentID(noschema.ACCOUNTS_COL, requestor),
		To:    driver.NewDocumentID(schema.CHATS_COL, chat.Uuid),
		Level: access.MGMT,
	})
	if err != nil {
		logger.Warn("Could not link account and chat", zap.String("chat", chat.Uuid), zap.String("account", requestor), zap.Error(err))
	}

	return &Chat{chat, meta}, nil
}

func (ctrl *ChatsController) Update(ctx context.Context, chat *chatpb.Chat) error {
	logger := ctrl.log.Named("UpdateChat")
	logger.Info("Updating chat", zap.String("id", chat.GetUuid()), zap.Any("chat", chat))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if !HasAccess(ctx, ctrl.db, schema.ACC2CHTS, requestor, chat.GetUuid(), access.MGMT) {
		return status.Error(codes.PermissionDenied, "Permission Denied")
	}

	_, err := ctrl.col.ReplaceDocument(ctx, chat.GetUuid(), chat)
	return err
}

func (ctrl *ChatsController) InviteUser(ctx context.Context, invite *chatpb.InviteChatRequest) error {
	logger := ctrl.log.Named("InviteUser")
	logger.Info("Inviting user to chat", zap.String("chat", invite.GetChatUuid()), zap.String("user", invite.GetUserUuid()))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if !HasAccess(ctx, ctrl.db, schema.ACC2CHTS,
		requestor, invite.GetChatUuid(), access.READ) {
		return status.Error(codes.PermissionDenied, "Permission Denied")
	}

	_, err := ctrl.acc2chts.CreateDocument(ctx, nograph.Access{
		From:  driver.NewDocumentID(noschema.ACCOUNTS_COL, invite.GetUserUuid()),
		To:    driver.NewDocumentID(schema.CHATS_COL, invite.GetChatUuid()),
		Level: access.READ,
	})

	return err
}

func (ctrl *ChatsMessagesController) Create(ctx context.Context, msg *chatpb.ChatMessage) (*ChatMessage, error) {
	logger := ctrl.log.Named("CreateChatMessage")
	logger.Info("Creating message", zap.Any("message", msg))
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	msg.From = requestor

	if !HasAccess(ctx, ctrl.db, schema.ACC2CHTS,
		requestor, msg.GetTo(), access.READ) {
		return nil, status.Error(codes.PermissionDenied, "Permission Denied")
	}

	meta, err := ctrl.col.CreateDocument(ctx, msg)
	if err != nil {
		return nil, err
	}
	msg.Uuid = meta.ID.Key()

	_, err = ctrl.cht2msg.CreateDocument(ctx, nograph.Access{
		From:  driver.NewDocumentID(schema.CHATS_COL, msg.GetTo()),
		To:    driver.NewDocumentID(schema.CHATS_MESSAGES_COL, msg.Uuid),
		Level: access.READ,
	})
	if err != nil {
		logger.Warn("Could not link chat and message", zap.String("chat", msg.To), zap.String("message", msg.Uuid))
	}

	_, err = ctrl.acc2msg.CreateDocument(ctx, nograph.Access{
		From:  driver.NewDocumentID(noschema.ACCOUNTS_COL, requestor),
		To:    driver.NewDocumentID(schema.CHATS_MESSAGES_COL, msg.Uuid),
		Level: access.MGMT,
	})
	if err != nil {
		logger.Warn("Could not link account and message", zap.String("account", requestor), zap.String("message", msg.Uuid))
	}

	return &ChatMessage{msg, meta}, nil
}

func (ctrl *ChatsMessagesController) Get(ctx context.Context, id string) (*ChatMessage, error) {
	logger := ctrl.log.Named("GetChatMessage")
	logger.Info("Getting chat message", zap.String("id", id))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if !HasAccess(ctx, ctrl.db, schema.ACC2MSG,
		requestor, id, access.READ) {
		return nil, status.Error(codes.PermissionDenied, "Permission Denied")
	}
	msg := &chatpb.ChatMessage{}
	meta, err := ctrl.col.ReadDocument(ctx, id, msg)
	if err != nil {
		return nil, err
	}

	msg.Uuid = meta.ID.Key()
	return &ChatMessage{msg, meta}, nil
}

func (ctrl *ChatsMessagesController) Delete(ctx context.Context, id string) error {
	logger := ctrl.log.Named("DeleteChatMessage")
	logger.Info("Deleting message", zap.String("id", id))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if !HasAccess(ctx, ctrl.db, schema.ACC2MSG,
		requestor, id, access.MGMT) {
		return status.Error(codes.PermissionDenied, "Permission Denied")
	}
	_, err := ctrl.col.RemoveDocument(ctx, id)
	return err
}

func (ctrl *ChatsMessagesController) Update(ctx context.Context, msg *chatpb.ChatMessage) error {
	logger := ctrl.log.Named("UpdateChatMessage")
	logger.Info("Updating message", zap.String("id", msg.GetUuid()), zap.Any("message", msg))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if !HasAccess(ctx, ctrl.db, schema.ACC2MSG,
		requestor, msg.GetUuid(), access.MGMT) {
		return status.Error(codes.PermissionDenied, "Permission Denied")
	}

	_, err := ctrl.col.ReplaceDocument(ctx, msg.GetUuid(), msg)
	return err
}

var listQuery = `
FOR message IN @@collection 
    FILTER message.to == @chat 
    RETURN message`

func (ctrl *ChatsMessagesController) List(ctx context.Context, req *chatpb.ListChatMessagesRequest) ([]*chatpb.ChatMessage, error) {
	logger := ctrl.log.Named("ListChatMessages")
	logger.Info("Fetching messages", zap.String("chat", req.GetChatUuid()))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	if !HasAccess(ctx, ctrl.db, schema.ACC2CHTS,
		requestor, req.GetChatUuid(), access.READ) {
		return nil, status.Error(codes.PermissionDenied, "Permission Denied")
	}

	c, err := ctrl.db.Query(ctx, listQuery, map[string]interface{}{
		"chat":        req.GetChatUuid(),
		"@collection": schema.CHATS_MESSAGES_COL,
	})
	if err != nil {
		return nil, err
	}

	messages := []*proto.ChatMessage{}
	for {
		message := &proto.ChatMessage{}
		_, err = c.ReadDocument(ctx, message)

		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			ctrl.log.Error("Failed to fetch messages from chat", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to fetch messages from chat")
		} else {
			messages = append(messages, message)
		}
	}

	return messages, nil
}
