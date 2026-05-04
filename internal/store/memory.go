package store

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/ingestion"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/registry"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/slack"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

const (
	defaultProposalSlotCap = 2
	proposalReviewSLA      = 24 * time.Hour
	proposalPromoterLease  = 5 * time.Minute
)

type Store interface {
	CompanyWikiStore
	ListEvents() []ingestion.EventEnvelope
	ListConversations() []conversation.Conversation
	GetConversation(conversationID string) (conversation.Conversation, bool)
	ListConversationEntries(conversationID string) []conversation.Entry
	ListCases() []conversation.Case
	GetCase(caseID string) (conversation.Case, bool)
	ListActionIntents() []action.Intent
	GetActionIntent(actionID string) (action.Intent, bool)
	ListActionResults(actionIntentID string) []action.Result
	ListDomainEvents() []transition.DomainEvent
	ListEffectExecutions() []transition.EffectExecution
	ListEffectExecutionsByAggregate(machineKind transition.MachineKind, aggregateID string) []transition.EffectExecution
	GetCommandReceipt(commandID string) (transition.CommandReceipt, bool)
	RecordCommandReceipt(item transition.CommandReceipt) (transition.CommandReceipt, bool, error)
	SubmitCommand(command transition.CommandEnvelope) (transition.CommandReceipt, error)
	SubmitCommandContext(ctx context.Context, command transition.CommandEnvelope) (transition.CommandReceipt, error)
	QueueEffectExecution(effect transition.EffectExecution) (transition.EffectExecution, bool, error)
	ClaimEffectExecution(effectID string, holder string, lease time.Duration) (transition.EffectExecution, bool, error)
	ClaimNextEffectExecution(holder string, lease time.Duration, queueNames []string, maxPerScope int) (transition.EffectExecution, bool, error)
	ClaimNextEffectExecutionForKinds(holder string, lease time.Duration, queueNames []string, maxPerScope int, selectors []EffectClaimSelector) (transition.EffectExecution, bool, error)
	DeferEffectExecution(effectID string, holder string, lease time.Duration, reason string) (transition.EffectExecution, error)
	CompleteEffectExecution(effectID string, holder string, resultRef string) (transition.EffectExecution, error)
	FailEffectExecution(effectID string, holder string, lastError string) (transition.EffectExecution, error)
	ListRunnerExecutions() []RunnerExecution
	ListActiveRunnerExecutions() []RunnerExecution
	GetRunnerExecution(executionID string) (RunnerExecution, bool)
	RecordRunnerExecution(item RunnerExecution) (RunnerExecution, error)
	RecordRunnerExecutionWithHolderCAS(item RunnerExecution, expectedOldHolder string, expectedHeartbeatAt *time.Time) (RunnerExecution, error)
	CancelRunnerExecutionsForCase(caseID string, exceptTraceID string, reason string) []RunnerExecution
	ReconcileWorkflowTrace(workflowID string) (events.Trace, bool, error)
	ListOutcomes() []outcome.Record
	ListKnowledgeEntries() []knowledge.Entry
	GetKnowledgeEntry(knowledgeID string) (knowledge.Entry, bool)
	ListKnowledgeEvidenceLinks(knowledgeID string) []knowledge.EvidenceLink
	ListKnowledgeReviews(knowledgeID string) []knowledge.Review
	ListHarnessProfiles() []harness.Profile
	GetHarnessProfile(profileID string) (harness.Profile, bool)
	ListHarnessOverlays() []harness.Overlay
	GetActiveHarnessOverlay(role string) (harness.Overlay, bool)
	ListHarnessExperiments() []harness.Experiment
	ListHarnessSessionBindings() []harness.SessionBinding
	ListHarnessExecutions() []harness.Execution
	ListHarnessExecutionObservations() []harness.ExecutionObservation
	RecordHarnessExecutionObservation(item harness.ExecutionObservation) (harness.ExecutionObservation, error)
	ListExecutionLedgerEvents() []events.ExecutionLedgerEvent
	ListExecutionLedgerEventsByTrace(traceID string) []events.ExecutionLedgerEvent
	RecordExecutionLedgerEvents(items []events.ExecutionLedgerEvent) error
	ListChangeAttempts() []improvement.ChangeAttempt
	GetChangeAttempt(attemptID string) (improvement.ChangeAttempt, bool)
	ListAttemptWorkspaces() []improvement.AttemptWorkspace
	RecordAttemptWorkspace(workspace improvement.AttemptWorkspace) (improvement.AttemptWorkspace, error)
	GetAttemptWorkspace(workspaceID string) (improvement.AttemptWorkspace, bool)
	GetAttemptWorkspaceByAttempt(attemptID string) (improvement.AttemptWorkspace, bool)
	ListIngestions() []slack.Ingestion
	ListWorkflowLines() []WorkflowLine
	GetWorkflowLine(caseID string) (WorkflowLine, bool)
	ListWorkflows() []Workflow
	ListAssignments() []Assignment
	ListThreadPolicies() []policy.ThreadPolicy
	ListChannelPolicies() []policy.ChannelPolicy
	ListOwnershipRecords() []registry.OwnershipRecord
	ListCapabilities() []registry.CapabilityRecord
	ListTemplates() []registry.WorkflowTemplate
	ListExperiments() []registry.ExperimentRecord
	ListTraces() []events.TraceSummary
	GetTrace(traceID string) (events.Trace, bool)
	ListRatings(traceID string) []review.HumanRating
	ListImprovementNotes(traceID string) []review.ImprovementNote
	ListFeedback(traceID string) []review.FeedbackRecord
	ListEvalSuites() []evals.Suite
	ListEvalRuns() []evals.Run
	ListEvalJudgments(evalRunID string) []evals.Judgment
	GetSettings() improvement.Settings
	ListCandidates() []improvement.Candidate
	ListRuntimeDiagnoses() []improvement.RuntimeDiagnosis
	GetRuntimeDiagnosis(diagnosisID string) (improvement.RuntimeDiagnosis, bool)
	ListProposalMemories() []review.ProposalMemory
	GetProposalSlots() ProposalSlotState
	ListProposals() []review.Proposal
	ListValidationRuns() []improvement.ValidationRun
	RecordValidationRun(run improvement.ValidationRun) (improvement.ValidationRun, error)
	ListRepoChangeJobs() []improvement.RepoChangeJob
	ListPRAttempts() []improvement.PRAttempt
	RecordPRAttempt(attempt improvement.PRAttempt) (improvement.PRAttempt, error)
	ListPostMergeReplays() []improvement.PostMergeReplay
	ResetAppData() (AppDataResetResult, error)
}

type MemoryStore struct {
	mu                                contextRWMutex
	events                            []ingestion.EventEnvelope
	sourceMirrorRecords               map[string]SourceMirrorRecord
	companyWikiSourceDocuments        map[string]CompanyWikiSourceDocument
	companyWikiSourceRevisions        map[string]CompanyWikiSourceRevision
	companyWikiSourceChunks           map[string]CompanyWikiSourceChunk
	companyWikiSourceChunksByRevision map[string][]CompanyWikiSourceChunk
	companyWikiPages                  map[string]CompanyWikiPage
	companyWikiPageBySlug             map[string]string
	companyWikiRevisions              map[string]CompanyWikiRevision
	companyWikiCitations              map[string]CompanyWikiCitation
	companyWikiManifest               map[string]CompanyWikiManifestEntry
	companyWikiAudits                 map[string]CompanyWikiAuditRecord
	companyWikiAuditByIdempotency     map[string]string
	conversations                     map[string]conversation.Conversation
	conversationEntries               []conversation.Entry
	cases                             map[string]conversation.Case
	ingestions                        []slack.Ingestion
	workflowLines                     map[string]WorkflowLine
	workflows                         []Workflow
	assignments                       []Assignment
	threadPolicies                    map[string]policy.ThreadPolicy
	channelPolicy                     []policy.ChannelPolicy
	ownership                         []registry.OwnershipRecord
	capabilities                      []registry.CapabilityRecord
	templates                         []registry.WorkflowTemplate
	experiments                       []registry.ExperimentRecord
	traces                            map[string]events.Trace
	ratings                           map[string][]review.HumanRating
	notes                             map[string][]review.ImprovementNote
	feedbackRecords                   map[string][]review.FeedbackRecord
	actionIntents                     map[string]action.Intent
	actionResults                     map[string][]action.Result
	domainEvents                      []transition.DomainEvent
	effectExecutions                  map[string]transition.EffectExecution
	runnerExecutions                  map[string]RunnerExecution
	commandReceipts                   map[string]transition.CommandReceipt
	outcomes                          map[string]outcome.Record
	knowledgeEntries                  map[string]knowledge.Entry
	knowledgeEvidence                 map[string][]knowledge.EvidenceLink
	knowledgeReviews                  map[string][]knowledge.Review
	harnessProfiles                   map[string]harness.Profile
	harnessOverlays                   map[string]harness.Overlay
	harnessExperiments                map[string]harness.Experiment
	harnessSessionBindings            map[string]harness.SessionBinding
	harnessExecutions                 []harness.Execution
	harnessExecutionObservations      []harness.ExecutionObservation
	executionLedgerEvents             []events.ExecutionLedgerEvent
	evalSuites                        []evals.Suite
	evalRuns                          map[string]evals.Run
	evalJudgments                     map[string][]evals.Judgment
	candidates                        map[string]improvement.Candidate
	runtimeDiagnoses                  map[string]improvement.RuntimeDiagnosis
	proposals                         map[string]review.Proposal
	changeAttempts                    map[string]improvement.ChangeAttempt
	attemptWorkspaces                 map[string]improvement.AttemptWorkspace
	validationRuns                    map[string]improvement.ValidationRun
	proposalMemory                    []review.ProposalMemory
	repoChangeJobs                    map[string]improvement.RepoChangeJob
	prAttempts                        map[string]improvement.PRAttempt
	postMergeReplay                   map[string]improvement.PostMergeReplay
	cronLeases                        map[string]improvement.CronLease
	settings                          improvement.Settings
}

func NewMemoryStore() *MemoryStore {
	s := newEmptyMemoryStore()
	s.seedDefaults()
	return s
}

func (s *MemoryStore) ResetAppData() (AppDataResetResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	replacement := newEmptyMemoryStore()
	s.events = replacement.events
	s.sourceMirrorRecords = replacement.sourceMirrorRecords
	s.conversations = replacement.conversations
	s.conversationEntries = replacement.conversationEntries
	s.cases = replacement.cases
	s.ingestions = replacement.ingestions
	s.workflowLines = replacement.workflowLines
	s.workflows = replacement.workflows
	s.assignments = replacement.assignments
	s.threadPolicies = replacement.threadPolicies
	s.channelPolicy = replacement.channelPolicy
	s.ownership = replacement.ownership
	s.capabilities = replacement.capabilities
	s.templates = replacement.templates
	s.experiments = replacement.experiments
	s.traces = replacement.traces
	s.ratings = replacement.ratings
	s.notes = replacement.notes
	s.feedbackRecords = replacement.feedbackRecords
	s.actionIntents = replacement.actionIntents
	s.actionResults = replacement.actionResults
	s.domainEvents = replacement.domainEvents
	s.effectExecutions = replacement.effectExecutions
	s.commandReceipts = replacement.commandReceipts
	s.outcomes = replacement.outcomes
	s.knowledgeEntries = replacement.knowledgeEntries
	s.knowledgeEvidence = replacement.knowledgeEvidence
	s.knowledgeReviews = replacement.knowledgeReviews
	s.harnessProfiles = replacement.harnessProfiles
	s.harnessOverlays = replacement.harnessOverlays
	s.harnessExperiments = replacement.harnessExperiments
	s.harnessSessionBindings = replacement.harnessSessionBindings
	s.harnessExecutions = replacement.harnessExecutions
	s.harnessExecutionObservations = replacement.harnessExecutionObservations
	s.executionLedgerEvents = replacement.executionLedgerEvents
	s.evalSuites = replacement.evalSuites
	s.evalRuns = replacement.evalRuns
	s.evalJudgments = replacement.evalJudgments
	s.candidates = replacement.candidates
	s.proposals = replacement.proposals
	s.changeAttempts = replacement.changeAttempts
	s.attemptWorkspaces = replacement.attemptWorkspaces
	s.validationRuns = replacement.validationRuns
	s.proposalMemory = replacement.proposalMemory
	s.repoChangeJobs = replacement.repoChangeJobs
	s.prAttempts = replacement.prAttempts
	s.postMergeReplay = replacement.postMergeReplay
	s.cronLeases = replacement.cronLeases
	s.settings = replacement.settings

	return AppDataResetResult{
		Backend:         "memory",
		ResetAt:         time.Now().UTC(),
		TruncatedTables: []string{"memory_store"},
		PreservedTables: []string{},
	}, nil
}

func (s *MemoryStore) seedDefaults() {
	now := time.Now().UTC()
	for _, profile := range harness.SeedProfiles(now) {
		s.harnessProfiles[profile.ID] = profile
	}
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
		{Name: "github.cli_credentials", Kind: "credential", AllowedBots: []string{"oncall", "fr", "arch"}, ApprovalNeeded: false},
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

	_, _ = s.createEventLocked(ingestion.EventEnvelope{
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
	_, _ = s.createEventLocked(ingestion.EventEnvelope{
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
		_, _ = s.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineProblemLine,
			AggregateID: trace.TraceID,
			CommandKind: string(transition.CommandProblemLineEvaluateTrace),
			CommandID:   fmt.Sprintf("cmd-seed-evaluate:%s", trace.TraceID),
			Actor:       "seed",
			OccurredAt:  now,
			Payload: map[string]any{
				"trigger": "seed",
			},
		})
	}
	_, _ = s.promoteCandidatesLocked("seed", 1)
}

func (s *MemoryStore) ListEvents() []ingestion.EventEnvelope {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]ingestion.EventEnvelope(nil), s.events...)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
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
	skipWorkflowMaterialization := boolFromMetadata(event.Metadata, "skip_workflow_materialization")
	if event.WorkflowHint == "" && !skipWorkflowMaterialization {
		event.WorkflowHint = deriveWorkflowHint(event.NormalizedProblemStatement)
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}
	s.events = append(s.events, event)
	if outcomeRecord, ok := s.outcomeFromEventLocked(event); ok {
		if _, err := s.recordOutcomeLocked(outcomeRecord); err != nil {
			return ingestion.EventEnvelope{}, err
		}
		return event, nil
	}
	if skipWorkflowMaterialization || strings.TrimSpace(event.WorkflowHint) == "" {
		return event, nil
	}
	s.materializeWorkflowLocked(event)
	return event, nil
}

func (s *MemoryStore) ListConversations() []conversation.Conversation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]conversation.Conversation, 0, len(s.conversations))
	for _, item := range s.conversations {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func (s *MemoryStore) GetConversation(conversationID string) (conversation.Conversation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.conversations[conversationID]
	return item, ok
}

func (s *MemoryStore) ListConversationEntries(conversationID string) []conversation.Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]conversation.Entry, 0)
	for _, item := range s.conversationEntries {
		if item.ConversationID == conversationID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) ListConversationEntriesPage(conversationID string, opts ConversationEntryPageOptions) ConversationEntryPage {
	entries := s.ListConversationEntries(conversationID)
	limit := opts.Limit
	if limit <= 0 {
		return ConversationEntryPage{Entries: entries, Limit: limit, HasMore: false}
	}
	hasMore := len(entries) > limit
	if hasMore {
		entries = entries[:limit]
	}
	return ConversationEntryPage{Entries: entries, Limit: limit, HasMore: hasMore}
}

func (s *MemoryStore) ListCases() []conversation.Case {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]conversation.Case, 0, len(s.cases))
	for _, item := range s.cases {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func (s *MemoryStore) GetCase(caseID string) (conversation.Case, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.cases[caseID]
	return item, ok
}

func (s *MemoryStore) ListActionIntents() []action.Intent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]action.Intent, 0, len(s.actionIntents))
	for _, item := range s.actionIntents {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) GetActionIntent(actionID string) (action.Intent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.actionIntents[actionID]
	return item, ok
}

func (s *MemoryStore) upsertActionIntentLocked(intent action.Intent) (action.Intent, error) {
	now := time.Now().UTC()
	if intent.ID == "" {
		intent.ID = nextID("action", len(s.actionIntents)+1)
	}
	if intent.Status == "" {
		intent.Status = action.StatusDrafted
	}
	if intent.CreatedAt.IsZero() {
		intent.CreatedAt = now
	}
	intent.UpdatedAt = now
	intent.EvidenceRefs = normalizeEvidenceRefs(intent.EvidenceRefs)
	intent.RequestPayload = cloneMetadata(intent.RequestPayload)
	s.actionIntents[intent.ID] = intent
	return intent, nil
}

func (s *MemoryStore) ListActionResults(actionIntentID string) []action.Result {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]action.Result(nil), s.actionResults[actionIntentID]...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].AttemptNumber == out[j].AttemptNumber {
			return out[i].StartedAt.Before(out[j].StartedAt)
		}
		return out[i].AttemptNumber < out[j].AttemptNumber
	})
	return out
}

func (s *MemoryStore) recordActionResultLocked(result action.Result) (action.Result, error) {
	if strings.TrimSpace(result.ActionIntentID) == "" {
		return action.Result{}, errors.New("action_intent_id is required")
	}
	intent, ok := s.actionIntents[result.ActionIntentID]
	if !ok {
		return action.Result{}, errors.New("action intent not found")
	}
	now := time.Now().UTC()
	if strings.TrimSpace(result.OperationID) != "" {
		for _, existing := range s.actionResults[result.ActionIntentID] {
			if strings.TrimSpace(existing.OperationID) == result.OperationID {
				return existing, nil
			}
		}
	}
	if result.ID == "" {
		result.ID = nextUUID("ares")
	}
	if result.AttemptNumber == 0 {
		result.AttemptNumber = len(s.actionResults[result.ActionIntentID]) + 1
	}
	if result.StartedAt.IsZero() {
		result.StartedAt = now
	}
	if result.CompletedAt.IsZero() {
		result.CompletedAt = now
	}
	s.actionResults[result.ActionIntentID] = append(s.actionResults[result.ActionIntentID], result)
	intent.UpdatedAt = result.CompletedAt
	switch result.Status {
	case action.StatusSucceeded:
		intent.Status = action.StatusSucceeded
	case action.StatusBlocked:
		intent.Status = action.StatusBlocked
	case action.StatusCanceled:
		intent.Status = action.StatusCanceled
	case action.StatusSuperseded:
		intent.Status = action.StatusSuperseded
	default:
		intent.Status = action.StatusFailed
	}
	s.actionIntents[intent.ID] = intent
	return result, nil
}

func (s *MemoryStore) ListOutcomes() []outcome.Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]outcome.Record, 0, len(s.outcomes))
	for _, item := range s.outcomes {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RecordedAt.After(out[j].RecordedAt) })
	return out
}

func (s *MemoryStore) recordOutcomeLocked(record outcome.Record) (outcome.Record, error) {
	now := time.Now().UTC()
	if strings.TrimSpace(record.OperationID) != "" {
		for _, existing := range s.outcomes {
			if strings.TrimSpace(existing.OperationID) == record.OperationID {
				return existing, nil
			}
		}
	}
	if record.ID == "" {
		record.ID = nextUUID("outcome")
	}
	if record.RecordedAt.IsZero() {
		record.RecordedAt = now
	}
	s.outcomes[record.ID] = record
	if record.CaseID != "" {
		caseRecord, ok := s.cases[record.CaseID]
		if ok {
			caseRecord.LatestOutcomeID = record.ID
			caseRecord.OutcomeScore = record.Score
			caseRecord.UpdatedAt = record.RecordedAt
			switch record.Verdict {
			case outcome.VerdictPositive:
				caseRecord.ResolutionState = conversation.ResolutionResolved
				caseRecord.Status = conversation.CaseResolved
				caseRecord.ResolvedAt = timePtr(record.RecordedAt)
			case outcome.VerdictNegative:
				caseRecord.ResolutionState = conversation.ResolutionRegressed
			case outcome.VerdictMixed:
				caseRecord.ResolutionState = conversation.ResolutionMonitoring
			default:
				caseRecord.ResolutionState = conversation.ResolutionUnresolved
			}
			s.cases[caseRecord.ID] = caseRecord
		}
	}
	if record.ProposalID != "" {
		if proposal, ok := s.proposals[record.ProposalID]; ok && record.OutcomeType == outcome.TypeProposalEffectiveness {
			if record.Verdict == outcome.VerdictPositive {
				proposal.NewEvidenceSinceLastRejection = true
			}
			s.proposals[proposal.ID] = proposal
		}
	}
	return record, nil
}

func (s *MemoryStore) ListKnowledgeEntries() []knowledge.Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]knowledge.Entry, 0, len(s.knowledgeEntries))
	for _, item := range s.knowledgeEntries {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func (s *MemoryStore) GetKnowledgeEntry(knowledgeID string) (knowledge.Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.knowledgeEntries[knowledgeID]
	return item, ok
}

func (s *MemoryStore) upsertKnowledgeEntryLocked(entry knowledge.Entry, links []knowledge.EvidenceLink) (knowledge.Entry, error) {
	now := time.Now().UTC()
	if entry.ID == "" {
		entry.ID = nextID("knowledge", len(s.knowledgeEntries)+1)
	}
	if entry.Tier == "" {
		entry.Tier = knowledge.TierWorking
	}
	if entry.Status == "" {
		entry.Status = knowledge.StatusDraft
	}
	if entry.SourceType == "" {
		entry.SourceType = knowledge.SourceAgent
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now
	entry.StructuredFacts = cloneMetadata(entry.StructuredFacts)
	s.knowledgeEntries[entry.ID] = entry
	normalized := make([]knowledge.EvidenceLink, 0, len(links))
	for _, link := range links {
		link.KnowledgeEntryID = entry.ID
		normalized = append(normalized, link)
	}
	if len(normalized) > 0 {
		s.knowledgeEvidence[entry.ID] = normalized
	}
	return entry, nil
}

func (s *MemoryStore) ListKnowledgeEvidenceLinks(knowledgeID string) []knowledge.EvidenceLink {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]knowledge.EvidenceLink(nil), s.knowledgeEvidence[knowledgeID]...)
}

func (s *MemoryStore) ListKnowledgeReviews(knowledgeID string) []knowledge.Review {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]knowledge.Review(nil), s.knowledgeReviews[knowledgeID]...)
}

func (s *MemoryStore) ListIngestions() []slack.Ingestion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]slack.Ingestion(nil), s.ingestions...)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) createIngestionLocked(envelope slack.SlackEnvelope) (slack.Ingestion, error) {
	conversationKey := slackConversationKey(envelope)
	prompt := slack.PromptEnvelopeFromValue(envelope.Prompt)
	normalizedText := firstNonEmpty(prompt.RenderedText, strings.TrimSpace(envelope.Text))
	event := ingestion.EventEnvelope{
		Source:                     ingestion.SourceSlack,
		SourceEventID:              envelope.TS,
		ThreadKey:                  conversationKey,
		DedupeKey:                  envelope.TS,
		Severity:                   severityFromText(normalizedText),
		NormalizedProblemStatement: normalizedText,
		OwnershipHint:              "platform",
		WorkflowHint:               deriveWorkflowHint(normalizedText),
		Metadata: map[string]interface{}{
			"team_id":             envelope.TeamID,
			"channel_id":          envelope.ChannelID,
			"user_id":             envelope.UserID,
			"thread_ts":           envelope.ThreadTS,
			"action_token":        envelope.ActionToken,
			"conversation_key":    conversationKey,
			"bot_role":            envelope.BotRole,
			"files":               envelope.Files,
			"entity_refs":         envelope.EntityRefs,
			"prompt_envelope":     prompt,
			"slack_user_names":    slack.PromptEnvelopeUserNames(prompt),
			"slack_channel_names": slack.PromptEnvelopeChannelNames(prompt),
		},
		CreatedAt: envelope.CreatedAt,
	}
	event.RawPayloadRef = fmt.Sprintf("memory://slack/%s/%s.json", envelope.ChannelID, strings.ReplaceAll(envelope.TS, ".", "-"))
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	created, err := s.createEventLocked(event)
	if err != nil {
		return slack.Ingestion{}, err
	}
	for _, item := range s.ingestions {
		if item.EventID == created.ID {
			return item, nil
		}
	}
	return slack.Ingestion{}, errors.New("ingestion materialization did not produce an ingestion row")
}

func (s *MemoryStore) outcomeFromEventLocked(event ingestion.EventEnvelope) (outcome.Record, bool) {
	now := event.CreatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if event.Source == ingestion.SourceGitHub {
		githubOutcome, ok := githubOutcomeFromMetadata(event.Metadata)
		if !ok {
			return outcome.Record{}, false
		}
		proposalID := firstNonEmpty(stringFromMetadata(event.Metadata, "proposal_id"), githubOutcome.ProposalID)
		if proposalID == "" {
			return outcome.Record{}, false
		}
		proposal, ok := s.proposals[proposalID]
		if !ok {
			return outcome.Record{}, false
		}
		attemptID := strings.TrimSpace(stringFromMetadata(event.Metadata, "attempt_id"))
		traceID := firstNonEmpty(proposal.OriginTraceID, proposal.TraceID)
		if attemptID != "" {
			if attempt, ok := s.changeAttempts[attemptID]; ok {
				traceID = firstNonEmpty(attempt.AttemptTraceID, traceID)
			}
		}
		if conv, ok := s.conversations[proposal.ConversationID]; ok {
			conv.LatestEventID = event.ID
			conv.UpdatedAt = now
			s.conversations[conv.ID] = conv
			_ = s.appendConversationEntryLocked(conv.ID, event, now)
		}
		return outcome.Record{
			Source:         string(event.Source),
			SourceEventID:  event.SourceEventID,
			ConversationID: proposal.ConversationID,
			CaseID:         proposal.CaseID,
			TraceID:        traceID,
			ProposalID:     proposalID,
			AttemptID:      attemptID,
			OutcomeType:    githubOutcome.OutcomeType,
			Verdict:        githubOutcome.Verdict,
			Score:          githubOutcome.Score,
			Summary:        githubOutcome.Summary,
			Details:        githubOutcome.Details,
			ExternalRef:    firstNonEmpty(stringFromMetadata(event.Metadata, "html_url"), stringFromMetadata(event.Metadata, "pr_url")),
			RecordedBy:     firstNonEmpty(stringFromMetadata(event.Metadata, "sender_login"), stringFromMetadata(event.Metadata, "user_id")),
			RecordedAt:     now,
		}, true
	}
	conversationKey := conversationKeyForEvent(event)
	conv, _ := s.resolveConversationLocked(event, now)
	conv.LatestEventID = event.ID
	conv.UpdatedAt = now
	s.conversations[conv.ID] = conv
	entry := s.appendConversationEntryLocked(conv.ID, event, now)
	caseRecord, _ := s.activeCaseForConversationLocked(conv.ID)
	base := outcome.Record{
		Source:         string(event.Source),
		SourceEventID:  event.SourceEventID,
		ConversationID: conv.ID,
		CaseID:         caseRecord.ID,
		TraceID:        caseRecord.LatestTraceID,
		RecordedBy:     stringFromMetadata(event.Metadata, "user_id"),
		RecordedAt:     now,
		ExternalRef:    conversationKey,
	}
	if event.Source == ingestion.SourceSlack {
		if slackOutcome, ok := slackOutcomeFromText(event.NormalizedProblemStatement); ok {
			base.OutcomeType = slackOutcome.OutcomeType
			base.Verdict = slackOutcome.Verdict
			base.Score = slackOutcome.Score
			base.Summary = slackOutcome.Summary
			base.Details = fmt.Sprintf("Slack outcome via conversation entry %s.", entry.ID)
			return base, true
		}
	}
	return outcome.Record{}, false
}

func (s *MemoryStore) materializeWorkflowLocked(event ingestion.EventEnvelope) {
	createdAt := event.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	intent := intentForWorkflowHint(event.WorkflowHint)
	assignedBot := assignedBotFor(event.WorkflowHint)
	channelID, _ := event.Metadata["channel_id"].(string)
	userID, _ := event.Metadata["user_id"].(string)
	threadTS, _ := event.Metadata["thread_ts"].(string)
	if raw, ok := event.Metadata["bot_role"].(string); ok && raw != "" {
		assignedBot = raw
	}
	if raw, ok := event.Metadata["bot_role"].(slack.BotRole); ok && raw != "" {
		assignedBot = string(raw)
	}
	if channelID == "" {
		channelID = string(event.Source)
	}
	if userID == "" {
		userID = "system"
	}
	if threadTS == "" && !strings.HasPrefix(channelID, "D") {
		threadTS = event.SourceEventID
	}

	conv, createdConversation := s.resolveConversationLocked(event, createdAt)
	entry := s.appendConversationEntryLocked(conv.ID, event, createdAt)
	caseRecord, _ := s.resolveCaseLocked(conv, event, createdAt)
	line := s.ensureWorkflowLineLocked(caseRecord, createdAt)

	ingestionID := firstNonEmpty(stringFromMetadata(event.Metadata, "ingestion_id"), nextID("ing", len(s.ingestions)+1))
	ingestionItem := slack.Ingestion{
		ID:             ingestionID,
		EventID:        event.ID,
		ConversationID: conv.ID,
		CaseID:         caseRecord.ID,
		ThreadKey:      conv.ExternalKey,
		ThreadTS:       threadTS,
		WorkflowHint:   event.WorkflowHint,
		Intent:         intent,
		BotRole:        slack.BotRole(assignedBot),
		Source:         string(event.Source),
		ChannelID:      channelID,
		UserID:         userID,
		Text:           event.NormalizedProblemStatement,
		EntityRefs:     slack.EntityRefsFromValue(event.Metadata["entity_refs"]),
		Prompt:         slack.PromptEnvelopeFromValue(event.Metadata["prompt_envelope"]),
		CreatedAt:      createdAt,
	}
	s.ingestions = append(s.ingestions, ingestionItem)

	s.assignments = append(s.assignments, Assignment{
		ID:             nextID("as", len(s.assignments)+1),
		ConversationID: conv.ID,
		CaseID:         caseRecord.ID,
		ThreadKey:      conv.ExternalKey,
		AssignedBot:    assignedBot,
		Confidence:     routeConfidenceForEvent(event),
		Rationale:      routingRationale(event),
		CreatedAt:      createdAt,
	})
	if _, ok := s.threadPolicies[conv.ExternalKey]; !ok {
		s.threadPolicies[conv.ExternalKey] = policy.ThreadPolicy{
			ThreadKey:         conv.ExternalKey,
			State:             policy.ThreadStateActive,
			OwnerBot:          assignedBot,
			LastPolicyVersion: "conversation-v2",
			UpdatedAt:         createdAt,
		}
	}

	traceID := nextID("trace", len(s.traces)+1)
	workflowID := firstNonEmpty(stringFromMetadata(event.Metadata, "workflow_id"), nextID("wf", len(s.workflows)+1))
	parentWorkflowID := firstNonEmpty(line.CurrentWorkflowID, line.LatestWorkflowID)
	supersedesTraceID := s.supersedeInFlightTracesLocked(caseRecord.ID, traceID, event.ID, createdAt)
	workflow := Workflow{
		ID:               workflowID,
		IngestionID:      ingestionID,
		TraceID:          traceID,
		ConversationID:   caseRecord.ConversationID,
		CaseID:           caseRecord.ID,
		ThreadKey:        conversationKeyForCase(caseRecord, s.conversations),
		Kind:             caseRecord.Kind,
		Intent:           caseRecord.Intent,
		AssignedBot:      caseRecord.AssignedBot,
		ApprovalMode:     caseRecord.ApprovalMode,
		ResponseMode:     caseRecord.ResponseMode,
		Status:           "queued",
		AttemptNumber:    line.AttemptCount + 1,
		ParentWorkflowID: parentWorkflowID,
		Version:          1,
		CreatedAt:        entry.CreatedAt,
		UpdatedAt:        createdAt,
	}
	line.ConversationID = conv.ID
	line.Status = workflowLineStatusFromState(transition.WorkflowLineStateActive)
	line.CurrentWorkflowID = workflow.ID
	line.LatestWorkflowID = workflow.ID
	line.AttemptCount++
	line.AutoRetryBudgetRemaining = workflowLineRetryBudgetRemaining(line.AttemptCount)
	line.LastFailureClass = ""
	line.NextRetryAction = ""
	line.RetryAfter = nil
	line.LineStopReason = ""
	line.UpdatedAt = createdAt
	line.CompletedAt = nil
	s.upsertWorkflowLineLocked(line)
	s.upsertWorkflowLocked(workflow)

	caseRecord.LatestTraceID = traceID
	caseRecord.UpdatedAt = createdAt
	s.cases[caseRecord.ID] = caseRecord
	conv.ActiveCaseID = caseRecord.ID
	conv.LatestEventID = event.ID
	conv.UpdatedAt = createdAt
	if createdConversation && conv.Title == "" {
		conv.Title = conversation.NormalizeTitle(caseRecord.Kind, event.NormalizedProblemStatement)
	}
	s.conversations[conv.ID] = conv

	traceStatus := events.StatusQueued
	trace := events.Trace{
		Summary: events.TraceSummary{
			TraceID:           traceID,
			IngestionID:       ingestionID,
			WorkflowID:        workflow.ID,
			ConversationID:    conv.ID,
			CaseID:            caseRecord.ID,
			TriggerEventID:    event.ID,
			SupersedesTraceID: supersedesTraceID,
			ThreadKey:         conv.ExternalKey,
			WorkflowKind:      caseRecord.Kind,
			Status:            traceStatus,
			StartedAt:         createdAt,
			EndedAt:           createdAt,
		},
		Events: []events.TraceEvent{
			{
				TraceID:        traceID,
				IngestionID:    ingestionID,
				WorkflowID:     workflow.ID,
				ConversationID: conv.ID,
				CaseID:         caseRecord.ID,
				TriggerEventID: event.ID,
				Plane:          "edge",
				Service:        "control-plane",
				Actor:          "event-ingestor",
				EventType:      "event.ingested",
				Status:         events.StatusCompleted,
				StartedAt:      createdAt,
				EndedAt:        timePtr(createdAt),
				PayloadRef:     event.RawPayloadRef,
				Description:    fmt.Sprintf("%s event normalized into conversation %s.", event.Source, conv.ExternalKey),
			},
			{
				TraceID:        traceID,
				IngestionID:    ingestionID,
				WorkflowID:     workflow.ID,
				ConversationID: conv.ID,
				CaseID:         caseRecord.ID,
				TriggerEventID: event.ID,
				Plane:          "control",
				Service:        "control-plane",
				Actor:          "router-policy",
				EventType:      "workflow.queued",
				Status:         events.StatusQueued,
				StartedAt:      createdAt,
				Description:    fmt.Sprintf("Queued %s trace for case %s and bot %s.", intent, caseRecord.ID, assignedBot),
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
				ID:             nextID("reason", len(s.traces)+1),
				TraceID:        traceID,
				WorkflowID:     workflow.ID,
				ConversationID: conv.ID,
				CaseID:         caseRecord.ID,
				StepType:       "goal_framing",
				Summary:        fmt.Sprintf("Treat event %s as %s within case %s.", event.ID, intent, caseRecord.ID),
				EvidenceRefs: []events.EvidenceRef{
					{Kind: "event", Ref: event.ID, Summary: event.NormalizedProblemStatement},
					{Kind: "conversation_entry", Ref: entry.ID, Summary: entry.Body},
				},
				Alternatives: []string{"incident", "feature_request", "question"},
				Confidence:   routeConfidenceForEvent(event),
				Decision:     fmt.Sprintf("route:%s assign:%s case:%s", intent, assignedBot, caseRecord.ID),
				CreatedAt:    createdAt,
			},
		},
	}
	recomputeTraceSummary(&trace)
	s.traces[traceID] = trace
}

func (s *MemoryStore) resolveConversationLocked(event ingestion.EventEnvelope, createdAt time.Time) (conversation.Conversation, bool) {
	externalKey := conversationKeyForEvent(event)
	for _, item := range s.conversations {
		if item.ExternalKey == externalKey {
			item.LatestEventID = event.ID
			item.UpdatedAt = createdAt
			item.Title = firstNonEmpty(item.Title, conversation.NormalizeTitle(event.WorkflowHint, event.NormalizedProblemStatement))
			item.ParticipantIDs = appendUnique(item.ParticipantIDs, stringFromMetadata(event.Metadata, "user_id"))
			s.conversations[item.ID] = item
			return item, false
		}
	}
	item := conversation.Conversation{
		ID:                   nextID("conv", len(s.conversations)+1),
		Source:               event.Source,
		ExternalKey:          externalKey,
		ExternalConversation: externalKey,
		Title:                conversation.NormalizeTitle(event.WorkflowHint, event.NormalizedProblemStatement),
		Status:               conversation.StatusActive,
		ParticipantIDs:       compactStrings([]string{stringFromMetadata(event.Metadata, "user_id")}),
		LatestEventID:        event.ID,
		CreatedAt:            createdAt,
		UpdatedAt:            createdAt,
	}
	s.conversations[item.ID] = item
	return item, true
}

func (s *MemoryStore) appendConversationEntryLocked(conversationID string, event ingestion.EventEnvelope, createdAt time.Time) conversation.Entry {
	return s.upsertConversationEntryLocked(externalConversationEntry(conversationID, event, createdAt))
}

func (s *MemoryStore) upsertConversationEntryLocked(entry conversation.Entry) conversation.Entry {
	for idx := range s.conversationEntries {
		if s.conversationEntries[idx].ID != entry.ID {
			continue
		}
		s.conversationEntries[idx] = entry
		return entry
	}
	s.conversationEntries = append(s.conversationEntries, entry)
	return entry
}

func (s *MemoryStore) resolveCaseLocked(conv conversation.Conversation, event ingestion.EventEnvelope, createdAt time.Time) (conversation.Case, bool) {
	if active, ok := s.activeCaseForConversationLocked(conv.ID); ok && active.Kind == event.WorkflowHint && active.Status == conversation.CaseActive {
		active.Summary = event.NormalizedProblemStatement
		active.UpdatedAt = createdAt
		active.OpenedByEventID = firstNonEmpty(active.OpenedByEventID, event.ID)
		s.cases[active.ID] = active
		return active, false
	}
	if active, ok := s.activeCaseForConversationLocked(conv.ID); ok {
		active.Status = conversation.CaseSuperseded
		active.ClosedByEventID = event.ID
		active.UpdatedAt = createdAt
		active.ClosedAt = timePtr(createdAt)
		s.cases[active.ID] = active
		if line, ok := s.workflowLines[active.ID]; ok {
			line.Status = workflowLineStatusFromState(transition.WorkflowLineStateSuperseded)
			line.LineStopReason = firstNonEmpty(line.LineStopReason, "case_superseded")
			line.UpdatedAt = createdAt
			line.CompletedAt = timePtr(createdAt)
			s.upsertWorkflowLineLocked(line)
		}
	}
	item := conversation.Case{
		ID:              nextID("case", len(s.cases)+1),
		ConversationID:  conv.ID,
		Kind:            event.WorkflowHint,
		Intent:          intentForWorkflowHint(event.WorkflowHint),
		Title:           conversation.NormalizeTitle(event.WorkflowHint, event.NormalizedProblemStatement),
		Summary:         event.NormalizedProblemStatement,
		Status:          conversation.CaseActive,
		ApprovalMode:    approvalModeForIntent(intentForWorkflowHint(event.WorkflowHint)),
		ResponseMode:    responseModeForIntent(intentForWorkflowHint(event.WorkflowHint)),
		AssignedBot:     assignedBotFor(event.WorkflowHint),
		OpenedByEventID: event.ID,
		ResolutionState: conversation.ResolutionUnresolved,
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt,
	}
	s.cases[item.ID] = item
	return item, true
}

func (s *MemoryStore) activeCaseForConversationLocked(conversationID string) (conversation.Case, bool) {
	if item, ok := s.conversations[conversationID]; ok && item.ActiveCaseID != "" {
		caseRecord, found := s.cases[item.ActiveCaseID]
		if found && caseRecord.Status == conversation.CaseActive {
			return caseRecord, true
		}
	}
	for _, item := range s.cases {
		if item.ConversationID == conversationID && item.Status == conversation.CaseActive {
			return item, true
		}
	}
	return conversation.Case{}, false
}

func (s *MemoryStore) ensureWorkflowForCaseLocked(caseRecord conversation.Case, entry conversation.Entry, createdAt time.Time, preferredWorkflowID string) Workflow {
	for i := range s.workflows {
		if s.workflows[i].CaseID != caseRecord.ID {
			continue
		}
		s.workflows[i].Version++
		s.workflows[i].ConversationID = caseRecord.ConversationID
		s.workflows[i].CaseID = caseRecord.ID
		s.workflows[i].ThreadKey = conversationKeyForCase(caseRecord, s.conversations)
		s.workflows[i].Kind = caseRecord.Kind
		s.workflows[i].Intent = caseRecord.Intent
		s.workflows[i].AssignedBot = caseRecord.AssignedBot
		s.workflows[i].ApprovalMode = caseRecord.ApprovalMode
		s.workflows[i].ResponseMode = caseRecord.ResponseMode
		s.workflows[i].UpdatedAt = createdAt
		return s.workflows[i]
	}
	item := Workflow{
		ID:             firstNonEmpty(preferredWorkflowID, nextID("wf", len(s.workflows)+1)),
		ConversationID: caseRecord.ConversationID,
		CaseID:         caseRecord.ID,
		ThreadKey:      conversationKeyForCase(caseRecord, s.conversations),
		Kind:           caseRecord.Kind,
		Intent:         caseRecord.Intent,
		AssignedBot:    caseRecord.AssignedBot,
		ApprovalMode:   caseRecord.ApprovalMode,
		ResponseMode:   caseRecord.ResponseMode,
		Status:         "queued",
		Version:        1,
		CreatedAt:      entry.CreatedAt,
		UpdatedAt:      createdAt,
	}
	s.workflows = append(s.workflows, item)
	return item
}

func (s *MemoryStore) upsertWorkflowLocked(item Workflow) {
	for i := range s.workflows {
		if s.workflows[i].ID == item.ID {
			if item.Version <= s.workflows[i].Version {
				item.Version = s.workflows[i].Version + 1
			}
			s.workflows[i] = item
			return
		}
	}
	if item.Version == 0 {
		item.Version = 1
	}
	s.workflows = append(s.workflows, item)
}

func (s *MemoryStore) supersedeInFlightTracesLocked(caseID string, nextTraceID string, triggerEventID string, createdAt time.Time) string {
	caseID = strings.TrimSpace(caseID)
	if caseID == "" {
		return ""
	}
	var supersedes string
	for traceID, trace := range s.traces {
		if trace.Summary.CaseID != caseID {
			continue
		}
		if trace.Summary.Status != events.StatusQueued && trace.Summary.Status != events.StatusRunning && trace.Summary.Status != events.StatusReplayed {
			continue
		}
		if supersedes == "" {
			supersedes = traceID
		}
		trace.Summary.Status = events.StatusSuppressed
		trace.Events = append(trace.Events, events.TraceEvent{
			TraceID:        traceID,
			IngestionID:    trace.Summary.IngestionID,
			WorkflowID:     trace.Summary.WorkflowID,
			ConversationID: trace.Summary.ConversationID,
			CaseID:         caseID,
			TriggerEventID: triggerEventID,
			Plane:          "control",
			Service:        "control-plane",
			Actor:          "supersession",
			EventType:      "trace.superseded",
			Status:         events.StatusSuppressed,
			StartedAt:      createdAt,
			EndedAt:        timePtr(createdAt),
			Description:    fmt.Sprintf("Superseded by newer trace %s for case %s.", nextTraceID, caseID),
		})
		recomputeTraceSummary(&trace)
		s.traces[traceID] = trace
		if updated, err := s.updateWorkflowStatusLocked(trace.Summary.WorkflowID, string(transition.WorkflowStateSuperseded), "superseded by successor trace"); err == nil {
			s.upsertWorkflowLocked(updated)
		}
	}
	for id, item := range s.actionIntents {
		if item.CaseID != caseID {
			continue
		}
		if item.Status == action.StatusSucceeded || item.Status == action.StatusFailed || item.Status == action.StatusBlocked || item.Status == action.StatusCanceled || item.Status == action.StatusSuperseded {
			continue
		}
		item.Status = action.StatusSuperseded
		item.SupersededByActionID = nextTraceID
		item.UpdatedAt = createdAt
		s.actionIntents[id] = item
		s.actionResults[id] = append(s.actionResults[id], action.Result{
			ID:             nextUUID("ares"),
			ActionIntentID: id,
			AttemptNumber:  len(s.actionResults[id]) + 1,
			Executor:       "supersession",
			Status:         action.StatusSuperseded,
			ErrorCode:      "trace_superseded",
			ErrorMessage:   fmt.Sprintf("Superseded by newer trace %s", nextTraceID),
			StartedAt:      createdAt,
			CompletedAt:    createdAt,
		})
	}
	runnerExecutionUpdates := map[string]RunnerExecution{}
	for id, item := range s.runnerExecutions {
		candidate := normalizeRunnerExecution(item)
		if candidate.CaseID != caseID || candidate.TraceID == nextTraceID || !runnerExecutionStatusActive(candidate.Status) {
			continue
		}
		if strings.ToLower(strings.TrimSpace(candidate.Status)) != "cancelling" {
			candidate.Status = "cancel_requested"
		}
		candidate.CancelRequested = true
		candidate.FailureClass = firstNonEmpty(candidate.FailureClass, "trace_superseded")
		candidate.UpdatedAt = createdAt
		runnerExecutionUpdates[id] = candidate
	}
	for id, item := range runnerExecutionUpdates {
		s.runnerExecutions[id] = item
	}
	return supersedes
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

func (s *MemoryStore) ListRatings(traceID string) []review.HumanRating {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]review.HumanRating(nil), s.ratings[traceID]...)
}

func (s *MemoryStore) ListImprovementNotes(traceID string) []review.ImprovementNote {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]review.ImprovementNote(nil), s.notes[traceID]...)
}

func (s *MemoryStore) ListFeedback(traceID string) []review.FeedbackRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]review.FeedbackRecord(nil), s.feedbackRecords[traceID]...)
}

func (s *MemoryStore) addRatingLocked(traceID string, rating review.HumanRating, createdAt time.Time) (review.HumanRating, error) {
	trace, ok := s.traces[traceID]
	if !ok {
		return review.HumanRating{}, errors.New("trace not found")
	}
	rating.TraceID = traceID
	rating.CreatedAt = createdAt
	s.ratings[traceID] = append(s.ratings[traceID], rating)
	trace.Summary.LastVerdict = rating.Verdict
	trace.Summary.Status = events.StatusInReview
	s.traces[traceID] = trace
	s.feedbackRecords[traceID] = append(s.feedbackRecords[traceID], review.FeedbackRecord{
		ID:             nextID("feedback", len(s.feedbackRecords[traceID])+1),
		ConversationID: trace.Summary.ConversationID,
		CaseID:         trace.Summary.CaseID,
		TraceID:        traceID,
		TargetType:     review.FeedbackTargetTrace,
		TargetID:       traceID,
		Score:          rating.Score,
		Verdict:        rating.Verdict,
		Labels:         append([]string(nil), rating.Labels...),
		Notes:          rating.Notes,
		ReviewerID:     rating.ReviewerID,
		CreatedAt:      rating.CreatedAt,
	})
	_, _, _ = s.evaluateTraceLocked(traceID, "human_rating")
	return rating, nil
}

func (s *MemoryStore) addImprovementNoteLocked(traceID string, note review.ImprovementNote, createdAt time.Time) (review.ImprovementNote, error) {
	if _, ok := s.traces[traceID]; !ok {
		return review.ImprovementNote{}, errors.New("trace not found")
	}
	note.TraceID = traceID
	note.CreatedAt = createdAt
	s.notes[traceID] = append(s.notes[traceID], note)
	trace := s.traces[traceID]
	s.feedbackRecords[traceID] = append(s.feedbackRecords[traceID], review.FeedbackRecord{
		ID:             nextID("feedback", len(s.feedbackRecords[traceID])+1),
		ConversationID: trace.Summary.ConversationID,
		CaseID:         trace.Summary.CaseID,
		TraceID:        traceID,
		TargetType:     review.FeedbackTargetTrace,
		TargetID:       traceID,
		Verdict:        note.Category,
		Labels:         []string{"improvement_note"},
		Notes:          note.Note,
		ReviewerID:     note.CreatedBy,
		CreatedAt:      note.CreatedAt,
	})
	return note, nil
}

func (s *MemoryStore) addFeedbackLocked(record review.FeedbackRecord, createdAt time.Time) (review.FeedbackRecord, error) {
	if strings.TrimSpace(record.TargetID) == "" {
		return review.FeedbackRecord{}, errors.New("target_id is required")
	}
	trace, err := s.resolveFeedbackTraceLocked(record)
	if err != nil {
		return review.FeedbackRecord{}, err
	}
	record.TraceID = trace.Summary.TraceID
	record.CaseID = trace.Summary.CaseID
	record.ConversationID = trace.Summary.ConversationID
	if record.ID == "" {
		record.ID = nextID("feedback", len(s.feedbackRecords[record.TraceID])+1)
	}
	record.CreatedAt = createdAt
	s.feedbackRecords[record.TraceID] = append(s.feedbackRecords[record.TraceID], record)
	return record, nil
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

func (s *MemoryStore) GetSettings() improvement.Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return normalizedSettings(s.settings)
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
		s.workflows[i].Version++
		if status == "completed" || status == "failed" || status == "needs_human" || status == "superseded" {
			s.workflows[i].CompletedAt = &now
		}
		return s.workflows[i], nil
	}
	return Workflow{}, errors.New("workflow not found")
}

func (s *MemoryStore) applyTraceUpdateLocked(traceID string, update TraceUpdate) (events.Trace, error) {
	trace, ok := s.traces[traceID]
	if !ok {
		return events.Trace{}, errors.New("trace not found")
	}
	now := time.Now().UTC()
	if update.Status != nil {
		if !isTerminalTraceStatusValue(trace.Summary.Status) || isTerminalTraceStatusValue(*update.Status) {
			trace.Summary.Status = *update.Status
		}
	}
	if update.LastVerdict != nil {
		trace.Summary.LastVerdict = *update.LastVerdict
	}
	trace.Events = append(trace.Events, update.Events...)
	trace.Artifacts = appendTraceArtifacts(trace.Artifacts, update.Artifacts)
	trace.Reasoning = append(trace.Reasoning, update.Reasoning...)
	trace.ToolCalls = append(trace.ToolCalls, update.ToolCalls...)
	trace.SlackActions = append(trace.SlackActions, update.SlackActions...)
	for _, action := range update.SlackActions {
		if trace.Summary.ConversationID == "" {
			continue
		}
		s.upsertConversationEntryLocked(slackActionConversationEntry(trace.Summary.ConversationID, trace.Summary.TriggerEventID, trace.Summary.TraceID, action))
	}
	recomputeTraceSummary(&trace)
	if trace.Summary.EndedAt.Before(now) && (trace.Summary.Status == events.StatusCompleted || trace.Summary.Status == events.StatusFailed || trace.Summary.Status == events.StatusNeedsHuman) {
		trace.Summary.EndedAt = now
	}
	s.traces[traceID] = trace
	if caseRecord, ok := s.cases[trace.Summary.CaseID]; ok {
		caseRecord.LatestTraceID = trace.Summary.TraceID
		caseRecord.UpdatedAt = now
		if trace.Summary.Status == events.StatusCompleted || trace.Summary.Status == events.StatusNeedsHuman {
			caseRecord.Status = conversation.CaseActive
		}
		s.cases[caseRecord.ID] = caseRecord
	}
	if update.WorkflowStatus != "" {
		if _, err := s.updateWorkflowStatusLocked(trace.Summary.WorkflowID, update.WorkflowStatus, update.WorkflowError); err != nil {
			return events.Trace{}, err
		}
	}
	return trace, nil
}

func appendTraceArtifacts(existing []events.Artifact, incoming ...[]events.Artifact) []events.Artifact {
	out := append([]events.Artifact{}, existing...)
	seen := map[string]struct{}{}
	for _, item := range out {
		key := TraceArtifactDedupKey(item)
		if key == "" {
			continue
		}
		seen[key] = struct{}{}
	}
	for _, group := range incoming {
		for _, item := range group {
			key := TraceArtifactDedupKey(item)
			if key == "" {
				out = append(out, item)
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, item)
		}
	}
	return out
}

func TraceArtifactDedupKey(item events.Artifact) string {
	if strings.TrimSpace(item.ID) != "" {
		return strings.TrimSpace(item.ID)
	}
	if strings.TrimSpace(item.URL) != "" {
		return strings.TrimSpace(item.Kind) + "|" + strings.TrimSpace(item.URL)
	}
	return ""
}

func isTerminalTraceStatusValue(status events.Status) bool {
	switch status {
	case events.StatusCompleted, events.StatusFailed, events.StatusNeedsHuman:
		return true
	default:
		return false
	}
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
	judgments := s.buildJudgments(trace, event, s.ratings[traceID], s.proposalMemory)
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
	for idx := range judgments {
		judgments[idx].EvalRunID = runID
	}
	s.evalRuns[runID] = run
	s.evalJudgments[runID] = judgments
	s.updateCandidateLocked(trace, event, run, judgments)
	return run, append([]evals.Judgment(nil), judgments...), nil
}

func (s *MemoryStore) buildJudgments(trace events.Trace, event *ingestion.EventEnvelope, ratings []review.HumanRating, memories []review.ProposalMemory) []evals.Judgment {
	now := time.Now().UTC()
	judgments := []evals.Judgment{}
	runtimeFailure, hasRuntimeFailure := s.runtimeFailureForTrace(trace)
	deterministicScore := 0.85
	deterministicReason := "Trace completed within expected policy, cost, and validation budget."
	if hasRuntimeFailure {
		deterministicScore = 0.18
		deterministicReason = fmt.Sprintf("Trace failed because RSI runtime execution broke in %s (%s).", runtimeFailure.Subsystem, runtimeFailure.FailureMode)
	} else if trace.Summary.Status == events.StatusFailed || trace.Summary.Status == events.StatusNeedsHuman {
		deterministicScore = 0.35
		deterministicReason = "Trace indicates an operational failure or unresolved handoff."
	}
	judgments = append(judgments, evals.Judgment{
		ID:        nextID("judge", 1+len(judgments)),
		Layer:     evals.LayerDeterministic,
		Category:  "state_accounting",
		Score:     deterministicScore,
		Passed:    deterministicScore >= 0.7,
		Rationale: deterministicReason,
		CreatedAt: now,
	})

	taskScore := 0.82
	taskReason := "Workflow output appears actionable."
	if hasRuntimeFailure {
		taskScore = 0.22
		taskReason = "RSI runtime failure prevented the workflow from completing the user-facing action."
	} else if len(ratings) > 0 {
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
		Category:  "reasoning_quality",
		Score:     taskScore,
		Passed:    taskScore >= 0.7,
		Rationale: taskReason,
		CreatedAt: now,
	})

	artifactScore := 0.82
	artifactReason := "No required artifact deliverable was missing."
	if traceHasUserFacingArtifact(trace) {
		artifactReason = "User-facing artifact deliverables were recorded on the trace."
	}
	judgments = append(judgments, evals.Judgment{
		ID:        nextID("judge", 1+len(judgments)),
		Layer:     evals.LayerTaskQuality,
		Category:  "artifact_quality",
		Score:     artifactScore,
		Passed:    artifactScore >= 0.7,
		Rationale: artifactReason,
		CreatedAt: now,
	})

	deliveryScore := 0.84
	deliveryReason := "No failed user-visible delivery was recorded."
	if traceHasFailedSlackDelivery(trace) {
		deliveryScore = 0.25
		deliveryReason = "The workflow produced a reply but Slack delivery failed or was not durably accounted for."
	} else if traceHasSuccessfulSlackDelivery(trace) {
		deliveryReason = "Slack delivery completed and was durably recorded."
	} else if trace.Summary.Status == events.StatusNeedsHuman || trace.Summary.Status == events.StatusFailed {
		deliveryScore = 0.5
		deliveryReason = "Workflow did not reach a user-visible delivery terminal state."
	}
	judgments = append(judgments, evals.Judgment{
		ID:        nextID("judge", 1+len(judgments)),
		Layer:     evals.LayerDeterministic,
		Category:  "delivery_reliability",
		Score:     deliveryScore,
		Passed:    deliveryScore >= 0.7,
		Rationale: deliveryReason,
		CreatedAt: now,
	})

	architectureScore := 0.8
	architectureReason := "Architecture boundary and recursive improvement controls look healthy."
	if hasRuntimeFailure && (runtimeFailure.Subsystem == "delivery" || runtimeFailure.FailureMode == "state_accounting_after_delivery") {
		architectureReason = "Architecture score held neutral because the failure belongs to delivery/state accounting, not reasoning architecture."
	} else if hasRuntimeFailure {
		architectureScore = 0.24
		architectureReason = fmt.Sprintf("Trace evidence points to an RSI %s runtime failure: %s.", runtimeFailure.Subsystem, runtimeFailure.FailureMode)
	} else if traceHasFailedSlackDelivery(trace) {
		architectureReason = "Architecture score held neutral because the failure belongs to delivery reliability."
	} else if event != nil && event.OwnershipHint == "platform" && (containsAny(event.NormalizedProblemStatement, []string{"proposal", "eval", "closed-loop", "architecture"}) || event.Source == ingestion.SourceSentry) {
		architectureScore = 0.54
		architectureReason = "Platform-level event suggests architectural debt or missing self-improvement guardrails."
	}
	if hasRecentRejectedMemory(memories, s.candidateKeyForTrace(trace, event)) {
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
	if trace.Summary.TriggerEventID != "" {
		for i := range s.events {
			if s.events[i].ID == trace.Summary.TriggerEventID {
				return &s.events[i]
			}
		}
	}
	for i := range s.ingestions {
		if s.ingestions[i].ID != trace.Summary.IngestionID {
			continue
		}
		if s.ingestions[i].EventID != "" {
			for j := range s.events {
				if s.events[j].ID == s.ingestions[i].EventID {
					return &s.events[j]
				}
			}
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
	key := s.candidateKeyForTrace(trace, event)
	failureMode := s.failureModeForTrace(trace, judgments)
	subsystem := s.subsystemForTrace(trace, event)
	interventionType := s.interventionTypeForTrace(trace, judgments)
	now := time.Now().UTC()
	candidate, ok := s.candidates[key]
	if !ok {
		candidate = improvement.Candidate{
			ID:               nextID("cand", len(s.candidates)+1),
			CandidateKey:     key,
			ConversationID:   trace.Summary.ConversationID,
			CaseID:           trace.Summary.CaseID,
			OriginTraceID:    trace.Summary.TraceID,
			EvidenceTraceIDs: []string{trace.Summary.TraceID},
			Subsystem:        subsystem,
			FailureMode:      failureMode,
			InterventionType: interventionType,
			TargetLayer:      targetLayerForCandidate(trace, subsystem, failureMode, interventionType),
			TargetKind:       targetKindForCandidate(trace, subsystem, failureMode, interventionType),
			TargetRef:        targetRefForCandidate(trace, subsystem, failureMode, interventionType),
			Status:           improvement.CandidateNeedsEvidence,
			Severity:         s.severityForTrace(trace, event),
			RiskTier:         improvement.RiskMedium,
			Hypothesis:       s.hypothesisForTrace(trace, event, judgments),
			ProposedScope:    s.proposedScopeForTrace(trace, event),
			CreatedAt:        now,
		}
	}
	candidate.RecurrenceCount++
	candidate.ExpectedImpact = maxFloat(candidate.ExpectedImpact, 1-run.OverallScore)
	candidate.NoveltyScore = noveltyScoreForCandidate(s.proposalMemory, candidate.CandidateKey)
	candidate.ConfidenceScore = confidenceScoreForCandidate(candidate.RecurrenceCount, judgments)
	candidate.FreshnessScore = 1.0
	candidate.PriorityScore = priorityScore(candidate.ExpectedImpact, candidate.NoveltyScore, candidate.ConfidenceScore, candidate.FreshnessScore, candidate.RecurrenceCount)
	candidate.ConversationID = firstNonEmpty(candidate.ConversationID, trace.Summary.ConversationID)
	candidate.CaseID = firstNonEmpty(candidate.CaseID, trace.Summary.CaseID)
	candidate.OriginTraceID = firstNonEmpty(candidate.OriginTraceID, trace.Summary.TraceID)
	candidate.LatestTraceID = trace.Summary.TraceID
	candidate.EvidenceTraceIDs = appendUnique(candidate.EvidenceTraceIDs, trace.Summary.TraceID)
	candidate.SourceEvalIDs = appendUnique(candidate.SourceEvalIDs, run.ID)
	candidate.EvidenceArtifactIDs = collectArtifactIDs(trace)
	candidate.PriorSimilarProposalIDs = similarProposalIDs(s.proposalMemory, candidate.CandidateKey)
	candidate.NewEvidenceSinceLastRejection = hasNewEvidenceSinceLastRejection(s.proposalMemory, candidate.CandidateKey, run.CreatedAt, candidate.RecurrenceCount)
	candidate.LastEvaluatedAt = now
	candidate.UpdatedAt = now
	if candidate.RecurrenceCount >= 2 || run.OverallScore < 0.65 || hasSevereFailedJudgment(judgments) {
		candidate.Status = improvement.CandidateQueued
	} else {
		candidate.Status = improvement.CandidateNeedsEvidence
	}
	if isUngroundedFailedWorkflowCandidate(candidate) {
		candidate.Status = improvement.CandidateNeedsEvidence
	}
	if len(candidate.PriorSimilarProposalIDs) > 0 && !candidate.NewEvidenceSinceLastRejection && rejectedRecently(s.proposalMemory, candidate.CandidateKey) {
		candidate.Status = improvement.CandidateNeedsEvidence
		candidate.PriorityScore *= 0.4
	}
	s.candidates[key] = candidate
}

func hasSevereFailedJudgment(judgments []evals.Judgment) bool {
	for _, judgment := range judgments {
		if !judgment.Passed && judgment.Score <= 0.5 {
			return true
		}
	}
	return false
}

func (s *MemoryStore) candidateKeyForTrace(trace events.Trace, event *ingestion.EventEnvelope) string {
	subsystem := s.subsystemForTrace(trace, event)
	failureMode := s.failureModeForTrace(trace, nil)
	if runtimeFailure, ok := s.runtimeFailureForTrace(trace); ok {
		return fmt.Sprintf("%s:%s:%s", firstNonEmpty(subsystem, runtimeFailure.Subsystem), s.interventionTypeForTrace(trace, nil), failureMode)
	}
	return fmt.Sprintf("%s:%s:%s", subsystem, trace.Summary.WorkflowKind, failureMode)
}

func (s *MemoryStore) subsystemForTrace(trace events.Trace, event *ingestion.EventEnvelope) string {
	if runtimeFailure, ok := s.runtimeFailureForTrace(trace); ok {
		return runtimeFailure.Subsystem
	}
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

func (s *MemoryStore) severityForTrace(trace events.Trace, event *ingestion.EventEnvelope) string {
	if _, ok := s.runtimeFailureForTrace(trace); ok {
		return string(ingestion.SeverityError)
	}
	if event != nil {
		return string(event.Severity)
	}
	if trace.Summary.Status == events.StatusFailed {
		return string(ingestion.SeverityError)
	}
	return string(ingestion.SeverityWarning)
}

func (s *MemoryStore) failureModeForTrace(trace events.Trace, judgments []evals.Judgment) string {
	if runtimeFailure, ok := s.runtimeFailureForTrace(trace); ok {
		return runtimeFailure.FailureMode
	}
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

func (s *MemoryStore) interventionTypeForTrace(trace events.Trace, judgments []evals.Judgment) string {
	if _, ok := s.runtimeFailureForTrace(trace); ok {
		return "policy_or_runtime_fix"
	}
	return interventionTypeForJudgments(judgments)
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

func (s *MemoryStore) hypothesisForTrace(trace events.Trace, event *ingestion.EventEnvelope, judgments []evals.Judgment) string {
	if runtimeFailure, ok := s.runtimeFailureForTrace(trace); ok {
		return runtimeFailure.Hypothesis
	}
	if event != nil && event.Source == ingestion.SourceSentry {
		return "Introduce stronger proposal slot gating and shared-state evaluation to avoid repeated recursive failures."
	}
	if containsAny(trace.Summary.WorkflowKind, []string{"architecture"}) {
		return "Improve architecture routing and add stronger closed-loop evaluation signals."
	}
	return "Tighten workflow policy and evaluation criteria to improve remediation quality."
}

func (s *MemoryStore) proposedScopeForTrace(trace events.Trace, event *ingestion.EventEnvelope) string {
	if runtimeFailure, ok := s.runtimeFailureForTrace(trace); ok {
		return runtimeFailure.ProposedScope
	}
	if event != nil && event.Source == ingestion.SourceSentry {
		return "platform + adapters"
	}
	return "whole_repo"
}

type runtimeFailureEvidence struct {
	Subsystem     string
	FailureMode   string
	Hypothesis    string
	ProposedScope string
}

func (s *MemoryStore) runtimeFailureForTrace(trace events.Trace) (runtimeFailureEvidence, bool) {
	if failure, ok := s.runtimeFailureFromWorkflowAttempts(trace); ok {
		return failure, true
	}
	if failure, ok := runtimeFailureFromActionPersistence(trace); ok {
		return failure, true
	}
	if failure, ok := runtimeFailureFromRSIToolCall(trace); ok {
		return failure, true
	}
	if failure, ok := runtimeFailureFromWorkflowFailure(trace); ok {
		return failure, true
	}
	return runtimeFailureEvidence{}, false
}

func (s *MemoryStore) runtimeFailureFromWorkflowAttempts(trace events.Trace) (runtimeFailureEvidence, bool) {
	for _, workflow := range s.workflowAttemptsForTraceLocked(trace) {
		failureClass := strings.TrimSpace(workflow.FailureClass)
		failureKind := strings.TrimSpace(stringFromDiagnostic(workflow.RunnerDiagnostics, "failure_kind"))
		providerParam := strings.TrimSpace(stringFromDiagnostic(workflow.RunnerDiagnostics, "provider_error_param"))
		switch {
		case (failureClass == "runner_invalid_request" || failureKind == "invalid_request") && providerParam == "tools[0].name":
			return runtimeFailureEvidence{
				Subsystem:     "runner",
				FailureMode:   "runner_invalid_tool_name_contract",
				Hypothesis:    "The model provider rejected tools[0].name because a non-native or malformed tool name crossed the runner boundary. Keep provider-facing tool names native and fail request admission when tool registration is invalid.",
				ProposedScope: "runner",
			}, true
		case failureClass == "runner_missing_structured_output":
			if traceHasUserFacingArtifact(trace) || traceHasSuccessfulSlackDelivery(trace) {
				return runtimeFailureEvidence{
					Subsystem:     "control-plane",
					FailureMode:   "state_accounting_after_delivery",
					Hypothesis:    "Reconcile terminal workflow state from durable artifact manifests and delivery receipts before scoring the parent runner response as failed.",
					ProposedScope: "control-plane + shared-store",
				}, true
			}
			return runtimeFailureEvidence{
				Subsystem:     "runner",
				FailureMode:   "runner_missing_structured_output",
				Hypothesis:    "Enforce the runner structured-output contract so control-plane failures surface as grounded evaluations instead of opaque workflow errors.",
				ProposedScope: "control-plane + runner",
			}, true
		case failureClass == "runner_structured_output_parse_failure":
			return runtimeFailureEvidence{
				Subsystem:     "runner",
				FailureMode:   "runner_structured_output_parse_failure",
				Hypothesis:    "Tighten the runner structured-output schema so control-plane can parse responses deterministically and recover from malformed outputs.",
				ProposedScope: "control-plane + runner",
			}, true
		case failureClass == "runner_transport_timeout":
			return runtimeFailureEvidence{
				Subsystem:     "runner",
				FailureMode:   "runner_transport_timeout",
				Hypothesis:    "Treat runner transport timeouts as terminal runtime evidence with explicit diagnostics so the platform can distinguish network failures from model-output problems.",
				ProposedScope: "runner + control-plane",
			}, true
		case failureClass == "reply_delivery_failed" || failureClass == "missing_reply_delivery" || failureClass == "runner_reply_delivery_uncertain":
			return runtimeFailureEvidence{
				Subsystem:     "delivery",
				FailureMode:   failureClass,
				Hypothesis:    "Treat Slack delivery attempts and provider errors as first-class receipts so failed DM/channel delivery routes create targeted delivery reliability pressure.",
				ProposedScope: "control-plane + native Hermes tools",
			}, true
		case failureClass == "runner_partial_completion_unrecoverable":
			return runtimeFailureEvidence{
				Subsystem:     "runner",
				FailureMode:   "runner_partial_completion_unrecoverable",
				Hypothesis:    "Bounded-stop workflow runs must finalize into a grounded partial completion before the hard deadline; reserve time for no-tools finalization and surface unrecoverable partial-finalization failures explicitly instead of collapsing them into generic timeout handling.",
				ProposedScope: "runner + control-plane",
			}, true
		}
	}
	return runtimeFailureEvidence{}, false
}

func (s *MemoryStore) workflowAttemptsForTraceLocked(trace events.Trace) []Workflow {
	attempts := make([]Workflow, 0)
	caseID := strings.TrimSpace(trace.Summary.CaseID)
	traceID := strings.TrimSpace(trace.Summary.TraceID)
	workflowID := strings.TrimSpace(trace.Summary.WorkflowID)
	for _, workflow := range s.workflows {
		switch {
		case caseID != "" && strings.TrimSpace(workflow.CaseID) == caseID:
			attempts = append(attempts, workflow)
		case traceID != "" && strings.TrimSpace(workflow.TraceID) == traceID:
			attempts = append(attempts, workflow)
		case workflowID != "" && strings.TrimSpace(workflow.ID) == workflowID:
			attempts = append(attempts, workflow)
		}
	}
	sort.Slice(attempts, func(i, j int) bool {
		if attempts[i].AttemptNumber == attempts[j].AttemptNumber {
			return attempts[i].CreatedAt.After(attempts[j].CreatedAt)
		}
		return attempts[i].AttemptNumber > attempts[j].AttemptNumber
	})
	return attempts
}

func runtimeFailureFromActionPersistence(trace events.Trace) (runtimeFailureEvidence, bool) {
	for _, event := range trace.Events {
		if event.EventType != "action.persistence_failed" {
			continue
		}
		subsystem := firstNonEmpty(traceEventTagValue(event.Description, "subsystem"), "control-plane")
		failureMode := firstNonEmpty(traceEventTagValue(event.Description, "failure_mode"), "action_result_persistence_failure")
		if failureMode == "action_result_primary_key_collision" {
			subsystem = "shared-store"
		}
		return runtimeFailureEvidence{
			Subsystem:     subsystem,
			FailureMode:   failureMode,
			Hypothesis:    runtimeFailureHypothesis(subsystem, failureMode),
			ProposedScope: runtimeFailureScope(subsystem, failureMode),
		}, true
	}
	return runtimeFailureEvidence{}, false
}

func runtimeFailureFromRSIToolCall(trace events.Trace) (runtimeFailureEvidence, bool) {
	for _, call := range trace.ToolCalls {
		toolName := strings.TrimSpace(call.ToolName)
		if !strings.HasPrefix(toolName, "rsi.") || !failedToolCallStatus(call.Status) {
			continue
		}
		if toolName == "rsi.workflow_context" && workflowContextBindingFailure(call) {
			return runtimeFailureEvidence{
				Subsystem:     "control-plane",
				FailureMode:   "workflow_context_binding_failure",
				Hypothesis:    "Bind RSI workflow-context reads to the active workflow and trace so the control loop always gathers scoped evidence before calling the runner.",
				ProposedScope: "control-plane + native Hermes tools",
			}, true
		}
		return runtimeFailureEvidence{
			Subsystem:     "control-plane",
			FailureMode:   runtimeFailureModeForToolName(toolName),
			Hypothesis:    fmt.Sprintf("Stabilize internal RSI tool reads so %s cannot fail mid-workflow and strand the control loop before evaluation.", toolName),
			ProposedScope: "control-plane + native Hermes tools",
		}, true
	}
	return runtimeFailureEvidence{}, false
}

func runtimeFailureFromWorkflowFailure(trace events.Trace) (runtimeFailureEvidence, bool) {
	for _, event := range trace.Events {
		if event.EventType != "workflow.failed" {
			continue
		}
		description := strings.ToLower(strings.TrimSpace(event.Description))
		switch {
		case strings.Contains(description, "runner response missing structured_output"):
			return runtimeFailureEvidence{
				Subsystem:     "runner",
				FailureMode:   "runner_missing_structured_output",
				Hypothesis:    "Enforce the runner structured-output contract so control-plane failures surface as grounded evaluations instead of opaque workflow errors.",
				ProposedScope: "control-plane + runner",
			}, true
		case strings.Contains(description, "parse runner structured_output"):
			return runtimeFailureEvidence{
				Subsystem:     "runner",
				FailureMode:   "runner_structured_output_parse_failure",
				Hypothesis:    "Tighten the runner structured-output schema so control-plane can parse responses deterministically and recover from malformed outputs.",
				ProposedScope: "control-plane + runner",
			}, true
		}
	}
	return runtimeFailureEvidence{}, false
}

func stringFromDiagnostic(values map[string]any, key string) string {
	if len(values) == 0 {
		return ""
	}
	raw, ok := values[key]
	if !ok || raw == nil {
		return ""
	}
	switch typed := raw.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func failedToolCallStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "failed", "error", "blocked":
		return true
	default:
		return false
	}
}

func workflowContextBindingFailure(call events.ToolCallRecord) bool {
	summary := strings.ToLower(strings.TrimSpace(call.Summary))
	if strings.Contains(summary, "requires workflow_id or trace_id") {
		return true
	}
	return traceToolRequestValue(call.Request, "workflow_id") == "" && traceToolRequestValue(call.Request, "trace_id") == ""
}

func traceToolRequestValue(input map[string]interface{}, key string) string {
	if len(input) == 0 {
		return ""
	}
	value, ok := input[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func runtimeFailureModeForToolName(toolName string) string {
	replacer := strings.NewReplacer(".", "_", "-", "_")
	return replacer.Replace(strings.ToLower(strings.TrimSpace(toolName))) + "_failure"
}

func traceEventTagValue(description string, key string) string {
	description = strings.TrimSpace(description)
	if description == "" || key == "" {
		return ""
	}
	marker := key + "="
	idx := strings.Index(strings.ToLower(description), strings.ToLower(marker))
	if idx < 0 {
		return ""
	}
	value := description[idx+len(marker):]
	if end := strings.IndexAny(value, " \n\t"); end >= 0 {
		value = value[:end]
	}
	return strings.TrimSpace(value)
}

func runtimeFailureHypothesis(subsystem string, failureMode string) string {
	switch failureMode {
	case "runner_invalid_tool_name_contract":
		return "Fail native Hermes request admission when a provider-facing tool name is invalid, and persist invalid-request diagnostics instead of collapsing them into structured-output parse failures."
	case "action_result_primary_key_collision":
		return "Fix shared-store action result keying and control-plane terminalization so action result collisions cannot wedge traces before eval."
	case "postgres_unique_constraint_violation":
		return "Harden RSI runtime persistence so unique-constraint failures surface as terminal evidence and cannot strand traces mid-flight."
	default:
		return fmt.Sprintf("Stabilize %s runtime persistence so platform failures become terminal evidence instead of wedging the recursive loop.", firstNonEmpty(subsystem, "control-plane"))
	}
}

func runtimeFailureScope(subsystem string, failureMode string) string {
	if failureMode == "runner_invalid_tool_name_contract" {
		return "runner"
	}
	if failureMode == "action_result_primary_key_collision" || subsystem == "shared-store" {
		return "control-plane + shared-store"
	}
	return firstNonEmpty(subsystem, "control-plane")
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

func traceHasUserFacingArtifact(trace events.Trace) bool {
	for _, artifact := range trace.Artifacts {
		kind := strings.TrimSpace(strings.ToLower(artifact.Kind))
		if kind == "" || kind == "event_payload" {
			continue
		}
		return true
	}
	for _, action := range trace.SlackActions {
		if len(action.ArtifactRefs) > 0 {
			return true
		}
	}
	return false
}

func traceHasSuccessfulSlackDelivery(trace events.Trace) bool {
	for _, action := range trace.SlackActions {
		if events.SlackDeliveryStatusSucceeded(action.SendStatus) {
			return true
		}
	}
	return false
}

func traceHasFailedSlackDelivery(trace events.Trace) bool {
	for _, action := range trace.SlackActions {
		status := strings.TrimSpace(action.SendStatus)
		if status == "" {
			continue
		}
		if !events.SlackDeliveryStatusSucceeded(status) {
			return true
		}
	}
	return false
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

func (s *MemoryStore) ListRuntimeDiagnoses() []improvement.RuntimeDiagnosis {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]improvement.RuntimeDiagnosis, 0, len(s.runtimeDiagnoses))
	for _, diagnosis := range s.runtimeDiagnoses {
		out = append(out, diagnosis)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}

func (s *MemoryStore) GetRuntimeDiagnosis(diagnosisID string) (improvement.RuntimeDiagnosis, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.runtimeDiagnoses[strings.TrimSpace(diagnosisID)]
	return item, ok
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
		if isUngroundedFailedWorkflowCandidate(candidate) {
			candidate.Status = improvement.CandidateNeedsEvidence
			candidate.UpdatedAt = now
			s.candidates[candidate.CandidateKey] = candidate
			continue
		}
		diagnosis, hasDiagnosis := s.runtimeDiagnosisForCandidateLocked(candidate)
		if hasDiagnosis {
			switch diagnosis.Status {
			case improvement.RuntimeDiagnosisQueued, improvement.RuntimeDiagnosisInvestigating:
				continue
			case improvement.RuntimeDiagnosisNeedsEvidence, improvement.RuntimeDiagnosisClosed:
				candidate.Status = improvement.CandidateNeedsEvidence
				candidate.LineStatus = improvement.LineNeedsEvidence
				candidate.UpdatedAt = now
				s.candidates[candidate.CandidateKey] = candidate
				continue
			case improvement.RuntimeDiagnosisPromoted:
				continue
			}
		}
		if s.hasActiveProposalForCandidateLocked(candidate.CandidateKey) {
			continue
		}
		if rejectedRecently(s.proposalMemory, candidate.CandidateKey) && !candidate.NewEvidenceSinceLastRejection {
			continue
		}
		kind := review.RecommendProposalIntervention(candidate)
		targetSurface := review.ProposalTargetSurfaceFromCandidate(candidate)
		summary := candidate.Hypothesis
		recommendedRationale := review.ProposalInterventionRationale(candidate, kind, targetSurface)
		validationPlan := review.ProposalValidationPlan(kind, targetSurface)
		if hasDiagnosis && diagnosis.Status == improvement.RuntimeDiagnosisGrounded {
			if strings.TrimSpace(diagnosis.Summary) != "" {
				summary = strings.TrimSpace(diagnosis.Summary)
			}
			if strings.TrimSpace(diagnosis.TargetSurface) != "" {
				targetSurface = strings.TrimSpace(diagnosis.TargetSurface)
			}
			if strings.TrimSpace(diagnosis.RecommendedFix) != "" {
				recommendedRationale = strings.TrimSpace(diagnosis.RecommendedFix)
			}
			if strings.TrimSpace(diagnosis.ValidationPlan) != "" {
				validationPlan = strings.TrimSpace(diagnosis.ValidationPlan)
			}
		}
		proposal := review.Proposal{
			ID:                               nextID("proposal", len(s.proposals)+1),
			TraceID:                          candidate.LatestTraceID,
			ConversationID:                   candidate.ConversationID,
			CaseID:                           candidate.CaseID,
			OriginTraceID:                    firstNonEmpty(candidate.OriginTraceID, candidate.LatestTraceID),
			EvidenceTraceIDs:                 append([]string(nil), candidate.EvidenceTraceIDs...),
			Title:                            proposalTitle(candidate),
			Category:                         candidate.InterventionType,
			Summary:                          summary,
			Status:                           review.ProposalPendingReview,
			CandidateKey:                     candidate.CandidateKey,
			TargetLayer:                      candidate.TargetLayer,
			TargetKind:                       candidate.TargetKind,
			TargetRef:                        candidate.TargetRef,
			SourceEvalIDs:                    append([]string(nil), candidate.SourceEvalIDs...),
			RiskTier:                         string(candidate.RiskTier),
			ProposedScope:                    candidate.ProposedScope,
			EvidenceArtifactIDs:              append([]string(nil), candidate.EvidenceArtifactIDs...),
			ActiveSlotConsuming:              true,
			ReviewDeadline:                   now.Add(proposalReviewSLA),
			PriorSimilarProposalIDs:          append([]string(nil), candidate.PriorSimilarProposalIDs...),
			NewEvidenceSinceLastRejection:    candidate.NewEvidenceSinceLastRejection,
			RecommendedInterventionKind:      kind,
			RecommendedInterventionRationale: recommendedRationale,
			TargetSurface:                    targetSurface,
			TouchedFiles:                     []string{},
			ValidationPlan:                   validationPlan,
			MaterialRiskSummary:              review.ProposalRiskSummary(string(candidate.RiskTier), targetSurface, kind),
			RecommendedDisposition:           review.ProposalDispositionForIntervention(kind),
			CreatedAt:                        now,
		}
		s.proposals[proposal.ID] = proposal
		candidate.Status = improvement.CandidatePromoted
		candidate.UpdatedAt = now
		s.candidates[candidate.CandidateKey] = candidate
		if hasDiagnosis && diagnosis.Status == improvement.RuntimeDiagnosisGrounded {
			diagnosis.Status = improvement.RuntimeDiagnosisPromoted
			diagnosis.PromotedAt = &now
			diagnosis.UpdatedAt = now
			s.runtimeDiagnoses[diagnosis.ID] = diagnosis
		}
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

func (s *MemoryStore) runtimeDiagnosisForCandidateLocked(candidate improvement.Candidate) (improvement.RuntimeDiagnosis, bool) {
	targetRepo := strings.TrimSpace(candidate.TargetRef)
	var best improvement.RuntimeDiagnosis
	found := false
	exactRepo := false
	for _, item := range s.runtimeDiagnoses {
		if item.CandidateKey != candidate.CandidateKey {
			continue
		}
		itemRepo := strings.TrimSpace(item.Repo)
		itemExactRepo := targetRepo != "" && itemRepo == targetRepo
		switch {
		case !found:
		case itemExactRepo && !exactRepo:
		case itemExactRepo == exactRepo && item.UpdatedAt.After(best.UpdatedAt):
		default:
			continue
		}
		best = item
		found = true
		exactRepo = itemExactRepo
	}
	return best, found
}

func isUngroundedFailedWorkflowCandidate(candidate improvement.Candidate) bool {
	if candidate.FailureMode != "failed_workflow" {
		return false
	}
	scope := strings.TrimSpace(candidate.ProposedScope)
	hypothesis := strings.TrimSpace(candidate.Hypothesis)
	if scope == "whole_repo" || scope == "platform + adapters" {
		return true
	}
	switch hypothesis {
	case "Introduce stronger proposal slot gating and shared-state evaluation to avoid repeated recursive failures.",
		"Improve architecture routing and add stronger closed-loop evaluation signals.",
		"Tighten workflow policy and evaluation criteria to improve remediation quality.":
		return true
	default:
		return false
	}
}

func proposalTitle(candidate improvement.Candidate) string {
	return fmt.Sprintf("Improve %s: %s", candidate.Subsystem, candidate.FailureMode)
}

func targetLayerForCandidate(trace events.Trace, subsystem, failureMode, interventionType string) harness.TargetLayer {
	lowerFailure := strings.ToLower(strings.TrimSpace(failureMode))
	lowerIntervention := strings.ToLower(strings.TrimSpace(interventionType))
	switch {
	case strings.Contains(lowerFailure, "memory"), strings.Contains(lowerFailure, "prompt"), strings.Contains(lowerFailure, "tool_selection"), strings.Contains(lowerFailure, "behavioral"):
		return harness.TargetLayerHarnessOverlay
	case strings.Contains(lowerIntervention, "overlay"), strings.Contains(lowerIntervention, "prompt"), strings.Contains(lowerIntervention, "behavior"):
		return harness.TargetLayerHarnessOverlay
	case strings.TrimSpace(trace.Summary.WorkflowKind) == "proposal":
		return harness.TargetLayerPlatformRuntime
	case strings.TrimSpace(subsystem) == "control-plane", strings.TrimSpace(subsystem) == "improvement-plane", strings.TrimSpace(subsystem) == "shared-store", strings.TrimSpace(subsystem) == "delivery":
		return harness.TargetLayerRepoChange
	default:
		return harness.TargetLayerRepoChange
	}
}

func targetKindForCandidate(trace events.Trace, subsystem, failureMode, interventionType string) string {
	layer := targetLayerForCandidate(trace, subsystem, failureMode, interventionType)
	if layer == harness.TargetLayerHarnessOverlay {
		return "runner_role"
	}
	return "repo"
}

func targetRefForCandidate(trace events.Trace, subsystem, failureMode, interventionType string) string {
	layer := targetLayerForCandidate(trace, subsystem, failureMode, interventionType)
	if layer == harness.TargetLayerHarnessOverlay {
		switch strings.TrimSpace(trace.Summary.WorkflowKind) {
		case "incident":
			return "prod"
		case "proposal":
			return "proposal"
		default:
			return "prod"
		}
	}
	return "rsi-agent-platform"
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

func buildRepoChangeContext(proposal review.Proposal, memories []review.ProposalMemory) string {
	context := fmt.Sprintf("Proposal %s recommends %s on %s. Rationale: %s.", proposal.ID, firstNonEmpty(string(proposal.RecommendedInterventionKind), string(review.InterventionRepoChange)), firstNonEmpty(proposal.TargetSurface, proposal.ProposedScope), firstNonEmpty(proposal.RecommendedInterventionRationale, proposal.Summary))
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
	return s.recordPRAttemptLocked(attempt)
}

func (s *MemoryStore) recordPRAttemptLocked(attempt improvement.PRAttempt) (improvement.PRAttempt, error) {
	if attempt.ID == "" {
		attempt.ID = nextID("pr", len(s.prAttempts)+1)
	}
	if attempt.CreatedAt.IsZero() {
		attempt.CreatedAt = time.Now().UTC()
	}
	if attempt.ProposalID != "" {
		if proposal, ok := s.proposals[attempt.ProposalID]; ok {
			attempt.ConversationID = firstNonEmpty(attempt.ConversationID, proposal.ConversationID)
			attempt.CaseID = firstNonEmpty(attempt.CaseID, proposal.CaseID)
			attempt.OriginTraceID = firstNonEmpty(attempt.OriginTraceID, proposal.OriginTraceID)
		}
	}
	s.prAttempts[attempt.ID] = attempt
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

type derivedOutcome struct {
	OutcomeType outcome.Type
	Verdict     outcome.Verdict
	Score       float64
	Summary     string
	Details     string
	ProposalID  string
}

func slackOutcomeFromText(text string) (derivedOutcome, bool) {
	normalized := strings.ToLower(strings.TrimSpace(text))
	normalized = strings.TrimPrefix(normalized, "rsi outcome:")
	normalized = strings.TrimPrefix(normalized, "rsi:")
	normalized = strings.TrimSpace(normalized)
	switch {
	case strings.HasPrefix(normalized, "accepted"):
		return derivedOutcome{OutcomeType: outcome.TypeAnswerQuality, Verdict: outcome.VerdictPositive, Score: 1, Summary: "Slack participant marked the answer accepted."}, true
	case strings.HasPrefix(normalized, "corrected"):
		return derivedOutcome{OutcomeType: outcome.TypeAnswerQuality, Verdict: outcome.VerdictNegative, Score: 0.2, Summary: "Slack participant corrected the answer."}, true
	case strings.HasPrefix(normalized, "mitigated"):
		return derivedOutcome{OutcomeType: outcome.TypeIncidentMitigation, Verdict: outcome.VerdictPositive, Score: 1, Summary: "Slack participant marked the incident mitigated."}, true
	case strings.HasPrefix(normalized, "still broken"):
		return derivedOutcome{OutcomeType: outcome.TypeIncidentMitigation, Verdict: outcome.VerdictNegative, Score: 0, Summary: "Slack participant reported the incident is still broken."}, true
	case strings.HasPrefix(normalized, "feature approved"):
		return derivedOutcome{OutcomeType: outcome.TypeFeatureDelivery, Verdict: outcome.VerdictPositive, Score: 0.8, Summary: "Slack participant approved the feature request direction."}, true
	case strings.HasPrefix(normalized, "feature declined"):
		return derivedOutcome{OutcomeType: outcome.TypeFeatureDelivery, Verdict: outcome.VerdictNegative, Score: 0.1, Summary: "Slack participant declined the feature request."}, true
	default:
		return derivedOutcome{}, false
	}
}

func githubOutcomeFromMetadata(metadata map[string]interface{}) (derivedOutcome, bool) {
	if strings.TrimSpace(stringFromMetadata(metadata, "proposal_id")) == "" {
		return derivedOutcome{}, false
	}
	actionValue := strings.ToLower(strings.TrimSpace(stringFromMetadata(metadata, "action")))
	eventType := strings.ToLower(strings.TrimSpace(stringFromMetadata(metadata, "event_type")))
	state := strings.ToLower(strings.TrimSpace(stringFromMetadata(metadata, "state")))
	merged := strings.ToLower(strings.TrimSpace(stringFromMetadata(metadata, "merged")))
	switch {
	case eventType == "pull_request" && actionValue == "opened":
		return derivedOutcome{
			OutcomeType: outcome.TypeProposalEffectiveness,
			Verdict:     outcome.VerdictUnresolved,
			Score:       0.4,
			Summary:     "GitHub pull request was opened.",
			Details:     "Proposal path reached draft PR open.",
			ProposalID:  stringFromMetadata(metadata, "proposal_id"),
		}, true
	case eventType == "pull_request" && actionValue == "reopened":
		return derivedOutcome{
			OutcomeType: outcome.TypeProposalEffectiveness,
			Verdict:     outcome.VerdictUnresolved,
			Score:       0.4,
			Summary:     "GitHub pull request reopened.",
			Details:     "Proposal path re-entered the open PR state.",
			ProposalID:  stringFromMetadata(metadata, "proposal_id"),
		}, true
	case eventType == "pull_request" && (actionValue == "closed" && (merged == "true" || state == "merged")):
		return derivedOutcome{
			OutcomeType: outcome.TypeProposalEffectiveness,
			Verdict:     outcome.VerdictPositive,
			Score:       1,
			Summary:     "GitHub pull request merged.",
			Details:     "Proposal path reached merge.",
			ProposalID:  stringFromMetadata(metadata, "proposal_id"),
		}, true
	case eventType == "pull_request" && actionValue == "closed":
		return derivedOutcome{
			OutcomeType: outcome.TypeProposalEffectiveness,
			Verdict:     outcome.VerdictNegative,
			Score:       0,
			Summary:     "GitHub pull request closed without merge.",
			Details:     "Proposal path did not merge.",
			ProposalID:  stringFromMetadata(metadata, "proposal_id"),
		}, true
	case (eventType == "check_run" || eventType == "check_suite" || eventType == "workflow_run") && isGitHubFailureConclusion(stringFromMetadata(metadata, "conclusion")):
		return derivedOutcome{
			OutcomeType: outcome.TypeProposalEffectiveness,
			Verdict:     outcome.VerdictNegative,
			Score:       0.2,
			Summary:     fmt.Sprintf("GitHub %s failed.", strings.ReplaceAll(eventType, "_", " ")),
			Details:     fmt.Sprintf("Proposal attempt hit CI failure with conclusion %s.", stringFromMetadata(metadata, "conclusion")),
			ProposalID:  stringFromMetadata(metadata, "proposal_id"),
		}, true
	default:
		return derivedOutcome{}, false
	}
}

func isGitHubFailureConclusion(conclusion string) bool {
	switch strings.ToLower(strings.TrimSpace(conclusion)) {
	case "failure", "cancelled", "timed_out", "startup_failure", "action_required":
		return true
	default:
		return false
	}
}

func normalizeEvidenceRefs(items []events.EvidenceRef) []events.EvidenceRef {
	if items == nil {
		return []events.EvidenceRef{}
	}
	return items
}

func nextID(prefix string, n int) string {
	if prefix == "action-result" {
		// Keep action_result IDs sequential for collision repro in tests and stage recovery.
		return fmt.Sprintf("%s-%03d", prefix, n)
	}
	return fmt.Sprintf("%s-%s", prefix, strings.ReplaceAll(uuid.NewString(), "-", ""))
}

func int64FromPayload(payload map[string]interface{}, key string) (int64, bool) {
	if payload == nil {
		return 0, false
	}
	raw, ok := payload[key]
	if !ok {
		return 0, false
	}
	switch typed := raw.(type) {
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case float64:
		return int64(typed), true
	case string:
		value, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if err != nil {
			return 0, false
		}
		return value, true
	default:
		return 0, false
	}
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

func slackConversationKey(envelope slack.SlackEnvelope) string {
	if strings.HasPrefix(envelope.ChannelID, "D") {
		return fmt.Sprintf("slack:dm:%s", envelope.ChannelID)
	}
	root := firstNonEmpty(envelope.ThreadTS, envelope.TS)
	return fmt.Sprintf("slack:thread:%s:%s", envelope.ChannelID, root)
}

func conversationKeyForEvent(event ingestion.EventEnvelope) string {
	if key := strings.TrimSpace(event.ThreadKey); key != "" {
		return key
	}
	if key := strings.TrimSpace(event.IncidentKey); key != "" {
		return key
	}
	return fmt.Sprintf("%s:%s", event.Source, event.SourceEventID)
}

func conversationKeyForCase(caseRecord conversation.Case, conversations map[string]conversation.Conversation) string {
	if item, ok := conversations[caseRecord.ConversationID]; ok {
		return item.ExternalKey
	}
	return caseRecord.ConversationID
}

func actorTypeForEvent(event ingestion.EventEnvelope) string {
	if event.Source == ingestion.SourceSlack {
		return "user"
	}
	return "system"
}

func stringFromMetadata(metadata map[string]interface{}, key string) string {
	if metadata == nil {
		return ""
	}
	switch value := metadata[key].(type) {
	case string:
		return value
	case bool:
		if value {
			return "true"
		}
		return "false"
	case fmt.Stringer:
		return value.String()
	case int:
		return fmt.Sprintf("%d", value)
	case int64:
		return fmt.Sprintf("%d", value)
	case float64:
		return fmt.Sprintf("%v", value)
	default:
		return ""
	}
}

func boolFromMetadata(metadata map[string]interface{}, key string) bool {
	if metadata == nil {
		return false
	}
	switch value := metadata[key].(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(strings.TrimSpace(value), "true")
	default:
		return false
	}
}

func cloneMetadata(metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		out[key] = value
	}
	return out
}

func compactStrings(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func (s *MemoryStore) traceForFeedbackTargetLocked(record review.FeedbackRecord) (events.Trace, bool) {
	if record.TraceID != "" {
		trace, ok := s.traces[record.TraceID]
		return trace, ok
	}
	if record.TargetType == review.FeedbackTargetActionIntent {
		intent, ok := s.actionIntents[record.TargetID]
		if !ok {
			return events.Trace{}, false
		}
		switch {
		case strings.TrimSpace(intent.TraceID) != "":
			trace, ok := s.traces[intent.TraceID]
			return trace, ok
		case strings.TrimSpace(intent.ProposalID) != "":
			proposal, ok := s.proposals[intent.ProposalID]
			if !ok {
				return events.Trace{}, false
			}
			trace, ok := s.traces[firstNonEmpty(proposal.OriginTraceID, proposal.TraceID)]
			return trace, ok
		case strings.TrimSpace(intent.CaseID) != "":
			return s.latestTraceLocked(func(trace events.Trace) bool {
				return trace.Summary.CaseID == intent.CaseID
			})
		case strings.TrimSpace(intent.ConversationID) != "":
			return s.latestTraceLocked(func(trace events.Trace) bool {
				return trace.Summary.ConversationID == intent.ConversationID
			})
		default:
			return events.Trace{}, false
		}
	}
	for _, trace := range s.traces {
		switch record.TargetType {
		case review.FeedbackTargetTrace:
			if trace.Summary.TraceID == record.TargetID {
				return trace, true
			}
		case review.FeedbackTargetReasoning:
			for _, step := range trace.Reasoning {
				if step.ID == record.TargetID {
					return trace, true
				}
			}
		case review.FeedbackTargetToolCall:
			for _, call := range trace.ToolCalls {
				if call.ID == record.TargetID || call.ToolCallID == record.TargetID {
					return trace, true
				}
			}
		case review.FeedbackTargetSlackAction:
			for _, action := range trace.SlackActions {
				if action.ID == record.TargetID {
					return trace, true
				}
			}
		}
	}
	return events.Trace{}, false
}

func (s *MemoryStore) resolveFeedbackTraceLocked(record review.FeedbackRecord) (events.Trace, error) {
	if trace, ok := s.traceForFeedbackTargetLocked(record); ok {
		return trace, nil
	}
	switch record.TargetType {
	case review.FeedbackTargetConversation:
		if _, ok := s.conversations[record.TargetID]; !ok {
			return events.Trace{}, errors.New("conversation not found")
		}
		trace, ok := s.latestTraceLocked(func(trace events.Trace) bool {
			return trace.Summary.ConversationID == record.TargetID
		})
		if !ok {
			return events.Trace{}, errors.New("trace not found for conversation")
		}
		return trace, nil
	case review.FeedbackTargetCase:
		if _, ok := s.cases[record.TargetID]; !ok {
			return events.Trace{}, errors.New("case not found")
		}
		trace, ok := s.latestTraceLocked(func(trace events.Trace) bool {
			return trace.Summary.CaseID == record.TargetID
		})
		if !ok {
			return events.Trace{}, errors.New("trace not found for case")
		}
		return trace, nil
	case review.FeedbackTargetProposal:
		proposal, ok := s.proposals[record.TargetID]
		if !ok {
			return events.Trace{}, errors.New("proposal not found")
		}
		trace, ok := s.traces[firstNonEmpty(proposal.OriginTraceID, proposal.TraceID)]
		if !ok {
			return events.Trace{}, errors.New("trace not found for proposal")
		}
		return trace, nil
	case review.FeedbackTargetTrace, review.FeedbackTargetReasoning, review.FeedbackTargetToolCall, review.FeedbackTargetActionIntent, review.FeedbackTargetSlackAction:
		return events.Trace{}, errors.New("trace not found")
	default:
		return events.Trace{}, errors.New("unsupported feedback target")
	}
}

func (s *MemoryStore) latestTraceLocked(match func(events.Trace) bool) (events.Trace, bool) {
	var (
		current events.Trace
		found   bool
	)
	for _, trace := range s.traces {
		if !match(trace) {
			continue
		}
		if !found || trace.Summary.StartedAt.After(current.Summary.StartedAt) || (trace.Summary.StartedAt.Equal(current.Summary.StartedAt) && trace.Summary.TraceID > current.Summary.TraceID) {
			current = trace
			found = true
		}
	}
	return current, found
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
