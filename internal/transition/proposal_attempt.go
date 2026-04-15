package transition

import (
	"fmt"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
)

type AttemptPhaseState string

const (
	AttemptPhasePlanning         AttemptPhaseState = "planning"
	AttemptPhaseWorkspaceOpening AttemptPhaseState = "workspace_opening"
	AttemptPhaseImplementing     AttemptPhaseState = "implementing"
	AttemptPhaseValidating       AttemptPhaseState = "validating"
	AttemptPhasePROpen           AttemptPhaseState = "pr_open"
	AttemptPhaseRetryDeciding    AttemptPhaseState = "retry_deciding"
	AttemptPhaseTerminal         AttemptPhaseState = "terminal"
)

type AttemptPhaseCommandKind string

const (
	CommandLineActivated                 AttemptPhaseCommandKind = "line_activated"
	CommandAttemptPlannedWorkspace       AttemptPhaseCommandKind = "attempt_planned_workspace"
	CommandAttemptPlannedImplement       AttemptPhaseCommandKind = "attempt_planned_implement"
	CommandWorkspaceOpenDeferred         AttemptPhaseCommandKind = "workspace_open_deferred"
	CommandWorkspaceCompletedLegacy      AttemptPhaseCommandKind = "workspace_completed_legacy"
	CommandWorkspaceReady                AttemptPhaseCommandKind = "workspace_ready"
	CommandImplementationDeferred        AttemptPhaseCommandKind = "implementation_deferred"
	CommandImplementationCompleted       AttemptPhaseCommandKind = "implementation_completed"
	CommandImplementationFailedRetryable AttemptPhaseCommandKind = "implementation_failed_retryable"
	CommandImplementationFailedReview    AttemptPhaseCommandKind = "implementation_failed_review"
	CommandValidationCompleted           AttemptPhaseCommandKind = "validation_completed"
	CommandValidationFailedRetryable     AttemptPhaseCommandKind = "validation_failed_retryable"
	CommandValidationFailedReview        AttemptPhaseCommandKind = "validation_failed_review"
)

type AttemptSnapshot struct {
	ProposalStatus       review.ProposalStatus
	AttemptState         improvement.ChangeAttemptState
	CurrentOperationKind string
}

type AttemptPhaseDecision struct {
	TransitionDecision
	NextPhase           AttemptPhaseState
	ExpectedProposal    review.ProposalStatus
	AllowedProposalNext []review.ProposalStatus
	AllowedAttemptNext  []improvement.ChangeAttemptState
}

type attemptReducer func(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision

func deriveAttemptPhase(snapshot AttemptSnapshot) AttemptPhaseState {
	switch snapshot.CurrentOperationKind {
	case "line_activate", "attempt_plan":
		return AttemptPhasePlanning
	case "workspace_open":
		return AttemptPhaseWorkspaceOpening
	case "implement_attempt":
		return AttemptPhaseImplementing
	case "workspace_validate":
		return AttemptPhaseValidating
	case "pr_open":
		return AttemptPhasePROpen
	}
	switch snapshot.AttemptState {
	case improvement.AttemptStatePatchPlan, improvement.AttemptStateOverlayPlan:
		switch snapshot.ProposalStatus {
		case review.ProposalRepoChangeQueued:
			return AttemptPhaseWorkspaceOpening
		case review.ProposalRepoChangeRunning:
			return AttemptPhaseImplementing
		default:
			return AttemptPhasePlanning
		}
	case improvement.AttemptStatePatchGenerated, improvement.AttemptStateOverlayGenerated:
		return AttemptPhaseValidating
	case improvement.AttemptStateValidationRunning, improvement.AttemptStateOverlayValidating:
		if snapshot.ProposalStatus == review.ProposalValidationPending {
			return AttemptPhasePROpen
		}
		return AttemptPhaseValidating
	case improvement.AttemptStatePROpen, improvement.AttemptStateCIObserving:
		return AttemptPhasePROpen
	case improvement.AttemptStateRetryDeciding:
		return AttemptPhaseRetryDeciding
	case improvement.AttemptStateSandboxFailed,
		improvement.AttemptStateCIFailed,
		improvement.AttemptStateClosedUnmerged,
		improvement.AttemptStateMerged,
		improvement.AttemptStateNeedsReview,
		improvement.AttemptStateAbandoned,
		improvement.AttemptStateSuperseded,
		improvement.AttemptStateOverlayActive:
		return AttemptPhaseTerminal
	default:
		return AttemptPhasePlanning
	}
}

var attemptTransitionTable = map[AttemptPhaseState]map[AttemptPhaseCommandKind]attemptReducer{
	AttemptPhasePlanning: {
		CommandLineActivated:           reduceLineActivated,
		CommandAttemptPlannedWorkspace: reduceAttemptPlannedWorkspace,
		CommandAttemptPlannedImplement: reduceAttemptPlannedImplement,
	},
	AttemptPhaseWorkspaceOpening: {
		CommandWorkspaceOpenDeferred:    reduceWorkspaceOpenDeferred,
		CommandWorkspaceCompletedLegacy: reduceWorkspaceCompletedLegacy,
		CommandWorkspaceReady:           reduceWorkspaceReady,
	},
	AttemptPhaseImplementing: {
		CommandImplementationDeferred:        reduceImplementationDeferred,
		CommandImplementationCompleted:       reduceImplementationCompleted,
		CommandImplementationFailedRetryable: reduceImplementationFailedRetryable,
		CommandImplementationFailedReview:    reduceImplementationFailedReview,
	},
	AttemptPhaseValidating: {
		CommandValidationCompleted:       reduceValidationCompleted,
		CommandValidationFailedRetryable: reduceValidationFailedRetryable,
		CommandValidationFailedReview:    reduceValidationFailedReview,
	},
}

func ReduceAttempt(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	currentPhase := deriveAttemptPhase(snapshot)
	commandKind := AttemptPhaseCommandKind(command.CommandKind)
	commands, ok := attemptTransitionTable[currentPhase]
	if !ok {
		return rejectAttemptDecision(currentPhase, commandKind, "no reducer registered for current phase")
	}
	reducer, ok := commands[commandKind]
	if !ok {
		return rejectAttemptDecision(currentPhase, commandKind, "command is not allowed in the current phase")
	}
	return reducer(snapshot, command)
}

func reduceLineActivated(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "activated proposal line and queued attempt planning",
			Events: []DomainEventDescriptor{{
				Kind: "attempt_plan_queued",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectQueueAttemptPhase,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "attempt_plan"),
			}},
		},
		NextPhase:           AttemptPhasePlanning,
		AllowedProposalNext: []review.ProposalStatus{snapshot.ProposalStatus},
		AllowedAttemptNext:  []improvement.ChangeAttemptState{improvement.AttemptStatePatchPlan, improvement.AttemptStateOverlayPlan},
	}
}

func reduceAttemptPlannedWorkspace(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "attempt planning completed and queued workspace opening",
			Events: []DomainEventDescriptor{{
				Kind: "workspace_open_queued",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectOpenWorkspace,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "workspace_open"),
			}},
		},
		NextPhase:           AttemptPhaseWorkspaceOpening,
		AllowedProposalNext: []review.ProposalStatus{snapshot.ProposalStatus},
		AllowedAttemptNext:  []improvement.ChangeAttemptState{improvement.AttemptStatePatchPlan, improvement.AttemptStateOverlayPlan},
	}
}

func reduceAttemptPlannedImplement(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "attempt planning completed and queued implementation",
			Events: []DomainEventDescriptor{{
				Kind: "implementation_queued",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectInvokeRunner,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "implement_attempt"),
			}},
		},
		NextPhase:           AttemptPhaseImplementing,
		AllowedProposalNext: []review.ProposalStatus{snapshot.ProposalStatus},
		AllowedAttemptNext:  []improvement.ChangeAttemptState{improvement.AttemptStatePatchPlan, improvement.AttemptStateOverlayPlan},
	}
}

func reduceWorkspaceOpenDeferred(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "workspace is not ready yet and the same phase was rescheduled",
			Events: []DomainEventDescriptor{{
				Kind: "workspace_open_deferred",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectScheduleRetry,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "workspace_open_retry"),
			}},
		},
		NextPhase:           AttemptPhaseWorkspaceOpening,
		AllowedProposalNext: []review.ProposalStatus{review.ProposalRepoChangeQueued},
		AllowedAttemptNext: []improvement.ChangeAttemptState{
			improvement.AttemptStatePatchPlan,
			improvement.AttemptStateOverlayPlan,
		},
	}
}

func reduceWorkspaceCompletedLegacy(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "legacy workspace_open completion without a successor was accepted so reconciliation can recover it",
			Events: []DomainEventDescriptor{{
				Kind: "workspace_open_completed_without_successor",
			}},
		},
		NextPhase:           AttemptPhaseWorkspaceOpening,
		AllowedProposalNext: []review.ProposalStatus{review.ProposalApproved, review.ProposalRepoChangeQueued},
		AllowedAttemptNext: []improvement.ChangeAttemptState{
			improvement.AttemptStatePatchPlan,
			improvement.AttemptStateOverlayPlan,
		},
	}
}

func reduceWorkspaceReady(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "workspace is ready and implementation may start",
			Events: []DomainEventDescriptor{{
				Kind: "implementation_queued",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectInvokeRunner,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "implement_attempt"),
			}},
		},
		NextPhase:           AttemptPhaseImplementing,
		AllowedProposalNext: []review.ProposalStatus{review.ProposalRepoChangeRunning},
		AllowedAttemptNext: []improvement.ChangeAttemptState{
			improvement.AttemptStatePatchPlan,
			improvement.AttemptStateOverlayPlan,
		},
	}
}

func reduceImplementationDeferred(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "implementation was deferred and remains resumable",
			Events: []DomainEventDescriptor{{
				Kind: "implementation_deferred",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectScheduleRetry,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "implement_attempt_retry"),
			}},
		},
		NextPhase:           AttemptPhaseImplementing,
		AllowedProposalNext: []review.ProposalStatus{review.ProposalRepoChangeRunning, review.ProposalApproved},
		AllowedAttemptNext: []improvement.ChangeAttemptState{
			improvement.AttemptStatePatchPlan,
			improvement.AttemptStateOverlayPlan,
		},
	}
}

func reduceImplementationCompleted(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "implementation completed and validation was queued",
			Events: []DomainEventDescriptor{{
				Kind: "workspace_validate_queued",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectWorkspaceValidate,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "workspace_validate"),
			}},
		},
		NextPhase:           AttemptPhaseValidating,
		AllowedProposalNext: []review.ProposalStatus{review.ProposalRepoChangeRunning, review.ProposalApproved},
		AllowedAttemptNext: []improvement.ChangeAttemptState{
			improvement.AttemptStatePatchGenerated,
			improvement.AttemptStateOverlayGenerated,
		},
	}
}

func reduceImplementationFailedRetryable(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "implementation failed with a retryable outcome",
			Events: []DomainEventDescriptor{{
				Kind: "attempt_failed_retryable",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectRefreshProjection,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "retryable_failure"),
			}},
		},
		NextPhase:           AttemptPhaseRetryDeciding,
		AllowedProposalNext: []review.ProposalStatus{review.ProposalApproved, review.ProposalRepoChangeRunning, review.ProposalFailedValidation},
		AllowedAttemptNext:  []improvement.ChangeAttemptState{improvement.AttemptStateSandboxFailed, improvement.AttemptStateCIFailed, improvement.AttemptStateClosedUnmerged, improvement.AttemptStateNeedsReview},
	}
}

func reduceImplementationFailedReview(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "implementation failed and needs line review",
			Events: []DomainEventDescriptor{{
				Kind: "attempt_failed_needs_review",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectRefreshProjection,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "needs_review"),
			}},
		},
		NextPhase:           AttemptPhaseTerminal,
		AllowedProposalNext: []review.ProposalStatus{review.ProposalPendingReview},
		AllowedAttemptNext:  []improvement.ChangeAttemptState{improvement.AttemptStateNeedsReview},
	}
}

func reduceValidationCompleted(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "validation succeeded and the governed PR open was queued",
			Events: []DomainEventDescriptor{{
				Kind: "draft_pr_open_queued",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectOpenDraftPR,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "pr_open"),
			}},
		},
		NextPhase:           AttemptPhasePROpen,
		AllowedProposalNext: []review.ProposalStatus{review.ProposalValidationPending},
		AllowedAttemptNext: []improvement.ChangeAttemptState{
			improvement.AttemptStateValidationRunning,
			improvement.AttemptStateOverlayValidating,
		},
	}
}

func reduceValidationFailedRetryable(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "validation failed with a retryable outcome",
			Events: []DomainEventDescriptor{{
				Kind: "validation_failed_retryable",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectRefreshProjection,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "validation_failed"),
			}},
		},
		NextPhase:           AttemptPhaseRetryDeciding,
		AllowedProposalNext: []review.ProposalStatus{review.ProposalApproved, review.ProposalFailedValidation},
		AllowedAttemptNext:  []improvement.ChangeAttemptState{improvement.AttemptStateSandboxFailed, improvement.AttemptStateCIFailed, improvement.AttemptStateClosedUnmerged, improvement.AttemptStateNeedsReview},
	}
}

func reduceValidationFailedReview(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "validation failed and line review is required",
			Events: []DomainEventDescriptor{{
				Kind: "validation_failed_needs_review",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectRefreshProjection,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "validation_needs_review"),
			}},
		},
		NextPhase:           AttemptPhaseTerminal,
		AllowedProposalNext: []review.ProposalStatus{review.ProposalPendingReview},
		AllowedAttemptNext:  []improvement.ChangeAttemptState{improvement.AttemptStateNeedsReview},
	}
}

func rejectAttemptDecision(phase AttemptPhaseState, command AttemptPhaseCommandKind, reason string) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionReject,
			Reason:       fmt.Sprintf("%s: phase=%s command=%s", reason, phase, command),
		},
		NextPhase: AttemptPhaseTerminal,
	}
}

func effectKey(command CommandEnvelope, suffix string) string {
	return fmt.Sprintf("%s:%s", command.CommandID, suffix)
}
