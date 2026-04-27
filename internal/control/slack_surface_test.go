package control

import (
	"context"
	"testing"
	"time"

	"github.com/slack-go/slack/slackevents"

	"github.com/piplabs/rsi-agent-platform/internal/app"
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

func TestSlackSurfaceBuildMentionEnvelopeAllowsMentionOnlySentinel(t *testing.T) {
	runtime := newSlackSurfaceRuntime(config.Config{
		SlackAppIdentity:       "rsi",
		AllowedSlackChannelIDs: []string{slackMentionsOnlySentinel},
	}, storepkg.NewMemoryStore())

	envelope, ok := runtime.buildMentionEnvelope("T123", &slackevents.AppMentionEvent{
		Channel:   "C999",
		User:      "U123",
		Text:      "<@U_RSI> Hello from a new training room",
		TimeStamp: "171000001.000100",
	})
	if !ok {
		t.Fatal("expected mention-only sentinel to allow any channel mention")
	}
	if envelope.ChannelID != "C999" {
		t.Fatalf("unexpected channel id: %s", envelope.ChannelID)
	}
}

func TestSlackSurfaceIgnoresAmbientMessageEvents(t *testing.T) {
	store := storepkg.NewMemoryStore()
	runtime := newSlackSurfaceRuntime(config.Config{
		SlackAppIdentity:       "rsi",
		AllowedSlackChannelIDs: []string{"C123"},
	}, store)
	before := len(store.ListIngestions())

	runtime.handleEventsAPIEvent(context.Background(), slackevents.EventsAPIEvent{
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

	runtime.handleEventsAPIEvent(context.Background(), slackevents.EventsAPIEvent{
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

func TestSlackSurfaceIngressContextOnlyUsesTimeoutForDurableAckMode(t *testing.T) {
	legacy := newSlackSurfaceRuntime(config.Config{
		SlackAckAfterDurableIngress:   false,
		SlackDurableIngressAckTimeout: time.Nanosecond,
	}, storepkg.NewMemoryStore())
	parent, parentCancel := context.WithCancel(context.Background())
	parentCancel()
	ctx, cancel, timeout := legacy.ingressContext(parent)
	defer cancel()
	if timeout != 0 {
		t.Fatalf("legacy ack-first mode timeout = %s, want none", timeout)
	}
	if _, ok := ctx.Deadline(); ok {
		t.Fatal("legacy ack-first mode should not cap ingestion after Slack has already been acked")
	}
	select {
	case <-ctx.Done():
		t.Fatal("legacy ack-first mode should not inherit parent cancellation after Slack has already been acked")
	default:
	}

	durable := newSlackSurfaceRuntime(config.Config{
		SlackAckAfterDurableIngress:   true,
		SlackDurableIngressAckTimeout: 50 * time.Millisecond,
	}, storepkg.NewMemoryStore())
	durableParent, durableParentCancel := context.WithCancel(context.Background())
	ctx, cancel, timeout = durable.ingressContext(durableParent)
	defer cancel()
	if timeout != 50*time.Millisecond {
		t.Fatalf("durable ack mode timeout = %s, want 50ms", timeout)
	}
	if _, ok := ctx.Deadline(); !ok {
		t.Fatal("durable ack mode should bound ingress before acknowledging Slack")
	}
	durableParentCancel()
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("durable ack mode should inherit parent cancellation before Slack is acknowledged")
	}
}

func TestSlackSurfaceDrainWatcherCancelsWhenDrainStartsEvenIfDrainFlagDisabled(t *testing.T) {
	app.StopDrainForTest()
	defer app.StopDrainForTest()

	runtime := newSlackSurfaceRuntime(config.Config{DrainEnabled: false}, storepkg.NewMemoryStore())
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	cancelled := make(chan struct{})
	go runtime.watchDrain(ctx, func() {
		stop()
		close(cancelled)
	})

	app.StartDrain()
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("expected Slack drain watcher to cancel after global drain starts")
	}
}

func TestSlackSurfaceDrainWatcherCancelsWhenDrainAlreadyStarted(t *testing.T) {
	app.StopDrainForTest()
	defer app.StopDrainForTest()
	app.StartDrain()

	runtime := newSlackSurfaceRuntime(config.Config{DrainEnabled: false}, storepkg.NewMemoryStore())
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	cancelled := make(chan struct{})
	go runtime.watchDrain(ctx, func() {
		stop()
		close(cancelled)
	})

	select {
	case <-cancelled:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected Slack drain watcher to cancel immediately when drain already started")
	}
}
