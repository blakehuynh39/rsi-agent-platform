package transition

import (
	"fmt"
)

type ProposalStatus string

const (
	ProposalQueuedForPromotion ProposalStatus = "queued_for_promotion"
	ProposalPendingReview      ProposalStatus = "pending_review"
	ProposalApproved           ProposalStatus = "approved"
	ProposalRepoChangeQueued   ProposalStatus = "repo_change_queued"
	ProposalRepoChangeRunning  ProposalStatus = "repo_change_running"
	ProposalValidationPending  ProposalStatus = "validation_pending"
	ProposalPROpen             ProposalStatus = "pr_open"
	ProposalDismissed          ProposalStatus = "dismissed"
	ProposalRejected           ProposalStatus = "rejected"
	ProposalSuperseded         ProposalStatus = "superseded"
	ProposalMerged             ProposalStatus = "merged"
	ProposalFailedValidation   ProposalStatus = "failed_validation"
	ProposalCanceled           ProposalStatus = "canceled"
)

type AttemptState string

const (
	AttemptStatePatchPlan           AttemptState = "patch_plan"
	AttemptStateInvestigateComplete AttemptState = "investigate_complete"
	AttemptStatePatchGenerated      AttemptState = "patch_generated"
	AttemptStateValidationRunning   AttemptState = "validation_running"
	AttemptStateCIObserving         AttemptState = "ci_observing"
	AttemptStateRetryDeciding       AttemptState = "retry_deciding"
	AttemptStateOverlayPlan         AttemptState = "overlay_plan"
	AttemptStateOverlayGenerated    AttemptState = "overlay_generated"
	AttemptStateOverlayValidating   AttemptState = "overlay_validating"
	AttemptStateOverlayActive       AttemptState = "overlay_active"
	AttemptStateSandboxFailed       AttemptState = "sandbox_failed"
	AttemptStatePROpen              AttemptState = "pr_open"
	AttemptStateCIFailed            AttemptState = "ci_failed"
	AttemptStateClosedUnmerged      AttemptState = "closed_unmerged"
	AttemptStateMerged              AttemptState = "merged"
	AttemptStateNeedsReview         AttemptState = "needs_review"
	AttemptStateAbandoned           AttemptState = "abandoned"
	AttemptStateSuperseded          AttemptState = "superseded"
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
	ProposalStatus       ProposalStatus
	AttemptState         AttemptState
	CurrentOperationKind string
}

type AttemptPhaseDecision struct {
	TransitionDecision
	NextPhase           AttemptPhaseState
	ExpectedProposal    ProposalStatus
	AllowedProposalNext []ProposalStatus
	AllowedAttemptNext  []AttemptState
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
	case AttemptStatePatchPlan, AttemptStateOverlayPlan:
		switch snapshot.ProposalStatus {
		case ProposalRepoChangeQueued:
			return AttemptPhaseWorkspaceOpening
		case ProposalRepoChangeRunning:
			return AttemptPhaseImplementing
		default:
			return AttemptPhasePlanning
		}
	case AttemptStatePatchGenerated, AttemptStateOverlayGenerated:
		return AttemptPhaseValidating
	case AttemptStateValidationRunning, AttemptStateOverlayValidating:
		if snapshot.ProposalStatus == ProposalValidationPending {
			return AttemptPhasePROpen
		}
		return AttemptPhaseValidating
	case AttemptStatePROpen, AttemptStateCIObserving:
		return AttemptPhasePROpen
	case AttemptStateRetryDeciding:
		return AttemptPhaseRetryDeciding
	case AttemptStateSandboxFailed,
		AttemptStateCIFailed,
		AttemptStateClosedUnmerged,
		AttemptStateMerged,
		AttemptStateNeedsReview,
		AttemptStateAbandoned,
		AttemptStateSuperseded,
		AttemptStateOverlayActive:
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
		AllowedProposalNext: []ProposalStatus{snapshot.ProposalStatus},
		AllowedAttemptNext:  []AttemptState{AttemptStatePatchPlan, AttemptStateOverlayPlan},
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
		AllowedProposalNext: []ProposalStatus{snapshot.ProposalStatus},
		AllowedAttemptNext:  []AttemptState{AttemptStatePatchPlan, AttemptStateOverlayPlan},
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
		AllowedProposalNext: []ProposalStatus{snapshot.ProposalStatus},
		AllowedAttemptNext:  []AttemptState{AttemptStatePatchPlan, AttemptStateOverlayPlan},
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
		AllowedProposalNext: []ProposalStatus{ProposalRepoChangeQueued},
		AllowedAttemptNext: []AttemptState{
			AttemptStatePatchPlan,
			AttemptStateOverlayPlan,
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
		AllowedProposalNext: []ProposalStatus{ProposalApproved, ProposalRepoChangeQueued},
		AllowedAttemptNext: []AttemptState{
			AttemptStatePatchPlan,
			AttemptStateOverlayPlan,
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
		AllowedProposalNext: []ProposalStatus{ProposalRepoChangeRunning},
		AllowedAttemptNext: []AttemptState{
			AttemptStatePatchPlan,
			AttemptStateOverlayPlan,
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
		AllowedProposalNext: []ProposalStatus{ProposalRepoChangeRunning, ProposalApproved},
		AllowedAttemptNext: []AttemptState{
			AttemptStatePatchPlan,
			AttemptStateOverlayPlan,
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
		AllowedProposalNext: []ProposalStatus{ProposalRepoChangeRunning, ProposalApproved},
		AllowedAttemptNext: []AttemptState{
			AttemptStatePatchGenerated,
			AttemptStateOverlayGenerated,
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
		AllowedProposalNext: []ProposalStatus{ProposalApproved, ProposalRepoChangeRunning, ProposalFailedValidation},
		AllowedAttemptNext:  []AttemptState{AttemptStateSandboxFailed, AttemptStateCIFailed, AttemptStateClosedUnmerged, AttemptStateNeedsReview},
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
		AllowedProposalNext: []ProposalStatus{ProposalPendingReview},
		AllowedAttemptNext:  []AttemptState{AttemptStateNeedsReview},
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
		AllowedProposalNext: []ProposalStatus{ProposalValidationPending},
		AllowedAttemptNext: []AttemptState{
			AttemptStateValidationRunning,
			AttemptStateOverlayValidating,
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
		AllowedProposalNext: []ProposalStatus{ProposalApproved, ProposalFailedValidation},
		AllowedAttemptNext:  []AttemptState{AttemptStateSandboxFailed, AttemptStateCIFailed, AttemptStateClosedUnmerged, AttemptStateNeedsReview},
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
		AllowedProposalNext: []ProposalStatus{ProposalPendingReview},
		AllowedAttemptNext:  []AttemptState{AttemptStateNeedsReview},
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
