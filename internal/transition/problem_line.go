package transition

type ProblemLineState string

const (
	ProblemLineStateObserving        ProblemLineState = "observing"
	ProblemLineStateReadyForProposal ProblemLineState = "ready_for_proposal"
	ProblemLineStateClosed           ProblemLineState = "closed"
)

type ProblemLineCommandKind string

const (
	CommandProblemLineEvaluateTrace         ProblemLineCommandKind = "problem_line_evaluate_trace"
	CommandProblemLinePromote               ProblemLineCommandKind = "problem_line_promote"
	CommandProblemLineRecordOutcome         ProblemLineCommandKind = "problem_line_record_outcome"
	CommandProblemLineProjectTrace          ProblemLineCommandKind = "problem_line_project_trace"
	CommandProblemLineRecordFeedback        ProblemLineCommandKind = "problem_line_record_feedback"
	CommandProblemLineRecordRating          ProblemLineCommandKind = "problem_line_record_rating"
	CommandProblemLineRecordImprovementNote ProblemLineCommandKind = "problem_line_record_improvement_note"
	CommandProblemLineScheduleReplay        ProblemLineCommandKind = "problem_line_schedule_replay"
)

type ProblemLineSnapshot struct {
	State                   ProblemLineState `json:"state,omitempty"`
	TraceExists             bool             `json:"trace_exists,omitempty"`
	SlotsAvailable          bool             `json:"slots_available,omitempty"`
	HasPromotableCandidates bool             `json:"has_promotable_candidates,omitempty"`
	PromotionLeaseBlocked   bool             `json:"promotion_lease_blocked,omitempty"`
}

type ProblemLineDecision struct {
	TransitionDecision
	NextState ProblemLineState `json:"next_state,omitempty"`
}

func ReduceProblemLine(snapshot ProblemLineSnapshot, command CommandEnvelope) ProblemLineDecision {
	current := snapshot.State
	if current == "" {
		current = ProblemLineStateObserving
	}
	switch ProblemLineCommandKind(command.CommandKind) {
	case CommandProblemLineEvaluateTrace:
		if !snapshot.TraceExists {
			return rejectProblemLine(current, "trace not found for evaluation")
		}
		return ProblemLineDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "trace evaluation applied to problem-line evidence",
				Events: []DomainEventDescriptor{{
					Kind: "problem_line_trace_evaluated",
				}},
				Effects: []EffectRequest{{
					Kind:           EffectInvokeRunner,
					Status:         EffectQueued,
					IdempotencyKey: command.CommandID,
				}},
			},
			NextState: current,
		}
	case CommandProblemLinePromote:
		if snapshot.PromotionLeaseBlocked {
			return rejectProblemLine(current, "proposal promoter lease already held")
		}
		if !snapshot.SlotsAvailable {
			return noopProblemLine(current, "proposal promotion blocked by active slot cap")
		}
		if !snapshot.HasPromotableCandidates {
			return noopProblemLine(current, "no promotable problem lines are queued")
		}
		return ProblemLineDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "queued problem lines promoted into governed proposals",
				Events: []DomainEventDescriptor{{
					Kind: "problem_line_promoted",
				}},
			},
			NextState: ProblemLineStateReadyForProposal,
		}
	case CommandProblemLineRecordOutcome:
		return ProblemLineDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "problem-line outcome recorded",
				Events: []DomainEventDescriptor{{
					Kind: "problem_line_outcome_recorded",
				}},
			},
			NextState: current,
		}
	case CommandProblemLineProjectTrace:
		if !snapshot.TraceExists {
			return rejectProblemLine(current, "trace not found for projection")
		}
		return ProblemLineDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "problem-line trace projection recorded",
				Events: []DomainEventDescriptor{{
					Kind: "problem_line_trace_projected",
				}},
			},
			NextState: current,
		}
	case CommandProblemLineRecordFeedback:
		if !snapshot.TraceExists {
			return rejectProblemLine(current, "trace not found for feedback")
		}
		return ProblemLineDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "feedback recorded against the trace-owned problem line",
				Events: []DomainEventDescriptor{{
					Kind: "problem_line_feedback_recorded",
				}},
			},
			NextState: current,
		}
	case CommandProblemLineRecordRating:
		if !snapshot.TraceExists {
			return rejectProblemLine(current, "trace not found for rating")
		}
		return ProblemLineDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "rating recorded against the trace-owned problem line",
				Events: []DomainEventDescriptor{{
					Kind: "problem_line_rating_recorded",
				}},
			},
			NextState: current,
		}
	case CommandProblemLineRecordImprovementNote:
		if !snapshot.TraceExists {
			return rejectProblemLine(current, "trace not found for improvement note")
		}
		return ProblemLineDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "improvement note recorded against the trace-owned problem line",
				Events: []DomainEventDescriptor{{
					Kind: "problem_line_improvement_note_recorded",
				}},
			},
			NextState: current,
		}
	case CommandProblemLineScheduleReplay:
		if !snapshot.TraceExists {
			return rejectProblemLine(current, "trace not found for replay scheduling")
		}
		return ProblemLineDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "trace replay scheduled from the problem-line command path",
				Events: []DomainEventDescriptor{{
					Kind: "problem_line_replay_scheduled",
				}},
				Commands: []CommandDescriptor{{
					MachineKind: MachineProblemLine,
					AggregateID: command.AggregateID,
					CommandKind: string(CommandProblemLineEvaluateTrace),
					CommandID:   command.CommandID + ":evaluate",
					Actor:       command.Actor,
					Payload: map[string]any{
						"trigger":      "replay",
						"requested_by": command.Actor,
					},
				}},
			},
			NextState: current,
		}
	default:
		return rejectProblemLine(current, "unsupported problem-line command for current state")
	}
}

func noopProblemLine(state ProblemLineState, reason string) ProblemLineDecision {
	return ProblemLineDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionNoop,
			Reason:       reason,
		},
		NextState: state,
	}
}

func rejectProblemLine(state ProblemLineState, reason string) ProblemLineDecision {
	return ProblemLineDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionReject,
			Reason:       reason,
		},
		NextState: state,
	}
}
