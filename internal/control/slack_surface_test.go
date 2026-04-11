package control

import (
	"testing"

	"github.com/slack-go/slack/slackevents"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestSlackSurfaceBuildEnvelopeFiltersAndMapsIdentity(t *testing.T) {
	runtime := newSlackSurfaceRuntime(config.Config{
		SlackAppIdentity:       "oncall",
		AllowedSlackChannelIDs: []string{"C123"},
	}, storepkg.NewMemoryStore())

	envelope, ok := runtime.buildEnvelope("T123", &slackevents.MessageEvent{
		Channel:   "C123",
		User:      "U123",
		Text:      "Investigate this alert",
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

	if _, ok := runtime.buildEnvelope("T123", &slackevents.MessageEvent{
		Channel:   "C999",
		User:      "U123",
		Text:      "Ignore me",
		TimeStamp: "171000001.000100",
	}); ok {
		t.Fatal("expected channel filter to reject message")
	}
}
