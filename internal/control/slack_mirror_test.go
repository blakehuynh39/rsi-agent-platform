package control

import (
	"context"
	"errors"
	"testing"
	"time"

	slackapi "github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/config"
)

type fakeSlackChannelLister struct {
	pages []fakeSlackChannelPage
	err   error
}

type fakeSlackChannelPage struct {
	channels []slackapi.Channel
	cursor   string
}

func (f *fakeSlackChannelLister) GetConversationsContext(ctx context.Context, params *slackapi.GetConversationsParameters) ([]slackapi.Channel, string, error) {
	_ = ctx
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
