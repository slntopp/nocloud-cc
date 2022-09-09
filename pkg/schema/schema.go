package schema

import "github.com/slntopp/nocloud/pkg/nocloud/schema"

const (
	CHATS_COL          = "Chats"
	CHATS_MESSAGES_COL = "ChatsMessages"
	ACC2CHTS           = schema.ACCOUNTS_COL + "2" + CHATS_COL
	ACC2MSG            = schema.ACCOUNTS_COL + "2" + CHATS_MESSAGES_COL
)
