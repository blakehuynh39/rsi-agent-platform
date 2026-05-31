package companyknowledge

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

const (
	SlackMessageSourceType = "slack_message"
	DefaultHonchoPeerID    = "slack_unknown"
)

var honchoNameAllowed = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type SlackMessageInput struct {
	WorkspaceID     string
	ChannelID       string
	ChannelType     string
	ChannelPrivate  bool
	ChannelIM       bool
	MirrorDiscovery string
	MirrorAllowed   bool
	MirrorDenied    bool
	TS              string
	ThreadTS        string
	UserID          string
	BotID           string
	Username        string
	Text            string
	EditedTS        string
	EventID         string
	Permalink       string
	ReplyCount      int
	Files           []SlackFileMetadata
	CreatedAt       time.Time
	Raw             map[string]any
}

type SlackFileMetadata struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Title     string `json:"title,omitempty"`
	MimeType  string `json:"mimetype,omitempty"`
	FileType  string `json:"filetype,omitempty"`
	Size      int    `json:"size,omitempty"`
	Permalink string `json:"permalink,omitempty"`
}

type SlackMirrorOptions struct {
	Environment     string
	HonchoWorkspace string
	Lease           time.Duration
}

type SlackMirrorResult struct {
	SourceKey        string
	SourceSessionKey string
	HonchoWorkspace  string
	HonchoSessionID  string
	HonchoMessageID  string
	SourceRevision   string
	Status           string
	Skipped          bool
	SkipReason       string
}

type SlackMirror struct {
	store  store.SourceMirrorWriteStore
	honcho HonchoCorpusClient
	opts   SlackMirrorOptions
}

type HonchoCorpusClient interface {
	EnsureWorkspace(id string, metadata map[string]any) (clients.HonchoWorkspace, error)
	EnsureSession(workspaceID string, sessionID string, metadata map[string]any) (clients.HonchoSession, error)
	CreateMessages(workspaceID string, sessionID string, messages []clients.HonchoMessageCreate) ([]clients.HonchoMessage, error)
}

func NewSlackMirror(state store.SourceMirrorWriteStore, honcho HonchoCorpusClient, opts SlackMirrorOptions) *SlackMirror {
	opts.Environment = strings.TrimSpace(opts.Environment)
	opts.HonchoWorkspace = HonchoCompatibleName("workspace", firstNonEmpty(opts.HonchoWorkspace, "rsi_company_knowledge"))
	if opts.Lease <= 0 {
		opts.Lease = 5 * time.Minute
	}
	return &SlackMirror{store: state, honcho: honcho, opts: opts}
}

func (m *SlackMirror) IngestMessage(ctx context.Context, input SlackMessageInput) (SlackMirrorResult, error) {
	_ = ctx
	if m == nil || m.store == nil || m.honcho == nil {
		return SlackMirrorResult{}, fmt.Errorf("slack mirror requires store and honcho client")
	}
	if err := validateSlackMessage(input); err != nil {
		return SlackMirrorResult{}, err
	}
	sourceKey := SlackMessageSourceKey(input.WorkspaceID, input.ChannelID, input.TS)
	sessionKey := SlackSessionSourceKey(input.WorkspaceID, input.ChannelID, input.EffectiveThreadTS(), input.IsThreaded())
	revision := SlackSourceRevision(input)
	honchoSessionID := HonchoCompatibleName("slack", sessionKey)
	metadata := SlackMessageMetadata(input, sourceKey, sessionKey, revision)
	content := SlackMessageContent(input)
	if strings.TrimSpace(content) == "" {
		return SlackMirrorResult{
			SourceKey:        sourceKey,
			SourceSessionKey: sessionKey,
			HonchoWorkspace:  m.opts.HonchoWorkspace,
			HonchoSessionID:  honchoSessionID,
			SourceRevision:   revision,
			Skipped:          true,
			SkipReason:       "empty_content",
		}, nil
	}
	record := store.SourceMirrorRecord{
		SourceType:       SlackMessageSourceType,
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
		return SlackMirrorResult{}, err
	}
	result := SlackMirrorResult{
		SourceKey:        sourceKey,
		SourceSessionKey: sessionKey,
		HonchoWorkspace:  claim.Record.HonchoWorkspace,
		HonchoSessionID:  claim.Record.HonchoSessionID,
		HonchoMessageID:  claim.Record.HonchoMessageID,
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
		return SlackMirrorResult{}, err
	}
	if _, err := m.honcho.EnsureSession(record.HonchoWorkspace, record.HonchoSessionID, map[string]any{
		"source":             "slack",
		"source_session_key": record.SourceSessionKey,
		"workspace":          record.Workspace,
		"environment":        record.Environment,
	}); err != nil {
		_, _ = m.store.FailSourceMirrorRecord(record.SourceType, record.SourceKey, err.Error(), map[string]any{"failure_stage": "ensure_session"})
		return SlackMirrorResult{}, err
	}
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = SlackTimestampToTime(input.TS)
	}
	messages, err := m.honcho.CreateMessages(record.HonchoWorkspace, record.HonchoSessionID, []clients.HonchoMessageCreate{
		{
			Content:   content,
			PeerID:    HonchoPeerIDForSlack(input),
			Metadata:  metadata,
			CreatedAt: &createdAt,
		},
	})
	if err != nil {
		_, _ = m.store.FailSourceMirrorRecord(record.SourceType, record.SourceKey, err.Error(), map[string]any{"failure_stage": "create_message"})
		return SlackMirrorResult{}, err
	}
	if len(messages) != 1 || strings.TrimSpace(messages[0].ID) == "" {
		err := fmt.Errorf("honcho create message returned %d messages with no stable id", len(messages))
		_, _ = m.store.FailSourceMirrorRecord(record.SourceType, record.SourceKey, err.Error(), map[string]any{"failure_stage": "create_message"})
		return SlackMirrorResult{}, err
	}
	completed, err := m.store.CompleteSourceMirrorRecord(record.SourceType, record.SourceKey, messages[0].ID, map[string]any{
		"honcho_message_id": messages[0].ID,
	})
	if err != nil {
		return SlackMirrorResult{}, err
	}
	result.HonchoMessageID = completed.HonchoMessageID
	result.Status = completed.Status
	result.Skipped = false
	result.SkipReason = ""
	return result, nil
}

func validateSlackMessage(input SlackMessageInput) error {
	if strings.TrimSpace(input.WorkspaceID) == "" {
		return fmt.Errorf("slack workspace id is required")
	}
	if strings.TrimSpace(input.ChannelID) == "" {
		return fmt.Errorf("slack channel id is required")
	}
	if strings.TrimSpace(input.TS) == "" {
		return fmt.Errorf("slack message ts is required")
	}
	return nil
}

func SlackMessageContent(input SlackMessageInput) string {
	if strings.TrimSpace(input.Text) != "" {
		return input.Text
	}
	if len(input.Files) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Slack file attachment")
	if len(input.Files) > 1 {
		b.WriteString("s")
	}
	b.WriteString(":")
	for _, file := range input.Files {
		label := firstNonEmpty(file.Title, file.Name, file.ID, "untitled")
		b.WriteString("\n- ")
		b.WriteString(label)
		details := []string{}
		if strings.TrimSpace(file.MimeType) != "" {
			details = append(details, strings.TrimSpace(file.MimeType))
		}
		if strings.TrimSpace(file.FileType) != "" {
			details = append(details, strings.TrimSpace(file.FileType))
		}
		if file.Size > 0 {
			details = append(details, strconv.Itoa(file.Size)+" bytes")
		}
		if len(details) > 0 {
			b.WriteString(" (")
			b.WriteString(strings.Join(details, ", "))
			b.WriteString(")")
		}
		if strings.TrimSpace(file.Permalink) != "" {
			b.WriteString(" ")
			b.WriteString(strings.TrimSpace(file.Permalink))
		}
	}
	return b.String()
}

func (m SlackMessageInput) EffectiveThreadTS() string {
	if strings.TrimSpace(m.ThreadTS) != "" {
		return strings.TrimSpace(m.ThreadTS)
	}
	return strings.TrimSpace(m.TS)
}

func (m SlackMessageInput) IsThreaded() bool {
	threadTS := strings.TrimSpace(m.ThreadTS)
	if threadTS == "" {
		return false
	}
	return threadTS != strings.TrimSpace(m.TS) || m.ReplyCount > 0
}

func SlackMessageSourceKey(workspaceID string, channelID string, ts string) string {
	return "slack:" + strings.TrimSpace(workspaceID) + ":" + strings.TrimSpace(channelID) + ":" + strings.TrimSpace(ts)
}

func SlackSessionSourceKey(workspaceID string, channelID string, threadTS string, threaded bool) string {
	if threaded {
		return "slack:" + strings.TrimSpace(workspaceID) + ":" + strings.TrimSpace(channelID) + ":" + strings.TrimSpace(threadTS)
	}
	return "slack:" + strings.TrimSpace(workspaceID) + ":" + strings.TrimSpace(channelID) + ":channel"
}

func SlackSourceRevision(input SlackMessageInput) string {
	if strings.TrimSpace(input.EditedTS) != "" {
		return "edited:" + strings.TrimSpace(input.EditedTS)
	}
	payload := map[string]any{
		"ts":    strings.TrimSpace(input.TS),
		"text":  input.Text,
		"files": input.Files,
	}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return "ts:" + strings.TrimSpace(input.TS) + ":sha256:" + hex.EncodeToString(sum[:])[:24]
}

func SlackMessageMetadata(input SlackMessageInput, sourceKey string, sessionKey string, revision string) map[string]any {
	metadata := map[string]any{
		"source":             "slack",
		"source_key":         sourceKey,
		"source_dedupe_key":  sourceKey,
		"source_revision":    revision,
		"source_session_key": sessionKey,
		"workspace_id":       strings.TrimSpace(input.WorkspaceID),
		"channel_id":         strings.TrimSpace(input.ChannelID),
		"channel_type":       strings.TrimSpace(input.ChannelType),
		"channel_private":    input.ChannelPrivate,
		"channel_im":         input.ChannelIM,
		"mirror_discovery":   strings.TrimSpace(input.MirrorDiscovery),
		"mirror_allowed":     input.MirrorAllowed,
		"mirror_denied":      input.MirrorDenied,
		"slack_ts":           strings.TrimSpace(input.TS),
		"edited_ts":          strings.TrimSpace(input.EditedTS),
		"thread_ts":          input.EffectiveThreadTS(),
		"permalink":          strings.TrimSpace(input.Permalink),
		"event_id":           strings.TrimSpace(input.EventID),
		"user_id":            strings.TrimSpace(input.UserID),
		"bot_id":             strings.TrimSpace(input.BotID),
		"username":           strings.TrimSpace(input.Username),
		"files":              input.Files,
	}
	if len(input.Raw) > 0 {
		metadata["raw_keys"] = sortedKeys(input.Raw)
		if checked, ok := input.Raw["channel_info_checked"]; ok {
			metadata["channel_info_checked"] = checked
		}
		if infoErr, ok := input.Raw["channel_info_error"]; ok {
			metadata["channel_info_error"] = infoErr
		}
	}
	return metadata
}

func SlackWikiSourceRevisionInput(input SlackMessageInput) store.CompanyWikiSourceRevisionInput {
	sourceKey := SlackMessageSourceKey(input.WorkspaceID, input.ChannelID, input.TS)
	sessionKey := SlackSessionSourceKey(input.WorkspaceID, input.ChannelID, input.EffectiveThreadTS(), input.IsThreaded())
	revision := SlackSourceRevision(input)
	title := "Slack thread " + strings.TrimSpace(input.ChannelID)
	if input.IsThreaded() {
		title += " " + input.EffectiveThreadTS()
	} else {
		title += " channel"
	}
	nativeLocator := firstNonEmpty(
		strings.TrimSpace(input.Permalink),
		"slack:"+strings.TrimSpace(input.ChannelID)+":"+input.EffectiveThreadTS()+":"+strings.TrimSpace(input.TS),
	)
	metadata := SlackMessageMetadata(input, sourceKey, sessionKey, revision)
	return store.CompanyWikiSourceRevisionInput{
		SourceType:        SlackMessageSourceType,
		DocumentSourceKey: sessionKey,
		SourceKey:         sourceKey,
		SourceSessionKey:  sessionKey,
		Workspace:         strings.TrimSpace(input.WorkspaceID),
		Title:             title,
		URL:               strings.TrimSpace(input.Permalink),
		SourceRevision:    sourceKey + ":" + revision,
		Content:           SlackMessageContent(input),
		NativeLocator:     nativeLocator,
		Metadata:          metadata,
		ObservedAt:        input.CreatedAt,
	}
}

func HonchoPeerIDForSlack(input SlackMessageInput) string {
	for _, value := range []string{input.UserID, input.BotID, input.Username} {
		if strings.TrimSpace(value) != "" {
			return HonchoCompatibleName("slack_peer", value)
		}
	}
	return DefaultHonchoPeerID
}

func HonchoCompatibleName(prefix string, raw string) string {
	prefix = sanitizeHonchoNamePart(prefix)
	if prefix == "" {
		prefix = "rsi"
	}
	raw = strings.TrimSpace(raw)
	sanitized := sanitizeHonchoNamePart(raw)
	if sanitized == "" {
		sanitized = "empty"
	}
	candidate := sanitized
	if len(candidate) > 100 || !honchoNameAllowed.MatchString(candidate) {
		sum := sha256.Sum256([]byte(raw))
		return trimHonchoName(prefix+"_"+hex.EncodeToString(sum[:])[:48], "")
	}
	if len(candidate) <= 100 {
		return candidate
	}
	sum := sha256.Sum256([]byte(raw))
	return trimHonchoName(candidate, hex.EncodeToString(sum[:])[:16])
}

func sanitizeHonchoNamePart(value string) string {
	value = strings.TrimSpace(value)
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return strings.Trim(b.String(), "_-")
}

func trimHonchoName(base string, suffix string) string {
	if suffix != "" {
		suffix = "_" + suffix
	}
	if len(base)+len(suffix) <= 100 {
		return base + suffix
	}
	limit := 100 - len(suffix)
	if limit <= 0 {
		return base[:100]
	}
	return strings.TrimRight(base[:limit], "_-") + suffix
}

func SlackTimestampToTime(ts string) time.Time {
	parts := strings.Split(strings.TrimSpace(ts), ".")
	if len(parts) == 0 || parts[0] == "" {
		return time.Time{}
	}
	seconds, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}
	}
	nanos := int64(0)
	if len(parts) > 1 {
		microsText := parts[1]
		if len(microsText) > 6 {
			microsText = microsText[:6]
		}
		for len(microsText) < 6 {
			microsText += "0"
		}
		micros, err := strconv.ParseInt(microsText, 10, 64)
		if err == nil {
			nanos = micros * int64(time.Microsecond)
		}
	}
	return time.Unix(seconds, nanos).UTC()
}

func sortedKeys(input map[string]any) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
