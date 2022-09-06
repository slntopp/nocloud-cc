package schema

import "github.com/slntopp/nocloud/pkg/nocloud/schema"

const (
	CHATS_COL          = "Chats"
	NS2CHTS            = schema.NAMESPACES_COL + "2" + CHATS_COL
	CHATS_MESSAGES_COL = "ChatsMessages"
	CHT2MSG            = CHATS_MESSAGES_COL + "2" + CHATS_COL
)
