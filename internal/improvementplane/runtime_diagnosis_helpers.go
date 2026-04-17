package improvementplane

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

func runtimeDiagnosisAggregateID(candidateKey string, repo string) string {
	candidateKey = strings.TrimSpace(candidateKey)
	repo = strings.TrimSpace(repo)
	if repo == "" {
		return candidateKey
	}
	return fmt.Sprintf("%s|%s", candidateKey, repo)
}

func runtimeDiagnosisSessionScopeID(candidateKey string, repo string) string {
	return runtimeDiagnosisAggregateID(candidateKey, repo)
}

func runtimeDiagnosisTargetRepo(cfg config.Config, candidate improvement.Candidate) string {
	return improvementTargetRepo(cfg, candidate.TargetLayer, candidate.TargetKind, candidate.TargetRef)
}

func runtimeDiagnosisCandidates(items []improvement.Candidate) []improvement.Candidate {
	out := make([]improvement.Candidate, 0, len(items))
	for _, item := range items {
		if !shouldQueueRuntimeDiagnosis(item) {
			continue
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PriorityScore == out[j].PriorityScore {
			return out[i].UpdatedAt.After(out[j].UpdatedAt)
		}
		return out[i].PriorityScore > out[j].PriorityScore
	})
	return out
}

func shouldQueueRuntimeDiagnosis(candidate improvement.Candidate) bool {
	if strings.TrimSpace(candidate.CandidateKey) == "" {
		return false
	}
	if candidate.Status != improvement.CandidateQueued {
		return false
	}
	if strings.TrimSpace(firstNonEmpty(candidate.LatestTraceID, candidate.OriginTraceID)) == "" {
		return false
	}
	if candidate.InterventionType == "policy_or_runtime_fix" {
		return true
	}
	switch strings.TrimSpace(candidate.Subsystem) {
	case "runner", "control-plane", "improvement-plane", "tool-gateway", "shared-store", "platform":
		return true
	default:
		return false
	}
}

func queueRuntimeDiagnoses(cfg config.Config, store storepkg.Store, actor string, occurredAt time.Time) (int, error) {
	if !cfg.RuntimeDiagnosisEnabled {
		return 0, nil
	}
	queued := 0
	for _, candidate := range runtimeDiagnosisCandidates(store.ListCandidates()) {
		repo := runtimeDiagnosisTargetRepo(cfg, candidate)
		aggregateID := runtimeDiagnosisAggregateID(candidate.CandidateKey, repo)
		if diagnosis, ok := store.GetRuntimeDiagnosis(aggregateID); ok && !runtimeDiagnosisNeedsRefresh(diagnosis, candidate) {
			continue
		}
		receipt, err := submitRuntimeDiagnosisCommand(
			store,
			aggregateID,
			transition.CommandRuntimeDiagnosisQueue,
			actor,
			occurredAt,
			fmt.Sprintf("cmd-runtime-diagnosis:queue:%s:%d", aggregateID, occurredAt.UnixNano()),
			map[string]any{
				"candidate_key":      candidate.CandidateKey,
				"repo":               repo,
				"conversation_id":    candidate.ConversationID,
				"case_id":            candidate.CaseID,
				"latest_trace_id":    firstNonEmpty(candidate.LatestTraceID, candidate.OriginTraceID),
				"session_scope_kind": "runtime_diagnosis",
				"session_scope_id":   runtimeDiagnosisSessionScopeID(candidate.CandidateKey, repo),
			},
		)
		if err != nil {
			return queued, err
		}
		if receipt.DecisionKind == transition.DecisionAdvance {
			queued++
		}
	}
	return queued, nil
}

func runtimeDiagnosisNeedsRefresh(diagnosis improvement.RuntimeDiagnosis, candidate improvement.Candidate) bool {
	latestTraceID := strings.TrimSpace(firstNonEmpty(candidate.LatestTraceID, candidate.OriginTraceID))
	if latestTraceID == "" {
		return false
	}
	if strings.TrimSpace(diagnosis.LatestTraceID) == latestTraceID {
		switch diagnosis.Status {
		case improvement.RuntimeDiagnosisQueued,
			improvement.RuntimeDiagnosisInvestigating,
			improvement.RuntimeDiagnosisGrounded,
			improvement.RuntimeDiagnosisPromoted,
			improvement.RuntimeDiagnosisNeedsEvidence,
			improvement.RuntimeDiagnosisClosed:
			return false
		}
	}
	return true
}

func submitRuntimeDiagnosisCommand(store storepkg.Store, aggregateID string, kind transition.RuntimeDiagnosisCommandKind, actor string, occurredAt time.Time, commandID string, payload map[string]any) (transition.CommandReceipt, error) {
	receipt, err := store.SubmitCommand(transition.CommandEnvelope{
		MachineKind: transition.MachineRuntimeDiagnosis,
		AggregateID: strings.TrimSpace(aggregateID),
		CommandKind: string(kind),
		CommandID:   strings.TrimSpace(commandID),
		Actor:       actor,
		OccurredAt:  occurredAt,
		Payload:     payload,
	})
	if err != nil {
		return transition.CommandReceipt{}, err
	}
	if receipt.DecisionKind == transition.DecisionReject {
		return transition.CommandReceipt{}, errors.New(receipt.Reason)
	}
	return receipt, nil
}

func findCandidate(items []improvement.Candidate, candidateKey string) (improvement.Candidate, bool) {
	candidateKey = strings.TrimSpace(candidateKey)
	for _, item := range items {
		if strings.TrimSpace(item.CandidateKey) == candidateKey {
			return item, true
		}
	}
	return improvement.Candidate{}, false
}

func runtimeDiagnosesForConversation(items []improvement.RuntimeDiagnosis, conversationID string) []improvement.RuntimeDiagnosis {
	out := make([]improvement.RuntimeDiagnosis, 0)
	conversationID = strings.TrimSpace(conversationID)
	for _, item := range items {
		if strings.TrimSpace(item.ConversationID) != conversationID {
			continue
		}
		out = append(out, item)
	}
	sortRuntimeDiagnoses(out)
	return out
}

func runtimeDiagnosesForCase(items []improvement.RuntimeDiagnosis, caseID string) []improvement.RuntimeDiagnosis {
	out := make([]improvement.RuntimeDiagnosis, 0)
	caseID = strings.TrimSpace(caseID)
	for _, item := range items {
		if strings.TrimSpace(item.CaseID) != caseID {
			continue
		}
		out = append(out, item)
	}
	sortRuntimeDiagnoses(out)
	return out
}

func runtimeDiagnosesForCandidate(items []improvement.RuntimeDiagnosis, candidateKey string) []improvement.RuntimeDiagnosis {
	out := make([]improvement.RuntimeDiagnosis, 0)
	candidateKey = strings.TrimSpace(candidateKey)
	for _, item := range items {
		if strings.TrimSpace(item.CandidateKey) != candidateKey {
			continue
		}
		out = append(out, item)
	}
	sortRuntimeDiagnoses(out)
	return out
}

func runtimeDiagnosesForTrace(items []improvement.RuntimeDiagnosis, trace events.Trace, candidateKey string) []improvement.RuntimeDiagnosis {
	out := make([]improvement.RuntimeDiagnosis, 0)
	traceID := strings.TrimSpace(trace.Summary.TraceID)
	caseID := strings.TrimSpace(trace.Summary.CaseID)
	candidateKey = strings.TrimSpace(candidateKey)
	seen := map[string]struct{}{}
	for _, item := range items {
		switch {
		case traceID != "" && strings.TrimSpace(item.LatestTraceID) == traceID:
		case caseID != "" && strings.TrimSpace(item.CaseID) == caseID:
		case candidateKey != "" && strings.TrimSpace(item.CandidateKey) == candidateKey:
		default:
			continue
		}
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		out = append(out, item)
	}
	sortRuntimeDiagnoses(out)
	return out
}

func latestRuntimeDiagnosisForCandidate(items []improvement.RuntimeDiagnosis, candidateKey string) *improvement.RuntimeDiagnosis {
	matches := runtimeDiagnosesForCandidate(items, candidateKey)
	if len(matches) == 0 {
		return nil
	}
	item := matches[0]
	return &item
}

func runtimeDiagnosisContextRefs(items []improvement.RuntimeDiagnosis, candidateKey string) []clients.RunnerContextRef {
	diagnoses := runtimeDiagnosesForCandidate(items, candidateKey)
	refs := make([]clients.RunnerContextRef, 0, len(diagnoses))
	for _, item := range diagnoses {
		refs = append(refs, clients.RunnerContextRef{
			Kind:           "runtime_diagnosis",
			Ref:            item.ID,
			Status:         string(item.Status),
			Summary:        item.Summary,
			Subsystem:      item.Subsystem,
			FailureMode:    item.FailureMode,
			TargetSurface:  item.TargetSurface,
			ValidationPlan: item.ValidationPlan,
		})
		if len(refs) == 4 {
			break
		}
	}
	return refs
}

func sortRuntimeDiagnoses(items []improvement.RuntimeDiagnosis) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].CreatedAt.After(items[j].CreatedAt)
		}
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
}

func runtimeDiagnosisRunnerTools(logFallbackEnabled bool) []string {
	tools := []string{
		"todo",
		"session_search",
		"rsi.trace_context",
		"rsi.workflow_context",
		"rsi.runner_execution",
		"repo.read_file",
		"repo.search",
	}
	if logFallbackEnabled {
		tools = append(tools, "kubernetes.logs")
	}
	return tools
}

func runtimeDiagnosisTargetLayer(candidate improvement.Candidate) harness.TargetLayer {
	if candidate.TargetLayer != "" {
		return candidate.TargetLayer
	}
	return harness.TargetLayerRepoChange
}
