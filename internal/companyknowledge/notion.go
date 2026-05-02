package companyknowledge

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

const (
	NotionDocumentSourceType = "notion_document"
	NotionMirrorObserverID   = "notion_mirror"
	NotionMirrorObservedID   = "story_company"
)

type NotionDocumentInput struct {
	WorkspaceID    string
	PageID         string
	RootID         string
	ParentID       string
	DatabaseID     string
	Title          string
	URL            string
	LastEditedTime string
	CreatedTime    string
	Content        string
	Hierarchy      []string
	Raw            map[string]any
}

type NotionMirrorOptions struct {
	Environment     string
	HonchoWorkspace string
	Lease           time.Duration
}

type NotionMirrorResult struct {
	SourceKey        string
	SourceSessionKey string
	HonchoWorkspace  string
	HonchoSessionID  string
	HonchoDocumentID string
	SourceRevision   string
	Status           string
	Skipped          bool
	SkipReason       string
}

type NotionMirror struct {
	store  store.SourceMirrorWriteStore
	honcho HonchoDocumentClient
	opts   NotionMirrorOptions
}

type HonchoDocumentClient interface {
	EnsureWorkspace(id string, metadata map[string]any) (clients.HonchoWorkspace, error)
	EnsureSession(workspaceID string, sessionID string, metadata map[string]any) (clients.HonchoSession, error)
	CreateConclusions(workspaceID string, conclusions []clients.HonchoConclusionCreate) ([]clients.HonchoConclusion, error)
}

func NewNotionMirror(state store.SourceMirrorWriteStore, honcho HonchoDocumentClient, opts NotionMirrorOptions) *NotionMirror {
	opts.Environment = strings.TrimSpace(opts.Environment)
	opts.HonchoWorkspace = HonchoCompatibleName("workspace", firstNonEmpty(opts.HonchoWorkspace, "rsi_company_knowledge"))
	if opts.Lease <= 0 {
		opts.Lease = 5 * time.Minute
	}
	return &NotionMirror{store: state, honcho: honcho, opts: opts}
}

func (m *NotionMirror) IngestDocument(ctx context.Context, input NotionDocumentInput) (NotionMirrorResult, error) {
	_ = ctx
	if m == nil || m.store == nil || m.honcho == nil {
		return NotionMirrorResult{}, fmt.Errorf("notion mirror requires store and honcho client")
	}
	if err := validateNotionDocument(input); err != nil {
		return NotionMirrorResult{}, err
	}
	sourceKey := NotionDocumentSourceKey(input.WorkspaceID, input.PageID)
	sessionKey := NotionDocumentSessionKey(input.WorkspaceID, input.PageID)
	revision := NotionDocumentSourceRevision(input)
	honchoSessionID := HonchoCompatibleName("notion", sessionKey)
	metadata := NotionDocumentMetadata(input, sourceKey, sessionKey, revision)
	record := store.SourceMirrorRecord{
		SourceType:       NotionDocumentSourceType,
		SourceKey:        sourceKey,
		Workspace:        strings.TrimSpace(input.WorkspaceID),
		Environment:      strings.TrimSpace(m.opts.Environment),
		SourceSessionKey: sessionKey,
		HonchoWorkspace:  m.opts.HonchoWorkspace,
		HonchoSessionID:  honchoSessionID,
		SourceRevision:   revision,
		Status:           store.SourceMirrorStatusPending,
		Metadata:         metadata,
	}
	claim, err := m.store.ClaimSourceMirrorRecord(record, m.opts.Lease)
	if err != nil {
		return NotionMirrorResult{}, err
	}
	result := NotionMirrorResult{
		SourceKey:        sourceKey,
		SourceSessionKey: sessionKey,
		HonchoWorkspace:  claim.Record.HonchoWorkspace,
		HonchoSessionID:  claim.Record.HonchoSessionID,
		HonchoDocumentID: claim.Record.HonchoObjectID,
		SourceRevision:   claim.Record.SourceRevision,
		Status:           claim.Record.Status,
		Skipped:          !claim.ShouldWrite,
		SkipReason:       claim.Reason,
	}
	if !claim.ShouldWrite {
		return result, nil
	}

	if _, err := m.honcho.EnsureWorkspace(record.HonchoWorkspace, map[string]any{
		"source":      "rsi_company_knowledge",
		"environment": record.Environment,
	}); err != nil {
		_, _ = m.store.FailSourceMirrorRecord(record.SourceType, record.SourceKey, err.Error(), map[string]any{"failure_stage": "ensure_workspace"})
		return NotionMirrorResult{}, err
	}
	if _, err := m.honcho.EnsureSession(record.HonchoWorkspace, record.HonchoSessionID, map[string]any{
		"source":             "notion",
		"source_session_key": record.SourceSessionKey,
		"workspace":          record.Workspace,
		"environment":        record.Environment,
		"source_page_id":     strings.TrimSpace(input.PageID),
		"source_root_id":     strings.TrimSpace(input.RootID),
	}); err != nil {
		_, _ = m.store.FailSourceMirrorRecord(record.SourceType, record.SourceKey, err.Error(), map[string]any{"failure_stage": "ensure_session"})
		return NotionMirrorResult{}, err
	}
	documents, err := m.honcho.CreateConclusions(record.HonchoWorkspace, []clients.HonchoConclusionCreate{
		{
			Content:    NotionDocumentConclusionContent(input),
			ObserverID: NotionMirrorObserverID,
			ObservedID: NotionMirrorObservedID,
			SessionID:  record.HonchoSessionID,
		},
	})
	if err != nil {
		_, _ = m.store.FailSourceMirrorRecord(record.SourceType, record.SourceKey, err.Error(), map[string]any{"failure_stage": "create_document"})
		return NotionMirrorResult{}, err
	}
	if len(documents) != 1 || strings.TrimSpace(documents[0].ID) == "" {
		err := fmt.Errorf("honcho create document returned %d documents with no stable id", len(documents))
		_, _ = m.store.FailSourceMirrorRecord(record.SourceType, record.SourceKey, err.Error(), map[string]any{"failure_stage": "create_document"})
		return NotionMirrorResult{}, err
	}
	completed, err := m.store.CompleteSourceMirrorObject(record.SourceType, record.SourceKey, "document", documents[0].ID, map[string]any{
		"honcho_document_id": documents[0].ID,
		"honcho_api_surface": "conclusions",
	})
	if err != nil {
		return NotionMirrorResult{}, err
	}
	result.HonchoDocumentID = completed.HonchoObjectID
	result.Status = completed.Status
	result.Skipped = false
	result.SkipReason = ""
	return result, nil
}

func validateNotionDocument(input NotionDocumentInput) error {
	if strings.TrimSpace(input.WorkspaceID) == "" {
		return fmt.Errorf("notion workspace id is required")
	}
	if strings.TrimSpace(input.PageID) == "" {
		return fmt.Errorf("notion page id is required")
	}
	if strings.TrimSpace(input.Content) == "" && strings.TrimSpace(input.Title) == "" {
		return fmt.Errorf("notion document content or title is required")
	}
	return nil
}

func NotionDocumentSourceKey(workspaceID string, pageID string) string {
	return "notion_document:" + strings.TrimSpace(workspaceID) + ":" + strings.TrimSpace(pageID)
}

func NotionDocumentSessionKey(workspaceID string, pageID string) string {
	return "notion:" + strings.TrimSpace(workspaceID) + ":" + strings.TrimSpace(pageID)
}

func NotionDocumentSourceRevision(input NotionDocumentInput) string {
	if strings.TrimSpace(input.LastEditedTime) != "" {
		return "last_edited_time:" + strings.TrimSpace(input.LastEditedTime)
	}
	payload := map[string]any{
		"page_id":  strings.TrimSpace(input.PageID),
		"title":    strings.TrimSpace(input.Title),
		"url":      strings.TrimSpace(input.URL),
		"content":  input.Content,
		"children": input.Hierarchy,
	}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return "content:sha256:" + hex.EncodeToString(sum[:])[:24]
}

func NotionDocumentMetadata(input NotionDocumentInput, sourceKey string, sessionKey string, revision string) map[string]any {
	metadata := map[string]any{
		"source":             "notion",
		"source_key":         sourceKey,
		"source_dedupe_key":  sourceKey,
		"source_revision":    revision,
		"source_session_key": sessionKey,
		"workspace_id":       strings.TrimSpace(input.WorkspaceID),
		"notion_page_id":     strings.TrimSpace(input.PageID),
		"notion_root_id":     strings.TrimSpace(input.RootID),
		"notion_parent_id":   strings.TrimSpace(input.ParentID),
		"notion_database_id": strings.TrimSpace(input.DatabaseID),
		"title":              strings.TrimSpace(input.Title),
		"url":                strings.TrimSpace(input.URL),
		"last_edited_time":   strings.TrimSpace(input.LastEditedTime),
		"created_time":       strings.TrimSpace(input.CreatedTime),
		"hierarchy":          input.Hierarchy,
	}
	if len(input.Raw) > 0 {
		metadata["raw_keys"] = sortedKeys(input.Raw)
	}
	return metadata
}

func NotionDocumentConclusionContent(input NotionDocumentInput) string {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = strings.TrimSpace(input.PageID)
	}
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString("Source: Notion\n")
	if strings.TrimSpace(input.URL) != "" {
		b.WriteString("URL: ")
		b.WriteString(strings.TrimSpace(input.URL))
		b.WriteString("\n")
	}
	b.WriteString("Notion page id: ")
	b.WriteString(strings.TrimSpace(input.PageID))
	b.WriteString("\n")
	if strings.TrimSpace(input.LastEditedTime) != "" {
		b.WriteString("Last edited: ")
		b.WriteString(strings.TrimSpace(input.LastEditedTime))
		b.WriteString("\n")
	}
	if len(input.Hierarchy) > 0 {
		b.WriteString("Hierarchy: ")
		b.WriteString(strings.Join(input.Hierarchy, " > "))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	if strings.TrimSpace(input.Content) != "" {
		b.WriteString(strings.TrimSpace(input.Content))
	} else {
		b.WriteString("(No extractable Notion page body text.)")
	}
	return b.String()
}
