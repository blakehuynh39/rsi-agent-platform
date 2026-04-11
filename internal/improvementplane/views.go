package improvementplane

import (
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type traceEvalSummary struct {
	RunID     string    `json:"run_id"`
	Verdict   string    `json:"verdict"`
	Score     float64   `json:"score"`
	CreatedAt time.Time `json:"created_at"`
	SuiteName string    `json:"suite_name"`
}

type traceListItem struct {
	TraceID          string            `json:"trace_id"`
	WorkflowID       string            `json:"workflow_id"`
	IngestionID      string            `json:"ingestion_id"`
	WorkflowKind     string            `json:"workflow_kind"`
	Status           events.Status     `json:"status"`
	ThreadKey        string            `json:"thread_key"`
	StartedAt        time.Time         `json:"started_at"`
	EventCount       int               `json:"event_count"`
	ReasoningCount   int               `json:"reasoning_count"`
	ToolCallCount    int               `json:"tool_call_count"`
	SlackActionCount int               `json:"slack_action_count"`
	LatestEval       *traceEvalSummary `json:"latest_eval,omitempty"`
}

type traceDetailResponse struct {
	Trace              events.Trace                `json:"trace"`
	LinkedEvalRuns     []evals.Run                 `json:"linked_eval_runs"`
	JudgmentsByEvalRun map[string][]evals.Judgment `json:"judgments_by_eval_run"`
	LinkedProposals    []review.Proposal           `json:"linked_proposals"`
	Ratings            []review.HumanRating        `json:"ratings"`
	ImprovementNotes   []review.ImprovementNote    `json:"improvement_notes"`
}

type proposalDetailResponse struct {
	Proposal              review.Proposal               `json:"proposal"`
	Reviews               []review.ProposalReview       `json:"reviews"`
	RelatedProposalMemory []review.ProposalMemory       `json:"related_proposal_memory"`
	RepoChangeJobs        []improvement.RepoChangeJob   `json:"repo_change_jobs"`
	PRAttempts            []improvement.PRAttempt       `json:"pr_attempts"`
	PostMergeReplays      []improvement.PostMergeReplay `json:"post_merge_replays"`
	LinkedTraceSummaries  []events.TraceSummary         `json:"linked_trace_summaries"`
	LinkedEvalRuns        []evals.Run                   `json:"linked_eval_runs"`
}

type runtimeRoleStatus struct {
	Role             string `json:"role"`
	ReportedRole     string `json:"reported_role,omitempty"`
	BaseURL          string `json:"base_url"`
	Status           string `json:"status"`
	Backend          string `json:"backend"`
	Provider         string `json:"provider"`
	Model            string `json:"model"`
	ProviderModel    string `json:"provider_model,omitempty"`
	APIMode          string `json:"api_mode,omitempty"`
	ReasoningEffort  string `json:"reasoning_effort"`
	Available        bool   `json:"available"`
	Healthy          bool   `json:"healthy"`
	OpenAIConfigured bool   `json:"openai_configured"`
	HermesAvailable  bool   `json:"hermes_available"`
	Error            string `json:"error,omitempty"`
}

func buildTraceList(store storepkg.Repository) []traceListItem {
	traces := store.ListTraces()
	latestEvalByTrace := latestEvalRunByTrace(store.ListEvalRuns())
	out := make([]traceListItem, 0, len(traces))
	for _, trace := range traces {
		item := traceListItem{
			TraceID:          trace.TraceID,
			WorkflowID:       trace.WorkflowID,
			IngestionID:      trace.IngestionID,
			WorkflowKind:     trace.WorkflowKind,
			Status:           trace.Status,
			ThreadKey:        trace.ThreadKey,
			StartedAt:        trace.StartedAt,
			EventCount:       trace.EventCount,
			ReasoningCount:   trace.ReasoningStepCount,
			ToolCallCount:    trace.ToolCallCount,
			SlackActionCount: trace.SlackActionCount,
		}
		if run, ok := latestEvalByTrace[trace.TraceID]; ok {
			item.LatestEval = &traceEvalSummary{
				RunID:     run.ID,
				Verdict:   run.OverallVerdict,
				Score:     run.OverallScore,
				CreatedAt: run.CreatedAt,
				SuiteName: run.SuiteName,
			}
		}
		out = append(out, item)
	}
	return out
}

func buildTraceDetail(store storepkg.Repository, traceID string) (traceDetailResponse, bool) {
	trace, ok := store.GetTrace(traceID)
	if !ok {
		return traceDetailResponse{}, false
	}
	trace = normalizeTrace(trace)
	runs := filterEvalRunsForTrace(store.ListEvalRuns(), traceID)
	judgments := map[string][]evals.Judgment{}
	for _, run := range runs {
		judgments[run.ID] = sliceOrEmpty(store.ListEvalJudgments(run.ID))
	}
	return traceDetailResponse{
		Trace:              trace,
		LinkedEvalRuns:     runs,
		JudgmentsByEvalRun: judgments,
		LinkedProposals:    normalizeProposals(filterProposalsForTrace(store.ListProposals(), traceID)),
		Ratings:            sliceOrEmpty(store.ListRatings(traceID)),
		ImprovementNotes:   sliceOrEmpty(store.ListImprovementNotes(traceID)),
	}, true
}

func buildProposalDetail(store storepkg.Repository, proposalID string) (proposalDetailResponse, bool) {
	proposal, ok := findProposal(store.ListProposals(), proposalID)
	if !ok {
		return proposalDetailResponse{}, false
	}
	traceSummaries := linkTraceSummaries(store.ListTraces(), proposal, store.ListPostMergeReplays())
	return proposalDetailResponse{
		Proposal:              normalizeProposal(proposal),
		Reviews:               sliceOrEmpty(proposal.Reviews),
		RelatedProposalMemory: sliceOrEmpty(filterProposalMemory(store.ListProposalMemories(), proposal.CandidateKey)),
		RepoChangeJobs:        sliceOrEmpty(filterRepoChangeJobs(store.ListRepoChangeJobs(), proposal.ID)),
		PRAttempts:            sliceOrEmpty(filterPRAttempts(store.ListPRAttempts(), proposal.ID)),
		PostMergeReplays:      sliceOrEmpty(filterPostMergeReplays(store.ListPostMergeReplays(), proposal.ID)),
		LinkedTraceSummaries:  sliceOrEmpty(traceSummaries),
		LinkedEvalRuns:        sliceOrEmpty(filterEvalRunsForProposal(store.ListEvalRuns(), proposal)),
	}, true
}

func buildRuntimeStatus(cfg config.Config) []runtimeRoleStatus {
	roleURLs := cfg.RunnerURLs()
	cache := map[string]clients.RuntimeResponse{}
	cacheErrs := map[string]error{}
	roles := []string{"prod", "proactive", "eval", "proposal"}
	out := make([]runtimeRoleStatus, 0, len(roles))
	for _, role := range roles {
		baseURL := roleURLs[role]
		if _, ok := cache[baseURL]; !ok && cacheErrs[baseURL] == nil {
			resp, err := clients.NewRunnerClient(baseURL).Runtime()
			if err != nil {
				cacheErrs[baseURL] = err
			} else {
				cache[baseURL] = resp
			}
		}
		item := runtimeRoleStatus{
			Role:            role,
			BaseURL:         baseURL,
			Status:          "unreachable",
			Model:           "openai/gpt-5.4",
			ReasoningEffort: "xhigh",
		}
		if err := cacheErrs[baseURL]; err != nil {
			item.Error = err.Error()
			out = append(out, item)
			continue
		}
		resp := cache[baseURL]
		item.ReportedRole = resp.Role
		item.Status = resp.Status
		item.Backend = resp.Backend
		item.Provider = resp.Provider
		item.Model = firstNonEmpty(strings.TrimSpace(resp.Model), item.Model)
		item.ProviderModel = resp.ProviderModel
		item.APIMode = resp.APIMode
		item.ReasoningEffort = firstNonEmpty(strings.TrimSpace(resp.ReasoningEffort), item.ReasoningEffort)
		item.Available = resp.Available
		item.Healthy = resp.Available && strings.EqualFold(resp.Status, "ok")
		item.OpenAIConfigured = resp.OpenAIConfigured
		item.HermesAvailable = resp.HermesAvailable
		out = append(out, item)
	}
	return out
}

func normalizeProposals(items []review.Proposal) []review.Proposal {
	out := make([]review.Proposal, 0, len(items))
	for _, item := range items {
		out = append(out, normalizeProposal(item))
	}
	return out
}

func normalizeProposal(item review.Proposal) review.Proposal {
	item.SourceEvalIDs = sliceOrEmpty(item.SourceEvalIDs)
	item.EvidenceArtifactIDs = sliceOrEmpty(item.EvidenceArtifactIDs)
	item.PriorSimilarProposalIDs = sliceOrEmpty(item.PriorSimilarProposalIDs)
	item.Reviews = sliceOrEmpty(item.Reviews)
	return item
}

func filterProposalsForTrace(items []review.Proposal, traceID string) []review.Proposal {
	out := make([]review.Proposal, 0)
	for _, item := range items {
		if item.TraceID == traceID {
			out = append(out, item)
		}
	}
	return out
}

func filterEvalRunsForTrace(items []evals.Run, traceID string) []evals.Run {
	out := make([]evals.Run, 0)
	for _, item := range items {
		if item.TraceID == traceID {
			out = append(out, item)
		}
	}
	return out
}

func filterEvalRunsForProposal(items []evals.Run, proposal review.Proposal) []evals.Run {
	byID := map[string]struct{}{}
	for _, sourceEvalID := range proposal.SourceEvalIDs {
		byID[sourceEvalID] = struct{}{}
	}
	out := make([]evals.Run, 0)
	for _, item := range items {
		if item.TraceID == proposal.TraceID {
			out = append(out, item)
			continue
		}
		if _, ok := byID[item.ID]; ok {
			out = append(out, item)
		}
	}
	return out
}

func latestEvalRunByTrace(items []evals.Run) map[string]evals.Run {
	out := map[string]evals.Run{}
	for _, item := range items {
		current, ok := out[item.TraceID]
		if !ok || item.CreatedAt.After(current.CreatedAt) {
			out[item.TraceID] = item
		}
	}
	return out
}

func filterRepoChangeJobs(items []improvement.RepoChangeJob, proposalID string) []improvement.RepoChangeJob {
	out := make([]improvement.RepoChangeJob, 0)
	for _, item := range items {
		if item.ProposalID == proposalID {
			out = append(out, item)
		}
	}
	return out
}

func filterPRAttempts(items []improvement.PRAttempt, proposalID string) []improvement.PRAttempt {
	out := make([]improvement.PRAttempt, 0)
	for _, item := range items {
		if item.ProposalID == proposalID {
			out = append(out, item)
		}
	}
	return out
}

func filterPostMergeReplays(items []improvement.PostMergeReplay, proposalID string) []improvement.PostMergeReplay {
	out := make([]improvement.PostMergeReplay, 0)
	for _, item := range items {
		if item.ProposalID == proposalID {
			out = append(out, item)
		}
	}
	return out
}

func linkTraceSummaries(items []events.TraceSummary, proposal review.Proposal, replays []improvement.PostMergeReplay) []events.TraceSummary {
	index := map[string]events.TraceSummary{}
	for _, item := range items {
		index[item.TraceID] = item
	}
	seen := map[string]struct{}{}
	out := make([]events.TraceSummary, 0)
	for _, traceID := range proposalLinkedTraceIDs(proposal, replays) {
		if _, ok := seen[traceID]; ok {
			continue
		}
		seen[traceID] = struct{}{}
		trace, ok := index[traceID]
		if ok {
			out = append(out, trace)
		}
	}
	return out
}

func proposalLinkedTraceIDs(proposal review.Proposal, replays []improvement.PostMergeReplay) []string {
	out := []string{proposal.TraceID}
	for _, replay := range replays {
		if replay.ProposalID == proposal.ID && strings.TrimSpace(replay.TraceID) != "" {
			out = append(out, replay.TraceID)
		}
	}
	return out
}

func runtimeSummary(status runtimeRoleStatus) string {
	if status.Error != "" {
		return fmt.Sprintf("%s unavailable: %s", status.Role, status.Error)
	}
	return fmt.Sprintf("%s -> %s %s effort=%s", status.Role, status.Backend, status.Model, status.ReasoningEffort)
}
