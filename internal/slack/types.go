package slack

import "time"

type BotRole string

type EntityKind string

const (
	EntityUser    EntityKind = "user"
	EntityChannel EntityKind = "channel"
)

type EntityRef struct {
	Kind   EntityKind `json:"kind"`
	ID     string     `json:"id"`
	Label  string     `json:"label,omitempty"`
	Source string     `json:"source,omitempty"`
}

const (
	BotOrchestrator BotRole = "orchestrator"
	BotOnCall       BotRole = "oncall"
	BotFR           BotRole = "fr"
	BotArch         BotRole = "arch"
)

type SlackEnvelope struct {
	BotRole     BotRole     `json:"bot_role"`
	TeamID      string      `json:"team_id"`
	ChannelID   string      `json:"channel_id"`
	ThreadTS    string      `json:"thread_ts"`
	ActionToken string      `json:"action_token,omitempty"`
	UserID      string      `json:"user_id"`
	Text        string      `json:"text"`
	TS          string      `json:"ts"`
	Files       []string    `json:"files"`
	EntityRefs  []EntityRef `json:"entity_refs,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

type Ingestion struct {
	ID             string      `json:"id"`
	EventID        string      `json:"event_id,omitempty"`
	ConversationID string      `json:"conversation_id,omitempty"`
	CaseID         string      `json:"case_id,omitempty"`
	ThreadKey      string      `json:"thread_key"`
	ThreadTS       string      `json:"thread_ts,omitempty"`
	WorkflowHint   string      `json:"workflow_hint"`
	Intent         string      `json:"intent,omitempty"`
	BotRole        BotRole     `json:"bot_role,omitempty"`
	Source         string      `json:"source"`
	ChannelID      string      `json:"channel_id"`
	UserID         string      `json:"user_id"`
	Text           string      `json:"text"`
	EntityRefs     []EntityRef `json:"entity_refs,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
}
