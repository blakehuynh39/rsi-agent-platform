package store

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func TestMemoryStoreSubmitCommandHarnessActivateOverlayPersistsOverlayAndExperiment(t *testing.T) {
	store := newEmptyMemoryStore()
	now := time.Now().UTC()

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineHarness,
		AggregateID: "overlay-1",
		CommandKind: string(transition.CommandHarnessActivateOverlay),
		CommandID:   "cmd-harness-activate",
		Actor:       "improvement-plane",
		OccurredAt:  now,
		Payload: map[string]any{
			"profile_id":            "harness-profile-prod",
			"role":                  "prod",
			"version":               "overlay-v1",
			"target_kind":           "runner_role",
			"target_ref":            "prod",
			"proposal_id":           "proposal-1",
			"prompt_fragments":      []string{"Ground answers in explicit evidence."},
			"tool_preference_order": []string{"repo.context"},
			"memory_read_enabled":   true,
			"memory_write_enabled":  false,
			"experiment_id":         "hexp-1",
			"attempt_id":            "attempt-1",
			"experiment_summary":    "Activated bounded overlay.",
			"experiment_metrics": map[string]any{
				"target_layer": "harness_overlay",
			},
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}
	if receipt.ResultRef != "overlay-1" {
		t.Fatalf("expected overlay result ref, got %q", receipt.ResultRef)
	}

	overlay, ok := store.harnessOverlays["overlay-1"]
	if !ok {
		t.Fatal("expected persisted harness overlay")
	}
	if overlay.Status != harness.OverlayStatusActive || overlay.Version != "overlay-v1" {
		t.Fatalf("unexpected overlay %+v", overlay)
	}
	if overlay.MemoryReadEnabled == nil || !*overlay.MemoryReadEnabled {
		t.Fatalf("expected memory_read_enabled=true, got %+v", overlay.MemoryReadEnabled)
	}
	if overlay.MemoryWriteEnabled == nil || *overlay.MemoryWriteEnabled {
		t.Fatalf("expected memory_write_enabled=false, got %+v", overlay.MemoryWriteEnabled)
	}

	experiment, ok := store.harnessExperiments["hexp-1"]
	if !ok {
		t.Fatal("expected persisted harness experiment")
	}
	if experiment.AttemptID != "attempt-1" || experiment.OverlayID != "overlay-1" {
		t.Fatalf("unexpected experiment %+v", experiment)
	}
}

func TestMemoryStoreSubmitCommandHarnessBindSessionPersistsBinding(t *testing.T) {
	store := newEmptyMemoryStore()
	now := time.Now().UTC()

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineHarness,
		AggregateID: "proposal|proposal|proposal-1",
		CommandKind: string(transition.CommandHarnessBindSession),
		CommandID:   "cmd-harness-bind",
		Actor:       "proposal",
		OccurredAt:  now,
		Payload: map[string]any{
			"role":                      "proposal",
			"scope_kind":                "proposal",
			"scope_id":                  "proposal-1",
			"hermes_session_id":         "hsess-1",
			"parent_session_id":         "hsess-parent",
			"memory_backend":            "honcho",
			"harness_profile_id":        "harness-profile-proposal",
			"effective_overlay_id":      "overlay-1",
			"effective_overlay_version": "overlay-v1",
			"last_used_at":              now,
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}

	binding, ok := store.harnessSessionBindings["proposal|proposal|proposal-1"]
	if !ok {
		t.Fatal("expected persisted harness session binding")
	}
	if binding.HermesSessionID != "hsess-1" || binding.HarnessProfileID != "harness-profile-proposal" {
		t.Fatalf("unexpected binding %+v", binding)
	}
}

func TestMemoryStoreSubmitCommandHarnessRecordExecutionPersistsExecution(t *testing.T) {
	store := newEmptyMemoryStore()
	now := time.Now().UTC()

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineHarness,
		AggregateID: "hexec-1",
		CommandKind: string(transition.CommandHarnessRecordExecution),
		CommandID:   "cmd-harness-exec",
		Actor:       "proposal",
		OccurredAt:  now,
		Payload: map[string]any{
			"operation_id":              "op-1",
			"trace_id":                  "trace-1",
			"proposal_id":               "proposal-1",
			"role":                      "proposal",
			"session_scope_kind":        "proposal",
			"session_scope_id":          "proposal-1",
			"hermes_session_id":         "hsess-1",
			"harness_profile_id":        "harness-profile-proposal",
			"effective_overlay_id":      "overlay-1",
			"effective_overlay_version": "overlay-v1",
			"memory_backend":            "honcho",
			"memory_reads": []harness.MemoryArtifact{{
				Kind:    "memory_read",
				Summary: "Loaded prior proposal memory.",
				Ref:     "mem-1",
			}},
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}

	executions := store.ListHarnessExecutions()
	if len(executions) != 1 {
		t.Fatalf("expected one persisted harness execution, got %d", len(executions))
	}
	if executions[0].ID != "hexec-1" || executions[0].OperationID != "op-1" {
		t.Fatalf("unexpected execution %+v", executions[0])
	}
}

func TestMemoryStoreSubmitCommandKnowledgeRecordDraftPersistsEntryAndEvidence(t *testing.T) {
	store := newEmptyMemoryStore()
	now := time.Now().UTC()

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineKnowledge,
		AggregateID: "knowledge-1",
		CommandKind: string(transition.CommandKnowledgeRecordDraft),
		CommandID:   "cmd-knowledge-draft",
		Actor:       "control-plane",
		OccurredAt:  now,
		Payload: map[string]any{
			"tier":        string(knowledge.TierWorking),
			"kind":        string(knowledge.KindFact),
			"scope_type":  string(knowledge.ScopeCase),
			"scope_id":    "case-1",
			"title":       "Recent workflow timeout",
			"summary":     "Timeout triggered while waiting on context action.",
			"body":        "The workflow hit its configured timeout before a tool result arrived.",
			"status":      string(knowledge.StatusDraft),
			"confidence":  0.73,
			"source_type": string(knowledge.SourceAgent),
			"evidence_links": []knowledge.EvidenceLink{{
				EvidenceType:     "trace",
				EvidenceID:       "trace-1",
				RelevanceSummary: "The failing trace captured the timeout.",
				EvidenceRef:      events.EvidenceRef{Kind: "trace", Ref: "trace-1", Summary: "Failing trace"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}

	entry, ok := store.GetKnowledgeEntry("knowledge-1")
	if !ok {
		t.Fatal("expected persisted knowledge entry")
	}
	if entry.Title != "Recent workflow timeout" || entry.Status != knowledge.StatusDraft {
		t.Fatalf("unexpected knowledge entry %+v", entry)
	}
	links := store.ListKnowledgeEvidenceLinks("knowledge-1")
	if len(links) != 1 || links[0].KnowledgeEntryID != "knowledge-1" {
		t.Fatalf("unexpected knowledge evidence %+v", links)
	}
}

func TestMemoryStoreSubmitCommandAttemptPROpenedPersistsPRAttempt(t *testing.T) {
	store := newEmptyMemoryStore()
	now := time.Now().UTC()
	store.proposals["proposal-1"] = review.Proposal{
		ID:               "proposal-1",
		Status:           review.ProposalValidationPending,
		TraceID:          "trace-1",
		TargetRef:        "rsi-agent-platform",
		CurrentAttemptID: "attempt-1",
		CreatedAt:        now,
		Version:          1,
	}
	store.changeAttempts["attempt-1"] = improvement.ChangeAttempt{
		ID:            "attempt-1",
		ProposalID:    "proposal-1",
		AttemptNumber: 1,
		State:         improvement.AttemptStateValidationRunning,
		BranchName:    "codex/proposal-1/attempt-01",
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	}

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAttempt,
		AggregateID: "attempt-1",
		CommandKind: string(transition.CommandAttemptPROpened),
		CommandID:   "cmd-attempt-pr-opened",
		Actor:       "improvement-plane",
		OccurredAt:  now,
		Payload: map[string]any{
			"pr_url":      "https://github.com/piplabs/rsi-agent-platform/pull/999",
			"head_sha":    "abc123",
			"repo":        "rsi-agent-platform",
			"branch_name": "codex/proposal-1/attempt-01",
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}

	attempts := store.ListPRAttempts()
	if len(attempts) != 1 {
		t.Fatalf("expected one pr attempt, got %d", len(attempts))
	}
	if attempts[0].AttemptID != "attempt-1" || attempts[0].PRURL == "" {
		t.Fatalf("unexpected pr attempt %+v", attempts[0])
	}
}

func TestMemoryStoreSubmitCommandProblemLineEvaluateTracePersistsEvalAndCandidate(t *testing.T) {
	store := NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	now := time.Now().UTC()

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: traceID,
		CommandKind: string(transition.CommandProblemLineEvaluateTrace),
		CommandID:   "cmd-problem-line-evaluate",
		Actor:       "improvement-plane",
		OccurredAt:  now,
		Payload: map[string]any{
			"trigger": "manual",
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}
	if strings.TrimSpace(receipt.ResultRef) == "" {
		t.Fatal("expected eval run result ref")
	}

	run, ok := findEvalRunByID(store.evalRuns, receipt.ResultRef)
	if !ok {
		t.Fatalf("expected eval run %s", receipt.ResultRef)
	}
	if run.TraceID != traceID || run.Trigger != "manual" {
		t.Fatalf("unexpected eval run %+v", run)
	}
	if judgments := store.evalJudgments[run.ID]; len(judgments) == 0 {
		t.Fatal("expected persisted eval judgments")
	}
	foundEffect := false
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind == transition.MachineProblemLine && effect.AggregateID == traceID && effect.EffectKind == transition.EffectInvokeRunner && effect.Status == transition.EffectQueued {
			if got := strings.TrimSpace(fmt.Sprint(effect.Payload["eval_run_id"])); got != run.ID {
				t.Fatalf("expected eval effect to carry run id %s, got %s", run.ID, got)
			}
			foundEffect = true
			break
		}
	}
	if !foundEffect {
		t.Fatal("expected queued problem-line invoke_runner effect")
	}
	foundCandidate := false
	for _, item := range store.candidates {
		if item.LatestTraceID == traceID && len(item.SourceEvalIDs) > 0 && item.SourceEvalIDs[len(item.SourceEvalIDs)-1] == run.ID {
			foundCandidate = true
			break
		}
	}
	if !foundCandidate {
		t.Fatal("expected trace evaluation to update a candidate")
	}
}

func TestMemoryStoreSubmitCommandProblemLinePromotePersistsPromotion(t *testing.T) {
	store := newEmptyMemoryStore()
	now := time.Now().UTC()
	store.settings.ActiveProposalCap = 2
	store.candidates["cand-1"] = improvement.Candidate{
		ID:                            "cand-1",
		CandidateKey:                  "cand-1",
		ConversationID:                "conv-1",
		CaseID:                        "case-1",
		OriginTraceID:                 "trace-1",
		LatestTraceID:                 "trace-1",
		EvidenceTraceIDs:              []string{"trace-1"},
		Subsystem:                     "control-plane",
		FailureMode:                   "failed_workflow",
		InterventionType:              "policy_or_runtime_fix",
		TargetLayer:                   harness.TargetLayerRepoChange,
		TargetKind:                    "repo",
		TargetRef:                     "rsi-agent-platform",
		Status:                        improvement.CandidateQueued,
		Hypothesis:                    "Fix the runtime persistence path.",
		ProposedScope:                 "control-plane + shared-store",
		RiskTier:                      improvement.RiskMedium,
		NewEvidenceSinceLastRejection: true,
		CreatedAt:                     now,
		UpdatedAt:                     now,
	}

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: "problem-lines",
		CommandKind: string(transition.CommandProblemLinePromote),
		CommandID:   "cmd-problem-line-promote",
		Actor:       "improvement-plane-cron",
		OccurredAt:  now,
		Payload: map[string]any{
			"requested_by": "improvement-plane-cron",
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}
	if len(store.proposals) != 1 {
		t.Fatalf("expected exactly one promoted proposal, got %d", len(store.proposals))
	}
	if store.candidates["cand-1"].Status != improvement.CandidatePromoted {
		t.Fatalf("expected candidate to be promoted, got %+v", store.candidates["cand-1"])
	}
	if lease, ok := store.cronLeases["improvement-plane-cron"]; !ok || lease.Holder != "improvement-plane-cron" {
		t.Fatalf("expected refreshed cron lease, got ok=%t lease=%+v", ok, lease)
	}
	foundPromotionEvent := false
	for _, item := range store.ListDomainEvents() {
		if item.CommandID == receipt.CommandID && item.EventKind == "problem_line_promoted" {
			foundPromotionEvent = true
			break
		}
	}
	if !foundPromotionEvent {
		t.Fatal("expected problem_line_promoted domain event")
	}
}

func TestMemoryStoreSubmitCommandProblemLineRecordOutcomePersistsOutcomeAndCase(t *testing.T) {
	store := NewMemoryStore()
	trace, ok := store.GetTrace(store.ListTraces()[0].TraceID)
	if !ok {
		t.Fatal("expected seeded trace")
	}
	now := time.Now().UTC()

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: trace.Summary.CaseID,
		CommandKind: string(transition.CommandProblemLineRecordOutcome),
		CommandID:   "cmd-problem-line-outcome",
		Actor:       "ui-operator",
		OccurredAt:  now,
		Payload: map[string]any{
			"conversation_id": trace.Summary.ConversationID,
			"case_id":         trace.Summary.CaseID,
			"trace_id":        trace.Summary.TraceID,
			"proposal_id":     store.ListProposals()[0].ID,
			"outcome_type":    string(outcome.TypeProposalEffectiveness),
			"verdict":         string(outcome.VerdictPositive),
			"score":           1.0,
			"summary":         "The proposal improved the system.",
			"source":          "operator",
			"recorded_by":     "ui-operator",
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if receipt.DecisionKind != transition.DecisionAdvance {
		t.Fatalf("expected advance receipt, got %+v", receipt)
	}
	if strings.TrimSpace(receipt.ResultRef) == "" {
		t.Fatal("expected outcome result ref")
	}
	record, ok := store.outcomes[receipt.ResultRef]
	if !ok {
		t.Fatalf("expected persisted outcome %s", receipt.ResultRef)
	}
	if record.CaseID != trace.Summary.CaseID || record.Verdict != outcome.VerdictPositive {
		t.Fatalf("unexpected outcome %+v", record)
	}
	caseRecord, ok := store.cases[trace.Summary.CaseID]
	if !ok {
		t.Fatalf("expected case %s", trace.Summary.CaseID)
	}
	if caseRecord.LatestOutcomeID != record.ID {
		t.Fatalf("expected case latest outcome %s, got %+v", record.ID, caseRecord)
	}
}
