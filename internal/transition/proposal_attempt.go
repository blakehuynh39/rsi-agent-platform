package transition

import (
	"fmt"
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
	"github.com/piplabs/rsi-agent-platform/internal/review"
)

type ProposalStatus = review.ProposalStatus

const (
	ProposalQueuedForPromotion = review.ProposalQueuedForPromotion
	ProposalPendingReview      = review.ProposalPendingReview
	ProposalApproved           = review.ProposalApproved
	ProposalRepoChangeQueued   = review.ProposalRepoChangeQueued
	ProposalRepoChangeRunning  = review.ProposalRepoChangeRunning
	ProposalValidationPending  = review.ProposalValidationPending
	ProposalPROpen             = review.ProposalPROpen
	ProposalDismissed          = review.ProposalDismissed
	ProposalRejected           = review.ProposalRejected
	ProposalSuperseded         = review.ProposalSuperseded
	ProposalMerged             = review.ProposalMerged
	ProposalFailedValidation   = review.ProposalFailedValidation
	ProposalCanceled           = review.ProposalCanceled
)

type AttemptState = improvement.ChangeAttemptState

const (
	AttemptStatePatchPlan           = improvement.AttemptStatePatchPlan
	AttemptStateInvestigateComplete = improvement.AttemptStateInvestigateComplete
	AttemptStatePatchGenerated      = improvement.AttemptStatePatchGenerated
	AttemptStateValidationRunning   = improvement.AttemptStateValidationRunning
	AttemptStateCIObserving         = improvement.AttemptStateCIObserving
	AttemptStateRetryDeciding       = improvement.AttemptStateRetryDeciding
	AttemptStateOverlayPlan         = improvement.AttemptStateOverlayPlan
	AttemptStateOverlayGenerated    = improvement.AttemptStateOverlayGenerated
	AttemptStateOverlayValidating   = improvement.AttemptStateOverlayValidating
	AttemptStateOverlayActive       = improvement.AttemptStateOverlayActive
	AttemptStateSandboxFailed       = improvement.AttemptStateSandboxFailed
	AttemptStatePROpen              = improvement.AttemptStatePROpen
	AttemptStateCIFailed            = improvement.AttemptStateCIFailed
	AttemptStateClosedUnmerged      = improvement.AttemptStateClosedUnmerged
	AttemptStateMerged              = improvement.AttemptStateMerged
	AttemptStateNeedsReview         = improvement.AttemptStateNeedsReview
	AttemptStateAbandoned           = improvement.AttemptStateAbandoned
	AttemptStateSuperseded          = improvement.AttemptStateSuperseded
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
	CommandLineActivated                    AttemptPhaseCommandKind = "line_activated"
	CommandAttemptPlannedWorkspace          AttemptPhaseCommandKind = "attempt_planned_workspace"
	CommandAttemptPlannedImplement          AttemptPhaseCommandKind = "attempt_planned_implement"
	CommandWorkspaceOpenDeferred            AttemptPhaseCommandKind = "workspace_open_deferred"
	CommandWorkspaceReady                   AttemptPhaseCommandKind = "workspace_ready"
	CommandWorkspaceFailedRetryable         AttemptPhaseCommandKind = "workspace_failed_retryable"
	CommandWorkspaceFailedReview            AttemptPhaseCommandKind = "workspace_failed_review"
	CommandWorkspaceMetadataSynced          AttemptPhaseCommandKind = "workspace_metadata_synced"
	CommandWorkspaceToolValidationStarted   AttemptPhaseCommandKind = "workspace_tool_validation_started"
	CommandWorkspaceToolValidationCompleted AttemptPhaseCommandKind = "workspace_tool_validation_completed"
	CommandWorkspaceToolValidationFailed    AttemptPhaseCommandKind = "workspace_tool_validation_failed"
	CommandAttemptRunnerStarted             AttemptPhaseCommandKind = "attempt_runner_started"
	CommandAttemptRunnerCompleted           AttemptPhaseCommandKind = "attempt_runner_completed"
	CommandOverlayActivated                 AttemptPhaseCommandKind = "overlay_activated"
	CommandImplementationDeferred           AttemptPhaseCommandKind = "implementation_deferred"
	CommandImplementationCompleted          AttemptPhaseCommandKind = "implementation_completed"
	CommandImplementationFailedRetryable    AttemptPhaseCommandKind = "implementation_failed_retryable"
	CommandImplementationFailedReview       AttemptPhaseCommandKind = "implementation_failed_review"
	CommandValidationStarted                AttemptPhaseCommandKind = "validation_started"
	CommandValidationCompleted              AttemptPhaseCommandKind = "validation_completed"
	CommandValidationFailedRetryable        AttemptPhaseCommandKind = "validation_failed_retryable"
	CommandValidationFailedReview           AttemptPhaseCommandKind = "validation_failed_review"
	CommandAttemptPROpened                  AttemptPhaseCommandKind = "attempt_pr_opened"
	CommandPROpenFailedRetryable            AttemptPhaseCommandKind = "pr_open_failed_retryable"
	CommandPROpenFailedReview               AttemptPhaseCommandKind = "pr_open_failed_review"
	CommandAttemptMerged                    AttemptPhaseCommandKind = "attempt_merged"
	CommandAttemptClosedUnmerged            AttemptPhaseCommandKind = "attempt_closed_unmerged"
	CommandAttemptCIFailed                  AttemptPhaseCommandKind = "attempt_ci_failed"
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
		CommandWorkspaceReady:           reduceWorkspaceReady,
		CommandWorkspaceFailedRetryable: reduceWorkspaceFailedRetryable,
		CommandWorkspaceFailedReview:    reduceWorkspaceFailedReview,
	},
	AttemptPhaseImplementing: {
		CommandWorkspaceMetadataSynced:          reduceWorkspaceMetadataSynced,
		CommandWorkspaceToolValidationStarted:   reduceWorkspaceToolValidationStarted,
		CommandWorkspaceToolValidationCompleted: reduceWorkspaceToolValidationCompleted,
		CommandWorkspaceToolValidationFailed:    reduceWorkspaceToolValidationFailed,
		CommandAttemptRunnerStarted:             reduceAttemptRunnerStarted,
		CommandAttemptRunnerCompleted:           reduceAttemptRunnerCompleted,
		CommandOverlayActivated:                 reduceOverlayActivated,
		CommandImplementationDeferred:           reduceImplementationDeferred,
		CommandImplementationCompleted:          reduceImplementationCompleted,
		CommandImplementationFailedRetryable:    reduceImplementationFailedRetryable,
		CommandImplementationFailedReview:       reduceImplementationFailedReview,
	},
	AttemptPhaseValidating: {
		CommandWorkspaceMetadataSynced:          reduceWorkspaceMetadataSynced,
		CommandWorkspaceToolValidationStarted:   reduceWorkspaceToolValidationStarted,
		CommandWorkspaceToolValidationCompleted: reduceWorkspaceToolValidationCompleted,
		CommandWorkspaceToolValidationFailed:    reduceWorkspaceToolValidationFailed,
		CommandValidationStarted:                reduceValidationStarted,
		CommandValidationCompleted:              reduceValidationCompleted,
		CommandValidationFailedRetryable:        reduceValidationFailedRetryable,
		CommandValidationFailedReview:           reduceValidationFailedReview,
		CommandAttemptPROpened:                  reduceAttemptPROpened,
		CommandAttemptMerged:                    reduceAttemptMerged,
		CommandAttemptClosedUnmerged:            reduceAttemptClosedUnmerged,
		CommandAttemptCIFailed:                  reduceAttemptCIFailed,
	},
	AttemptPhasePROpen: {
		CommandAttemptPROpened:       reduceAttemptPROpened,
		CommandPROpenFailedRetryable: reducePROpenFailedRetryable,
		CommandPROpenFailedReview:    reducePROpenFailedReview,
		CommandAttemptMerged:         reduceAttemptMerged,
		CommandAttemptClosedUnmerged: reduceAttemptClosedUnmerged,
		CommandAttemptCIFailed:       reduceAttemptCIFailed,
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
				Kind:           EffectOpenWorkspace,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "workspace_open"),
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

func reduceWorkspaceFailedRetryable(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "workspace opening failed with a retryable outcome",
			Events: []DomainEventDescriptor{{
				Kind: "workspace_failed_retryable",
			}},
		},
		NextPhase:           AttemptPhaseRetryDeciding,
		AllowedProposalNext: []ProposalStatus{ProposalApproved, ProposalRepoChangeQueued, ProposalRepoChangeRunning, ProposalFailedValidation},
		AllowedAttemptNext:  []AttemptState{AttemptStateSandboxFailed, AttemptStateNeedsReview},
	}
}

func reduceWorkspaceFailedReview(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "workspace opening failed and line review is required",
			Events: []DomainEventDescriptor{{
				Kind: "workspace_failed_needs_review",
			}},
		},
		NextPhase:           AttemptPhaseTerminal,
		AllowedProposalNext: []ProposalStatus{ProposalPendingReview},
		AllowedAttemptNext:  []AttemptState{AttemptStateNeedsReview},
	}
}

func reduceAttemptRunnerStarted(snapshot AttemptSnapshot, _ CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "implementation runner started",
			Events: []DomainEventDescriptor{{
				Kind: "attempt_runner_started",
			}},
		},
		NextPhase:           AttemptPhaseImplementing,
		AllowedProposalNext: []ProposalStatus{snapshot.ProposalStatus},
		AllowedAttemptNext:  []AttemptState{snapshot.AttemptState},
	}
}

func reduceWorkspaceMetadataSynced(snapshot AttemptSnapshot, _ CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "workspace metadata synced from governed tool execution",
			Events: []DomainEventDescriptor{{
				Kind: "workspace_metadata_synced",
			}},
		},
		NextPhase:           deriveAttemptPhase(snapshot),
		AllowedProposalNext: []ProposalStatus{snapshot.ProposalStatus},
		AllowedAttemptNext:  []AttemptState{snapshot.AttemptState},
	}
}

func reduceWorkspaceToolValidationStarted(snapshot AttemptSnapshot, _ CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "workspace tool validation started",
			Events: []DomainEventDescriptor{{
				Kind: "workspace_tool_validation_started",
			}},
		},
		NextPhase:           deriveAttemptPhase(snapshot),
		AllowedProposalNext: []ProposalStatus{snapshot.ProposalStatus},
		AllowedAttemptNext:  []AttemptState{snapshot.AttemptState},
	}
}

func reduceWorkspaceToolValidationCompleted(snapshot AttemptSnapshot, _ CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "workspace tool validation completed",
			Events: []DomainEventDescriptor{{
				Kind: "workspace_tool_validation_completed",
			}},
		},
		NextPhase:           deriveAttemptPhase(snapshot),
		AllowedProposalNext: []ProposalStatus{snapshot.ProposalStatus},
		AllowedAttemptNext:  []AttemptState{snapshot.AttemptState},
	}
}

func reduceWorkspaceToolValidationFailed(snapshot AttemptSnapshot, _ CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "workspace tool validation failed",
			Events: []DomainEventDescriptor{{
				Kind: "workspace_tool_validation_failed",
			}},
		},
		NextPhase:           deriveAttemptPhase(snapshot),
		AllowedProposalNext: []ProposalStatus{snapshot.ProposalStatus},
		AllowedAttemptNext:  []AttemptState{snapshot.AttemptState},
	}
}

func reduceAttemptRunnerCompleted(snapshot AttemptSnapshot, _ CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "implementation runner completed",
			Events: []DomainEventDescriptor{{
				Kind: "attempt_runner_completed",
			}},
		},
		NextPhase:           AttemptPhaseImplementing,
		AllowedProposalNext: []ProposalStatus{snapshot.ProposalStatus},
		AllowedAttemptNext:  []AttemptState{snapshot.AttemptState},
	}
}

func reduceOverlayActivated(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "harness overlay was activated for the approved attempt",
			Events: []DomainEventDescriptor{{
				Kind: "overlay_activated",
			}},
		},
		NextPhase:           AttemptPhaseTerminal,
		AllowedProposalNext: []ProposalStatus{ProposalApproved, ProposalMerged},
		AllowedAttemptNext:  []AttemptState{AttemptStateOverlayActive},
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
				Kind:           EffectInvokeRunner,
				Status:         EffectQueued,
				IdempotencyKey: effectKey(command, "implement_attempt"),
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
		},
		NextPhase:           AttemptPhaseTerminal,
		AllowedProposalNext: []ProposalStatus{ProposalPendingReview},
		AllowedAttemptNext:  []AttemptState{AttemptStateNeedsReview},
	}
}

func reduceValidationStarted(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	effects := []EffectRequest{}
	if hasSandboxObservationMetadata(command) {
		effects = append(effects, EffectRequest{
			Kind:           EffectObserveWorkspaceValidation,
			Status:         EffectQueued,
			IdempotencyKey: effectKey(command, "observe_workspace_validation"),
		})
	}
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "validation started inside the governed sandbox",
			Events: []DomainEventDescriptor{{
				Kind: "validation_started",
			}},
			Effects: effects,
		},
		NextPhase:           AttemptPhaseValidating,
		AllowedProposalNext: []ProposalStatus{ProposalRepoChangeQueued, ProposalRepoChangeRunning},
		AllowedAttemptNext: []AttemptState{
			AttemptStatePatchGenerated,
			AttemptStateOverlayGenerated,
		},
	}
}

func hasSandboxObservationMetadata(command CommandEnvelope) bool {
	payload := command.Payload
	if payload == nil {
		return false
	}
	namespace, _ := payload["sandbox_namespace"].(string)
	jobName, _ := payload["sandbox_job_name"].(string)
	namespace = strings.TrimSpace(namespace)
	jobName = strings.TrimSpace(jobName)
	return namespace != "" && jobName != ""
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
		AllowedProposalNext: []ProposalStatus{ProposalRepoChangeRunning, ProposalValidationPending},
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
		},
		NextPhase:           AttemptPhaseTerminal,
		AllowedProposalNext: []ProposalStatus{ProposalPendingReview},
		AllowedAttemptNext:  []AttemptState{AttemptStateNeedsReview},
	}
}

func reduceAttemptPROpened(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "governed pr open was observed for the attempt",
			Events: []DomainEventDescriptor{{
				Kind: "attempt_pr_opened",
			}},
		},
		NextPhase: AttemptPhasePROpen,
		AllowedProposalNext: []ProposalStatus{
			ProposalApproved,
			ProposalValidationPending,
			ProposalPROpen,
		},
		AllowedAttemptNext: []AttemptState{
			AttemptStateCIObserving,
			AttemptStatePROpen,
		},
	}
}

func reducePROpenFailedReview(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "governed pr open failed and line review is required",
			Events: []DomainEventDescriptor{{
				Kind: "pr_open_failed_needs_review",
			}},
		},
		NextPhase:           AttemptPhaseTerminal,
		AllowedProposalNext: []ProposalStatus{ProposalPendingReview},
		AllowedAttemptNext:  []AttemptState{AttemptStateNeedsReview},
	}
}

func reducePROpenFailedRetryable(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "governed pr open failed with a retryable outcome",
			Events: []DomainEventDescriptor{{
				Kind: "pr_open_failed_retryable",
			}},
		},
		NextPhase:           AttemptPhaseRetryDeciding,
		AllowedProposalNext: []ProposalStatus{ProposalApproved, ProposalValidationPending, ProposalPROpen, ProposalFailedValidation},
		AllowedAttemptNext:  []AttemptState{AttemptStateNeedsReview},
	}
}

func reduceAttemptMerged(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "merged pr was observed for the attempt",
			Events: []DomainEventDescriptor{{
				Kind: "attempt_merged",
			}},
		},
		NextPhase: AttemptPhaseTerminal,
		AllowedProposalNext: []ProposalStatus{
			ProposalApproved,
			ProposalPROpen,
			ProposalMerged,
		},
		AllowedAttemptNext: []AttemptState{AttemptStateMerged},
	}
}

func reduceAttemptClosedUnmerged(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "closed unmerged pr was observed for the attempt",
			Events: []DomainEventDescriptor{{
				Kind: "attempt_closed_unmerged",
			}},
		},
		NextPhase: AttemptPhaseTerminal,
		AllowedProposalNext: []ProposalStatus{
			ProposalApproved,
			ProposalPROpen,
			ProposalPendingReview,
		},
		AllowedAttemptNext: []AttemptState{AttemptStateClosedUnmerged},
	}
}

func reduceAttemptCIFailed(snapshot AttemptSnapshot, command CommandEnvelope) AttemptPhaseDecision {
	return AttemptPhaseDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "ci failure was observed for the attempt",
			Events: []DomainEventDescriptor{{
				Kind: "attempt_ci_failed",
			}},
		},
		NextPhase: AttemptPhaseTerminal,
		AllowedProposalNext: []ProposalStatus{
			ProposalApproved,
			ProposalPROpen,
			ProposalPendingReview,
		},
		AllowedAttemptNext: []AttemptState{AttemptStateCIFailed},
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
