package store

import (
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/operation"
	"github.com/piplabs/rsi-agent-platform/internal/policy"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
)

func TestMemoryStoreRatingAndReplay(t *testing.T) {
	store := NewMemoryStore()
	traces := store.ListTraces()
	if len(traces) == 0 {
		t.Fatal("expected seeded traces")
	}
	traceID := traces[0].TraceID

	if _, err := store.AddRating(traceID, review.HumanRating{
		Score:      4,
		Verdict:    "partial",
		Labels:     []string{"needs-human"},
		Notes:      "Useful investigation, incomplete mitigation.",
		ReviewerID: "alice",
	}); err != nil {
		t.Fatalf("AddRating() error = %v", err)
	}

	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatal("expected trace to exist")
	}
	if trace.Summary.LastVerdict != "partial" {
		t.Fatalf("expected last verdict to be updated, got %q", trace.Summary.LastVerdict)
	}

	item, err := store.ScheduleReplay(traceID, "alice")
	if err != nil {
		t.Fatalf("ScheduleReplay() error = %v", err)
	}
	if item.TraceID != traceID {
		t.Fatalf("unexpected trace id: %s", item.TraceID)
	}
}

func TestMemoryStoreSetThreadState(t *testing.T) {
	store := NewMemoryStore()

	item, err := store.SetThreadState("slack:CENG:171000001.000100", policy.ThreadStateMuted, "")
	if err != nil {
		t.Fatalf("SetThreadState() error = %v", err)
	}
	if !item.Muted {
		t.Fatal("expected muted flag to be set")
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

	result, err := store.RunProposalPromoter("test-promoter")
	if err != nil {
		t.Fatalf("RunProposalPromoter() error = %v", err)
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

	if _, err := store.RunProposalPromoter("test-worker"); err == nil {
		t.Fatal("expected promoter lease conflict")
	}
}

func TestRejectedProposalRequiresNewEvidence(t *testing.T) {
	store := NewMemoryStore()
	proposals := store.ListProposals()
	if len(proposals) == 0 {
		t.Fatal("expected seeded proposals")
	}
	proposal := proposals[0]

	if _, err := store.ReviewProposal(proposal.ID, review.ProposalReview{
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

	result, err := store.PromoteCandidates("alice", 2)
	if err != nil {
		t.Fatalf("PromoteCandidates() error = %v", err)
	}
	for _, promotedID := range result.PromotedIDs {
		if store.proposals[promotedID].CandidateKey == proposal.CandidateKey {
			t.Fatal("expected rejected candidate to stay blocked without new evidence")
		}
	}
}

func TestSettingsBackedProposalCap(t *testing.T) {
	store := NewMemoryStore()

	settings, err := store.UpdateSettings(improvement.Settings{ActiveProposalCap: 1})
	if err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}
	if settings.ActiveProposalCap != 1 {
		t.Fatalf("expected active proposal cap to be 1, got %d", settings.ActiveProposalCap)
	}

	slots := store.GetProposalSlots()
	if slots.Cap != 1 {
		t.Fatalf("expected slot cap to be 1, got %d", slots.Cap)
	}
}

func TestApproveProposalQueuesMaterializationWork(t *testing.T) {
	store := NewMemoryStore()
	proposals := store.ListProposals()
	if len(proposals) == 0 {
		t.Fatal("expected seeded proposals")
	}

	proposal, err := store.ReviewProposal(proposals[0].ID, review.ProposalReview{
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

	items := store.ListWorkItems()
	found := false
	for _, item := range items {
		if item.Queue == queue.ProposalQueue && item.ProposalID == proposal.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected proposal materialization work item to be queued")
	}
}

func TestRetryProposalRepoChangeRequeuesSandboxWork(t *testing.T) {
	store := NewMemoryStore()
	proposal := store.ListProposals()[0]

	if _, err := store.ReviewProposal(proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with repo-change work.",
		ReviewerID: "alice",
	}); err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	job, err := store.MaterializeApprovedProposal(proposal.ID, "alice")
	if err != nil {
		t.Fatalf("MaterializeApprovedProposal() error = %v", err)
	}
	if _, err := store.UpdateRepoChangeJobStatus(job.ID, string(review.ProposalFailedValidation)); err != nil {
		t.Fatalf("UpdateRepoChangeJobStatus() error = %v", err)
	}
	if _, err := store.UpdateProposalStatus(proposal.ID, review.ProposalFailedValidation); err != nil {
		t.Fatalf("UpdateProposalStatus() error = %v", err)
	}

	item, err := store.RetryProposalRepoChange(proposal.ID, "alice")
	if err != nil {
		t.Fatalf("RetryProposalRepoChange() error = %v", err)
	}
	if item.Queue != queue.SandboxQueue || item.Kind != "repo_change_job" {
		t.Fatalf("expected sandbox repo_change_job work item, got %+v", item)
	}
	if item.OperationID == "" {
		t.Fatalf("expected retry sandbox work item for job %s to be operation-backed", job.ID)
	}

	reloaded := store.ListRepoChangeJobs()[0]
	if reloaded.Status != string(review.ProposalRepoChangeQueued) {
		t.Fatalf("expected repo change job to return to repo_change_queued, got %s", reloaded.Status)
	}
	retryProposal := store.ListProposals()[0]
	if retryProposal.Status != review.ProposalRepoChangeQueued {
		t.Fatalf("expected proposal to return to repo_change_queued, got %s", retryProposal.Status)
	}

	again, err := store.RetryProposalRepoChange(proposal.ID, "alice")
	if err != nil {
		t.Fatalf("second RetryProposalRepoChange() error = %v", err)
	}
	if again.ID != item.ID {
		t.Fatalf("expected deduped retry work item %s, got %s", item.ID, again.ID)
	}
}

func TestReconcileProposalAttemptPhaseRequeuesCompletedWorkspaceOpenWhenWorkspaceNotReady(t *testing.T) {
	store := NewMemoryStore()
	proposal := seedProposalAttemptPhaseReconcileFixture(t, store, improvement.WorkspaceQueued)

	item, queued, err := store.ReconcileProposalAttemptPhase(proposal.ID, "tester")
	if err != nil {
		t.Fatalf("ReconcileProposalAttemptPhase() error = %v", err)
	}
	if !queued {
		t.Fatal("expected reconcile to requeue workspace_open")
	}
	if item.Kind != "workspace_open" || item.Status != queue.WorkQueued {
		t.Fatalf("expected queued workspace_open item, got %+v", item)
	}
	op, ok := store.GetOperation(item.OperationID)
	if !ok {
		t.Fatalf("expected operation %s", item.OperationID)
	}
	if op.OperationKind != "workspace_open" || op.Status != operation.StatusQueued {
		t.Fatalf("expected queued workspace_open operation, got %+v", op)
	}
}

func TestReconcileProposalAttemptPhaseQueuesImplementAttemptWhenWorkspaceReady(t *testing.T) {
	store := NewMemoryStore()
	proposal := seedProposalAttemptPhaseReconcileFixture(t, store, improvement.WorkspaceReady)

	item, queued, err := store.ReconcileProposalAttemptPhase(proposal.ID, "tester")
	if err != nil {
		t.Fatalf("ReconcileProposalAttemptPhase() error = %v", err)
	}
	if !queued {
		t.Fatal("expected reconcile to queue implement_attempt")
	}
	if item.Kind != "implement_attempt" || item.Status != queue.WorkQueued {
		t.Fatalf("expected queued implement_attempt item, got %+v", item)
	}
	op, ok := store.GetOperation(item.OperationID)
	if !ok {
		t.Fatalf("expected operation %s", item.OperationID)
	}
	if op.OperationKind != "implement_attempt" || op.Status != operation.StatusQueued {
		t.Fatalf("expected queued implement_attempt operation, got %+v", op)
	}
	ops := store.ListOperationsByScope(operation.ScopeAttempt, proposal.CurrentAttemptID)
	count := 0
	for _, candidate := range ops {
		if candidate.OperationKind == "implement_attempt" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one implement_attempt operation, got %d", count)
	}
}

func TestReconcileProposalAttemptPhaseRequeuesRunningWorkspaceOpenWhenWorkItemCanceled(t *testing.T) {
	store := NewMemoryStore()
	proposal := seedProposalAttemptPhaseReconcileFixture(t, store, improvement.WorkspaceQueued)

	op := store.operations["op-workspace-open-1"]
	now := time.Now().UTC()
	op.Status = operation.StatusRunning
	op.Holder = "worker-a"
	op.CompletedAt = nil
	op.UpdatedAt = now
	store.operations[op.ID] = op

	item := store.workItems["work-workspace-open-1"]
	item.Status = queue.WorkCanceled
	item.LeaseOwner = ""
	item.LeaseExpiresAt = nil
	item.LastError = "operation already terminal"
	item.CompletedAt = &now
	item.UpdatedAt = now
	store.workItems[item.ID] = item

	requeued, queued, err := store.ReconcileProposalAttemptPhase(proposal.ID, "tester")
	if err != nil {
		t.Fatalf("ReconcileProposalAttemptPhase() error = %v", err)
	}
	if !queued {
		t.Fatal("expected reconcile to requeue running workspace_open")
	}
	if requeued.ID != item.ID || requeued.Status != queue.WorkQueued {
		t.Fatalf("expected canceled workspace_open work item to be reopened, got %+v", requeued)
	}
	reopenedOp, ok := store.GetOperation(op.ID)
	if !ok {
		t.Fatalf("expected operation %s", op.ID)
	}
	if reopenedOp.Status != operation.StatusQueued || reopenedOp.Holder != "" {
		t.Fatalf("expected running workspace_open operation to be reopened, got %+v", reopenedOp)
	}
}

func TestEnqueueWorkItemAllowsRequeueAfterCompleted(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	first, err := store.EnqueueWorkItem(queue.WorkItem{
		Queue:  queue.ProposalQueue,
		Kind:   "approved_proposal",
		Status: queue.WorkQueued,
		Payload: map[string]any{
			"dedupe_key": "proposal-runner:proposal-test",
		},
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("EnqueueWorkItem() error = %v", err)
	}
	if _, err := store.CompleteWorkItem(first.ID); err != nil {
		t.Fatalf("CompleteWorkItem() error = %v", err)
	}
	second, err := store.EnqueueWorkItem(queue.WorkItem{
		Queue:  queue.ProposalQueue,
		Kind:   "approved_proposal",
		Status: queue.WorkQueued,
		Payload: map[string]any{
			"dedupe_key": "proposal-runner:proposal-test",
		},
		CreatedAt: now.Add(time.Second),
		UpdatedAt: now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("second EnqueueWorkItem() error = %v", err)
	}
	if second.ID == first.ID {
		t.Fatalf("expected completed work item %s not to block requeue", first.ID)
	}
}

func seedProposalAttemptPhaseReconcileFixture(t *testing.T, store *MemoryStore, workspaceStatus improvement.AttemptWorkspaceStatus) review.Proposal {
	t.Helper()
	base := store.ListProposals()[0]
	approved, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	now := time.Now().UTC()
	attempt := normalizeChangeAttempt(improvement.ChangeAttempt{
		ID:             "attempt-reconcile-1",
		ProposalID:     approved.ID,
		CandidateKey:   approved.CandidateKey,
		AttemptNumber:  1,
		TargetLayer:    approved.TargetLayer,
		TargetKind:     approved.TargetKind,
		TargetRef:      approved.TargetRef,
		Trigger:        improvement.AttemptTriggerProposalApproved,
		State:          improvement.AttemptStatePatchPlan,
		AttemptTraceID: firstNonEmpty(approved.OriginTraceID, approved.TraceID),
		BranchName:     "codex/reconcile-attempt-01",
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	workspace := normalizeAttemptWorkspace(improvement.AttemptWorkspace{
		ID:               "workspace-reconcile-1",
		AttemptID:        attempt.ID,
		ProposalID:       approved.ID,
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		Status:           workspaceStatus,
		AllowedPathGlobs: []string{"internal/**"},
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	job := improvement.RepoChangeJob{
		ID:               "job-reconcile-1",
		ProposalID:       approved.ID,
		AttemptID:        attempt.ID,
		ConversationID:   approved.ConversationID,
		CaseID:           approved.CaseID,
		OriginTraceID:    approved.TraceID,
		CandidateKey:     approved.CandidateKey,
		Status:           string(review.ProposalRepoChangeQueued),
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		AllowedPathGlobs: []string{"internal/**"},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	approved.CurrentAttemptID = attempt.ID
	approved.AttemptCount = 1
	approved.Status = review.ProposalRepoChangeQueued
	approved.ActiveSlotConsuming = true
	store.changeAttempts[attempt.ID] = attempt
	store.attemptWorkspaces[workspace.ID] = workspace
	store.repoChangeJobs[job.ID] = job
	store.proposals[approved.ID] = approved
	workspaceOpenOp := operation.Execution{
		ID:            "op-workspace-open-1",
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attempt.ID,
		OperationKind: "workspace_open",
		OperationKey:  "workspace_open",
		Status:        operation.StatusCompleted,
		Queue:         queue.ProposalQueue,
		RequestedBy:   "tester",
		TraceID:       approved.TraceID,
		ProposalID:    approved.ID,
		AttemptID:     attempt.ID,
		CreatedAt:     now,
		UpdatedAt:     now,
		CompletedAt:   &now,
	}
	store.operations[workspaceOpenOp.ID] = workspaceOpenOp
	store.workItems["work-workspace-open-1"] = queue.WorkItem{
		ID:          "work-workspace-open-1",
		OperationID: workspaceOpenOp.ID,
		Queue:       queue.ProposalQueue,
		Kind:        "workspace_open",
		Status:      queue.WorkCompleted,
		TraceID:     approved.TraceID,
		ProposalID:  approved.ID,
		Payload: map[string]any{
			"attempt_id":   attempt.ID,
			"workspace_id": workspace.ID,
		},
		RequestedBy: "tester",
		CreatedAt:   now,
		UpdatedAt:   now,
		CompletedAt: &now,
	}
	return approved
}

func TestEnqueueWorkItemReusesFailedOperationScopedItem(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	op, _, err := store.GetOrCreateOperation(operation.Execution{
		ID:            "op-requeue",
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       "attempt-1",
		OperationKind: "sandbox_launch",
		OperationKey:  "sandbox_launch",
		Status:        operation.StatusQueued,
		Queue:         queue.SandboxQueue,
		RequestedBy:   "tester",
		TraceID:       "trace-1",
		ProposalID:    "proposal-1",
		AttemptID:     "attempt-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		t.Fatalf("GetOrCreateOperation() error = %v", err)
	}
	first, err := store.EnqueueWorkItem(queue.WorkItem{
		ID:          "work-op-requeue",
		OperationID: op.ID,
		Queue:       queue.SandboxQueue,
		Kind:        "repo_change_job",
		Status:      queue.WorkQueued,
		ProposalID:  "proposal-1",
		TraceID:     "trace-1",
		Payload:     map[string]any{"phase": 1},
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("EnqueueWorkItem() error = %v", err)
	}
	if _, err := store.FailWorkItem(first.ID, "temporary failure"); err != nil {
		t.Fatalf("FailWorkItem() error = %v", err)
	}
	second, err := store.EnqueueWorkItem(queue.WorkItem{
		ID:          "work-op-requeue-new",
		OperationID: op.ID,
		Queue:       queue.SandboxQueue,
		Kind:        "repo_change_job",
		Status:      queue.WorkQueued,
		ProposalID:  "proposal-1",
		TraceID:     "trace-1",
		Payload:     map[string]any{"phase": 2},
		CreatedAt:   now.Add(time.Second),
		UpdatedAt:   now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("second EnqueueWorkItem() error = %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected operation-backed work item reuse, got %s want %s", second.ID, first.ID)
	}
	reloadedOp, ok := store.GetOperation(op.ID)
	if !ok {
		t.Fatal("expected operation to remain present")
	}
	if reloadedOp.Status != operation.StatusQueued {
		t.Fatalf("expected requeued operation status, got %s", reloadedOp.Status)
	}
}

func TestRescheduleWorkItemRequeuesOperation(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	op, _, err := store.GetOrCreateOperation(operation.Execution{
		ID:            "op-1",
		ScopeKind:     operation.ScopeProposal,
		ScopeID:       "proposal-1",
		OperationKind: "line_activate",
		OperationKey:  "attempt-01",
		Status:        operation.StatusQueued,
		Queue:         queue.ProposalQueue,
		RequestedBy:   "tester",
		TraceID:       "trace-1",
		ProposalID:    "proposal-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		t.Fatalf("GetOrCreateOperation() error = %v", err)
	}
	item, err := store.EnqueueWorkItem(queue.WorkItem{
		ID:          "work-1",
		OperationID: op.ID,
		Queue:       queue.ProposalQueue,
		Kind:        "approved_proposal",
		Status:      queue.WorkQueued,
		ProposalID:  "proposal-1",
		TraceID:     "trace-1",
		CreatedAt:   now,
		UpdatedAt:   now,
		RequestedBy: "tester",
	})
	if err != nil {
		t.Fatalf("EnqueueWorkItem() error = %v", err)
	}
	claimed, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.ProposalQueue}, "worker-a", 30*time.Second)
	if err != nil || !ok {
		t.Fatalf("ClaimNextWorkItem() ok=%t err=%v", ok, err)
	}
	if claimed.ID != item.ID {
		t.Fatalf("claimed wrong item: %+v", claimed)
	}
	if _, err := store.RescheduleWorkItem(item.ID, map[string]interface{}{"workspace_id": "workspace-1"}, "workspace initializing", time.Time{}); err != nil {
		t.Fatalf("RescheduleWorkItem() error = %v", err)
	}
	rescheduledOp, ok := store.GetOperation(op.ID)
	if !ok {
		t.Fatal("expected operation to remain present")
	}
	if rescheduledOp.Status != operation.StatusQueued {
		t.Fatalf("expected operation to be requeued, got %s", rescheduledOp.Status)
	}
	reclaimed, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.ProposalQueue}, "worker-b", 30*time.Second)
	if err != nil || !ok {
		t.Fatalf("second ClaimNextWorkItem() ok=%t err=%v", ok, err)
	}
	if reclaimed.ID != item.ID {
		t.Fatalf("expected reclaimed item %s, got %s", item.ID, reclaimed.ID)
	}
}

func TestEvaluateTraceRuntimeFailureCreatesRootCauseCandidate(t *testing.T) {
	store := NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatalf("expected trace %s", traceID)
	}

	description := `subsystem=shared-store failure_mode=action_result_primary_key_collision provider=github action_intent_id=action-002 work_item_id=work-003 kind=tool_read sqlstate=23505 constraint=action_result_pkey table=action_result error="duplicate key value violates unique constraint \"action_result_pkey\""`
	if _, err := store.ApplyTraceUpdate(traceID, TraceUpdate{
		Status:         statusPtr(events.StatusNeedsHuman),
		WorkflowStatus: "needs-human",
		WorkflowError:  description,
		Events: []events.TraceEvent{
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
				StartedAt:      time.Now().UTC(),
				Description:    description,
			},
		},
	}); err != nil {
		t.Fatalf("ApplyTraceUpdate() error = %v", err)
	}

	run, judgments, err := store.EvaluateTrace(traceID, "test")
	if err != nil {
		t.Fatalf("EvaluateTrace() error = %v", err)
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

func statusPtr(status events.Status) *events.Status {
	return &status
}
