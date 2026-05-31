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
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
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

func ensureProposalAttemptForTest(cfg config.Config, store *storepkg.MemoryStore, proposal review.Proposal, payload map[string]any) (improvement.ChangeAttempt, events.Trace, error) {
	refreshed, ok := findProposal(store.ListProposals(), proposal.ID)
	if !ok {
		return improvement.ChangeAttempt{}, events.Trace{}, fmt.Errorf("proposal %s not found", proposal.ID)
	}
	proposal = refreshed
	if strings.TrimSpace(proposal.CurrentAttemptID) == "" {
		receipt, err := store.SubmitCommand(transition.CommandEnvelope{
			MachineKind: transition.MachineProposalLine,
			AggregateID: proposal.ID,
			CommandKind: string(transition.CommandProposalResumeExecution),
			CommandID:   fmt.Sprintf("cmd-test-proposal-resume:%s", proposal.ID),
			Actor:       cfg.ServiceName,
			OccurredAt:  time.Now().UTC(),
			Payload:     payload,
		})
		if err != nil {
			return improvement.ChangeAttempt{}, events.Trace{}, err
		}
		if receipt.DecisionKind == transition.DecisionReject {
			return improvement.ChangeAttempt{}, events.Trace{}, errors.New(receipt.Reason)
		}
		proposal, ok = findProposal(store.ListProposals(), proposal.ID)
		if !ok {
			return improvement.ChangeAttempt{}, events.Trace{}, fmt.Errorf("proposal %s not found after resume", proposal.ID)
		}
	}
	attemptID := strings.TrimSpace(proposal.CurrentAttemptID)
	if attemptID == "" {
		return improvement.ChangeAttempt{}, events.Trace{}, errors.New("proposal did not materialize an attempt")
	}
	attempt, ok := store.GetChangeAttempt(attemptID)
	if !ok {
		return improvement.ChangeAttempt{}, events.Trace{}, fmt.Errorf("attempt %s not found", attemptID)
	}
	trace, ok := store.GetTrace(attempt.AttemptTraceID)
	if !ok {
		return improvement.ChangeAttempt{}, events.Trace{}, fmt.Errorf("attempt trace %s not found", attempt.AttemptTraceID)
	}
	return attempt, trace, nil
}

func submitProblemLineCommandForTest(t *testing.T, store *storepkg.MemoryStore, aggregateID string, kind transition.ProblemLineCommandKind, commandID string, payload map[string]any) transition.CommandReceipt {
	t.Helper()
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: aggregateID,
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

func submitSettingsCommandForTest(t *testing.T, store *storepkg.MemoryStore, commandID string, payload map[string]any) {
	t.Helper()
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineSettings,
		AggregateID: "default",
		CommandKind: string(transition.CommandSettingsUpdate),
		CommandID:   commandID,
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
		Payload:     payload,
	})
	if err != nil {
		t.Fatalf("SubmitCommand(%s) error = %v", transition.CommandSettingsUpdate, err)
	}
	if receipt.DecisionKind == transition.DecisionReject {
		t.Fatalf("SubmitCommand(%s) rejected: %s", transition.CommandSettingsUpdate, receipt.Reason)
	}
}

func seedHarnessOverlayProposalForTest(t *testing.T, store *storepkg.MemoryStore) review.Proposal {
	t.Helper()

	slackEventIDs := map[string]struct{}{}
	for _, event := range store.ListEvents() {
		if event.Source == "slack" {
			slackEventIDs[event.ID] = struct{}{}
		}
	}
	var trace events.Trace
	foundTrace := false
	for _, candidate := range store.ListTraces() {
		if _, ok := slackEventIDs[candidate.TriggerEventID]; !ok {
			continue
		}
		loadedTrace, ok := store.GetTrace(candidate.TraceID)
		if !ok {
			t.Fatalf("expected trace %s", candidate.TraceID)
		}
		trace = loadedTrace
		foundTrace = true
		break
	}
	if !foundTrace {
		t.Fatal("expected a seeded slack trace")
	}

	submitSettingsCommandForTest(t, store, "cmd-test-settings-overlay-cap", map[string]any{
		"active_proposal_cap": 4,
	})

	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineProblemLine,
		AggregateID: trace.Summary.TraceID,
		CommandKind: string(transition.CommandProblemLineRecordRating),
		CommandID:   "cmd-test-overlay-rating",
		Actor:       "tester",
		OccurredAt:  time.Now().UTC(),
		Payload: map[string]any{
			"score":       1,
			"verdict":     "behavioral_regression",
			"labels":      []string{"overlay"},
			"notes":       "Prompt and workflow behavior degraded despite a nominally completed trace.",
			"reviewer_id": "tester",
		},
	})
	if err != nil {
		t.Fatalf("SubmitCommand(problem_line_record_rating) error = %v", err)
	}
	if receipt.DecisionKind == transition.DecisionReject {
		t.Fatalf("SubmitCommand(problem_line_record_rating) rejected: %s", receipt.Reason)
	}

	var overlayCandidate improvement.Candidate
	foundCandidate := false
	for _, candidate := range store.ListCandidates() {
		if candidate.LatestTraceID != trace.Summary.TraceID || candidate.TargetLayer != harness.TargetLayerHarnessOverlay {
			continue
		}
		overlayCandidate = candidate
		foundCandidate = true
		break
	}
	if !foundCandidate {
		t.Fatal("expected harness-overlay candidate after low-rating replay")
	}

	knownProposalIDs := map[string]struct{}{}
	for _, proposal := range store.ListProposals() {
		knownProposalIDs[proposal.ID] = struct{}{}
	}

	promoteReceipt := submitProblemLineCommandForTest(t, store, "tester", transition.CommandProblemLinePromote, "cmd-test-overlay-promote", map[string]any{
		"requested_by": "tester",
		"limit":        4,
	})
	if promoteReceipt.DecisionKind == transition.DecisionReject {
		t.Fatalf("SubmitCommand(%s) rejected: %s", transition.CommandProblemLinePromote, promoteReceipt.Reason)
	}

	for _, proposal := range store.ListProposals() {
		if _, ok := knownProposalIDs[proposal.ID]; ok {
			continue
		}
		if proposal.CandidateKey == overlayCandidate.CandidateKey {
			return proposal
		}
	}
	t.Fatalf("expected promoted proposal for candidate %s", overlayCandidate.CandidateKey)
	return review.Proposal{}
}

func loadProposalAttemptForTest(t *testing.T, store *storepkg.MemoryStore, proposalID string, attemptID string) (review.Proposal, improvement.ChangeAttempt) {
	t.Helper()
	proposal, ok := findProposal(store.ListProposals(), proposalID)
	if !ok {
		t.Fatalf("expected proposal %s", proposalID)
	}
	attempt, ok := store.GetChangeAttempt(attemptID)
	if !ok {
		t.Fatalf("expected attempt %s", attemptID)
	}
	return proposal, attempt
}

func loadRepoChangeJobForTest(t *testing.T, store *storepkg.MemoryStore, attemptID string, jobID string) improvement.RepoChangeJob {
	t.Helper()
	if jobID != "" {
		if job, ok := findRepoChangeJob(store.ListRepoChangeJobs(), jobID); ok {
			return job
		}
	}
	if job, ok := findRepoChangeJobByAttempt(store.ListRepoChangeJobs(), attemptID); ok {
		return job
	}
	t.Fatalf("expected repo change job for attempt %s", attemptID)
	return improvement.RepoChangeJob{}
}

func advanceAttemptToPatchGeneratedForTest(
	t *testing.T,
	cfg config.Config,
	store *storepkg.MemoryStore,
	proposal review.Proposal,
	attempt improvement.ChangeAttempt,
	trace events.Trace,
	workspaceID string,
	jobID string,
	repo string,
	baseRef string,
	namespace string,
	jobName string,
	podName string,
	changePlan string,
	validationPlan string,
) (review.Proposal, improvement.ChangeAttempt, improvement.RepoChangeJob) {
	t.Helper()

	repo = firstNonEmpty(repo, "rsi-agent-platform")
	baseRef = firstNonEmpty(baseRef, "main")
	workspaceID = firstNonEmpty(workspaceID, "workspace-"+attempt.ID)
	jobID = firstNonEmpty(jobID, "job-"+attempt.ID)
	namespace = firstNonEmpty(namespace, "rsi-platform")
	jobName = firstNonEmpty(jobName, "workspace-job-"+attempt.ID)
	podName = firstNonEmpty(podName, "workspace-pod-"+attempt.ID)
	changePlan = firstNonEmpty(changePlan, "Apply the approved remediation.")
	validationPlan = firstNonEmpty(validationPlan, "Run governed validation.")

	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalRepoChangeQueued, "cmd-test-proposal-queued-"+attempt.ID)
	if err := submitAttemptCommand(store, attempt, transition.CommandWorkspaceReady, cfg.ServiceName, time.Now().UTC(), map[string]any{
		"workspace_id":       workspaceID,
		"job_id":             jobID,
		"trace_id":           trace.Summary.TraceID,
		"repo":               repo,
		"base_ref":           baseRef,
		"branch_name":        attempt.BranchName,
		"sandbox_namespace":  namespace,
		"sandbox_job_name":   jobName,
		"sandbox_pod_name":   podName,
		"validation_ref":     fmt.Sprintf("%s/%s", namespace, jobName),
		"validation_summary": "Workspace ready.",
		"allowed_path_globs": []string{"internal/**"},
	}); err != nil {
		t.Fatalf("submitAttemptCommand(workspace_ready) error = %v", err)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalRepoChangeRunning, "cmd-test-proposal-running-"+attempt.ID)
	if err := submitAttemptCommand(store, attempt, transition.CommandImplementationCompleted, cfg.ServiceName, time.Now().UTC(), map[string]any{
		"job_id":             jobID,
		"trace_id":           trace.Summary.TraceID,
		"repo":               repo,
		"base_ref":           baseRef,
		"branch_name":        attempt.BranchName,
		"change_plan":        changePlan,
		"validation_plan":    validationPlan,
		"diff_summary":       "1 file changed",
		"changed_files":      []string{"internal/store/commands.go"},
		"validation_ref":     fmt.Sprintf("%s/%s", namespace, jobName),
		"validation_summary": validationPlan,
	}); err != nil {
		t.Fatalf("submitAttemptCommand(implementation_completed) error = %v", err)
	}

	proposal, attempt = loadProposalAttemptForTest(t, store, proposal.ID, attempt.ID)
	return proposal, attempt, loadRepoChangeJobForTest(t, store, attempt.ID, jobID)
}

func advanceAttemptToValidationRunningForTest(
	t *testing.T,
	cfg config.Config,
	store *storepkg.MemoryStore,
	proposal review.Proposal,
	attempt improvement.ChangeAttempt,
	trace events.Trace,
	workspaceID string,
	jobID string,
	repo string,
	baseRef string,
	namespace string,
	jobName string,
	podName string,
	changePlan string,
	validationPlan string,
) (review.Proposal, improvement.ChangeAttempt, improvement.RepoChangeJob) {
	t.Helper()

	proposal, attempt, job := advanceAttemptToPatchGeneratedForTest(t, cfg, store, proposal, attempt, trace, workspaceID, jobID, repo, baseRef, namespace, jobName, podName, changePlan, validationPlan)
	if err := submitAttemptCommand(store, attempt, transition.CommandValidationStarted, cfg.ServiceName, time.Now().UTC(), map[string]any{
		"job_id":             job.ID,
		"sandbox_namespace":  firstNonEmpty(job.SandboxNamespace, namespace, "rsi-platform"),
		"sandbox_job_name":   firstNonEmpty(job.SandboxJobName, jobName),
		"sandbox_pod_name":   firstNonEmpty(job.SandboxPodName, podName),
		"validation_ref":     firstNonEmpty(job.ValidationRef, fmt.Sprintf("%s/%s", firstNonEmpty(job.SandboxNamespace, namespace, "rsi-platform"), firstNonEmpty(job.SandboxJobName, jobName))),
		"validation_summary": firstNonEmpty(validationPlan, "Run governed validation."),
	}); err != nil {
		t.Fatalf("submitAttemptCommand(validation_started) error = %v", err)
	}

	proposal, attempt = loadProposalAttemptForTest(t, store, proposal.ID, attempt.ID)
	return proposal, attempt, loadRepoChangeJobForTest(t, store, attempt.ID, job.ID)
}

func advanceAttemptToValidationPendingForTest(
	t *testing.T,
	cfg config.Config,
	store *storepkg.MemoryStore,
	proposal review.Proposal,
	attempt improvement.ChangeAttempt,
	trace events.Trace,
	workspaceID string,
	jobID string,
	repo string,
	baseRef string,
	namespace string,
	jobName string,
	podName string,
	changePlan string,
	validationPlan string,
) (review.Proposal, improvement.ChangeAttempt, improvement.RepoChangeJob) {
	t.Helper()

	proposal, attempt, job := advanceAttemptToPatchGeneratedForTest(t, cfg, store, proposal, attempt, trace, workspaceID, jobID, repo, baseRef, namespace, jobName, podName, changePlan, validationPlan)
	if err := submitAttemptCommand(store, attempt, transition.CommandValidationCompleted, cfg.ServiceName, time.Now().UTC(), map[string]any{
		"job_id":             job.ID,
		"repo":               firstNonEmpty(job.Repo, repo, "rsi-agent-platform"),
		"branch_name":        firstNonEmpty(job.BranchName, attempt.BranchName),
		"base_ref":           firstNonEmpty(job.BaseRef, baseRef, "main"),
		"sandbox_namespace":  firstNonEmpty(job.SandboxNamespace, namespace, "rsi-platform"),
		"sandbox_job_name":   firstNonEmpty(job.SandboxJobName, jobName),
		"sandbox_pod_name":   firstNonEmpty(job.SandboxPodName, podName),
		"validation_ref":     firstNonEmpty(job.ValidationRef, fmt.Sprintf("%s/%s", firstNonEmpty(job.SandboxNamespace, namespace, "rsi-platform"), firstNonEmpty(job.SandboxJobName, jobName))),
		"validation_summary": firstNonEmpty(validationPlan, "Run governed validation."),
	}); err != nil {
		t.Fatalf("submitAttemptCommand(validation_completed) error = %v", err)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalValidationPending, "cmd-test-proposal-validation-pending-"+attempt.ID)

	proposal, attempt = loadProposalAttemptForTest(t, store, proposal.ID, attempt.ID)
	return proposal, attempt, loadRepoChangeJobForTest(t, store, attempt.ID, job.ID)
}

func sandboxObservationRequestForTest(effectID string, operationID string, attempt improvement.ChangeAttempt, trace events.Trace, job improvement.RepoChangeJob) sandboxObservationRequest {
	return sandboxObservationRequest{
		EffectID:    effectID,
		OperationID: operationID,
		ProposalID:  attempt.ProposalID,
		AttemptID:   attempt.ID,
		TraceID:     trace.Summary.TraceID,
		JobID:       job.ID,
		JobName:     job.SandboxJobName,
		Namespace:   job.SandboxNamespace,
		Repo:        firstNonEmpty(job.Repo, "rsi-agent-platform"),
		BranchName:  firstNonEmpty(job.BranchName, attempt.BranchName),
		BaseRef:     firstNonEmpty(job.BaseRef, "main"),
	}
}

func draftPROpenRequestForTest(effectID string, operationID string, attempt improvement.ChangeAttempt, trace events.Trace, jobID string, repo string, branchName string, baseRef string, title string, body string) draftPROpenRequest {
	return draftPROpenRequest{
		EffectID:    effectID,
		OperationID: operationID,
		ProposalID:  attempt.ProposalID,
		AttemptID:   attempt.ID,
		TraceID:     trace.Summary.TraceID,
		JobID:       jobID,
		Repo:        repo,
		BranchName:  branchName,
		BaseRef:     baseRef,
		Title:       title,
		Body:        body,
	}
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
	proposal, err := storepkg.ReviewProposalForTesting(store, base.ID, review.ProposalReview{
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
	attempt, _, err := ensureProposalAttemptForTest(config.Config{ServiceName: "improvement-plane"}, store, proposal, map[string]any{
		"trigger": string(improvement.AttemptTriggerProposalApproved),
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	cfg := config.Config{
		ServiceName:         "improvement-plane",
		SandboxPollInterval: 2 * time.Second,
		SandboxNamespace:    "rsi-platform",
	}
	proposal, attempt, repoJob := advanceAttemptToPatchGeneratedForTest(
		t,
		cfg,
		store,
		proposal,
		attempt,
		sourceTrace,
		"workspace-watch-1",
		"job-watch-1",
		"rsi-agent-platform",
		"main",
		"rsi-platform",
		"sandbox-watch-1",
		"sandbox-watch-1-pod",
		"Update the formal transition path.",
		"Run the sandbox validation flow.",
	)
	if err := submitAttemptCommand(store, attempt, transition.CommandValidationStarted, cfg.ServiceName, time.Now().UTC(), map[string]any{
		"operation_id":      "op-sandbox-watch-defer",
		"job_id":            repoJob.ID,
		"sandbox_namespace": firstNonEmpty(repoJob.SandboxNamespace, "rsi-platform"),
		"sandbox_job_name":  firstNonEmpty(repoJob.SandboxJobName, "sandbox-watch-1"),
		"sandbox_pod_name":  firstNonEmpty(repoJob.SandboxPodName, "sandbox-watch-1-pod"),
		"validation_ref":    firstNonEmpty(repoJob.ValidationRef, "rsi-platform/sandbox-watch-1"),
	}); err != nil {
		t.Fatalf("submitAttemptCommand(validation_started) error = %v", err)
	}
	effect, claimed, err := claimNextImprovementEffect(cfg, store, "tester", 30*time.Second, cfg.SandboxPollInterval)
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
	proposal, err := storepkg.ReviewProposalForTesting(store, base.ID, review.ProposalReview{
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
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, map[string]any{
		"trigger": string(improvement.AttemptTriggerProposalApproved),
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	proposal, attempt, repoJob := advanceAttemptToPatchGeneratedForTest(
		t,
		cfg,
		store,
		proposal,
		attempt,
		attemptTrace,
		"workspace-sandbox-launch-1",
		"job-sandbox-launch-1",
		"rsi-agent-platform",
		"main",
		"rsi-platform",
		"workspace-job-sandbox-launch-1",
		"workspace-pod-sandbox-launch-1",
		"Update the formal transition path.",
		"Run the sandbox validation flow.",
	)
	operationID := "op-sandbox-launch-1"
	session := sandbox.Session{
		ID:        "sandbox-session-1",
		Namespace: "rsi-platform",
		PodName:   "sandbox-pod-1",
		Status:    "running",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := applySandboxLaunchSuccess(cfg, store, proposal, attempt, sourceTrace, repoJob, sandboxObservationRequestForTest("work-sandbox-launch-1", operationID, attempt, sourceTrace, repoJob), session); err != nil {
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

	if receipt, ok := store.GetCommandReceipt("cmd-attempt:" + attempt.ID + ":validation_started:" + operationID); !ok || receipt.MachineKind != transition.MachineAttempt {
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
	proposal, err := storepkg.ReviewProposalForTesting(store, base.ID, review.ProposalReview{
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
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, map[string]any{
		"trigger": string(improvement.AttemptTriggerProposalApproved),
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	proposal, attempt, repoJob := advanceAttemptToPatchGeneratedForTest(
		t,
		cfg,
		store,
		proposal,
		attempt,
		attemptTrace,
		"workspace-watch-success-1",
		"job-watch-success-1",
		"rsi-agent-platform",
		"main",
		"rsi-platform",
		"sandbox-watch-success-1",
		"sandbox-watch-success-1-pod",
		"Update the formal transition path.",
		"Run the sandbox validation flow.",
	)
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
	if err := applySandboxLaunchSuccess(cfg, store, proposal, attempt, sourceTrace, repoJob, sandboxObservationRequestForTest("work-sandbox-launch-success", operationID, attempt, sourceTrace, repoJob), sandbox.Session{
		ID:        "sandbox-session-success",
		Namespace: "rsi-platform",
		PodName:   "sandbox-watch-success-1-pod",
		Status:    "running",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("applySandboxLaunchSuccess() error = %v", err)
	}
	effect, claimed, err := claimNextImprovementEffect(cfg, store, "tester", 30*time.Second, cfg.SandboxPollInterval)
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
	proposal, err := storepkg.ReviewProposalForTesting(store, base.ID, review.ProposalReview{
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
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, map[string]any{
		"trigger": string(improvement.AttemptTriggerProposalApproved),
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	proposal, attempt, repoJob := advanceAttemptToPatchGeneratedForTest(
		t,
		cfg,
		store,
		proposal,
		attempt,
		attemptTrace,
		"workspace-watch-fail-1",
		"job-watch-fail-1",
		"rsi-agent-platform",
		"main",
		"rsi-platform",
		"sandbox-watch-fail-1",
		"sandbox-watch-fail-1-pod",
		"Update the formal transition path.",
		"Run the sandbox validation flow.",
	)
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
	if err := applySandboxLaunchSuccess(cfg, store, proposal, attempt, sourceTrace, repoJob, sandboxObservationRequestForTest("work-sandbox-launch-fail", operationID, attempt, sourceTrace, repoJob), sandbox.Session{
		ID:        "sandbox-session-fail",
		Namespace: "rsi-platform",
		PodName:   "sandbox-watch-fail-1-pod",
		Status:    "running",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("applySandboxLaunchSuccess() error = %v", err)
	}
	effect, claimed, err := claimNextImprovementEffect(cfg, store, "tester", 30*time.Second, cfg.SandboxPollInterval)
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

func TestRecordAttemptFailureUsesFormalAttemptAndProposalCommands(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{ServiceName: "improvement-plane"}
	base := store.ListProposals()[0]
	trace, ok := store.GetTrace(base.TraceID)
	if !ok {
		t.Fatalf("expected trace %s", base.TraceID)
	}
	proposal, err := storepkg.ReviewProposalForTesting(store, base.ID, review.ProposalReview{
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
	proposal, attempt, _ = advanceAttemptToValidationRunningForTest(
		t,
		cfg,
		store,
		proposal,
		attempt,
		trace,
		"workspace-failure-1",
		"job-failure-1",
		"rsi-agent-platform",
		"main",
		"rsi-platform",
		"sandbox-failure-1",
		"sandbox-failure-1-pod",
		"Update the formal transition path.",
		"Run the sandbox validation flow.",
	)

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
	proposal, err := storepkg.ReviewProposalForTesting(store, base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	proposal.TargetRef = "prod"
	proposal.TargetKind = "role"
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, map[string]any{
		"trigger": string(improvement.AttemptTriggerProposalApproved),
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	proposal, attempt, _ = advanceAttemptToValidationPendingForTest(
		t,
		cfg,
		store,
		proposal,
		attempt,
		attemptTrace,
		"workspace-pr-open-1",
		"job-pr-open-1",
		"rsi-agent-platform",
		"main",
		"rsi-platform",
		"sandbox-pr-open-1",
		"sandbox-pr-open-1-pod",
		"Update the formal transition path.",
		"Run the sandbox validation flow.",
	)
	operationID := "op-pr-open-1"
	effectID := "work-pr-open-1"
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

	if err := processDraftPROpen(cfg, store, toolClient, draftPROpenRequestForTest(effectID, operationID, attempt, attemptTrace, "job-pr-open-1", "rsi-agent-platform", attempt.BranchName, "main", "Fix formal transition gap", "Draft PR body")); err != nil {
		t.Fatalf("processDraftPROpen() error = %v", err)
	}

	attempt, ok := store.GetChangeAttempt(attempt.ID)
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
			receiptID = improvementActionCommandID(actionID, kind, operationID)
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

	attemptReceiptID := "cmd-attempt:" + attempt.ID + ":" + string(transition.CommandAttemptPROpened) + ":" + operationID
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
	proposal, err := storepkg.ReviewProposalForTesting(store, base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	proposal.TargetRef = "prod"
	proposal.TargetKind = "role"
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, map[string]any{
		"trigger": string(improvement.AttemptTriggerProposalApproved),
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	proposal, attempt, _ = advanceAttemptToValidationPendingForTest(
		t,
		cfg,
		store,
		proposal,
		attempt,
		attemptTrace,
		"workspace-pr-open-fail-1",
		"job-pr-open-fail-1",
		"rsi-agent-platform",
		"main",
		"rsi-platform",
		"sandbox-pr-open-fail-1",
		"sandbox-pr-open-fail-1-pod",
		"Update the formal transition path.",
		"Run the sandbox validation flow.",
	)
	operationID := "op-pr-open-fail-1"
	effectID := "work-pr-open-fail-1"
	toolClient := fakeToolClient{
		results: map[string]storepkg.ToolResult{
			"github.create_pr": {
				Status:   "blocked",
				Summary:  "Draft PR open blocked.",
				Provider: "github",
			},
		},
	}

	if err := processDraftPROpen(cfg, store, toolClient, draftPROpenRequestForTest(effectID, operationID, attempt, attemptTrace, "", "rsi-agent-platform", attempt.BranchName, "main", "Fix formal transition gap", "Draft PR body")); err != nil {
		t.Fatalf("processDraftPROpen() error = %v", err)
	}

	attempt, ok := store.GetChangeAttempt(attempt.ID)
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
	base := seedHarnessOverlayProposalForTest(t, store)
	proposal, err := storepkg.ReviewProposalForTesting(store, base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, map[string]any{
		"trigger": string(improvement.AttemptTriggerProposalApproved),
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	if attempt.State != improvement.AttemptStateOverlayPlan {
		t.Fatalf("expected overlay_plan attempt state, got %s", attempt.State)
	}
	submitProposalStatusForTest(t, store, proposal.ID, review.ProposalRepoChangeRunning, "cmd-test-proposal-overlay-running")

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
	operationID := "op-overlay-1"
	if err := processHarnessOverlayProposal(cfg, store, attemptTrace, proposal, attempt, operationID, clients.RunnerResponse{OK: true}, runnerOutput, time.Now().UTC()); err != nil {
		t.Fatalf("processHarnessOverlayProposal() error = %v", err)
	}

	attempt, ok := store.GetChangeAttempt(attempt.ID)
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
			receiptID = improvementActionCommandID(actionID, kind, operationID)
		}
		if receipt, ok := store.GetCommandReceipt(receiptID); !ok || receipt.MachineKind != transition.MachineAction {
			t.Fatalf("expected harness overlay action receipt for %s, got ok=%t receipt=%+v", kind, ok, receipt)
		}
	}
	attemptReceiptID := "cmd-attempt:" + attempt.ID + ":" + string(transition.CommandOverlayActivated) + ":" + operationID
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
	approved, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
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
	approved, err := storepkg.ReviewProposalForTesting(store, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with governed remediation.",
		ReviewerID: "alice",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, approved, map[string]any{
		"trigger": string(improvement.AttemptTriggerProposalApproved),
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
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

func TestBuildProposalRunnerTaskPreservesRepoAndContextBudget(t *testing.T) {
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
	proposal, err := storepkg.ReviewProposalForTesting(store, base.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	attempt, attemptTrace, err := ensureProposalAttemptForTest(cfg, store, proposal, map[string]any{
		"trigger": string(improvement.AttemptTriggerProposalApproved),
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

func TestBuildEvalRunnerTaskPreservesContextRefs(t *testing.T) {
	store := storepkg.NewMemoryStore()
	proposals := store.ListProposals()
	if len(proposals) == 0 {
		t.Fatal("expected seeded proposal")
	}
	if _, err := storepkg.ReviewProposalForTesting(store, proposals[0].ID, review.ProposalReview{
		Decision:     string(review.ProposalRejected),
		Rationale:    "Seed proposal memory for eval context.",
		ReviewerID:   "tester",
		FailureClass: "needs_stronger_evidence",
	}); err != nil {
		t.Fatalf("ReviewProposalForTesting() error = %v", err)
	}
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
	}, store, trace, run, judgments, run.ID, run.Trigger, "")

	if task.TimeoutSeconds != 300 {
		t.Fatalf("eval timeout = %d, want 300", task.TimeoutSeconds)
	}
	assertContextRefKind(t, task.ContextRefs, "target_repo")
	assertContextRefKind(t, task.ContextRefs, "workflow_attempt")
	assertContextRefKind(t, task.ContextRefs, "proposal_memory")
}

func TestEnsureAttemptWorkspaceReturnsPersistenceError(t *testing.T) {
	base := storepkg.NewMemoryStore()
	cfg := config.Config{ServiceName: "improvement-plane"}
	proposal := base.ListProposals()[0]
	approved, err := storepkg.ReviewProposalForTesting(base, proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed.",
		ReviewerID: "tester",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	attempt, _, err := ensureProposalAttemptForTest(cfg, base, approved, map[string]any{
		"trigger": string(improvement.AttemptTriggerProposalApproved),
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	submitProposalStatusForTest(t, base, approved.ID, review.ProposalRepoChangeQueued, "cmd-test-proposal-queued-workspace-persist")
	if err := submitAttemptCommand(base, attempt, transition.CommandWorkspaceOpenDeferred, cfg.ServiceName, time.Now().UTC(), map[string]any{
		"workspace_id":        fmt.Sprintf("workspace-%s", attempt.ID),
		"job_id":              "job-1",
		"repo":                "rsi-agent-platform",
		"base_ref":            "main",
		"branch_name":         attempt.BranchName,
		"workspace_namespace": "rsi-platform",
		"workspace_job_name":  "workspace-job-1",
	}); err != nil {
		t.Fatalf("submitAttemptCommand(workspace_open_deferred) error = %v", err)
	}
	queuedWorkspace, ok := base.GetAttemptWorkspaceByAttempt(attempt.ID)
	if !ok {
		t.Fatalf("expected queued workspace for attempt %s", attempt.ID)
	}
	if queuedWorkspace.JobName == "" {
		t.Fatalf("expected queued workspace job name, got %+v", queuedWorkspace)
	}
	approved, attempt = loadProposalAttemptForTest(t, base, approved.ID, attempt.ID)
	workspace, ready, err := ensureAttemptWorkspace(cfg, base, workspaceLaunchStub{}, nil, approved, attempt, approved.TraceID)
	if err != nil {
		t.Fatalf("ensureAttemptWorkspace() error = %v", err)
	}
	if !ready {
		t.Fatal("expected resolved workspace to be ready")
	}
	if workspace.PodName != "workspace-pod-1" || workspace.Status != improvement.WorkspaceReady {
		t.Fatalf("unexpected workspace %+v", workspace)
	}
	stored, ok := base.GetAttemptWorkspaceByAttempt(attempt.ID)
	if !ok {
		t.Fatal("expected stored workspace")
	}
	if stored.PodName != "" || stored.Status != improvement.WorkspaceQueued {
		t.Fatalf("expected ensureAttemptWorkspace to avoid persisting directly, got %+v", stored)
	}
}

func assertContextRefKind(t *testing.T, refs []clients.RunnerContextRef, target string) {
	t.Helper()
	for _, item := range refs {
		if item.Kind == target {
			return
		}
	}
	t.Fatalf("expected context ref kind %q in %#v", target, refs)
}
