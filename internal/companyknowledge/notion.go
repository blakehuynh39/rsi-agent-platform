package companyknowledge

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

const (
	NotionDocumentSourceType  = "notion_document"
	NotionCrawlMissSourceType = "notion_crawl_miss"
	NotionMirrorObserverID    = "notion_mirror"
	NotionMirrorObservedID    = "story_company"

	NotionObjectKindPage     = "page"
	NotionObjectKindDatabase = "database"
	NotionTraversalComplete  = "complete"
	NotionTraversalTruncated = "truncated"
)

type NotionObjectInput struct {
	WorkspaceID        string
	ObjectKind         string
	ObjectID           string
	PageID             string
	RootID             string
	ParentID           string
	DatabaseID         string
	Title              string
	URL                string
	LastEditedTime     string
	CreatedTime        string
	Content            string
	SchemaSummary      string
	SchemaHash         string
	TraversalStatus    string
	Truncated          bool
	Hierarchy          []string
	OutboundReferences []NotionOutboundReference
	Raw                map[string]any
}

type NotionDocumentInput = NotionObjectInput

type NotionOutboundReference struct {
	ReferenceKind    string `json:"reference_kind"`
	SourceBlockID    string `json:"source_block_id,omitempty"`
	SourceProperty   string `json:"source_property,omitempty"`
	TargetID         string `json:"target_id,omitempty"`
	TargetURL        string `json:"target_url,omitempty"`
	TargetObjectKind string `json:"target_object_kind,omitempty"`
	Traversed        bool   `json:"traversed"`
	Reason           string `json:"reason,omitempty"`
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

func (m *NotionMirror) HonchoWorkspace() string {
	if m == nil {
		return ""
	}
	return m.opts.HonchoWorkspace
}

func (m *NotionMirror) IngestDocument(ctx context.Context, input NotionDocumentInput) (NotionMirrorResult, error) {
	_ = ctx
	if m == nil || m.store == nil || m.honcho == nil {
		return NotionMirrorResult{}, fmt.Errorf("notion mirror requires store and honcho client")
	}
	input = normalizeNotionObjectInput(input)
	if err := validateNotionDocument(input); err != nil {
		return NotionMirrorResult{}, err
	}
	sourceKey := NotionObjectSourceKey(input.WorkspaceID, input.ObjectKind, input.ObjectID)
	sessionKey := NotionObjectSessionKey(input.WorkspaceID, input.ObjectKind, input.ObjectID)
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
		"source_object_kind": input.ObjectKind,
		"source_object_id":   input.ObjectID,
		"source_page_id":     strings.TrimSpace(input.PageID),
		"source_database_id": strings.TrimSpace(input.DatabaseID),
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
	switch strings.TrimSpace(input.ObjectKind) {
	case NotionObjectKindPage:
		if strings.TrimSpace(input.PageID) == "" || strings.TrimSpace(input.ObjectID) == "" {
			return fmt.Errorf("notion page id is required")
		}
	case NotionObjectKindDatabase:
		if strings.TrimSpace(input.DatabaseID) == "" || strings.TrimSpace(input.ObjectID) == "" {
			return fmt.Errorf("notion database id is required")
		}
	default:
		return fmt.Errorf("unsupported notion object kind %q", input.ObjectKind)
	}
	if strings.TrimSpace(input.Content) == "" && strings.TrimSpace(input.Title) == "" {
		return fmt.Errorf("notion document content or title is required")
	}
	return nil
}

func normalizeNotionObjectInput(input NotionObjectInput) NotionObjectInput {
	input.WorkspaceID = strings.TrimSpace(input.WorkspaceID)
	input.ObjectKind = strings.TrimSpace(input.ObjectKind)
	input.ObjectID = strings.TrimSpace(input.ObjectID)
	input.PageID = strings.TrimSpace(input.PageID)
	input.DatabaseID = strings.TrimSpace(input.DatabaseID)
	if input.ObjectKind == "" {
		if input.DatabaseID != "" && input.PageID == "" {
			input.ObjectKind = NotionObjectKindDatabase
		} else {
			input.ObjectKind = NotionObjectKindPage
		}
	}
	if input.ObjectKind == NotionObjectKindDatabase {
		if input.DatabaseID == "" {
			input.DatabaseID = input.ObjectID
		}
		if input.ObjectID == "" {
			input.ObjectID = input.DatabaseID
		}
	} else {
		if input.PageID == "" {
			input.PageID = input.ObjectID
		}
		if input.ObjectID == "" {
			input.ObjectID = input.PageID
		}
		input.ObjectKind = NotionObjectKindPage
	}
	if input.TraversalStatus == "" {
		if input.Truncated {
			input.TraversalStatus = NotionTraversalTruncated
		} else {
			input.TraversalStatus = NotionTraversalComplete
		}
	}
	return input
}

func NotionDocumentSourceKey(workspaceID string, pageID string) string {
	return NotionObjectSourceKey(workspaceID, NotionObjectKindPage, pageID)
}

func NotionObjectSourceKey(workspaceID string, objectKind string, objectID string) string {
	workspaceID = strings.TrimSpace(workspaceID)
	objectKind = strings.TrimSpace(objectKind)
	objectID = strings.TrimSpace(objectID)
	if objectKind == NotionObjectKindDatabase {
		return "notion_document:" + workspaceID + ":database:" + objectID
	}
	return "notion_document:" + workspaceID + ":" + objectID
}

func NotionDocumentSessionKey(workspaceID string, pageID string) string {
	return NotionObjectSessionKey(workspaceID, NotionObjectKindPage, pageID)
}

func NotionObjectSessionKey(workspaceID string, objectKind string, objectID string) string {
	workspaceID = strings.TrimSpace(workspaceID)
	objectKind = strings.TrimSpace(objectKind)
	objectID = strings.TrimSpace(objectID)
	if objectKind == NotionObjectKindDatabase {
		return "notion:" + workspaceID + ":database:" + objectID
	}
	return "notion:" + workspaceID + ":" + objectID
}

func NotionCrawlMissSourceKey(workspaceID string, rootID string, targetID string) string {
	return "notion_crawl_miss:" + strings.TrimSpace(workspaceID) + ":" + strings.TrimSpace(rootID) + ":" + strings.TrimSpace(targetID)
}

func NotionDocumentSourceRevision(input NotionDocumentInput) string {
	input = normalizeNotionObjectInput(input)
	if input.ObjectKind == NotionObjectKindDatabase {
		parts := []string{}
		if strings.TrimSpace(input.LastEditedTime) != "" {
			parts = append(parts, "last_edited_time:"+strings.TrimSpace(input.LastEditedTime))
		}
		if strings.TrimSpace(input.SchemaHash) != "" {
			parts = append(parts, "schema_hash:"+strings.TrimSpace(input.SchemaHash))
		}
		if len(parts) > 0 {
			return strings.Join(parts, ";")
		}
	}
	if strings.TrimSpace(input.LastEditedTime) != "" {
		return "last_edited_time:" + strings.TrimSpace(input.LastEditedTime)
	}
	payload := map[string]any{
		"object_kind": input.ObjectKind,
		"object_id":   input.ObjectID,
		"title":       strings.TrimSpace(input.Title),
		"url":         strings.TrimSpace(input.URL),
		"content":     input.Content,
		"children":    input.Hierarchy,
		"schema_hash": strings.TrimSpace(input.SchemaHash),
	}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return "content:sha256:" + hex.EncodeToString(sum[:])[:24]
}

func NotionDocumentMetadata(input NotionDocumentInput, sourceKey string, sessionKey string, revision string) map[string]any {
	input = normalizeNotionObjectInput(input)
	metadata := map[string]any{
		"source":             "notion",
		"source_key":         sourceKey,
		"source_dedupe_key":  sourceKey,
		"source_revision":    revision,
		"source_session_key": sessionKey,
		"workspace_id":       strings.TrimSpace(input.WorkspaceID),
		"object_kind":        input.ObjectKind,
		"object_id":          input.ObjectID,
		"notion_page_id":     strings.TrimSpace(input.PageID),
		"notion_root_id":     strings.TrimSpace(input.RootID),
		"notion_parent_id":   strings.TrimSpace(input.ParentID),
		"notion_database_id": strings.TrimSpace(input.DatabaseID),
		"title":              strings.TrimSpace(input.Title),
		"url":                strings.TrimSpace(input.URL),
		"last_edited_time":   strings.TrimSpace(input.LastEditedTime),
		"created_time":       strings.TrimSpace(input.CreatedTime),
		"hierarchy":          input.Hierarchy,
		"traversal_status":   strings.TrimSpace(input.TraversalStatus),
		"truncated":          input.Truncated,
	}
	if strings.TrimSpace(input.SchemaHash) != "" {
		metadata["schema_hash"] = strings.TrimSpace(input.SchemaHash)
	}
	if strings.TrimSpace(input.SchemaSummary) != "" {
		metadata["schema_summary"] = strings.TrimSpace(input.SchemaSummary)
	}
	if len(input.OutboundReferences) > 0 {
		metadata["outbound_references"] = input.OutboundReferences
	}
	if len(input.Raw) > 0 {
		metadata["raw_keys"] = sortedKeys(input.Raw)
	}
	return metadata
}

func NotionDocumentConclusionContent(input NotionDocumentInput) string {
	input = normalizeNotionObjectInput(input)
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = strings.TrimSpace(input.ObjectID)
	}
	sourceKey := NotionObjectSourceKey(input.WorkspaceID, input.ObjectKind, input.ObjectID)
	sessionKey := NotionObjectSessionKey(input.WorkspaceID, input.ObjectKind, input.ObjectID)
	revision := NotionDocumentSourceRevision(input)
	provenance := map[string]any{
		"source":             "notion",
		"source_type":        NotionDocumentSourceType,
		"source_key":         sourceKey,
		"source_session_key": sessionKey,
		"source_revision":    revision,
		"workspace_id":       input.WorkspaceID,
		"object_kind":        input.ObjectKind,
		"object_id":          input.ObjectID,
		"page_id":            input.PageID,
		"database_id":        input.DatabaseID,
		"root_id":            strings.TrimSpace(input.RootID),
		"url":                strings.TrimSpace(input.URL),
		"title":              title,
		"traversal_status":   strings.TrimSpace(input.TraversalStatus),
	}
	if strings.TrimSpace(input.SchemaHash) != "" {
		provenance["schema_hash"] = strings.TrimSpace(input.SchemaHash)
	}
	rawProvenance, _ := json.Marshal(provenance)
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString("```rsi-source-provenance-json\n")
	b.Write(rawProvenance)
	b.WriteString("\n```\n\n")
	b.WriteString("Source: Notion\n")
	if strings.TrimSpace(input.URL) != "" {
		b.WriteString("URL: ")
		b.WriteString(strings.TrimSpace(input.URL))
		b.WriteString("\n")
	}
	b.WriteString("Notion object kind: ")
	b.WriteString(input.ObjectKind)
	b.WriteString("\n")
	b.WriteString("Notion object id: ")
	b.WriteString(strings.TrimSpace(input.ObjectID))
	b.WriteString("\n")
	if input.ObjectKind == NotionObjectKindPage {
		b.WriteString("Notion page id: ")
		b.WriteString(strings.TrimSpace(input.PageID))
		b.WriteString("\n")
	}
	if input.ObjectKind == NotionObjectKindDatabase {
		b.WriteString("Notion database id: ")
		b.WriteString(strings.TrimSpace(input.DatabaseID))
		b.WriteString("\n")
	}
	if strings.TrimSpace(input.LastEditedTime) != "" {
		b.WriteString("Last edited: ")
		b.WriteString(strings.TrimSpace(input.LastEditedTime))
		b.WriteString("\n")
	}
	if strings.TrimSpace(input.SchemaSummary) != "" {
		b.WriteString("Database schema summary:\n")
		b.WriteString(strings.TrimSpace(input.SchemaSummary))
		b.WriteString("\n")
	}
	if len(input.Hierarchy) > 0 {
		b.WriteString("Hierarchy: ")
		b.WriteString(strings.Join(input.Hierarchy, " > "))
		b.WriteString("\n")
	}
	if strings.TrimSpace(input.TraversalStatus) != "" && input.TraversalStatus != NotionTraversalComplete {
		b.WriteString("Traversal status: ")
		b.WriteString(strings.TrimSpace(input.TraversalStatus))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	if strings.TrimSpace(input.Content) != "" {
		b.WriteString(strings.TrimSpace(input.Content))
	} else {
		if input.ObjectKind == NotionObjectKindDatabase {
			b.WriteString("Row pages in this database are mirrored as separate Notion documents.")
		} else {
			b.WriteString("(No extractable Notion page body text.)")
		}
	}
	return b.String()
}

func NotionDatabaseSchemaSummary(properties map[string]any) (string, string) {
	canonical := canonicalizeNotionProperties(properties)
	raw, _ := json.Marshal(canonical)
	sum := sha256.Sum256(raw)
	schemaHash := "sha256:" + hex.EncodeToString(sum[:])[:24]
	if len(canonical) == 0 {
		return "(No database properties exposed by Notion API.)", schemaHash
	}
	lines := make([]string, 0, len(canonical))
	for _, property := range canonical {
		name, _ := property["name"].(string)
		kind, _ := property["type"].(string)
		line := "- " + name + ": " + kind
		if values, ok := property["values"].([]string); ok && len(values) > 0 {
			line += " [" + strings.Join(values, ", ") + "]"
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n"), schemaHash
}

func canonicalizeNotionProperties(properties map[string]any) []map[string]any {
	names := sortedKeys(properties)
	out := make([]map[string]any, 0, len(names))
	for _, name := range names {
		raw, _ := properties[name].(map[string]any)
		entry := map[string]any{"name": name}
		kind, _ := raw["type"].(string)
		entry["type"] = strings.TrimSpace(kind)
		if payload, ok := raw[kind].(map[string]any); ok {
			entry["values"] = canonicalNotionPropertyValues(payload)
		}
		out = append(out, entry)
	}
	return out
}

func canonicalNotionPropertyValues(payload map[string]any) []string {
	values := []string{}
	for _, key := range []string{"options", "groups"} {
		rawValues, ok := payload[key].([]any)
		if !ok {
			continue
		}
		for _, rawValue := range rawValues {
			valueMap, _ := rawValue.(map[string]any)
			name, _ := valueMap["name"].(string)
			if strings.TrimSpace(name) != "" {
				values = append(values, strings.TrimSpace(name))
			}
		}
	}
	if len(values) == 0 {
		return nil
	}
	sort.Strings(values)
	return values
}
