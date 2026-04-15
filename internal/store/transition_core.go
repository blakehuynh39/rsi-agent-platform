package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/operation"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
	"github.com/piplabs/rsi-agent-platform/internal/review"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

type transitionPersistBundle struct {
	Events  []transition.DomainEvent
	Effects []transition.EffectExecution
}

func persistDomainEvents(tx *sql.Tx, items []transition.DomainEvent) error {
	for _, item := range items {
		if _, err := tx.Exec(`insert into domain_event (id, machine_kind, aggregate_id, aggregate_version, event_kind, command_id, causation_id, payload, created_at)
			values ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9)
			on conflict (id) do nothing`,
			item.ID,
			string(item.MachineKind),
			item.AggregateID,
			item.AggregateVersion,
			item.EventKind,
			firstNonEmpty(item.CommandID),
			firstNonEmpty(item.CausationID),
			jsonString(item.Payload),
			item.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func persistEffectExecutions(tx *sql.Tx, items []transition.EffectExecution) error {
	for _, item := range items {
		if _, err := tx.Exec(`insert into effect_execution (id, machine_kind, aggregate_id, attempt_id, effect_kind, status, holder, idempotency_key, payload, result_ref, last_error, retry_count, created_at, updated_at, started_at, lease_expires_at, completed_at)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10,$11,$12,$13,$14,$15,$16,$17)
			on conflict (idempotency_key) do update set
				status = excluded.status,
				holder = excluded.holder,
				payload = excluded.payload,
				result_ref = excluded.result_ref,
				last_error = excluded.last_error,
				retry_count = excluded.retry_count,
				updated_at = excluded.updated_at,
				started_at = excluded.started_at,
				lease_expires_at = excluded.lease_expires_at,
				completed_at = excluded.completed_at`,
			item.ID,
			string(item.MachineKind),
			item.AggregateID,
			firstNonEmpty(item.AttemptID),
			string(item.EffectKind),
			string(item.Status),
			firstNonEmpty(item.Holder),
			item.IdempotencyKey,
			jsonString(item.Payload),
			firstNonEmpty(item.ResultRef),
			firstNonEmpty(item.LastError),
			item.RetryCount,
			item.CreatedAt,
			item.UpdatedAt,
			nullTime(item.StartedAt),
			nullTime(item.LeaseExpiresAt),
			nullTime(item.CompletedAt),
		); err != nil {
			return err
		}
	}
	return nil
}

func persistCommandReceipts(tx *sql.Tx, items []transition.CommandReceipt) error {
	for _, item := range items {
		if _, err := tx.Exec(`insert into command_receipt (command_id, machine_kind, aggregate_id, command_kind, causation_id, actor, decision_kind, reason, aggregate_version, result_ref, created_at, updated_at)
			values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			on conflict (command_id) do update set
				machine_kind = excluded.machine_kind,
				aggregate_id = excluded.aggregate_id,
				command_kind = excluded.command_kind,
				causation_id = excluded.causation_id,
				actor = excluded.actor,
				decision_kind = excluded.decision_kind,
				reason = excluded.reason,
				aggregate_version = excluded.aggregate_version,
				result_ref = excluded.result_ref,
				updated_at = excluded.updated_at`,
			item.CommandID,
			string(item.MachineKind),
			item.AggregateID,
			item.CommandKind,
			firstNonEmpty(item.CausationID),
			firstNonEmpty(item.Actor),
			string(item.DecisionKind),
			firstNonEmpty(item.Reason),
			item.AggregateVersion,
			firstNonEmpty(item.ResultRef),
			item.CreatedAt,
			item.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

func proposalTransitionSnapshots(proposalID string, attemptID string, proposals map[string]review.Proposal, attempts map[string]improvement.ChangeAttempt) (*review.Proposal, *improvement.ChangeAttempt, error) {
	var proposal *review.Proposal
	if trimmed := strings.TrimSpace(proposalID); trimmed != "" {
		item, ok := proposals[trimmed]
		if !ok {
			return nil, nil, fmt.Errorf("proposal %s not found for transition bundle", trimmed)
		}
		copy := item
		proposal = &copy
	}
	var attempt *improvement.ChangeAttempt
	if trimmed := strings.TrimSpace(attemptID); trimmed != "" {
		item, ok := attempts[trimmed]
		if !ok {
			return proposal, nil, fmt.Errorf("attempt %s not found for transition bundle", trimmed)
		}
		copy := item
		attempt = &copy
	}
	return proposal, attempt, nil
}

func proposalPhaseAdvanceCommand(currentKind string, nextKind string) (transition.AttemptPhaseCommandKind, error) {
	switch strings.TrimSpace(currentKind) {
	case "line_activate":
		if strings.TrimSpace(nextKind) == "attempt_plan" {
			return transition.CommandLineActivated, nil
		}
	case "attempt_plan":
		switch strings.TrimSpace(nextKind) {
		case "workspace_open":
			return transition.CommandAttemptPlannedWorkspace, nil
		case "implement_attempt":
			return transition.CommandAttemptPlannedImplement, nil
		}
	case "workspace_open":
		if strings.TrimSpace(nextKind) == "" {
			return transition.CommandWorkspaceCompletedLegacy, nil
		}
		if strings.TrimSpace(nextKind) == "implement_attempt" {
			return transition.CommandWorkspaceReady, nil
		}
	case "implement_attempt":
		if strings.TrimSpace(nextKind) == "workspace_validate" {
			return transition.CommandImplementationCompleted, nil
		}
	case "workspace_validate":
		if strings.TrimSpace(nextKind) == "pr_open" {
			return transition.CommandValidationCompleted, nil
		}
	}
	return "", fmt.Errorf("unsupported proposal phase advance transition %q -> %q", currentKind, nextKind)
}

func proposalPhaseDeferCommand(currentKind string) (transition.AttemptPhaseCommandKind, error) {
	switch strings.TrimSpace(currentKind) {
	case "workspace_open":
		return transition.CommandWorkspaceOpenDeferred, nil
	case "implement_attempt":
		return transition.CommandImplementationDeferred, nil
	default:
		return "", fmt.Errorf("unsupported proposal phase defer transition %q", currentKind)
	}
}

func proposalPhaseFailureCommand(currentKind string, proposal *review.Proposal, _ *improvement.ChangeAttempt) (transition.AttemptPhaseCommandKind, error) {
	needsReview := proposal != nil && proposal.Status == review.ProposalPendingReview
	switch strings.TrimSpace(currentKind) {
	case "implement_attempt":
		if needsReview {
			return transition.CommandImplementationFailedReview, nil
		}
		return transition.CommandImplementationFailedRetryable, nil
	case "workspace_validate":
		if needsReview {
			return transition.CommandValidationFailedReview, nil
		}
		return transition.CommandValidationFailedRetryable, nil
	default:
		return "", fmt.Errorf("unsupported proposal phase failure transition %q", currentKind)
	}
}

func validateAttemptDecision(decision transition.AttemptPhaseDecision, proposal *review.Proposal, attempt *improvement.ChangeAttempt) error {
	if decision.DecisionKind != transition.DecisionAdvance {
		return fmt.Errorf("transition reducer rejected proposal attempt command: %s", decision.Reason)
	}
	if proposal != nil {
		currentProposalStatus := transitionProposalStatusOr(proposal)
		if len(decision.AllowedProposalNext) > 0 {
			for _, allowed := range decision.AllowedProposalNext {
				if currentProposalStatus == allowed {
					goto attemptCheck
				}
			}
			return fmt.Errorf("proposal transition mismatch: got %s want one of %v", currentProposalStatus, decision.AllowedProposalNext)
		}
		if decision.ExpectedProposal != "" && currentProposalStatus != decision.ExpectedProposal {
			return fmt.Errorf("proposal transition mismatch: got %s want %s", currentProposalStatus, decision.ExpectedProposal)
		}
	}
attemptCheck:
	if attempt != nil && len(decision.AllowedAttemptNext) > 0 {
		currentAttemptState := transitionAttemptStateOr(attempt)
		for _, allowed := range decision.AllowedAttemptNext {
			if currentAttemptState == allowed {
				return nil
			}
		}
		return fmt.Errorf("attempt transition mismatch: got %s want one of %v", currentAttemptState, decision.AllowedAttemptNext)
	}
	return nil
}

func buildTransitionBundle(now time.Time, command transition.CommandEnvelope, decision transition.AttemptPhaseDecision, proposal *review.Proposal, attempt *improvement.ChangeAttempt, currentItem queue.WorkItem, currentOp operation.Execution, nextItem *queue.WorkItem, nextOp *operation.Execution, lastError string) (transitionPersistBundle, error) {
	if err := validateAttemptDecision(decision, proposal, attempt); err != nil {
		return transitionPersistBundle{}, err
	}
	bundle := transitionPersistBundle{}
	basePayload := map[string]any{
		"command_kind":         command.CommandKind,
		"current_operation":    currentOp.OperationKind,
		"current_operation_id": currentOp.ID,
		"current_work_item_id": currentItem.ID,
		"reason":               decision.Reason,
	}
	if nextOp != nil {
		basePayload["next_operation"] = nextOp.OperationKind
		basePayload["next_operation_id"] = nextOp.ID
	}
	if nextItem != nil {
		basePayload["next_work_item_id"] = nextItem.ID
		basePayload["next_queue"] = string(nextItem.Queue)
	}
	if proposal != nil {
		bundle.Events = append(bundle.Events, transition.DomainEvent{
			ID:               fmt.Sprintf("evt:%s:proposal", command.CommandID),
			MachineKind:      transition.MachineProposalLine,
			AggregateID:      proposal.ID,
			AggregateVersion: proposal.Version,
			EventKind:        firstEventKind(decision.Events, "proposal_transition_applied"),
			CommandID:        command.CommandID,
			CausationID:      command.CausationID,
			Payload:          cloneMetadata(basePayload),
			CreatedAt:        now,
		})
	}
	if attempt != nil {
		payload := cloneMetadata(basePayload)
		payload["attempt_state"] = string(attempt.State)
		bundle.Events = append(bundle.Events, transition.DomainEvent{
			ID:               fmt.Sprintf("evt:%s:attempt", command.CommandID),
			MachineKind:      transition.MachineAttempt,
			AggregateID:      attempt.ID,
			AggregateVersion: attempt.Version,
			EventKind:        firstEventKind(decision.Events, "attempt_transition_applied"),
			CommandID:        command.CommandID,
			CausationID:      command.CausationID,
			Payload:          payload,
			CreatedAt:        now,
		})
	}
	for idx, effect := range decision.Effects {
		payload := cloneMetadata(effect.Payload)
		if payload == nil {
			payload = map[string]any{}
		}
		payload["command_kind"] = command.CommandKind
		payload["current_operation"] = currentOp.OperationKind
		if nextOp != nil {
			payload["operation_id"] = nextOp.ID
			payload["operation_kind"] = nextOp.OperationKind
		} else {
			payload["operation_id"] = currentOp.ID
			payload["operation_kind"] = currentOp.OperationKind
		}
		if nextItem != nil {
			payload["work_item_id"] = nextItem.ID
			payload["queue"] = string(nextItem.Queue)
		} else {
			payload["work_item_id"] = currentItem.ID
			payload["queue"] = string(currentItem.Queue)
		}
		aggregateID := command.AggregateID
		if aggregateID == "" && attempt != nil {
			aggregateID = attempt.ID
		}
		bundle.Effects = append(bundle.Effects, transition.EffectExecution{
			ID:             nextUUID("eff"),
			MachineKind:    transition.MachineAttempt,
			AggregateID:    aggregateID,
			AttemptID:      firstNonEmpty(command.AggregateID, attemptIDOrEmpty(attempt)),
			EffectKind:     effect.Kind,
			Status:         effect.Status,
			IdempotencyKey: fmt.Sprintf("%s:%s:%d", aggregateID, effect.IdempotencyKey, idx),
			Payload:        payload,
			ResultRef:      firstNonEmpty(resultRef(nextItem, currentItem), resultRefOp(nextOp, currentOp)),
			LastError:      strings.TrimSpace(lastError),
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}
	return bundle, nil
}

func firstEventKind(items []transition.DomainEventDescriptor, fallback string) string {
	if len(items) == 0 {
		return fallback
	}
	return firstNonEmpty(items[0].Kind, fallback)
}

func resultRef(nextItem *queue.WorkItem, currentItem queue.WorkItem) string {
	if nextItem != nil {
		return nextItem.ID
	}
	return currentItem.ID
}

func resultRefOp(nextOp *operation.Execution, currentOp operation.Execution) string {
	if nextOp != nil {
		return nextOp.ID
	}
	return currentOp.ID
}

func attemptIDOrEmpty(attempt *improvement.ChangeAttempt) string {
	if attempt == nil {
		return ""
	}
	return attempt.ID
}
