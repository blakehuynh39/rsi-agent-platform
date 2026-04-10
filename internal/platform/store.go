package platform

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/registry"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/slack"
)

type Store interface {
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
	ListProposals() []review.Proposal
	ReviewProposal(proposalID string, decision review.ProposalReview) (review.Proposal, error)
	ExecuteTool(name string, input map[string]interface{}) ToolResult
}

type MemoryStore struct {
	mu             sync.RWMutex
	ingestions     []slack.Ingestion
	workflows      []Workflow
	assignments    []Assignment
	threadPolicies map[string]policy.ThreadPolicy
	channelPolicy  []policy.ChannelPolicy
	ownership      []registry.OwnershipRecord
	capabilities   []registry.CapabilityRecord
	templates      []registry.WorkflowTemplate
	experiments    []registry.ExperimentRecord
	traces         map[string]events.Trace
	ratings        map[string][]review.HumanRating
	notes          map[string][]review.ImprovementNote
	proposals      map[string]review.Proposal
}

func NewMemoryStore() *MemoryStore {
	now := time.Now().UTC()
	traceID := "trace-oncall-001"
	ended := now.Add(-29 * time.Minute)
	store := &MemoryStore{
		threadPolicies: map[string]policy.ThreadPolicy{
			"slack:CENG:171000001.000100": {
				ThreadKey:         "slack:CENG:171000001.000100",
				State:             policy.ThreadStateActive,
				OwnerBot:          "oncall",
				LastPolicyVersion: "v1",
				UpdatedAt:         now.Add(-30 * time.Minute),
			},
			"slack:CENG:171000002.000100": {
				ThreadKey:         "slack:CENG:171000002.000100",
				State:             policy.ThreadStateObserveOnly,
				OwnerBot:          "arch",
				LastPolicyVersion: "v1",
				UpdatedAt:         now.Add(-15 * time.Minute),
			},
		},
		channelPolicy: []policy.ChannelPolicy{
			{
				ChannelID:            "CENG",
				ProactiveEnabled:     true,
				AutoPostAllowed:      true,
				AllowedWorkflowKinds: []string{"incident", "feature-request", "architecture"},
				UpdatedAt:            now.Add(-2 * time.Hour),
			},
		},
		ownership: []registry.OwnershipRecord{
			{Domain: "depin-backend", OwnerTeam: "platform", EscalationSlack: "#platform-alerts"},
			{Domain: "story-stage", OwnerTeam: "infra", EscalationSlack: "#stage-oncall"},
		},
		capabilities: []registry.CapabilityRecord{
			{Name: "sentry.query", Kind: "tool", AllowedBots: []string{"oncall"}, ApprovalNeeded: false},
			{Name: "k8s.logs", Kind: "tool", AllowedBots: []string{"oncall"}, ApprovalNeeded: false},
			{Name: "github.create_issue", Kind: "tool", AllowedBots: []string{"fr"}, ApprovalNeeded: false},
			{Name: "github.create_pr", Kind: "tool", AllowedBots: []string{"oncall", "fr"}, ApprovalNeeded: true},
			{Name: "repo.answer_question", Kind: "skill", AllowedBots: []string{"arch"}, ApprovalNeeded: false},
		},
		templates: []registry.WorkflowTemplate{
			{Name: "incident-oncall", Kind: "incident", Description: "Investigate stage issues and propose remediation", Steps: []string{"ingest", "route", "debug", "propose"}},
			{Name: "feature-request", Kind: "feature-request", Description: "Turn asks into grounded FRs", Steps: []string{"ingest", "ground", "summarize", "issue"}},
			{Name: "architecture-question", Kind: "architecture", Description: "Answer repo/architecture questions", Steps: []string{"ingest", "ground", "answer"}},
		},
		experiments: []registry.ExperimentRecord{
			{Name: "arch-routing-threshold", Candidate: "v2", Baseline: "v1", State: "review"},
		},
		traces:   map[string]events.Trace{},
		ratings:  map[string][]review.HumanRating{},
		notes:    map[string][]review.ImprovementNote{},
		proposals: map[string]review.Proposal{},
	}

	store.ingestions = []slack.Ingestion{
		{
			ID:           "ing-001",
			ThreadKey:    "slack:CENG:171000001.000100",
			WorkflowHint: "incident",
			Source:       "slack",
			ChannelID:    "CENG",
			UserID:       "U123",
			Text:         "Investigate why staging homepage is failing and propose a fix.",
			CreatedAt:    now.Add(-35 * time.Minute),
		},
		{
			ID:           "ing-002",
			ThreadKey:    "slack:CENG:171000002.000100",
			WorkflowHint: "architecture",
			Source:       "slack",
			ChannelID:    "CENG",
			UserID:       "U456",
			Text:         "How does depin-backend issue backend JWTs after Dynamic auth?",
			CreatedAt:    now.Add(-18 * time.Minute),
		},
	}

	store.workflows = []Workflow{
		{ID: "wf-001", ThreadKey: "slack:CENG:171000001.000100", Kind: "incident", AssignedBot: "oncall", Status: "running", CreatedAt: now.Add(-34 * time.Minute)},
		{ID: "wf-002", ThreadKey: "slack:CENG:171000002.000100", Kind: "architecture", AssignedBot: "arch", Status: "completed", CreatedAt: now.Add(-17 * time.Minute)},
	}

	store.assignments = []Assignment{
		{ID: "as-001", ThreadKey: "slack:CENG:171000001.000100", AssignedBot: "oncall", Confidence: 0.97, Rationale: "Matched incident keywords and stage/debug context", CreatedAt: now.Add(-34 * time.Minute)},
		{ID: "as-002", ThreadKey: "slack:CENG:171000002.000100", AssignedBot: "arch", Confidence: 0.93, Rationale: "Question references architecture and auth internals", CreatedAt: now.Add(-17 * time.Minute)},
	}

	store.traces[traceID] = events.Trace{
		Summary: events.TraceSummary{
			TraceID:       traceID,
			IngestionID:   "ing-001",
			WorkflowID:    "wf-001",
			ThreadKey:     "slack:CENG:171000001.000100",
			WorkflowKind:  "incident",
			Status:        events.StatusRunning,
			StartedAt:     now.Add(-34 * time.Minute),
			EndedAt:       ended,
			EventCount:    4,
			ArtifactCount: 2,
		},
		Events: []events.TraceEvent{
			{TraceID: traceID, IngestionID: "ing-001", WorkflowID: "wf-001", Plane: "edge", Service: "workflow-api", Actor: "orchestrator", EventType: "slack.ingested", Status: events.StatusCompleted, StartedAt: now.Add(-34 * time.Minute), EndedAt: timePtr(now.Add(-34 * time.Minute + 300*time.Millisecond)), PayloadRef: "s3://rsi-agent-platform-stage-artifacts/payloads/ing-001.json", LatencyMs: 300, Description: "Slack thread normalized into an ingestion envelope."},
			{TraceID: traceID, IngestionID: "ing-001", WorkflowID: "wf-001", Plane: "control", Service: "control-plane", Actor: "router-policy", EventType: "workflow.routed", Status: events.StatusCompleted, StartedAt: now.Add(-33 * time.Minute), EndedAt: timePtr(now.Add(-33 * time.Minute + 42*time.Millisecond)), LatencyMs: 42, Description: "Assigned to oncall with incident template."},
			{TraceID: traceID, IngestionID: "ing-001", WorkflowID: "wf-001", Plane: "execution", Service: "runner", Actor: "oncall", EventType: "runner.started", Status: events.StatusRunning, StartedAt: now.Add(-32 * time.Minute), CostTokens: 1820, Description: "Hermes runner started incident workflow."},
			{TraceID: traceID, IngestionID: "ing-001", WorkflowID: "wf-001", Plane: "execution", Service: "tool-gateway", Actor: "oncall", EventType: "tool.executed", Status: events.StatusCompleted, StartedAt: now.Add(-31 * time.Minute), EndedAt: timePtr(now.Add(-31 * time.Minute + 180*time.Millisecond)), ArtifactRef: "artifact-log-snapshot", LatencyMs: 180, Description: "Fetched Kubernetes logs and Sentry issue details."},
		},
		Artifacts: []events.Artifact{
			{ID: "artifact-log-snapshot", TraceID: traceID, Kind: "logs", ContentType: "text/plain", URL: "s3://rsi-agent-platform-stage-artifacts/traces/trace-oncall-001/logs.txt", SizeBytes: 2048, Source: "tool-gateway"},
			{ID: "artifact-pr-diff", TraceID: traceID, Kind: "diff", ContentType: "text/x-diff", URL: "s3://rsi-agent-platform-stage-artifacts/traces/trace-oncall-001/patch.diff", SizeBytes: 1024, Source: "sandbox-runtime"},
		},
	}

	store.proposals["proposal-001"] = review.Proposal{
		ID:        "proposal-001",
		TraceID:   traceID,
		Title:     "Tighten proactive incident routing for stage-only alerts",
		Category:  "policy",
		Summary:   "Increase routing confidence threshold for auto-join outside explicitly owned incident channels.",
		Status:    "pending-review",
		CreatedAt: now.Add(-10 * time.Minute),
	}

	return store
}

func timePtr(t time.Time) *time.Time {
	return &t
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
	ingestion := slack.Ingestion{
		ID:           fmt.Sprintf("ing-%03d", len(s.ingestions)+1),
		ThreadKey:    fmt.Sprintf("slack:%s:%s", envelope.ChannelID, envelope.ThreadTS),
		WorkflowHint: deriveWorkflowHint(envelope.Text),
		Source:       "slack",
		ChannelID:    envelope.ChannelID,
		UserID:       envelope.UserID,
		Text:         envelope.Text,
		CreatedAt:    time.Now().UTC(),
	}
	s.ingestions = append(s.ingestions, ingestion)
	wf := Workflow{
		ID:          fmt.Sprintf("wf-%03d", len(s.workflows)+1),
		ThreadKey:   ingestion.ThreadKey,
		Kind:        ingestion.WorkflowHint,
		AssignedBot: assignedBotFor(ingestion.WorkflowHint),
		Status:      "queued",
		CreatedAt:   ingestion.CreatedAt,
	}
	s.workflows = append(s.workflows, wf)
	s.assignments = append(s.assignments, Assignment{
		ID:          fmt.Sprintf("as-%03d", len(s.assignments)+1),
		ThreadKey:   ingestion.ThreadKey,
		AssignedBot: wf.AssignedBot,
		Confidence:  0.82,
		Rationale:   "Created from workflow-api ingestion stub",
		CreatedAt:   ingestion.CreatedAt,
	})
	s.threadPolicies[ingestion.ThreadKey] = policy.ThreadPolicy{
		ThreadKey:         ingestion.ThreadKey,
		State:             policy.ThreadStateActive,
		OwnerBot:          wf.AssignedBot,
		LastPolicyVersion: "v1",
		UpdatedAt:         ingestion.CreatedAt,
	}
	return ingestion
}

func deriveWorkflowHint(text string) string {
	switch {
	case containsAny(text, []string{"incident", "failing", "broken", "debug", "outage"}):
		return "incident"
	case containsAny(text, []string{"feature", "request", "product", "need"}):
		return "feature-request"
	default:
		return "architecture"
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
	return len(sub) > 0 && len(s) >= len(sub) && (containsFold(s, sub))
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
	if _, ok := s.traces[traceID]; !ok {
		return queue.WorkItem{}, errors.New("trace not found")
	}
	item := queue.WorkItem{
		ID:           fmt.Sprintf("replay-%s-%d", traceID, time.Now().Unix()),
		Queue:        queue.EvalQueue,
		Kind:         "trace_replay",
		TraceID:      traceID,
		RequestedBy:  requestedBy,
		ApprovalMode: "ui",
		CreatedAt:    time.Now().UTC(),
	}
	trace := s.traces[traceID]
	trace.Summary.Status = events.StatusReplayed
	s.traces[traceID] = trace
	return item, nil
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
	proposal.Reviews = append(proposal.Reviews, decision)
	proposal.Reviewer = decision.ReviewerID
	proposal.Status = decision.Decision
	s.proposals[proposalID] = proposal
	return proposal, nil
}

func (s *MemoryStore) ExecuteTool(name string, input map[string]interface{}) ToolResult {
	return ToolResult{
		Name:       name,
		Approved:   name != "github.create_pr",
		ExecutedAt: time.Now().UTC(),
		Input:      input,
		Output: map[string]interface{}{
			"status":  "ok",
			"message": fmt.Sprintf("stubbed tool execution for %s", name),
		},
	}
}

