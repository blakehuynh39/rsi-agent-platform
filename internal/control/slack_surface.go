package control

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/sync/errgroup"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/debuglog"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

const slackMentionsOnlySentinel = "MENTIONS_ONLY"

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
		return errors.New("slack-surface mode requires RSI_SLACK_BOT_TOKEN")
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
			"service":                   cfg.ServiceName,
			"mode":                      "slack-surface",
			"slack_app_identity":        cfg.SlackAppIdentity,
			"socket_mode_enabled":       cfg.SlackSocketModeEnabled,
			"allowed_slack_channel_ids": cfg.AllowedSlackChannelIDs,
		})
	})
	r.Get("/api/slack-surface/channels", func(w http.ResponseWriter, r *http.Request) {
		app.WriteJSON(w, http.StatusOK, map[string]any{"allowed_channel_ids": cfg.AllowedSlackChannelIDs})
	})
	return r
}

type slackSurfaceRuntime struct {
	cfg         config.Config
	store       storepkg.Store
	client      *socketmode.Client
	resolver    slackpkg.EntityResolver
	allowedChan map[string]struct{}
}

func newSlackSurfaceRuntime(cfg config.Config, store storepkg.Store) *slackSurfaceRuntime {
	api := slack.New(cfg.SlackBotToken, slack.OptionAppLevelToken(cfg.SlackAppToken))
	allowed := make(map[string]struct{}, len(cfg.AllowedSlackChannelIDs))
	for _, channelID := range cfg.AllowedSlackChannelIDs {
		channelID = strings.TrimSpace(channelID)
		if channelID != "" {
			allowed[channelID] = struct{}{}
		}
	}
	return &slackSurfaceRuntime{
		cfg:         cfg,
		store:       store,
		client:      socketmode.New(api),
		resolver:    slackpkg.NewEntityResolver(firstNonEmpty(cfg.SlackUserToken, cfg.SlackBotToken)),
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
			}
		}
	}
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
	switch event := eventsAPIEvent.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		if event == nil {
			return true
		}
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
		receipt, err := s.submitIngressSlackCommandWithAckBudget(ctx, envelope, createdAt)
		if err != nil {
			log.Printf("slack-surface identity=%s ingestion error=%v", s.cfg.SlackAppIdentity, err)
			return false
		}
		created, err := loadSlackIngestionFromReceipt(s.store, receipt)
		if err != nil {
			log.Printf("slack-surface identity=%s ingestion load error=%v", s.cfg.SlackAppIdentity, err)
		} else {
			log.Printf("slack-surface identity=%s ingested thread=%s ingestion=%s", s.cfg.SlackAppIdentity, created.ThreadKey, created.ID)
		}
		return true
	case *slackevents.MessageEvent:
		if event == nil {
			return true
		}
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
		receipt, err := s.submitIngressSlackCommandWithAckBudget(ctx, envelope, createdAt)
		if err != nil {
			log.Printf("slack-surface identity=%s ingestion error=%v", s.cfg.SlackAppIdentity, err)
			return false
		}
		created, err := loadSlackIngestionFromReceipt(s.store, receipt)
		if err != nil {
			log.Printf("slack-surface identity=%s ingestion load error=%v", s.cfg.SlackAppIdentity, err)
		} else {
			log.Printf("slack-surface identity=%s ingested thread=%s ingestion=%s", s.cfg.SlackAppIdentity, created.ThreadKey, created.ID)
		}
		return true
	default:
		return true
	}
}

func (s *slackSurfaceRuntime) submitIngressSlackCommandWithAckBudget(parentCtx context.Context, envelope slackpkg.SlackEnvelope, createdAt time.Time) (transition.CommandReceipt, error) {
	ctx, cancel, _ := s.ingressContext(parentCtx)
	defer cancel()
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
		return transition.CommandReceipt{}, err
	}
	return receipt, nil
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
		log.Printf("slack-surface identity=%s ignored mention channel=%s allowed=%v", s.cfg.SlackAppIdentity, event.Channel, s.cfg.AllowedSlackChannelIDs)
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
