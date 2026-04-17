package store

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
)

func externalConversationEntry(conversationID string, event ingestion.EventEnvelope, createdAt time.Time) conversation.Entry {
	return conversation.Entry{
		ID:             conversationEntryID("external_event", conversationID, event.ID, firstNonEmpty(event.SourceEventID, event.DedupeKey)),
		ConversationID: conversationID,
		EventID:        event.ID,
		Source:         event.Source,
		SourceEventID:  event.SourceEventID,
		EntryType:      "external_event",
		ActorID:        stringFromMetadata(event.Metadata, "user_id"),
		ActorType:      actorTypeForEvent(event),
		Body:           event.NormalizedProblemStatement,
		Metadata:       cloneMetadata(event.Metadata),
		CreatedAt:      createdAt,
	}
}

func slackActionConversationEntry(conversationID string, triggerEventID string, traceID string, action events.SlackActionRecord) conversation.Entry {
	sourceEventID := firstNonEmpty(action.IdempotencyKey, triggerEventID, traceID, "slack-action")
	return conversation.Entry{
		ID:             conversationEntryID("slack_action", conversationID, sourceEventID, firstNonEmpty(traceID, triggerEventID)),
		ConversationID: conversationID,
		EventID:        triggerEventID,
		TraceID:        traceID,
		Source:         ingestion.SourceSlack,
		SourceEventID:  sourceEventID,
		EntryType:      "slack_action",
		ActorID:        action.ChannelID,
		ActorType:      "bot",
		Body:           firstNonEmpty(action.FinalBody, action.DraftBody),
		Metadata: map[string]interface{}{
			"send_status":    action.SendStatus,
			"policy_verdict": action.PolicyVerdict,
			"thread_ts":      action.ThreadTS,
		},
		CreatedAt: action.CreatedAt,
	}
}

func conversationEntryID(kind string, conversationID string, primaryRef string, secondaryRef string) string {
	sum := sha1.Sum([]byte(strings.Join([]string{
		strings.TrimSpace(kind),
		strings.TrimSpace(conversationID),
		strings.TrimSpace(primaryRef),
		strings.TrimSpace(secondaryRef),
	}, "\x1f")))
	return fmt.Sprintf("entry-%x", sum[:10])
}
