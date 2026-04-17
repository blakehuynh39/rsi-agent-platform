package improvementplane

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/events"
	"github.com/piplabs/rsi-agent-platform/internal/harness"
	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/operation"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/sandbox"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

const (
	proposalOperationLineActivate      = "line_activate"
	proposalOperationAttemptPlan       = "attempt_plan"
	proposalOperationWorkspaceOpen     = "workspace_open"
	proposalOperationImplementAttempt  = "implement_attempt"
	proposalOperationWorkspaceValidate = "workspace_validate"
)

var errProposalAttemptNotMaterialized = errors.New("proposal attempt not materialized")
var errProposalPhaseHandled = errors.New("proposal phase handled")

type legacyWorkItem struct {
	ID             string
	Queue          queue.QueueName
	Kind           string
	Status         queue.WorkItemStatus
	TraceID        string
	ThreadKey      string
	ConversationID string
	CaseID         string
	TriggerEventID string
	ProposalID     string
	RequestedBy    string
	ApprovalMode   string
	OperationID    string
	WorkflowID     string
	IngestionID    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Payload        map[string]any
}

func registerImprovementOperationForTest(_ any, op operation.Execution, item legacyWorkItem) (legacyWorkItem, error) {
	item.OperationID = firstNonEmpty(strings.TrimSpace(op.ID), fmt.Sprintf("op-test-%s-%s", strings.TrimSpace(op.ScopeID), strings.TrimSpace(op.OperationKind)))
	return item, nil
}

func proposalAttemptPhaseWork(cfg config.Config, proposal review.Proposal, trace events.Trace, attempt improvement.ChangeAttempt, operationKind string, payload map[string]any) (operation.Execution, legacyWorkItem) {
	if payload == nil {
		payload = map[string]any{}
	}
	payload["attempt_id"] = attempt.ID
	return operation.Execution{
			ScopeKind:     operation.ScopeAttempt,
			ScopeID:       attempt.ID,
			OperationKind: operationKind,
			OperationKey:  operationKind,
			Status:        operation.StatusQueued,
			Queue:         queue.ProposalQueue,
			RequestedBy:   cfg.ServiceName,
			TraceID:       trace.Summary.TraceID,
			ProposalID:    proposal.ID,
			AttemptID:     attempt.ID,
		}, legacyWorkItem{
			Kind:       operationKind,
			TraceID:    trace.Summary.TraceID,
			ProposalID: proposal.ID,
			Payload:    payload,
		}
}

func ensureProposalAttempt(cfg config.Config, store storepkg.Store, proposal review.Proposal, sourceTrace events.Trace, item legacyWorkItem) (improvement.ChangeAttempt, events.Trace, error) {
	_ = cfg
	_ = sourceTrace
	_ = item
	if strings.TrimSpace(proposal.CurrentAttemptID) != "" {
		if existing, ok := store.GetChangeAttempt(proposal.CurrentAttemptID); ok && !isAttemptTerminal(existing.State) {
			trace, ok := store.GetTrace(existing.AttemptTraceID)
			if !ok {
				return improvement.ChangeAttempt{}, events.Trace{}, fmt.Errorf("attempt trace %s not found", existing.AttemptTraceID)
			}
			return existing, trace, nil
		}
	}
	if latest, ok := latestAttemptForProposal(store, proposal.ID); ok && !isAttemptTerminal(latest.State) {
		trace, ok := store.GetTrace(latest.AttemptTraceID)
		if !ok {
			return improvement.ChangeAttempt{}, events.Trace{}, fmt.Errorf("attempt trace %s not found", latest.AttemptTraceID)
		}
		return latest, trace, nil
	}
	return improvement.ChangeAttempt{}, events.Trace{}, errProposalAttemptNotMaterialized
}

func prepareProposalAttempt(cfg config.Config, proposal review.Proposal, sourceTrace events.Trace, item legacyWorkItem) (improvement.ChangeAttempt, storepkg.DerivedTraceRequest) {
	nextNumber := int(intFromAny(item.Payload["attempt_number"]))
	parentAttemptID := stringValue(item.Payload["parent_attempt"])
	if nextNumber <= 0 {
		nextNumber = proposal.AttemptCount + 1
	}
	if nextNumber <= 0 {
		nextNumber = 1
	}
	trigger := improvement.AttemptTriggerProposalApproved
	if raw := strings.TrimSpace(stringValue(item.Payload["trigger"])); raw != "" {
		trigger = improvement.ChangeAttemptTrigger(raw)
	}
	now := time.Now().UTC()
	attempt := improvement.ChangeAttempt{
		ID:              fmt.Sprintf("attempt-%s-%02d", strings.ReplaceAll(strings.TrimSpace(proposal.ID), "/", "-"), nextNumber),
		ProposalID:      proposal.ID,
		CandidateKey:    proposal.CandidateKey,
		AttemptNumber:   nextNumber,
		TargetLayer:     proposal.TargetLayer,
		TargetKind:      proposal.TargetKind,
		TargetRef:       proposal.TargetRef,
		Trigger:         trigger,
		State:           improvement.AttemptStatePatchPlan,
		ParentAttemptID: parentAttemptID,
		BranchName:      buildAttemptBranchName(proposal.ID, nextNumber),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if proposal.RecommendedInterventionKind == review.InterventionHarnessOverlay {
		attempt.State = improvement.AttemptStateOverlayPlan
	}
	description := fmt.Sprintf("Queued remediation attempt %d for proposal %s triggered by %s.", nextNumber, proposal.ID, trigger)
	traceReq := storepkg.DerivedTraceRequest{
		SourceTraceID:  sourceTrace.Summary.TraceID,
		ProposalID:     proposal.ID,
		AttemptID:      attempt.ID,
		ConversationID: proposal.ConversationID,
		CaseID:         proposal.CaseID,
		ThreadKey:      sourceTrace.Summary.ThreadKey,
		WorkflowKind:   "proposal_attempt",
		RequestedBy:    cfg.ServiceName,
		Description:    description,
		TriggerEventID: sourceTrace.Summary.TriggerEventID,
		IngestionID:    sourceTrace.Summary.IngestionID,
		CreatedAt:      now,
	}
	return attempt, traceReq
}

func sandboxObservationRequestForTest(item legacyWorkItem) sandboxObservationRequest {
	return sandboxObservationRequest{
		EffectID:    item.ID,
		OperationID: item.OperationID,
		ProposalID:  item.ProposalID,
		AttemptID:   stringValue(item.Payload["attempt_id"]),
		TraceID:     item.TraceID,
		JobID:       stringValue(item.Payload["job_id"]),
		JobName:     firstNonEmpty(stringValue(item.Payload["job_name"]), stringValue(item.Payload["sandbox_job_name"])),
		Namespace:   firstNonEmpty(stringValue(item.Payload["namespace"]), stringValue(item.Payload["sandbox_namespace"])),
		Repo:        stringValue(item.Payload["repo"]),
		BranchName:  stringValue(item.Payload["branch_name"]),
		BaseRef:     stringValue(item.Payload["base_ref"]),
	}
}

func draftPROpenRequestForTest(item legacyWorkItem) draftPROpenRequest {
	return draftPROpenRequest{
		EffectID:    item.ID,
		OperationID: item.OperationID,
		ProposalID:  item.ProposalID,
		AttemptID:   stringValue(item.Payload["attempt_id"]),
		TraceID:     item.TraceID,
		JobID:       stringValue(item.Payload["job_id"]),
		Repo:        stringValue(item.Payload["repo"]),
		BranchName:  stringValue(item.Payload["branch_name"]),
		BaseRef:     stringValue(item.Payload["base_ref"]),
		Title:       stringValue(item.Payload["title"]),
		Body:        stringValue(item.Payload["body"]),
	}
}

func processProposalItem(cfg config.Config, store *storepkg.MemoryStore, runner runnerExecutor, toolClient toolExecutor, launcher sandbox.Launcher, launcherErr error, item legacyWorkItem) error {
	attemptID := firstNonEmpty(stringValue(item.Payload["attempt_id"]), stringValue(item.Payload["current_attempt_id"]))
	if attemptID == "" {
		return fmt.Errorf("proposal item %s missing attempt_id", item.ID)
	}
	proposal, ok := findProposal(store.ListProposals(), item.ProposalID)
	if !ok {
		return fmt.Errorf("proposal %s not found", item.ProposalID)
	}
	switch item.Kind {
	case proposalOperationAttemptPlan:
		attempt, ok := store.GetChangeAttempt(attemptID)
		if !ok {
			return fmt.Errorf("attempt %s not found", attemptID)
		}
		commandKind := transition.CommandAttemptPlannedWorkspace
		if proposal.RecommendedInterventionKind == review.InterventionHarnessOverlay || proposal.TargetLayer == harness.TargetLayerHarnessOverlay {
			commandKind = transition.CommandAttemptPlannedImplement
		}
		payload := clonePayload(item.Payload)
		payload["operation_id"] = item.OperationID
		if err := submitAttemptCommand(store, attempt, commandKind, cfg.ServiceName, time.Now().UTC(), payload); err != nil {
			return err
		}
		return errProposalPhaseHandled
	case proposalOperationWorkspaceOpen, proposalOperationImplementAttempt, proposalOperationWorkspaceValidate:
	default:
		return fmt.Errorf("unsupported proposal item kind %s", item.Kind)
	}

	effectKind := transition.EffectKind("")
	switch item.Kind {
	case proposalOperationWorkspaceOpen:
		effectKind = transition.EffectOpenWorkspace
	case proposalOperationImplementAttempt:
		effectKind = transition.EffectInvokeRunner
	case proposalOperationWorkspaceValidate:
		effectKind = transition.EffectWorkspaceValidate
	}

	payload := clonePayload(item.Payload)
	payload["attempt_id"] = attemptID
	payload["operation_id"] = item.OperationID
	payload["trace_id"] = item.TraceID
	payload["proposal_id"] = item.ProposalID
	if item.WorkflowID != "" {
		payload["workflow_id"] = item.WorkflowID
	}
	if item.IngestionID != "" {
		payload["ingestion_id"] = item.IngestionID
	}
	if item.ConversationID != "" {
		payload["conversation_id"] = item.ConversationID
	}
	if item.CaseID != "" {
		payload["case_id"] = item.CaseID
	}
	if item.TriggerEventID != "" {
		payload["trigger_event_id"] = item.TriggerEventID
	}
	if item.ThreadKey != "" {
		payload["thread_key"] = item.ThreadKey
	}
	effect := transition.EffectExecution{
		ID:          firstNonEmpty(item.ID, fmt.Sprintf("eff-test-%s-%s", attemptID, item.Kind)),
		MachineKind: transition.MachineAttempt,
		AggregateID: attemptID,
		AttemptID:   attemptID,
		EffectKind:  effectKind,
		Status:      transition.EffectQueued,
		Payload:     payload,
	}
	if err := processImprovementEffect(cfg, store, nil, runner, toolClient, launcher, launcherErr, effect); err != nil {
		return err
	}
	return errProposalPhaseHandled
}
