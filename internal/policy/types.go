package policy

import "time"

type ThreadState string

const (
	ThreadStateActive          ThreadState = "ACTIVE"
	ThreadStateMuted           ThreadState = "MUTED"
	ThreadStateMuteUntilMention ThreadState = "MUTE_UNTIL_MENTION"
	ThreadStateClosed          ThreadState = "CLOSED"
	ThreadStateObserveOnly     ThreadState = "OBSERVE_ONLY"
)

type ThreadPolicy struct {
	ThreadKey         string      `json:"thread_key"`
	State             ThreadState `json:"state"`
	OwnerBot          string      `json:"owner_bot"`
	Muted             bool        `json:"muted"`
	CloseReason       string      `json:"close_reason,omitempty"`
	LastPolicyVersion string      `json:"last_policy_version"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

type ChannelPolicy struct {
	ChannelID            string    `json:"channel_id"`
	ProactiveEnabled     bool      `json:"proactive_enabled"`
	AutoPostAllowed      bool      `json:"auto_post_allowed"`
	AllowedWorkflowKinds []string  `json:"allowed_workflow_kinds"`
	UpdatedAt            time.Time `json:"updated_at"`
}

