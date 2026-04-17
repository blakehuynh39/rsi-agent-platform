package control

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/sync/errgroup"

	"github.com/piplabs/rsi-agent-platform/internal/app"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	slackpkg "github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
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

	runtime := newSlackSurfaceRuntime(cfg, store)
	var group errgroup.Group
	group.Go(func() error {
		log.Printf("starting %s slack-surface identity=%s on :%d", cfg.ServiceName, cfg.SlackAppIdentity, cfg.HTTPPort)
		return app.ListenAndServe(cfg, newSlackSurfaceRouter(cfg))
	})
	group.Go(runtime.run)
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
		allowedChan: allowed,
	}
}

func (s *slackSurfaceRuntime) run() error {
	go s.client.Run()
	for evt := range s.client.Events {
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
			s.client.Ack(*evt.Request)
			s.handleEventsAPIEvent(eventsAPIEvent)
		}
	}
	return errors.New("slack socket mode event channel closed")
}

func (s *slackSurfaceRuntime) handleEventsAPIEvent(eventsAPIEvent slackevents.EventsAPIEvent) {
	if eventsAPIEvent.Type != slackevents.CallbackEvent {
		return
	}
	switch event := eventsAPIEvent.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		if event == nil {
			return
		}
		envelope, ok := s.buildMentionEnvelope(eventsAPIEvent.TeamID, event)
		if !ok {
			return
		}
		createdAt := envelope.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		receipt, err := submitIngressSlackCommand(
			s.cfg,
			s.store,
			envelope,
			s.cfg.ServiceName,
			createdAt,
			"cmd-ingress:slack:"+ingressAggregateID("slack", firstNonEmpty(envelope.TS, envelope.ThreadTS, envelope.ChannelID)),
		)
		if err != nil {
			log.Printf("slack-surface identity=%s ingestion error=%v", s.cfg.SlackAppIdentity, err)
			return
		}
		created, err := loadSlackIngestionFromReceipt(s.store, receipt)
		if err != nil {
			log.Printf("slack-surface identity=%s ingestion load error=%v", s.cfg.SlackAppIdentity, err)
			return
		}
		log.Printf("slack-surface identity=%s ingested thread=%s ingestion=%s", s.cfg.SlackAppIdentity, created.ThreadKey, created.ID)
	case *slackevents.MessageEvent:
		if event == nil {
			return
		}
		envelope, ok := s.buildDirectMessageEnvelope(eventsAPIEvent.TeamID, event)
		if !ok {
			return
		}
		createdAt := envelope.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		receipt, err := submitIngressSlackCommand(
			s.cfg,
			s.store,
			envelope,
			s.cfg.ServiceName,
			createdAt,
			"cmd-ingress:slack:"+ingressAggregateID("slack", firstNonEmpty(envelope.TS, envelope.ThreadTS, envelope.ChannelID)),
		)
		if err != nil {
			log.Printf("slack-surface identity=%s ingestion error=%v", s.cfg.SlackAppIdentity, err)
			return
		}
		created, err := loadSlackIngestionFromReceipt(s.store, receipt)
		if err != nil {
			log.Printf("slack-surface identity=%s ingestion load error=%v", s.cfg.SlackAppIdentity, err)
			return
		}
		log.Printf("slack-surface identity=%s ingested thread=%s ingestion=%s", s.cfg.SlackAppIdentity, created.ThreadKey, created.ID)
	default:
		return
	}
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
		CreatedAt:   parseSlackTimestamp(event.TimeStamp),
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
		CreatedAt:   parseSlackTimestamp(event.TimeStamp),
	}, true
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
