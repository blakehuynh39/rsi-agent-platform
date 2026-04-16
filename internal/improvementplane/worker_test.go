package improvementplane

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/operation"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/runnerutil"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
	batchv1 "k8s.io/api/batch/v1"
)

type fakeLauncher struct {
	observation sandbox.JobObservation
	err         error
}

type operationWorkQueueStore interface {
	GetOrCreateOperation(operation.Execution) (operation.Execution, bool, error)
	EnqueueWorkItem(queue.WorkItem) (queue.WorkItem, error)
}

func enqueueImprovementOperationWork(store operationWorkQueueStore, op operation.Execution, item queue.WorkItem) (queue.WorkItem, error) {
	created, _, err := store.GetOrCreateOperation(op)
	if err != nil {
		return queue.WorkItem{}, err
	}
	item.OperationID = created.ID
	return store.EnqueueWorkItem(item)
}

func submitProposalStatusForTest(t *testing.T, store *storepkg.MemoryStore, proposalID string, status review.ProposalStatus, commandID string) {
	t.Helper()
	proposal, ok := findProposal(store.ListProposals(), proposalID)
	if !ok {
		t.Fatalf("expected proposal %s", proposalID)
	}
	step := 0
	for proposal.Status != status {
		step++
		kind, err := nextProposalStatusCommandForTest(proposal.Status, status)
		if err != nil {
			t.Fatalf("nextProposalStatusCommandForTest(%s -> %s) error = %v", proposal.Status, status, err)
		}
		receipt, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineProposalLine,
			AggregateID: proposalID,
			CommandKind: string(kind),
			CommandID:   fmt.Sprintf("%s-%02d", commandID, step),
			Actor:       "tester",
			OccurredAt:  time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("SubmitCommand(%s) error = %v", kind, err)
		}
		if receipt.DecisionKind == transition.DecisionReject {
			t.Fatalf("SubmitCommand(%s) rejected: %s", kind, receipt.Reason)
		}
		proposal, ok = findProposal(store.ListProposals(), proposalID)
		if !ok {
			t.Fatalf("expected updated proposal %s", proposalID)
		}
	}
}

func nextProposalStatusCommandForTest(current review.ProposalStatus, target review.ProposalStatus) (transition.ProposalLineCommandKind, error) {
	switch target {
	case review.ProposalRepoChangeQueued:
		switch current {
		case review.ProposalApproved, review.ProposalFailedValidation:
			return transition.CommandProposalMarkRepoChangeQueued, nil
		case review.ProposalRepoChangeQueued:
			return "", nil
		}
	case review.ProposalRepoChangeRunning:
		switch current {
		case review.ProposalApproved, review.ProposalFailedValidation:
			return transition.CommandProposalMarkRepoChangeQueued, nil
		case review.ProposalRepoChangeQueued:
			return transition.CommandProposalMarkRepoChangeRunning, nil
		case review.ProposalRepoChangeRunning:
			return "", nil
		}
	case review.ProposalValidationPending:
		switch current {
		case review.ProposalApproved, review.ProposalFailedValidation:
			return transition.CommandProposalMarkRepoChangeQueued, nil
		case review.ProposalRepoChangeQueued:
			return transition.CommandProposalMarkRepoChangeRunning, nil
		case review.ProposalRepoChangeRunning:
			return transition.CommandProposalMarkValidationPending, nil
		case review.ProposalValidationPending:
			return "", nil
		}
	case review.ProposalFailedValidation:
		switch current {
		case review.ProposalApproved:
			return transition.CommandProposalMarkRepoChangeQueued, nil
		case review.ProposalRepoChangeQueued, review.ProposalRepoChangeRunning, review.ProposalValidationPending:
			return transition.CommandProposalMarkFailedValidation, nil
		case review.ProposalFailedValidation:
			return "", nil
		}
	}
	return "", fmt.Errorf("unsupported proposal status transition %s -> %s", current, target)
}

func (f fakeLauncher) Launch(context.Context, sandbox.JobRequest) (sandbox.Session, *batchv1.Job, error) {
	return sandbox.Session{}, nil, errors.New("not implemented")
}

func (f fakeLauncher) GetJob(context.Context, string, string) (*batchv1.Job, error) {
	return nil, errors.New("not implemented")
}

func (f fakeLauncher) ObserveJob(context.Context, string, string) (sandbox.JobObservation, error) {
	return f.observation, f.err
}

func (f fakeLauncher) ResolvePod(context.Context, string, string) (string, error) {
	if strings.TrimSpace(f.observation.PodName) == "" {
		return "", errors.New("not implemented")
	}
	return f.observation.PodName, f.err
}

func (f fakeLauncher) Exec(context.Context, string, string, []string) (sandbox.ExecResult, error) {
	return sandbox.ExecResult{}, errors.New("not implemented")
}

type workspaceLaunchStub struct{}

func (workspaceLaunchStub) Launch(context.Context, sandbox.JobRequest) (sandbox.Session, *batchv1.Job, error) {
	now := time.Now().UTC()
	return sandbox.Session{
		ID:        "workspace-session-1",
		Namespace: "rsi-platform",
		PodName:   "workspace-pod-1",
		Status:    "running",
		CreatedAt: now,
		UpdatedAt: now,
	}, &batchv1.Job{}, nil
}

func (workspaceLaunchStub) GetJob(context.Context, string, string) (*batchv1.Job, error) {
	return nil, errors.New("not implemented")
}

func (workspaceLaunchStub) ObserveJob(context.Context, string, string) (sandbox.JobObservation, error) {
	return sandbox.JobObservation{}, errors.New("not implemented")
}

func (workspaceLaunchStub) ResolvePod(context.Context, string, string) (string, error) {
	return "workspace-pod-1", nil
}

func (workspaceLaunchStub) Exec(context.Context, string, string, []string) (sandbox.ExecResult, error) {
	return sandbox.ExecResult{}, errors.New("not implemented")
}

type fakeRunner struct {
	resp  clients.RunnerResponse
	err   error
	tasks []clients.RunnerTask
}

func (f *fakeRunner) Execute(task clients.RunnerTask) (clients.RunnerResponse, error) {
	f.tasks = append(f.tasks, task)
	return f.resp, f.err
}

type fakeToolClient struct {
	results map[string]storepkg.ToolResult
	errors  map[string]error
}

func (f fakeToolClient) Execute(name string, input map[string]any) (storepkg.ToolResult, error) {
	if err := f.errors[name]; err != nil {
		return storepkg.ToolResult{}, err
	}
	if result, ok := f.results[name]; ok {
		result.Input = input
		if result.Name == "" {
			result.Name = name
		}
		if result.ToolCallID == "" {
			result.ToolCallID = "tool-" + strings.ReplaceAll(name, ".", "-")
		}
		return result, nil
	}
	return storepkg.ToolResult{}, errors.New("unexpected tool call: " + name)
}

func ensureProposalAttemptForTest(cfg config.Config, store *storepkg.MemoryStore, proposal review.Proposal, sourceTrace events.Trace, item queue.WorkItem) (improvement.ChangeAttempt, events.Trace, error) {
	attempt, attemptTrace, err := ensureProposalAttempt(cfg, store, proposal, sourceTrace, item)
	if err == nil {
		return attempt, attemptTrace, nil
	}
	if !errors.Is(err, errProposalAttemptNotMaterialized) {
		return improvement.ChangeAttempt{}, events.Trace{}, err
	}
	attempt, traceReq := prepareProposalAttempt(cfg, proposal, sourceTrace, item)
	attemptTrace, _, err = store.CreateDerivedTrace(traceReq)
	if err != nil {
		return improvement.ChangeAttempt{}, events.Trace{}, err
	}
	attempt.AttemptTraceID = attemptTrace.Summary.TraceID
	attempt, err = store.UpsertChangeAttempt(attempt)
	if err != nil {
		return improvement.ChangeAttempt{}, events.Trace{}, err
	}
	return attempt, attemptTrace, nil
}

func assertQueuedAttemptBootstrapEffect(t *testing.T, store *storepkg.MemoryStore, attemptID string) transition.EffectExecution {
	t.Helper()
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind != transition.MachineAttempt || effect.AggregateID != attemptID || effect.Status != transition.EffectQueued {
			continue
		}
		switch effect.EffectKind {
		case transition.EffectOpenWorkspace, transition.EffectInvokeRunner:
			return effect
		}
	}
	t.Fatalf("expected queued attempt bootstrap effect for attempt %s", attemptID)
	return transition.EffectExecution{}
}

func requireQueuedAttemptEffect(t *testing.T, store *storepkg.MemoryStore, attemptID string, kind transition.EffectKind) transition.EffectExecution {
	t.Helper()
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind == transition.MachineAttempt && effect.AggregateID == attemptID && effect.Status == transition.EffectQueued && effect.EffectKind == kind {
			return effect
		}
	}
	t.Fatalf("expected queued %s effect for attempt %s", kind, attemptID)
	return transition.EffectExecution{}
}

func TestProcessWorkspaceValidationObservationEffectDefersOnNonTerminalObservation(t *testing.T) {
	store := storepkg.NewMemoryStore()
	base := store.ListProposals()[0]
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	sourceTrace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	attempt, _, err := ensureProposalAttemptForTest(config.Config{ServiceName: "improvement-plane"}, store, proposal, sourceTrace, queue.WorkItem{
		ProposalID: proposal.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	attempt.State = improvement.AttemptStatePatchGenerated
	attempt.ValidationPlan = "Run the sandbox validation flow."
	attempt.UpdatedAt = time.Now().UTC()
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		t.Fatalf("UpsertChangeAttempt() error = %v", err)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalRepoChangeQueued, "cmd-test-proposal-queued-observe-defer")
	if _, err := store.UpsertRepoChangeJob(improvement.RepoChangeJob{
		ID:               "job-watch-1",
		ProposalID:       proposal.ID,
		AttemptID:        attempt.ID,
		ConversationID:   proposal.ConversationID,
		CaseID:           proposal.CaseID,
		OriginTraceID:    sourceTrace.Summary.TraceID,
		CandidateKey:     proposal.CandidateKey,
		Status:           string(review.ProposalRepoChangeQueued),
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		SandboxNamespace: "rsi-platform",
		SandboxJobName:   "sandbox-watch-1",
		SandboxPodName:   "sandbox-watch-1-pod",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertRepoChangeJob() error = %v", err)
	}
	cfg := config.Config{
		ServiceName:         "improvement-plane",
		SandboxPollInterval: 2 * time.Second,
		SandboxNamespace:    "rsi-platform",
	}
	if err := submitAttemptCommand(store, attempt, transition.CommandValidationStarted, cfg.ServiceName, time.Now().UTC(), map[string]any{
		"operation_id":      "op-sandbox-watch-defer",
		"job_id":            "job-watch-1",
		"sandbox_namespace": "rsi-platform",
		"sandbox_job_name":  "sandbox-watch-1",
		"sandbox_pod_name":  "sandbox-watch-1-pod",
		"validation_ref":    "rsi-platform/sandbox-watch-1",
	}); err != nil {
		t.Fatalf("submitAttemptCommand(validation_started) error = %v", err)
	}
	effect, claimed, err := claimNextImprovementEffect(store, "tester", 30*time.Second, cfg.SandboxPollInterval)
	if err != nil {
		t.Fatalf("claimNextImprovementEffect() error = %v", err)
	}
	if !claimed {
		t.Fatal("expected claimed observe effect")
	}
	err = processImprovementEffect(cfg, store, nil, nil, nil, fakeLauncher{
		observation: sandbox.JobObservation{
			Namespace: "rsi-platform",
			JobName:   "sandbox-watch-1",
			PodName:   "sandbox-watch-1-pod",
		},
	}, nil, effect)
	if !errors.Is(err, errDeferredEffect) {
		t.Fatalf("processImprovementEffect() error = %v, want errDeferredEffect", err)
	}
	effects := store.ListEffectExecutions()
	if len(effects) == 0 {
		t.Fatal("expected persisted effects")
	}
	var observed transition.EffectExecution
	found := false
	for _, candidate := range effects {
		if candidate.ID == effect.ID {
			observed = candidate
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected effect %s after defer", effect.ID)
	}
	if observed.Status != transition.EffectRunning {
		t.Fatalf("expected observe effect running, got %s", observed.Status)
	}
	if observed.LeaseExpiresAt == nil {
		t.Fatal("expected observe effect lease expiry after defer")
	}
	if !observed.LeaseExpiresAt.After(time.Now().UTC()) {
		t.Fatalf("expected observe effect lease expiry in the future, got %v", observed.LeaseExpiresAt)
	}
}

func TestApplySandboxLaunchSuccessUsesFormalAttemptAndProposalCommands(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{ServiceName: "improvement-plane"}
	base := store.ListProposals()[0]
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	sourceTrace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, sourceTrace, queue.WorkItem{
		ProposalID: proposal.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	attempt.State = improvement.AttemptStatePatchGenerated
	attempt.ChangePlan = "Update the formal transition path."
	attempt.ValidationPlan = "Run the sandbox validation flow."
	attempt.UpdatedAt = time.Now().UTC()
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		t.Fatalf("UpsertChangeAttempt() error = %v", err)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalRepoChangeQueued, "cmd-test-proposal-queued")
	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s", proposal.ID)
	}
	repoJob, err := store.UpsertRepoChangeJob(improvement.RepoChangeJob{
		ID:             "job-sandbox-launch-1",
		ProposalID:     proposal.ID,
		AttemptID:      attempt.ID,
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		OriginTraceID:  attemptTrace.Summary.TraceID,
		CandidateKey:   proposal.CandidateKey,
		Status:         string(review.ProposalRepoChangeQueued),
		Repo:           "rsi-agent-platform",
		BaseRef:        "main",
		BranchName:     attempt.BranchName,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("UpsertRepoChangeJob() error = %v", err)
	}
	item := queue.WorkItem{
		ID:          "work-sandbox-launch-1",
		Queue:       queue.SandboxQueue,
		Kind:        "repo_change_job",
		Status:      queue.WorkQueued,
		TraceID:     sourceTrace.Summary.TraceID,
		ProposalID:  proposal.ID,
		OperationID: "op-sandbox-launch-1",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Payload: map[string]any{
			"attempt_id": attempt.ID,
			"job_id":     repoJob.ID,
		},
	}
	session := sandbox.Session{
		ID:        "sandbox-session-1",
		Namespace: "rsi-platform",
		PodName:   "sandbox-pod-1",
		Status:    "running",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := applySandboxLaunchSuccess(cfg, store, proposal, attempt, sourceTrace, repoJob, item, session); err != nil {
		t.Fatalf("applySandboxLaunchSuccess() error = %v", err)
	}

	attempt, ok = store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatalf("expected attempt %s", attempt.ID)
	}
	if attempt.State != improvement.AttemptStateValidationRunning {
		t.Fatalf("expected validation_running attempt state, got %s", attempt.State)
	}

	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s after sandbox launch", proposal.ID)
	}
	if proposal.Status != review.ProposalRepoChangeRunning {
		t.Fatalf("expected proposal repo_change_running, got %s", proposal.Status)
	}

	repoJob, ok = findRepoChangeJob(store.ListRepoChangeJobs(), repoJob.ID)
	if !ok {
		t.Fatalf("expected repo change job %s", repoJob.ID)
	}
	if repoJob.Status != string(review.ProposalRepoChangeRunning) {
		t.Fatalf("expected repo job repo_change_running, got %s", repoJob.Status)
	}

	attemptTrace, ok = store.GetTrace(attempt.AttemptTraceID)
	if !ok {
		t.Fatalf("expected attempt trace %s", attempt.AttemptTraceID)
	}
	foundSandboxStart := false
	foundSandboxArtifact := false
	for _, event := range attemptTrace.Events {
		if event.EventType == "sandbox.job.started" {
			foundSandboxStart = true
			break
		}
	}
	for _, artifact := range attemptTrace.Artifacts {
		if artifact.Kind == "sandbox_job" {
			foundSandboxArtifact = true
			break
		}
	}
	if !foundSandboxStart {
		t.Fatal("expected sandbox.job.started event on attempt trace")
	}
	if !foundSandboxArtifact {
		t.Fatal("expected sandbox_job artifact on attempt trace")
	}

	if receipt, ok := store.GetCommandReceipt("cmd-attempt:" + attempt.ID + ":validation_started:" + item.OperationID); !ok || receipt.MachineKind != transition.MachineAttempt {
		t.Fatalf("expected validation_started attempt command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
	if receipt, ok := store.GetCommandReceipt("cmd-proposal-repo-change-running:" + proposal.ID + ":" + attempt.ID); !ok || receipt.MachineKind != transition.MachineProposalLine {
		t.Fatalf("expected repo_change_running proposal command receipt, got ok=%t receipt=%+v", ok, receipt)
	}

	foundObserveEffect := false
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind == transition.MachineAttempt && effect.AggregateID == attempt.ID && effect.EffectKind == transition.EffectObserveWorkspaceValidation && effect.Status == transition.EffectQueued {
			foundObserveEffect = true
			break
		}
	}
	if !foundObserveEffect {
		t.Fatal("expected queued observe_workspace_validation effect")
	}
}

func TestProcessWorkspaceValidationObservationEffectUsesFormalAttemptAndProposalCommands(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		ServiceName:         "improvement-plane",
		SandboxPollInterval: 2 * time.Second,
	}
	base := store.ListProposals()[0]
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	sourceTrace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, sourceTrace, queue.WorkItem{
		ProposalID: proposal.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	attempt.State = improvement.AttemptStatePatchGenerated
	attempt.ValidationPlan = "Run the sandbox validation flow."
	attempt.UpdatedAt = time.Now().UTC()
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		t.Fatalf("UpsertChangeAttempt() error = %v", err)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalRepoChangeQueued, "cmd-test-proposal-queued-watch")
	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s", proposal.ID)
	}
	if _, err := store.UpsertRepoChangeJob(improvement.RepoChangeJob{
		ID:               "job-watch-success-1",
		ProposalID:       proposal.ID,
		AttemptID:        attempt.ID,
		ConversationID:   proposal.ConversationID,
		CaseID:           proposal.CaseID,
		OriginTraceID:    attemptTrace.Summary.TraceID,
		CandidateKey:     proposal.CandidateKey,
		Status:           string(review.ProposalRepoChangeRunning),
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		SandboxNamespace: "rsi-platform",
		SandboxJobName:   "sandbox-watch-success-1",
		SandboxPodName:   "sandbox-watch-success-1-pod",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertRepoChangeJob() error = %v", err)
	}
	sandboxAction, err := ensureImprovementActionIntent(store, improvementActionIntentBase(
		cfg.ServiceName,
		proposal,
		sourceTrace,
		attempt.ID,
		action.KindSandboxLaunch,
		"rsi-platform/job-watch-success-1",
		action.StatusQueued,
		"Launch the sandbox job to validate the approved repo change.",
		"sandbox:job-watch-success-1",
		map[string]any{
			"job_id":      "job-watch-success-1",
			"attempt_id":  attempt.ID,
			"repo":        "rsi-agent-platform",
			"branch_name": attempt.BranchName,
			"base_ref":    "main",
		},
		[]events.EvidenceRef{
			{Kind: "proposal", Ref: proposal.ID, Summary: proposal.CandidateKey},
			{Kind: "trace", Ref: sourceTrace.Summary.TraceID, Summary: sourceTrace.Summary.WorkflowKind},
		},
		time.Now().UTC(),
	))
	if err != nil {
		t.Fatalf("ensureImprovementActionIntent() error = %v", err)
	}
	operationID := "op-sandbox-launch-success"
	if _, err := submitImprovementActionCommand(store, sandboxAction.ID, transition.CommandActionStart, cfg.ServiceName, time.Now().UTC(), map[string]any{
		"operation_id": operationID,
		"attempt_id":   attempt.ID,
	}); err != nil {
		t.Fatalf("submitImprovementActionCommand(action_started) error = %v", err)
	}
	if err := applySandboxLaunchSuccess(cfg, store, proposal, attempt, sourceTrace, improvement.RepoChangeJob{
		ID:               "job-watch-success-1",
		ProposalID:       proposal.ID,
		AttemptID:        attempt.ID,
		ConversationID:   proposal.ConversationID,
		CaseID:           proposal.CaseID,
		OriginTraceID:    attemptTrace.Summary.TraceID,
		CandidateKey:     proposal.CandidateKey,
		Status:           string(review.ProposalRepoChangeQueued),
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		SandboxNamespace: "rsi-platform",
		SandboxJobName:   "sandbox-watch-success-1",
		SandboxPodName:   "sandbox-watch-success-1-pod",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}, queue.WorkItem{
		ID:          "work-sandbox-launch-success",
		Queue:       queue.SandboxQueue,
		Kind:        "repo_change_job",
		Status:      queue.WorkQueued,
		TraceID:     sourceTrace.Summary.TraceID,
		ProposalID:  proposal.ID,
		OperationID: operationID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Payload: map[string]any{
			"attempt_id": attempt.ID,
			"job_id":     "job-watch-success-1",
		},
	}, sandbox.Session{
		ID:        "sandbox-session-success",
		Namespace: "rsi-platform",
		PodName:   "sandbox-watch-success-1-pod",
		Status:    "running",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("applySandboxLaunchSuccess() error = %v", err)
	}
	effect, claimed, err := claimNextImprovementEffect(store, "tester", 30*time.Second, cfg.SandboxPollInterval)
	if err != nil || !claimed {
		t.Fatalf("claimNextImprovementEffect() claimed=%t err=%v", claimed, err)
	}
	if effect.EffectKind != transition.EffectObserveWorkspaceValidation {
		t.Fatalf("expected observe_workspace_validation effect, got %+v", effect)
	}
	err = processImprovementEffect(cfg, store, nil, nil, nil, fakeLauncher{
		observation: sandbox.JobObservation{
			Namespace:    "rsi-platform",
			JobName:      "sandbox-watch-success-1",
			PodName:      "sandbox-watch-success-1-pod",
			JobSucceeded: true,
			Logs:         "ok",
		},
	}, nil, effect)
	if err != nil {
		t.Fatalf("processImprovementEffect() error = %v", err)
	}

	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s", proposal.ID)
	}
	if proposal.Status != review.ProposalValidationPending {
		t.Fatalf("expected proposal validation_pending, got %s", proposal.Status)
	}
	attempt, ok = store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatalf("expected attempt %s", attempt.ID)
	}
	if attempt.State != improvement.AttemptStateValidationRunning {
		t.Fatalf("expected validation_running attempt state, got %s", attempt.State)
	}
	if attempt.ValidationSummary == "" {
		t.Fatal("expected validation summary to be recorded")
	}

	if receipt, ok := store.GetCommandReceipt("cmd-attempt:" + attempt.ID + ":validation_completed:" + operationID); !ok || receipt.MachineKind != transition.MachineAttempt {
		t.Fatalf("expected validation_completed attempt command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
	if receipt, ok := store.GetCommandReceipt("cmd-proposal-validation-pending:" + proposal.ID + ":" + attempt.ID); !ok || receipt.MachineKind != transition.MachineProposalLine {
		t.Fatalf("expected validation_pending proposal command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
	actionReceiptID := improvementActionCommandID(sandboxAction.ID, transition.CommandActionSucceed, operationID)
	if receipt, ok := store.GetCommandReceipt(actionReceiptID); !ok || receipt.MachineKind != transition.MachineAction {
		t.Fatalf("expected sandbox action success command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
	sandboxAction, ok = store.GetActionIntent(sandboxAction.ID)
	if !ok {
		t.Fatalf("expected sandbox action intent %s", sandboxAction.ID)
	}
	if sandboxAction.Status != action.StatusSucceeded {
		t.Fatalf("expected sandbox action succeeded, got %s", sandboxAction.Status)
	}
	if sandboxAction.ProposalID != proposal.ID {
		t.Fatalf("expected sandbox action proposal id %s, got %s", proposal.ID, sandboxAction.ProposalID)
	}
	results := store.ListActionResults(sandboxAction.ID)
	if len(results) != 1 || results[0].Status != action.StatusSucceeded {
		t.Fatalf("expected one succeeded sandbox action result, got %+v", results)
	}

	foundPROpenEffect := false
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind == transition.MachineAttempt && effect.AggregateID == attempt.ID && effect.EffectKind == transition.EffectOpenDraftPR && effect.Status == transition.EffectQueued {
			if got := strings.TrimSpace(stringValue(effect.Payload["repo"])); got != "rsi-agent-platform" {
				t.Fatalf("expected PR-open effect repo rsi-agent-platform, got %s", got)
			}
			if got := strings.TrimSpace(stringValue(effect.Payload["branch_name"])); got != attempt.BranchName {
				t.Fatalf("expected PR-open effect branch %s, got %s", attempt.BranchName, got)
			}
			foundPROpenEffect = true
			break
		}
	}
	if !foundPROpenEffect {
		t.Fatal("expected queued open_draft_pr effect")
	}

	attemptTrace, ok = store.GetTrace(attempt.AttemptTraceID)
	if !ok {
		t.Fatalf("expected attempt trace %s", attempt.AttemptTraceID)
	}
	foundPRQueued := false
	foundSandboxLogs := false
	for _, event := range attemptTrace.Events {
		if event.EventType == "github.pr.queued" {
			foundPRQueued = true
			break
		}
	}
	for _, artifact := range attemptTrace.Artifacts {
		if artifact.Kind == "sandbox_job_logs" || artifact.Kind == "sandbox_job_status" {
			foundSandboxLogs = true
			break
		}
	}
	if !foundPRQueued {
		t.Fatal("expected github.pr.queued event on attempt trace")
	}
	if !foundSandboxLogs {
		t.Fatal("expected sandbox observation artifacts on attempt trace")
	}
}

func TestProcessWorkspaceValidationObservationEffectProjectsSandboxFailureViaAttemptCommand(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		ServiceName:         "improvement-plane",
		SandboxPollInterval: 2 * time.Second,
		SandboxNamespace:    "rsi-platform",
	}
	base := store.ListProposals()[0]
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	sourceTrace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, sourceTrace, queue.WorkItem{
		ProposalID: proposal.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	attempt.State = improvement.AttemptStatePatchGenerated
	attempt.ValidationPlan = "Run the sandbox validation flow."
	attempt.UpdatedAt = time.Now().UTC()
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		t.Fatalf("UpsertChangeAttempt() error = %v", err)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalRepoChangeQueued, "cmd-test-proposal-queued-validate")
	if _, err := store.UpsertRepoChangeJob(improvement.RepoChangeJob{
		ID:               "job-watch-fail-1",
		ProposalID:       proposal.ID,
		AttemptID:        attempt.ID,
		ConversationID:   proposal.ConversationID,
		CaseID:           proposal.CaseID,
		OriginTraceID:    attemptTrace.Summary.TraceID,
		CandidateKey:     proposal.CandidateKey,
		Status:           string(review.ProposalRepoChangeRunning),
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		SandboxNamespace: "rsi-platform",
		SandboxJobName:   "sandbox-watch-fail-1",
		SandboxPodName:   "sandbox-watch-fail-1-pod",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertRepoChangeJob() error = %v", err)
	}
	sandboxAction, err := ensureImprovementActionIntent(store, improvementActionIntentBase(
		cfg.ServiceName,
		proposal,
		sourceTrace,
		attempt.ID,
		action.KindSandboxLaunch,
		"rsi-platform/job-watch-fail-1",
		action.StatusQueued,
		"Launch the sandbox job to validate the approved repo change.",
		"sandbox:job-watch-fail-1",
		map[string]any{
			"job_id":      "job-watch-fail-1",
			"attempt_id":  attempt.ID,
			"repo":        "rsi-agent-platform",
			"branch_name": attempt.BranchName,
			"base_ref":    "main",
		},
		[]events.EvidenceRef{
			{Kind: "proposal", Ref: proposal.ID, Summary: proposal.CandidateKey},
			{Kind: "trace", Ref: sourceTrace.Summary.TraceID, Summary: sourceTrace.Summary.WorkflowKind},
		},
		time.Now().UTC(),
	))
	if err != nil {
		t.Fatalf("ensureImprovementActionIntent() error = %v", err)
	}
	operationID := "op-sandbox-launch-fail"
	if _, err := submitImprovementActionCommand(store, sandboxAction.ID, transition.CommandActionStart, cfg.ServiceName, time.Now().UTC(), map[string]any{
		"operation_id": operationID,
		"attempt_id":   attempt.ID,
	}); err != nil {
		t.Fatalf("submitImprovementActionCommand(action_started) error = %v", err)
	}
	if err := applySandboxLaunchSuccess(cfg, store, proposal, attempt, sourceTrace, improvement.RepoChangeJob{
		ID:               "job-watch-fail-1",
		ProposalID:       proposal.ID,
		AttemptID:        attempt.ID,
		ConversationID:   proposal.ConversationID,
		CaseID:           proposal.CaseID,
		OriginTraceID:    attemptTrace.Summary.TraceID,
		CandidateKey:     proposal.CandidateKey,
		Status:           string(review.ProposalRepoChangeQueued),
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		SandboxNamespace: "rsi-platform",
		SandboxJobName:   "sandbox-watch-fail-1",
		SandboxPodName:   "sandbox-watch-fail-1-pod",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}, queue.WorkItem{
		ID:          "work-sandbox-launch-fail",
		Queue:       queue.SandboxQueue,
		Kind:        "repo_change_job",
		Status:      queue.WorkQueued,
		TraceID:     sourceTrace.Summary.TraceID,
		ProposalID:  proposal.ID,
		OperationID: operationID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Payload: map[string]any{
			"attempt_id": attempt.ID,
			"job_id":     "job-watch-fail-1",
		},
	}, sandbox.Session{
		ID:        "sandbox-session-fail",
		Namespace: "rsi-platform",
		PodName:   "sandbox-watch-fail-1-pod",
		Status:    "running",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("applySandboxLaunchSuccess() error = %v", err)
	}
	effect, claimed, err := claimNextImprovementEffect(store, "tester", 30*time.Second, cfg.SandboxPollInterval)
	if err != nil || !claimed {
		t.Fatalf("claimNextImprovementEffect() claimed=%t err=%v", claimed, err)
	}
	if effect.EffectKind != transition.EffectObserveWorkspaceValidation {
		t.Fatalf("expected observe_workspace_validation effect, got %+v", effect)
	}
	err = processImprovementEffect(cfg, store, nil, nil, nil, fakeLauncher{
		observation: sandbox.JobObservation{
			Namespace:         "rsi-platform",
			JobName:           "sandbox-watch-fail-1",
			PodName:           "sandbox-watch-fail-1-pod",
			JobFailed:         true,
			PodPhase:          "Failed",
			TerminationReason: "Error",
			Logs:              "boom",
		},
	}, nil, effect)
	if err != nil {
		t.Fatalf("processImprovementEffect() error = %v", err)
	}

	attempt, ok = store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatalf("expected attempt %s", attempt.ID)
	}
	if attempt.FailureClass != "sandbox_failure" {
		t.Fatalf("expected sandbox_failure class, got %s", attempt.FailureClass)
	}

	trace, ok := store.GetTrace(attempt.AttemptTraceID)
	if !ok {
		t.Fatalf("expected attempt trace %s", attempt.AttemptTraceID)
	}
	sandboxFailedEvents := 0
	foundAttemptFailed := false
	for _, event := range trace.Events {
		switch event.EventType {
		case "sandbox.job.failed":
			sandboxFailedEvents++
		case "change_attempt.failed":
			foundAttemptFailed = true
		}
	}
	if sandboxFailedEvents != 1 {
		t.Fatalf("expected exactly one sandbox.job.failed event on attempt trace, got %d", sandboxFailedEvents)
	}
	if !foundAttemptFailed {
		t.Fatal("expected change_attempt.failed event on attempt trace")
	}
	actionReceiptID := improvementActionCommandID(sandboxAction.ID, transition.CommandActionFail, operationID)
	if receipt, ok := store.GetCommandReceipt(actionReceiptID); !ok || receipt.MachineKind != transition.MachineAction {
		t.Fatalf("expected sandbox action failure command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
	sandboxAction, ok = store.GetActionIntent(sandboxAction.ID)
	if !ok {
		t.Fatalf("expected sandbox action intent %s", sandboxAction.ID)
	}
	if sandboxAction.Status != action.StatusFailed {
		t.Fatalf("expected sandbox action failed, got %s", sandboxAction.Status)
	}
	results := store.ListActionResults(sandboxAction.ID)
	if len(results) != 1 || results[0].Status != action.StatusFailed {
		t.Fatalf("expected one failed sandbox action result, got %+v", results)
	}
}

func TestProcessEvalItemRequeuesStalledApprovedProposal(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	proposals := store.ListProposals()
	if len(proposals) == 0 {
		t.Fatal("expected seeded proposals")
	}
	proposal := proposals[0]
	trace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	var err error
	proposal, err = store.ReviewProposal(proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "alice",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok || proposal.CurrentAttemptID == "" {
		t.Fatalf("expected current attempt after approval resume, got %+v", proposal)
	}
	attempt, ok := store.GetChangeAttempt(proposal.CurrentAttemptID)
	if !ok {
		t.Fatalf("expected attempt %s", proposal.CurrentAttemptID)
	}
	workspaceOpen, err := enqueueImprovementOperationWork(store, operation.Execution{
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attempt.ID,
		OperationKind: proposalOperationWorkspaceOpen,
		OperationKey:  proposalOperationWorkspaceOpen,
		Status:        operation.StatusQueued,
		Queue:         queue.ProposalQueue,
		RequestedBy:   "tester",
		TraceID:       attempt.AttemptTraceID,
		ProposalID:    proposal.ID,
		AttemptID:     attempt.ID,
	}, queue.WorkItem{
		ID:         "work-stalled-workspace-open",
		Queue:      queue.ProposalQueue,
		Kind:       proposalOperationWorkspaceOpen,
		Status:     queue.WorkQueued,
		TraceID:    attempt.AttemptTraceID,
		ProposalID: proposal.ID,
		Payload:    map[string]any{"attempt_id": attempt.ID},
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		t.Fatalf("enqueueImprovementOperationWork(workspace_open) error = %v", err)
	}
	now = time.Now().UTC()
	workspace, err := store.UpsertAttemptWorkspace(improvement.AttemptWorkspace{
		ID:               "workspace-stalled-1",
		AttemptID:        attempt.ID,
		ProposalID:       proposal.ID,
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		Status:           improvement.WorkspaceQueued,
		AllowedPathGlobs: []string{"internal/**"},
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		t.Fatalf("UpsertAttemptWorkspace() error = %v", err)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProposalLine,
		AggregateID: proposal.ID,
		CommandKind: string(transition.CommandProposalMarkRepoChangeRunning),
		CommandID:   "cmd-test-proposal-repo-change-running",
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SubmitCommand(proposal_mark_repo_change_running) error = %v", err)
	}
	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s after repo change running command", proposal.ID)
	}
	cfg := config.Config{ServiceName: "improvement-plane"}
	nextOp, nextItem := proposalAttemptPhaseWork(cfg, proposal, trace, attempt, proposalOperationImplementAttempt, map[string]any{
		"workspace_id": workspace.ID,
	})
	if err := store.AdvanceProposalAttemptPhase(storepkg.ProposalAttemptPhaseAdvance{
		ProposalID:    proposal.ID,
		WorkItemID:    workspaceOpen.ID,
		OperationID:   workspaceOpen.OperationID,
		Proposal:      &proposal,
		Workspace:     &workspace,
		NextOperation: &nextOp,
		NextWorkItem:  &nextItem,
	}); err != nil {
		t.Fatalf("AdvanceProposalAttemptPhase(workspace_open) error = %v", err)
	}
	evalCommandID := "cmd-problem-line:evaluate:" + trace.Summary.TraceID + ":manual-effect"
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: trace.Summary.TraceID,
		CommandKind: string(transition.CommandProblemLineEvaluateTrace),
		CommandID:   evalCommandID,
		Actor:       cfg.ServiceName,
		OccurredAt:  now.Add(time.Second),
		Payload: map[string]any{
			"trigger": "evaluate_trace",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(problem_line_evaluate_trace) error = %v", err)
	}
	effect, claimed, err := claimNextImprovementEffect(store, "tester", 30*time.Second, 30*time.Second)
	if err != nil {
		t.Fatalf("claimNextImprovementEffect() error = %v", err)
	}
	if !claimed {
		t.Fatal("expected queued problem-line eval effect")
	}
	if err := processImprovementEffect(cfg, store, nil, nil, nil, nil, nil, effect); err != nil {
		t.Fatalf("processImprovementEffect() error = %v", err)
	}
	if receipt, ok := store.GetCommandReceipt(evalCommandID); !ok || receipt.MachineKind != transition.MachineProblemLine {
		t.Fatalf("expected problem-line eval command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
	if receipt, ok := store.GetCommandReceipt("cmd-problem-line:trace:" + trace.Summary.TraceID + ":" + effect.ID); !ok || receipt.MachineKind != transition.MachineProblemLine {
		t.Fatalf("expected problem-line trace projection receipt, got ok=%t receipt=%+v", ok, receipt)
	}
	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok || strings.TrimSpace(proposal.CurrentAttemptID) == "" {
		t.Fatalf("expected proposal %s with current attempt after eval", proposal.ID)
	}
	assertQueuedAttemptBootstrapEffect(t, store, proposal.CurrentAttemptID)
}

func TestRecordAttemptFailureUsesFormalAttemptAndProposalCommands(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{ServiceName: "improvement-plane"}
	base := store.ListProposals()[0]
	trace, ok := store.GetTrace(base.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", base.TraceID)
	}
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok || proposal.CurrentAttemptID == "" {
		t.Fatalf("expected current attempt after approval resume, got %+v", proposal)
	}
	attempt, ok := store.GetChangeAttempt(proposal.CurrentAttemptID)
	if !ok {
		t.Fatalf("expected attempt %s", proposal.CurrentAttemptID)
	}
	attempt.State = improvement.AttemptStateValidationRunning
	attempt.UpdatedAt = time.Now().UTC()
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		t.Fatalf("UpsertChangeAttempt() error = %v", err)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalRepoChangeRunning, "cmd-test-proposal-validation-running-1")
	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected updated proposal %s", proposal.ID)
	}

	if err := recordAttemptFailure(cfg, store, proposal, attempt, trace, "sandbox_failure", "Sandbox validation failed.", false, improvement.AttemptTriggerSandboxFailed); err != nil {
		t.Fatalf("recordAttemptFailure() error = %v", err)
	}

	attempt, ok = store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatalf("expected attempt %s after failure", attempt.ID)
	}
	if attempt.State != improvement.AttemptStateSandboxFailed {
		t.Fatalf("expected sandbox_failed attempt state, got %s", attempt.State)
	}
	if attempt.RetryDecision != "auto_retry" {
		t.Fatalf("expected auto_retry decision, got %s", attempt.RetryDecision)
	}
	if attempt.FailureClass != "sandbox_failure" {
		t.Fatalf("expected sandbox_failure class, got %s", attempt.FailureClass)
	}

	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s after failure", proposal.ID)
	}
	if proposal.Status != review.ProposalApproved {
		t.Fatalf("expected proposal to remain approved for retryable failure, got %s", proposal.Status)
	}

	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok || strings.TrimSpace(proposal.CurrentAttemptID) == "" {
		t.Fatalf("expected proposal %s with current attempt after retryable failure", proposal.ID)
	}
	assertQueuedAttemptBootstrapEffect(t, store, proposal.CurrentAttemptID)

	traceID := firstNonEmpty(attempt.AttemptTraceID, trace.Summary.TraceID)
	trace, ok = store.GetTrace(traceID)
	if !ok {
		t.Fatalf("expected trace %s", traceID)
	}
	foundFailureEvent := false
	for _, event := range trace.Events {
		if event.EventType == "change_attempt.failed" {
			foundFailureEvent = true
		}
	}
	if !foundFailureEvent {
		t.Fatal("expected change_attempt.failed event to be projected from attempt command")
	}

	receipt, ok := store.GetCommandReceipt("cmd-proposal-attempt-failure:" + attempt.ID + ":sandbox_failure")
	if !ok || receipt.MachineKind != transition.MachineProposalLine {
		t.Fatalf("expected proposal failure command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
}

func TestProcessDraftPROpenUsesFormalAttemptAndProposalCommands(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{ServiceName: "improvement-plane"}
	base := store.ListProposals()[0]
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	proposal.TargetRef = "prod"
	proposal.TargetKind = "role"
	sourceTrace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, sourceTrace, queue.WorkItem{
		ProposalID: proposal.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	attempt.State = improvement.AttemptStateValidationRunning
	attempt.UpdatedAt = time.Now().UTC()
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		t.Fatalf("UpsertChangeAttempt() error = %v", err)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalValidationPending, "cmd-test-proposal-validation-pending-2")
	if _, err := store.UpsertRepoChangeJob(improvement.RepoChangeJob{
		ID:             "job-pr-open-1",
		ProposalID:     proposal.ID,
		AttemptID:      attempt.ID,
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		OriginTraceID:  attemptTrace.Summary.TraceID,
		CandidateKey:   proposal.CandidateKey,
		Status:         string(review.ProposalValidationPending),
		Repo:           "rsi-agent-platform",
		BaseRef:        "main",
		BranchName:     attempt.BranchName,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertRepoChangeJob() error = %v", err)
	}
	item, err := enqueueImprovementOperationWork(store, operation.Execution{
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attempt.ID,
		OperationKind: "pr_open",
		OperationKey:  "pr_open",
		Status:        operation.StatusQueued,
		Queue:         queue.ImprovementActionQueue,
		RequestedBy:   cfg.ServiceName,
		TraceID:       attemptTrace.Summary.TraceID,
		ProposalID:    proposal.ID,
		AttemptID:     attempt.ID,
	}, queue.WorkItem{
		ID:         "work-pr-open-1",
		Queue:      queue.ImprovementActionQueue,
		Kind:       "draft_pr_open",
		Status:     queue.WorkQueued,
		TraceID:    attemptTrace.Summary.TraceID,
		ProposalID: proposal.ID,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		Payload: map[string]any{
			"attempt_id":  attempt.ID,
			"job_id":      "job-pr-open-1",
			"repo":        "rsi-agent-platform",
			"branch_name": attempt.BranchName,
			"base_ref":    "main",
			"title":       "Fix formal transition gap",
			"body":        "Draft PR body",
		},
	})
	if err != nil {
		t.Fatalf("enqueueImprovementOperationWork() error = %v", err)
	}
	toolClient := fakeToolClient{
		results: map[string]storepkg.ToolResult{
			"github.create_pr": {
				Available:   true,
				Status:      "ok",
				Provider:    "github",
				ProviderRef: "pull/123",
				Output: map[string]any{
					"pr_url": "https://github.com/piplabs/rsi-agent-platform/pull/123",
					"response": map[string]any{
						"head": map[string]any{
							"sha": "abc123",
						},
					},
				},
			},
		},
	}

	if err := processDraftPROpen(cfg, store, toolClient, item); err != nil {
		t.Fatalf("processDraftPROpen() error = %v", err)
	}

	attempt, ok = store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatalf("expected attempt %s", attempt.ID)
	}
	if attempt.State != improvement.AttemptStateCIObserving {
		t.Fatalf("expected attempt state ci_observing, got %s", attempt.State)
	}
	if attempt.PRURL != "https://github.com/piplabs/rsi-agent-platform/pull/123" {
		t.Fatalf("unexpected attempt pr url %q", attempt.PRURL)
	}
	if attempt.HeadSHA != "abc123" {
		t.Fatalf("unexpected attempt head sha %q", attempt.HeadSHA)
	}

	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s", proposal.ID)
	}
	if proposal.Status != review.ProposalPROpen {
		t.Fatalf("expected proposal pr_open, got %s", proposal.Status)
	}

	repoJob, ok := findRepoChangeJob(store.ListRepoChangeJobs(), "job-pr-open-1")
	if !ok {
		t.Fatal("expected repo change job after PR open")
	}
	if repoJob.Status != string(review.ProposalPROpen) {
		t.Fatalf("expected repo job pr_open, got %s", repoJob.Status)
	}

	foundPRAttempt := false
	for _, prAttempt := range store.ListPRAttempts() {
		if prAttempt.AttemptID == attempt.ID && prAttempt.PRURL == "https://github.com/piplabs/rsi-agent-platform/pull/123" {
			foundPRAttempt = true
			break
		}
	}
	if !foundPRAttempt {
		t.Fatal("expected persisted pr attempt")
	}
	actionID := improvementActionIntentIDFromIdempotencyKey(fmt.Sprintf("pr:%s:%s", attempt.ID, attempt.BranchName))
	actionIntent, ok := store.GetActionIntent(actionID)
	if !ok {
		t.Fatalf("expected draft pr action intent %s", actionID)
	}
	if actionIntent.Status != action.StatusSucceeded {
		t.Fatalf("expected draft pr action succeeded, got %s", actionIntent.Status)
	}
	if actionIntent.ProposalID != proposal.ID {
		t.Fatalf("expected draft pr action proposal id %s, got %s", proposal.ID, actionIntent.ProposalID)
	}
	for _, kind := range []transition.ActionExecutionCommandKind{
		transition.CommandActionQueue,
		transition.CommandActionStart,
		transition.CommandActionSucceed,
	} {
		receiptID := improvementActionCommandID(actionID, kind, "")
		if kind != transition.CommandActionQueue {
			receiptID = improvementActionCommandID(actionID, kind, item.OperationID)
		}
		if receipt, ok := store.GetCommandReceipt(receiptID); !ok || receipt.MachineKind != transition.MachineAction {
			t.Fatalf("expected draft pr action receipt for %s, got ok=%t receipt=%+v", kind, ok, receipt)
		}
	}

	trace, ok := store.GetTrace(attempt.AttemptTraceID)
	if !ok {
		t.Fatalf("expected attempt trace %s", attempt.AttemptTraceID)
	}
	foundStartedEvent := false
	foundCompletedEvent := false
	for _, event := range trace.Events {
		if event.EventType == "github.pr.started" {
			foundStartedEvent = true
		}
		if event.EventType == "github.pr.completed" {
			foundCompletedEvent = true
			break
		}
	}
	if !foundStartedEvent {
		t.Fatal("expected github.pr.started event on attempt trace")
	}
	if !foundCompletedEvent {
		t.Fatal("expected github.pr.completed event on attempt trace")
	}

	attemptReceiptID := "cmd-attempt:" + attempt.ID + ":" + string(transition.CommandAttemptPROpened) + ":" + item.OperationID
	receipt, ok := store.GetCommandReceipt(attemptReceiptID)
	if !ok || receipt.MachineKind != transition.MachineAttempt {
		t.Fatalf("expected attempt pr-open command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
	proposalReceiptID := "cmd-proposal-pr-open:" + proposal.ID + ":" + attempt.ID
	receipt, ok = store.GetCommandReceipt(proposalReceiptID)
	if !ok || receipt.MachineKind != transition.MachineProposalLine {
		t.Fatalf("expected proposal pr-open command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
}

func TestProcessDraftPROpenFailureUsesFormalAttemptFailureCommands(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{ServiceName: "improvement-plane"}
	base := store.ListProposals()[0]
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	proposal.TargetRef = "prod"
	proposal.TargetKind = "role"
	sourceTrace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, sourceTrace, queue.WorkItem{
		ProposalID: proposal.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	attempt.State = improvement.AttemptStateValidationRunning
	attempt.UpdatedAt = time.Now().UTC()
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		t.Fatalf("UpsertChangeAttempt() error = %v", err)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalValidationPending, "cmd-test-proposal-validation-pending-3")
	item, err := enqueueImprovementOperationWork(store, operation.Execution{
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attempt.ID,
		OperationKind: "pr_open",
		OperationKey:  "pr_open",
		Status:        operation.StatusQueued,
		Queue:         queue.ImprovementActionQueue,
		RequestedBy:   cfg.ServiceName,
		TraceID:       attemptTrace.Summary.TraceID,
		ProposalID:    proposal.ID,
		AttemptID:     attempt.ID,
	}, queue.WorkItem{
		ID:         "work-pr-open-fail-1",
		Queue:      queue.ImprovementActionQueue,
		Kind:       "draft_pr_open",
		Status:     queue.WorkQueued,
		TraceID:    attemptTrace.Summary.TraceID,
		ProposalID: proposal.ID,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		Payload: map[string]any{
			"attempt_id":  attempt.ID,
			"repo":        "rsi-agent-platform",
			"branch_name": attempt.BranchName,
			"base_ref":    "main",
			"title":       "Fix formal transition gap",
			"body":        "Draft PR body",
		},
	})
	if err != nil {
		t.Fatalf("enqueueImprovementOperationWork() error = %v", err)
	}
	toolClient := fakeToolClient{
		results: map[string]storepkg.ToolResult{
			"github.create_pr": {
				Status:   "blocked",
				Summary:  "Draft PR open blocked.",
				Provider: "github",
			},
		},
	}

	if err := processDraftPROpen(cfg, store, toolClient, item); err != nil {
		t.Fatalf("processDraftPROpen() error = %v", err)
	}

	attempt, ok = store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatalf("expected attempt %s", attempt.ID)
	}
	if attempt.State != improvement.AttemptStateNeedsReview {
		t.Fatalf("expected attempt state needs_review, got %s", attempt.State)
	}
	if attempt.FailureClass != "stale_branch" {
		t.Fatalf("expected stale_branch failure class, got %s", attempt.FailureClass)
	}

	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s", proposal.ID)
	}
	if proposal.Status != review.ProposalApproved {
		t.Fatalf("expected proposal approved, got %s", proposal.Status)
	}

	trace, ok := store.GetTrace(attempt.AttemptTraceID)
	if !ok {
		t.Fatalf("expected attempt trace %s", attempt.AttemptTraceID)
	}
	foundBlocked := false
	foundAttemptFailed := false
	for _, event := range trace.Events {
		switch event.EventType {
		case "github.pr.blocked":
			foundBlocked = true
		case "change_attempt.failed":
			foundAttemptFailed = true
		}
	}
	if !foundBlocked {
		t.Fatal("expected github.pr.blocked event on attempt trace")
	}
	if !foundAttemptFailed {
		t.Fatal("expected change_attempt.failed event on attempt trace")
	}
	if trace.Summary.Status != events.StatusFailed {
		t.Fatalf("expected attempt trace status failed, got %s", trace.Summary.Status)
	}

	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok || strings.TrimSpace(proposal.CurrentAttemptID) == "" {
		t.Fatalf("expected proposal %s with current attempt after pr_open failure", proposal.ID)
	}
	assertQueuedAttemptBootstrapEffect(t, store, proposal.CurrentAttemptID)
}

func TestProcessHarnessOverlayProposalUsesFormalAttemptAndProposalCommands(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{ServiceName: "improvement-plane"}
	base := store.ListProposals()[0]
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	proposal.TargetRef = "prod"
	proposal.TargetKind = "role"
	sourceTrace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, sourceTrace, queue.WorkItem{
		ProposalID: proposal.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	attempt.State = improvement.AttemptStateOverlayPlan
	attempt.UpdatedAt = time.Now().UTC()
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		t.Fatalf("UpsertChangeAttempt() error = %v", err)
	}
	if _, err := enqueueImprovementOperationWork(store, operation.Execution{
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attempt.ID,
		OperationKind: proposalOperationImplementAttempt,
		OperationKey:  proposalOperationImplementAttempt,
		Status:        operation.StatusQueued,
		Queue:         queue.ProposalQueue,
		RequestedBy:   cfg.ServiceName,
		TraceID:       attemptTrace.Summary.TraceID,
		ProposalID:    proposal.ID,
		AttemptID:     attempt.ID,
	}, queue.WorkItem{
		ID:         "work-overlay-1",
		Queue:      queue.ProposalQueue,
		Kind:       proposalOperationImplementAttempt,
		Status:     queue.WorkQueued,
		TraceID:    attemptTrace.Summary.TraceID,
		ProposalID: proposal.ID,
		Payload: map[string]any{
			"attempt_id": attempt.ID,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("enqueueImprovementOperationWork() error = %v", err)
	}

	runnerOutput := runnerutil.StructuredOutput{
		ChangePlan:     "Activate a production-safe harness overlay.",
		ValidationPlan: "Observe overlay activation.",
		FinalAnswer:    "Use a bounded overlay to steer the runtime.",
		ProposedActions: []runnerutil.ProposedAction{
			{
				Kind: string(action.KindHarnessOverlay),
				RequestPayload: map[string]any{
					"version":               "overlay-v1",
					"prompt_fragments":      []string{"Prioritize direct RSI evidence."},
					"few_shot_snippets":     []string{"Example snippet"},
					"tool_preference_order": []string{"rsi.runtime_health", "rsi.workflow_context"},
					"retrieval_bias":        "recent",
					"reasoning_verbosity":   "verbose",
					"memory_read_enabled":   true,
					"memory_write_enabled":  false,
				},
			},
		},
	}
	if err := processHarnessOverlayProposal(cfg, store, attemptTrace, proposal, attempt, clients.RunnerResponse{OK: true}, runnerOutput, time.Now().UTC()); err != nil {
		t.Fatalf("processHarnessOverlayProposal() error = %v", err)
	}

	attempt, ok = store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatalf("expected attempt %s", attempt.ID)
	}
	if attempt.State != improvement.AttemptStateOverlayActive {
		t.Fatalf("expected overlay_active attempt state, got %s", attempt.State)
	}
	if attempt.ChangePlan != runnerOutput.ChangePlan {
		t.Fatalf("expected attempt change plan %q, got %q", runnerOutput.ChangePlan, attempt.ChangePlan)
	}
	if attempt.ValidationPlan != runnerOutput.ValidationPlan {
		t.Fatalf("expected attempt validation plan %q, got %q", runnerOutput.ValidationPlan, attempt.ValidationPlan)
	}
	if overlayID := stringValue(attempt.OverlayPayload["overlay_id"]); overlayID == "" {
		t.Fatal("expected attempt overlay payload to include overlay_id")
	}

	proposal, ok = findProposal(store.ListProposals(), proposal.ID)
	if !ok {
		t.Fatalf("expected proposal %s", proposal.ID)
	}
	if proposal.Status != review.ProposalMerged {
		t.Fatalf("expected proposal merged, got %s", proposal.Status)
	}

	foundOverlay := false
	for _, overlay := range store.ListHarnessOverlays() {
		if overlay.ProposalID == proposal.ID && overlay.Version == "overlay-v1" {
			foundOverlay = true
			break
		}
	}
	if !foundOverlay {
		t.Fatal("expected persisted harness overlay")
	}

	foundExperiment := false
	for _, experiment := range store.ListHarnessExperiments() {
		if experiment.ProposalID == proposal.ID && experiment.AttemptID == attempt.ID {
			foundExperiment = true
			break
		}
	}
	if !foundExperiment {
		t.Fatal("expected persisted harness experiment")
	}

	trace, ok := store.GetTrace(attempt.AttemptTraceID)
	if !ok {
		t.Fatalf("expected attempt trace %s", attempt.AttemptTraceID)
	}
	foundRunnerCompleted := false
	foundOverlayActivated := false
	for _, event := range trace.Events {
		switch event.EventType {
		case "runner.completed":
			foundRunnerCompleted = true
		case "harness.overlay.activated":
			foundOverlayActivated = true
		}
	}
	if !foundRunnerCompleted {
		t.Fatal("expected runner.completed event on attempt trace")
	}
	if !foundOverlayActivated {
		t.Fatal("expected harness.overlay.activated event on attempt trace")
	}

	currentOp, ok := latestActiveAttemptOperation(store, attempt.ID)
	if !ok {
		t.Fatalf("expected active operation for attempt %s", attempt.ID)
	}
	actionID := improvementActionIntentIDFromIdempotencyKey(fmt.Sprintf("harness-overlay:%s", proposal.ID))
	actionIntent, ok := store.GetActionIntent(actionID)
	if !ok {
		t.Fatalf("expected harness overlay action intent %s", actionID)
	}
	if actionIntent.Status != action.StatusSucceeded {
		t.Fatalf("expected harness overlay action succeeded, got %s", actionIntent.Status)
	}
	if actionIntent.ProposalID != proposal.ID {
		t.Fatalf("expected harness overlay action proposal id %s, got %s", proposal.ID, actionIntent.ProposalID)
	}
	for _, kind := range []transition.ActionExecutionCommandKind{
		transition.CommandActionQueue,
		transition.CommandActionStart,
		transition.CommandActionSucceed,
	} {
		receiptID := improvementActionCommandID(actionID, kind, "")
		if kind != transition.CommandActionQueue {
			receiptID = improvementActionCommandID(actionID, kind, currentOp.ID)
		}
		if receipt, ok := store.GetCommandReceipt(receiptID); !ok || receipt.MachineKind != transition.MachineAction {
			t.Fatalf("expected harness overlay action receipt for %s, got ok=%t receipt=%+v", kind, ok, receipt)
		}
	}
	attemptReceiptID := "cmd-attempt:" + attempt.ID + ":" + string(transition.CommandOverlayActivated) + ":" + currentOp.ID
	receipt, ok := store.GetCommandReceipt(attemptReceiptID)
	if !ok || receipt.MachineKind != transition.MachineAttempt {
		t.Fatalf("expected overlay activation attempt command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
	if receipt, ok := store.GetCommandReceipt("cmd-harness-overlay:" + proposal.ID + ":" + attempt.ID); !ok || receipt.MachineKind != transition.MachineHarness {
		t.Fatalf("expected harness overlay command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
	proposalReceiptID := "cmd-proposal-overlay-merged:" + proposal.ID + ":" + attempt.ID
	receipt, ok = store.GetCommandReceipt(proposalReceiptID)
	if !ok || receipt.MachineKind != transition.MachineProposalLine {
		t.Fatalf("expected overlay merge proposal command receipt, got ok=%t receipt=%+v", ok, receipt)
	}
}

func TestReviewProposalQueuesAttemptBootstrapEffect(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]
	approved, err := store.ReviewProposal(proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with governed remediation.",
		ReviewerID: "alice",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	approved, ok := findProposal(store.ListProposals(), approved.ID)
	if !ok || approved.CurrentAttemptID == "" {
		t.Fatalf("expected current attempt after approval resume, got %+v", approved)
	}
	assertQueuedAttemptBootstrapEffect(t, store, approved.CurrentAttemptID)
}

func TestProcessImplementAttemptEffectRequiresWorkspaceMutation(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		ServiceName:               "improvement-plane",
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		DefaultReasoningVerbosity: "verbose",
	}
	proposal := store.ListProposals()[0]
	approved, err := store.ReviewProposal(proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with governed remediation.",
		ReviewerID: "alice",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	sourceTrace, ok := store.GetTrace(approved.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", approved.TraceID)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, approved, sourceTrace, queue.WorkItem{
		ProposalID: approved.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	if _, err := store.UpsertAttemptWorkspace(improvement.AttemptWorkspace{
		ID:               "workspace-1",
		AttemptID:        attempt.ID,
		ProposalID:       approved.ID,
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		Namespace:        "rsi-platform",
		JobName:          "workspace-job-1",
		PodName:          "workspace-pod-1",
		Status:           improvement.WorkspaceReady,
		AllowedPathGlobs: []string{"internal/**"},
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertAttemptWorkspace() error = %v", err)
	}
	submitProposalStatusForTest(t, store, approved.ID, review.ProposalRepoChangeQueued, "cmd-test-proposal-queued-noop-diff")
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAttempt,
		AggregateID: attempt.ID,
		CommandKind: string(transition.CommandWorkspaceReady),
		CommandID:   "cmd-test-attempt-workspace-ready-noop-diff",
		Actor:       cfg.ServiceName,
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"workspace_id":       "workspace-1",
			"trace_id":           attemptTrace.Summary.TraceID,
			"repo":               "rsi-agent-platform",
			"base_ref":           "main",
			"branch_name":        attempt.BranchName,
			"sandbox_namespace":  "rsi-platform",
			"sandbox_job_name":   "workspace-job-1",
			"sandbox_pod_name":   "workspace-pod-1",
			"validation_ref":     "rsi-platform/workspace-pod-1",
			"validation_summary": "Workspace ready.",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(workspace_ready) error = %v", err)
	}
	submitProposalStatusForTest(t, store, approved.ID, review.ProposalRepoChangeRunning, "cmd-test-proposal-running-noop-diff")
	effect := requireQueuedAttemptEffect(t, store, attempt.ID, transition.EffectInvokeRunner)
	effect, claimed, err := store.ClaimEffectExecution(effect.ID, "tester", 30*time.Second)
	if err != nil || !claimed {
		t.Fatalf("ClaimEffectExecution(invoke_runner) claimed=%t err=%v", claimed, err)
	}
	runner := &fakeRunner{
		resp: clients.RunnerResponse{
			OK: true,
			Raw: map[string]any{
				"structured_output": map[string]any{
					"change_plan":     "Update the direct-write persistence path.",
					"validation_plan": "go test ./...",
					"proposed_actions": []map[string]any{
						{
							"kind": "draft_pr_open",
							"request_payload": map[string]any{
								"title":       "Fix action_result_pkey collision",
								"body":        "Draft PR body",
								"branch_name": attempt.BranchName,
								"base_ref":    "main",
							},
						},
					},
					"retry_assessment": map[string]any{},
				},
			},
		},
	}
	if err := processImplementAttemptEffect(cfg, store, runner, fakeToolClient{}, effect); err != nil {
		t.Fatalf("processImplementAttemptEffect() error = %v", err)
	}
	persisted, ok := store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatalf("expected attempt %s", attempt.ID)
	}
	if persisted.FailureClass != "no_op_diff" {
		t.Fatalf("failure class = %q, want no_op_diff", persisted.FailureClass)
	}
}

func TestProcessProposalImplementAttemptQueuesValidationAfterWorkspaceMutation(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		ServiceName:               "improvement-plane",
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		DefaultReasoningVerbosity: "verbose",
	}
	proposal := store.ListProposals()[0]
	approved, err := store.ReviewProposal(proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with governed remediation.",
		ReviewerID: "alice",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	sourceTrace, ok := store.GetTrace(approved.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", approved.TraceID)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, approved, sourceTrace, queue.WorkItem{
		ProposalID: approved.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: attemptTrace.Summary.TraceID,
		CommandKind: string(transition.CommandProblemLineProjectTrace),
		CommandID:   "cmd-attempt-trace-tool-write",
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"trace_id": attemptTrace.Summary.TraceID,
			"tool_calls": []events.ToolCallRecord{
				{
					ID:         "tool-write-1",
					TraceID:    attemptTrace.Summary.TraceID,
					WorkflowID: attemptTrace.Summary.WorkflowID,
					ToolName:   "workspace.write_file",
					ToolCallID: "tool-write-1",
					Request:    map[string]interface{}{"path": "internal/store/postgres.go"},
					Status:     "ok",
					CreatedAt:  time.Now().UTC(),
				},
			},
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(problem_line_project_trace) error = %v", err)
	}
	if _, err := store.UpsertAttemptWorkspace(improvement.AttemptWorkspace{
		ID:               "workspace-2",
		AttemptID:        attempt.ID,
		ProposalID:       approved.ID,
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		Namespace:        "rsi-platform",
		JobName:          "workspace-job-2",
		PodName:          "workspace-pod-2",
		Status:           improvement.WorkspaceReady,
		AllowedPathGlobs: []string{"internal/**"},
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertAttemptWorkspace() error = %v", err)
	}
	item, err := enqueueImprovementOperationWork(store, operation.Execution{
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attempt.ID,
		OperationKind: proposalOperationImplementAttempt,
		OperationKey:  proposalOperationImplementAttempt,
		Status:        operation.StatusQueued,
		Queue:         queue.ProposalQueue,
		RequestedBy:   cfg.ServiceName,
		TraceID:       attemptTrace.Summary.TraceID,
		ProposalID:    approved.ID,
		AttemptID:     attempt.ID,
	}, queue.WorkItem{
		ID:         "work-implement-2",
		Queue:      queue.ProposalQueue,
		Kind:       proposalOperationImplementAttempt,
		Status:     queue.WorkQueued,
		TraceID:    attemptTrace.Summary.TraceID,
		ProposalID: approved.ID,
		Payload: map[string]any{
			"attempt_id":   attempt.ID,
			"workspace_id": "workspace-2",
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("enqueueImprovementOperationWork() error = %v", err)
	}
	runner := &fakeRunner{
		resp: clients.RunnerResponse{
			OK: true,
			Raw: map[string]any{
				"structured_output": map[string]any{
					"change_plan":     "Apply an in-scope direct write fix.",
					"validation_plan": "go test ./...",
					"proposed_actions": []map[string]any{
						{
							"kind": "draft_pr_open",
							"request_payload": map[string]any{
								"title":       "Fix action_result_pkey collision",
								"body":        "Draft PR body",
								"branch_name": attempt.BranchName,
								"base_ref":    "main",
							},
						},
					},
					"retry_assessment": map[string]any{},
				},
			},
		},
	}
	toolClient := fakeToolClient{
		results: map[string]storepkg.ToolResult{
			"workspace.git_diff": {
				Name:       "workspace.git_diff",
				ToolCallID: "tool-diff-1",
				Status:     "ok",
				Summary:    "Workspace diff ready.",
				Output: map[string]interface{}{
					"head_sha":      "abc123",
					"changed_files": []string{"internal/store/postgres.go"},
					"diff_summary":  "1 file changed",
					"patch":         "diff --git a/internal/store/postgres.go b/internal/store/postgres.go\n+++ b/internal/store/postgres.go\n@@\n+fix\n",
				},
			},
		},
	}
	if err := processProposalItem(cfg, store, runner, toolClient, nil, nil, item); !errors.Is(err, errProposalPhaseHandled) {
		t.Fatalf("processProposalItem(implement_attempt) error = %v", err)
	}
	foundValidate := false
	for _, work := range store.ListWorkItems() {
		if work.Queue == queue.ProposalQueue && work.Kind == proposalOperationWorkspaceValidate && work.ProposalID == approved.ID && work.Status == queue.WorkQueued {
			foundValidate = true
		}
	}
	if !foundValidate {
		t.Fatal("expected workspace_validate phase to be queued")
	}
}

func TestBuildProposalRunnerTaskUsesReadOnlyToolBudget(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposal := store.ListProposals()[0]
	trace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	attempt := improvement.ChangeAttempt{
		ID:            "attempt-test-1",
		ProposalID:    proposal.ID,
		CandidateKey:  proposal.CandidateKey,
		AttemptNumber: 1,
		TargetLayer:   proposal.TargetLayer,
		TargetKind:    proposal.TargetKind,
		TargetRef:     proposal.TargetRef,
	}
	task := buildProposalRunnerTask(config.Config{
		Environment:               "stage",
		DefaultRepo:               "depin-backend",
		AllowedTargetRepos:        []string{"depin-backend", "rsi-agent-platform"},
		DefaultReasoningVerbosity: "verbose",
	}, store, trace, proposal, attempt, nil, store.ListProposalMemories())

	if task.TimeoutSeconds != 420 {
		t.Fatalf("proposal timeout = %d, want 420", task.TimeoutSeconds)
	}
	if len(task.AllowedTools) == 0 || len(task.ToolAllowlist) == 0 {
		t.Fatalf("expected proposal read-only tools, got allowed=%v allowlist=%v", task.AllowedTools, task.ToolAllowlist)
	}
	assertContains(t, task.ToolAllowlist, "repo.context")
	assertContains(t, task.ToolAllowlist, "rsi.trace_context")
	assertContains(t, task.ToolAllowlist, "rsi.candidate_context")
	assertContains(t, task.ToolAllowlist, "rsi.proposal_memory")
	if task.Repo != "rsi-agent-platform" {
		t.Fatalf("proposal task repo = %q, want rsi-agent-platform", task.Repo)
	}
	if len(task.RepoAllowlist) != 1 || task.RepoAllowlist[0] != "rsi-agent-platform" {
		t.Fatalf("proposal repo allowlist = %v, want [rsi-agent-platform]", task.RepoAllowlist)
	}
	if !strings.Contains(task.Prompt, "authoritative target repository is rsi-agent-platform") {
		t.Fatalf("proposal prompt missing target-repo guard: %s", task.Prompt)
	}
	expectedSessionScopeID := proposal.CandidateKey + "|repo:rsi-agent-platform|v2"
	if task.SessionScopeID != expectedSessionScopeID {
		t.Fatalf("proposal session scope id = %q", task.SessionScopeID)
	}
	if task.UserPeerID != "candidate:"+expectedSessionScopeID {
		t.Fatalf("proposal user peer id = %q", task.UserPeerID)
	}
}

func TestProcessProposalWorkspaceValidateUsesFormalValidationStartCommand(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{ServiceName: "improvement-plane"}
	base := store.ListProposals()[0]
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	sourceTrace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, sourceTrace, queue.WorkItem{
		ProposalID: proposal.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	attempt.State = improvement.AttemptStatePatchGenerated
	attempt.ValidationPlan = "go test ./..."
	attempt.BranchName = "codex/test-validation-start"
	attempt.UpdatedAt = time.Now().UTC()
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		t.Fatalf("UpsertChangeAttempt() error = %v", err)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalRepoChangeRunning, "cmd-test-proposal-running-workspace")
	if _, err := store.UpsertAttemptWorkspace(improvement.AttemptWorkspace{
		ID:               "workspace-validate-1",
		AttemptID:        attempt.ID,
		ProposalID:       proposal.ID,
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		Namespace:        "rsi-platform",
		JobName:          "workspace-job-validate-1",
		PodName:          "workspace-pod-validate-1",
		Status:           improvement.WorkspaceReady,
		AllowedPathGlobs: []string{"internal/**"},
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertAttemptWorkspace() error = %v", err)
	}
	item, err := enqueueImprovementOperationWork(store, operation.Execution{
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attempt.ID,
		OperationKind: proposalOperationWorkspaceValidate,
		OperationKey:  proposalOperationWorkspaceValidate,
		Status:        operation.StatusQueued,
		Queue:         queue.ProposalQueue,
		RequestedBy:   cfg.ServiceName,
		TraceID:       attemptTrace.Summary.TraceID,
		ProposalID:    proposal.ID,
		AttemptID:     attempt.ID,
	}, queue.WorkItem{
		ID:         "work-validate-start-1",
		Queue:      queue.ProposalQueue,
		Kind:       proposalOperationWorkspaceValidate,
		Status:     queue.WorkQueued,
		TraceID:    attemptTrace.Summary.TraceID,
		ProposalID: proposal.ID,
		Payload: map[string]any{
			"attempt_id":         attempt.ID,
			"workspace_id":       "workspace-validate-1",
			"validation_command": "go test ./...",
			"branch_name":        attempt.BranchName,
			"base_ref":           "main",
			"title":              "Fix formal validation start",
			"body":               "Draft PR body",
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("enqueueImprovementOperationWork() error = %v", err)
	}

	toolClient := fakeToolClient{
		results: map[string]storepkg.ToolResult{
			"workspace.run_validation": {
				Name:       "workspace.run_validation",
				ToolCallID: "tool-validate-1",
				Status:     "ok",
				Summary:    "Validation passed.",
				Output: map[string]any{
					"stdout": "ok",
				},
			},
		},
	}

	if err := processProposalItem(cfg, store, nil, toolClient, nil, nil, item); !errors.Is(err, errProposalPhaseHandled) {
		t.Fatalf("processProposalItem(workspace_validate) error = %v", err)
	}

	attempt, ok = store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatalf("expected attempt %s", attempt.ID)
	}
	if attempt.State != improvement.AttemptStateValidationRunning {
		t.Fatalf("expected validation_running attempt state, got %s", attempt.State)
	}
	if receipt, ok := store.GetCommandReceipt("cmd-attempt:" + attempt.ID + ":validation_started:" + item.OperationID); !ok || receipt.MachineKind != transition.MachineAttempt {
		t.Fatalf("expected validation_started attempt command receipt, got ok=%t receipt=%+v", ok, receipt)
	}

	attemptTrace, ok = store.GetTrace(attempt.AttemptTraceID)
	if !ok {
		t.Fatalf("expected attempt trace %s", attempt.AttemptTraceID)
	}
	foundValidationStarted := false
	for _, event := range attemptTrace.Events {
		if event.EventType == "workspace.validation.started" {
			foundValidationStarted = true
			break
		}
	}
	if !foundValidationStarted {
		t.Fatal("expected workspace.validation.started event on attempt trace")
	}

	foundPROpenEffect := false
	for _, effect := range store.ListEffectExecutions() {
		if effect.MachineKind == transition.MachineAttempt && effect.EffectKind == transition.EffectOpenDraftPR && effect.AggregateID == attempt.ID {
			foundPROpenEffect = true
			break
		}
	}
	if !foundPROpenEffect {
		t.Fatal("expected queued open_draft_pr effect")
	}
}

func TestProcessProposalWorkspaceOpenEffectRequeuesOnDefer(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{ServiceName: "improvement-plane", WorkItemLeaseDuration: 30 * time.Second}
	base := store.ListProposals()[0]
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	sourceTrace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, sourceTrace, queue.WorkItem{
		ProposalID: proposal.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	if _, err := store.UpsertAttemptWorkspace(improvement.AttemptWorkspace{
		ID:               "workspace-effect-defer",
		AttemptID:        attempt.ID,
		ProposalID:       proposal.ID,
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		Namespace:        "rsi-platform",
		JobName:          "workspace-job-effect-defer",
		Status:           improvement.WorkspaceQueued,
		AllowedPathGlobs: []string{"internal/**"},
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertAttemptWorkspace() error = %v", err)
	}
	item, err := enqueueImprovementOperationWork(store, operation.Execution{
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attempt.ID,
		OperationKind: proposalOperationAttemptPlan,
		OperationKey:  proposalOperationAttemptPlan,
		Status:        operation.StatusQueued,
		Queue:         queue.ProposalQueue,
		RequestedBy:   cfg.ServiceName,
		TraceID:       attemptTrace.Summary.TraceID,
		ProposalID:    proposal.ID,
		AttemptID:     attempt.ID,
	}, queue.WorkItem{
		ID:         "work-effect-workspace-open-attempt-plan",
		Queue:      queue.ProposalQueue,
		Kind:       proposalOperationAttemptPlan,
		Status:     queue.WorkQueued,
		TraceID:    attemptTrace.Summary.TraceID,
		ProposalID: proposal.ID,
		Payload: map[string]any{
			"attempt_id": attempt.ID,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("enqueueImprovementOperationWork() error = %v", err)
	}
	if err := processProposalItem(cfg, store, nil, nil, nil, nil, item); !errors.Is(err, errProposalPhaseHandled) {
		t.Fatalf("processProposalItem(attempt_plan) error = %v", err)
	}
	effect, claimed, err := claimNextImprovementEffect(store, "tester", 30*time.Second, 30*time.Second)
	if err != nil {
		t.Fatalf("claimNextImprovementEffect() error = %v", err)
	}
	if !claimed {
		t.Fatal("expected queued open_workspace effect")
	}
	if effect.EffectKind != transition.EffectOpenWorkspace {
		t.Fatalf("effect kind = %s, want %s", effect.EffectKind, transition.EffectOpenWorkspace)
	}
	if err := processImprovementEffect(cfg, store, nil, nil, fakeToolClient{}, fakeLauncher{}, nil, effect); err != nil {
		t.Fatalf("processImprovementEffect() error = %v", err)
	}

	workspaceEffectCompleted := false
	workspaceEffectRequeued := false
	for _, candidate := range store.ListEffectExecutions() {
		if candidate.ID == effect.ID {
			workspaceEffectCompleted = candidate.Status == transition.EffectCompleted
			continue
		}
		if candidate.MachineKind == transition.MachineAttempt && candidate.AggregateID == attempt.ID && candidate.EffectKind == transition.EffectOpenWorkspace && candidate.Status == transition.EffectQueued {
			workspaceEffectRequeued = true
		}
	}
	if !workspaceEffectCompleted {
		t.Fatalf("expected effect %s to complete after deferral", effect.ID)
	}
	if !workspaceEffectRequeued {
		t.Fatal("expected deferred workspace_open to queue a follow-on effect")
	}
	workspaceWorkItemID := strings.TrimSpace(stringValue(effect.Payload["work_item_id"]))
	workspaceWorkItem, ok := findWorkItemByID(store.ListWorkItems(), workspaceWorkItemID)
	if !ok {
		t.Fatalf("expected workspace_open work item %s", workspaceWorkItemID)
	}
	if workspaceWorkItem.Status != queue.WorkQueued {
		t.Fatalf("expected workspace_open work item to remain queued, got %s", workspaceWorkItem.Status)
	}
}

func TestProcessProposalImplementAttemptEffectQueuesValidation(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		ServiceName:               "improvement-plane",
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		DefaultReasoningVerbosity: "verbose",
		WorkItemLeaseDuration:     30 * time.Second,
	}
	base := store.ListProposals()[0]
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	sourceTrace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, sourceTrace, queue.WorkItem{
		ProposalID: proposal.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: attemptTrace.Summary.TraceID,
		CommandKind: string(transition.CommandProblemLineProjectTrace),
		CommandID:   "cmd-attempt-trace-tool-write-effect-implement",
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"trace_id": attemptTrace.Summary.TraceID,
			"tool_calls": []events.ToolCallRecord{
				{
					ID:         "tool-write-effect-implement",
					TraceID:    attemptTrace.Summary.TraceID,
					WorkflowID: attemptTrace.Summary.WorkflowID,
					ToolName:   "workspace.write_file",
					ToolCallID: "tool-write-effect-implement",
					Request:    map[string]any{"path": "internal/store/postgres.go"},
					Status:     "ok",
					CreatedAt:  time.Now().UTC(),
				},
			},
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(problem_line_project_trace) error = %v", err)
	}
	if _, err := store.UpsertAttemptWorkspace(improvement.AttemptWorkspace{
		ID:               "workspace-effect-implement",
		AttemptID:        attempt.ID,
		ProposalID:       proposal.ID,
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		Namespace:        "rsi-platform",
		JobName:          "workspace-job-effect-implement",
		PodName:          "workspace-pod-effect-implement",
		Status:           improvement.WorkspaceReady,
		AllowedPathGlobs: []string{"internal/**"},
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertAttemptWorkspace() error = %v", err)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalRepoChangeQueued, "cmd-test-proposal-queued-effect-implement")
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineAttempt,
		AggregateID: attempt.ID,
		CommandKind: string(transition.CommandWorkspaceReady),
		CommandID:   "cmd-test-attempt-workspace-ready-effect-implement",
		Actor:       cfg.ServiceName,
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"workspace_id":       "workspace-effect-implement",
			"trace_id":           attemptTrace.Summary.TraceID,
			"repo":               "rsi-agent-platform",
			"base_ref":           "main",
			"branch_name":        attempt.BranchName,
			"sandbox_namespace":  "rsi-platform",
			"sandbox_job_name":   "workspace-job-effect-implement",
			"sandbox_pod_name":   "workspace-pod-effect-implement",
			"validation_ref":     "rsi-platform/workspace-pod-effect-implement",
			"validation_summary": "Workspace ready.",
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(workspace_ready) error = %v", err)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalRepoChangeRunning, "cmd-test-proposal-running-effect-implement")
	invokeRunner := requireQueuedAttemptEffect(t, store, attempt.ID, transition.EffectInvokeRunner)
	effect, claimed, err := store.ClaimEffectExecution(invokeRunner.ID, "tester", 30*time.Second)
	if err != nil {
		t.Fatalf("ClaimEffectExecution(invoke_runner) error = %v", err)
	}
	if !claimed {
		t.Fatal("expected queued implement_attempt effect")
	}
	effect.Payload["work_item_id"] = ""
	runner := &fakeRunner{
		resp: clients.RunnerResponse{
			OK: true,
			Raw: map[string]any{
				"structured_output": map[string]any{
					"change_plan":     "Apply the direct-write fix.",
					"validation_plan": "go test ./...",
					"proposed_actions": []map[string]any{
						{
							"kind": "draft_pr_open",
							"request_payload": map[string]any{
								"title":       "Formalize attempt effect execution",
								"body":        "Draft PR body",
								"branch_name": attempt.BranchName,
								"base_ref":    "main",
							},
						},
					},
					"retry_assessment": map[string]any{},
				},
			},
		},
	}
	toolClient := fakeToolClient{
		results: map[string]storepkg.ToolResult{
			"workspace.git_diff": {
				Name:       "workspace.git_diff",
				ToolCallID: "tool-diff-effect-implement",
				Status:     "ok",
				Summary:    "Workspace diff ready.",
				Output: map[string]any{
					"head_sha":      "abc123",
					"changed_files": []string{"internal/store/postgres.go"},
					"diff_summary":  "1 file changed",
					"patch":         "diff --git a/internal/store/postgres.go b/internal/store/postgres.go\n+++ b/internal/store/postgres.go\n@@\n+fix\n",
				},
			},
		},
	}
	if err := processImprovementEffect(cfg, store, nil, runner, toolClient, nil, nil, effect); err != nil {
		t.Fatalf("processImprovementEffect() error = %v", err)
	}
	foundCompleted := false
	foundValidateEffect := false
	for _, candidate := range store.ListEffectExecutions() {
		if candidate.ID == effect.ID {
			foundCompleted = candidate.Status == transition.EffectCompleted
		}
		if candidate.MachineKind == transition.MachineAttempt && candidate.AggregateID == attempt.ID && candidate.EffectKind == transition.EffectWorkspaceValidate && candidate.Status == transition.EffectQueued {
			foundValidateEffect = true
		}
	}
	if !foundCompleted {
		t.Fatalf("expected effect %s to be completed", effect.ID)
	}
	if !foundValidateEffect {
		t.Fatal("expected workspace_validate effect to be queued")
	}
}

func TestProcessProposalWorkspaceValidateEffectQueuesDraftPROpen(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{
		ServiceName:               "improvement-plane",
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		DefaultReasoningVerbosity: "verbose",
		WorkItemLeaseDuration:     30 * time.Second,
	}
	base := store.ListProposals()[0]
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	sourceTrace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, sourceTrace, queue.WorkItem{
		ProposalID: proposal.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	if _, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: attemptTrace.Summary.TraceID,
		CommandKind: string(transition.CommandProblemLineProjectTrace),
		CommandID:   "cmd-attempt-trace-tool-write-effect-validate",
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"trace_id": attemptTrace.Summary.TraceID,
			"tool_calls": []events.ToolCallRecord{
				{
					ID:         "tool-write-effect-validate",
					TraceID:    attemptTrace.Summary.TraceID,
					WorkflowID: attemptTrace.Summary.WorkflowID,
					ToolName:   "workspace.write_file",
					ToolCallID: "tool-write-effect-validate",
					Request:    map[string]any{"path": "internal/store/postgres.go"},
					Status:     "ok",
					CreatedAt:  time.Now().UTC(),
				},
			},
		},
	}); err != nil {
		t.Fatalf("SubmitCommand(problem_line_project_trace) error = %v", err)
	}
	if _, err := store.UpsertAttemptWorkspace(improvement.AttemptWorkspace{
		ID:               "workspace-effect-validate",
		AttemptID:        attempt.ID,
		ProposalID:       proposal.ID,
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		Namespace:        "rsi-platform",
		JobName:          "workspace-job-effect-validate",
		PodName:          "workspace-pod-effect-validate",
		Status:           improvement.WorkspaceReady,
		AllowedPathGlobs: []string{"internal/**"},
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertAttemptWorkspace() error = %v", err)
	}
	item, err := enqueueImprovementOperationWork(store, operation.Execution{
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attempt.ID,
		OperationKind: proposalOperationImplementAttempt,
		OperationKey:  proposalOperationImplementAttempt,
		Status:        operation.StatusQueued,
		Queue:         queue.ProposalQueue,
		RequestedBy:   cfg.ServiceName,
		TraceID:       attemptTrace.Summary.TraceID,
		ProposalID:    proposal.ID,
		AttemptID:     attempt.ID,
	}, queue.WorkItem{
		ID:         "work-effect-validate-implement",
		Queue:      queue.ProposalQueue,
		Kind:       proposalOperationImplementAttempt,
		Status:     queue.WorkQueued,
		TraceID:    attemptTrace.Summary.TraceID,
		ProposalID: proposal.ID,
		Payload: map[string]any{
			"attempt_id":   attempt.ID,
			"workspace_id": "workspace-effect-validate",
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("enqueueImprovementOperationWork() error = %v", err)
	}
	runner := &fakeRunner{
		resp: clients.RunnerResponse{
			OK: true,
			Raw: map[string]any{
				"structured_output": map[string]any{
					"change_plan":     "Apply the direct-write fix.",
					"validation_plan": "go test ./...",
					"proposed_actions": []map[string]any{
						{
							"kind": "draft_pr_open",
							"request_payload": map[string]any{
								"title":       "Formalize validation effect execution",
								"body":        "Draft PR body",
								"branch_name": attempt.BranchName,
								"base_ref":    "main",
							},
						},
					},
					"retry_assessment": map[string]any{},
				},
			},
		},
	}
	toolClient := fakeToolClient{
		results: map[string]storepkg.ToolResult{
			"workspace.git_diff": {
				Name:       "workspace.git_diff",
				ToolCallID: "tool-diff-effect-validate",
				Status:     "ok",
				Summary:    "Workspace diff ready.",
				Output: map[string]any{
					"head_sha":      "abc123",
					"changed_files": []string{"internal/store/postgres.go"},
					"diff_summary":  "1 file changed",
					"patch":         "diff --git a/internal/store/postgres.go b/internal/store/postgres.go\n+++ b/internal/store/postgres.go\n@@\n+fix\n",
				},
			},
			"workspace.run_validation": {
				Name:       "workspace.run_validation",
				ToolCallID: "tool-validate-effect-validate",
				Status:     "ok",
				Summary:    "Validation passed.",
				Output: map[string]any{
					"stdout": "ok",
				},
			},
		},
	}
	if err := processProposalItem(cfg, store, runner, toolClient, nil, nil, item); !errors.Is(err, errProposalPhaseHandled) {
		t.Fatalf("processProposalItem(implement_attempt) error = %v", err)
	}
	effect, claimed, err := claimNextImprovementEffect(store, "tester", 30*time.Second, 30*time.Second)
	if err != nil {
		t.Fatalf("claimNextImprovementEffect() error = %v", err)
	}
	if !claimed {
		t.Fatal("expected queued workspace_validate effect")
	}
	if effect.EffectKind != transition.EffectWorkspaceValidate {
		t.Fatalf("effect kind = %s, want %s", effect.EffectKind, transition.EffectWorkspaceValidate)
	}
	effect.Payload["work_item_id"] = ""
	if err := processImprovementEffect(cfg, store, nil, nil, toolClient, nil, nil, effect); err != nil {
		t.Fatalf("processImprovementEffect() error = %v", err)
	}
	foundCompleted := false
	foundPROpenEffect := false
	for _, candidate := range store.ListEffectExecutions() {
		if candidate.ID == effect.ID {
			foundCompleted = candidate.Status == transition.EffectCompleted
		}
		if candidate.MachineKind == transition.MachineAttempt && candidate.AggregateID == attempt.ID && candidate.EffectKind == transition.EffectOpenDraftPR && candidate.Status == transition.EffectQueued {
			foundPROpenEffect = true
		}
	}
	if !foundCompleted {
		t.Fatalf("expected effect %s to be completed", effect.ID)
	}
	if !foundPROpenEffect {
		t.Fatalf("expected queued open_draft_pr effect, got effects=%+v", store.ListEffectExecutions())
	}
}

func TestProcessDraftPROpenEffectUsesEffectExecution(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{ServiceName: "improvement-plane", WorkItemLeaseDuration: 30 * time.Second}
	base := store.ListProposals()[0]
	proposal, err := store.ReviewProposal(base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	sourceTrace, ok := store.GetTrace(proposal.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", proposal.TraceID)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, sourceTrace, queue.WorkItem{
		ProposalID: proposal.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	attempt.State = improvement.AttemptStatePatchGenerated
	attempt.ValidationPlan = "go test ./..."
	attempt.BranchName = "codex/test-effect-pr-open"
	attempt.UpdatedAt = time.Now().UTC()
	if _, err := store.UpsertChangeAttempt(attempt); err != nil {
		t.Fatalf("UpsertChangeAttempt() error = %v", err)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalRepoChangeRunning, "cmd-test-proposal-running-effect-pr")
	if _, err := store.UpsertAttemptWorkspace(improvement.AttemptWorkspace{
		ID:               "workspace-effect-pr-open",
		AttemptID:        attempt.ID,
		ProposalID:       proposal.ID,
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       attempt.BranchName,
		Namespace:        "rsi-platform",
		JobName:          "workspace-job-effect-pr-open",
		PodName:          "workspace-pod-effect-pr-open",
		Status:           improvement.WorkspaceReady,
		AllowedPathGlobs: []string{"internal/**"},
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertAttemptWorkspace() error = %v", err)
	}
	item, err := enqueueImprovementOperationWork(store, operation.Execution{
		ScopeKind:     operation.ScopeAttempt,
		ScopeID:       attempt.ID,
		OperationKind: proposalOperationWorkspaceValidate,
		OperationKey:  proposalOperationWorkspaceValidate,
		Status:        operation.StatusQueued,
		Queue:         queue.ProposalQueue,
		RequestedBy:   cfg.ServiceName,
		TraceID:       attemptTrace.Summary.TraceID,
		ProposalID:    proposal.ID,
		AttemptID:     attempt.ID,
	}, queue.WorkItem{
		ID:         "work-effect-pr-open-validate",
		Queue:      queue.ProposalQueue,
		Kind:       proposalOperationWorkspaceValidate,
		Status:     queue.WorkQueued,
		TraceID:    attemptTrace.Summary.TraceID,
		ProposalID: proposal.ID,
		Payload: map[string]any{
			"attempt_id":         attempt.ID,
			"workspace_id":       "workspace-effect-pr-open",
			"validation_command": "go test ./...",
			"branch_name":        attempt.BranchName,
			"base_ref":           "main",
			"title":              "Open PR through effect execution",
			"body":               "Draft PR body",
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("enqueueImprovementOperationWork() error = %v", err)
	}
	toolClient := fakeToolClient{
		results: map[string]storepkg.ToolResult{
			"workspace.run_validation": {
				Name:       "workspace.run_validation",
				ToolCallID: "tool-validate-effect-pr",
				Status:     "ok",
				Summary:    "Validation passed.",
				Output: map[string]any{
					"stdout": "ok",
				},
			},
			"github.create_pr": {
				Available:   true,
				Status:      "ok",
				Provider:    "github",
				ProviderRef: "pull/456",
				Output: map[string]any{
					"pr_url": "https://github.com/piplabs/rsi-agent-platform/pull/456",
					"response": map[string]any{
						"head": map[string]any{
							"sha": "def456",
						},
					},
				},
			},
		},
	}

	if err := processProposalItem(cfg, store, nil, toolClient, nil, nil, item); !errors.Is(err, errProposalPhaseHandled) {
		t.Fatalf("processProposalItem(workspace_validate) error = %v", err)
	}
	effect, claimed, err := claimNextImprovementEffect(store, "tester", 30*time.Second, 30*time.Second)
	if err != nil {
		t.Fatalf("claimNextImprovementEffect() error = %v", err)
	}
	if !claimed {
		t.Fatal("expected queued open_draft_pr effect")
	}
	if payload, ok := effect.Payload["work_item_payload"].(map[string]interface{}); ok {
		for _, key := range []string{"attempt_id", "job_id", "job_name", "namespace", "repo", "branch_name", "base_ref", "title", "body"} {
			if _, exists := effect.Payload[key]; exists {
				continue
			}
			if value, exists := payload[key]; exists {
				effect.Payload[key] = value
			}
		}
	}
	effect.Payload["work_item_id"] = ""
	delete(effect.Payload, "work_item_payload")
	if err := processImprovementEffect(cfg, store, nil, nil, toolClient, nil, nil, effect); err != nil {
		t.Fatalf("processImprovementEffect() error = %v", err)
	}

	effects := store.ListEffectExecutions()
	foundCompleted := false
	for _, candidate := range effects {
		if candidate.ID == effect.ID {
			foundCompleted = candidate.Status == transition.EffectCompleted
			break
		}
	}
	if !foundCompleted {
		t.Fatalf("expected effect %s to be completed, effects=%+v", effect.ID, effects)
	}

	attempt, ok = store.GetChangeAttempt(attempt.ID)
	if !ok {
		t.Fatalf("expected attempt %s", attempt.ID)
	}
	if attempt.State != improvement.AttemptStateCIObserving {
		t.Fatalf("expected attempt state ci_observing, got %s", attempt.State)
	}
	if attempt.PRURL != "https://github.com/piplabs/rsi-agent-platform/pull/456" {
		t.Fatalf("unexpected attempt pr url %q", attempt.PRURL)
	}
}

func TestBuildEvalRunnerTaskUsesReadOnlyToolBudget(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: traceID,
		CommandKind: string(transition.CommandProblemLineEvaluateTrace),
		CommandID:   "cmd-eval-runner-task-evaluate",
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"trigger": "incident-response",
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand(problem_line_evaluate_trace) error = %v", err)
	}
	var run evals.Run
	foundRun := false
	for _, candidate := range store.ListEvalRuns() {
		if candidate.ID == receipt.ResultRef {
			run = candidate
			foundRun = true
			break
		}
	}
	if !foundRun {
		t.Fatalf("expected eval run %s", receipt.ResultRef)
	}
	judgments := store.ListEvalJudgments(run.ID)
	trace, ok := store.GetTrace(traceID)
	if !ok {
		t.Fatalf("expected trace %s", traceID)
	}
	task := buildEvalRunnerTask(config.Config{
		Environment:               "stage",
		DefaultRepo:               "rsi-agent-platform",
		AllowedTargetRepos:        []string{"rsi-agent-platform"},
		DefaultReasoningVerbosity: "verbose",
	}, store, trace, run, judgments, queue.WorkItem{Kind: "evaluate_trace"})

	if task.TimeoutSeconds != 300 {
		t.Fatalf("eval timeout = %d, want 300", task.TimeoutSeconds)
	}
	assertContains(t, task.ToolAllowlist, "knowledge.context")
	assertContains(t, task.ToolAllowlist, "rsi.trace_context")
	assertContains(t, task.ToolAllowlist, "rsi.candidate_context")
}

func TestEnsureAttemptWorkspaceReturnsPersistenceError(t *testing.T) {
	base := storepkg.NewMemoryStore()
	if _, err := base.UpsertAttemptWorkspace(improvement.AttemptWorkspace{
		ID:         "workspace-1",
		AttemptID:  "attempt-1",
		ProposalID: "proposal-1",
		Repo:       "rsi-agent-platform",
		BaseRef:    "main",
		BranchName: "codex/proposal-1/attempt-01",
		Namespace:  "rsi-platform",
		JobName:    "workspace-job-1",
		Status:     improvement.WorkspaceQueued,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertAttemptWorkspace() error = %v", err)
	}
	cfg := config.Config{ServiceName: "improvement-plane"}
	proposal := review.Proposal{
		ID:                          "proposal-1",
		ConversationID:              "conv-1",
		CaseID:                      "case-1",
		TraceID:                     "trace-1",
		OriginTraceID:               "trace-1",
		CandidateKey:                "candidate-1",
		TargetLayer:                 "repo_change",
		TargetKind:                  "repo",
		TargetRef:                   "rsi-agent-platform",
		RecommendedInterventionKind: review.InterventionRepoChange,
		Summary:                     "Fix workspace persistence.",
		ValidationPlan:              "make test",
	}
	attempt := improvement.ChangeAttempt{
		ID:         "attempt-1",
		ProposalID: proposal.ID,
		BranchName: "codex/proposal-1/attempt-01",
	}
	workspace, ready, err := ensureAttemptWorkspace(cfg, base, workspaceLaunchStub{}, nil, proposal, attempt, "trace-1")
	if err != nil {
		t.Fatalf("ensureAttemptWorkspace() error = %v", err)
	}
	if !ready {
		t.Fatal("expected resolved workspace to be ready")
	}
	if workspace.PodName != "workspace-pod-1" || workspace.Status != improvement.WorkspaceReady {
		t.Fatalf("unexpected workspace %+v", workspace)
	}
	stored, ok := base.GetAttemptWorkspaceByAttempt("attempt-1")
	if !ok {
		t.Fatal("expected stored workspace")
	}
	if stored.PodName != "" || stored.Status != improvement.WorkspaceQueued {
		t.Fatalf("expected ensureAttemptWorkspace to avoid persisting directly, got %+v", stored)
	}
}

func assertContains(t *testing.T, values []string, target string) {
	t.Helper()
	for _, item := range values {
		if item == target {
			return
		}
	}
	t.Fatalf("expected %q in %v", target, values)
}
