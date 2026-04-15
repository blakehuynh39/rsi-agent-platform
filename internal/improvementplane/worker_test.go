package improvementplane

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/operation"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	"github.com/piplabs/rsi-agent-platform/internal/slack"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	batchv1 "k8s.io/api/batch/v1"
)

type fakeLauncher struct {
	observation sandbox.JobObservation
	err         error
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

type workspaceUpsertFailStore struct {
	storepkg.Store
	err error
}

func (s workspaceUpsertFailStore) UpsertAttemptWorkspace(improvement.AttemptWorkspace) (improvement.AttemptWorkspace, error) {
	return improvement.AttemptWorkspace{}, s.err
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

func TestProcessSandboxWatchReschedulesCurrentLeaseUntilTerminal(t *testing.T) {
	store := storepkg.NewMemoryStore()
	now := time.Now().UTC()
	if _, err := store.CreateIngestion(slack.SlackEnvelope{
		BotRole:   slack.BotOrchestrator,
		TeamID:    "T-stage",
		ChannelID: "D-stage",
		UserID:    "U-stage",
		Text:      "Investigate recursive proposal sandbox progress.",
		ThreadTS:  "1711000000.000100",
		TS:        "1711000000.000100",
		CreatedAt: now,
	}); err != nil {
		t.Fatalf("CreateIngestion() error = %v", err)
	}
	trace := store.ListTraces()[0]
	if _, err := store.UpsertRepoChangeJob(improvement.RepoChangeJob{
		ID:               "job-watch-1",
		ProposalID:       "proposal-watch-1",
		ConversationID:   trace.ConversationID,
		CaseID:           trace.CaseID,
		OriginTraceID:    trace.TraceID,
		CandidateKey:     "shared-store:policy_or_runtime_fix:action_result_primary_key_collision",
		Status:           string(review.ProposalRepoChangeRunning),
		Repo:             "rsi-agent-platform",
		BaseRef:          "main",
		BranchName:       "codex/proposal-watch-1",
		ContextSummary:   "Validate sandbox watch rescheduling.",
		SandboxNamespace: "rsi-platform",
		SandboxJobName:   "sandbox-watch-1",
		SandboxPodName:   "sandbox-watch-1-pod",
		CreatedAt:        now,
		UpdatedAt:        now,
	}); err != nil {
		t.Fatalf("UpsertRepoChangeJob() error = %v", err)
	}
	queued, err := store.EnqueueWorkItem(queue.WorkItem{
		Queue:      queue.SandboxQueue,
		Kind:       "watch_sandbox_job",
		Status:     queue.WorkQueued,
		TraceID:    trace.TraceID,
		ProposalID: "proposal-watch-1",
		Payload: map[string]any{
			"job_id":      "job-watch-1",
			"job_name":    "sandbox-watch-1",
			"namespace":   "rsi-platform",
			"repo":        "rsi-agent-platform",
			"branch_name": "codex/proposal-watch-1",
			"base_ref":    "main",
			"dedupe_key":  "sandbox-watch:job-watch-1",
		},
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("EnqueueWorkItem() error = %v", err)
	}
	claimed, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.SandboxQueue}, "tester", 30*time.Second)
	if err != nil {
		t.Fatalf("ClaimNextWorkItem() error = %v", err)
	}
	if !ok {
		t.Fatal("expected claimed watch item")
	}
	cfg := config.Config{
		ServiceName:         "improvement-plane",
		SandboxPollInterval: 2 * time.Second,
	}
	err = processSandboxWatch(cfg, store, fakeLauncher{
		observation: sandbox.JobObservation{
			Namespace: "rsi-platform",
			JobName:   "sandbox-watch-1",
			PodName:   "sandbox-watch-1-pod",
		},
	}, nil, claimed)
	if !errors.Is(err, errDeferredWorkItem) {
		t.Fatalf("processSandboxWatch() error = %v, want errDeferredWorkItem", err)
	}
	var (
		found       queue.WorkItem
		foundCount  int
		retryAfter  int64
		retryExists bool
	)
	for _, item := range store.ListWorkItems() {
		if item.ID != queued.ID {
			continue
		}
		found = item
		foundCount++
		if raw, ok := item.Payload["retry_after_unix"].(int64); ok {
			retryAfter = raw
			retryExists = true
		}
		if raw, ok := item.Payload["retry_after_unix"].(float64); ok {
			retryAfter = int64(raw)
			retryExists = true
		}
	}
	if foundCount != 1 {
		t.Fatalf("expected exactly one rescheduled watch item, found %d", foundCount)
	}
	if found.Status != queue.WorkQueued {
		t.Fatalf("watch item status = %s, want %s", found.Status, queue.WorkQueued)
	}
	if found.LeaseOwner != "" || found.LeaseExpiresAt != nil {
		t.Fatalf("expected rescheduled watch item lease cleared, got owner=%q lease=%v", found.LeaseOwner, found.LeaseExpiresAt)
	}
	if found.CompletedAt != nil {
		t.Fatalf("expected rescheduled watch item to remain incomplete, got completed_at=%v", found.CompletedAt)
	}
	if !retryExists {
		t.Fatal("expected retry_after_unix payload on rescheduled watch item")
	}
	if retryAfter <= now.Unix() {
		t.Fatalf("retry_after_unix = %d, want value after %d", retryAfter, now.Unix())
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
	var queued queue.WorkItem
	foundQueued := false
	for _, item := range store.ListWorkItems() {
		if item.Queue == queue.ProposalQueue && item.Kind == "approved_proposal" && item.ProposalID == proposal.ID && item.Status == queue.WorkQueued {
			queued = item
			foundQueued = true
			break
		}
	}
	if !foundQueued {
		t.Fatalf("expected queued proposal work item for %s", proposal.ID)
	}
	if _, err := store.FailWorkItem(queued.ID, "simulated old runtime failure"); err != nil {
		t.Fatalf("FailWorkItem() error = %v", err)
	}
	cfg := config.Config{ServiceName: "improvement-plane"}
	evalItem := queue.WorkItem{
		ID:        "eval-manual",
		Queue:     queue.EvalQueue,
		Kind:      "evaluate_trace",
		Status:    queue.WorkQueued,
		TraceID:   trace.Summary.TraceID,
		CreatedAt: now.Add(time.Second),
		UpdatedAt: now.Add(time.Second),
	}
	if err := processEvalItem(cfg, store, nil, evalItem); err != nil {
		t.Fatalf("processEvalItem() error = %v", err)
	}
	foundQueued = false
	for _, item := range store.ListWorkItems() {
		if item.Queue == queue.ProposalQueue && item.Kind == "approved_proposal" && item.ProposalID == proposal.ID && item.Status == queue.WorkQueued {
			foundQueued = true
		}
	}
	if !foundQueued {
		t.Fatal("expected stalled approved proposal to be requeued after eval")
	}
}

func TestProcessProposalItemQueuesExplicitPhases(t *testing.T) {
	store := storepkg.NewMemoryStore()
	cfg := config.Config{ServiceName: "improvement-plane"}
	proposal := store.ListProposals()[0]
	approved, err := store.ReviewProposal(proposal.ID, review.ProposalReview{
		Decision:   string(review.ProposalApproved),
		Rationale:  "Proceed with governed remediation.",
		ReviewerID: "alice",
	})
	if err != nil {
		t.Fatalf("ReviewProposal() error = %v", err)
	}
	lineItem, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.ProposalQueue}, "tester", 30*time.Second)
	if err != nil || !ok {
		t.Fatalf("ClaimNextWorkItem() ok=%t err=%v", ok, err)
	}
	if err := processProposalItem(cfg, store, nil, nil, nil, nil, lineItem); err != nil {
		t.Fatalf("processProposalItem(line_activate) error = %v", err)
	}
	if _, err := store.CompleteWorkItem(lineItem.ID); err != nil {
		t.Fatalf("CompleteWorkItem(line_activate) error = %v", err)
	}
	approved, ok = findProposal(store.ListProposals(), approved.ID)
	if !ok || approved.CurrentAttemptID == "" {
		t.Fatalf("expected current attempt after line activation, got %+v", approved)
	}
	attemptPlan, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.ProposalQueue}, "tester", 30*time.Second)
	if err != nil || !ok {
		t.Fatalf("ClaimNextWorkItem(attempt_plan) ok=%t err=%v", ok, err)
	}
	if attemptPlan.Kind != proposalOperationAttemptPlan {
		t.Fatalf("attempt plan kind = %q, want %q", attemptPlan.Kind, proposalOperationAttemptPlan)
	}
	if err := processProposalItem(cfg, store, nil, nil, nil, nil, attemptPlan); err != nil {
		t.Fatalf("processProposalItem(attempt_plan) error = %v", err)
	}
	if _, err := store.CompleteWorkItem(attemptPlan.ID); err != nil {
		t.Fatalf("CompleteWorkItem(attempt_plan) error = %v", err)
	}
	foundWorkspaceOpen := false
	for _, item := range store.ListWorkItems() {
		if item.Queue == queue.ProposalQueue && item.Kind == proposalOperationWorkspaceOpen && item.ProposalID == approved.ID && item.Status == queue.WorkQueued {
			foundWorkspaceOpen = true
		}
	}
	if !foundWorkspaceOpen {
		t.Fatal("expected workspace_open phase to be queued")
	}
}

func TestProcessProposalImplementAttemptRequiresWorkspaceMutation(t *testing.T) {
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
	attempt, attemptTrace, err := ensureProposalAttempt(cfg, store, approved, sourceTrace, queue.WorkItem{
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
	op, _, err := store.GetOrCreateOperation(operation.Execution{
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
	})
	if err != nil {
		t.Fatalf("GetOrCreateOperation() error = %v", err)
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
	item := queue.WorkItem{
		ID:          "work-implement-1",
		OperationID: op.ID,
		Queue:       queue.ProposalQueue,
		Kind:        proposalOperationImplementAttempt,
		Status:      queue.WorkQueued,
		TraceID:     attemptTrace.Summary.TraceID,
		ProposalID:  approved.ID,
		Payload: map[string]any{
			"attempt_id":   attempt.ID,
			"workspace_id": "workspace-1",
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := processProposalItem(cfg, store, runner, fakeToolClient{}, nil, nil, item); err != nil {
		t.Fatalf("processProposalItem(implement_attempt) error = %v", err)
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
	attempt, attemptTrace, err := ensureProposalAttempt(cfg, store, approved, sourceTrace, queue.WorkItem{
		ProposalID: approved.ID,
		Payload:    map[string]any{"trigger": string(improvement.AttemptTriggerProposalApproved)},
	})
	if err != nil {
		t.Fatalf("ensureProposalAttempt() error = %v", err)
	}
	_, _ = store.ApplyTraceUpdate(attemptTrace.Summary.TraceID, storepkg.TraceUpdate{
		ToolCalls: []events.ToolCallRecord{
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
	})
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
	op, _, err := store.GetOrCreateOperation(operation.Execution{
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
	})
	if err != nil {
		t.Fatalf("GetOrCreateOperation() error = %v", err)
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
	item := queue.WorkItem{
		ID:          "work-implement-2",
		OperationID: op.ID,
		Queue:       queue.ProposalQueue,
		Kind:        proposalOperationImplementAttempt,
		Status:      queue.WorkQueued,
		TraceID:     attemptTrace.Summary.TraceID,
		ProposalID:  approved.ID,
		Payload: map[string]any{
			"attempt_id":   attempt.ID,
			"workspace_id": "workspace-2",
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := processProposalItem(cfg, store, runner, toolClient, nil, nil, item); err != nil {
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

func TestBuildEvalRunnerTaskUsesReadOnlyToolBudget(t *testing.T) {
	store := storepkg.NewMemoryStore()
	traceID := store.ListTraces()[0].TraceID
	run, judgments, err := store.EvaluateTrace(traceID, "incident-response")
	if err != nil {
		t.Fatalf("EvaluateTrace() error = %v", err)
	}
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
	wantErr := errors.New("workspace upsert failed")
	_, _, err := ensureAttemptWorkspace(cfg, workspaceUpsertFailStore{Store: base, err: wantErr}, workspaceLaunchStub{}, nil, proposal, attempt, "trace-1")
	if err == nil || !strings.Contains(err.Error(), wantErr.Error()) {
		t.Fatalf("ensureAttemptWorkspace() error = %v, want %v", err, wantErr)
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
