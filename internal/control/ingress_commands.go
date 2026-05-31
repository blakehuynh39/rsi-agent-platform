package control

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func submitIngressEventCommand(cfg config.Config, store storepkg.Store, event ingestion.EventEnvelope, actor string, occurredAt time.Time, commandID string) (transition.CommandReceipt, error) {
	aggregateID := ingressAggregateID(string(event.Source), firstNonEmpty(event.DedupeKey, event.SourceEventID, event.ID))
	payload := map[string]any{
		"event_id":                     event.ID,
		"source":                       string(event.Source),
		"source_event_id":              event.SourceEventID,
		"thread_key":                   event.ThreadKey,
		"incident_key":                 event.IncidentKey,
		"dedupe_key":                   event.DedupeKey,
		"severity":                     string(event.Severity),
		"normalized_problem_statement": event.NormalizedProblemStatement,
		"ownership_hint":               event.OwnershipHint,
		"raw_payload_ref":              event.RawPayloadRef,
		"workflow_hint":                event.WorkflowHint,
		"metadata":                     event.Metadata,
		"created_at":                   event.CreatedAt,
	}
	mergeIngressRuntimePayload(payload, cfg)
	return submitIngressCommand(store, aggregateID, transition.CommandIngressRecordEvent, actor, occurredAt, commandID, payload)
}

func submitIngressSlackCommand(ctx context.Context, cfg config.Config, store storepkg.Store, envelope slack.SlackEnvelope, actor string, occurredAt time.Time, commandID string) (transition.CommandReceipt, error) {
	select {
	case <-ctx.Done():
		return transition.CommandReceipt{}, ctx.Err()
	default:
	}
	aggregateID := ingressAggregateID("slack", firstNonEmpty(envelope.TS, envelope.ThreadTS, envelope.ChannelID))
	payload := map[string]any{
		"bot_role":        string(envelope.BotRole),
		"team_id":         envelope.TeamID,
		"channel_id":      envelope.ChannelID,
		"thread_ts":       envelope.ThreadTS,
		"action_token":    envelope.ActionToken,
		"user_id":         envelope.UserID,
		"text":            envelope.Text,
		"ts":              envelope.TS,
		"files":           envelope.Files,
		"entity_refs":     envelope.EntityRefs,
		"prompt_envelope": envelope.Prompt,
		"created_at":      envelope.CreatedAt,
	}
	mergeIngressRuntimePayload(payload, cfg)
	receipt, err := submitIngressCommandContext(ctx, store, aggregateID, transition.CommandIngressRecordSlack, actor, occurredAt, commandID, payload)
	return receipt, err
}

func submitIngressCommand(store storepkg.Store, aggregateID string, kind transition.IngressCommandKind, actor string, occurredAt time.Time, commandID string, payload map[string]any) (transition.CommandReceipt, error) {
	return submitIngressCommandContext(context.Background(), store, aggregateID, kind, actor, occurredAt, commandID, payload)
}

func submitIngressCommandContext(ctx context.Context, store storepkg.Store, aggregateID string, kind transition.IngressCommandKind, actor string, occurredAt time.Time, commandID string, payload map[string]any) (transition.CommandReceipt, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	receipt, err := store.SubmitCommandContext(ctx, transition.CommandEnvelope{
		MachineKind: transition.MachineIngress,
		AggregateID: strings.TrimSpace(aggregateID),
		CommandKind: string(kind),
		CommandID:   strings.TrimSpace(commandID),
		Actor:       strings.TrimSpace(actor),
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
	if err != nil {
		return transition.CommandReceipt{}, err
	}
	if receipt.DecisionKind == transition.DecisionReject {
		return transition.CommandReceipt{}, errors.New(receipt.Reason)
	}
	return receipt, nil
}

func loadIngressEventFromReceipt(store storepkg.Store, receipt transition.CommandReceipt) (ingestion.EventEnvelope, error) {
	eventID := strings.TrimSpace(receipt.ResultRef)
	if eventID == "" {
		return ingestion.EventEnvelope{}, errors.New("missing ingress event result ref")
	}
	for _, item := range store.ListEvents() {
		if item.ID == eventID {
			return item, nil
		}
	}
	return ingestion.EventEnvelope{}, fmt.Errorf("event %s not found", eventID)
}

func loadSlackIngestionFromReceipt(store storepkg.Store, receipt transition.CommandReceipt) (slack.Ingestion, error) {
	ingestionID := strings.TrimSpace(receipt.ResultRef)
	if ingestionID == "" {
		return slack.Ingestion{}, errors.New("missing ingress ingestion result ref")
	}
	for _, item := range store.ListIngestions() {
		if item.ID == ingestionID {
			return item, nil
		}
	}
	return slack.Ingestion{}, fmt.Errorf("ingestion %s not found", ingestionID)
}

func ingressAggregateID(source string, key string) string {
	source = strings.TrimSpace(source)
	key = strings.TrimSpace(key)
	if source == "" {
		source = "system"
	}
	if key == "" {
		key = "unknown"
	}
	return fmt.Sprintf("%s:%s", source, key)
}

func mergeIngressRuntimePayload(payload map[string]any, cfg config.Config) {
	payload["default_repo"] = cfg.DefaultRepo
	payload["allowed_target_repos"] = append([]string(nil), cfg.AllowedTargetRepos...)
	payload["knowledge_base_url"] = cfg.DefaultKnowledgeBaseURL
	payload["sandbox_namespace"] = cfg.SandboxNamespace
}
