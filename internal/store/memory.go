package store

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/registry"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/slack"
)

const (
	defaultProposalSlotCap = 2
	proposalReviewSLA      = 24 * time.Hour
	proposalPromoterLease  = 5 * time.Minute
)

type Store interface {
	ListEvents() []ingestion.EventEnvelope
	CreateEvent(event ingestion.EventEnvelope) (ingestion.EventEnvelope, error)
	ListIngestions() []slack.Ingestion
	CreateIngestion(envelope slack.SlackEnvelope) slack.Ingestion
	ListWorkflows() []Workflow
	ListAssignments() []Assignment
	ListThreadPolicies() []policy.ThreadPolicy
	ListChannelPolicies() []policy.ChannelPolicy
	SetThreadState(threadKey string, state policy.ThreadState, owner string) (policy.ThreadPolicy, error)
	ListOwnershipRecords() []registry.OwnershipRecord
	ListCapabilities() []registry.CapabilityRecord
	ListTemplates() []registry.WorkflowTemplate
	ListExperiments() []registry.ExperimentRecord
	ListTraces() []events.TraceSummary
	GetTrace(traceID string) (events.Trace, bool)
	AddRating(traceID string, rating review.HumanRating) (review.HumanRating, error)
	AddImprovementNote(traceID string, note review.ImprovementNote) (review.ImprovementNote, error)
	ScheduleReplay(traceID string, requestedBy string) (queue.WorkItem, error)
	ListEvalSuites() []evals.Suite
	ListEvalRuns() []evals.Run
	ListEvalJudgments(evalRunID string) []evals.Judgment
	EvaluateTrace(traceID string, trigger string) (evals.Run, []evals.Judgment, error)
	GetSettings() improvement.Settings
	UpdateSettings(settings improvement.Settings) (improvement.Settings, error)
	ListWorkItems() []queue.WorkItem
	EnqueueWorkItem(item queue.WorkItem) (queue.WorkItem, error)
	ClaimNextWorkItem(queues []queue.QueueName, holder string, lease time.Duration) (queue.WorkItem, bool, error)
	CompleteWorkItem(id string) (queue.WorkItem, error)
	FailWorkItem(id string, lastError string) (queue.WorkItem, error)
	UpdateWorkflowStatus(workflowID string, status string, lastError string) (Workflow, error)
	ApplyTraceUpdate(traceID string, update TraceUpdate) (events.Trace, error)
	ListCandidates() []improvement.Candidate
	ListProposalMemories() []review.ProposalMemory
	GetProposalSlots() ProposalSlotState
	PromoteCandidates(requestedBy string, limit int) (PromotionResult, error)
	RunProposalPromoter(holder string) (PromotionResult, error)
	ListProposals() []review.Proposal
	ReviewProposal(proposalID string, decision review.ProposalReview) (review.Proposal, error)
	UpdateProposalStatus(proposalID string, status review.ProposalStatus) (review.Proposal, error)
	MaterializeApprovedProposal(proposalID string, requestedBy string) (improvement.RepoChangeJob, error)
	ListRepoChangeJobs() []improvement.RepoChangeJob
	UpdateRepoChangeJobStatus(jobID string, status string) (improvement.RepoChangeJob, error)
	ListPRAttempts() []improvement.PRAttempt
	RecordPRAttempt(attempt improvement.PRAttempt) (improvement.PRAttempt, error)
	ListPostMergeReplays() []improvement.PostMergeReplay
	ExecuteTool(name string, input map[string]interface{}) ToolResult
}

type MemoryStore struct {
	mu              sync.RWMutex
	events          []ingestion.EventEnvelope
	ingestions      []slack.Ingestion
	workflows       []Workflow
	assignments     []Assignment
	threadPolicies  map[string]policy.ThreadPolicy
	channelPolicy   []policy.ChannelPolicy
	ownership       []registry.OwnershipRecord
	capabilities    []registry.CapabilityRecord
	templates       []registry.WorkflowTemplate
	experiments     []registry.ExperimentRecord
	traces          map[string]events.Trace
	ratings         map[string][]review.HumanRating
	notes           map[string][]review.ImprovementNote
	evalSuites      []evals.Suite
	evalRuns        map[string]evals.Run
	evalJudgments   map[string][]evals.Judgment
	workItems       map[string]queue.WorkItem
	candidates      map[string]improvement.Candidate
	proposals       map[string]review.Proposal
	proposalMemory  []review.ProposalMemory
	repoChangeJobs  map[string]improvement.RepoChangeJob
	prAttempts      map[string]improvement.PRAttempt
	postMergeReplay map[string]improvement.PostMergeReplay
	cronLeases      map[string]improvement.CronLease
	settings        improvement.Settings
}

func NewMemoryStore() *MemoryStore {
	s := newEmptyMemoryStore()
	s.seedDefaults()
	return s
}

func (s *MemoryStore) seedDefaults() {
	now := time.Now().UTC()
	s.channelPolicy = []policy.ChannelPolicy{
		{
			ChannelID:            "CENG",
			ProactiveEnabled:     true,
			AutoPostAllowed:      true,
			AllowedWorkflowKinds: []string{"incident", "feature-request", "architecture"},
			UpdatedAt:            now.Add(-2 * time.Hour),
		},
	}
	s.ownership = []registry.OwnershipRecord{
		{Domain: "platform", OwnerTeam: "platform", EscalationSlack: "#platform-alerts"},
		{Domain: "depin-backend", OwnerTeam: "platform", EscalationSlack: "#platform-alerts"},
		{Domain: "story-stage", OwnerTeam: "infra", EscalationSlack: "#stage-oncall"},
	}
	s.capabilities = []registry.CapabilityRecord{
		{Name: "sentry.query", Kind: "tool", AllowedBots: []string{"oncall"}, ApprovalNeeded: false},
		{Name: "k8s.logs", Kind: "tool", AllowedBots: []string{"oncall"}, ApprovalNeeded: false},
		{Name: "github.create_issue", Kind: "tool", AllowedBots: []string{"fr"}, ApprovalNeeded: false},
		{Name: "github.create_pr", Kind: "tool", AllowedBots: []string{"oncall", "fr", "arch"}, ApprovalNeeded: true},
		{Name: "repo.answer_question", Kind: "skill", AllowedBots: []string{"arch"}, ApprovalNeeded: false},
	}
	s.templates = []registry.WorkflowTemplate{
		{Name: "incident-oncall", Kind: "incident", Description: "Investigate incidents and propose remediation", Steps: []string{"ingest", "route", "debug", "propose"}},
		{Name: "feature-request", Kind: "feature-request", Description: "Turn asks into grounded FRs", Steps: []string{"ingest", "ground", "summarize", "issue"}},
		{Name: "architecture-question", Kind: "architecture", Description: "Answer repo/architecture questions", Steps: []string{"ingest", "ground", "answer"}},
	}
	s.experiments = []registry.ExperimentRecord{
		{Name: "arch-routing-threshold", Candidate: "v2", Baseline: "v1", State: "review"},
	}
	s.evalSuites = []evals.Suite{
		{Name: "incident-response", Description: "Evaluate incident routing and remediation quality", EventKinds: []string{"incident", "sentry"}, Layers: []evals.Layer{evals.LayerDeterministic, evals.LayerTaskQuality, evals.LayerArchitecture}},
		{Name: "architecture-review", Description: "Evaluate architecture analysis quality", EventKinds: []string{"architecture", "slack"}, Layers: []evals.Layer{evals.LayerDeterministic, evals.LayerTaskQuality, evals.LayerArchitecture}},
		{Name: "proposal-quality", Description: "Evaluate repo-change proposal readiness", EventKinds: []string{"proposal"}, Layers: []evals.Layer{evals.LayerDeterministic, evals.LayerTaskQuality, evals.LayerArchitecture}},
	}

	_, _ = s.CreateEvent(ingestion.EventEnvelope{
		Source:                     ingestion.SourceSlack,
		SourceEventID:              "slack-171000001.000100",
		ThreadKey:                  "slack:CENG:171000001.000100",
		DedupeKey:                  "slack:CENG:171000001.000100",
		Severity:                   ingestion.SeverityError,
		NormalizedProblemStatement: "Investigate why staging homepage is failing and propose a fix.",
		OwnershipHint:              "platform",
		RawPayloadRef:              "s3://rsi-agent-platform-stage-artifacts/payloads/ing-001.json",
		WorkflowHint:               "incident",
		Metadata: map[string]interface{}{
			"channel_id": "CENG",
			"user_id":    "U123",
			"thread_ts":  "171000001.000100",
		},
		CreatedAt: now.Add(-35 * time.Minute),
	})
	_, _ = s.CreateEvent(ingestion.EventEnvelope{
		Source:                     ingestion.SourceSentry,
		SourceEventID:              "sentry-issue-2413",
		IncidentKey:                "sentry:issue-2413",
		DedupeKey:                  "sentry:issue-2413",
		Severity:                   ingestion.SeverityCritical,
		NormalizedProblemStatement: "Repeated improvement-plane failures indicate the platform lacks a durable closed-loop proposal gate.",
		OwnershipHint:              "platform",
		RawPayloadRef:              "s3://rsi-agent-platform-stage-artifacts/payloads/sentry-2413.json",
		WorkflowHint:               "incident",
		Metadata: map[string]interface{}{
			"service": "improvement-plane",
			"alert":   "proposal-slot-drift",
		},
		CreatedAt: now.Add(-20 * time.Minute),
	})
	for _, trace := range s.ListTraces() {
		_, _, _ = s.EvaluateTrace(trace.TraceID, "seed")
	}
	_, _ = s.PromoteCandidates("seed", 1)
}

func (s *MemoryStore) ListEvents() []ingestion.EventEnvelope {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]ingestion.EventEnvelope(nil), s.events...)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) CreateEvent(event ingestion.EventEnvelope) (ingestion.EventEnvelope, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createEventLocked(event)
}

func (s *MemoryStore) createEventLocked(event ingestion.EventEnvelope) (ingestion.EventEnvelope, error) {
	for _, existing := range s.events {
		if existing.Source == event.Source && existing.DedupeKey == event.DedupeKey {
			return existing, nil
		}
	}
	now := time.Now().UTC()
	if event.ID == "" {
		event.ID = nextID("evt", len(s.events)+1)
	}
	if event.Source == "" {
		event.Source = ingestion.SourceSystem
	}
	if event.SourceEventID == "" {
		event.SourceEventID = event.ID
	}
	if event.DedupeKey == "" {
		event.DedupeKey = event.SourceEventID
	}
	if event.Severity == "" {
		event.Severity = ingestion.SeverityWarning
	}
	if event.WorkflowHint == "" {
		event.WorkflowHint = deriveWorkflowHint(event.NormalizedProblemStatement)
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}
	s.events = append(s.events, event)
	s.materializeWorkflowLocked(event)
	return event, nil
}

func (s *MemoryStore) ListIngestions() []slack.Ingestion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]slack.Ingestion(nil), s.ingestions...)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) CreateIngestion(envelope slack.SlackEnvelope) slack.Ingestion {
	s.mu.Lock()
	defer s.mu.Unlock()
	event := ingestion.EventEnvelope{
		Source:                     ingestion.SourceSlack,
		SourceEventID:              envelope.TS,
		ThreadKey:                  fmt.Sprintf("slack:%s:%s", envelope.ChannelID, envelope.ThreadTS),
		DedupeKey:                  fmt.Sprintf("slack:%s:%s", envelope.ChannelID, envelope.ThreadTS),
		Severity:                   severityFromText(envelope.Text),
		NormalizedProblemStatement: envelope.Text,
		OwnershipHint:              "platform",
		WorkflowHint:               deriveWorkflowHint(envelope.Text),
		Metadata: map[string]interface{}{
			"team_id":    envelope.TeamID,
			"channel_id": envelope.ChannelID,
			"user_id":    envelope.UserID,
			"thread_ts":  envelope.ThreadTS,
			"bot_role":   envelope.BotRole,
			"files":      envelope.Files,
		},
		CreatedAt: envelope.CreatedAt,
	}
	event.RawPayloadRef = fmt.Sprintf("memory://slack/%s/%s.json", envelope.ChannelID, strings.ReplaceAll(envelope.TS, ".", "-"))
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	created, _ := s.createEventLocked(event)
	for _, item := range s.ingestions {
		if item.EventID == created.ID {
			return item
		}
	}
	return slack.Ingestion{}
}

func (s *MemoryStore) materializeWorkflowLocked(event ingestion.EventEnvelope) {
	createdAt := event.CreatedAt
	ingestionID := nextID("ing", len(s.ingestions)+1)
	channelID, _ := event.Metadata["channel_id"].(string)
	userID, _ := event.Metadata["user_id"].(string)
	threadTS, _ := event.Metadata["thread_ts"].(string)
	botRoleValue, _ := event.Metadata["bot_role"].(slack.BotRole)
	if botRoleValue == "" {
		if raw, ok := event.Metadata["bot_role"].(string); ok {
			botRoleValue = slack.BotRole(raw)
		}
	}
	if channelID == "" {
		channelID = string(event.Source)
	}
	if userID == "" {
		userID = "system"
	}
	threadKey := event.ThreadKey
	if threadKey == "" {
		threadKey = event.IncidentKey
	}
	if threadKey == "" {
		threadKey = fmt.Sprintf("%s:%s", event.Source, event.SourceEventID)
	}
	intent := intentForWorkflowHint(event.WorkflowHint)
	assignedBot := assignedBotFor(event.WorkflowHint)
	if botRoleValue != "" {
		assignedBot = string(botRoleValue)
	}
	if threadTS == "" {
		threadTS = event.SourceEventID
	}

	ingestionItem := slack.Ingestion{
		ID:           ingestionID,
		EventID:      event.ID,
		ThreadKey:    threadKey,
		ThreadTS:     threadTS,
		WorkflowHint: event.WorkflowHint,
		Intent:       intent,
		BotRole:      slack.BotRole(assignedBot),
		Source:       string(event.Source),
		ChannelID:    channelID,
		UserID:       userID,
		Text:         event.NormalizedProblemStatement,
		CreatedAt:    createdAt,
	}
	s.ingestions = append(s.ingestions, ingestionItem)

	workflowID := nextID("wf", len(s.workflows)+1)
	traceID := nextID("trace", len(s.traces)+1)
	workflow := Workflow{
		ID:           workflowID,
		IngestionID:  ingestionID,
		TraceID:      traceID,
		ThreadKey:    threadKey,
		Kind:         event.WorkflowHint,
		Intent:       intent,
		AssignedBot:  assignedBot,
		ApprovalMode: approvalModeForIntent(intent),
		ResponseMode: responseModeForIntent(intent),
		Status:       "queued",
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
	}
	s.workflows = append(s.workflows, workflow)
	s.assignments = append(s.assignments, Assignment{
		ID:          nextID("as", len(s.assignments)+1),
		ThreadKey:   threadKey,
		AssignedBot: assignedBot,
		Confidence:  routeConfidenceForEvent(event),
		Rationale:   routingRationale(event),
		CreatedAt:   createdAt,
	})
	s.threadPolicies[threadKey] = policy.ThreadPolicy{
		ThreadKey:         threadKey,
		State:             policy.ThreadStateActive,
		OwnerBot:          assignedBot,
		LastPolicyVersion: "v2",
		UpdatedAt:         createdAt,
	}

	traceStatus := events.StatusQueued
	trace := events.Trace{
		Summary: events.TraceSummary{
			TraceID:      traceID,
			IngestionID:  ingestionID,
			WorkflowID:   workflowID,
			ThreadKey:    threadKey,
			WorkflowKind: event.WorkflowHint,
			Status:       traceStatus,
			StartedAt:    createdAt,
			EndedAt:      createdAt,
		},
		Events: []events.TraceEvent{
			{
				TraceID:     traceID,
				IngestionID: ingestionID,
				WorkflowID:  workflowID,
				Plane:       "edge",
				Service:     "control-plane",
				Actor:       "slack-ingestor",
				EventType:   "event.ingested",
				Status:      events.StatusCompleted,
				StartedAt:   createdAt,
				EndedAt:     timePtr(createdAt),
				PayloadRef:  event.RawPayloadRef,
				Description: fmt.Sprintf("%s event normalized into the event bus.", event.Source),
			},
			{
				TraceID:     traceID,
				IngestionID: ingestionID,
				WorkflowID:  workflowID,
				Plane:       "control",
				Service:     "control-plane",
				Actor:       "router-policy",
				EventType:   "workflow.queued",
				Status:      events.StatusQueued,
				StartedAt:   createdAt,
				Description: fmt.Sprintf("Queued %s workflow for bot %s.", intent, assignedBot),
			},
		},
		Artifacts: []events.Artifact{
			{
				ID:          nextID("artifact", len(s.traces)+1),
				TraceID:     traceID,
				Kind:        "event_payload",
				ContentType: "application/json",
				URL:         event.RawPayloadRef,
				SizeBytes:   2048,
				Source:      "event-bus",
			},
		},
		Reasoning: []events.ReasoningStep{
			{
				ID:         nextID("reason", len(s.traces)+1),
				TraceID:    traceID,
				WorkflowID: workflowID,
				StepType:   "intent_classified",
				Summary:    fmt.Sprintf("Classified incoming event as %s for workflow kind %s.", intent, event.WorkflowHint),
				EvidenceRefs: []events.EvidenceRef{
					{Kind: "event", Ref: event.ID, Summary: event.NormalizedProblemStatement},
				},
				Alternatives: []string{"incident", "feature_request", "question"},
				Confidence:   routeConfidenceForEvent(event),
				Decision:     fmt.Sprintf("route:%s assign:%s", intent, assignedBot),
				CreatedAt:    createdAt,
			},
		},
	}
	recomputeTraceSummary(&trace)
	s.traces[traceID] = trace
	_, _ = s.enqueueWorkItemLocked(queue.WorkItem{
		Queue:        queue.WorkflowQueue,
		Kind:         "workflow",
		Status:       queue.WorkQueued,
		TraceID:      traceID,
		WorkflowID:   workflowID,
		IngestionID:  ingestionID,
		ThreadKey:    threadKey,
		Intent:       intent,
		RepoScope:    event.OwnershipHint,
		RequestedBy:  "event_ingested",
		ApprovalMode: workflow.ApprovalMode,
		ResponseMode: workflow.ResponseMode,
		Payload: map[string]interface{}{
			"event_id":        event.ID,
			"workflow_hint":   event.WorkflowHint,
			"assigned_bot":    assignedBot,
			"channel_id":      channelID,
			"thread_ts":       threadTS,
			"problem":         event.NormalizedProblemStatement,
			"source":          event.Source,
			"raw_payload_ref": event.RawPayloadRef,
		},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
}

func severityFromText(text string) ingestion.Severity {
	switch {
	case containsAny(text, []string{"critical", "outage", "failing", "alert"}):
		return ingestion.SeverityCritical
	case containsAny(text, []string{"error", "incident", "broken"}):
		return ingestion.SeverityError
	default:
		return ingestion.SeverityWarning
	}
}

func routeConfidenceForEvent(event ingestion.EventEnvelope) float64 {
	switch event.WorkflowHint {
	case "incident":
		return 0.97
	case "feature-request":
		return 0.84
	default:
		return 0.91
	}
}

func routingRationale(event ingestion.EventEnvelope) string {
	switch event.WorkflowHint {
	case "incident":
		return "Matched operational incident patterns and severity indicators."
	case "feature-request":
		return "Matched product/request language."
	default:
		return "Matched architecture/repo reasoning signals."
	}
}

func (s *MemoryStore) ListWorkflows() []Workflow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]Workflow(nil), s.workflows...)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) ListAssignments() []Assignment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]Assignment(nil), s.assignments...)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) ListThreadPolicies() []policy.ThreadPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]policy.ThreadPolicy, 0, len(s.threadPolicies))
	for _, item := range s.threadPolicies {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func (s *MemoryStore) ListChannelPolicies() []policy.ChannelPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]policy.ChannelPolicy(nil), s.channelPolicy...)
}

func (s *MemoryStore) SetThreadState(threadKey string, state policy.ThreadState, owner string) (policy.ThreadPolicy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.threadPolicies[threadKey]
	if !ok {
		return policy.ThreadPolicy{}, errors.New("thread policy not found")
	}
	item.State = state
	item.Muted = state == policy.ThreadStateMuted || state == policy.ThreadStateMuteUntilMention
	if owner != "" {
		item.OwnerBot = owner
	}
	item.UpdatedAt = time.Now().UTC()
	s.threadPolicies[threadKey] = item
	return item, nil
}

func (s *MemoryStore) ListOwnershipRecords() []registry.OwnershipRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]registry.OwnershipRecord(nil), s.ownership...)
}

func (s *MemoryStore) ListCapabilities() []registry.CapabilityRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]registry.CapabilityRecord(nil), s.capabilities...)
}

func (s *MemoryStore) ListTemplates() []registry.WorkflowTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]registry.WorkflowTemplate(nil), s.templates...)
}

func (s *MemoryStore) ListExperiments() []registry.ExperimentRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]registry.ExperimentRecord(nil), s.experiments...)
}

func (s *MemoryStore) ListTraces() []events.TraceSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]events.TraceSummary, 0, len(s.traces))
	for _, trace := range s.traces {
		summary := trace.Summary
		if ratings := s.ratings[summary.TraceID]; len(ratings) > 0 {
			summary.LastVerdict = ratings[len(ratings)-1].Verdict
		}
		out = append(out, summary)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.After(out[j].StartedAt) })
	return out
}

func (s *MemoryStore) GetTrace(traceID string) (events.Trace, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	trace, ok := s.traces[traceID]
	return trace, ok
}

func (s *MemoryStore) AddRating(traceID string, rating review.HumanRating) (review.HumanRating, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	trace, ok := s.traces[traceID]
	if !ok {
		return review.HumanRating{}, errors.New("trace not found")
	}
	rating.TraceID = traceID
	rating.CreatedAt = time.Now().UTC()
	s.ratings[traceID] = append(s.ratings[traceID], rating)
	trace.Summary.LastVerdict = rating.Verdict
	trace.Summary.Status = events.StatusInReview
	s.traces[traceID] = trace
	_, _, _ = s.evaluateTraceLocked(traceID, "human_rating")
	return rating, nil
}

func (s *MemoryStore) AddImprovementNote(traceID string, note review.ImprovementNote) (review.ImprovementNote, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.traces[traceID]; !ok {
		return review.ImprovementNote{}, errors.New("trace not found")
	}
	note.TraceID = traceID
	note.CreatedAt = time.Now().UTC()
	s.notes[traceID] = append(s.notes[traceID], note)
	return note, nil
}

func (s *MemoryStore) ScheduleReplay(traceID string, requestedBy string) (queue.WorkItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	trace, ok := s.traces[traceID]
	if !ok {
		return queue.WorkItem{}, errors.New("trace not found")
	}
	item := queue.WorkItem{
		Queue:        queue.EvalQueue,
		Kind:         "trace_replay",
		Status:       queue.WorkQueued,
		TraceID:      traceID,
		WorkflowID:   trace.Summary.WorkflowID,
		IngestionID:  trace.Summary.IngestionID,
		ThreadKey:    trace.Summary.ThreadKey,
		RequestedBy:  requestedBy,
		ApprovalMode: "ui",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	created, err := s.enqueueWorkItemLocked(item)
	if err != nil {
		return queue.WorkItem{}, err
	}
	stepStatus := events.StatusReplayed
	trace.Summary.Status = events.StatusReplayed
	trace.Events = append(trace.Events, events.TraceEvent{
		TraceID:     traceID,
		IngestionID: trace.Summary.IngestionID,
		WorkflowID:  trace.Summary.WorkflowID,
		Plane:       "improvement",
		Service:     "improvement-plane",
		Actor:       "replay-scheduler",
		EventType:   "replay.queued",
		Status:      stepStatus,
		StartedAt:   time.Now().UTC(),
		Description: "Replay queued for later evaluation.",
	})
	recomputeTraceSummary(&trace)
	s.traces[traceID] = trace
	return created, nil
}

func (s *MemoryStore) ListEvalSuites() []evals.Suite {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]evals.Suite(nil), s.evalSuites...)
}

func (s *MemoryStore) ListEvalRuns() []evals.Run {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]evals.Run, 0, len(s.evalRuns))
	for _, run := range s.evalRuns {
		out = append(out, run)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) ListEvalJudgments(evalRunID string) []evals.Judgment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]evals.Judgment(nil), s.evalJudgments[evalRunID]...)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) EvaluateTrace(traceID string, trigger string) (evals.Run, []evals.Judgment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.evaluateTraceLocked(traceID, trigger)
}

func (s *MemoryStore) GetSettings() improvement.Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return normalizedSettings(s.settings)
}

func (s *MemoryStore) UpdateSettings(settings improvement.Settings) (improvement.Settings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	settings = normalizedSettings(settings)
	settings.UpdatedAt = time.Now().UTC()
	s.settings = settings
	return settings, nil
}

func (s *MemoryStore) ListWorkItems() []queue.WorkItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]queue.WorkItem, 0, len(s.workItems))
	for _, item := range s.workItems {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (s *MemoryStore) EnqueueWorkItem(item queue.WorkItem) (queue.WorkItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enqueueWorkItemLocked(item)
}

func (s *MemoryStore) enqueueWorkItemLocked(item queue.WorkItem) (queue.WorkItem, error) {
	now := time.Now().UTC()
	if item.ID == "" {
		item.ID = nextID("work", len(s.workItems)+1)
	}
	if item.Status == "" {
		item.Status = queue.WorkQueued
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	if item.Payload == nil {
		item.Payload = map[string]interface{}{}
	}
	s.workItems[item.ID] = item
	return item, nil
}

func (s *MemoryStore) ClaimNextWorkItem(queuesToCheck []queue.QueueName, holder string, lease time.Duration) (queue.WorkItem, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if holder == "" {
		return queue.WorkItem{}, false, errors.New("holder is required")
	}
	if lease <= 0 {
		lease = 30 * time.Second
	}
	queueFilter := map[queue.QueueName]struct{}{}
	for _, name := range queuesToCheck {
		queueFilter[name] = struct{}{}
	}
	now := time.Now().UTC()
	candidates := make([]queue.WorkItem, 0, len(s.workItems))
	for _, item := range s.workItems {
		if len(queueFilter) > 0 {
			if _, ok := queueFilter[item.Queue]; !ok {
				continue
			}
		}
		expired := item.Status == queue.WorkLeased && item.LeaseExpiresAt != nil && item.LeaseExpiresAt.Before(now)
		if item.Status == queue.WorkQueued || expired {
			candidates = append(candidates, item)
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].CreatedAt.Equal(candidates[j].CreatedAt) {
			return candidates[i].ID < candidates[j].ID
		}
		return candidates[i].CreatedAt.Before(candidates[j].CreatedAt)
	})
	if len(candidates) == 0 {
		return queue.WorkItem{}, false, nil
	}
	item := candidates[0]
	expires := now.Add(lease)
	item.Status = queue.WorkLeased
	item.Attempts++
	item.LeaseOwner = holder
	item.LeaseExpiresAt = &expires
	item.UpdatedAt = now
	s.workItems[item.ID] = item
	return item, true, nil
}

func (s *MemoryStore) CompleteWorkItem(id string) (queue.WorkItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.workItems[id]
	if !ok {
		return queue.WorkItem{}, errors.New("work item not found")
	}
	now := time.Now().UTC()
	item.Status = queue.WorkCompleted
	item.UpdatedAt = now
	item.CompletedAt = &now
	item.LeaseOwner = ""
	item.LeaseExpiresAt = nil
	s.workItems[id] = item
	return item, nil
}

func (s *MemoryStore) FailWorkItem(id string, lastError string) (queue.WorkItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.workItems[id]
	if !ok {
		return queue.WorkItem{}, errors.New("work item not found")
	}
	now := time.Now().UTC()
	item.Status = queue.WorkFailed
	item.LastError = lastError
	item.UpdatedAt = now
	item.CompletedAt = &now
	item.LeaseOwner = ""
	item.LeaseExpiresAt = nil
	s.workItems[id] = item
	return item, nil
}

func (s *MemoryStore) UpdateWorkflowStatus(workflowID string, status string, lastError string) (Workflow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.updateWorkflowStatusLocked(workflowID, status, lastError)
}

func (s *MemoryStore) updateWorkflowStatusLocked(workflowID string, status string, lastError string) (Workflow, error) {
	for i := range s.workflows {
		if s.workflows[i].ID != workflowID {
			continue
		}
		now := time.Now().UTC()
		s.workflows[i].Status = status
		s.workflows[i].LastError = lastError
		s.workflows[i].UpdatedAt = now
		if status == "completed" || status == "failed" {
			s.workflows[i].CompletedAt = &now
		}
		return s.workflows[i], nil
	}
	return Workflow{}, errors.New("workflow not found")
}

func (s *MemoryStore) ApplyTraceUpdate(traceID string, update TraceUpdate) (events.Trace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	trace, ok := s.traces[traceID]
	if !ok {
		return events.Trace{}, errors.New("trace not found")
	}
	now := time.Now().UTC()
	if update.Status != nil {
		trace.Summary.Status = *update.Status
	}
	if update.LastVerdict != nil {
		trace.Summary.LastVerdict = *update.LastVerdict
	}
	trace.Events = append(trace.Events, update.Events...)
	trace.Artifacts = append(trace.Artifacts, update.Artifacts...)
	trace.Reasoning = append(trace.Reasoning, update.Reasoning...)
	trace.ToolCalls = append(trace.ToolCalls, update.ToolCalls...)
	trace.SlackActions = append(trace.SlackActions, update.SlackActions...)
	recomputeTraceSummary(&trace)
	if trace.Summary.EndedAt.Before(now) && (trace.Summary.Status == events.StatusCompleted || trace.Summary.Status == events.StatusFailed || trace.Summary.Status == events.StatusNeedsHuman) {
		trace.Summary.EndedAt = now
	}
	s.traces[traceID] = trace
	if update.WorkflowStatus != "" {
		if _, err := s.updateWorkflowStatusLocked(trace.Summary.WorkflowID, update.WorkflowStatus, update.WorkflowError); err != nil {
			return events.Trace{}, err
		}
	}
	return trace, nil
}

func (s *MemoryStore) evaluateTraceLocked(traceID string, trigger string) (evals.Run, []evals.Judgment, error) {
	trace, ok := s.traces[traceID]
	if !ok {
		return evals.Run{}, nil, errors.New("trace not found")
	}
	event := s.findEventByTraceLocked(trace)
	suiteName := suiteNameForTrace(trace, event)
	createdAt := time.Now().UTC()
	runID := nextID("eval", len(s.evalRuns)+1)
	judgments := buildJudgments(trace, event, s.ratings[traceID], s.proposalMemory)
	overallScore := 0.0
	overallVerdict := "pass"
	for _, judgment := range judgments {
		overallScore += judgment.Score
		if !judgment.Passed {
			overallVerdict = "needs_improvement"
		}
	}
	if len(judgments) > 0 {
		overallScore = overallScore / float64(len(judgments))
	}
	run := evals.Run{
		ID:             runID,
		TraceID:        traceID,
		SuiteName:      suiteName,
		Status:         evals.StatusCompleted,
		Trigger:        trigger,
		OverallScore:   overallScore,
		OverallVerdict: overallVerdict,
		CreatedAt:      createdAt,
		CompletedAt:    createdAt,
	}
	if event != nil {
		run.EventID = event.ID
	}
	s.evalRuns[runID] = run
	s.evalJudgments[runID] = judgments
	s.updateCandidateLocked(trace, event, run, judgments)
	return run, append([]evals.Judgment(nil), judgments...), nil
}

func buildJudgments(trace events.Trace, event *ingestion.EventEnvelope, ratings []review.HumanRating, memories []review.ProposalMemory) []evals.Judgment {
	now := time.Now().UTC()
	judgments := []evals.Judgment{}
	deterministicScore := 0.85
	deterministicReason := "Trace completed within expected policy, cost, and validation budget."
	if trace.Summary.Status == events.StatusFailed || trace.Summary.Status == events.StatusNeedsHuman {
		deterministicScore = 0.35
		deterministicReason = "Trace indicates an operational failure or unresolved handoff."
	}
	judgments = append(judgments, evals.Judgment{
		ID:        nextID("judge", 1+len(judgments)),
		Layer:     evals.LayerDeterministic,
		Category:  "policy_and_reliability",
		Score:     deterministicScore,
		Passed:    deterministicScore >= 0.7,
		Rationale: deterministicReason,
		CreatedAt: now,
	})

	taskScore := 0.82
	taskReason := "Workflow output appears actionable."
	if len(ratings) > 0 {
		last := ratings[len(ratings)-1]
		taskScore = float64(last.Score) / 5.0
		taskReason = fmt.Sprintf("Human review verdict=%s", last.Verdict)
	} else if event != nil && containsAny(event.NormalizedProblemStatement, []string{"failing", "alert", "critical", "closed-loop"}) {
		taskScore = 0.58
		taskReason = "High-severity event without strong evidence of a complete remediation."
	}
	judgments = append(judgments, evals.Judgment{
		ID:        nextID("judge", 1+len(judgments)),
		Layer:     evals.LayerTaskQuality,
		Category:  "task_quality",
		Score:     taskScore,
		Passed:    taskScore >= 0.7,
		Rationale: taskReason,
		CreatedAt: now,
	})

	architectureScore := 0.8
	architectureReason := "Architecture boundary and recursive improvement controls look healthy."
	if event != nil && event.OwnershipHint == "platform" && (containsAny(event.NormalizedProblemStatement, []string{"proposal", "eval", "closed-loop", "architecture"}) || event.Source == ingestion.SourceSentry) {
		architectureScore = 0.54
		architectureReason = "Platform-level event suggests architectural debt or missing self-improvement guardrails."
	}
	if hasRecentRejectedMemory(memories, candidateKeyForTrace(trace, event)) {
		architectureScore -= 0.08
		architectureReason = "Similar platform issue was rejected before; stronger novelty is required."
	}
	if architectureScore < 0 {
		architectureScore = 0
	}
	judgments = append(judgments, evals.Judgment{
		ID:        nextID("judge", 1+len(judgments)),
		Layer:     evals.LayerArchitecture,
		Category:  "architecture_health",
		Score:     architectureScore,
		Passed:    architectureScore >= 0.7,
		Rationale: architectureReason,
		CreatedAt: now,
	})

	return judgments
}

func hasRecentRejectedMemory(memories []review.ProposalMemory, candidateKey string) bool {
	for _, memory := range memories {
		if memory.CandidateKey == candidateKey && memory.Disposition == review.ProposalRejected && time.Since(memory.CreatedAt) < 30*24*time.Hour {
			return true
		}
	}
	return false
}

func suiteNameForTrace(trace events.Trace, event *ingestion.EventEnvelope) string {
	if event != nil && event.Source == ingestion.SourceSentry {
		return "incident-response"
	}
	switch trace.Summary.WorkflowKind {
	case "incident":
		return "incident-response"
	case "feature-request":
		return "proposal-quality"
	default:
		return "architecture-review"
	}
}

func (s *MemoryStore) findEventByTraceLocked(trace events.Trace) *ingestion.EventEnvelope {
	for i := range s.ingestions {
		if s.ingestions[i].ID != trace.Summary.IngestionID {
			continue
		}
		for j := range s.events {
			if sameThread(s.events[j], s.ingestions[i].ThreadKey) && sameCreatedWindow(s.events[j].CreatedAt, s.ingestions[i].CreatedAt) {
				return &s.events[j]
			}
		}
	}
	return nil
}

func sameThread(event ingestion.EventEnvelope, threadKey string) bool {
	return event.ThreadKey == threadKey || event.IncidentKey == threadKey || fmt.Sprintf("%s:%s", event.Source, event.SourceEventID) == threadKey
}

func sameCreatedWindow(a, b time.Time) bool {
	return a.Equal(b) || a.Sub(b) < time.Second && b.Sub(a) < time.Second
}

func (s *MemoryStore) updateCandidateLocked(trace events.Trace, event *ingestion.EventEnvelope, run evals.Run, judgments []evals.Judgment) {
	failureMode := failureModeForTrace(trace, judgments)
	key := fmt.Sprintf("%s:%s:%s", subsystemForTrace(trace, event), trace.Summary.WorkflowKind, failureMode)
	now := time.Now().UTC()
	candidate, ok := s.candidates[key]
	if !ok {
		candidate = improvement.Candidate{
			ID:               nextID("cand", len(s.candidates)+1),
			CandidateKey:     key,
			Subsystem:        subsystemForTrace(trace, event),
			FailureMode:      failureMode,
			InterventionType: interventionTypeForJudgments(judgments),
			Status:           improvement.CandidateNeedsEvidence,
			Severity:         severityForTrace(trace, event),
			RiskTier:         improvement.RiskMedium,
			Hypothesis:       hypothesisForTrace(trace, event, judgments),
			ProposedScope:    proposedScopeForTrace(trace, event),
			CreatedAt:        now,
		}
	}
	candidate.RecurrenceCount++
	candidate.ExpectedImpact = maxFloat(candidate.ExpectedImpact, 1-run.OverallScore)
	candidate.NoveltyScore = noveltyScoreForCandidate(s.proposalMemory, candidate.CandidateKey)
	candidate.ConfidenceScore = confidenceScoreForCandidate(candidate.RecurrenceCount, judgments)
	candidate.FreshnessScore = 1.0
	candidate.PriorityScore = priorityScore(candidate.ExpectedImpact, candidate.NoveltyScore, candidate.ConfidenceScore, candidate.FreshnessScore, candidate.RecurrenceCount)
	candidate.LatestTraceID = trace.Summary.TraceID
	candidate.SourceEvalIDs = appendUnique(candidate.SourceEvalIDs, run.ID)
	candidate.EvidenceArtifactIDs = collectArtifactIDs(trace)
	candidate.PriorSimilarProposalIDs = similarProposalIDs(s.proposalMemory, candidate.CandidateKey)
	candidate.NewEvidenceSinceLastRejection = hasNewEvidenceSinceLastRejection(s.proposalMemory, candidate.CandidateKey, run.CreatedAt, candidate.RecurrenceCount)
	candidate.LastEvaluatedAt = now
	candidate.UpdatedAt = now
	if candidate.RecurrenceCount >= 2 || run.OverallScore < 0.65 {
		candidate.Status = improvement.CandidateQueued
	} else {
		candidate.Status = improvement.CandidateNeedsEvidence
	}
	if len(candidate.PriorSimilarProposalIDs) > 0 && !candidate.NewEvidenceSinceLastRejection && rejectedRecently(s.proposalMemory, candidate.CandidateKey) {
		candidate.Status = improvement.CandidateNeedsEvidence
		candidate.PriorityScore *= 0.4
	}
	s.candidates[key] = candidate
}

func candidateKeyForTrace(trace events.Trace, event *ingestion.EventEnvelope) string {
	subsystem := subsystemForTrace(trace, event)
	return fmt.Sprintf("%s:%s:%s", subsystem, trace.Summary.WorkflowKind, failureModeForTrace(trace, nil))
}

func subsystemForTrace(trace events.Trace, event *ingestion.EventEnvelope) string {
	if event != nil && event.OwnershipHint != "" {
		return event.OwnershipHint
	}
	switch trace.Summary.WorkflowKind {
	case "incident":
		return "platform"
	default:
		return "architecture"
	}
}

func severityForTrace(trace events.Trace, event *ingestion.EventEnvelope) string {
	if event != nil {
		return string(event.Severity)
	}
	if trace.Summary.Status == events.StatusFailed {
		return string(ingestion.SeverityError)
	}
	return string(ingestion.SeverityWarning)
}

func failureModeForTrace(trace events.Trace, judgments []evals.Judgment) string {
	if trace.Summary.Status == events.StatusFailed {
		return "failed_workflow"
	}
	for _, judgment := range judgments {
		if !judgment.Passed {
			return judgment.Category
		}
	}
	if trace.Summary.LastVerdict != "" {
		return trace.Summary.LastVerdict
	}
	return "quality_gap"
}

func interventionTypeForJudgments(judgments []evals.Judgment) string {
	for _, judgment := range judgments {
		if judgment.Layer == evals.LayerArchitecture && !judgment.Passed {
			return "architecture_refactor"
		}
		if judgment.Layer == evals.LayerDeterministic && !judgment.Passed {
			return "policy_or_runtime_fix"
		}
	}
	return "prompt_or_workflow_tune"
}

func hypothesisForTrace(trace events.Trace, event *ingestion.EventEnvelope, judgments []evals.Judgment) string {
	if event != nil && event.Source == ingestion.SourceSentry {
		return "Introduce stronger proposal slot gating and shared-state evaluation to avoid repeated recursive failures."
	}
	if containsAny(trace.Summary.WorkflowKind, []string{"architecture"}) {
		return "Improve architecture routing and add stronger closed-loop evaluation signals."
	}
	return "Tighten workflow policy and evaluation criteria to improve remediation quality."
}

func proposedScopeForTrace(trace events.Trace, event *ingestion.EventEnvelope) string {
	if event != nil && event.Source == ingestion.SourceSentry {
		return "platform + adapters"
	}
	return "whole_repo"
}

func noveltyScoreForCandidate(memories []review.ProposalMemory, key string) float64 {
	if len(similarProposalIDs(memories, key)) == 0 {
		return 0.92
	}
	if rejectedRecently(memories, key) {
		return 0.35
	}
	return 0.62
}

func rejectedRecently(memories []review.ProposalMemory, key string) bool {
	for _, memory := range memories {
		if memory.CandidateKey == key && memory.Disposition == review.ProposalRejected && time.Since(memory.CreatedAt) < 30*24*time.Hour {
			return true
		}
	}
	return false
}

func confidenceScoreForCandidate(recurrence int, judgments []evals.Judgment) float64 {
	score := 0.5 + float64(recurrence)*0.1
	failed := 0
	for _, judgment := range judgments {
		if !judgment.Passed {
			failed++
		}
	}
	score += float64(failed) * 0.08
	if score > 1 {
		return 1
	}
	return score
}

func priorityScore(expectedImpact, novelty, confidence, freshness float64, recurrence int) float64 {
	return expectedImpact*0.35 + novelty*0.2 + confidence*0.25 + freshness*0.1 + minFloat(float64(recurrence)/5.0, 0.1)
}

func collectArtifactIDs(trace events.Trace) []string {
	out := make([]string, 0, len(trace.Artifacts))
	for _, artifact := range trace.Artifacts {
		out = append(out, artifact.ID)
	}
	return out
}

func similarProposalIDs(memories []review.ProposalMemory, key string) []string {
	out := []string{}
	for _, memory := range memories {
		if memory.CandidateKey == key {
			out = append(out, memory.ProposalID)
		}
	}
	return out
}

func hasNewEvidenceSinceLastRejection(memories []review.ProposalMemory, key string, evaluatedAt time.Time, recurrence int) bool {
	latest := time.Time{}
	for _, memory := range memories {
		if memory.CandidateKey == key && (memory.Disposition == review.ProposalRejected || memory.Disposition == review.ProposalDismissed) && memory.CreatedAt.After(latest) {
			latest = memory.CreatedAt
		}
	}
	if latest.IsZero() {
		return true
	}
	return recurrence >= 2 && evaluatedAt.After(latest)
}

func (s *MemoryStore) ListCandidates() []improvement.Candidate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]improvement.Candidate, 0, len(s.candidates))
	for _, candidate := range s.candidates {
		out = append(out, candidate)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PriorityScore == out[j].PriorityScore {
			return out[i].UpdatedAt.After(out[j].UpdatedAt)
		}
		return out[i].PriorityScore > out[j].PriorityScore
	})
	return out
}

func (s *MemoryStore) ListProposalMemories() []review.ProposalMemory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]review.ProposalMemory(nil), s.proposalMemory...)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) GetProposalSlots() ProposalSlotState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.proposalSlotsLocked(time.Now().UTC())
}

func (s *MemoryStore) proposalSlotsLocked(now time.Time) ProposalSlotState {
	settings := normalizedSettings(s.settings)
	out := ProposalSlotState{Cap: settings.ActiveProposalCap}
	for _, proposal := range s.proposals {
		if review.ConsumesActiveProposalSlot(proposal.Status) {
			out.Active++
			out.ActiveProposalIDs = append(out.ActiveProposalIDs, proposal.ID)
			if proposal.Status == review.ProposalPendingReview && !proposal.ReviewDeadline.IsZero() && proposal.ReviewDeadline.Before(now) {
				out.StaleProposalIDs = append(out.StaleProposalIDs, proposal.ID)
			}
		}
	}
	sort.Strings(out.ActiveProposalIDs)
	sort.Strings(out.StaleProposalIDs)
	out.Available = out.Cap - out.Active
	if out.Available < 0 {
		out.Available = 0
	}
	return out
}

func (s *MemoryStore) PromoteCandidates(requestedBy string, limit int) (PromotionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.promoteCandidatesLocked(requestedBy, limit)
}

func (s *MemoryStore) promoteCandidatesLocked(requestedBy string, limit int) (PromotionResult, error) {
	now := time.Now().UTC()
	slots := s.proposalSlotsLocked(now)
	result := PromotionResult{BlockedByCap: slots.Available == 0, StaleProposalIDs: slots.StaleProposalIDs}
	if slots.Available == 0 {
		return result, nil
	}
	candidates := make([]improvement.Candidate, 0, len(s.candidates))
	for _, candidate := range s.candidates {
		candidates = append(candidates, candidate)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].PriorityScore == candidates[j].PriorityScore {
			return candidates[i].UpdatedAt.After(candidates[j].UpdatedAt)
		}
		return candidates[i].PriorityScore > candidates[j].PriorityScore
	})
	allowed := slots.Available
	if limit > 0 {
		allowed = minInt(limit, slots.Available)
	}
	for _, candidate := range candidates {
		if allowed == 0 {
			break
		}
		if candidate.Status != improvement.CandidateQueued {
			continue
		}
		if s.hasActiveProposalForCandidateLocked(candidate.CandidateKey) {
			continue
		}
		if rejectedRecently(s.proposalMemory, candidate.CandidateKey) && !candidate.NewEvidenceSinceLastRejection {
			continue
		}
		proposal := review.Proposal{
			ID:                            nextID("proposal", len(s.proposals)+1),
			TraceID:                       candidate.LatestTraceID,
			Title:                         proposalTitle(candidate),
			Category:                      candidate.InterventionType,
			Summary:                       candidate.Hypothesis,
			Status:                        review.ProposalPendingReview,
			CandidateKey:                  candidate.CandidateKey,
			SourceEvalIDs:                 append([]string(nil), candidate.SourceEvalIDs...),
			RiskTier:                      string(candidate.RiskTier),
			ProposedScope:                 candidate.ProposedScope,
			EvidenceArtifactIDs:           append([]string(nil), candidate.EvidenceArtifactIDs...),
			ActiveSlotConsuming:           true,
			ReviewDeadline:                now.Add(proposalReviewSLA),
			PriorSimilarProposalIDs:       append([]string(nil), candidate.PriorSimilarProposalIDs...),
			NewEvidenceSinceLastRejection: candidate.NewEvidenceSinceLastRejection,
			CreatedAt:                     now,
		}
		s.proposals[proposal.ID] = proposal
		candidate.Status = improvement.CandidatePromoted
		candidate.UpdatedAt = now
		s.candidates[candidate.CandidateKey] = candidate
		result.Promoted++
		result.PromotedIDs = append(result.PromotedIDs, proposal.ID)
		allowed--
	}
	result.BlockedByCap = result.Promoted == 0 && slots.Available == 0
	return result, nil
}

func (s *MemoryStore) hasActiveProposalForCandidateLocked(candidateKey string) bool {
	for _, proposal := range s.proposals {
		if proposal.CandidateKey == candidateKey && review.ConsumesActiveProposalSlot(proposal.Status) {
			return true
		}
	}
	return false
}

func proposalTitle(candidate improvement.Candidate) string {
	return fmt.Sprintf("Improve %s: %s", candidate.Subsystem, candidate.FailureMode)
}

func (s *MemoryStore) RunProposalPromoter(holder string) (PromotionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if holder == "" {
		holder = "improvement-plane-cron"
	}
	if lease, ok := s.cronLeases["improvement-plane-cron"]; ok && lease.ExpiresAt.After(now) && lease.Holder != holder {
		return PromotionResult{}, errors.New("proposal promoter lease already held")
	}
	s.cronLeases["improvement-plane-cron"] = improvement.CronLease{
		Name:      "improvement-plane-cron",
		Holder:    holder,
		ExpiresAt: now.Add(proposalPromoterLease),
	}
	return s.promoteCandidatesLocked(holder, normalizedSettings(s.settings).ActiveProposalCap)
}

func (s *MemoryStore) ListProposals() []review.Proposal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]review.Proposal, 0, len(s.proposals))
	for _, proposal := range s.proposals {
		out = append(out, proposal)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) ReviewProposal(proposalID string, decision review.ProposalReview) (review.Proposal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	proposal, ok := s.proposals[proposalID]
	if !ok {
		return review.Proposal{}, errors.New("proposal not found")
	}
	decision.ProposalID = proposalID
	decision.CreatedAt = time.Now().UTC()
	if len(decision.FailureClasses) == 0 && decision.FailureClass != "" {
		decision.FailureClasses = []string{decision.FailureClass}
	}
	proposal.Reviews = append(proposal.Reviews, decision)
	proposal.Reviewer = decision.ReviewerID
	proposal.Status = review.ProposalStatus(decision.Decision)
	proposal.ActiveSlotConsuming = review.ConsumesActiveProposalSlot(proposal.Status)
	s.proposals[proposalID] = proposal

	memory := review.ProposalMemory{
		ID:                nextID("memory", len(s.proposalMemory)+1),
		ProposalID:        proposalID,
		CandidateKey:      proposal.CandidateKey,
		Hypothesis:        proposal.Summary,
		DiffSummary:       proposal.ProposedScope,
		ReviewRationale:   decision.Rationale,
		Disposition:       proposal.Status,
		DispositionReason: decision.Rationale,
		FailureClass:      decision.FailureClass,
		FailureClasses:    append([]string(nil), decision.FailureClasses...),
		SourceEvalIDs:     append([]string(nil), proposal.SourceEvalIDs...),
		LinkedArtifactIDs: append([]string(nil), proposal.EvidenceArtifactIDs...),
		LinkedProposalIDs: append([]string(nil), proposal.PriorSimilarProposalIDs...),
		CreatedAt:         decision.CreatedAt,
	}
	s.proposalMemory = append(s.proposalMemory, memory)

	candidate := s.candidates[proposal.CandidateKey]
	switch proposal.Status {
	case review.ProposalApproved:
		candidate.Status = improvement.CandidatePromoted
		_, _ = s.enqueueWorkItemLocked(queue.WorkItem{
			Queue:        queue.ProposalQueue,
			Kind:         "approved_proposal",
			Status:       queue.WorkQueued,
			TraceID:      proposal.TraceID,
			ProposalID:   proposal.ID,
			RequestedBy:  decision.ReviewerID,
			ApprovalMode: "human_review",
			CreatedAt:    decision.CreatedAt,
			UpdatedAt:    decision.CreatedAt,
			Payload: map[string]interface{}{
				"candidate_key": proposal.CandidateKey,
				"risk_tier":     proposal.RiskTier,
			},
		})
	case review.ProposalRejected, review.ProposalDismissed:
		candidate.Status = improvement.CandidateNeedsEvidence
		candidate.NewEvidenceSinceLastRejection = false
	case review.ProposalMerged:
		candidate.Status = improvement.CandidateDormant
		replayID := nextID("pmr", len(s.postMergeReplay)+1)
		s.postMergeReplay[replayID] = improvement.PostMergeReplay{
			ID:             replayID,
			ProposalID:     proposal.ID,
			TraceID:        proposal.TraceID,
			BaselineScore:  latestEvalScoreForTrace(s.evalRuns, proposal.TraceID),
			CandidateScore: minFloat(1.0, latestEvalScoreForTrace(s.evalRuns, proposal.TraceID)+0.15),
			Improved:       true,
			CreatedAt:      decision.CreatedAt,
		}
	default:
		candidate.Status = improvement.CandidateDormant
	}
	candidate.UpdatedAt = decision.CreatedAt
	s.candidates[proposal.CandidateKey] = candidate
	return proposal, nil
}

func (s *MemoryStore) UpdateProposalStatus(proposalID string, status review.ProposalStatus) (review.Proposal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	proposal, ok := s.proposals[proposalID]
	if !ok {
		return review.Proposal{}, errors.New("proposal not found")
	}
	proposal.Status = status
	proposal.ActiveSlotConsuming = review.ConsumesActiveProposalSlot(status)
	s.proposals[proposalID] = proposal
	return proposal, nil
}

func (s *MemoryStore) MaterializeApprovedProposal(proposalID string, requestedBy string) (improvement.RepoChangeJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	proposal, ok := s.proposals[proposalID]
	if !ok {
		return improvement.RepoChangeJob{}, errors.New("proposal not found")
	}
	if proposal.Status != review.ProposalApproved && proposal.Status != review.ProposalRepoChangeQueued && proposal.Status != review.ProposalRepoChangeRunning && proposal.Status != review.ProposalValidationPending && proposal.Status != review.ProposalPROpen {
		return improvement.RepoChangeJob{}, fmt.Errorf("proposal %s is not approved for materialization", proposal.Status)
	}
	for _, existing := range s.repoChangeJobs {
		if existing.ProposalID == proposalID {
			return existing, nil
		}
	}
	now := time.Now().UTC()
	jobID := nextID("job", len(s.repoChangeJobs)+1)
	job := improvement.RepoChangeJob{
		ID:               jobID,
		ProposalID:       proposal.ID,
		CandidateKey:     proposal.CandidateKey,
		Status:           string(review.ProposalRepoChangeQueued),
		Repo:             proposalRepo(proposal),
		BaseRef:          "main",
		BranchName:       fmt.Sprintf("codex/%s", proposal.ID),
		AllowedPathGlobs: []string{"cmd/**", "internal/**", "runner/**", "ui/**", "README.md", "Makefile"},
		ContextSummary:   buildRepoChangeContext(proposal, s.proposalMemory),
		CreatedAt:        now,
	}
	s.repoChangeJobs[jobID] = job
	proposal.Status = review.ProposalRepoChangeQueued
	proposal.ActiveSlotConsuming = true
	s.proposals[proposal.ID] = proposal
	_, _ = s.enqueueWorkItemLocked(queue.WorkItem{
		Queue:        queue.SandboxQueue,
		Kind:         "repo_change_job",
		Status:       queue.WorkQueued,
		TraceID:      proposal.TraceID,
		ProposalID:   proposal.ID,
		RepoScope:    job.Repo,
		RequestedBy:  requestedBy,
		ApprovalMode: "approved",
		CreatedAt:    now,
		UpdatedAt:    now,
		Payload: map[string]interface{}{
			"branch_name": job.BranchName,
			"base_ref":    job.BaseRef,
			"job_id":      job.ID,
		},
	})
	return job, nil
}

func buildRepoChangeContext(proposal review.Proposal, memories []review.ProposalMemory) string {
	context := fmt.Sprintf("Proposal %s targets %s with scope %s.", proposal.ID, proposal.CandidateKey, proposal.ProposedScope)
	for _, memory := range memories {
		if memory.CandidateKey == proposal.CandidateKey && (memory.Disposition == review.ProposalRejected || memory.Disposition == review.ProposalDismissed) {
			context += fmt.Sprintf(" Prior %s rationale: %s.", memory.Disposition, memory.ReviewRationale)
		}
	}
	return context
}

func latestEvalScoreForTrace(runs map[string]evals.Run, traceID string) float64 {
	latest := time.Time{}
	score := 0.0
	for _, run := range runs {
		if run.TraceID == traceID && run.CreatedAt.After(latest) {
			latest = run.CreatedAt
			score = run.OverallScore
		}
	}
	return score
}

func (s *MemoryStore) ListRepoChangeJobs() []improvement.RepoChangeJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]improvement.RepoChangeJob, 0, len(s.repoChangeJobs))
	for _, item := range s.repoChangeJobs {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) UpdateRepoChangeJobStatus(jobID string, status string) (improvement.RepoChangeJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.repoChangeJobs[jobID]
	if !ok {
		return improvement.RepoChangeJob{}, errors.New("repo change job not found")
	}
	item.Status = status
	s.repoChangeJobs[jobID] = item
	return item, nil
}

func (s *MemoryStore) ListPRAttempts() []improvement.PRAttempt {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]improvement.PRAttempt, 0, len(s.prAttempts))
	for _, item := range s.prAttempts {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) RecordPRAttempt(attempt improvement.PRAttempt) (improvement.PRAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if attempt.ID == "" {
		attempt.ID = nextID("pr", len(s.prAttempts)+1)
	}
	if attempt.CreatedAt.IsZero() {
		attempt.CreatedAt = time.Now().UTC()
	}
	s.prAttempts[attempt.ID] = attempt
	if attempt.ProposalID != "" {
		if proposal, ok := s.proposals[attempt.ProposalID]; ok && attempt.Status == string(review.ProposalPROpen) {
			proposal.Status = review.ProposalPROpen
			proposal.ActiveSlotConsuming = true
			s.proposals[proposal.ID] = proposal
		}
	}
	return attempt, nil
}

func (s *MemoryStore) ListPostMergeReplays() []improvement.PostMergeReplay {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]improvement.PostMergeReplay, 0, len(s.postMergeReplay))
	for _, item := range s.postMergeReplay {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) ExecuteTool(name string, input map[string]interface{}) ToolResult {
	now := time.Now().UTC()
	toolCallID := nextID("toolcall", len(s.prAttempts)+len(s.repoChangeJobs)+len(s.events)+1)
	approved := name != "github.create_pr"
	approvalState := "not_required"
	summary := fmt.Sprintf("tool execution for %s", name)
	rawArtifactRefs := []string{}
	if name == "github.create_pr" {
		approvalState = "human_approved"
		if proposalID, _ := input["proposal_id"].(string); proposalID != "" {
			if proposal, ok := s.proposals[proposalID]; ok && proposal.Status == review.ProposalApproved {
				slots := s.proposalSlotsLocked(now)
				if slots.Active > slots.Cap {
					approved = false
					approvalState = "blocked_by_cap"
					summary = "proposal cap exceeded; refusing to open another PR-backed proposal"
				}
				if approved {
					proposal.Status = review.ProposalPROpen
					proposal.ActiveSlotConsuming = true
					s.proposals[proposalID] = proposal
				}
				approved = true
				attemptID := nextID("pr", len(s.prAttempts)+1)
				s.prAttempts[attemptID] = improvement.PRAttempt{
					ID:               attemptID,
					ProposalID:       proposalID,
					Repo:             "rsi-agent-platform",
					BranchName:       fmt.Sprintf("codex/%s", proposalID),
					PRURL:            fmt.Sprintf("https://github.com/piplabs/rsi-agent-platform/pull/%d", len(s.prAttempts)+100),
					Status:           string(review.ProposalPROpen),
					ValidationStatus: "pending",
					CreatedAt:        now,
				}
				summary = fmt.Sprintf("draft PR opened for proposal %s", proposalID)
			}
		}
		if !approved && summary == fmt.Sprintf("tool execution for %s", name) {
			approvalState = "missing_or_unapproved_proposal"
			summary = "github.create_pr requires an approved active proposal"
		}
	}
	return ToolResult{
		Name:          name,
		ToolCallID:    toolCallID,
		Approved:      approved,
		ApprovalState: approvalState,
		ExecutedAt:    now,
		Input:         input,
		Output: map[string]interface{}{
			"status":  "ok",
			"message": summary,
		},
		Summary:         summary,
		RawArtifactRefs: rawArtifactRefs,
		Metadata: map[string]interface{}{
			"tool_name": name,
		},
	}
}

func nextID(prefix string, n int) string {
	return fmt.Sprintf("%s-%03d", prefix, n)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func deriveWorkflowHint(text string) string {
	switch {
	case containsAny(text, []string{"incident", "failing", "broken", "debug", "outage", "alert", "critical"}):
		return "incident"
	case containsAny(text, []string{"feature", "request", "product", "need"}):
		return "feature-request"
	default:
		return "architecture"
	}
}

func intentForWorkflowHint(kind string) string {
	switch kind {
	case "incident":
		return "incident"
	case "feature-request":
		return "feature_request"
	default:
		return "question"
	}
}

func approvalModeForIntent(intent string) string {
	switch intent {
	case "feature_request":
		return "human_required"
	default:
		return "policy_gated"
	}
}

func responseModeForIntent(intent string) string {
	switch intent {
	case "incident":
		return "thread_updates"
	default:
		return "reply_in_thread"
	}
}

func assignedBotFor(kind string) string {
	switch kind {
	case "incident":
		return "oncall"
	case "feature-request":
		return "fr"
	default:
		return "arch"
	}
}

func containsAny(text string, needles []string) bool {
	for _, needle := range needles {
		if needle != "" && stringContainsFold(text, needle) {
			return true
		}
	}
	return false
}

func stringContainsFold(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && containsFold(s, sub)
}

func containsFold(s, substr string) bool {
	sRunes := []rune(s)
	subRunes := []rune(substr)
	for i := 0; i <= len(sRunes)-len(subRunes); i++ {
		match := true
		for j := range subRunes {
			if lowerRune(sRunes[i+j]) != lowerRune(subRunes[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func lowerRune(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

func recomputeTraceSummary(trace *events.Trace) {
	if trace == nil {
		return
	}
	trace.Summary.EventCount = len(trace.Events)
	trace.Summary.ArtifactCount = len(trace.Artifacts)
	trace.Summary.ReasoningStepCount = len(trace.Reasoning)
	trace.Summary.ToolCallCount = len(trace.ToolCalls)
	trace.Summary.SlackActionCount = len(trace.SlackActions)

	latest := trace.Summary.StartedAt
	for _, event := range trace.Events {
		if event.StartedAt.After(latest) {
			latest = event.StartedAt
		}
		if event.EndedAt != nil && event.EndedAt.After(latest) {
			latest = *event.EndedAt
		}
	}
	for _, item := range trace.Reasoning {
		if item.CreatedAt.After(latest) {
			latest = item.CreatedAt
		}
	}
	for _, item := range trace.ToolCalls {
		if item.CreatedAt.After(latest) {
			latest = item.CreatedAt
		}
	}
	for _, item := range trace.SlackActions {
		if item.CreatedAt.After(latest) {
			latest = item.CreatedAt
		}
	}
	if latest.IsZero() {
		latest = time.Now().UTC()
	}
	trace.Summary.EndedAt = latest
}

func normalizedSettings(settings improvement.Settings) improvement.Settings {
	if settings.ActiveProposalCap <= 0 {
		settings.ActiveProposalCap = defaultProposalSlotCap
	}
	if settings.UpdatedAt.IsZero() {
		settings.UpdatedAt = time.Now().UTC()
	}
	return settings
}

func proposalRepo(proposal review.Proposal) string {
	scope := strings.TrimSpace(strings.ToLower(proposal.ProposedScope))
	switch {
	case strings.Contains(scope, "story-deployments"):
		return "story-deployments"
	case strings.Contains(scope, "story-infra-aws"):
		return "story-infra-aws"
	case strings.Contains(scope, "story-api"):
		return "story-api"
	case strings.Contains(scope, "story-orchestration-service"):
		return "story-orchestration-service"
	case strings.Contains(scope, "depin-backend"):
		return "depin-backend"
	default:
		return "rsi-agent-platform"
	}
}

func appendUnique(existing []string, values ...string) []string {
	seen := map[string]struct{}{}
	for _, value := range existing {
		seen[value] = struct{}{}
	}
	out := append([]string(nil), existing...)
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
