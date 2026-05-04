package control

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	slackapi "github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

type fakeSlackChannelLister struct {
	pages []fakeSlackChannelPage
	err   error
	calls []slackapi.GetConversationsParameters
}

type fakeSlackChannelPage struct {
	channels []slackapi.Channel
	cursor   string
}

func (f *fakeSlackChannelLister) GetConversationsContext(ctx context.Context, params *slackapi.GetConversationsParameters) ([]slackapi.Channel, string, error) {
	_ = ctx
	if params != nil {
		f.calls = append(f.calls, *params)
	}
	if f.err != nil {
		return nil, "", f.err
	}
	index := 0
	if params != nil && params.Cursor != "" {
		for candidate, page := range f.pages {
			if page.cursor == params.Cursor {
				index = candidate + 1
				break
			}
		}
	}
	if index >= len(f.pages) {
		return nil, "", nil
	}
	page := f.pages[index]
	return page.channels, page.cursor, nil
}

func TestSlackMirrorChannelsDefaultToJoinedDiscovery(t *testing.T) {
	lister := &fakeSlackChannelLister{
		pages: []fakeSlackChannelPage{
			{
				channels: []slackapi.Channel{
					{GroupConversation: slackapi.GroupConversation{Conversation: slackapi.Conversation{ID: "CJOINED"}, Name: "joined"}, IsMember: true},
					{GroupConversation: slackapi.GroupConversation{Conversation: slackapi.Conversation{ID: "CNOTMEMBER"}, Name: "not-member"}, IsMember: false},
					{GroupConversation: slackapi.GroupConversation{Conversation: slackapi.Conversation{ID: "CARCHIVED"}, Name: "archived", IsArchived: true}, IsMember: true},
				},
				cursor: "next",
			},
			{
				channels: []slackapi.Channel{
					{GroupConversation: slackapi.GroupConversation{Conversation: slackapi.Conversation{ID: "CPRIVATE"}, Name: "private"}, IsMember: true},
				},
			},
		},
	}

	got, err := slackMirrorChannels(context.Background(), config.Config{
		SlackMirrorChannelAllowlist: []string{"CEXPLICIT", "CJOINED"},
		SlackMirrorChannelDenylist:  []string{"CJOINED"},
	}, lister)
	if err != nil {
		t.Fatalf("slackMirrorChannels() error = %v", err)
	}
	want := []string{"CPRIVATE"}
	if len(got) != len(want) {
		t.Fatalf("channels = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("channels = %v, want %v", got, want)
		}
	}
}

func TestSlackMirrorChannelsJoinedPublicExcludesPrivateChannels(t *testing.T) {
	lister := &fakeSlackChannelLister{
		pages: []fakeSlackChannelPage{{
			channels: []slackapi.Channel{
				{GroupConversation: slackapi.GroupConversation{Conversation: slackapi.Conversation{ID: "CPUBLIC"}, Name: "public"}, IsMember: true},
				{GroupConversation: slackapi.GroupConversation{Conversation: slackapi.Conversation{ID: "CPRIVATE", IsPrivate: true}, Name: "private"}, IsMember: true},
			},
		}},
	}
	got, err := slackMirrorChannels(context.Background(), config.Config{
		SlackMirrorChannelDiscovery: "joined_public",
	}, lister)
	if err != nil {
		t.Fatalf("slackMirrorChannels() error = %v", err)
	}
	if len(got) != 1 || got[0] != "CPUBLIC" {
		t.Fatalf("channels = %v, want [CPUBLIC]", got)
	}
	if len(lister.calls) != 1 || len(lister.calls[0].Types) != 1 || lister.calls[0].Types[0] != "public_channel" {
		t.Fatalf("joined_public should only request public channels, calls=%+v", lister.calls)
	}
}

func TestSlackMirrorChannelsExplicitDiscoveryUsesAllowlistOnly(t *testing.T) {
	got, err := slackMirrorChannels(context.Background(), config.Config{
		SlackMirrorChannelDiscovery: "explicit",
		SlackMirrorChannelAllowlist: []string{"C2", "C1", "C2"},
		SlackMirrorChannelDenylist:  []string{"C2"},
	}, &fakeSlackChannelLister{err: errors.New("should not list")})
	if err != nil {
		t.Fatalf("slackMirrorChannels() error = %v", err)
	}
	if len(got) != 1 || got[0] != "C1" {
		t.Fatalf("channels = %v, want [C1]", got)
	}
}

func TestSlackMirrorChannelAllowedByConfigDefaultsToJoinedUnlessDenied(t *testing.T) {
	cfg := config.Config{SlackMirrorChannelDenylist: []string{"CNOPE"}}
	if !slackMirrorChannelAllowedByConfig(cfg, "CANY") {
		t.Fatal("joined discovery should allow any non-denied channel event")
	}
	if slackMirrorChannelAllowedByConfig(cfg, "CNOPE") {
		t.Fatal("denylist should reject channel event")
	}

	cfg.SlackMirrorChannelDiscovery = "explicit"
	cfg.SlackMirrorChannelAllowlist = []string{"CYES"}
	if !slackMirrorChannelAllowedByConfig(cfg, "CYES") {
		t.Fatal("explicit allowlist should allow listed channel")
	}
	if slackMirrorChannelAllowedByConfig(cfg, "CANY") {
		t.Fatal("explicit discovery should reject unlisted channel")
	}
}

func TestApplySlackMirrorPolicyMetadataUsesRealChannelPrivacy(t *testing.T) {
	input := companyknowledge.SlackMessageInput{ChannelID: "CPRIVATE"}
	applySlackMirrorPolicyMetadata(&input, config.Config{}, slackMirrorChannelMetadata{
		ChannelID:      "CPRIVATE",
		ChannelType:    "private_channel",
		ChannelPrivate: true,
		InfoChecked:    true,
	})
	if !input.ChannelPrivate || input.ChannelType != "private_channel" {
		t.Fatalf("expected private channel metadata, got %+v", input)
	}
	if input.MirrorDenied || !input.MirrorAllowed {
		t.Fatalf("joined mirror should keep ingestion allowed while source policy can reject private synthesis, got allowed=%t denied=%t", input.MirrorAllowed, input.MirrorDenied)
	}
}

func TestApplySlackMirrorPolicyMetadataDeniesUnknownChannelInfoForWikiPolicy(t *testing.T) {
	input := companyknowledge.SlackMessageInput{ChannelID: "CUNKNOWN"}
	applySlackMirrorPolicyMetadata(&input, config.Config{}, slackMirrorChannelMetadata{
		ChannelID:       "CUNKNOWN",
		ChannelType:     "unknown",
		ChannelPrivate:  true,
		PolicyUntrusted: true,
		InfoError:       "missing_scope",
	})
	if !input.MirrorDenied || input.MirrorAllowed {
		t.Fatalf("untrusted channel info should deny wiki policy metadata, got allowed=%t denied=%t", input.MirrorAllowed, input.MirrorDenied)
	}
	if !input.ChannelPrivate || input.ChannelType != "unknown" {
		t.Fatalf("unknown channel should be marked private by default, got %+v", input)
	}
	if input.Raw["channel_info_error"] != "missing_scope" {
		t.Fatalf("channel info error not persisted in raw metadata: %+v", input.Raw)
	}
}

func TestSlackWikiPublishBatchCompilesAfterPageBatch(t *testing.T) {
	state := store.NewMemoryStore()
	batch := newSlackWikiPublishBatch(config.Config{CompanyWikiRoot: t.TempDir()}, state)
	inputs := []companyknowledge.SlackMessageInput{
		{WorkspaceID: "T123", ChannelID: "C123", TS: "1777600000.000001", Text: "First deploy decision."},
		{WorkspaceID: "T123", ChannelID: "C123", TS: "1777600001.000001", Text: "Second deploy decision."},
		{WorkspaceID: "T123", ChannelID: "C123", TS: "1777600002.000001", Text: "Third deploy decision."},
	}
	for _, input := range inputs {
		if err := batch.record(context.Background(), input); err != nil {
			t.Fatalf("record() error = %v", err)
		}
	}
	entries, err := state.ListCompanyWikiManifestEntries()
	if err != nil {
		t.Fatalf("ListCompanyWikiManifestEntries() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("batch should not publish before publish() is called, entries=%+v", entries)
	}
	if err := batch.publish(context.Background()); err != nil {
		t.Fatalf("publish() error = %v", err)
	}
	entries, err = state.ListCompanyWikiManifestEntries()
	if err != nil {
		t.Fatalf("ListCompanyWikiManifestEntries() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one compiled channel page, got %+v", entries)
	}
	page, found, err := state.GetCompanyWikiPage(entries[0].WikiPageID)
	if err != nil || !found {
		t.Fatalf("GetCompanyWikiPage() found=%t err=%v", found, err)
	}
	if page.Revision.RevisionNumber != 1 {
		t.Fatalf("expected one wiki revision for batched messages, got %d", page.Revision.RevisionNumber)
	}
	for _, want := range []string{"First deploy decision.", "Second deploy decision.", "Third deploy decision."} {
		if !strings.Contains(page.Revision.Body, want) {
			t.Fatalf("compiled body missing %q:\n%s", want, page.Revision.Body)
		}
	}
}

func TestSlackMirrorHistoryWindowResumesBackfillBeforeOldestSeenPage(t *testing.T) {
	checkpoint := slackMirrorCheckpoint{}
	oldest, latest, mode := slackMirrorHistoryWindow(checkpoint)
	if oldest != "" || latest != "" || mode != "backfill" {
		t.Fatalf("empty checkpoint should start backfill, got oldest=%q latest=%q mode=%q", oldest, latest, mode)
	}

	updateSlackMirrorCheckpointProgress(&checkpoint, "T123", "C123", "backfill", "1777700000.000000", "1777600000.000000", 200, 12, false)
	if checkpoint.BackfillComplete {
		t.Fatal("in-progress backfill should not be complete")
	}
	if checkpoint.BackfillBeforeTS != "1777600000.000000" {
		t.Fatalf("expected backfill_before_ts to advance to page oldest ts, got %q", checkpoint.BackfillBeforeTS)
	}
	if checkpoint.LastMirroredTS != "1777700000.000000" {
		t.Fatalf("expected newest high-watermark to be retained, got %q", checkpoint.LastMirroredTS)
	}

	oldest, latest, mode = slackMirrorHistoryWindow(checkpoint)
	if oldest != "" || latest != "1777600000.000000" || mode != "backfill" {
		t.Fatalf("in-progress checkpoint should continue older history, got oldest=%q latest=%q mode=%q", oldest, latest, mode)
	}
}

func TestSlackMirrorBackfillCompletionSwitchesFutureRunsToIncremental(t *testing.T) {
	checkpoint := slackMirrorCheckpoint{
		ChannelID:        "C123",
		WorkspaceID:      "T123",
		LastMirroredTS:   "1777700000.000000",
		BackfillBeforeTS: "1777000000.000000",
	}

	updateSlackMirrorCheckpointProgress(&checkpoint, "T123", "C123", "backfill", "1777700000.000000", "", 33, 4, true)
	if !checkpoint.BackfillComplete {
		t.Fatal("completed backfill should be marked complete")
	}
	if checkpoint.BackfillBeforeTS != "" {
		t.Fatalf("completed backfill should clear backfill_before_ts, got %q", checkpoint.BackfillBeforeTS)
	}
	if checkpoint.LastCompletedAt.IsZero() {
		t.Fatal("completed backfill should record completion time")
	}

	oldest, latest, mode := slackMirrorHistoryWindow(checkpoint)
	if oldest != "1777700000.000000" || latest != "" || mode != "incremental" {
		t.Fatalf("completed checkpoint should use incremental window, got oldest=%q latest=%q mode=%q", oldest, latest, mode)
	}
}

func TestSlackMirrorTimestampBoundsIgnoreEmptyTimestamps(t *testing.T) {
	oldest, newest := slackMirrorMessageTimestampBounds([]slackapi.Message{
		{Msg: slackapi.Msg{Timestamp: "1777600000.000000"}},
		{Msg: slackapi.Msg{Timestamp: ""}},
		{Msg: slackapi.Msg{Timestamp: "1777700000.000000"}},
		{Msg: slackapi.Msg{Timestamp: "1777500000.000000"}},
	})
	if oldest != "1777500000.000000" || newest != "1777700000.000000" {
		t.Fatalf("unexpected bounds oldest=%q newest=%q", oldest, newest)
	}
}

func TestSlackMirrorTimestampIsSlackCompatible(t *testing.T) {
	got := slackMirrorTimestamp(time.Unix(1777700000, 123456789).UTC())
	if got != "1777700000.123456" {
		t.Fatalf("unexpected Slack timestamp %q", got)
	}
}
