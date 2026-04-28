package store

import "github.com/piplabs/rsi-agent-platform/internal/conversation"

type ConversationEntryPageOptions struct {
	Limit int
}

type ConversationEntryPage struct {
	Entries []conversation.Entry
	Limit   int
	HasMore bool
}

type ConversationEntryPager interface {
	ListConversationEntriesPage(conversationID string, opts ConversationEntryPageOptions) ConversationEntryPage
}
