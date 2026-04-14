package improvementplane

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
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
	}, store, trace, proposal, attempt, store.ListProposalMemories())

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

func assertContains(t *testing.T, values []string, target string) {
	t.Helper()
	for _, item := range values {
		if item == target {
			return
		}
	}
	t.Fatalf("expected %q in %v", target, values)
}
