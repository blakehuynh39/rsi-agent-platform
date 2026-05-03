package companyknowledge

import (
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

type fakeHonchoCorpus struct {
	workspaceCalls int
	sessionCalls   int
	createCalls    int
	nextID         int
	messages       []clients.HonchoMessageCreate
}

func (f *fakeHonchoCorpus) EnsureWorkspace(id string, metadata map[string]any) (clients.HonchoWorkspace, error) {
	f.workspaceCalls++
	return clients.HonchoWorkspace{ID: id, Metadata: metadata}, nil
}

func (f *fakeHonchoCorpus) EnsureSession(workspaceID string, sessionID string, metadata map[string]any) (clients.HonchoSession, error) {
	f.sessionCalls++
	return clients.HonchoSession{ID: sessionID, WorkspaceID: workspaceID, Metadata: metadata}, nil
}

func (f *fakeHonchoCorpus) CreateMessages(workspaceID string, sessionID string, messages []clients.HonchoMessageCreate) ([]clients.HonchoMessage, error) {
	f.createCalls++
	f.messages = append(f.messages, messages...)
	out := make([]clients.HonchoMessage, 0, len(messages))
	for range messages {
		f.nextID++
		out = append(out, clients.HonchoMessage{ID: "msg_" + string(rune('a'+f.nextID-1)), WorkspaceID: workspaceID, SessionID: sessionID})
	}
	return out, nil
}

func TestSlackMirrorIngestIsIdempotentBySourceKeyAndRevision(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeHonchoCorpus{}
	mirror := NewSlackMirror(state, honcho, SlackMirrorOptions{
		Environment:     "stage",
		HonchoWorkspace: "rsi_company_knowledge",
		Lease:           time.Minute,
	})
	input := SlackMessageInput{
		WorkspaceID: "T123",
		ChannelID:   "C123",
		TS:          "1777650186.068179",
		UserID:      "U123",
		Text:        "how should CORS work for /auth/exchange?",
	}
	first, err := mirror.IngestMessage(nil, input)
	if err != nil {
		t.Fatalf("first ingest error = %v", err)
	}
	second, err := mirror.IngestMessage(nil, input)
	if err != nil {
		t.Fatalf("second ingest error = %v", err)
	}
	if first.Skipped {
		t.Fatalf("first ingest skipped unexpectedly: %+v", first)
	}
	if !second.Skipped || second.SkipReason != "already_complete" {
		t.Fatalf("second ingest = %+v, want already_complete skip", second)
	}
	if honcho.createCalls != 1 {
		t.Fatalf("CreateMessages calls = %d, want 1", honcho.createCalls)
	}
	record, found, err := state.GetSourceMirrorRecord(SlackMessageSourceType, SlackMessageSourceKey("T123", "C123", "1777650186.068179"))
	if err != nil || !found {
		t.Fatalf("record found=%v err=%v", found, err)
	}
	if record.Status != store.SourceMirrorStatusComplete || record.HonchoMessageID == "" {
		t.Fatalf("record = %+v, want complete with honcho id", record)
	}
}

func TestSlackMirrorRevisionChangeCreatesNewHonchoMessageAndUpdatesPointer(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeHonchoCorpus{}
	mirror := NewSlackMirror(state, honcho, SlackMirrorOptions{Environment: "stage"})
	input := SlackMessageInput{WorkspaceID: "T123", ChannelID: "C123", TS: "1777650186.068179", UserID: "U123", Text: "first"}
	first, err := mirror.IngestMessage(nil, input)
	if err != nil {
		t.Fatalf("first ingest error = %v", err)
	}
	input.Text = "edited"
	input.EditedTS = "1777650190.000001"
	second, err := mirror.IngestMessage(nil, input)
	if err != nil {
		t.Fatalf("second ingest error = %v", err)
	}
	if first.HonchoMessageID == second.HonchoMessageID {
		t.Fatalf("honcho message id did not advance across revision: first=%s second=%s", first.HonchoMessageID, second.HonchoMessageID)
	}
	if honcho.createCalls != 2 {
		t.Fatalf("CreateMessages calls = %d, want 2", honcho.createCalls)
	}
	record, found, err := state.GetSourceMirrorRecord(SlackMessageSourceType, SlackMessageSourceKey("T123", "C123", "1777650186.068179"))
	if err != nil || !found {
		t.Fatalf("record found=%v err=%v", found, err)
	}
	if record.SourceRevision != "edited:1777650190.000001" {
		t.Fatalf("source revision = %q", record.SourceRevision)
	}
	if record.HonchoMessageID != second.HonchoMessageID {
		t.Fatalf("record honcho id = %q, want %q", record.HonchoMessageID, second.HonchoMessageID)
	}
}

func TestSlackMirrorFileOnlyMessageWritesFileMetadataContent(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeHonchoCorpus{}
	mirror := NewSlackMirror(state, honcho, SlackMirrorOptions{Environment: "stage"})
	input := SlackMessageInput{
		WorkspaceID: "T123",
		ChannelID:   "C123",
		TS:          "1777650186.068179",
		UserID:      "U123",
		Files: []SlackFileMetadata{
			{
				ID:        "F123",
				Title:     "deploy-log.txt",
				MimeType:  "text/plain",
				Size:      42,
				Permalink: "https://slack.example/files/F123",
			},
		},
	}

	result, err := mirror.IngestMessage(nil, input)
	if err != nil {
		t.Fatalf("ingest error = %v", err)
	}
	if result.Skipped {
		t.Fatalf("file-only message skipped unexpectedly: %+v", result)
	}
	if honcho.createCalls != 1 || len(honcho.messages) != 1 {
		t.Fatalf("CreateMessages calls=%d messages=%d, want one message", honcho.createCalls, len(honcho.messages))
	}
	content := honcho.messages[0].Content
	if content == "" || content == input.Text {
		t.Fatalf("content = %q, want generated file metadata content", content)
	}
	for _, want := range []string{"deploy-log.txt", "text/plain", "https://slack.example/files/F123"} {
		if !strings.Contains(content, want) {
			t.Fatalf("content = %q, want %q", content, want)
		}
	}
}

func TestSlackMirrorEmptyMessageSkipsWithoutClaimingRecord(t *testing.T) {
	state := store.NewMemoryStore()
	honcho := &fakeHonchoCorpus{}
	mirror := NewSlackMirror(state, honcho, SlackMirrorOptions{Environment: "stage"})
	input := SlackMessageInput{
		WorkspaceID: "T123",
		ChannelID:   "C123",
		TS:          "1777650186.068179",
		UserID:      "U123",
	}

	result, err := mirror.IngestMessage(nil, input)
	if err != nil {
		t.Fatalf("ingest error = %v", err)
	}
	if !result.Skipped || result.SkipReason != "empty_content" {
		t.Fatalf("result = %+v, want empty_content skip", result)
	}
	if honcho.createCalls != 0 {
		t.Fatalf("CreateMessages calls = %d, want 0", honcho.createCalls)
	}
	_, found, err := state.GetSourceMirrorRecord(SlackMessageSourceType, SlackMessageSourceKey("T123", "C123", "1777650186.068179"))
	if err != nil {
		t.Fatalf("GetSourceMirrorRecord() error = %v", err)
	}
	if found {
		t.Fatal("empty message should not create a source mirror record")
	}
}

func TestSlackSessionMappingAndHonchoIDEncoding(t *testing.T) {
	thread := SlackSessionSourceKey("T123", "C123", "1777650186.068179", true)
	if thread != "slack:T123:C123:1777650186.068179" {
		t.Fatalf("thread session key = %q", thread)
	}
	channel := SlackSessionSourceKey("T123", "C123", "", false)
	if channel != "slack:T123:C123:channel" {
		t.Fatalf("channel session key = %q", channel)
	}
	encoded := HonchoCompatibleName("slack", thread)
	if len(encoded) > 100 {
		t.Fatalf("encoded id too long: %q", encoded)
	}
	for _, r := range encoded {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') && r != '_' && r != '-' {
			t.Fatalf("encoded id %q contains invalid rune %q", encoded, r)
		}
	}
}
