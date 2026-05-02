package control

import (
	"testing"
	"time"

	slackapi "github.com/slack-go/slack"
)

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
