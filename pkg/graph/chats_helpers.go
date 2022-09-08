package graph

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-cc/pkg/schema"
	"github.com/slntopp/nocloud/pkg/graph"
)

func hasAccess(ctx context.Context, account string, node string, level int32, db driver.Database, col string) bool {
	key := fmt.Sprintf("%s/0", col)
	return graph.HasAccess(ctx, db, account, key, level)
}

func (ctrl *ChatsController) HasAccess(ctx context.Context, account string, node string, level int32) bool {
	return hasAccess(ctx, account, node, level, ctrl.db, schema.CHATS_COL)
}

func (ctrl *ChatsMessagesController) HasAccess(ctx context.Context, account string, node string, level int32) bool {
	return hasAccess(ctx, account, node, level, ctrl.db, schema.CHATS_MESSAGES_COL)
}
