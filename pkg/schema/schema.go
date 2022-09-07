package schema

import "github.com/slntopp/nocloud/pkg/nocloud/schema"

const (
	CHATS_COL          = "Chats"
	CHATS_MESSAGES_COL = "ChatsMessages"
	NS2CHTS            = schema.NAMESPACES_COL + "2" + CHATS_COL

	// Provide read access to all group members
	CHT2MSG = CHATS_MESSAGES_COL + "2" + CHATS_COL
	// Provide edit access to own messages
	NS2MSG = schema.NAMESPACES_COL + "2" + CHATS_COL
)
