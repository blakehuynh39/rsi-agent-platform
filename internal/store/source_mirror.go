package store

import "time"

const (
	SourceMirrorStatusPending  = "pending"
	SourceMirrorStatusComplete = "complete"
	SourceMirrorStatusFailed   = "failed"
	SourceMirrorStatusStale    = "stale"

	SourceMirrorMinimumSchemaVersion int64 = 32
)

type SourceMirrorRecord struct {
	SourceType       string         `json:"source_type"`
	SourceKey        string         `json:"source_key"`
	Workspace        string         `json:"workspace"`
	Environment      string         `json:"environment"`
	SourceSessionKey string         `json:"source_session_key"`
	HonchoWorkspace  string         `json:"honcho_workspace"`
	HonchoSessionID  string         `json:"honcho_session_id"`
	HonchoMessageID  string         `json:"honcho_message_id,omitempty"`
	HonchoObjectType string         `json:"honcho_object_type,omitempty"`
	HonchoObjectID   string         `json:"honcho_object_id,omitempty"`
	SourceRevision   string         `json:"source_revision"`
	Status           string         `json:"status"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	LastError        string         `json:"last_error,omitempty"`
	CreatedAt        time.Time      `json:"created_at,omitempty"`
	UpdatedAt        time.Time      `json:"updated_at,omitempty"`
}

type SourceMirrorClaimResult struct {
	Record      SourceMirrorRecord `json:"record"`
	ShouldWrite bool               `json:"should_write"`
	Reason      string             `json:"reason"`
}

type SourceMirrorWriteStore interface {
	ClaimSourceMirrorRecord(record SourceMirrorRecord, lease time.Duration) (SourceMirrorClaimResult, error)
	CompleteSourceMirrorRecord(sourceType string, sourceKey string, honchoMessageID string, metadata map[string]any) (SourceMirrorRecord, error)
	CompleteSourceMirrorObject(sourceType string, sourceKey string, honchoObjectType string, honchoObjectID string, metadata map[string]any) (SourceMirrorRecord, error)
	FailSourceMirrorRecord(sourceType string, sourceKey string, lastError string, metadata map[string]any) (SourceMirrorRecord, error)
	MarkSourceMirrorRecordStale(record SourceMirrorRecord, lastError string, metadata map[string]any) (SourceMirrorRecord, error)
	GetSourceMirrorRecord(sourceType string, sourceKey string) (SourceMirrorRecord, bool, error)
}

type SourceMirrorStatusStore interface {
	ListSourceMirrorRecords(sourceTypes []string, limit int) ([]SourceMirrorRecord, error)
}
