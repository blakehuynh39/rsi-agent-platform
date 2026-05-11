package control

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type fakeSlackPost struct {
	channelID string
	values    url.Values
}

type fakeSlackPoster struct {
	calls []fakeSlackPost
	err   error
}

type fakeSurfaceHonchoCorpus struct {
	createCalls int
}

func (f *fakeSurfaceHonchoCorpus) EnsureWorkspace(id string, metadata map[string]any) (clients.HonchoWorkspace, error) {
	return clients.HonchoWorkspace{ID: id, Metadata: metadata}, nil
}

func (f *fakeSurfaceHonchoCorpus) EnsureSession(workspaceID string, sessionID string, metadata map[string]any) (clients.HonchoSession, error) {
	return clients.HonchoSession{ID: sessionID, WorkspaceID: workspaceID, Metadata: metadata}, nil
}

func (f *fakeSurfaceHonchoCorpus) CreateMessages(workspaceID string, sessionID string, messages []clients.HonchoMessageCreate) ([]clients.HonchoMessage, error) {
	f.createCalls++
	return []clients.HonchoMessage{{ID: "msg_surface_1", WorkspaceID: workspaceID, SessionID: sessionID}}, nil
}

func (f *fakeSlackPoster) PostMessageContext(_ context.Context, channelID string, options ...slack.MsgOption) (string, string, error) {
	_, values, _ := slack.UnsafeApplyMsgOptions("xoxb-test", channelID, "https://slack.com/api/", options...)
	f.calls = append(f.calls, fakeSlackPost{channelID: channelID, values: values})
	if f.err != nil {
		return "", "", f.err
	}
	return channelID, "171000001.000200", nil
}

func (f *fakeSlackPoster) UpdateMessageContext(_ context.Context, channelID, timestamp string, options ...slack.MsgOption) (string, string, string, error) {
	_, values, _ := slack.UnsafeApplyMsgOptions("xoxb-test", channelID, "https://slack.com/api/", options...)
	f.calls = append(f.calls, fakeSlackPost{channelID: channelID, values: values})
	if f.err != nil {
		return "", "", "", f.err
	}
	return channelID, timestamp, "", nil
}

func TestSlackSurfaceBuildMentionEnvelopeFiltersAndMapsIdentity(t *testing.T) {
	runtime := newSlackSurfaceRuntime(config.Config{
		SlackAppIdentity:              "oncall",
		SlackIngressAllowedChannelIDs: []string{"C123"},
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

func TestOperatorTraceURLBuilder(t *testing.T) {
	traceURL := operatorTraceURL("https://staging-rsi-platform.storyprotocol.net/", "conv-1", "trace-1")
	parsed, err := url.Parse(traceURL)
	if err != nil {
		t.Fatalf("operatorTraceURL returned invalid URL %q: %v", traceURL, err)
	}
	if got := parsed.Scheme + "://" + parsed.Host + parsed.Path; got != "https://staging-rsi-platform.storyprotocol.net/sessions" {
		t.Fatalf("unexpected base URL: %s", got)
	}
	values := parsed.Query()
	if values.Get("tab") != "conversations" || values.Get("conversation") != "conv-1" || values.Get("trace") != "trace-1" {
		t.Fatalf("unexpected trace URL query: %s", values.Encode())
	}
}

func TestSlackSurfaceOperatorTraceACKIsIdempotent(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		ServiceName:      "rsi-slack-surface",
		SlackAppIdentity: "rsi",
		PublicBaseURL:    "https://staging-rsi-platform.storyprotocol.net",
	}
	runtime := newSlackSurfaceRuntime(cfg, store)
	poster := &fakeSlackPoster{}
	runtime.slackAPI = poster
	now := time.Unix(171000001, 100000000).UTC()
	receipt, err := submitIngressSlackCommand(
		context.Background(),
		cfg,
		store,
		slackpkg.SlackEnvelope{
			BotRole:   slackpkg.BotOrchestrator,
			TeamID:    "T123",
			ChannelID: "C123",
			ThreadTS:  "171000001.000100",
			UserID:    "U123",
			Text:      "<@U_RSI> run this",
			TS:        "171000001.000100",
			CreatedAt: now,
		},
		"test",
		now,
		"cmd-test-operator-ack",
	)
	if err != nil {
		t.Fatalf("submitIngressSlackCommand() error = %v", err)
	}
	ingestion, err := loadSlackIngestionFromReceipt(store, receipt)
	if err != nil {
		t.Fatalf("loadSlackIngestionFromReceipt() error = %v", err)
	}

	runtime.postOperatorTraceACK(context.Background(), ingestion)
	runtime.postOperatorTraceACK(context.Background(), ingestion)

	if len(poster.calls) != 1 {
		t.Fatalf("expected one Slack ACK post after duplicate calls, got %d", len(poster.calls))
	}
	call := poster.calls[0]
	if call.channelID != "C123" || call.values.Get("thread_ts") != "171000001.000100" {
		t.Fatalf("ACK posted to wrong target: channel=%s values=%s", call.channelID, call.values.Encode())
	}
	if !strings.Contains(call.values.Get("text"), "https://staging-rsi-platform.storyprotocol.net/sessions?") {
		t.Fatalf("ACK text missing trace URL: %q", call.values.Get("text"))
	}
	trace, ok := traceSummaryForIngestion(store, ingestion)
	if !ok {
		t.Fatal("expected trace for ingestion")
	}
	ledger := store.ListExecutionLedgerEventsByTrace(trace.TraceID)
	if len(ledger) != 1 {
		t.Fatalf("expected one ACK ledger event, got %#v", ledger)
	}
	if ledger[0].Kind != "operator_ack.slack" || ledger[0].Status != "delivered" {
		t.Fatalf("unexpected ACK ledger event: %#v", ledger[0])
	}
}

func TestSlackSurfacePostsOperatorACKAfterDurableIngress(t *testing.T) {
	store := storepkg.NewMemoryStore()
	runtime := newSlackSurfaceRuntime(config.Config{
		ServiceName:                   "rsi-slack-surface",
		SlackAppIdentity:              "rsi",
		PublicBaseURL:                 "https://staging-rsi-platform.storyprotocol.net",
		SlackIngressAllowedChannelIDs: []string{"C123"},
	}, store)
	poster := &fakeSlackPoster{}
	runtime.slackAPI = poster

	runtime.handleEventsAPIEvent(context.Background(), slackevents.EventsAPIEvent{
		Type:   slackevents.CallbackEvent,
		TeamID: "T123",
		InnerEvent: slackevents.EventsAPIInnerEvent{
			Type: "app_mention",
			Data: &slackevents.AppMentionEvent{
				Channel:   "C123",
				User:      "U123",
				Text:      "<@U_RSI> investigate this",
				TimeStamp: "171000001.000100",
			},
		},
	})

	deadline := time.Now().Add(2 * time.Second)
	for len(poster.calls) == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if len(poster.calls) != 1 {
		t.Fatalf("expected Slack ingress to post exactly one trace ACK after durable ingress, got %d calls", len(poster.calls))
	}
	call := poster.calls[0]
	if call.channelID != "C123" || call.values.Get("thread_ts") != "171000001.000100" {
		t.Fatalf("ACK posted to wrong target: channel=%s values=%s", call.channelID, call.values.Encode())
	}
}

func TestSlackSurfaceDBReadDenyDoesNotSetApprovalFields(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	request, _, err := store.UpsertDBReadRequest(storepkg.DBReadCreateInput{
		IdempotencyKey: "db-read-deny-test",
		Target:         "depin-stage",
		Purpose:        "query",
		SQL:            "select 1",
		SQLSHA256:      "sha256:abc",
		Requester:      "hermes",
		ChannelID:      "C123",
		ThreadTS:       "171000001.000100",
		ExpiresAt:      now.Add(time.Hour),
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	attempt := storepkg.NewDBReadValidationAttempt(request, storepkg.DBReadValidationStatusSucceeded, "target_prepare", "", nil, now)
	if _, err := store.AppendDBReadValidationAttempt(attempt); err != nil {
		t.Fatal(err)
	}
	request, err = store.TransitionDBReadRequest(request.ID, storepkg.DBReadStatePendingApproval, storepkg.DBReadStatePendingApproval, func(item *storepkg.DBReadRequest) error {
		item.SlackMessageChannelID = "C123"
		item.SlackMessageTS = "171000001.000200"
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	runtime := newSlackSurfaceRuntime(config.Config{
		SlackAppIdentity:           "rsi",
		DBReadApproverSlackUserIDs: []string{"UADMIN"},
	}, store)
	runtime.slackAPI = &fakeSlackPoster{}

	runtime.handleInteractiveCallback(context.Background(), slack.InteractionCallback{
		Type: slack.InteractionTypeBlockActions,
		User: slack.User{ID: "UADMIN"},
		ActionCallback: slack.ActionCallbacks{BlockActions: []*slack.BlockAction{{
			ActionID: dbReadSlackDenyAction,
			Value:    request.ID,
		}}},
	})

	updated, ok := store.GetDBReadRequest(request.ID)
	if !ok {
		t.Fatal("expected DB read request")
	}
	if updated.State != storepkg.DBReadStateDenied {
		t.Fatalf("expected denied state, got %s", updated.State)
	}
	if updated.ApprovedBySlackUserID != "" || updated.ApprovedAt != nil {
		t.Fatalf("deny should not set approval fields: approved_by=%q approved_at=%v", updated.ApprovedBySlackUserID, updated.ApprovedAt)
	}
}

func TestSlackSurfaceBuildMentionEnvelopeAllowsMentionOnlySentinel(t *testing.T) {
	runtime := newSlackSurfaceRuntime(config.Config{
		SlackAppIdentity:              "rsi",
		SlackIngressAllowedChannelIDs: []string{slackMentionsOnlySentinel},
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
		SlackAppIdentity:              "rsi",
		SlackIngressAllowedChannelIDs: []string{"C123"},
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

func TestSlackSurfaceMirrorsAmbientAllowlistedMessagesWithoutIngress(t *testing.T) {
	store := storepkg.NewMemoryStore()
	honcho := &fakeSurfaceHonchoCorpus{}
	runtime := newSlackSurfaceRuntime(config.Config{
		SlackAppIdentity:            "rsi",
		SlackMirrorEnabled:          true,
		SlackMirrorChannelAllowlist: []string{"C123"},
		HonchoWorkspaceID:           "rsi_company_knowledge",
		HonchoBaseURL:               "http://honcho.test",
		Environment:                 "stage",
	}, store)
	runtime.mirror = companyknowledge.NewSlackMirror(store, honcho, companyknowledge.SlackMirrorOptions{
		Environment:     "stage",
		HonchoWorkspace: "rsi_company_knowledge",
	})
	before := len(store.ListIngestions())

	runtime.mirrorSlackMessageEvent("T123", "Ev123", &slackevents.MessageEvent{
		Channel:        "C123",
		User:           "U123",
		Text:           "ambient channel message relevant to later RSI work",
		TimeStamp:      "171000001.000100",
		EventTimeStamp: "171000001.000101",
	})

	sourceKey := companyknowledge.SlackMessageSourceKey("T123", "C123", "171000001.000100")
	deadline := time.Now().Add(2 * time.Second)
	var record storepkg.SourceMirrorRecord
	var found bool
	for time.Now().Before(deadline) {
		var err error
		record, found, err = store.GetSourceMirrorRecord(companyknowledge.SlackMessageSourceType, sourceKey)
		if err != nil {
			t.Fatalf("GetSourceMirrorRecord() error = %v", err)
		}
		if found && record.Status == storepkg.SourceMirrorStatusComplete {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !found || record.Status != storepkg.SourceMirrorStatusComplete {
		t.Fatalf("expected completed source mirror record, found=%v record=%#v", found, record)
	}
	if got := len(store.ListIngestions()); got != before {
		t.Fatalf("ambient mirror should not create RSI ingress, before=%d after=%d", before, got)
	}
	if honcho.createCalls != 1 {
		t.Fatalf("CreateMessages calls = %d, want 1", honcho.createCalls)
	}
	if record.Metadata["event_id"] != "Ev123" {
		t.Fatalf("metadata event_id = %#v, want Ev123", record.Metadata["event_id"])
	}
}

func TestSlackSurfaceAcceptsDirectMessages(t *testing.T) {
	store := storepkg.NewMemoryStore()
	runtime := newSlackSurfaceRuntime(config.Config{
		SlackAppIdentity:              "rsi",
		SlackIngressAllowedChannelIDs: []string{"C123"},
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
			if ingestion.ThreadTS != "" {
				t.Fatalf("top-level DM should not be forced into a thread, got thread_ts=%q", ingestion.ThreadTS)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected DM ingestion to be present")
	}
}

func TestSlackSurfaceIgnoresLowSignalDirectMessages(t *testing.T) {
	store := storepkg.NewMemoryStore()
	runtime := newSlackSurfaceRuntime(config.Config{SlackAppIdentity: "rsi"}, store)
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
				Text:        "\U0001F44D",
				TimeStamp:   "171000002.000100",
			},
		},
	})

	if got := len(store.ListIngestions()); got != before {
		t.Fatalf("expected no ingestion for low-signal DM, before=%d after=%d", before, got)
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
