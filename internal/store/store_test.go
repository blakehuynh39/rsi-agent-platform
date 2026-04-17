package store

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func queueActionIntentForTest(t *testing.T, store interface {
	SubmitCommand(transition.CommandEnvelope) (transition.CommandReceipt, error)
	GetActionIntent(string) (action.Intent, bool)
}, intent action.Intent, commandID string) action.Intent {
	t.Helper()
	occurredAt := intent.CreatedAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	if intent.ID == "" {
		intent.ID = nextID("action", 0)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: intent.ID,
		CommandKind: string(transition.CommandActionQueue),
		CommandID:   commandID,
		OccurredAt:  occurredAt,
		Payload: map[string]any{
			"owner_plane":     intent.OwnerPlane,
			"conversation_id": intent.ConversationID,
			"case_id":         intent.CaseID,
			"trace_id":        intent.TraceID,
			"proposal_id":     intent.ProposalID,
			"attempt_id":      intent.AttemptID,
			"kind":            string(intent.Kind),
			"phase_key":       intent.PhaseKey,
			"target_ref":      intent.TargetRef,
			"request_payload": intent.RequestPayload,
			"idempotency_key": intent.IdempotencyKey,
			"approval_mode":   intent.ApprovalMode,
			"approval_state":  intent.ApprovalState,
			"requested_by":    intent.RequestedBy,
			"rationale":       intent.Rationale,
			"evidence_refs":   intent.EvidenceRefs,
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(action_queued) error = %v", err)
	}
	item, ok := store.GetActionIntent(intent.ID)
	if !ok {
		t.Fatalf("expected action intent %s", intent.ID)
	}
	return item
}

func submitActionCommandForTest(t *testing.T, store interface {
	SubmitCommand(transition.CommandEnvelope) (transition.CommandReceipt, error)
}, actionID string, kind transition.ActionExecutionCommandKind, commandID string, occurredAt time.Time, payload map[string]any) transition.CommandReceipt {
	t.Helper()
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: actionID,
		CommandKind: string(kind),
		CommandID:   commandID,
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
	if err != nil {
		t.Fatalf("SubmitCommand(%s) error = %v", kind, err)
	}
	return receipt
}

func submitProposalCommandForTest(t *testing.T, store interface {
	SubmitCommand(transition.CommandEnvelope) (transition.CommandReceipt, error)
}, proposalID string, kind transition.ProposalLineCommandKind, commandID string, payload map[string]any) transition.CommandReceipt {
	t.Helper()
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposalID,
		CommandKind: string(kind),
		CommandID:   commandID,
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
		Payload:     payload,
	})
	if err != nil {
		t.Fatalf("SubmitCommand(%s) error = %v", kind, err)
	}
	return receipt
}

func submitProblemLineCommandForTest(t *testing.T, store interface {
	SubmitCommand(transition.CommandEnvelope) (transition.CommandReceipt, error)
}, aggregateID string, kind transition.ProblemLineCommandKind, commandID string, actor string, occurredAt time.Time, payload map[string]any) transition.CommandReceipt {
	t.Helper()
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: aggregateID,
		CommandKind: string(kind),
		CommandID:   commandID,
		Actor:       firstNonEmpty(actor, "tester"),
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
	if err != nil {
		t.Fatalf("SubmitCommand(%s) error = %v", kind, err)
	}
	return receipt
}

func submitKnowledgeCommandForTest(t *testing.T, store interface {
	SubmitCommand(transition.CommandEnvelope) (transition.CommandReceipt, error)
}, knowledgeID string, kind transition.KnowledgeCommandKind, commandID string, actor string, occurredAt time.Time, payload map[string]any) transition.CommandReceipt {
	t.Helper()
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineKnowledge,
		AggregateID: knowledgeID,
		CommandKind: string(kind),
		CommandID:   commandID,
		Actor:       firstNonEmpty(actor, "tester"),
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
	if err != nil {
		t.Fatalf("SubmitCommand(%s) error = %v", kind, err)
	}
	return receipt
}

func submitSettingsCommandForTest(t *testing.T, store interface {
	SubmitCommand(transition.CommandEnvelope) (transition.CommandReceipt, error)
}, commandID string, actor string, occurredAt time.Time, payload map[string]any) transition.CommandReceipt {
	t.Helper()
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineSettings,
		AggregateID: "settings",
		CommandKind: string(transition.CommandSettingsUpdate),
		CommandID:   commandID,
		Actor:       firstNonEmpty(actor, "tester"),
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
	if err != nil {
		t.Fatalf("SubmitCommand(settings_update) error = %v", err)
	}
	return receipt
}

func submitIngressCommandForTest(t *testing.T, store interface {
	SubmitCommand(transition.CommandEnvelope) (transition.CommandReceipt, error)
}, aggregateID string, kind transition.IngressCommandKind, commandID string, actor string, occurredAt time.Time, payload map[string]any) transition.CommandReceipt {
	t.Helper()
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineIngress,
		AggregateID: aggregateID,
		CommandKind: string(kind),
		CommandID:   commandID,
		Actor:       firstNonEmpty(actor, "tester"),
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
	if err != nil {
		t.Fatalf("SubmitCommand(%s) error = %v", kind, err)
	}
	return receipt
}

func findEvalRunForReceipt(store interface {
	ListEvalRuns() []evals.Run
	ListEvalJudgments(string) []evals.Judgment
}, receipt transition.CommandReceipt) (evals.Run, []evals.Judgment, bool) {
	for _, run := range store.ListEvalRuns() {
		if run.ID == receipt.ResultRef {
			return run, store.ListEvalJudgments(run.ID), true
		}
	}
	return evals.Run{}, nil, false
}

func loadPromotionResultForReceipt(store interface {
	ListDomainEvents() []transition.DomainEvent
	GetProposalSlots() ProposalSlotState
}, receipt transition.CommandReceipt) (PromotionResult, error) {
	if receipt.DecisionKind == transition.DecisionNoop {
		slots := store.GetProposalSlots()
		return PromotionResult{
			BlockedByCap:     slots.Available == 0,
			StaleProposalIDs: slots.StaleProposalIDs,
		}, nil
	}
	for _, item := range store.ListDomainEvents() {
		if strings.TrimSpace(item.CommandID) != strings.TrimSpace(receipt.CommandID) {
			continue
		}
		if strings.TrimSpace(item.EventKind) != "problem_line_promoted" {
			continue
		}
		return PromotionResult{
			Promoted:         int(int64FromAnyForTest(item.Payload["promoted"])),
			PromotedIDs:      stringSliceFromAnyForTest(item.Payload["promoted_ids"]),
			BlockedByCap:     boolFromAnyForTest(item.Payload["blocked_by_cap"]),
			StaleProposalIDs: stringSliceFromAnyForTest(item.Payload["stale_proposal_ids"]),
		}, nil
	}
	return PromotionResult{}, errors.New("promotion result event not found")
}

func boolFromAnyForTest(raw any) bool {
	value, ok := raw.(bool)
	return ok && value
}

func int64FromAnyForTest(raw any) int64 {
	switch value := raw.(type) {
	case int:
		return int64(value)
	case int64:
		return value
	case float64:
		return int64(value)
	default:
		return 0
	}
}

func stringSliceFromAnyForTest(raw any) []string {
	switch value := raw.(type) {
	case []string:
		return append([]string(nil), value...)
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			text := strings.TrimSpace(item.(string))
			if text != "" {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func TestMemoryStoreRatingAndReplay(t *testing.T) {
	store := NewMemoryStore()
	traces := store.ListTraces()
	if len(traces) == 0 {
		t.Fatal("expected seeded traces")
	}
	traceID := traces[0].TraceID

	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: traceID,
		CommandKind: string(transition.CommandProblemLineRecordRating),
		CommandID:   "cmd-problem-line-rating",
		Actor:       "alice",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"score":       4,
			"verdict":     "partial",
			"labels":      []string{"needs-human"},
			"notes":       "Useful investigation, incomplete mitigation.",
			"reviewer_id": "alice",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(problem_line_record_rating) error = %v", err)
	}

	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if trace.Summary.LastVerdict != "partial" {
		t.Fatalf("expected last verdict to be updated, got %q", trace.Summary.LastVerdict)
	}

	receipt := submitProblemLineCommandForTest(t, store, traceID, transition.CommandProblemLineScheduleReplay, "cmd-problem-line-replay", "alice", time.Now().UTC(), map[string]any{
		"requested_by": "alice",
	})
	if receipt.ResultRef != traceID {
		t.Fatalf("expected replay receipt to reference trace %s, got %s", traceID, receipt.ResultRef)
	}
	if _, ok := store.GetCommandReceipt(receipt.CommandID + ":evaluate"); !ok {
		t.Fatalf("expected follow-on eval receipt for %s", receipt.CommandID)
	}
	foundEvalEffect := false
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind == transition.MachineProblemLine && effect.AggregateID == traceID && effect.EffectKind == transition.EffectInvokeRunner && effect.Status == transition.EffectQueued {
			foundEvalEffect = true
			break
		}
	}
	if !foundEvalEffect {
		t.Fatal("expected queued eval runner effect after replay")
	}
}

func TestMemoryStoreAddFeedbackConversationAnchorsTrace(t *testing.T) {
	store := NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatal("expected seeded trace")
	}
	receipt := submitProblemLineCommandForTest(t, store, traceID, transition.CommandProblemLineRecordFeedback, "cmd-problem-line-feedback-conversation", "alice", time.Now().UTC(), map[string]any{
		"target_type": string(review.FeedbackTargetConversation),
		"target_id":   trace.Summary.ConversationID,
		"verdict":     "useful",
		"reviewer_id": "alice",
	})
	items := store.ListFeedback(traceID)
	var item review.FeedbackRecord
	found := false
	for _, candidate := range items {
		if candidate.ID == receipt.ResultRef {
			item = candidate
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected feedback %s", receipt.ResultRef)
	}
	if item.TraceID != traceID {
		t.Fatalf("expected feedback to anchor to trace %s, got %s", traceID, item.TraceID)
	}
}

func TestMemoryStoreSubmitCommandProblemLineFeedbackPersistsRecord(t *testing.T) {
	store := NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatal("expected seeded trace")
	}
	now := time.Now().UTC()
	intent := queueActionIntentForTest(t, store, action.Intent{
		ConversationID: trace.Summary.ConversationID,
		CaseID:         trace.Summary.CaseID,
		TraceID:        traceID,
		Kind:           action.KindToolRead,
		CreatedAt:      now,
	}, "cmd-problem-feedback-action")
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: traceID,
		CommandKind: string(transition.CommandProblemLineRecordFeedback),
		CommandID:   "cmd-problem-feedback-store",
		Actor:       "alice",
		OccurredAt:  now,
		Payload: map[string]any{
			"trace_id":    traceID,
			"target_type": string(review.FeedbackTargetActionIntent),
			"target_id":   intent.ID,
			"notes":       "grounded feedback",
			"reviewer_id": "alice",
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}
	items := store.ListFeedback(traceID)
	if len(items) != 1 {
		t.Fatalf("expected one feedback record, got %d", len(items))
	}
	if items[0].ID != receipt.ResultRef {
		t.Fatalf("expected result ref %s, got %+v", receipt.ResultRef, items[0])
	}
	if items[0].TargetType != review.FeedbackTargetActionIntent {
		t.Fatalf("expected action_intent target type, got %+v", items[0])
	}
}

func TestMemoryStoreSubmitCommandProblemLineReplayQueuesEvalCommandAndEffect(t *testing.T) {
	store := NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: traceID,
		CommandKind: string(transition.CommandProblemLineScheduleReplay),
		CommandID:   "cmd-problem-replay-store",
		Actor:       "alice",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"requested_by": "alice",
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}
	if receipt.ResultRef != traceID {
		t.Fatalf("expected trace result ref, got %s", receipt.ResultRef)
	}
	if _, ok := store.GetCommandReceipt(receipt.CommandID + ":evaluate"); !ok {
		t.Fatalf("expected follow-on eval receipt for %s", receipt.CommandID)
	}
	foundEffect := false
	for _, item := range store.ListEffectExecutions() {
		if item.MachineKind == transition.MachineProblemLine && item.AggregateID == traceID && item.EffectKind == transition.EffectInvokeRunner && item.Status == transition.EffectQueued {
			foundEffect = true
			break
		}
	}
	if !foundEffect {
		t.Fatal("expected replay to queue a problem-line invoke_runner effect")
	}
}

func TestMemoryStoreSetThreadState(t *testing.T) {
	store := NewMemoryStore()

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineThreadPolicy,
		AggregateID: "slack:CENG:171000001.000100",
		CommandKind: string(transition.CommandThreadMute),
		CommandID:   "cmd-thread-policy-set-state",
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"owner_bot": "tester",
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand(thread_mute) error = %v", err)
	}
	item, ok := findThreadPolicyByKey(store.ListThreadPolicies(), receipt.AggregateID)
	if !ok {
		t.Fatalf("expected thread policy %s", receipt.AggregateID)
	}
	if !item.Muted {
		t.Fatal("expected muted flag to be set")
	}
}

func TestMemoryStoreSubmitCommandThreadPolicyMuteIsIdempotent(t *testing.T) {
	store := NewMemoryStore()
	command := transition.CommandEnvelope{
		MachineKind: transition.MachineThreadPolicy,
		AggregateID: "slack:CENG:171000001.000100",
		CommandKind: string(transition.CommandThreadMute),
		CommandID:   "cmd-thread-mute",
		OccurredAt:  time.Now().UTC(),
	}

	receipt, err := store.SubmitCommand(command)
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}
	item, ok := findThreadPolicyByKey(store.ListThreadPolicies(), command.AggregateID)
	if !ok {
		t.Fatal("expected thread policy to exist")
	}
	if item.State != policy.ThreadStateMuted || !item.Muted {
		t.Fatalf("expected muted thread policy, got %+v", item)
	}

	again, err := store.SubmitCommand(command)
	if err != nil {
		t.Fatalf("SubmitCommand(duplicate) error = %v", err)
	}
	if again.CommandID != receipt.CommandID || again.DecisionKind != receipt.DecisionKind {
		t.Fatalf("expected duplicate command receipt to be reused, got %+v", again)
	}
}

func TestMemoryStoreSubmitCommandKnowledgeApprovePromotesEntry(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	knowledgeID := "knowledge-test-entry"
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineKnowledge,
		AggregateID: knowledgeID,
		CommandKind: string(transition.CommandKnowledgeRecordDraft),
		CommandID:   "cmd-knowledge-draft",
		Actor:       "tester",
		OccurredAt:  now,
		Payload: map[string]any{
			"tier":        string(knowledge.TierWorking),
			"kind":        string(knowledge.KindFact),
			"scope_type":  string(knowledge.ScopeGlobal),
			"title":       "Test knowledge",
			"status":      string(knowledge.StatusDraft),
			"source_type": string(knowledge.SourceAgent),
			"created_at":  now,
			"updated_at":  now,
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(knowledge_record_draft) error = %v", err)
	}

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineKnowledge,
		AggregateID: knowledgeID,
		CommandKind: string(transition.CommandKnowledgeApprove),
		CommandID:   "cmd-knowledge-approve",
		Actor:       "reviewer",
		OccurredAt:  now,
		Payload: map[string]any{
			"rationale": "Grounded and reusable.",
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}

	reloaded, ok := store.GetKnowledgeEntry(knowledgeID)
	if !ok {
		t.Fatal("expected knowledge entry to exist")
	}
	if reloaded.Status != knowledge.StatusCanonical || reloaded.Tier != knowledge.TierCanonical {
		t.Fatalf("expected canonical knowledge, got %+v", reloaded)
	}
	if reviews := store.ListKnowledgeReviews(knowledgeID); len(reviews) != 1 {
		t.Fatalf("expected one persisted knowledge review, got %d", len(reviews))
	}
}

func TestMemoryStoreSubmitCommandProposalApproveMaterializesAttemptAndQueuesFirstExecutablePhase(t *testing.T) {
	store := NewMemoryStore()
	proposal := store.ListProposals()[0]

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposal.ID,
		CommandKind: string(transition.CommandProposalApproveIntervention),
		CommandID:   "cmd-proposal-approve",
		Actor:       "reviewer",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"rationale": "Proceed with bounded remediation.",
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}

	updated, ok := findProposalByID(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatal("expected proposal to exist")
	}
	if updated.Status != review.ProposalApproved {
		t.Fatalf("expected approved proposal, got %s", updated.Status)
	}
	if strings.TrimSpace(updated.CurrentAttemptID) == "" {
		t.Fatalf("expected approved proposal to materialize a current attempt, got %+v", updated)
	}
	attempt, ok := store.GetChangeAttempt(updated.CurrentAttemptID)
	if !ok {
		t.Fatalf("expected change attempt %s", updated.CurrentAttemptID)
	}
	if strings.TrimSpace(attempt.AttemptTraceID) == "" {
		t.Fatalf("expected attempt %s to have a derived trace, got %+v", attempt.ID, attempt)
	}
	if _, ok := store.GetTrace(attempt.AttemptTraceID); !ok {
		t.Fatalf("expected derived attempt trace %s", attempt.AttemptTraceID)
	}

	foundAttemptEffect := false
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind != transition.MachineAttempt || effect.AggregateID != attempt.ID || effect.Status != transition.EffectQueued {
			continue
		}
		switch effect.EffectKind {
		case transition.EffectOpenWorkspace, transition.EffectInvokeRunner:
			foundAttemptEffect = true
		default:
			t.Fatalf("unexpected first attempt bootstrap effect %s", effect.EffectKind)
		}
	}
	if !foundAttemptEffect {
		t.Fatal("expected approved proposal command to queue the first attempt effect directly")
	}
}

func TestMemoryStoreSubmitCommandWorkflowCompletesThroughReducerStates(t *testing.T) {
	store := NewMemoryStore()
	workflow := store.ListWorkflows()[0]
	trace, ok := store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", workflow.TraceID)
	}
	initialEvents := len(trace.Events)
	now := time.Now().UTC()

	commands := []transition.CommandEnvelope{
		{
			MachineKind: transition.MachineWorkflow,
			AggregateID: workflow.ID,
			CommandKind: string(transition.CommandWorkflowStarted),
			CommandID:   "cmd-workflow-started",
			OccurredAt:  now,
		},
		{
			MachineKind: transition.MachineWorkflow,
			AggregateID: workflow.ID,
			CommandKind: string(transition.CommandContextSkipped),
			CommandID:   "cmd-workflow-context-skipped",
			OccurredAt:  now.Add(time.Second),
		},
		{
			MachineKind: transition.MachineWorkflow,
			AggregateID: workflow.ID,
			CommandKind: string(transition.CommandRunnerCompleted),
			CommandID:   "cmd-workflow-runner-completed",
			OccurredAt:  now.Add(2 * time.Second),
		},
		{
			MachineKind: transition.MachineWorkflow,
			AggregateID: workflow.ID,
			CommandKind: string(transition.CommandReplyPosted),
			CommandID:   "cmd-workflow-reply-posted",
			OccurredAt:  now.Add(3 * time.Second),
		},
	}
	for _, command := range commands {
		if _, err := store.SubmitCommand(command); err != nil {
			t.Fatalf("SubmitCommand(%s) error = %v", command.CommandKind, err)
		}
	}

	updated, ok := findWorkflowByID(store.ListWorkflows(), workflow.ID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if updated.Status != string(transition.WorkflowStateCompleted) {
		t.Fatalf("expected completed workflow state, got %s", updated.Status)
	}
	if updated.CompletedAt == nil {
		t.Fatal("expected completed workflow to set completed_at")
	}
	if len(store.ListEvalRuns()) == 0 {
		t.Fatal("expected workflow completion to trigger immediate problem-line evaluation")
	}
	trace, ok = store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected updated trace %s", workflow.TraceID)
	}
	if trace.Summary.Status != events.StatusCompleted {
		t.Fatalf("expected completed trace summary, got %s", trace.Summary.Status)
	}
	if len(trace.Events) != initialEvents+2 {
		t.Fatalf("expected workflow started and completed events, got %d new events", len(trace.Events)-initialEvents)
	}
	if trace.Events[len(trace.Events)-2].EventType != "workflow.started" {
		t.Fatalf("expected workflow.started event, got %s", trace.Events[len(trace.Events)-2].EventType)
	}
	if trace.Events[len(trace.Events)-1].EventType != "workflow.completed" {
		t.Fatalf("expected workflow.completed event, got %s", trace.Events[len(trace.Events)-1].EventType)
	}
}

func TestMemoryStoreSubmitCommandWorkflowFailurePersistsLastError(t *testing.T) {
	store := NewMemoryStore()
	workflow := store.ListWorkflows()[0]
	trace, ok := store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", workflow.TraceID)
	}
	initialEvents := len(trace.Events)
	now := time.Now().UTC()

	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineWorkflow,
		AggregateID: workflow.ID,
		CommandKind: string(transition.CommandWorkflowFailed),
		CommandID:   "cmd-workflow-failed",
		OccurredAt:  now,
		Payload: map[string]any{
			"last_error":         "runner response missing structured_output",
			"runner_diagnostics": map[string]any{"failure_kind": "structured_output_parse_failure"},
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(workflow_failed) error = %v", err)
	}

	updated, ok := findWorkflowByID(store.ListWorkflows(), workflow.ID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if updated.Status != string(transition.WorkflowStateFailed) {
		t.Fatalf("expected failed workflow state, got %s", updated.Status)
	}
	if updated.LastError != "runner response missing structured_output" {
		t.Fatalf("expected workflow last error to persist, got %q", updated.LastError)
	}
	if updated.RunnerDiagnostics["failure_kind"] != "structured_output_parse_failure" {
		t.Fatalf("expected runner diagnostics to persist, got %#v", updated.RunnerDiagnostics)
	}
	if updated.CompletedAt == nil {
		t.Fatal("expected failed workflow to set completed_at")
	}
	if len(store.ListEvalRuns()) == 0 {
		t.Fatal("expected failed workflow to trigger immediate problem-line evaluation")
	}
	trace, ok = store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected updated trace %s", workflow.TraceID)
	}
	if trace.Summary.Status != events.StatusFailed {
		t.Fatalf("expected failed trace summary, got %s", trace.Summary.Status)
	}
	if len(trace.Events) != initialEvents+1 {
		t.Fatalf("expected one projected workflow failure event, got %d", len(trace.Events)-initialEvents)
	}
	lastEvent := trace.Events[len(trace.Events)-1]
	if lastEvent.EventType != "workflow.failed" {
		t.Fatalf("expected workflow.failed event, got %s", lastEvent.EventType)
	}
	if !strings.Contains(lastEvent.Description, "runner response missing structured_output") {
		t.Fatalf("expected failure description to persist, got %q", lastEvent.Description)
	}
}

func TestMemoryStoreSubmitCommandContextTransitionsProjectTraceArtifacts(t *testing.T) {
	t.Run("context actions queued", func(t *testing.T) {
		store := NewMemoryStore()
		workflow := store.ListWorkflows()[0]
		trace, ok := store.GetTrace(workflow.TraceID)
		if !ok {
			t.Fatalf("expected trace %s", workflow.TraceID)
		}
		now := time.Now().UTC()
		toolRequested := events.TraceEvent{
			TraceID:     trace.Summary.TraceID,
			IngestionID: trace.Summary.IngestionID,
			WorkflowID:  trace.Summary.WorkflowID,
			Plane:       "control",
			Service:     "tool-gateway",
			Actor:       "arch",
			EventType:   "tool.requested",
			Status:      events.StatusQueued,
			StartedAt:   now,
			Description: "Requested github.repo_activity.",
		}
		if _, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineWorkflow,
			AggregateID: workflow.ID,
			CommandKind: string(transition.CommandWorkflowStarted),
			CommandID:   "cmd-context-projection-started",
			OccurredAt:  now,
		}); err != nil {
			t.Fatalf("SubmitCommand(workflow_started) error = %v", err)
		}
		if _, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineWorkflow,
			AggregateID: workflow.ID,
			CommandKind: string(transition.CommandContextActionsQueued),
			CommandID:   "cmd-context-projection-actions",
			OccurredAt:  now,
			Payload: map[string]any{
				"tool_count":   1,
				"trace_events": []events.TraceEvent{toolRequested},
			},
		}); err != nil {
			t.Fatalf("SubmitCommand(context_actions_queued) error = %v", err)
		}
		trace, ok = store.GetTrace(workflow.TraceID)
		if !ok {
			t.Fatalf("expected trace %s", workflow.TraceID)
		}
		for _, event := range trace.Events {
			if event.EventType == "tool.requested" {
				return
			}
		}
		t.Fatal("expected tool.requested event to be projected from context_actions_queued")
	})

	t.Run("context skipped", func(t *testing.T) {
		store := NewMemoryStore()
		workflow := store.ListWorkflows()[0]
		trace, ok := store.GetTrace(workflow.TraceID)
		if !ok {
			t.Fatalf("expected trace %s", workflow.TraceID)
		}
		now := time.Now().UTC()
		runnerStarted := events.TraceEvent{
			TraceID:     trace.Summary.TraceID,
			IngestionID: trace.Summary.IngestionID,
			WorkflowID:  trace.Summary.WorkflowID,
			Plane:       "execution",
			Service:     "runner",
			Actor:       "arch",
			EventType:   "runner.started",
			Status:      events.StatusRunning,
			StartedAt:   now.Add(time.Second),
			Description: "Runner task dispatched with verbose reasoning enabled.",
		}
		if _, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineWorkflow,
			AggregateID: workflow.ID,
			CommandKind: string(transition.CommandWorkflowStarted),
			CommandID:   "cmd-context-skipped-started",
			OccurredAt:  now,
		}); err != nil {
			t.Fatalf("SubmitCommand(workflow_started) error = %v", err)
		}
		if _, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineWorkflow,
			AggregateID: workflow.ID,
			CommandKind: string(transition.CommandContextSkipped),
			CommandID:   "cmd-context-projection-skipped",
			OccurredAt:  now.Add(time.Second),
			Payload: map[string]any{
				"tool_count":   0,
				"trace_events": []events.TraceEvent{runnerStarted},
			},
		}); err != nil {
			t.Fatalf("SubmitCommand(context_skipped) error = %v", err)
		}
		trace, ok = store.GetTrace(workflow.TraceID)
		if !ok {
			t.Fatalf("expected trace %s", workflow.TraceID)
		}
		for _, event := range trace.Events {
			if event.EventType == "runner.started" {
				return
			}
		}
		t.Fatal("expected runner.started event to be projected from context_skipped")
	})
}

func TestMemoryStoreSubmitCommandWorkflowBlockedProjectsNeedsHumanTrace(t *testing.T) {
	store := NewMemoryStore()
	workflow := store.ListWorkflows()[0]
	trace, ok := store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", workflow.TraceID)
	}
	initialEvents := len(trace.Events)
	now := time.Now().UTC()

	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineWorkflow,
		AggregateID: workflow.ID,
		CommandKind: string(transition.CommandWorkflowBlocked),
		CommandID:   "cmd-workflow-blocked",
		OccurredAt:  now,
		Payload: map[string]any{
			"last_error": "channel_autopost_disabled",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(workflow_blocked) error = %v", err)
	}

	updated, ok := findWorkflowByID(store.ListWorkflows(), workflow.ID)
	if !ok {
		t.Fatal("expected workflow to exist")
	}
	if updated.Status != string(transition.WorkflowStateNeedsHuman) {
		t.Fatalf("expected needs_human workflow state, got %s", updated.Status)
	}
	trace, ok = store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected updated trace %s", workflow.TraceID)
	}
	if trace.Summary.Status != events.StatusNeedsHuman {
		t.Fatalf("expected needs-human trace summary, got %s", trace.Summary.Status)
	}
	if len(trace.Events) != initialEvents+1 {
		t.Fatalf("expected one projected workflow blocked event, got %d", len(trace.Events)-initialEvents)
	}
	lastEvent := trace.Events[len(trace.Events)-1]
	if lastEvent.EventType != "workflow.blocked" {
		t.Fatalf("expected workflow.blocked event, got %s", lastEvent.EventType)
	}
	if !strings.Contains(lastEvent.Description, "channel_autopost_disabled") {
		t.Fatalf("expected block description to persist, got %q", lastEvent.Description)
	}
}

func TestMemoryStoreSubmitCommandWorkflowBlockedProjectsAttachedFailureArtifacts(t *testing.T) {
	store := NewMemoryStore()
	workflow := store.ListWorkflows()[0]
	trace, ok := store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", workflow.TraceID)
	}
	now := time.Now().UTC()
	detailEvent := events.TraceEvent{
		TraceID:        trace.Summary.TraceID,
		IngestionID:    trace.Summary.IngestionID,
		WorkflowID:     trace.Summary.WorkflowID,
		ConversationID: trace.Summary.ConversationID,
		CaseID:         trace.Summary.CaseID,
		TriggerEventID: trace.Summary.TriggerEventID,
		Plane:          "control",
		Service:        "control-plane",
		Actor:          "action-worker",
		EventType:      "action.persistence_failed",
		Status:         events.StatusNeedsHuman,
		StartedAt:      now,
		Description:    "subsystem=shared-store failure_mode=action_result_primary_key_collision sqlstate=23505",
	}

	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineWorkflow,
		AggregateID: workflow.ID,
		CommandKind: string(transition.CommandWorkflowBlocked),
		CommandID:   "cmd-workflow-blocked-with-detail",
		OccurredAt:  now,
		Payload: map[string]any{
			"last_error":   "waiting on operator guidance",
			"trace_events": []events.TraceEvent{detailEvent},
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(workflow_blocked) error = %v", err)
	}

	trace, ok = store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected updated trace %s", workflow.TraceID)
	}
	found := false
	for _, event := range trace.Events {
		if event.EventType == "action.persistence_failed" {
			found = true
			if !strings.Contains(event.Description, "sqlstate=23505") {
				t.Fatalf("expected SQLSTATE detail in failure event, got %q", event.Description)
			}
		}
	}
	if !found {
		t.Fatal("expected action.persistence_failed event to be projected with workflow_blocked")
	}
}

func TestMemoryStoreSubmitCommandActionQueueCreatesIntent(t *testing.T) {
	store := NewMemoryStore()
	trace := store.ListTraces()[0]
	now := time.Now().UTC()
	command := transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: "action-queued-test",
		CommandKind: string(transition.CommandActionQueue),
		CommandID:   "cmd-action-queue",
		OccurredAt:  now,
		Payload: map[string]any{
			"owner_plane":     "control",
			"conversation_id": trace.ConversationID,
			"case_id":         trace.CaseID,
			"trace_id":        trace.TraceID,
			"proposal_id":     "proposal-action-queued-test",
			"kind":            string(action.KindToolRead),
			"phase_key":       "collect_context",
			"target_ref":      "github.repo_activity",
			"request_payload": map[string]any{"repo": "piplabs/rsi-agent-platform"},
			"idempotency_key": "action-queued-test",
			"approval_mode":   "not_required",
			"approval_state":  "not_required",
			"requested_by":    "control-plane",
			"rationale":       "Collect context via github.repo_activity.",
			"evidence_refs":   []events.EvidenceRef{{Kind: "trace", Ref: trace.TraceID}},
		},
	}

	receipt, err := store.SubmitCommand(command)
	if err != nil {
		t.Fatalf("SubmitCommand(action_queued) error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}

	intent, ok := store.GetActionIntent(command.AggregateID)
	if !ok {
		t.Fatalf("expected action intent %s to exist", command.AggregateID)
	}
	if intent.Status != action.StatusQueued {
		t.Fatalf("expected queued action status, got %s", intent.Status)
	}
	if intent.IdempotencyKey != "action-queued-test" {
		t.Fatalf("expected idempotency key to persist, got %q", intent.IdempotencyKey)
	}
	if intent.ProposalID != "proposal-action-queued-test" {
		t.Fatalf("expected proposal id to persist, got %q", intent.ProposalID)
	}
	if got := intent.RequestPayload["repo"]; got != "piplabs/rsi-agent-platform" {
		t.Fatalf("expected request payload to persist, got %#v", intent.RequestPayload)
	}
}

func TestMemoryStoreSubmitCommandActionQueueIsIdempotent(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	command := transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: "action-queued-idempotent",
		CommandKind: string(transition.CommandActionQueue),
		CommandID:   "cmd-action-queue-idempotent",
		OccurredAt:  now,
		Payload: map[string]any{
			"owner_plane":     "control",
			"trace_id":        store.ListTraces()[0].TraceID,
			"kind":            string(action.KindSlackPost),
			"target_ref":      "CENG",
			"request_payload": map[string]any{"body": "reply"},
			"idempotency_key": "action-queued-idempotent",
			"approval_mode":   "not_required",
			"approval_state":  "approved",
			"requested_by":    "control-plane",
		},
	}

	if _, err := store.SubmitCommand(command); err != nil {
		t.Fatalf("first SubmitCommand(action_queued) error = %v", err)
	}
	again, err := store.SubmitCommand(command)
	if err != nil {
		t.Fatalf("second SubmitCommand(action_queued) error = %v", err)
	}
	if again.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected duplicate command receipt to reuse advance decision, got %+v", again)
	}
	if intents := store.ListActionIntents(); len(intents) != 1 {
		t.Fatalf("expected one action intent after duplicate queue command, got %d", len(intents))
	}
}

func TestMemoryStoreSubmitCommandActionExecutionPersistsResult(t *testing.T) {
	store := NewMemoryStore()
	intent := queueActionIntentForTest(t, store, action.Intent{
		OwnerPlane:     "control",
		TraceID:        store.ListTraces()[0].TraceID,
		Kind:           action.KindToolRead,
		TargetRef:      "github.repo_activity",
		IdempotencyKey: "action-intent-test",
		CreatedAt:      time.Now().UTC(),
	}, "cmd-action-intent-test")

	started := time.Now().UTC()
	completed := started.Add(time.Second)
	commands := []transition.CommandEnvelope{
		{
			MachineKind: transition.MachineAction,
			AggregateID: intent.ID,
			CommandKind: string(transition.CommandActionStart),
			CommandID:   "cmd-action-start",
			OccurredAt:  started,
			Payload: map[string]any{
				"operation_id": "op-action-test",
			},
		},
		{
			MachineKind: transition.MachineAction,
			AggregateID: intent.ID,
			CommandKind: string(transition.CommandActionSucceed),
			CommandID:   "cmd-action-succeed",
			OccurredAt:  completed,
			Payload: map[string]any{
				"operation_id": "op-action-test",
				"executor":     "tool-gateway",
				"provider":     "github",
				"provider_ref": "provider-ref",
				"started_at":   started,
				"completed_at": completed,
			},
		},
	}
	for _, command := range commands {
		if _, err := store.SubmitCommand(command); err != nil {
			t.Fatalf("SubmitCommand(%s) error = %v", command.CommandKind, err)
		}
	}

	updated, ok := store.GetActionIntent(intent.ID)
	if !ok {
		t.Fatal("expected action intent to exist")
	}
	if updated.Status != action.StatusSucceeded {
		t.Fatalf("expected succeeded action status, got %s", updated.Status)
	}
	results := store.ListActionResults(intent.ID)
	if len(results) != 1 {
		t.Fatalf("expected one persisted action result, got %d", len(results))
	}
	if results[0].OperationID != "op-action-test" {
		t.Fatalf("expected persisted operation id, got %q", results[0].OperationID)
	}
}

func TestMemoryStoreSubmitCommandActionFailureWithoutResult(t *testing.T) {
	store := NewMemoryStore()
	intent := queueActionIntentForTest(t, store, action.Intent{
		OwnerPlane:     "control",
		TraceID:        store.ListTraces()[0].TraceID,
		Kind:           action.KindToolRead,
		TargetRef:      "github.repo_activity",
		IdempotencyKey: "action-intent-fail-test",
		CreatedAt:      time.Now().UTC(),
	}, "cmd-action-intent-fail-test")
	submitActionCommandForTest(t, store, intent.ID, transition.CommandActionStart, "cmd-action-intent-fail-start", time.Now().UTC(), map[string]any{
		"operation_id": "op-action-fail",
	})

	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAction,
		AggregateID: intent.ID,
		CommandKind: string(transition.CommandActionFail),
		CommandID:   "cmd-action-fail",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"operation_id":   "op-action-fail",
			"policy_verdict": "action_result_primary_key_collision",
			"record_result":  false,
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(action_failed) error = %v", err)
	}

	updated, ok := store.GetActionIntent(intent.ID)
	if !ok {
		t.Fatal("expected action intent to exist")
	}
	if updated.Status != action.StatusFailed {
		t.Fatalf("expected failed action status, got %s", updated.Status)
	}
	if updated.PolicyVerdict != "action_result_primary_key_collision" {
		t.Fatalf("expected policy verdict to persist, got %q", updated.PolicyVerdict)
	}
	if results := store.ListActionResults(intent.ID); len(results) != 0 {
		t.Fatalf("expected no action result rows for failure without result, got %d", len(results))
	}
}

func TestMemoryStoreSubmitCommandProjectsToolReadTrace(t *testing.T) {
	store := NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatalf("expected trace %s", traceID)
	}
	now := time.Now().UTC()
	intent := queueActionIntentForTest(t, store, action.Intent{
		OwnerPlane:     "control",
		TraceID:        traceID,
		ConversationID: trace.Summary.ConversationID,
		CaseID:         trace.Summary.CaseID,
		Kind:           action.KindToolRead,
		TargetRef:      "github.repo_activity",
		RequestPayload: map[string]any{"repo": "piplabs/rsi-agent-platform"},
		IdempotencyKey: "action-project-tool-trace",
		CreatedAt:      now,
	}, "cmd-action-project-tool-trace")

	initialEvents := len(trace.Events)
	initialCalls := len(trace.ToolCalls)
	started := now
	completed := now.Add(time.Second)
	for _, command := range []transition.CommandEnvelope{
		{
			MachineKind: transition.MachineAction,
			AggregateID: intent.ID,
			CommandKind: string(transition.CommandActionStart),
			CommandID:   "cmd-project-tool-start",
			OccurredAt:  started,
			Payload:     map[string]any{"operation_id": "op-project-tool"},
		},
		{
			MachineKind: transition.MachineAction,
			AggregateID: intent.ID,
			CommandKind: string(transition.CommandActionSucceed),
			CommandID:   "cmd-project-tool-success",
			OccurredAt:  completed,
			Payload: map[string]any{
				"operation_id":    "op-project-tool",
				"executor":        "tool-gateway",
				"provider":        "github",
				"provider_ref":    "tool-provider-ref",
				"tool_call_id":    "tool-call-1",
				"summary":         "Fetched repository activity.",
				"request_payload": map[string]any{"repo": "piplabs/rsi-agent-platform"},
				"raw_artifact_refs": []string{
					"artifact://tool-output",
				},
				"started_at":   started,
				"completed_at": completed,
			},
		},
	} {
		if _, err := store.SubmitCommand(command); err != nil {
			t.Fatalf("SubmitCommand(%s) error = %v", command.CommandKind, err)
		}
	}

	trace, ok = store.GetTrace(traceID)
	if !ok {
		t.Fatalf("expected updated trace %s", traceID)
	}
	if len(trace.Events) != initialEvents+1 {
		t.Fatalf("expected one projected tool event, got %d events", len(trace.Events)-initialEvents)
	}
	lastEvent := trace.Events[len(trace.Events)-1]
	if lastEvent.EventType != "tool.completed" {
		t.Fatalf("expected projected tool.completed event, got %s", lastEvent.EventType)
	}
	if len(trace.ToolCalls) != initialCalls+1 {
		t.Fatalf("expected one projected tool call, got %d", len(trace.ToolCalls)-initialCalls)
	}
	lastCall := trace.ToolCalls[len(trace.ToolCalls)-1]
	if lastCall.ToolCallID != "tool-call-1" {
		t.Fatalf("expected projected tool call id, got %q", lastCall.ToolCallID)
	}
	if lastCall.Summary != "Fetched repository activity." {
		t.Fatalf("expected projected summary, got %q", lastCall.Summary)
	}
	if got := lastCall.Request["repo"]; got != "piplabs/rsi-agent-platform" {
		t.Fatalf("expected projected request payload, got %#v", lastCall.Request)
	}
}

func TestMemoryStoreSubmitCommandProjectsSlackReplyTrace(t *testing.T) {
	store := NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatalf("expected trace %s", traceID)
	}
	now := time.Now().UTC()
	intent := queueActionIntentForTest(t, store, action.Intent{
		OwnerPlane:     "control",
		TraceID:        traceID,
		ConversationID: trace.Summary.ConversationID,
		CaseID:         trace.Summary.CaseID,
		Kind:           action.KindSlackPost,
		TargetRef:      "CENG",
		RequestPayload: map[string]any{
			"channel_id": "CENG",
			"thread_ts":  "171000001.000100",
			"draft_body": "Draft reply",
			"final_body": "Final reply",
			"body":       "Final reply",
		},
		IdempotencyKey: "action-project-slack-trace",
		CreatedAt:      now,
	}, "cmd-action-project-slack-trace")

	initialEvents := len(trace.Events)
	initialActions := len(trace.SlackActions)
	started := now
	completed := now.Add(time.Second)
	for _, command := range []transition.CommandEnvelope{
		{
			MachineKind: transition.MachineAction,
			AggregateID: intent.ID,
			CommandKind: string(transition.CommandActionStart),
			CommandID:   "cmd-project-slack-start",
			OccurredAt:  started,
			Payload:     map[string]any{"operation_id": "op-project-slack"},
		},
		{
			MachineKind: transition.MachineAction,
			AggregateID: intent.ID,
			CommandKind: string(transition.CommandActionSucceed),
			CommandID:   "cmd-project-slack-success",
			OccurredAt:  completed,
			Payload: map[string]any{
				"operation_id":  "op-project-slack",
				"executor":      "tool-gateway",
				"provider":      "slack",
				"provider_ref":  "thread-reply-ref",
				"summary":       "Posted Slack reply.",
				"channel_id":    "CENG",
				"thread_ts":     "171000001.000100",
				"draft_body":    "Draft reply",
				"final_body":    "Final reply",
				"send_status":   "posted",
				"artifact_refs": []string{"artifact://slack-post"},
				"started_at":    started,
				"completed_at":  completed,
			},
		},
	} {
		if _, err := store.SubmitCommand(command); err != nil {
			t.Fatalf("SubmitCommand(%s) error = %v", command.CommandKind, err)
		}
	}

	trace, ok = store.GetTrace(traceID)
	if !ok {
		t.Fatalf("expected updated trace %s", traceID)
	}
	if len(trace.Events) != initialEvents+1 {
		t.Fatalf("expected one projected slack event, got %d events", len(trace.Events)-initialEvents)
	}
	lastEvent := trace.Events[len(trace.Events)-1]
	if lastEvent.EventType != "slack.reply.posted" {
		t.Fatalf("expected projected slack.reply.posted event, got %s", lastEvent.EventType)
	}
	if len(trace.SlackActions) != initialActions+1 {
		t.Fatalf("expected one projected slack action, got %d", len(trace.SlackActions)-initialActions)
	}
	lastAction := trace.SlackActions[len(trace.SlackActions)-1]
	if lastAction.FinalBody != "Final reply" {
		t.Fatalf("expected projected final body, got %q", lastAction.FinalBody)
	}
	if lastAction.SendStatus != "posted" {
		t.Fatalf("expected projected send status, got %q", lastAction.SendStatus)
	}
}

func TestProposalCapEnforced(t *testing.T) {
	store := NewMemoryStore()

	slots := store.GetProposalSlots()
	if slots.Active < 2 {
		store.proposals["proposal-test-cap"] = review.Proposal{
			ID:                  "proposal-test-cap",
			TraceID:             store.ListTraces()[0].TraceID,
			Title:               "Synthetic cap filler",
			Category:            "policy_or_runtime_fix",
			Summary:             "Fill the second slot for cap enforcement.",
			Status:              review.ProposalPendingReview,
			CandidateKey:        "synthetic:incident:failed_workflow",
			ActiveSlotConsuming: true,
			CreatedAt:           time.Now().UTC(),
		}
		slots = store.GetProposalSlots()
	}
	if slots.Active != 2 {
		t.Fatalf("expected 2 active slots for cap test, got %d", slots.Active)
	}

	receipt := submitProblemLineCommandForTest(t, store, "test-promoter", transition.CommandProblemLinePromote, "cmd-problem-line-promote-cap", "test-promoter", time.Now().UTC(), map[string]any{
		"requested_by": "test-promoter",
		"limit":        2,
	})
	result, err := loadPromotionResultForReceipt(store, receipt)
	if err != nil {
		t.Fatalf("loadPromotionResultForReceipt() error = %v", err)
	}
	if !result.BlockedByCap {
		t.Fatal("expected promoter to be blocked by the slot cap")
	}
	if result.Promoted != 0 {
		t.Fatalf("expected no new proposals, got %d", result.Promoted)
	}
}

func TestProposalPromoterLease(t *testing.T) {
	store := NewMemoryStore()

	store.cronLeases["improvement-plane-cron"] = improvement.CronLease{
		Name:      "improvement-plane-cron",
		Holder:    "other-worker",
		ExpiresAt: time.Now().UTC().Add(time.Minute),
	}

	receipt := submitProblemLineCommandForTest(t, store, "test-worker", transition.CommandProblemLinePromote, "cmd-problem-line-promote-lease", "test-worker", time.Now().UTC(), map[string]any{
		"requested_by": "test-worker",
	})
	if receipt.DecisionKind != transition.DecisionReject {
		t.Fatalf("expected rejected promote receipt, got %+v", receipt)
	}
}

func TestRejectedProposalRequiresNewEvidence(t *testing.T) {
	store := NewMemoryStore()
	proposals := store.ListProposals()
	if len(proposals) == 0 {
		t.Fatal("expected seeded proposals")
	}
	proposal := proposals[0]

	if _, err := ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalRejected),
		Rationale:  "Too similar to prior attempt.",
		ReviewerID: "alice",
	}); err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}

	candidate := store.candidates[proposal.CandidateKey]
	candidate.Status = improvement.CandidateQueued
	candidate.NewEvidenceSinceLastRejection = false
	store.candidates[proposal.CandidateKey] = candidate

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: "alice",
		CommandKind: string(transition.CommandProblemLinePromote),
		CommandID:   "cmd-problem-line-promote",
		Actor:       "alice",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"requested_by": "alice",
			"limit":        2,
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand(problem_line_promote) error = %v", err)
	}
	result, err := loadPromotionResultForReceipt(store, receipt)
	if err != nil {
		t.Fatalf("loadPromotionResultForReceipt() error = %v", err)
	}
	for _, promotedID := range result.PromotedIDs {
		if store.proposals[promotedID].CandidateKey == proposal.CandidateKey {
			t.Fatal("expected rejected candidate to stay blocked without new evidence")
		}
	}
}

func TestSettingsBackedProposalCap(t *testing.T) {
	store := NewMemoryStore()

	settingsReceipt := submitSettingsCommandForTest(t, store, "cmd-settings-update", "tester", time.Now().UTC(), map[string]any{
		"active_proposal_cap": 1,
	})
	if settingsReceipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", settingsReceipt)
	}
	settings := store.GetSettings()
	if settings.ActiveProposalCap != 1 {
		t.Fatalf("expected active proposal cap to be 1, got %d", settings.ActiveProposalCap)
	}

	slots := store.GetProposalSlots()
	if slots.Cap != 1 {
		t.Fatalf("expected slot cap to be 1, got %d", slots.Cap)
	}
}

func TestApproveProposalMaterializesAttemptThroughEffects(t *testing.T) {
	store := NewMemoryStore()
	proposals := store.ListProposals()
	if len(proposals) == 0 {
		t.Fatal("expected seeded proposals")
	}

	proposal, err := ReviewProposalForTesting(store, proposals[0].ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with repo-change work.",
		ReviewerID: "alice",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	if proposal.Status != review.ProposalApproved {
		t.Fatalf("expected approved proposal, got %s", proposal.Status)
	}
	if strings.TrimSpace(proposal.CurrentAttemptID) == "" {
		t.Fatalf("expected approved proposal to materialize a current attempt, got %+v", proposal)
	}
	attempt, ok := store.GetChangeAttempt(proposal.CurrentAttemptID)
	if !ok {
		t.Fatalf("expected change attempt %s", proposal.CurrentAttemptID)
	}
	found := false
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind != transition.MachineAttempt || effect.AggregateID != attempt.ID || effect.Status != transition.EffectQueued {
			continue
		}
		switch effect.EffectKind {
		case transition.EffectOpenWorkspace, transition.EffectInvokeRunner:
			found = true
		default:
			t.Fatalf("unexpected approval bootstrap effect %s", effect.EffectKind)
		}
	}
	if !found {
		t.Fatal("expected approval to queue the first attempt effect directly")
	}
}

func TestRetryProposalRepoChangeAfterFailedValidationMaterializesNewAttempt(t *testing.T) {
	store := NewMemoryStore()
	proposal := store.ListProposals()[0]

	approved, err := ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with repo-change work.",
		ReviewerID: "alice",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	currentAttemptID := strings.TrimSpace(approved.CurrentAttemptID)
	if currentAttemptID == "" {
		t.Fatalf("expected current attempt after approval, got %+v", approved)
	}
	currentAttempt, ok := store.GetChangeAttempt(currentAttemptID)
	if !ok {
		t.Fatalf("expected attempt %s after approval", currentAttemptID)
	}

	if _, _, err := AdvanceProposalToFailedValidationForTesting(store, approved.ID, time.Now().UTC()); err != nil {
		t.Fatalf("AdvanceProposalToFailedValidationForTesting() error = %v", err)
	}

	submitProposalCommandForTest(t, store, approved.ID, transition.CommandProposalRetryAttempt, "cmd-proposal-retry-after-failed-validation", nil)

	reloadedProposal, ok := findProposalByID(store.ListProposals(), approved.ID)
	if !ok {
		t.Fatalf("expected proposal %s after retry", approved.ID)
	}
	if strings.TrimSpace(reloadedProposal.CurrentAttemptID) == "" || reloadedProposal.CurrentAttemptID == currentAttempt.ID {
		t.Fatalf("expected retry to materialize a new current attempt, got %q", reloadedProposal.CurrentAttemptID)
	}
}

func findThreadPolicyByKey(items []policy.ThreadPolicy, threadKey string) (policy.ThreadPolicy, bool) {
	for _, item := range items {
		if item.ThreadKey == threadKey {
			return item, true
		}
	}
	return policy.ThreadPolicy{}, false
}

func findProposalByID(items []review.Proposal, proposalID string) (review.Proposal, bool) {
	for _, item := range items {
		if item.ID == proposalID {
			return item, true
		}
	}
	return review.Proposal{}, false
}

func TestEvaluateTraceRuntimeFailureCreatesRootCauseCandidate(t *testing.T) {
	store := NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatalf("expected trace %s", traceID)
	}

	description := `subsystem=shared-store failure_mode=action_result_primary_key_collision provider=github action_intent_id=action-002 effect_execution_id=eff-003 kind=tool_read sqlstate=23505 constraint=action_result_pkey table=action_result error="duplicate key value violates unique constraint \"action_result_pkey\""`
	projectedAt := time.Now().UTC()
	if receipt := submitProblemLineCommandForTest(t, store, traceID, transition.CommandProblemLineProjectTrace, "cmd-problem-line-project-runtime-failure", "tester", projectedAt, map[string]any{
		"trace_id":        traceID,
		"trace_status":    string(events.StatusNeedsHuman),
		"workflow_status": "needs-human",
		"workflow_error":  description,
		"trace_events": []events.TraceEvent{
			{
				TraceID:        trace.Summary.TraceID,
				IngestionID:    trace.Summary.IngestionID,
				WorkflowID:     trace.Summary.WorkflowID,
				ConversationID: trace.Summary.ConversationID,
				CaseID:         trace.Summary.CaseID,
				TriggerEventID: trace.Summary.TriggerEventID,
				Plane:          "control",
				Service:        "control-plane",
				Actor:          "action-worker",
				EventType:      "action.persistence_failed",
				Status:         events.StatusNeedsHuman,
				StartedAt:      projectedAt,
				Description:    description,
			},
		},
	}); receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}

	evalReceipt := submitProblemLineCommandForTest(t, store, traceID, transition.CommandProblemLineEvaluateTrace, "cmd-problem-line-evaluate-runtime-failure", "tester", time.Now().UTC(), map[string]any{
		"trigger": "test",
	})
	run, judgments, ok := findEvalRunForReceipt(store, evalReceipt)
	if !ok {
		t.Fatalf("expected eval run %s", evalReceipt.ResultRef)
	}
	if run.OverallScore >= 0.65 {
		t.Fatalf("expected failing runtime trace to score below promotion threshold, got %.2f (%#v)", run.OverallScore, judgments)
	}

	var candidate improvement.Candidate
	found := false
	for _, item := range store.ListCandidates() {
		if item.LatestTraceID == traceID && item.FailureMode == "action_result_primary_key_collision" {
			candidate = item
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected runtime failure candidate to be created")
	}
	if candidate.CandidateKey != "shared-store:policy_or_runtime_fix:action_result_primary_key_collision" {
		t.Fatalf("unexpected candidate key: %s", candidate.CandidateKey)
	}
	if candidate.Subsystem != "shared-store" {
		t.Fatalf("expected shared-store subsystem, got %s", candidate.Subsystem)
	}
	if candidate.InterventionType != "policy_or_runtime_fix" {
		t.Fatalf("expected policy_or_runtime_fix intervention, got %s", candidate.InterventionType)
	}
	if candidate.Status != improvement.CandidateQueued {
		t.Fatalf("expected queued candidate, got %s", candidate.Status)
	}
	if !strings.Contains(candidate.Hypothesis, "action result") {
		t.Fatalf("expected hypothesis to mention action result failure, got %q", candidate.Hypothesis)
	}
}

func TestEvaluateTraceWorkflowContextBindingFailureCreatesRootCauseCandidate(t *testing.T) {
	store := NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatalf("expected trace %s", traceID)
	}

	now := time.Now().UTC()
	if receipt := submitProblemLineCommandForTest(t, store, traceID, transition.CommandProblemLineProjectTrace, "cmd-problem-line-project-workflow-context", "tester", now, map[string]any{
		"trace_id":        traceID,
		"trace_status":    string(events.StatusFailed),
		"workflow_status": "failed",
		"workflow_error":  "runner response missing structured_output",
		"trace_events": []events.TraceEvent{
			{
				TraceID:        trace.Summary.TraceID,
				IngestionID:    trace.Summary.IngestionID,
				WorkflowID:     trace.Summary.WorkflowID,
				ConversationID: trace.Summary.ConversationID,
				CaseID:         trace.Summary.CaseID,
				TriggerEventID: trace.Summary.TriggerEventID,
				Plane:          "control",
				Service:        "tool-gateway",
				Actor:          "arch",
				EventType:      "tool.failed",
				Status:         events.StatusNeedsHuman,
				StartedAt:      now,
				Description:    "Workflow context requires workflow_id or trace_id bound to a workflow.",
			},
			{
				TraceID:        trace.Summary.TraceID,
				IngestionID:    trace.Summary.IngestionID,
				WorkflowID:     trace.Summary.WorkflowID,
				ConversationID: trace.Summary.ConversationID,
				CaseID:         trace.Summary.CaseID,
				TriggerEventID: trace.Summary.TriggerEventID,
				Plane:          "control",
				Service:        "control-plane",
				Actor:          "worker",
				EventType:      "workflow.failed",
				Status:         events.StatusFailed,
				StartedAt:      now,
				Description:    "runner response missing structured_output",
			},
		},
		"tool_calls": []events.ToolCallRecord{
			{
				ID:         "tool-record-workflow-context",
				TraceID:    trace.Summary.TraceID,
				WorkflowID: trace.Summary.WorkflowID,
				ToolName:   "rsi.workflow_context",
				ToolCallID: "tool-call-workflow-context",
				Request: map[string]interface{}{
					"repo": "depin-backend",
				},
				Summary:   "Workflow context requires workflow_id or trace_id bound to a workflow.",
				Status:    "failed",
				CreatedAt: now,
			},
		},
	}); receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}

	evalReceipt := submitProblemLineCommandForTest(t, store, traceID, transition.CommandProblemLineEvaluateTrace, "cmd-problem-line-evaluate-workflow-context", "tester", time.Now().UTC(), map[string]any{
		"trigger": "test",
	})
	run, judgments, ok := findEvalRunForReceipt(store, evalReceipt)
	if !ok {
		t.Fatalf("expected eval run %s", evalReceipt.ResultRef)
	}
	if run.OverallScore >= 0.65 {
		t.Fatalf("expected workflow-context binding failure to score below promotion threshold, got %.2f (%#v)", run.OverallScore, judgments)
	}

	var candidate improvement.Candidate
	found := false
	for _, item := range store.ListCandidates() {
		if item.LatestTraceID == traceID && item.FailureMode == "workflow_context_binding_failure" {
			candidate = item
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected workflow-context binding failure candidate to be created")
	}
	if candidate.CandidateKey != "control-plane:policy_or_runtime_fix:workflow_context_binding_failure" {
		t.Fatalf("unexpected candidate key: %s", candidate.CandidateKey)
	}
	if candidate.Subsystem != "control-plane" {
		t.Fatalf("expected control-plane subsystem, got %s", candidate.Subsystem)
	}
	if candidate.Status != improvement.CandidateQueued {
		t.Fatalf("expected queued candidate, got %s", candidate.Status)
	}
	if !strings.Contains(strings.ToLower(candidate.Hypothesis), "workflow") {
		t.Fatalf("expected hypothesis to mention workflow binding, got %q", candidate.Hypothesis)
	}
	if candidate.ProposedScope != "control-plane + tool-gateway" {
		t.Fatalf("expected bounded scope, got %q", candidate.ProposedScope)
	}
}

func TestEvaluateTraceRunnerInvalidToolNameContractUsesWorkflowAttemptDiagnostics(t *testing.T) {
	store := NewMemoryStore()
	workflow := store.ListWorkflows()[0]
	trace, ok := store.GetTrace(workflow.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", workflow.TraceID)
	}
	now := time.Now().UTC()
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineWorkflow,
		AggregateID: workflow.ID,
		CommandKind: string(transition.CommandWorkflowFailed),
		CommandID:   "cmd-workflow-invalid-tool-name",
		OccurredAt:  now,
		Payload: map[string]any{
			"last_error":      "OpenAI rejected tools[0].name",
			"failure_class":   "runner_invalid_request",
			"failure_summary": "OpenAI rejected tools[0].name because dotted RSI tool names crossed the provider boundary.",
			"runner_diagnostics": map[string]any{
				"failure_kind":           "invalid_request",
				"provider_status_code":   400,
				"provider_error_param":   "tools[0].name",
				"provider_error_code":    "invalid_value",
				"provider_error_message": "Invalid 'tools[0].name': string does not match pattern '^[A-Za-z0-9_-]+$'",
				"invalid_tool_names":     []any{"repo.context", "rsi.workflow_context"},
			},
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(workflow_failed invalid tool name) error = %v", err)
	}

	evalReceipt := submitProblemLineCommandForTest(t, store, trace.Summary.TraceID, transition.CommandProblemLineEvaluateTrace, "cmd-problem-line-evaluate-invalid-tool-name", "tester", time.Now().UTC(), map[string]any{
		"trigger": "test",
	})
	run, judgments, ok := findEvalRunForReceipt(store, evalReceipt)
	if !ok {
		t.Fatalf("expected eval run %s", evalReceipt.ResultRef)
	}
	if run.OverallScore >= 0.65 {
		t.Fatalf("expected invalid tool-name trace to score below promotion threshold, got %.2f (%#v)", run.OverallScore, judgments)
	}

	var candidate improvement.Candidate
	found := false
	for _, item := range store.ListCandidates() {
		if item.LatestTraceID == trace.Summary.TraceID && item.FailureMode == "runner_invalid_tool_name_contract" {
			candidate = item
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected runner invalid tool-name contract candidate to be created")
	}
	if candidate.CandidateKey != "runner:policy_or_runtime_fix:runner_invalid_tool_name_contract" {
		t.Fatalf("unexpected candidate key: %s", candidate.CandidateKey)
	}
	if candidate.Subsystem != "runner" {
		t.Fatalf("expected runner subsystem, got %s", candidate.Subsystem)
	}
	if candidate.Status != improvement.CandidateQueued {
		t.Fatalf("expected queued candidate, got %s", candidate.Status)
	}
	if !strings.Contains(candidate.Hypothesis, "tools[0].name") {
		t.Fatalf("expected hypothesis to mention tools[0].name, got %q", candidate.Hypothesis)
	}
	if candidate.ProposedScope != "runner + control-plane + improvement-plane" {
		t.Fatalf("expected bounded scope, got %q", candidate.ProposedScope)
	}
}

func TestUngroundedFailedWorkflowCandidateStaysNeedsEvidenceAndDoesNotPromote(t *testing.T) {
	store := NewMemoryStore()
	store.proposals = map[string]review.Proposal{}
	for key, item := range store.candidates {
		item.Status = improvement.CandidateNeedsEvidence
		store.candidates[key] = item
	}
	workflow := store.ListWorkflows()[0]
	initialProposalCount := len(store.ListProposals())
	now := time.Now().UTC()
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineWorkflow,
		AggregateID: workflow.ID,
		CommandKind: string(transition.CommandWorkflowFailed),
		CommandID:   "cmd-workflow-generic-failure",
		OccurredAt:  now,
		Payload: map[string]any{
			"last_error": "workflow failed",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(workflow_failed generic) error = %v", err)
	}

	evalReceipt := submitProblemLineCommandForTest(t, store, workflow.TraceID, transition.CommandProblemLineEvaluateTrace, "cmd-problem-line-evaluate-generic-failure", "tester", now.Add(time.Second), map[string]any{
		"trigger": "test",
	})
	if _, _, ok := findEvalRunForReceipt(store, evalReceipt); !ok {
		t.Fatalf("expected eval run %s", evalReceipt.ResultRef)
	}

	var candidate improvement.Candidate
	found := false
	for _, item := range store.ListCandidates() {
		if item.LatestTraceID == workflow.TraceID && item.FailureMode == "failed_workflow" {
			candidate = item
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected generic failed_workflow candidate to be created")
	}
	if candidate.Status != improvement.CandidateNeedsEvidence {
		t.Fatalf("expected candidate to stay needs_evidence, got %s", candidate.Status)
	}
	if candidate.ProposedScope != "whole_repo" && candidate.ProposedScope != "platform + adapters" {
		t.Fatalf("expected generic broad scope, got %q", candidate.ProposedScope)
	}

	promoteReceipt := submitProblemLineCommandForTest(t, store, "test-promoter", transition.CommandProblemLinePromote, "cmd-problem-line-promote-generic-failure", "tester", now.Add(2*time.Second), map[string]any{
		"limit": 1,
	})
	result, err := loadPromotionResultForReceipt(store, promoteReceipt)
	if err != nil {
		t.Fatalf("loadPromotionResultForReceipt() error = %v", err)
	}
	if result.Promoted != 0 {
		t.Fatalf("expected no promoted proposals, got %+v", result)
	}
	if len(store.ListProposals()) != initialProposalCount {
		t.Fatalf("expected proposal count to remain %d, got %d", initialProposalCount, len(store.ListProposals()))
	}
}

func statusPtr(status events.Status) *events.Status {
	return &status
}
