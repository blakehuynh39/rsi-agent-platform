package control

import (
	"testing"

	"github.com/slack-go/slack/slackevents"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestSlackSurfaceBuildMentionEnvelopeFiltersAndMapsIdentity(t *testing.T) {
	runtime := newSlackSurfaceRuntime(config.Config{
		SlackAppIdentity:       "oncall",
		AllowedSlackChannelIDs: []string{"C123"},
	}, storepkg.NewMemoryStore())

	envelope, ok := runtime.buildMentionEnvelope("T123", &slackevents.AppMentionEvent{
		Channel:   "C123",
		User:      "U123",
		Text:      "<@U_RSI> Investigate this alert",
		TimeStamp: "171000001.000100",
	})
	if !ok {
		t.Fatal("expected envelope to be created")
	}
	if envelope.BotRole != "oncall" {
		t.Fatalf("unexpected bot role: %s", envelope.BotRole)
	}
	if envelope.ThreadTS == "" {
		t.Fatal("expected thread ts to be populated")
	}

	if _, ok := runtime.buildMentionEnvelope("T123", &slackevents.AppMentionEvent{
		Channel:   "C999",
		User:      "U123",
		Text:      "<@U_RSI> Ignore me",
		TimeStamp: "171000001.000100",
	}); ok {
		t.Fatal("expected channel filter to reject message")
	}
}

func TestSlackSurfaceIgnoresAmbientMessageEvents(t *testing.T) {
	store := storepkg.NewMemoryStore()
	runtime := newSlackSurfaceRuntime(config.Config{
		SlackAppIdentity:       "rsi",
		AllowedSlackChannelIDs: []string{"C123"},
	}, store)
	before := len(store.ListIngestions())

	runtime.handleEventsAPIEvent(slackevents.EventsAPIEvent{
		Type:   slackevents.CallbackEvent,
		TeamID: "T123",
		InnerEvent: slackevents.EventsAPIInnerEvent{
			Type: "message",
			Data: &slackevents.MessageEvent{
				Channel:   "C123",
				User:      "U123",
				Text:      "ambient channel message",
				TimeStamp: "171000001.000100",
			},
		},
	})

	if got := len(store.ListIngestions()); got != before {
		t.Fatalf("expected no new ingestions for ambient messages, before=%d after=%d", before, got)
	}
}

func TestSlackSurfaceAcceptsDirectMessages(t *testing.T) {
	store := storepkg.NewMemoryStore()
	runtime := newSlackSurfaceRuntime(config.Config{
		SlackAppIdentity:       "rsi",
		AllowedSlackChannelIDs: []string{"C123"},
	}, store)
	before := len(store.ListIngestions())

	runtime.handleEventsAPIEvent(slackevents.EventsAPIEvent{
		Type:   slackevents.CallbackEvent,
		TeamID: "T123",
		InnerEvent: slackevents.EventsAPIInnerEvent{
			Type: "message",
			Data: &slackevents.MessageEvent{
				Channel:     "D123",
				ChannelType: "im",
				User:        "U123",
				Text:        "can you explain what happened in #ops",
				TimeStamp:   "171000002.000100",
			},
		},
	})

	ingestions := store.ListIngestions()
	if got := len(ingestions); got != before+1 {
		t.Fatalf("expected one new ingestion for direct messages, before=%d after=%d", before, got)
	}

	found := false
	for _, ingestion := range ingestions {
		if ingestion.ChannelID == "D123" && ingestion.Text == "can you explain what happened in #ops" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected DM ingestion to be present")
	}
}
