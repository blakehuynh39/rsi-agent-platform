package improvementplane

import (
	"context"
	"errors"
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
