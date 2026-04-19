package improvementplane

import (
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
)

type stubSlackTranscriptResolver struct {
	userNames    map[string]string
	channelNames map[string]string
}

func (s stubSlackTranscriptResolver) UserName(userID string) (string, bool) {
	name, ok := s.userNames[userID]
	return name, ok
}

func (s stubSlackTranscriptResolver) ChannelName(channelID string) (string, bool) {
	name, ok := s.channelNames[channelID]
	return name, ok
}

func TestEnrichSlackTranscriptEntriesAddsResolvedSlackLabels(t *testing.T) {
	entries := []conversation.Entry{
		{
			ID:     "entry-1",
			Source: ingestion.SourceSlack,
			Body:   "Ping <@U0ASDQKU3UL> in <#C0AKH5SNGKH> and <#C0AL7EKNHDF>.",
			Metadata: map[string]interface{}{
				"thread_ts": "171000001.000100",
			},
		},
	}

	got := enrichSlackTranscriptEntries(entries, stubSlackTranscriptResolver{
		userNames: map[string]string{
			"U0ASDQKU3UL": "blake",
		},
		channelNames: map[string]string{
			"C0AKH5SNGKH": "depin-backend",
			"C0AL7EKNHDF": "numo-project",
		},
	})

	if got[0].Metadata["thread_ts"] != "171000001.000100" {
		t.Fatalf("expected original metadata to be preserved, got %#v", got[0].Metadata)
	}
	userNames := metadataStringMap(got[0].Metadata[slackUserNamesMetadataKey])
	channelNames := metadataStringMap(got[0].Metadata[slackChannelNamesMetadataKey])
	if userNames["U0ASDQKU3UL"] != "blake" {
		t.Fatalf("expected resolved user name, got %#v", userNames)
	}
	if channelNames["C0AKH5SNGKH"] != "depin-backend" || channelNames["C0AL7EKNHDF"] != "numo-project" {
		t.Fatalf("expected resolved channel names, got %#v", channelNames)
	}
	if _, exists := entries[0].Metadata[slackUserNamesMetadataKey]; exists {
		t.Fatalf("expected enrichment to avoid mutating input slice, got %#v", entries[0].Metadata)
	}
}

func TestEnrichSlackTranscriptEntriesAddsResolvedSlackLabelsForPlainTextBodies(t *testing.T) {
	entries := []conversation.Entry{
		{
			ID:     "entry-plain",
			Source: ingestion.SourceSlack,
			Body:   "Hello @U0ASDQKU3UL, please review #C0AKH5SNGKH and #C0AL7EKNHDF.",
			Metadata: map[string]interface{}{
				"entity_refs": []map[string]any{
					{"kind": "user", "id": "U0ASDQKU3UL", "source": "plain_text"},
					{"kind": "channel", "id": "C0AKH5SNGKH", "source": "plain_text"},
					{"kind": "channel", "id": "C0AL7EKNHDF", "source": "plain_text"},
				},
			},
		},
	}

	got := enrichSlackTranscriptEntries(entries, stubSlackTranscriptResolver{
		userNames: map[string]string{
			"U0ASDQKU3UL": "blake",
		},
		channelNames: map[string]string{
			"C0AKH5SNGKH": "depin-backend",
			"C0AL7EKNHDF": "numo-project",
		},
	})

	userNames := metadataStringMap(got[0].Metadata[slackUserNamesMetadataKey])
	channelNames := metadataStringMap(got[0].Metadata[slackChannelNamesMetadataKey])
	if userNames["U0ASDQKU3UL"] != "blake" {
		t.Fatalf("expected resolved user name for plain text transcript, got %#v", userNames)
	}
	if channelNames["C0AKH5SNGKH"] != "depin-backend" || channelNames["C0AL7EKNHDF"] != "numo-project" {
		t.Fatalf("expected resolved channel names for plain text transcript, got %#v", channelNames)
	}
}

func TestEnrichSlackTranscriptEntriesSkipsNonSlackEntries(t *testing.T) {
	entries := []conversation.Entry{
		{
			ID:     "entry-1",
			Source: ingestion.SourceGitHub,
			Body:   "Mention <@U0ASDQKU3UL> should stay untouched.",
		},
	}

	got := enrichSlackTranscriptEntries(entries, stubSlackTranscriptResolver{
		userNames: map[string]string{"U0ASDQKU3UL": "blake"},
	})

	if got[0].Metadata != nil {
		t.Fatalf("expected non-slack entry metadata to remain nil, got %#v", got[0].Metadata)
	}
}
