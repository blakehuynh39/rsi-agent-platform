package control

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/sync/errgroup"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/companyknowledge"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/debuglog"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

const slackMentionsOnlySentinel = "MENTIONS_ONLY"

type slackMessagePoster interface {
	PostMessageContext(ctx context.Context, channelID string, options ...slack.MsgOption) (string, string, error)
	UpdateMessageContext(ctx context.Context, channelID, timestamp string, options ...slack.MsgOption) (string, string, string, error)
}

func RunSlackSurface(cfg config.Config, store storepkg.Store) error {
	if !cfg.SlackSocketModeEnabled {
		return errors.New("slack-surface mode requires RSI_SLACK_SOCKET_MODE_ENABLED=true")
	}
	if strings.TrimSpace(cfg.SlackAppIdentity) == "" {
		return errors.New("slack-surface mode requires RSI_SLACK_APP_IDENTITY")
	}
	if strings.TrimSpace(cfg.SlackAppToken) == "" {
		return errors.New("slack-surface mode requires RSI_SLACK_APP_TOKEN")
	}
	if strings.TrimSpace(cfg.SlackBotToken) == "" {
		return errors.New("slack-surface mode requires SLACK_BOT_TOKEN")
	}
	if cfg.SlackMirrorEnabled {
		if _, ok := store.(storepkg.SourceMirrorWriteStore); !ok {
			return errors.New("slack-surface mirror requires a source-mirror-capable store")
		}
	}

	if cfg.DrainEnabled {
		app.InstallSignalDrain()
	}
	runtime := newSlackSurfaceRuntime(cfg, store)
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	group, ctx := errgroup.WithContext(rootCtx)
	group.Go(func() error {
		log.Printf("starting %s slack-surface identity=%s on :%d", cfg.ServiceName, cfg.SlackAppIdentity, cfg.HTTPPort)
		err := app.ListenAndServe(cfg, newSlackSurfaceRouter(cfg))
		cancel()
		return err
	})
	group.Go(func() error {
		err := runtime.run(ctx)
		if err != nil {
			app.StartDrain()
			cancel()
		}
		return err
	})
	return group.Wait()
}

func newSlackSurfaceRouter(cfg config.Config) http.Handler {
	r := app.NewBaseRouter(cfg)
	r.Get("/api/slack-surface", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{
			"service":                           cfg.ServiceName,
			"mode":                              "slack-surface",
			"slack_app_identity":                cfg.SlackAppIdentity,
			"socket_mode_enabled":               cfg.SlackSocketModeEnabled,
			"slack_ingress_allowed_channel_ids": cfg.SlackIngressAllowedChannelIDs,
		})
	})
	r.Get("/api/slack-surface/channels", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{"ingress_allowed_channel_ids": cfg.SlackIngressAllowedChannelIDs})
	})
	return r
}

type slackSurfaceRuntime struct {
	cfg         config.Config
	store       storepkg.Store
	client      *socketmode.Client
	slackAPI    slackMessagePoster
	resolver    slackpkg.EntityResolver
	mirror      *companyknowledge.SlackMirror
	allowedChan map[string]struct{}
}

func newSlackSurfaceRuntime(cfg config.Config, store storepkg.Store) *slackSurfaceRuntime {
	api := slack.New(cfg.SlackBotToken, slack.OptionAppLevelToken(cfg.SlackAppToken))
	allowed := make(map[string]struct{}, len(cfg.SlackIngressAllowedChannelIDs))
	for _, channelID := range cfg.SlackIngressAllowedChannelIDs {
		channelID = strings.TrimSpace(channelID)
		if channelID != "" {
			allowed[channelID] = struct{}{}
		}
	}
	var mirror *companyknowledge.SlackMirror
	if cfg.SlackMirrorEnabled {
		mirrorStore, ok := store.(storepkg.SourceMirrorWriteStore)
		if !ok {
			panic("unreachable: slack mirror enabled but store does not support source mirror idempotency")
		}
		mirror = companyknowledge.NewSlackMirror(
			mirrorStore,
			clients.NewHonchoClientWithAPIKey(cfg.HonchoBaseURL, cfg.HonchoAPIKey),
			companyknowledge.SlackMirrorOptions{
				Environment:     cfg.Environment,
				HonchoWorkspace: cfg.HonchoWorkspaceID,
			},
		)
	}
	return &slackSurfaceRuntime{
		cfg:         cfg,
		store:       store,
		client:      socketmode.New(api),
		slackAPI:    api,
		resolver:    slackpkg.NewEntityResolver(cfg.SlackBotToken),
		mirror:      mirror,
		allowedChan: allowed,
	}
}

func (s *slackSurfaceRuntime) run(parentCtx context.Context) error {
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()
	go s.watchDrain(ctx, cancel)
	go s.client.RunContext(ctx)
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt, ok := <-s.client.Events:
			if !ok {
				if ctx.Err() != nil || app.IsDraining() {
					return nil
				}
				return errors.New("slack socket mode event channel closed")
			}
			if ctx.Err() != nil || app.IsDraining() {
				return nil
			}
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				log.Printf("slack-surface identity=%s connecting", s.cfg.SlackAppIdentity)
			case socketmode.EventTypeConnected:
				log.Printf("slack-surface identity=%s connected", s.cfg.SlackAppIdentity)
			case socketmode.EventTypeConnectionError:
				log.Printf("slack-surface identity=%s connection error: %v", s.cfg.SlackAppIdentity, evt.Data)
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					s.client.Ack(*evt.Request)
					continue
				}
				if s.cfg.SlackAckAfterDurableIngress {
					if s.handleEventsAPIEvent(ctx, eventsAPIEvent) {
						s.client.Ack(*evt.Request)
					}
					continue
				}
				s.client.Ack(*evt.Request)
				s.handleEventsAPIEvent(ctx, eventsAPIEvent)
			case socketmode.EventTypeInteractive:
				s.client.Ack(*evt.Request)
				callback, ok := evt.Data.(slack.InteractionCallback)
				if !ok {
					log.Printf("slack-surface identity=%s unsupported interactive payload=%T", s.cfg.SlackAppIdentity, evt.Data)
					continue
				}
				go s.handleInteractiveCallback(context.Background(), callback)
			}
		}
	}
}

func (s *slackSurfaceRuntime) handleInteractiveCallback(ctx context.Context, callback slack.InteractionCallback) {
	if callback.Type != slack.InteractionTypeBlockActions {
		return
	}
	if len(callback.ActionCallback.BlockActions) == 0 {
		return
	}
	action := callback.ActionCallback.BlockActions[0]
	if action == nil {
		return
	}
	if action.ActionID != dbReadSlackApproveAction && action.ActionID != dbReadSlackDenyAction {
		return
	}
	requestID := strings.TrimSpace(action.Value)
	request, ok := s.store.GetDBReadRequest(requestID)
	if !ok {
		log.Printf("slack-surface identity=%s db-read action=%s unknown request=%s", s.cfg.SlackAppIdentity, action.ActionID, requestID)
		return
	}
	userID := strings.TrimSpace(callback.User.ID)
	if !dbReadApproverAllowed(s.cfg, userID) {
		log.Printf("slack-surface identity=%s db-read action=%s unauthorized user=%s request=%s", s.cfg.SlackAppIdentity, action.ActionID, userID, requestID)
		_, _, _ = s.slackAPI.PostMessageContext(
			ctx,
			firstNonEmpty(request.SlackMessageChannelID, request.ChannelID),
			slack.MsgOptionText("You are not authorized to approve RSI DB read requests.", false),
			slack.MsgOptionTS(firstNonEmpty(request.SlackMessageTS, request.ThreadTS)),
		)
		return
	}
	now := time.Now().UTC()
	if request.State != storepkg.DBReadStatePendingApproval {
		_ = updateDBReadSlackCard(ctx, s.slackAPI, request, "stale action ignored; current state is `"+string(request.State)+"`")
		return
	}
	if request.ExpiresAt.Before(now) {
		updated, err := s.store.TransitionDBReadRequest(request.ID, storepkg.DBReadStatePendingApproval, storepkg.DBReadStateExpired, func(item *storepkg.DBReadRequest) error {
			item.ErrorMessage = "approval expired before Slack action"
			return nil
		})
		if err == nil {
			_ = updateDBReadSlackCard(ctx, s.slackAPI, updated, "expired")
			markDBReadExternalToolOutcome(s.cfg, s.store, updated, storepkg.ExternalToolOutcomeExpired, "approval expired before Slack action")
		}
		return
	}
	switch action.ActionID {
	case dbReadSlackApproveAction:
		updated, err := s.store.TransitionDBReadRequest(request.ID, storepkg.DBReadStatePendingApproval, storepkg.DBReadStateApproved, func(item *storepkg.DBReadRequest) error {
			item.ApprovedBySlackUserID = userID
			item.ApprovedAt = &now
			return nil
		})
		if err != nil {
			log.Printf("slack-surface identity=%s db-read approve request=%s error=%v", s.cfg.SlackAppIdentity, request.ID, err)
			return
		}
		_ = updateDBReadSlackCard(ctx, s.slackAPI, updated, "approved by `<@"+userID+">`; queued for execution")
		if pause, ok := s.store.GetExternalToolPauseByDBReadRequestID(updated.ID); ok {
			_, _ = s.store.UpdateExternalToolPause(pause.ID, func(item *storepkg.ExternalToolPause) error {
				item.ApprovalStatus = storepkg.ExternalToolApprovalApproved
				item.ApprovalRef = userID
				return nil
			})
		}
	case dbReadSlackDenyAction:
		updated, err := s.store.TransitionDBReadRequest(request.ID, storepkg.DBReadStatePendingApproval, storepkg.DBReadStateDenied, func(item *storepkg.DBReadRequest) error {
			item.ErrorMessage = "denied by Slack approver"
			return nil
		})
		if err != nil {
			log.Printf("slack-surface identity=%s db-read deny request=%s error=%v", s.cfg.SlackAppIdentity, request.ID, err)
			return
		}
		_ = updateDBReadSlackCard(ctx, s.slackAPI, updated, "denied by `<@"+userID+">`")
		markDBReadExternalToolOutcome(s.cfg, s.store, updated, storepkg.ExternalToolOutcomeDenied, "denied by Slack approver")
	}
}

func dbReadApproverAllowed(cfg config.Config, userID string) bool {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return false
	}
	for _, allowed := range cfg.DBReadApproverSlackUserIDs {
		if strings.TrimSpace(allowed) == userID {
			return true
		}
	}
	return false
}

func (s *slackSurfaceRuntime) watchDrain(ctx context.Context, cancel context.CancelFunc) {
	select {
	case <-ctx.Done():
		return
	case <-app.DrainStarted():
		cancel()
		return
	}
}

func (s *slackSurfaceRuntime) handleEventsAPIEvent(ctx context.Context, eventsAPIEvent slackevents.EventsAPIEvent) bool {
	if eventsAPIEvent.Type != slackevents.CallbackEvent {
		return true
	}
	eventID := slackEventsAPIEventID(eventsAPIEvent)
	switch event := eventsAPIEvent.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		if event == nil {
			return true
		}
		s.mirrorSlackAppMentionEvent(eventsAPIEvent.TeamID, eventID, event)
		envelope, ok := s.buildMentionEnvelope(eventsAPIEvent.TeamID, event)
		if !ok {
			return true
		}
		envelope.Prompt = slackpkg.CanonicalizePromptEnvelope(envelope, s.resolver)
		if s.cfg.VerboseTraceLogging {
			log.Printf(
				"slack-surface identity=%s app_mention_envelope=%s",
				s.cfg.SlackAppIdentity,
				debuglog.JSON(envelope, s.cfg.VerboseTraceLogLimit),
			)
		}
		createdAt := envelope.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		created, direct, err := s.ingestSlackEnvelopeWithAckBudget(ctx, envelope, createdAt)
		if err != nil {
			log.Printf("slack-surface identity=%s ingestion error=%v", s.cfg.SlackAppIdentity, err)
			return false
		}
		log.Printf("slack-surface identity=%s ingested thread=%s ingestion=%s", s.cfg.SlackAppIdentity, created.ThreadKey, created.ID)
		go s.postOperatorTraceACK(context.Background(), created)
		if direct {
			go s.startDirectSlackWorkflow(context.Background(), created)
		}
		return true
	case *slackevents.MessageEvent:
		if event == nil {
			return true
		}
		s.mirrorSlackMessageEvent(eventsAPIEvent.TeamID, eventID, event)
		envelope, ok := s.buildDirectMessageEnvelope(eventsAPIEvent.TeamID, event)
		if !ok {
			return true
		}
		envelope.Prompt = slackpkg.CanonicalizePromptEnvelope(envelope, s.resolver)
		if s.cfg.VerboseTraceLogging {
			log.Printf(
				"slack-surface identity=%s dm_envelope=%s",
				s.cfg.SlackAppIdentity,
				debuglog.JSON(envelope, s.cfg.VerboseTraceLogLimit),
			)
		}
		createdAt := envelope.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		created, direct, err := s.ingestSlackEnvelopeWithAckBudget(ctx, envelope, createdAt)
		if err != nil {
			log.Printf("slack-surface identity=%s ingestion error=%v", s.cfg.SlackAppIdentity, err)
			return false
		}
		log.Printf("slack-surface identity=%s ingested thread=%s ingestion=%s", s.cfg.SlackAppIdentity, created.ThreadKey, created.ID)
		go s.postOperatorTraceACK(context.Background(), created)
		if direct {
			go s.startDirectSlackWorkflow(context.Background(), created)
		}
		return true
	default:
		return true
	}
}

func slackEventsAPIEventID(event slackevents.EventsAPIEvent) string {
	switch data := event.Data.(type) {
	case slackevents.EventsAPICallbackEvent:
		return strings.TrimSpace(data.EventID)
	case *slackevents.EventsAPICallbackEvent:
		if data != nil {
			return strings.TrimSpace(data.EventID)
		}
	case map[string]any:
		if value, ok := data["event_id"].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (s *slackSurfaceRuntime) mirrorSlackAppMentionEvent(teamID string, eventID string, event *slackevents.AppMentionEvent) {
	if s == nil || s.mirror == nil || event == nil {
		return
	}
	if !s.slackMirrorChannelAllowed(event.Channel) {
		return
	}
	input := companyknowledge.SlackMessageInput{
		WorkspaceID: firstNonEmpty(event.SourceTeam, event.UserTeam, teamID),
		ChannelID:   event.Channel,
		TS:          event.TimeStamp,
		ThreadTS:    event.ThreadTimeStamp,
		UserID:      event.User,
		BotID:       event.BotID,
		Text:        event.Text,
		EventID:     firstNonEmpty(eventID, event.EventTimeStamp),
		Files:       slackFileMetadata(event.Files),
		CreatedAt:   companyknowledge.SlackTimestampToTime(event.TimeStamp),
	}
	if event.Edited != nil {
		input.EditedTS = event.Edited.TimeStamp
	}
	go s.mirrorSlackInput(input)
}

func (s *slackSurfaceRuntime) mirrorSlackMessageEvent(teamID string, eventID string, event *slackevents.MessageEvent) {
	if s == nil || s.mirror == nil || event == nil {
		return
	}
	if !s.slackMirrorChannelAllowed(event.Channel) {
		return
	}
	msg := slack.Message{}
	if event.Message != nil {
		msg.Msg = *event.Message
	} else {
		msg.Msg = slack.Msg{
			Type:            event.Type,
			Channel:         event.Channel,
			User:            event.User,
			Text:            event.Text,
			Timestamp:       event.TimeStamp,
			ThreadTimestamp: event.ThreadTimeStamp,
			SubType:         event.SubType,
			BotID:           event.BotID,
			Username:        event.Username,
			Permalink:       event.Permalink,
		}
	}
	if strings.TrimSpace(msg.Channel) == "" {
		msg.Channel = event.Channel
	}
	input := slackInputFromMessage(firstNonEmpty(event.SourceTeam, event.UserTeam, teamID), event.Channel, msg, firstNonEmpty(eventID, event.EventTimeStamp))
	go s.mirrorSlackInput(input)
}

func (s *slackSurfaceRuntime) mirrorSlackInput(input companyknowledge.SlackMessageInput) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err := s.mirror.IngestMessage(ctx, input); err != nil {
		log.Printf("slack-surface identity=%s mirror failed channel=%s ts=%s error=%v", s.cfg.SlackAppIdentity, input.ChannelID, input.TS, err)
	}
}

func (s *slackSurfaceRuntime) slackMirrorChannelAllowed(channelID string) bool {
	if s == nil {
		return false
	}
	return slackMirrorChannelAllowedByConfig(s.cfg, channelID)
}

func slackFileMetadata(files []slack.File) []companyknowledge.SlackFileMetadata {
	out := make([]companyknowledge.SlackFileMetadata, 0, len(files))
	for _, file := range files {
		out = append(out, companyknowledge.SlackFileMetadata{
			ID:        file.ID,
			Name:      file.Name,
			Title:     file.Title,
			MimeType:  file.Mimetype,
			FileType:  file.Filetype,
			Size:      file.Size,
			Permalink: file.Permalink,
		})
	}
	return out
}

func (s *slackSurfaceRuntime) postOperatorTraceACK(parentCtx context.Context, ingestion slackpkg.Ingestion) {
	trace, ok := traceSummaryForIngestion(s.store, ingestion)
	if !ok {
		log.Printf("slack-surface identity=%s operator_ack skipped ingestion=%s reason=trace_not_found", s.cfg.SlackAppIdentity, ingestion.ID)
		return
	}
	postOperatorTraceACKForTrace(parentCtx, s.cfg, s.store, s.slackAPI, trace, ingestion)
}

func postOperatorTraceACKForTrace(parentCtx context.Context, cfg config.Config, store storepkg.Store, slackAPI slackMessagePoster, trace events.TraceSummary, ingestion slackpkg.Ingestion) {
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	traceURL := operatorTraceURL(cfg.PublicBaseURL, trace.ConversationID, trace.TraceID)
	if traceURL == "" {
		recordOperatorTraceACK(cfg, store, trace, ingestion, "failed", map[string]any{
			"error": "RSI_PUBLIC_BASE_URL is not configured",
		})
		return
	}
	if slackAPI == nil {
		recordOperatorTraceACK(cfg, store, trace, ingestion, "failed", map[string]any{
			"error":     "Slack API client is not configured",
			"trace_url": traceURL,
		})
		return
	}
	claimID := "cmd-slack-operator-ack:" + strings.TrimSpace(ingestion.ID)
	now := time.Now().UTC()
	_, claimed, err := store.RecordCommandReceipt(transition.CommandReceipt{
		CommandID:        claimID,
		MachineKind:      transition.MachineIngress,
		AggregateID:      ingestion.ID,
		CommandKind:      "slack_operator_ack",
		Actor:            cfg.ServiceName,
		DecisionKind:     transition.DecisionAdvance,
		Reason:           "visible Slack trace ACK claimed",
		AggregateVersion: 1,
		ResultRef:        trace.TraceID,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		log.Printf("slack operator_ack claim error service=%s identity=%s ingestion=%s trace=%s error=%v", cfg.ServiceName, cfg.SlackAppIdentity, ingestion.ID, trace.TraceID, err)
		recordOperatorTraceACK(cfg, store, trace, ingestion, "failed", map[string]any{"error": err.Error(), "trace_url": traceURL})
		return
	}
	if !claimed {
		return
	}

	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancel()
	body := operatorTraceACKBody(traceURL)
	options := []slack.MsgOption{
		slack.MsgOptionText(body, false),
		slack.MsgOptionDisableLinkUnfurl(),
		slack.MsgOptionDisableMediaUnfurl(),
	}
	if threadTS := strings.TrimSpace(ingestion.ThreadTS); threadTS != "" {
		options = append(options, slack.MsgOptionTS(threadTS))
	}
	channelID, messageTS, err := slackAPI.PostMessageContext(ctx, ingestion.ChannelID, options...)
	if err != nil {
		log.Printf("slack operator_ack failed service=%s identity=%s ingestion=%s trace=%s channel=%s thread=%s error=%v", cfg.ServiceName, cfg.SlackAppIdentity, ingestion.ID, trace.TraceID, ingestion.ChannelID, ingestion.ThreadTS, err)
		recordOperatorTraceACK(cfg, store, trace, ingestion, "failed", map[string]any{
			"error":      err.Error(),
			"trace_url":  traceURL,
			"channel_id": ingestion.ChannelID,
			"thread_ts":  ingestion.ThreadTS,
		})
		return
	}
	recordOperatorTraceACK(cfg, store, trace, ingestion, "delivered", map[string]any{
		"channel_id": firstNonEmpty(channelID, ingestion.ChannelID),
		"thread_ts":  ingestion.ThreadTS,
		"message_ts": messageTS,
		"trace_url":  traceURL,
		"body":       body,
	})
}

func recordOperatorTraceACK(cfg config.Config, store storepkg.Store, trace events.TraceSummary, ingestion slackpkg.Ingestion, status string, payload map[string]any) {
	if strings.TrimSpace(trace.TraceID) == "" {
		return
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload["ingestion_id"] = ingestion.ID
	payload["conversation_id"] = trace.ConversationID
	payload["trace_id"] = trace.TraceID
	err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ExecutionID:    "slack-operator-ack:" + firstNonEmpty(ingestion.ID, trace.TraceID),
			TraceID:        trace.TraceID,
			WorkflowID:     trace.WorkflowID,
			PhaseID:        "ingress",
			Kind:           "operator_ack.slack",
			Status:         status,
			Seq:            1,
			IdempotencyKey: "slack-operator-ack:" + firstNonEmpty(ingestion.ID, trace.TraceID),
			Payload:        payload,
			RecordedAt:     time.Now().UTC(),
		},
	})
	if err != nil {
		log.Printf("slack operator_ack ledger error service=%s identity=%s ingestion=%s trace=%s error=%v", cfg.ServiceName, cfg.SlackAppIdentity, ingestion.ID, trace.TraceID, err)
	}
}

func postWorkflowPlatformFailureNotice(cfg config.Config, store storepkg.Store, trace events.TraceSummary, ingestion slackpkg.Ingestion, failureClass string, summary string) {
	if strings.TrimSpace(ingestion.ChannelID) == "" || strings.TrimSpace(ingestion.ID) == "" {
		return
	}
	traceURL := operatorTraceURL(cfg.PublicBaseURL, trace.ConversationID, trace.TraceID)
	if traceURL == "" {
		recordPlatformFailureNotice(cfg, store, trace, ingestion, "failed", map[string]any{
			"error":         "RSI_PUBLIC_BASE_URL is not configured",
			"failure_class": failureClass,
		})
		return
	}
	if strings.TrimSpace(cfg.SlackBotToken) == "" {
		recordPlatformFailureNotice(cfg, store, trace, ingestion, "failed", map[string]any{
			"error":         "SLACK_BOT_TOKEN is not configured",
			"trace_url":     traceURL,
			"failure_class": failureClass,
		})
		return
	}
	claimID := "cmd-slack-platform-failure:" + strings.TrimSpace(trace.TraceID) + ":" + strings.TrimSpace(failureClass)
	now := time.Now().UTC()
	_, claimed, err := store.RecordCommandReceipt(transition.CommandReceipt{
		CommandID:        claimID,
		MachineKind:      transition.MachineIngress,
		AggregateID:      ingestion.ID,
		CommandKind:      "slack_platform_failure_notice",
		Actor:            cfg.ServiceName,
		DecisionKind:     transition.DecisionAdvance,
		Reason:           "visible Slack platform failure notice claimed",
		AggregateVersion: 1,
		ResultRef:        trace.TraceID,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		log.Printf("slack platform_failure_notice claim error service=%s identity=%s ingestion=%s trace=%s error=%v", cfg.ServiceName, cfg.SlackAppIdentity, ingestion.ID, trace.TraceID, err)
		recordPlatformFailureNotice(cfg, store, trace, ingestion, "failed", map[string]any{"error": err.Error(), "trace_url": traceURL, "failure_class": failureClass})
		return
	}
	if !claimed {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	body := platformFailureNoticeBody(traceURL, failureClass, summary)
	options := []slack.MsgOption{
		slack.MsgOptionText(body, false),
		slack.MsgOptionDisableLinkUnfurl(),
		slack.MsgOptionDisableMediaUnfurl(),
	}
	if threadTS := strings.TrimSpace(ingestion.ThreadTS); threadTS != "" {
		options = append(options, slack.MsgOptionTS(threadTS))
	}
	channelID, messageTS, err := slack.New(cfg.SlackBotToken).PostMessageContext(ctx, ingestion.ChannelID, options...)
	if err != nil {
		log.Printf("slack platform_failure_notice failed service=%s identity=%s ingestion=%s trace=%s channel=%s thread=%s error=%v", cfg.ServiceName, cfg.SlackAppIdentity, ingestion.ID, trace.TraceID, ingestion.ChannelID, ingestion.ThreadTS, err)
		recordPlatformFailureNotice(cfg, store, trace, ingestion, "failed", map[string]any{
			"error":         err.Error(),
			"trace_url":     traceURL,
			"failure_class": failureClass,
			"channel_id":    ingestion.ChannelID,
			"thread_ts":     ingestion.ThreadTS,
		})
		return
	}
	recordPlatformFailureNotice(cfg, store, trace, ingestion, "delivered", map[string]any{
		"channel_id":    firstNonEmpty(channelID, ingestion.ChannelID),
		"thread_ts":     ingestion.ThreadTS,
		"message_ts":    messageTS,
		"trace_url":     traceURL,
		"body":          body,
		"failure_class": failureClass,
	})
}

func recordPlatformFailureNotice(cfg config.Config, store storepkg.Store, trace events.TraceSummary, ingestion slackpkg.Ingestion, status string, payload map[string]any) {
	if strings.TrimSpace(trace.TraceID) == "" {
		return
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload["ingestion_id"] = ingestion.ID
	payload["conversation_id"] = trace.ConversationID
	payload["trace_id"] = trace.TraceID
	err := store.RecordExecutionLedgerEvents([]events.ExecutionLedgerEvent{
		{
			ExecutionID:    "slack-platform-failure:" + firstNonEmpty(ingestion.ID, trace.TraceID),
			TraceID:        trace.TraceID,
			WorkflowID:     trace.WorkflowID,
			PhaseID:        "ingress",
			Kind:           "platform_failure.slack",
			Status:         status,
			Seq:            1,
			IdempotencyKey: "slack-platform-failure:" + firstNonEmpty(ingestion.ID, trace.TraceID),
			Payload:        payload,
			RecordedAt:     time.Now().UTC(),
		},
	})
	if err != nil {
		log.Printf("slack platform_failure_notice ledger error service=%s identity=%s ingestion=%s trace=%s error=%v", cfg.ServiceName, cfg.SlackAppIdentity, ingestion.ID, trace.TraceID, err)
	}
}

func platformFailureNoticeBody(traceURL string, failureClass string, summary string) string {
	reason := firstNonEmpty(strings.TrimSpace(summary), strings.TrimSpace(failureClass), "the executor stopped before producing a final reply")
	return fmt.Sprintf("RSI run interrupted before Hermes could finish. This is an operational failure notice, not a Hermes answer. <%s|Open trace>. `%s`", traceURL, reason)
}

func traceSummaryForIngestion(store storepkg.Store, ingestion slackpkg.Ingestion) (events.TraceSummary, bool) {
	ingestionID := strings.TrimSpace(ingestion.ID)
	eventID := strings.TrimSpace(ingestion.EventID)
	conversationID := strings.TrimSpace(ingestion.ConversationID)
	caseID := strings.TrimSpace(ingestion.CaseID)

	for _, trace := range store.ListTraces() {
		if strings.TrimSpace(trace.IngestionID) == ingestionID {
			return trace, true
		}
		if eventID != "" && strings.TrimSpace(trace.TriggerEventID) == eventID {
			return trace, true
		}
		if conversationID != "" && strings.TrimSpace(trace.ConversationID) == conversationID && strings.TrimSpace(trace.CaseID) == caseID {
			return trace, true
		}
	}
	return events.TraceSummary{}, false
}

func operatorTraceURL(publicBaseURL string, conversationID string, traceID string) string {
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base == "" || strings.TrimSpace(conversationID) == "" || strings.TrimSpace(traceID) == "" {
		return ""
	}
	query := url.Values{}
	query.Set("tab", "conversations")
	query.Set("conversation", strings.TrimSpace(conversationID))
	query.Set("trace", strings.TrimSpace(traceID))
	return base + "/sessions?" + query.Encode()
}

func operatorTraceACKBody(traceURL string) string {
	if strings.TrimSpace(traceURL) == "" {
		return "Tracking this RSI run."
	}
	return fmt.Sprintf("Tracking this RSI run: <%s|open trace>", traceURL)
}

func (s *slackSurfaceRuntime) ingestSlackEnvelopeWithAckBudget(parentCtx context.Context, envelope slackpkg.SlackEnvelope, createdAt time.Time) (slackpkg.Ingestion, bool, error) {
	ctx, cancel, _ := s.ingressContext(parentCtx)
	defer cancel()
	if direct, ok := s.store.(storepkg.DirectSlackIngressStore); ok {
		created, err := direct.CreateSlackIngestionDirect(ctx, envelope)
		return created, true, err
	}
	receipt, err := submitIngressSlackCommand(
		ctx,
		s.cfg,
		s.store,
		envelope,
		s.cfg.ServiceName,
		createdAt,
		"cmd-ingress:slack:"+ingressAggregateID("slack", firstNonEmpty(envelope.TS, envelope.ThreadTS, envelope.ChannelID)),
	)
	if err != nil {
		return slackpkg.Ingestion{}, false, err
	}
	created, err := loadSlackIngestionFromReceipt(s.store, receipt)
	if err != nil {
		log.Printf("slack-surface identity=%s ingestion read-back failed after successful write ingestion=%s error=%v", s.cfg.SlackAppIdentity, receipt.ResultRef, err)
		return slackpkg.Ingestion{ID: receipt.ResultRef}, false, nil
	}
	return created, false, nil
}

func (s *slackSurfaceRuntime) startDirectSlackWorkflow(parentCtx context.Context, ingestion slackpkg.Ingestion) {
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	select {
	case <-parentCtx.Done():
		return
	default:
	}
	workflow, ok := workflowForIngestion(s.store, ingestion)
	if !ok {
		log.Printf("slack-surface identity=%s direct_ingress workflow_start skipped ingestion=%s reason=workflow_not_found", s.cfg.SlackAppIdentity, ingestion.ID)
		return
	}
	if strings.TrimSpace(workflow.Status) != "queued" {
		return
	}
	startedAt := ingestion.CreatedAt
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	if err := startWorkflowViaCommand(s.cfg, s.store, workflow.ID, startedAt, queue.WorkflowQueue); err != nil {
		log.Printf("slack-surface identity=%s direct_ingress workflow_start error ingestion=%s workflow=%s error=%v", s.cfg.SlackAppIdentity, ingestion.ID, workflow.ID, err)
	}
}

func workflowForIngestion(store storepkg.Store, ingestion slackpkg.Ingestion) (storepkg.Workflow, bool) {
	ingestionID := strings.TrimSpace(ingestion.ID)
	if ingestionID == "" {
		return storepkg.Workflow{}, false
	}
	if getter, ok := store.(interface {
		GetWorkflowByIngestionID(string) (storepkg.Workflow, bool)
	}); ok {
		return getter.GetWorkflowByIngestionID(ingestionID)
	}
	for _, workflow := range store.ListWorkflows() {
		if strings.TrimSpace(workflow.IngestionID) == ingestionID {
			return workflow, true
		}
	}
	return storepkg.Workflow{}, false
}

func (s *slackSurfaceRuntime) ingressContext(parentCtx context.Context) (context.Context, context.CancelFunc, time.Duration) {
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	if !s.cfg.SlackAckAfterDurableIngress {
		return context.Background(), func() {}, 0
	}
	timeout := s.cfg.SlackDurableIngressAckTimeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	return ctx, cancel, timeout
}

func (s *slackSurfaceRuntime) buildMentionEnvelope(teamID string, event *slackevents.AppMentionEvent) (slackpkg.SlackEnvelope, bool) {
	if event == nil {
		return slackpkg.SlackEnvelope{}, false
	}
	if event.BotID != "" {
		return slackpkg.SlackEnvelope{}, false
	}
	if !s.mentionChannelAllowed(event.Channel) {
		log.Printf("slack-surface identity=%s ignored mention channel=%s allowed=%v", s.cfg.SlackAppIdentity, event.Channel, s.cfg.SlackIngressAllowedChannelIDs)
		return slackpkg.SlackEnvelope{}, false
	}
	threadTS := strings.TrimSpace(event.ThreadTimeStamp)
	if threadTS == "" {
		threadTS = strings.TrimSpace(event.TimeStamp)
	}
	return slackpkg.SlackEnvelope{
		BotRole:     slackRoleFromIdentity(s.cfg.SlackAppIdentity),
		TeamID:      teamID,
		ChannelID:   event.Channel,
		ThreadTS:    threadTS,
		ActionToken: slackActionToken(event.AssistantThread),
		UserID:      event.User,
		Text:        event.Text,
		TS:          event.TimeStamp,
		EntityRefs:  slackpkg.ExtractEntityRefs(event.Text),
		CreatedAt:   parseSlackTimestamp(event.TimeStamp),
		Prompt:      slackpkg.SlackPromptEnvelope{},
	}, true
}

func (s *slackSurfaceRuntime) buildDirectMessageEnvelope(teamID string, event *slackevents.MessageEvent) (slackpkg.SlackEnvelope, bool) {
	if event == nil {
		return slackpkg.SlackEnvelope{}, false
	}
	if !event.IsIM() {
		return slackpkg.SlackEnvelope{}, false
	}
	if event.SubType != "" || event.BotID != "" {
		return slackpkg.SlackEnvelope{}, false
	}
	if lowSignalDirectMessage(event.Text) {
		log.Printf("slack-surface identity=%s ignored low-signal direct message channel=%s ts=%s", s.cfg.SlackAppIdentity, event.Channel, event.TimeStamp)
		return slackpkg.SlackEnvelope{}, false
	}
	threadTS := strings.TrimSpace(event.ThreadTimeStamp)
	return slackpkg.SlackEnvelope{
		BotRole:     slackRoleFromIdentity(s.cfg.SlackAppIdentity),
		TeamID:      teamID,
		ChannelID:   event.Channel,
		ThreadTS:    threadTS,
		ActionToken: slackActionToken(event.AssistantThread),
		UserID:      event.User,
		Text:        event.Text,
		TS:          event.TimeStamp,
		EntityRefs:  slackpkg.ExtractEntityRefs(event.Text),
		CreatedAt:   parseSlackTimestamp(event.TimeStamp),
		Prompt:      slackpkg.SlackPromptEnvelope{},
	}, true
}

func lowSignalDirectMessage(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return true
	}
	hasLetterOrDigit := false
	for _, r := range trimmed {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			hasLetterOrDigit = true
			break
		}
	}
	return !hasLetterOrDigit
}

func slackActionToken(thread *slackevents.AssistantThreadActionToken) string {
	if thread == nil {
		return ""
	}
	return strings.TrimSpace(thread.ActionToken)
}

func (s *slackSurfaceRuntime) mentionChannelAllowed(channelID string) bool {
	if len(s.allowedChan) == 0 {
		return true
	}
	if _, ok := s.allowedChan[slackMentionsOnlySentinel]; ok {
		return true
	}
	_, ok := s.allowedChan[channelID]
	return ok
}

func slackRoleFromIdentity(identity string) slackpkg.BotRole {
	switch strings.ToLower(strings.TrimSpace(identity)) {
	case string(slackpkg.BotOrchestrator):
		return slackpkg.BotOrchestrator
	case string(slackpkg.BotOnCall):
		return slackpkg.BotOnCall
	case string(slackpkg.BotFR):
		return slackpkg.BotFR
	case string(slackpkg.BotArch):
		return slackpkg.BotArch
	default:
		return slackpkg.BotOrchestrator
	}
}

func parseSlackTimestamp(ts string) time.Time {
	if strings.TrimSpace(ts) == "" {
		return time.Now().UTC()
	}
	value, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return time.Now().UTC()
	}
	seconds := int64(value)
	nanos := int64((value - float64(seconds)) * float64(time.Second))
	return time.Unix(seconds, nanos).UTC()
}

func (s *slackSurfaceRuntime) String() string {
	return fmt.Sprintf("slack-surface[%s]", s.cfg.SlackAppIdentity)
}
