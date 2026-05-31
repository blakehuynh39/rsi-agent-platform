package transition

import "github.com/piplabs/rsi-agent-platform/internal/review"

type ProposalLineCommandKind string

const (
	CommandProposalApproveIntervention   ProposalLineCommandKind = "proposal_approve_intervention"
	CommandProposalRejectLine            ProposalLineCommandKind = "proposal_reject_line"
	CommandProposalDismissLine           ProposalLineCommandKind = "proposal_dismiss_line"
	CommandProposalStopLine              ProposalLineCommandKind = "proposal_stop_line"
	CommandProposalRetryAttempt          ProposalLineCommandKind = "proposal_retry_attempt"
	CommandProposalResumeExecution       ProposalLineCommandKind = "proposal_resume_execution"
	CommandProposalMarkRepoChangeQueued  ProposalLineCommandKind = "proposal_mark_repo_change_queued"
	CommandProposalMarkRepoChangeRunning ProposalLineCommandKind = "proposal_mark_repo_change_running"
	CommandProposalMarkValidationPending ProposalLineCommandKind = "proposal_mark_validation_pending"
	CommandProposalMarkFailedValidation  ProposalLineCommandKind = "proposal_mark_failed_validation"
	CommandProposalMarkPROpen            ProposalLineCommandKind = "proposal_mark_pr_open"
	CommandProposalMarkMerged            ProposalLineCommandKind = "proposal_mark_merged"
	CommandProposalRetryableFailure      ProposalLineCommandKind = "proposal_retryable_failure"
	CommandProposalNeedsReview           ProposalLineCommandKind = "proposal_needs_review"
)

type ProposalLineSnapshot struct {
	State            review.ProposalStatus           `json:"state"`
	InterventionKind review.ProposalInterventionKind `json:"intervention_kind,omitempty"`
}

type ProposalLineDecision struct {
	TransitionDecision
	NextState review.ProposalStatus `json:"next_state,omitempty"`
}

func ReduceProposalLine(snapshot ProposalLineSnapshot, command CommandEnvelope) ProposalLineDecision {
	commandKind := ProposalLineCommandKind(command.CommandKind)
	switch snapshot.State {
	case review.ProposalPendingReview:
		switch commandKind {
		case CommandProposalApproveIntervention:
			return ProposalLineDecision{
				TransitionDecision: TransitionDecision{
					DecisionKind: DecisionAdvance,
					Reason:       "proposal line approved for execution",
					Events: []DomainEventDescriptor{{
						Kind: "proposal_line_approved",
					}},
					Commands: proposalResumeCommands(command.AggregateID, snapshot.InterventionKind),
				},
				NextState: review.ProposalApproved,
			}
		case CommandProposalRejectLine:
			return proposalLineAdvance(review.ProposalRejected, "proposal line rejected", "proposal_line_rejected")
		case CommandProposalDismissLine:
			return proposalLineAdvance(review.ProposalDismissed, "proposal line dismissed", "proposal_line_dismissed")
		case CommandProposalStopLine:
			return proposalLineAdvance(review.ProposalCanceled, "proposal line stopped", "proposal_line_stopped")
		}
	case review.ProposalApproved:
		switch commandKind {
		case CommandProposalApproveIntervention:
			return proposalLineNoop(snapshot.State, "proposal line already approved")
		case CommandProposalMarkRepoChangeQueued:
			return proposalLineAdvance(review.ProposalRepoChangeQueued, "proposal line moved to repo_change_queued from execution progress", "proposal_line_repo_change_queued")
		case CommandProposalMarkRepoChangeRunning:
			return proposalLineAdvance(review.ProposalRepoChangeRunning, "proposal line moved to repo_change_running from execution progress", "proposal_line_repo_change_running")
		case CommandProposalMarkPROpen:
			return proposalLineAdvance(review.ProposalPROpen, "proposal line moved to pr_open from external outcome", "proposal_line_pr_opened")
		case CommandProposalMarkMerged:
			return proposalLineAdvance(review.ProposalMerged, "proposal line merged from external outcome", "proposal_line_merged")
		case CommandProposalRetryableFailure:
			return ProposalLineDecision{
				TransitionDecision: TransitionDecision{
					DecisionKind: DecisionAdvance,
					Reason:       "proposal line recorded a retryable failure and remains approved",
					Events: []DomainEventDescriptor{{
						Kind: "proposal_line_retryable_failure",
					}},
					Commands: proposalResumeCommands(command.AggregateID, snapshot.InterventionKind),
				},
				NextState: review.ProposalApproved,
			}
		case CommandProposalNeedsReview:
			return proposalLineAdvance(review.ProposalPendingReview, "proposal line returned to pending_review from external outcome", "proposal_line_needs_review")
		case CommandProposalStopLine:
			return proposalLineAdvance(review.ProposalCanceled, "proposal line stopped", "proposal_line_stopped")
		case CommandProposalRetryAttempt:
			return ProposalLineDecision{
				TransitionDecision: TransitionDecision{
					DecisionKind: DecisionAdvance,
					Reason:       "retry requested for approved proposal line",
					Events: []DomainEventDescriptor{{
						Kind: "proposal_line_retry_requested",
					}},
					Commands: proposalResumeCommands(command.AggregateID, snapshot.InterventionKind),
				},
				NextState: snapshot.State,
			}
		case CommandProposalResumeExecution:
			return proposalLineAdvance(snapshot.State, "proposal line execution resumed through internal command progression", "proposal_line_execution_resumed")
		}
	case review.ProposalRepoChangeQueued, review.ProposalRepoChangeRunning, review.ProposalValidationPending, review.ProposalFailedValidation, review.ProposalPROpen:
		switch commandKind {
		case CommandProposalMarkRepoChangeQueued:
			switch snapshot.State {
			case review.ProposalRepoChangeQueued:
				return proposalLineNoop(snapshot.State, "proposal line already repo_change_queued")
			case review.ProposalFailedValidation:
				return proposalLineAdvance(review.ProposalRepoChangeQueued, "proposal line returned to repo_change_queued from failed_validation", "proposal_line_repo_change_queued")
			}
		case CommandProposalMarkRepoChangeRunning:
			switch snapshot.State {
			case review.ProposalRepoChangeQueued:
				return proposalLineAdvance(review.ProposalRepoChangeRunning, "proposal line moved to repo_change_running from execution progress", "proposal_line_repo_change_running")
			case review.ProposalRepoChangeRunning:
				return proposalLineNoop(snapshot.State, "proposal line already repo_change_running")
			}
		case CommandProposalMarkValidationPending:
			switch snapshot.State {
			case review.ProposalRepoChangeRunning:
				return proposalLineAdvance(review.ProposalValidationPending, "proposal line moved to validation_pending from execution progress", "proposal_line_validation_pending")
			case review.ProposalValidationPending:
				return proposalLineNoop(snapshot.State, "proposal line already validation_pending")
			}
		case CommandProposalMarkFailedValidation:
			switch snapshot.State {
			case review.ProposalRepoChangeQueued, review.ProposalRepoChangeRunning, review.ProposalValidationPending:
				return proposalLineAdvance(review.ProposalFailedValidation, "proposal line moved to failed_validation from execution progress", "proposal_line_failed_validation")
			case review.ProposalFailedValidation:
				return proposalLineNoop(snapshot.State, "proposal line already failed_validation")
			}
		case CommandProposalMarkPROpen:
			if snapshot.State == review.ProposalPROpen {
				return proposalLineNoop(snapshot.State, "proposal line already pr_open")
			}
			return proposalLineAdvance(review.ProposalPROpen, "proposal line moved to pr_open from external outcome", "proposal_line_pr_opened")
		case CommandProposalMarkMerged:
			if snapshot.State == review.ProposalMerged {
				return proposalLineNoop(snapshot.State, "proposal line already merged")
			}
			return proposalLineAdvance(review.ProposalMerged, "proposal line merged from external outcome", "proposal_line_merged")
		case CommandProposalRetryableFailure:
			return ProposalLineDecision{
				TransitionDecision: TransitionDecision{
					DecisionKind: DecisionAdvance,
					Reason:       "proposal line recorded a retryable failure and remains approved",
					Events: []DomainEventDescriptor{{
						Kind: "proposal_line_retryable_failure",
					}},
					Commands: proposalResumeCommands(command.AggregateID, snapshot.InterventionKind),
				},
				NextState: review.ProposalApproved,
			}
		case CommandProposalNeedsReview:
			return proposalLineAdvance(review.ProposalPendingReview, "proposal line returned to pending_review from external outcome", "proposal_line_needs_review")
		case CommandProposalRetryAttempt:
			if snapshot.State == review.ProposalFailedValidation {
				return ProposalLineDecision{
					TransitionDecision: TransitionDecision{
						DecisionKind: DecisionAdvance,
						Reason:       "retry requested for failed validation line",
						Events: []DomainEventDescriptor{{
							Kind: "proposal_line_retry_requested",
						}},
						Commands: proposalResumeCommands(command.AggregateID, snapshot.InterventionKind),
					},
					NextState: review.ProposalApproved,
				}
			}
			return proposalLineNoop(snapshot.State, "proposal line execution already in progress")
		case CommandProposalStopLine:
			return proposalLineAdvance(review.ProposalCanceled, "proposal line stopped", "proposal_line_stopped")
		case CommandProposalResumeExecution:
			return proposalLineAdvance(snapshot.State, "proposal line execution resumed through internal command progression", "proposal_line_execution_resumed")
		}
	case review.ProposalDismissed, review.ProposalRejected, review.ProposalMerged, review.ProposalCanceled, review.ProposalSuperseded:
		switch commandKind {
		case CommandProposalStopLine:
			return proposalLineNoop(snapshot.State, "proposal line already terminal")
		case CommandProposalMarkRepoChangeRunning, CommandProposalMarkValidationPending, CommandProposalMarkPROpen, CommandProposalMarkMerged, CommandProposalRetryableFailure, CommandProposalNeedsReview, CommandProposalResumeExecution:
			return proposalLineNoop(snapshot.State, "proposal line already terminal")
		}
	}
	return ProposalLineDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionReject,
			Reason:       "unsupported proposal line command for current state",
		},
		NextState: snapshot.State,
	}
}

func proposalLineAdvance(next review.ProposalStatus, reason string, eventKind string) ProposalLineDecision {
	return ProposalLineDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       reason,
			Events: []DomainEventDescriptor{{
				Kind: eventKind,
			}},
		},
		NextState: next,
	}
}

func proposalLineNoop(state review.ProposalStatus, reason string) ProposalLineDecision {
	return ProposalLineDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionNoop,
			Reason:       reason,
		},
		NextState: state,
	}
}

func proposalResumeCommands(aggregateID string, interventionKind review.ProposalInterventionKind) []CommandDescriptor {
	if aggregateID == "" {
		return nil
	}
	if !review.ProposalExecutableIntervention(interventionKind) {
		return nil
	}
	return []CommandDescriptor{{
		MachineKind: MachineProposalLine,
		AggregateID: aggregateID,
		CommandKind: string(CommandProposalResumeExecution),
	}}
}
