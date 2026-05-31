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

type PromptEntity struct {
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
}

type SlackPromptEnvelope struct {
	ChannelID         string         `json:"channel_id,omitempty"`
	ChannelName       string         `json:"channel_name,omitempty"`
	ThreadTS          string         `json:"thread_ts,omitempty"`
	SenderUserID      string         `json:"sender_user_id,omitempty"`
	SenderDisplayName string         `json:"sender_display_name,omitempty"`
	RawText           string         `json:"raw_text,omitempty"`
	RenderedText      string         `json:"rendered_text,omitempty"`
	MentionedChannels []PromptEntity `json:"mentioned_channels,omitempty"`
	MentionedUsers    []PromptEntity `json:"mentioned_users,omitempty"`
	Permalink         string         `json:"permalink,omitempty"`
}

const (
	BotOrchestrator BotRole = "orchestrator"
	BotOnCall       BotRole = "oncall"
	BotFR           BotRole = "fr"
	BotArch         BotRole = "arch"
)

type SlackEnvelope struct {
	BotRole     BotRole             `json:"bot_role"`
	TeamID      string              `json:"team_id"`
	ChannelID   string              `json:"channel_id"`
	ThreadTS    string              `json:"thread_ts"`
	ActionToken string              `json:"action_token,omitempty"`
	UserID      string              `json:"user_id"`
	Text        string              `json:"text"`
	TS          string              `json:"ts"`
	Files       []string            `json:"files"`
	EntityRefs  []EntityRef         `json:"entity_refs,omitempty"`
	Prompt      SlackPromptEnvelope `json:"prompt_envelope,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
}

type Ingestion struct {
	ID             string              `json:"id"`
	EventID        string              `json:"event_id,omitempty"`
	ConversationID string              `json:"conversation_id,omitempty"`
	CaseID         string              `json:"case_id,omitempty"`
	ThreadKey      string              `json:"thread_key"`
	ThreadTS       string              `json:"thread_ts,omitempty"`
	WorkflowHint   string              `json:"workflow_hint"`
	Intent         string              `json:"intent,omitempty"`
	BotRole        BotRole             `json:"bot_role,omitempty"`
	Source         string              `json:"source"`
	ChannelID      string              `json:"channel_id"`
	UserID         string              `json:"user_id"`
	Text           string              `json:"text"`
	EntityRefs     []EntityRef         `json:"entity_refs,omitempty"`
	Prompt         SlackPromptEnvelope `json:"prompt_envelope,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
}
