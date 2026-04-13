package improvementplane

import (
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/knowledge"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
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

type traceAttemptSummary struct {
	TraceID           string            `json:"trace_id"`
	ConversationID    string            `json:"conversation_id"`
	CaseID            string            `json:"case_id"`
	TriggerEventID    string            `json:"trigger_event_id"`
	SupersedesTraceID string            `json:"supersedes_trace_id,omitempty"`
	WorkflowKind      string            `json:"workflow_kind"`
	Status            events.Status     `json:"status"`
	ThreadKey         string            `json:"thread_key"`
	StartedAt         time.Time         `json:"started_at"`
	EventCount        int               `json:"event_count"`
	ReasoningCount    int               `json:"reasoning_count"`
	ToolCallCount     int               `json:"tool_call_count"`
	SlackActionCount  int               `json:"slack_action_count"`
	LatestEval        *traceEvalSummary `json:"latest_eval,omitempty"`
}

type caseSummary struct {
	CaseID             string                       `json:"case_id"`
	ConversationID     string                       `json:"conversation_id"`
	Kind               string                       `json:"kind"`
	Intent             string                       `json:"intent"`
	Title              string                       `json:"title"`
	Summary            string                       `json:"summary"`
	Status             conversation.CaseStatus      `json:"status"`
	ResolutionState    conversation.ResolutionState `json:"resolution_state"`
	AssignedBot        string                       `json:"assigned_bot"`
	LatestTraceID      string                       `json:"latest_trace_id,omitempty"`
	LatestTraceVerdict string                       `json:"latest_trace_verdict,omitempty"`
	LatestOutcomeID    string                       `json:"latest_outcome_id,omitempty"`
	OutcomeScore       float64                      `json:"outcome_score,omitempty"`
	Recurrence         int                          `json:"recurrence"`
	LinkedProposalIDs  []string                     `json:"linked_proposal_ids"`
	UpdatedAt          time.Time                    `json:"updated_at"`
}

type conversationListItem struct {
	ConversationID     string       `json:"conversation_id"`
	Source             string       `json:"source"`
	ExternalKey        string       `json:"external_key"`
	Title              string       `json:"title"`
	Status             string       `json:"status"`
	ActiveCase         *caseSummary `json:"active_case,omitempty"`
	LatestMessageAt    time.Time    `json:"latest_message_at"`
	LatestTraceVerdict string       `json:"latest_trace_verdict,omitempty"`
	OpenTraceCount     int          `json:"open_trace_count"`
	ProposalCount      int          `json:"proposal_count"`
}

type conversationDetailResponse struct {
	Conversation     conversation.Conversation `json:"conversation"`
	ActiveCase       *caseSummary              `json:"active_case,omitempty"`
	Cases            []caseSummary             `json:"cases"`
	Transcript       []conversation.Entry      `json:"transcript"`
	TraceAttempts    []traceAttemptSummary     `json:"trace_attempts"`
	ActionIntents    []action.Intent           `json:"action_intents"`
	ActionResults    []action.Result           `json:"action_results"`
	Outcomes         []outcome.Record          `json:"outcomes"`
	KnowledgeEntries []knowledge.Entry         `json:"knowledge_entries"`
	LinkedProposals  []review.Proposal         `json:"linked_proposals"`
}

type caseDetailResponse struct {
	Case             caseSummary           `json:"case"`
	Conversation     conversationListItem  `json:"conversation"`
	TraceAttempts    []traceAttemptSummary `json:"trace_attempts"`
	LatestEvalRuns   []evals.Run           `json:"latest_eval_runs"`
	ActionIntents    []action.Intent       `json:"action_intents"`
	ActionResults    []action.Result       `json:"action_results"`
	Outcomes         []outcome.Record      `json:"outcomes"`
	KnowledgeEntries []knowledge.Entry     `json:"knowledge_entries"`
	LinkedProposals  []review.Proposal     `json:"linked_proposals"`
}

type traceDetailResponse struct {
	Trace              events.Trace                `json:"trace"`
	Conversation       conversationListItem        `json:"conversation"`
	Case               *caseSummary                `json:"case,omitempty"`
	TranscriptSlice    []conversation.Entry        `json:"transcript_slice"`
	LinkedEvalRuns     []evals.Run                 `json:"linked_eval_runs"`
	JudgmentsByEvalRun map[string][]evals.Judgment `json:"judgments_by_eval_run"`
	ActionIntents      []action.Intent             `json:"action_intents"`
	ActionResults      []action.Result             `json:"action_results"`
	Outcomes           []outcome.Record            `json:"outcomes"`
	KnowledgeEntries   []knowledge.Entry           `json:"knowledge_entries"`
	FeedbackRecords    []review.FeedbackRecord     `json:"feedback_records"`
	LinkedProposals    []review.Proposal           `json:"linked_proposals"`
}

type proposalDetailResponse struct {
	Proposal              review.Proposal               `json:"proposal"`
	Reviews               []review.ProposalReview       `json:"reviews"`
	RelatedProposalMemory []review.ProposalMemory       `json:"related_proposal_memory"`
	RepoChangeJobs        []improvement.RepoChangeJob   `json:"repo_change_jobs"`
	PRAttempts            []improvement.PRAttempt       `json:"pr_attempts"`
	PostMergeReplays      []improvement.PostMergeReplay `json:"post_merge_replays"`
	LinkedTraceSummaries  []traceAttemptSummary         `json:"linked_trace_summaries"`
	LinkedEvalRuns        []evals.Run                   `json:"linked_eval_runs"`
	ActionIntents         []action.Intent               `json:"action_intents"`
	ActionResults         []action.Result               `json:"action_results"`
	Outcomes              []outcome.Record              `json:"outcomes"`
	KnowledgeEntries      []knowledge.Entry             `json:"knowledge_entries"`
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

func buildConversationList(store storepkg.Repository) []conversationListItem {
	conversations := store.ListConversations()
	traces := store.ListTraces()
	proposals := normalizeProposals(store.ListProposals())
	latestEvalByTrace := latestEvalRunByTrace(store.ListEvalRuns())
	traceSummaries := buildTraceSummaries(traces, latestEvalByTrace)
	tracesByConversation := map[string][]traceAttemptSummary{}
	for _, item := range traceSummaries {
		tracesByConversation[item.ConversationID] = append(tracesByConversation[item.ConversationID], item)
	}
	proposalsByConversation := map[string]int{}
	for _, proposal := range proposals {
		if proposal.ConversationID != "" {
			proposalsByConversation[proposal.ConversationID]++
		}
	}
	caseIndex := buildCaseSummaryIndex(store, traces, proposals)
	out := make([]conversationListItem, 0, len(conversations))
	for _, item := range conversations {
		tracesForConversation := tracesByConversation[item.ID]
		activeCase := caseIndex[item.ActiveCaseID]
		latestMessageAt := item.UpdatedAt
		latestTraceVerdict := ""
		openTraceCount := 0
		for _, trace := range tracesForConversation {
			if trace.StartedAt.After(latestMessageAt) {
				latestMessageAt = trace.StartedAt
			}
			if latestTraceVerdict == "" && trace.LatestEval != nil {
				latestTraceVerdict = trace.LatestEval.Verdict
			}
			if isOpenTraceStatus(trace.Status) {
				openTraceCount++
			}
		}
		out = append(out, conversationListItem{
			ConversationID:     item.ID,
			Source:             string(item.Source),
			ExternalKey:        item.ExternalKey,
			Title:              firstNonEmptyString(item.Title, item.ExternalKey),
			Status:             string(item.Status),
			ActiveCase:         activeCase,
			LatestMessageAt:    latestMessageAt,
			LatestTraceVerdict: latestTraceVerdict,
			OpenTraceCount:     openTraceCount,
			ProposalCount:      proposalsByConversation[item.ID],
		})
	}
	return out
}

func buildConversationDetail(store storepkg.Repository, conversationID string) (conversationDetailResponse, bool) {
	item, ok := store.GetConversation(conversationID)
	if !ok {
		return conversationDetailResponse{}, false
	}
	proposals := normalizeProposals(store.ListProposals())
	traces := traceSummariesForConversation(store.ListTraces(), conversationID)
	latestEvalByTrace := latestEvalRunByTrace(store.ListEvalRuns())
	traceSummaries := buildTraceSummaries(traces, latestEvalByTrace)
	caseIndex := buildCaseSummaryIndex(store, store.ListTraces(), proposals)
	cases := casesForConversation(store.ListCases(), conversationID, caseIndex)
	return conversationDetailResponse{
		Conversation:     item,
		ActiveCase:       caseIndex[item.ActiveCaseID],
		Cases:            cases,
		Transcript:       sliceOrEmpty(store.ListConversationEntries(conversationID)),
		TraceAttempts:    traceSummaries,
		ActionIntents:    sliceOrEmpty(listActionIntents(store, actionFilters{ConversationID: conversationID})),
		ActionResults:    sliceOrEmpty(flattenActionResults(store, listActionIntents(store, actionFilters{ConversationID: conversationID}))),
		Outcomes:         sliceOrEmpty(listOutcomes(store, conversationID, "", "", "")),
		KnowledgeEntries: sliceOrEmpty(relatedKnowledgeEntries(store, conversationID, "", "", "")),
		LinkedProposals:  filterProposalsByConversation(proposals, conversationID),
	}, true
}

func buildCaseList(store storepkg.Repository) []caseSummary {
	return caseSummaries(store, store.ListCases())
}

func buildCaseDetail(store storepkg.Repository, caseID string) (caseDetailResponse, bool) {
	caseIndex := buildCaseSummaryIndex(store, store.ListTraces(), normalizeProposals(store.ListProposals()))
	caseItem, ok := caseIndex[caseID]
	if !ok {
		return caseDetailResponse{}, false
	}
	conversationSummary, ok := findConversationSummary(buildConversationList(store), caseItem.ConversationID)
	if !ok {
		return caseDetailResponse{}, false
	}
	traces := traceSummariesForCase(store.ListTraces(), caseID)
	latestEvalByTrace := latestEvalRunByTrace(store.ListEvalRuns())
	traceSummaries := buildTraceSummaries(traces, latestEvalByTrace)
	return caseDetailResponse{
		Case:             *caseItem,
		Conversation:     conversationSummary,
		TraceAttempts:    traceSummaries,
		LatestEvalRuns:   latestEvalRunsForTraceSet(store.ListEvalRuns(), traceSummaries),
		ActionIntents:    sliceOrEmpty(listActionIntents(store, actionFilters{CaseID: caseID})),
		ActionResults:    sliceOrEmpty(flattenActionResults(store, listActionIntents(store, actionFilters{CaseID: caseID}))),
		Outcomes:         sliceOrEmpty(listOutcomes(store, "", caseID, "", "")),
		KnowledgeEntries: sliceOrEmpty(relatedKnowledgeEntries(store, conversationSummary.ConversationID, caseID, "", "")),
		LinkedProposals:  filterProposalsByCase(normalizeProposals(store.ListProposals()), caseID),
	}, true
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
	conversations := buildConversationList(store)
	conversationSummary, _ := findConversationSummary(conversations, trace.Summary.ConversationID)
	caseIndex := buildCaseSummaryIndex(store, store.ListTraces(), normalizeProposals(store.ListProposals()))
	actionIntents := listActionIntents(store, actionFilters{TraceID: traceID})
	outcomes := listOutcomes(store, trace.Summary.ConversationID, trace.Summary.CaseID, traceID, "")
	extraEvidence := make([]string, 0, len(trace.Reasoning)+len(trace.ToolCalls)+len(outcomes))
	for _, step := range trace.Reasoning {
		extraEvidence = append(extraEvidence, step.ID)
	}
	for _, call := range trace.ToolCalls {
		extraEvidence = append(extraEvidence, call.ID)
	}
	for _, item := range outcomes {
		extraEvidence = append(extraEvidence, item.ID)
	}
	return traceDetailResponse{
		Trace:              trace,
		Conversation:       conversationSummary,
		Case:               caseIndex[trace.Summary.CaseID],
		TranscriptSlice:    transcriptSlice(store.ListConversationEntries(trace.Summary.ConversationID), trace.Summary.TriggerEventID),
		LinkedEvalRuns:     runs,
		JudgmentsByEvalRun: judgments,
		ActionIntents:      sliceOrEmpty(actionIntents),
		ActionResults:      sliceOrEmpty(flattenActionResults(store, actionIntents)),
		Outcomes:           sliceOrEmpty(outcomes),
		KnowledgeEntries:   sliceOrEmpty(relatedKnowledgeEntries(store, trace.Summary.ConversationID, trace.Summary.CaseID, traceID, "", extraEvidence...)),
		FeedbackRecords:    sliceOrEmpty(store.ListFeedback(traceID)),
		LinkedProposals:    filterProposalsForTrace(normalizeProposals(store.ListProposals()), traceID),
	}, true
}

func buildProposalDetail(store storepkg.Repository, proposalID string) (proposalDetailResponse, bool) {
	proposal, ok := findProposalView(store.ListProposals(), proposalID)
	if !ok {
		return proposalDetailResponse{}, false
	}
	traceSummaries := linkTraceSummaries(store.ListTraces(), proposal, store.ListPostMergeReplays())
	actionIntents := listActionIntents(store, actionFilters{ProposalID: proposal.ID})
	outcomes := listOutcomes(store, proposal.ConversationID, proposal.CaseID, "", proposal.ID)
	extraEvidence := append([]string{}, proposal.EvidenceTraceIDs...)
	extraEvidence = appendUniqueStrings(extraEvidence, proposal.OriginTraceID, proposal.TraceID)
	for _, item := range outcomes {
		extraEvidence = append(extraEvidence, item.ID)
	}
	return proposalDetailResponse{
		Proposal:              normalizeProposal(proposal),
		Reviews:               sliceOrEmpty(proposal.Reviews),
		RelatedProposalMemory: sliceOrEmpty(filterProposalMemory(store.ListProposalMemories(), proposal.CandidateKey)),
		RepoChangeJobs:        sliceOrEmpty(filterRepoChangeJobs(store.ListRepoChangeJobs(), proposal.ID)),
		PRAttempts:            sliceOrEmpty(filterPRAttempts(store.ListPRAttempts(), proposal.ID)),
		PostMergeReplays:      sliceOrEmpty(filterPostMergeReplays(store.ListPostMergeReplays(), proposal.ID)),
		LinkedTraceSummaries:  traceSummaries,
		LinkedEvalRuns:        sliceOrEmpty(filterEvalRunsForProposal(store.ListEvalRuns(), proposal)),
		ActionIntents:         sliceOrEmpty(actionIntents),
		ActionResults:         sliceOrEmpty(flattenActionResults(store, actionIntents)),
		Outcomes:              sliceOrEmpty(outcomes),
		KnowledgeEntries:      sliceOrEmpty(relatedKnowledgeEntries(store, proposal.ConversationID, proposal.CaseID, proposal.OriginTraceID, proposal.ID, extraEvidence...)),
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
		item.Model = firstNonEmptyString(strings.TrimSpace(resp.Model), item.Model)
		item.ProviderModel = resp.ProviderModel
		item.APIMode = resp.APIMode
		item.ReasoningEffort = firstNonEmptyString(strings.TrimSpace(resp.ReasoningEffort), item.ReasoningEffort)
		item.Available = resp.Available
		item.Healthy = resp.Available && strings.EqualFold(resp.Status, "ok")
		item.OpenAIConfigured = resp.OpenAIConfigured
		item.HermesAvailable = resp.HermesAvailable
		out = append(out, item)
	}
	return out
}

func buildCaseSummaryIndex(store storepkg.Repository, traces []events.TraceSummary, proposals []review.Proposal) map[string]*caseSummary {
	latestEvalByTrace := latestEvalRunByTrace(store.ListEvalRuns())
	proposalIDsByCase := map[string][]string{}
	for _, proposal := range proposals {
		if proposal.CaseID == "" {
			continue
		}
		proposalIDsByCase[proposal.CaseID] = appendUniqueStrings(proposalIDsByCase[proposal.CaseID], proposal.ID)
	}
	traceVerdictByCase := map[string]string{}
	recurrenceByCase := map[string]int{}
	for _, trace := range traces {
		recurrenceByCase[trace.CaseID]++
		if trace.CaseID == "" {
			continue
		}
		if current, ok := traceVerdictByCase[trace.CaseID]; ok && current != "" {
			continue
		}
		if run, ok := latestEvalByTrace[trace.TraceID]; ok {
			traceVerdictByCase[trace.CaseID] = run.OverallVerdict
		}
	}
	out := map[string]*caseSummary{}
	for _, item := range store.ListCases() {
		summary := &caseSummary{
			CaseID:             item.ID,
			ConversationID:     item.ConversationID,
			Kind:               item.Kind,
			Intent:             item.Intent,
			Title:              item.Title,
			Summary:            item.Summary,
			Status:             item.Status,
			ResolutionState:    item.ResolutionState,
			AssignedBot:        item.AssignedBot,
			LatestTraceID:      item.LatestTraceID,
			LatestTraceVerdict: traceVerdictByCase[item.ID],
			LatestOutcomeID:    item.LatestOutcomeID,
			OutcomeScore:       item.OutcomeScore,
			Recurrence:         recurrenceByCase[item.ID],
			LinkedProposalIDs:  sliceOrEmpty(proposalIDsByCase[item.ID]),
			UpdatedAt:          item.UpdatedAt,
		}
		out[item.ID] = summary
	}
	return out
}

func caseSummaries(store storepkg.Repository, cases []conversation.Case) []caseSummary {
	index := buildCaseSummaryIndex(store, store.ListTraces(), normalizeProposals(store.ListProposals()))
	out := make([]caseSummary, 0, len(cases))
	for _, item := range cases {
		if summary, ok := index[item.ID]; ok {
			out = append(out, *summary)
		}
	}
	return out
}

func casesForConversation(cases []conversation.Case, conversationID string, index map[string]*caseSummary) []caseSummary {
	out := make([]caseSummary, 0)
	for _, item := range cases {
		if item.ConversationID != conversationID {
			continue
		}
		if summary, ok := index[item.ID]; ok {
			out = append(out, *summary)
		}
	}
	return out
}

func buildTraceSummaries(traces []events.TraceSummary, latestEvalByTrace map[string]evals.Run) []traceAttemptSummary {
	out := make([]traceAttemptSummary, 0, len(traces))
	for _, trace := range traces {
		item := traceAttemptSummary{
			TraceID:           trace.TraceID,
			ConversationID:    trace.ConversationID,
			CaseID:            trace.CaseID,
			TriggerEventID:    trace.TriggerEventID,
			SupersedesTraceID: trace.SupersedesTraceID,
			WorkflowKind:      trace.WorkflowKind,
			Status:            trace.Status,
			ThreadKey:         trace.ThreadKey,
			StartedAt:         trace.StartedAt,
			EventCount:        trace.EventCount,
			ReasoningCount:    trace.ReasoningStepCount,
			ToolCallCount:     trace.ToolCallCount,
			SlackActionCount:  trace.SlackActionCount,
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

func traceSummariesForConversation(traces []events.TraceSummary, conversationID string) []events.TraceSummary {
	out := make([]events.TraceSummary, 0)
	for _, item := range traces {
		if item.ConversationID == conversationID {
			out = append(out, item)
		}
	}
	return out
}

func traceSummariesForCase(traces []events.TraceSummary, caseID string) []events.TraceSummary {
	out := make([]events.TraceSummary, 0)
	for _, item := range traces {
		if item.CaseID == caseID {
			out = append(out, item)
		}
	}
	return out
}

func filterProposalsForTrace(items []review.Proposal, traceID string) []review.Proposal {
	out := make([]review.Proposal, 0)
	for _, item := range items {
		if item.TraceID == traceID || item.OriginTraceID == traceID || slicesContain(item.EvidenceTraceIDs, traceID) {
			out = append(out, item)
		}
	}
	return out
}

func filterProposalsByConversation(items []review.Proposal, conversationID string) []review.Proposal {
	out := make([]review.Proposal, 0)
	for _, item := range items {
		if item.ConversationID == conversationID {
			out = append(out, item)
		}
	}
	return out
}

func filterProposalsByCase(items []review.Proposal, caseID string) []review.Proposal {
	out := make([]review.Proposal, 0)
	for _, item := range items {
		if item.CaseID == caseID {
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
		if item.TraceID == proposal.TraceID || item.TraceID == proposal.OriginTraceID || slicesContain(proposal.EvidenceTraceIDs, item.TraceID) {
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

func latestEvalRunsForTraceSet(items []evals.Run, traces []traceAttemptSummary) []evals.Run {
	traceSet := map[string]struct{}{}
	for _, item := range traces {
		traceSet[item.TraceID] = struct{}{}
	}
	latest := map[string]evals.Run{}
	for _, item := range items {
		if _, ok := traceSet[item.TraceID]; !ok {
			continue
		}
		current, ok := latest[item.TraceID]
		if !ok || item.CreatedAt.After(current.CreatedAt) {
			latest[item.TraceID] = item
		}
	}
	out := make([]evals.Run, 0, len(latest))
	for _, item := range latest {
		out = append(out, item)
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

func linkTraceSummaries(items []events.TraceSummary, proposal review.Proposal, replays []improvement.PostMergeReplay) []traceAttemptSummary {
	index := map[string]events.TraceSummary{}
	for _, item := range items {
		index[item.TraceID] = item
	}
	out := make([]events.TraceSummary, 0)
	seen := map[string]struct{}{}
	for _, traceID := range proposalLinkedTraceIDs(proposal, replays) {
		if _, ok := seen[traceID]; ok {
			continue
		}
		seen[traceID] = struct{}{}
		if trace, ok := index[traceID]; ok {
			out = append(out, trace)
		}
	}
	return buildTraceSummaries(out, map[string]evals.Run{})
}

func proposalLinkedTraceIDs(proposal review.Proposal, replays []improvement.PostMergeReplay) []string {
	out := []string{firstNonEmptyString(proposal.OriginTraceID, proposal.TraceID)}
	out = append(out, proposal.EvidenceTraceIDs...)
	for _, replay := range replays {
		if replay.ProposalID == proposal.ID && strings.TrimSpace(replay.TraceID) != "" {
			out = append(out, replay.TraceID)
		}
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
	item.EvidenceTraceIDs = sliceOrEmpty(item.EvidenceTraceIDs)
	item.Reviews = sliceOrEmpty(item.Reviews)
	return item
}

func normalizeTrace(trace events.Trace) events.Trace {
	trace.Events = sliceOrEmpty(trace.Events)
	trace.Artifacts = sliceOrEmpty(trace.Artifacts)
	trace.Reasoning = sliceOrEmpty(trace.Reasoning)
	trace.ToolCalls = sliceOrEmpty(trace.ToolCalls)
	trace.SlackActions = sliceOrEmpty(trace.SlackActions)
	return trace
}

func transcriptSlice(entries []conversation.Entry, triggerEventID string) []conversation.Entry {
	if len(entries) == 0 {
		return []conversation.Entry{}
	}
	if triggerEventID == "" {
		if len(entries) <= 12 {
			return entries
		}
		return entries[len(entries)-12:]
	}
	position := -1
	for i, item := range entries {
		if item.EventID == triggerEventID {
			position = i
		}
	}
	if position == -1 {
		if len(entries) <= 12 {
			return entries
		}
		return entries[len(entries)-12:]
	}
	start := position - 6
	if start < 0 {
		start = 0
	}
	end := position + 1
	if end < len(entries) {
		end = minInt(end+5, len(entries))
	}
	return entries[start:end]
}

func findConversationSummary(items []conversationListItem, conversationID string) (conversationListItem, bool) {
	for _, item := range items {
		if item.ConversationID == conversationID {
			return item, true
		}
	}
	return conversationListItem{}, false
}

func findProposalView(items []review.Proposal, proposalID string) (review.Proposal, bool) {
	for _, item := range items {
		if item.ID == proposalID {
			return item, true
		}
	}
	return review.Proposal{}, false
}

func isOpenTraceStatus(status events.Status) bool {
	switch status {
	case events.StatusQueued, events.StatusRunning, events.StatusNeedsHuman, events.StatusInReview, events.StatusReplayed:
		return true
	default:
		return false
	}
}

func runtimeSummary(status runtimeRoleStatus) string {
	if status.Error != "" {
		return fmt.Sprintf("%s unavailable: %s", status.Role, status.Error)
	}
	return fmt.Sprintf("%s -> %s %s effort=%s", status.Role, status.Backend, status.Model, status.ReasoningEffort)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func slicesContain(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func appendUniqueStrings(existing []string, values ...string) []string {
	seen := map[string]struct{}{}
	out := append([]string(nil), existing...)
	for _, item := range existing {
		seen[item] = struct{}{}
	}
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
