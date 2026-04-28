package store

import (
	"database/sql"
	"log"
	"sort"

	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
)

func (p *PostgresStore) ListConversationEntriesPage(conversationID string, opts ConversationEntryPageOptions) ConversationEntryPage {
	limit := opts.Limit
	if limit <= 0 {
		return ConversationEntryPage{Entries: p.ListConversationEntries(conversationID), Limit: limit}
	}
	rows, err := p.db.Query(`select id, conversation_id, event_id, trace_id, source, source_event_id, entry_type, actor_id, actor_type, body, metadata, created_at
from conversation_entry
where conversation_id = $1
order by created_at asc, id asc
limit $2`, conversationID, limit+1)
	if err != nil {
		log.Printf("store postgres list conversation entries page fallback conversation_id=%s error=%v", conversationID, err)
		entries, hasMore := pageConversationEntries(p.ListConversationEntries(conversationID), limit)
		return ConversationEntryPage{Entries: entries, Limit: limit, HasMore: hasMore}
	}
	defer rows.Close()
	entries := make([]conversation.Entry, 0, limit+1)
	for rows.Next() {
		item, scanErr := scanConversationEntry(rows)
		if scanErr != nil {
			log.Printf("store postgres scan conversation entry page fallback conversation_id=%s error=%v", conversationID, scanErr)
			entries, hasMore := pageConversationEntries(p.ListConversationEntries(conversationID), limit)
			return ConversationEntryPage{Entries: entries, Limit: limit, HasMore: hasMore}
		}
		entries = append(entries, item)
	}
	if err := rows.Err(); err != nil {
		log.Printf("store postgres read conversation entry page fallback conversation_id=%s error=%v", conversationID, err)
		entries, hasMore := pageConversationEntries(p.ListConversationEntries(conversationID), limit)
		return ConversationEntryPage{Entries: entries, Limit: limit, HasMore: hasMore}
	}
	hasMore := len(entries) > limit
	if hasMore {
		entries = entries[:limit]
	}
	return ConversationEntryPage{Entries: entries, Limit: limit, HasMore: hasMore}
}

type conversationEntryScanner interface {
	Scan(dest ...any) error
}

func scanConversationEntry(row conversationEntryScanner) (conversation.Entry, error) {
	var item conversation.Entry
	var eventID, traceID, actorID, actorType sql.NullString
	var source string
	var metadata []byte
	if err := row.Scan(&item.ID, &item.ConversationID, &eventID, &traceID, &source, &item.SourceEventID, &item.EntryType, &actorID, &actorType, &item.Body, &metadata, &item.CreatedAt); err != nil {
		return conversation.Entry{}, err
	}
	item.EventID = eventID.String
	item.TraceID = traceID.String
	item.Source = ingestion.Source(source)
	item.ActorID = actorID.String
	item.ActorType = actorType.String
	item.Metadata = decodeJSON(metadata, map[string]interface{}{})
	return item, nil
}

func pageConversationEntries(entries []conversation.Entry, limit int) ([]conversation.Entry, bool) {
	sort.Slice(entries, func(i, j int) bool { return entries[i].CreatedAt.Before(entries[j].CreatedAt) })
	hasMore := limit > 0 && len(entries) > limit
	if hasMore {
		return entries[:limit], true
	}
	return entries, false
}
