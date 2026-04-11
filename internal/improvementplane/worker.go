package improvementplane

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/runnerutil"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func RunWorker(cfg config.Config, store storepkg.Store) error {
	workerID := fmt.Sprintf("%s-worker", cfg.ServiceName)
	runnerClients := map[string]*clients.RunnerClient{
		"eval":     clients.NewRunnerClient(cfg.RunnerURLForRole("eval")),
		"proposal": clients.NewRunnerClient(cfg.RunnerURLForRole("proposal")),
	}
	toolClient := clients.NewToolGatewayClient(cfg.ToolGatewayBaseURL)
	launcher, launcherErr := sandbox.NewLauncher(cfg)
	for {
		item, ok, err := store.ClaimNextWorkItem([]queue.QueueName{queue.EvalQueue, queue.ProposalQueue, queue.SandboxQueue}, workerID, cfg.WorkItemLeaseDuration)
		if err != nil {
			return err
		}
		if !ok {
			time.Sleep(cfg.WorkerPollInterval)
			continue
		}
		if err := processImprovementItem(cfg, store, runnerClients, toolClient, launcher, launcherErr, item); err != nil {
			log.Printf("improvement-plane worker item=%s error=%v", item.ID, err)
			_, _ = store.FailWorkItem(item.ID, err.Error())
			continue
		}
		_, _ = store.CompleteWorkItem(item.ID)
	}
}

func processImprovementItem(cfg config.Config, store storepkg.Store, runnerClients map[string]*clients.RunnerClient, toolClient *clients.ToolGatewayClient, launcher sandbox.Launcher, launcherErr error, item queue.WorkItem) error {
	switch item.Queue {
	case queue.EvalQueue:
		return processEvalItem(cfg, store, runnerClients["eval"], item)
	case queue.ProposalQueue:
		return processProposalItem(cfg, store, runnerClients["proposal"], item)
	case queue.SandboxQueue:
		return processSandboxItem(cfg, store, toolClient, launcher, launcherErr, item)
	default:
		return fmt.Errorf("unsupported improvement work queue %s", item.Queue)
	}
}

func processEvalItem(cfg config.Config, store storepkg.Store, runnerClient *clients.RunnerClient, item queue.WorkItem) error {
	trace, ok := store.GetTrace(item.TraceID)
	if !ok {
		return fmt.Errorf("trace %s not found", item.TraceID)
	}
	started := time.Now().UTC()
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "improvement",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "eval.started",
				Status:      events.StatusRunning,
				StartedAt:   started,
				Description: fmt.Sprintf("Started eval item kind=%s.", item.Kind),
			},
		},
	})
	run, judgments, err := store.EvaluateTrace(item.TraceID, item.Kind)
	if err != nil {
		return err
	}
	runnerStarted := time.Now().UTC()
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "execution",
				Service:     "runner",
				Actor:       "eval",
				EventType:   "runner.started",
				Status:      events.StatusRunning,
				StartedAt:   runnerStarted,
				Description: "Eval summary dispatched to eval runner.",
			},
		},
	})
	var (
		runnerResp   clients.RunnerResponse
		runnerOutput runnerutil.StructuredOutput
		runnerErr    error
	)
	if runnerClient != nil {
		runnerResp, runnerErr = runnerClient.Execute(buildEvalRunnerTask(cfg, trace, run, judgments, item))
		if runnerErr == nil {
			runnerOutput = runnerutil.ParseStructuredOutput(runnerResp)
		}
	}
	completed := time.Now().UTC()
	reasoning := []events.ReasoningStep{
		{
			ID:         fmt.Sprintf("reason-eval-%d", completed.UnixNano()),
			TraceID:    trace.Summary.TraceID,
			WorkflowID: trace.Summary.WorkflowID,
			StepType:   "eval_summary",
			Summary:    firstNonEmpty(runnerOutput.FinalAnswer, runnerOutput.ContextSummary, fmt.Sprintf("Recorded %d judgments with overall score %.2f.", len(judgments), run.OverallScore)),
			Confidence: confidenceOr(run.OverallScore, runnerOutput.Confidence),
			Decision:   run.OverallVerdict,
			CreatedAt:  completed,
		},
	}
	if runnerErr == nil {
		reasoning = append(runnerutil.ToTraceReasoning(trace.Summary.TraceID, trace.Summary.WorkflowID, runnerOutput, completed), reasoning...)
		if strings.TrimSpace(runnerOutput.SelfCritique) != "" {
			reasoning = append(reasoning, events.ReasoningStep{
				ID:         fmt.Sprintf("reason-eval-self-%d", completed.UnixNano()),
				TraceID:    trace.Summary.TraceID,
				WorkflowID: trace.Summary.WorkflowID,
				StepType:   "self_critique",
				Summary:    runnerOutput.SelfCritique,
				Confidence: confidenceOr(run.OverallScore, runnerOutput.Confidence),
				CreatedAt:  completed,
			})
		}
	}
	runnerEvent := events.TraceEvent{
		TraceID:     trace.Summary.TraceID,
		IngestionID: trace.Summary.IngestionID,
		WorkflowID:  trace.Summary.WorkflowID,
		Plane:       "execution",
		Service:     "runner",
		Actor:       "eval",
		EventType:   "runner.completed",
		Status:      events.StatusCompleted,
		StartedAt:   runnerStarted,
		EndedAt:     ptrTime(completed),
		Description: fmt.Sprintf("Eval runner returned summary using %s.", runnerRuntimeLabel(runnerResp)),
	}
	if runnerErr != nil {
		runnerEvent.EventType = "runner.failed"
		runnerEvent.Status = events.StatusFailed
		runnerEvent.Description = fmt.Sprintf("Eval runner unavailable; kept deterministic results only: %v", runnerErr)
	}
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "improvement",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "eval.completed",
				Status:      events.StatusCompleted,
				StartedAt:   started,
				EndedAt:     ptrTime(completed),
				Description: fmt.Sprintf("Eval %s completed with verdict %s.", run.ID, run.OverallVerdict),
			},
			runnerEvent,
		},
		Reasoning: reasoning,
	})
	return nil
}

func processProposalItem(cfg config.Config, store storepkg.Store, runnerClient *clients.RunnerClient, item queue.WorkItem) error {
	if item.ProposalID == "" {
		return fmt.Errorf("proposal work item %s missing proposal_id", item.ID)
	}
	proposal, ok := findProposal(store.ListProposals(), item.ProposalID)
	if !ok {
		return fmt.Errorf("proposal %s not found", item.ProposalID)
	}
	proposalTraceID := item.TraceID
	if proposalTraceID == "" {
		proposalTraceID = proposal.TraceID
	}
	trace, ok := store.GetTrace(proposalTraceID)
	if !ok {
		return nil
	}
	memories := filterProposalMemory(store.ListProposalMemories(), proposal.CandidateKey)
	runnerStarted := time.Now().UTC()
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "execution",
				Service:     "runner",
				Actor:       "proposal",
				EventType:   "runner.started",
				Status:      events.StatusRunning,
				StartedAt:   runnerStarted,
				Description: "Proposal materialization dispatched to proposal runner.",
			},
		},
	})
	var (
		runnerResp   clients.RunnerResponse
		runnerOutput runnerutil.StructuredOutput
		runnerErr    error
	)
	if runnerClient != nil {
		runnerResp, runnerErr = runnerClient.Execute(buildProposalRunnerTask(cfg, trace, proposal, memories))
		if runnerErr == nil {
			runnerOutput = runnerutil.ParseStructuredOutput(runnerResp)
		}
	}
	job, err := store.MaterializeApprovedProposal(item.ProposalID, cfg.ServiceName)
	if err != nil {
		return err
	}
	manifest := sandbox.BuildJob(cfg, sandbox.JobRequest{
		TraceID:      trace.Summary.TraceID,
		ProposalID:   item.ProposalID,
		Repo:         job.Repo,
		BaseRef:      job.BaseRef,
		RequestedBy:  cfg.ServiceName,
		ArtifactPath: fmt.Sprintf("memory://sandbox/%s", job.ID),
		Commands: []string{
			"bash",
			"-lc",
			fmt.Sprintf("echo materialized proposal %s for repo %s", item.ProposalID, job.Repo),
		},
	})
	now := time.Now().UTC()
	_, _ = store.UpdateRepoChangeJobStatus(job.ID, string(review.ProposalRepoChangeQueued))
	reasoning := []events.ReasoningStep{
		{
			ID:         fmt.Sprintf("reason-proposal-%d", now.UnixNano()),
			TraceID:    trace.Summary.TraceID,
			WorkflowID: trace.Summary.WorkflowID,
			StepType:   "proposal_materialized",
			Summary:    firstNonEmpty(runnerOutput.FinalAnswer, runnerOutput.ContextSummary, fmt.Sprintf("Approved proposal %s moved into repo-change queue.", item.ProposalID)),
			Confidence: confidenceOr(0.86, runnerOutput.Confidence),
			Decision:   job.BranchName,
			CreatedAt:  now,
		},
	}
	if runnerErr == nil {
		reasoning = append(runnerutil.ToTraceReasoning(trace.Summary.TraceID, trace.Summary.WorkflowID, runnerOutput, now), reasoning...)
		if strings.TrimSpace(runnerOutput.SelfCritique) != "" {
			reasoning = append(reasoning, events.ReasoningStep{
				ID:         fmt.Sprintf("reason-proposal-self-%d", now.UnixNano()),
				TraceID:    trace.Summary.TraceID,
				WorkflowID: trace.Summary.WorkflowID,
				StepType:   "self_critique",
				Summary:    runnerOutput.SelfCritique,
				Confidence: confidenceOr(0.86, runnerOutput.Confidence),
				CreatedAt:  now,
			})
		}
	}
	runnerEvent := events.TraceEvent{
		TraceID:     trace.Summary.TraceID,
		IngestionID: trace.Summary.IngestionID,
		WorkflowID:  trace.Summary.WorkflowID,
		Plane:       "execution",
		Service:     "runner",
		Actor:       "proposal",
		EventType:   "runner.completed",
		Status:      events.StatusCompleted,
		StartedAt:   runnerStarted,
		EndedAt:     ptrTime(now),
		Description: fmt.Sprintf("Proposal runner returned repo-change rationale using %s.", runnerRuntimeLabel(runnerResp)),
	}
	if runnerErr != nil {
		runnerEvent.EventType = "runner.failed"
		runnerEvent.Status = events.StatusFailed
		runnerEvent.Description = fmt.Sprintf("Proposal runner unavailable; using deterministic repo-change materialization only: %v", runnerErr)
	}
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			runnerEvent,
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "improvement",
				Service:     cfg.ServiceName,
				Actor:       "worker",
				EventType:   "repo_change.queued",
				Status:      events.StatusQueued,
				StartedAt:   now,
				Description: fmt.Sprintf("Materialized repo change job %s.", job.ID),
			},
		},
		Artifacts: []events.Artifact{
			{
				ID:          fmt.Sprintf("artifact-sandbox-%d", now.UnixNano()),
				TraceID:     trace.Summary.TraceID,
				Kind:        "sandbox_job_manifest",
				ContentType: "application/json",
				URL:         fmt.Sprintf("memory://sandbox/%s/manifest.json", job.ID),
				SizeBytes:   int64(len(fmt.Sprintf("%v", manifest))),
				Source:      "improvement-plane",
			},
		},
		Reasoning: reasoning,
	})
	return nil
}

func processSandboxItem(cfg config.Config, store storepkg.Store, toolClient *clients.ToolGatewayClient, launcher sandbox.Launcher, launcherErr error, item queue.WorkItem) error {
	switch item.Kind {
	case "repo_change_job":
		return processSandboxLaunch(cfg, store, launcher, launcherErr, item)
	case "watch_sandbox_job":
		return processSandboxWatch(cfg, store, toolClient, launcher, launcherErr, item)
	default:
		return fmt.Errorf("unsupported sandbox item kind %s", item.Kind)
	}
}

func processSandboxLaunch(cfg config.Config, store storepkg.Store, launcher sandbox.Launcher, launcherErr error, item queue.WorkItem) error {
	if launcherErr != nil {
		return launcherErr
	}
	if launcher == nil {
		return fmt.Errorf("sandbox launcher not configured")
	}
	jobID := stringValue(item.Payload["job_id"])
	if jobID == "" {
		return fmt.Errorf("sandbox work item missing job_id")
	}
	repoJob, ok := findRepoChangeJob(store.ListRepoChangeJobs(), jobID)
	if !ok {
		return fmt.Errorf("repo change job %s not found", jobID)
	}
	trace, ok := store.GetTrace(item.TraceID)
	if !ok {
		return fmt.Errorf("trace %s not found", item.TraceID)
	}
	request := sandbox.JobRequest{
		TraceID:      trace.Summary.TraceID,
		ProposalID:   item.ProposalID,
		Repo:         repoJob.Repo,
		BaseRef:      repoJob.BaseRef,
		RequestedBy:  cfg.ServiceName,
		ArtifactPath: fmt.Sprintf("memory://sandbox/%s", repoJob.ID),
		Env: map[string]string{
			"GITHUB_TOKEN":        cfg.GitHubToken,
			"GITHUB_OWNER":        cfg.GitHubOwner,
			"GITHUB_COMMIT_USER":  cfg.GitHubCommitUser,
			"GITHUB_COMMIT_EMAIL": cfg.GitHubCommitEmail,
			"RSI_BRANCH_NAME":     repoJob.BranchName,
			"RSI_CONTEXT_SUMMARY": repoJob.ContextSummary,
			"RSI_REPO":            repoJob.Repo,
			"RSI_BASE_REF":        repoJob.BaseRef,
			"RSI_PROPOSAL_ID":     item.ProposalID,
		},
		Commands: repoChangeCommands(item.ProposalID),
	}
	session, _, err := launcher.Launch(context.Background(), request)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	_, _ = store.UpdateRepoChangeJobStatus(repoJob.ID, string(review.ProposalRepoChangeRunning))
	_, _ = store.UpdateProposalStatus(item.ProposalID, review.ProposalRepoChangeRunning)
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "execution",
				Service:     cfg.ServiceName,
				Actor:       "sandbox-launcher",
				EventType:   "sandbox.job.started",
				Status:      events.StatusRunning,
				StartedAt:   now,
				Description: fmt.Sprintf("Launched sandbox job %s in namespace %s.", session.PodName, session.Namespace),
			},
		},
		Artifacts: []events.Artifact{
			{
				ID:          fmt.Sprintf("artifact-sandbox-launch-%d", now.UnixNano()),
				TraceID:     trace.Summary.TraceID,
				Kind:        "sandbox_job",
				ContentType: "text/plain",
				URL:         fmt.Sprintf("k8s://%s/jobs/%s", session.Namespace, session.PodName),
				SizeBytes:   0,
				Source:      "sandbox-runtime",
			},
		},
		Reasoning: []events.ReasoningStep{
			{
				ID:         fmt.Sprintf("reason-sandbox-launch-%d", now.UnixNano()),
				TraceID:    trace.Summary.TraceID,
				WorkflowID: trace.Summary.WorkflowID,
				StepType:   "sandbox_launch",
				Summary:    fmt.Sprintf("Launched real sandbox job for repo %s branch %s.", repoJob.Repo, repoJob.BranchName),
				Confidence: 0.88,
				Decision:   session.PodName,
				CreatedAt:  now,
			},
		},
	})
	_, err = store.EnqueueWorkItem(queue.WorkItem{
		Queue:      queue.SandboxQueue,
		Kind:       "watch_sandbox_job",
		Status:     queue.WorkQueued,
		TraceID:    trace.Summary.TraceID,
		ProposalID: item.ProposalID,
		Payload: map[string]interface{}{
			"job_name":    session.PodName,
			"namespace":   session.Namespace,
			"repo":        repoJob.Repo,
			"branch_name": repoJob.BranchName,
			"job_id":      repoJob.ID,
		},
		CreatedAt: now,
		UpdatedAt: now,
	})
	return err
}

func processSandboxWatch(cfg config.Config, store storepkg.Store, toolClient *clients.ToolGatewayClient, launcher sandbox.Launcher, launcherErr error, item queue.WorkItem) error {
	if launcherErr != nil {
		return launcherErr
	}
	if launcher == nil {
		return fmt.Errorf("sandbox launcher not configured")
	}
	jobName := stringValue(item.Payload["job_name"])
	namespace := stringValue(item.Payload["namespace"])
	repo := stringValue(item.Payload["repo"])
	branchName := stringValue(item.Payload["branch_name"])
	jobID := stringValue(item.Payload["job_id"])
	if jobName == "" || namespace == "" {
		return fmt.Errorf("sandbox watch item missing job metadata")
	}
	job, err := launcher.GetJob(context.Background(), namespace, jobName)
	if err != nil {
		return err
	}
	trace, ok := store.GetTrace(item.TraceID)
	if !ok {
		return fmt.Errorf("trace %s not found", item.TraceID)
	}
	if job.Status.Succeeded == 0 && job.Status.Failed == 0 {
		time.Sleep(cfg.SandboxPollInterval)
		_, err = store.EnqueueWorkItem(queue.WorkItem{
			Queue:      queue.SandboxQueue,
			Kind:       "watch_sandbox_job",
			Status:     queue.WorkQueued,
			TraceID:    item.TraceID,
			ProposalID: item.ProposalID,
			Payload:    item.Payload,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		})
		return err
	}
	now := time.Now().UTC()
	if job.Status.Failed > 0 {
		_, _ = store.UpdateRepoChangeJobStatus(jobID, string(review.ProposalFailedValidation))
		_, _ = store.UpdateProposalStatus(item.ProposalID, review.ProposalFailedValidation)
		_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
			Status: ptrStatus(events.StatusFailed),
			Events: []events.TraceEvent{
				{
					TraceID:     trace.Summary.TraceID,
					IngestionID: trace.Summary.IngestionID,
					WorkflowID:  trace.Summary.WorkflowID,
					Plane:       "execution",
					Service:     cfg.ServiceName,
					Actor:       "sandbox-launcher",
					EventType:   "sandbox.job.failed",
					Status:      events.StatusFailed,
					StartedAt:   now,
					Description: fmt.Sprintf("Sandbox job %s failed.", jobName),
				},
			},
		})
		return nil
	}

	_, _ = store.UpdateRepoChangeJobStatus(jobID, string(review.ProposalValidationPending))
	_, _ = store.UpdateProposalStatus(item.ProposalID, review.ProposalValidationPending)
	prResult, err := toolClient.Execute("github.create_pr", map[string]any{
		"proposal_id": item.ProposalID,
		"repo":        repo,
		"branch_name": branchName,
		"base_ref":    "main",
		"title":       fmt.Sprintf("RSI proposal %s for %s", item.ProposalID, repo),
		"body":        fmt.Sprintf("Automated draft PR for proposal %s after sandbox validation.", item.ProposalID),
	})
	if err != nil {
		return err
	}
	prURL := stringValue(prResult.Output["pr_url"])
	if _, err := store.RecordPRAttempt(buildPRAttempt(item.ProposalID, repo, branchName, prURL)); err != nil {
		return err
	}
	_, _ = store.UpdateRepoChangeJobStatus(jobID, string(review.ProposalPROpen))
	_, _ = store.UpdateProposalStatus(item.ProposalID, review.ProposalPROpen)
	_, _ = store.ApplyTraceUpdate(trace.Summary.TraceID, storepkg.TraceUpdate{
		Events: []events.TraceEvent{
			{
				TraceID:     trace.Summary.TraceID,
				IngestionID: trace.Summary.IngestionID,
				WorkflowID:  trace.Summary.WorkflowID,
				Plane:       "execution",
				Service:     cfg.ServiceName,
				Actor:       "sandbox-launcher",
				EventType:   "sandbox.job.succeeded",
				Status:      events.StatusCompleted,
				StartedAt:   now,
				Description: fmt.Sprintf("Sandbox job %s succeeded and draft PR opened.", jobName),
			},
		},
		Reasoning: []events.ReasoningStep{
			{
				ID:         fmt.Sprintf("reason-pr-open-%d", now.UnixNano()),
				TraceID:    trace.Summary.TraceID,
				WorkflowID: trace.Summary.WorkflowID,
				StepType:   "pr_opened",
				Summary:    fmt.Sprintf("Opened real draft PR for branch %s.", branchName),
				Confidence: 0.9,
				Decision:   prURL,
				CreatedAt:  now,
			},
		},
	})
	return nil
}

func buildPRAttempt(proposalID string, repo string, branchName string, prURL string) improvement.PRAttempt {
	return improvement.PRAttempt{
		ProposalID:       proposalID,
		Repo:             repo,
		BranchName:       branchName,
		PRURL:            prURL,
		Status:           string(review.ProposalPROpen),
		ValidationStatus: "pending",
	}
}

func findRepoChangeJob(items []improvement.RepoChangeJob, jobID string) (improvement.RepoChangeJob, bool) {
	for _, item := range items {
		if item.ID == jobID {
			return item, true
		}
	}
	return improvement.RepoChangeJob{}, false
}

func findProposal(items []review.Proposal, proposalID string) (review.Proposal, bool) {
	for _, item := range items {
		if item.ID == proposalID {
			return item, true
		}
	}
	return review.Proposal{}, false
}

func filterProposalMemory(items []review.ProposalMemory, candidateKey string) []review.ProposalMemory {
	out := make([]review.ProposalMemory, 0)
	for _, item := range items {
		if item.CandidateKey == candidateKey {
			out = append(out, item)
		}
	}
	return out
}

func buildEvalRunnerTask(cfg config.Config, trace events.Trace, run evals.Run, judgments []evals.Judgment, item queue.WorkItem) clients.RunnerTask {
	contextRefs := make([]map[string]any, 0, len(judgments)+1)
	contextRefs = append(contextRefs, map[string]any{
		"kind":      "eval_run",
		"ref":       run.ID,
		"summary":   run.OverallVerdict,
		"trace_id":  trace.Summary.TraceID,
		"suite":     run.SuiteName,
		"trigger":   run.Trigger,
		"work_kind": item.Kind,
	})
	for _, judgment := range judgments {
		contextRefs = append(contextRefs, map[string]any{
			"kind":     "eval_judgment",
			"ref":      judgment.ID,
			"layer":    judgment.Layer,
			"category": judgment.Category,
			"score":    judgment.Score,
			"summary":  judgment.Rationale,
		})
	}
	prompt := fmt.Sprintf(
		"Summarize the completed eval for trace %s. Workflow=%s status=%s thread=%s. Eval run=%s suite=%s verdict=%s score=%.2f. Judgments=%v. Explain what the eval found, what evidence mattered, and whether the trace should create or strengthen improvement pressure.",
		trace.Summary.TraceID,
		trace.Summary.WorkflowKind,
		trace.Summary.Status,
		trace.Summary.ThreadKey,
		run.ID,
		run.SuiteName,
		run.OverallVerdict,
		run.OverallScore,
		judgmentDigest(judgments),
	)
	return clients.RunnerTask{
		TaskType:            "eval",
		Repo:                cfg.DefaultRepo,
		RepoRef:             "main",
		Prompt:              prompt,
		SystemMessage:       "Return explicit visible reasoning only. Do not include hidden chain-of-thought. Produce a JSON object with visible_reasoning, reply_draft, final_answer, confidence, context_summary, and self_critique.",
		TimeoutSeconds:      120,
		ExpectedOutputs:     []string{"visible_reasoning", "final_answer"},
		ArtifactDestination: fmt.Sprintf("trace:%s:eval:%s", trace.Summary.TraceID, run.ID),
		ContextSummary: fmt.Sprintf(
			"Eval %s completed with verdict=%s score=%.2f across %d judgments.",
			run.ID,
			run.OverallVerdict,
			run.OverallScore,
			len(judgments),
		),
		Intent:             trace.Summary.WorkflowKind,
		TraceID:            trace.Summary.TraceID,
		WorkflowID:         trace.Summary.WorkflowID,
		RepoAllowlist:      cfg.AllowedTargetRepos,
		ResponseMode:       "analysis",
		ContextRefs:        contextRefs,
		ApprovalMode:       "deterministic",
		ReasoningVerbosity: cfg.DefaultReasoningVerbosity,
	}
}

func buildProposalRunnerTask(cfg config.Config, trace events.Trace, proposal review.Proposal, memories []review.ProposalMemory) clients.RunnerTask {
	rejectedContext := make([]map[string]any, 0, len(memories))
	for _, memory := range memories {
		rejectedContext = append(rejectedContext, map[string]any{
			"proposal_id":   memory.ProposalID,
			"disposition":   memory.Disposition,
			"rationale":     memory.ReviewRationale,
			"failure_class": firstNonEmpty(memory.FailureClass, strings.Join(memory.FailureClasses, ",")),
		})
	}
	contextRefs := []map[string]any{
		{
			"kind":          "proposal",
			"ref":           proposal.ID,
			"summary":       proposal.Summary,
			"candidate_key": proposal.CandidateKey,
			"risk_tier":     proposal.RiskTier,
			"scope":         proposal.ProposedScope,
		},
	}
	prompt := fmt.Sprintf(
		"Summarize why approved proposal %s should move into repo-change work. Candidate=%s risk=%s scope=%s summary=%s. Explain what makes this change justified now, what prior rejections or dismissals matter, and what a reviewer should inspect before the PR path continues.",
		proposal.ID,
		proposal.CandidateKey,
		proposal.RiskTier,
		proposal.ProposedScope,
		proposal.Summary,
	)
	return clients.RunnerTask{
		TaskType:                "proposal",
		Repo:                    cfg.DefaultRepo,
		RepoRef:                 "main",
		Prompt:                  prompt,
		SystemMessage:           "Return explicit visible reasoning only. Do not include hidden chain-of-thought. Produce a JSON object with visible_reasoning, reply_draft, final_answer, confidence, context_summary, and self_critique.",
		TimeoutSeconds:          120,
		ExpectedOutputs:         []string{"visible_reasoning", "final_answer"},
		ArtifactDestination:     fmt.Sprintf("trace:%s:proposal:%s", trace.Summary.TraceID, proposal.ID),
		ContextSummary:          fmt.Sprintf("Approved proposal %s for candidate %s is entering repo-change queue.", proposal.ID, proposal.CandidateKey),
		RejectedProposalContext: rejectedContext,
		Intent:                  trace.Summary.WorkflowKind,
		TraceID:                 trace.Summary.TraceID,
		WorkflowID:              trace.Summary.WorkflowID,
		RepoAllowlist:           cfg.AllowedTargetRepos,
		ResponseMode:            "analysis",
		ContextRefs:             contextRefs,
		ApprovalMode:            "human_review",
		ReasoningVerbosity:      cfg.DefaultReasoningVerbosity,
	}
}

func judgmentDigest(items []evals.Judgment) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, fmt.Sprintf("%s/%s=%.2f (%s)", item.Layer, item.Category, item.Score, item.Rationale))
	}
	return out
}

func repoChangeCommands(proposalID string) []string {
	script := `
set -euo pipefail
mkdir -p /workspace
cd /workspace
rm -rf repo
git clone "https://x-access-token:${GITHUB_TOKEN}@github.com/${GITHUB_OWNER}/${RSI_REPO}.git" repo
cd repo
git checkout -B "${RSI_BRANCH_NAME}" "origin/${RSI_BASE_REF}"
mkdir -p .rsi
printf "%s\n" "${RSI_CONTEXT_SUMMARY}" > .rsi/proposal-context.txt
git config user.name "${GITHUB_COMMIT_USER}"
git config user.email "${GITHUB_COMMIT_EMAIL}"
git add .rsi/proposal-context.txt
git commit -m "chore: seed RSI proposal context for ` + proposalID + `" || true
git push origin HEAD
`
	return []string{"bash", "-lc", script}
}

func stringValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprintf("%v", value)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func confidenceOr(fallback float64, candidate float64) float64 {
	if candidate > 0 {
		return candidate
	}
	return fallback
}

func runnerRuntimeLabel(resp clients.RunnerResponse) string {
	if resp.Provider == "" {
		return "unknown runtime"
	}
	model := strings.TrimSpace(stringValue(resp.Raw["model"]))
	effort := strings.TrimSpace(stringValue(resp.Raw["reasoning_effort"]))
	if model == "" {
		return resp.Provider
	}
	if effort == "" {
		return fmt.Sprintf("%s %s", resp.Provider, model)
	}
	return fmt.Sprintf("%s %s effort=%s", resp.Provider, model, effort)
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func ptrStatus(status events.Status) *events.Status {
	return &status
}
