package improvementplane

import (
	"github.com/piplabs/rsi-agent-platform/internal/action"
	"github.com/piplabs/rsi-agent-platform/internal/conversation"
	"github.com/piplabs/rsi-agent-platform/internal/evals"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/outcome"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func storeTraceSummariesForConversation(store storepkg.Repository, conversationID string) []events.TraceSummary {
	if reader, ok := store.(interface {
		ListTracesByConversation(string) []events.TraceSummary
	}); ok {
		return reader.ListTracesByConversation(conversationID)
	}
	return traceSummariesForConversation(store.ListTraces(), conversationID)
}

func storeTraceSummariesForCase(store storepkg.Repository, caseID string) []events.TraceSummary {
	if reader, ok := store.(interface {
		ListTracesByCase(string) []events.TraceSummary
	}); ok {
		return reader.ListTracesByCase(caseID)
	}
	return traceSummariesForCase(store.ListTraces(), caseID)
}

func storeCasesForConversation(store storepkg.Repository, conversationID string) []conversation.Case {
	if reader, ok := store.(interface {
		ListCasesByConversation(string) []conversation.Case
	}); ok {
		return reader.ListCasesByConversation(conversationID)
	}
	out := make([]conversation.Case, 0)
	for _, item := range store.ListCases() {
		if item.ConversationID == conversationID {
			out = append(out, item)
		}
	}
	return out
}

func storeWorkflowsForConversation(store storepkg.Repository, conversationID string) []storepkg.Workflow {
	if reader, ok := store.(interface {
		ListWorkflowsByConversation(string) []storepkg.Workflow
	}); ok {
		return reader.ListWorkflowsByConversation(conversationID)
	}
	out := make([]storepkg.Workflow, 0)
	for _, item := range store.ListWorkflows() {
		if item.ConversationID == conversationID {
			out = append(out, item)
		}
	}
	return out
}

func storeWorkflowsForCase(store storepkg.Repository, caseID string) []storepkg.Workflow {
	if reader, ok := store.(interface {
		ListWorkflowsByCase(string) []storepkg.Workflow
	}); ok {
		return reader.ListWorkflowsByCase(caseID)
	}
	out := make([]storepkg.Workflow, 0)
	for _, item := range store.ListWorkflows() {
		if item.CaseID == caseID {
			out = append(out, item)
		}
	}
	return out
}

func storeEvalRunsForTrace(store storepkg.Repository, traceID string) []evals.Run {
	if reader, ok := store.(interface{ ListEvalRunsByTrace(string) []evals.Run }); ok {
		return reader.ListEvalRunsByTrace(traceID)
	}
	return filterEvalRunsForTrace(store.ListEvalRuns(), traceID)
}

func storeEvalRunsForTraceSummaries(store storepkg.Repository, traces []events.TraceSummary) []evals.Run {
	traceIDs := traceIDsFromSummaries(traces)
	if reader, ok := store.(interface{ ListEvalRunsByTraceIDs([]string) []evals.Run }); ok {
		return reader.ListEvalRunsByTraceIDs(traceIDs)
	}
	traceSet := map[string]struct{}{}
	for _, traceID := range traceIDs {
		traceSet[traceID] = struct{}{}
	}
	out := make([]evals.Run, 0)
	for _, item := range store.ListEvalRuns() {
		if _, ok := traceSet[item.TraceID]; ok {
			out = append(out, item)
		}
	}
	return out
}

func storeProposalsByConversation(store storepkg.Repository, conversationID string) []review.Proposal {
	if reader, ok := store.(interface {
		ListProposalsByConversation(string) []review.Proposal
	}); ok {
		return normalizeProposals(reader.ListProposalsByConversation(conversationID))
	}
	return filterProposalsByConversation(normalizeProposals(store.ListProposals()), conversationID)
}

func storeProposalsByCase(store storepkg.Repository, caseID string) []review.Proposal {
	if reader, ok := store.(interface {
		ListProposalsByCase(string) []review.Proposal
	}); ok {
		return normalizeProposals(reader.ListProposalsByCase(caseID))
	}
	return filterProposalsByCase(normalizeProposals(store.ListProposals()), caseID)
}

func storeProposalsByTrace(store storepkg.Repository, traceID string) []review.Proposal {
	if reader, ok := store.(interface {
		ListProposalsByTrace(string) []review.Proposal
	}); ok {
		return normalizeProposals(reader.ListProposalsByTrace(traceID))
	}
	return filterProposalsForTrace(normalizeProposals(store.ListProposals()), traceID)
}

func storeHarnessExecutionsForTraces(store storepkg.Repository, traces []events.TraceSummary) []harness.Execution {
	return storeHarnessExecutionsForTraceIDs(store, traceIDsFromSummaries(traces))
}

func storeHarnessExecutionsForTraceIDs(store storepkg.Repository, traceIDs []string) []harness.Execution {
	if reader, ok := store.(interface {
		ListHarnessExecutionsByTraceIDs([]string) []harness.Execution
	}); ok {
		return reader.ListHarnessExecutionsByTraceIDs(traceIDs)
	}
	return store.ListHarnessExecutions()
}

func storeHarnessExecutionObservationsForTraces(store storepkg.Repository, traces []events.TraceSummary) []harness.ExecutionObservation {
	return storeHarnessExecutionObservationsForTraceIDs(store, traceIDsFromSummaries(traces))
}

func storeHarnessExecutionObservationsForTraceIDs(store storepkg.Repository, traceIDs []string) []harness.ExecutionObservation {
	if reader, ok := store.(interface {
		ListHarnessExecutionObservationsByTraceIDs([]string) []harness.ExecutionObservation
	}); ok {
		return reader.ListHarnessExecutionObservationsByTraceIDs(traceIDs)
	}
	return store.ListHarnessExecutionObservations()
}

func storeActionIntents(store storepkg.Repository, filters actionFilters) []action.Intent {
	switch {
	case filters.TraceID != "":
		if reader, ok := store.(interface{ ListActionIntentsByTrace(string) []action.Intent }); ok {
			return reader.ListActionIntentsByTrace(filters.TraceID)
		}
	case filters.ProposalID != "":
		if reader, ok := store.(interface{ ListActionIntentsByProposal(string) []action.Intent }); ok {
			return reader.ListActionIntentsByProposal(filters.ProposalID)
		}
	case filters.CaseID != "":
		if reader, ok := store.(interface{ ListActionIntentsByCase(string) []action.Intent }); ok {
			return reader.ListActionIntentsByCase(filters.CaseID)
		}
	case filters.ConversationID != "":
		if reader, ok := store.(interface{ ListActionIntentsByConversation(string) []action.Intent }); ok {
			return reader.ListActionIntentsByConversation(filters.ConversationID)
		}
	}
	return store.ListActionIntents()
}

func storeOutcomes(store storepkg.Repository, conversationID string, caseID string, traceID string, proposalID string) []outcome.Record {
	switch {
	case traceID != "":
		if reader, ok := store.(interface{ ListOutcomesByTrace(string) []outcome.Record }); ok {
			return reader.ListOutcomesByTrace(traceID)
		}
	case proposalID != "":
		if reader, ok := store.(interface{ ListOutcomesByProposal(string) []outcome.Record }); ok {
			return reader.ListOutcomesByProposal(proposalID)
		}
	case caseID != "":
		if reader, ok := store.(interface{ ListOutcomesByCase(string) []outcome.Record }); ok {
			return reader.ListOutcomesByCase(caseID)
		}
	case conversationID != "":
		if reader, ok := store.(interface{ ListOutcomesByConversation(string) []outcome.Record }); ok {
			return reader.ListOutcomesByConversation(conversationID)
		}
	}
	return store.ListOutcomes()
}

func storeLatestCandidateForTrace(store storepkg.Store, traceID string) (improvement.Candidate, bool) {
	if reader, ok := store.(interface {
		LatestCandidateForTrace(string) (improvement.Candidate, bool)
	}); ok {
		return reader.LatestCandidateForTrace(traceID)
	}
	return latestCandidateForTraceByScan(store, traceID)
}

func storeRuntimeDiagnosesForTrace(store storepkg.Repository, trace events.Trace, candidateKey string) []improvement.RuntimeDiagnosis {
	if reader, ok := store.(interface {
		ListRuntimeDiagnosesForTraceContext(string, string, string) []improvement.RuntimeDiagnosis
	}); ok {
		return reader.ListRuntimeDiagnosesForTraceContext(trace.Summary.TraceID, trace.Summary.CaseID, candidateKey)
	}
	return runtimeDiagnosesForTrace(store.ListRuntimeDiagnoses(), trace, candidateKey)
}

func buildConversationSummaryForID(store storepkg.Repository, conversationID string) (conversationListItem, bool) {
	item, ok := store.GetConversation(conversationID)
	if !ok {
		return conversationListItem{}, false
	}
	traces := storeTraceSummariesForConversation(store, conversationID)
	proposals := storeProposalsByConversation(store, conversationID)
	latestEvalByTrace := latestEvalRunByTrace(storeEvalRunsForTraceSummaries(store, traces))
	activeCase, _ := caseSummaryForCaseID(store, item.ActiveCaseID)
	latestMessageAt := item.UpdatedAt
	latestTraceVerdict := ""
	openTraceCount := 0
	for _, trace := range traces {
		if trace.StartedAt.After(latestMessageAt) {
			latestMessageAt = trace.StartedAt
		}
		if latestTraceVerdict == "" {
			if run, ok := latestEvalByTrace[trace.TraceID]; ok {
				latestTraceVerdict = run.OverallVerdict
			}
		}
		if isOpenTraceStatus(trace.Status) {
			openTraceCount++
		}
	}
	return conversationListItem{
		ConversationID:     item.ID,
		Source:             string(item.Source),
		ExternalKey:        item.ExternalKey,
		Title:              firstNonEmptyString(item.Title, item.ExternalKey),
		Status:             string(item.Status),
		ActiveCase:         activeCase,
		LatestMessageAt:    latestMessageAt,
		LatestTraceVerdict: latestTraceVerdict,
		OpenTraceCount:     openTraceCount,
		ProposalCount:      len(proposals),
	}, true
}

func caseSummaryForCaseID(store storepkg.Repository, caseID string) (*caseSummary, bool) {
	if caseID == "" {
		return nil, false
	}
	item, ok := store.GetCase(caseID)
	if !ok {
		return nil, false
	}
	traces := storeTraceSummariesForCase(store, caseID)
	proposals := storeProposalsByCase(store, caseID)
	summary := caseSummaryFromRecord(item, traces, latestEvalRunByTrace(storeEvalRunsForTraceSummaries(store, traces)), proposals)
	return &summary, true
}

func caseSummaryIndexForConversation(store storepkg.Repository, conversationID string, traces []events.TraceSummary, proposals []review.Proposal) map[string]*caseSummary {
	if traces == nil {
		traces = storeTraceSummariesForConversation(store, conversationID)
	}
	if proposals == nil {
		proposals = storeProposalsByConversation(store, conversationID)
	}
	latestEvalByTrace := latestEvalRunByTrace(storeEvalRunsForTraceSummaries(store, traces))
	tracesByCase := map[string][]events.TraceSummary{}
	for _, trace := range traces {
		tracesByCase[trace.CaseID] = append(tracesByCase[trace.CaseID], trace)
	}
	proposalsByCase := map[string][]review.Proposal{}
	for _, proposal := range proposals {
		proposalsByCase[proposal.CaseID] = append(proposalsByCase[proposal.CaseID], proposal)
	}
	out := map[string]*caseSummary{}
	for _, item := range storeCasesForConversation(store, conversationID) {
		summary := caseSummaryFromRecord(item, tracesByCase[item.ID], latestEvalByTrace, proposalsByCase[item.ID])
		out[item.ID] = &summary
	}
	return out
}

func caseSummaryFromRecord(item conversation.Case, traces []events.TraceSummary, latestEvalByTrace map[string]evals.Run, proposals []review.Proposal) caseSummary {
	traceVerdict := ""
	recurrence := 0
	for _, trace := range traces {
		recurrence++
		if traceVerdict == "" {
			if run, ok := latestEvalByTrace[trace.TraceID]; ok {
				traceVerdict = run.OverallVerdict
			}
		}
	}
	proposalIDs := make([]string, 0, len(proposals))
	for _, proposal := range proposals {
		proposalIDs = appendUniqueStrings(proposalIDs, proposal.ID)
	}
	return caseSummary{
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
		LatestTraceVerdict: traceVerdict,
		LatestOutcomeID:    item.LatestOutcomeID,
		OutcomeScore:       item.OutcomeScore,
		Recurrence:         recurrence,
		LinkedProposalIDs:  sliceOrEmpty(proposalIDs),
		UpdatedAt:          item.UpdatedAt,
	}
}

func traceIDsFromSummaries(traces []events.TraceSummary) []string {
	out := make([]string, 0, len(traces))
	seen := map[string]struct{}{}
	for _, trace := range traces {
		if trace.TraceID == "" {
			continue
		}
		if _, ok := seen[trace.TraceID]; ok {
			continue
		}
		seen[trace.TraceID] = struct{}{}
		out = append(out, trace.TraceID)
	}
	return out
}
