package transition

type WorkflowStateKind string

const (
	WorkflowStateQueued            WorkflowStateKind = "queued"
	WorkflowStateCollectingContext WorkflowStateKind = "collecting_context"
	WorkflowStateWaitingOnActions  WorkflowStateKind = "waiting_on_actions"
	WorkflowStateExecuting         WorkflowStateKind = "executing"
	WorkflowStateReplyPending      WorkflowStateKind = "reply_pending"
	WorkflowStateNeedsHuman        WorkflowStateKind = "needs_human"
	WorkflowStateCompleted         WorkflowStateKind = "completed"
	WorkflowStateFailed            WorkflowStateKind = "failed"
	WorkflowStateSuperseded        WorkflowStateKind = "superseded"
)

type WorkflowCommandKind string

const (
	CommandWorkflowStarted                          WorkflowCommandKind = "workflow_started"
	CommandContextActionsQueued                     WorkflowCommandKind = "context_actions_queued"
	CommandContextSkipped                           WorkflowCommandKind = "context_skipped"
	CommandContextCompleted                         WorkflowCommandKind = "context_completed"
	CommandWorkflowExecutionCompleted               WorkflowCommandKind = "workflow_execution_completed"
	CommandWorkflowExecutionCompletedPartial        WorkflowCommandKind = "workflow_execution_completed_partial"
	CommandWorkflowExecutionCompletedNoReply        WorkflowCommandKind = "workflow_execution_completed_no_reply"
	CommandWorkflowExecutionCompletedPartialNoReply WorkflowCommandKind = "workflow_execution_completed_partial_no_reply"
	CommandReplyPosted                              WorkflowCommandKind = "reply_posted"
	CommandReplyPostedPartial                       WorkflowCommandKind = "reply_posted_partial"
	CommandWorkflowExecutionNeedsHuman              WorkflowCommandKind = "workflow_execution_needs_human"
	CommandWorkflowExecutionFailed                  WorkflowCommandKind = "workflow_execution_failed"
	CommandWorkflowSuperseded                       WorkflowCommandKind = "workflow_superseded"
)

type WorkflowSnapshot struct {
	State                WorkflowStateKind `json:"state"`
	TraceID              string            `json:"trace_id,omitempty"`
	CurrentOperationKind string            `json:"current_operation_kind,omitempty"`
}

type WorkflowDecision struct {
	TransitionDecision
	NextState     WorkflowStateKind   `json:"next_state,omitempty"`
	ExpectedState WorkflowStateKind   `json:"expected_state,omitempty"`
	AllowedNext   []WorkflowStateKind `json:"allowed_next,omitempty"`
}

type workflowReducer func(snapshot WorkflowSnapshot, command CommandEnvelope) WorkflowDecision

var workflowReducers = map[WorkflowStateKind]map[WorkflowCommandKind]workflowReducer{
	WorkflowStateQueued: {
		CommandWorkflowStarted:             reduceWorkflowStarted,
		CommandWorkflowExecutionFailed:     reduceWorkflowFailed,
		CommandWorkflowExecutionNeedsHuman: reduceWorkflowNeedsHuman,
		CommandWorkflowSuperseded:          reduceWorkflowSuperseded,
	},
	WorkflowStateCollectingContext: {
		CommandContextActionsQueued:        reduceContextActionsQueued,
		CommandContextSkipped:              reduceContextSkipped,
		CommandWorkflowExecutionFailed:     reduceWorkflowFailed,
		CommandWorkflowExecutionNeedsHuman: reduceWorkflowNeedsHuman,
		CommandWorkflowSuperseded:          reduceWorkflowSuperseded,
	},
	WorkflowStateWaitingOnActions: {
		CommandContextCompleted:            reduceContextCompleted,
		CommandWorkflowExecutionFailed:     reduceWorkflowFailed,
		CommandWorkflowExecutionNeedsHuman: reduceWorkflowNeedsHuman,
		CommandWorkflowSuperseded:          reduceWorkflowSuperseded,
	},
	WorkflowStateExecuting: {
		CommandWorkflowExecutionCompleted:               reduceExecutionCompleted,
		CommandWorkflowExecutionCompletedPartial:        reduceExecutionCompletedPartial,
		CommandWorkflowExecutionCompletedNoReply:        reduceExecutionCompletedNoReply,
		CommandWorkflowExecutionCompletedPartialNoReply: reduceExecutionCompletedPartialNoReply,
		CommandWorkflowExecutionFailed:                  reduceWorkflowFailed,
		CommandWorkflowExecutionNeedsHuman:              reduceWorkflowNeedsHuman,
		CommandWorkflowSuperseded:                       reduceWorkflowSuperseded,
	},
	WorkflowStateReplyPending: {
		CommandReplyPosted:                 reduceReplyPosted,
		CommandReplyPostedPartial:          reduceReplyPostedPartial,
		CommandWorkflowExecutionFailed:     reduceWorkflowFailed,
		CommandWorkflowExecutionNeedsHuman: reduceWorkflowNeedsHuman,
		CommandWorkflowSuperseded:          reduceWorkflowSuperseded,
	},
	WorkflowStateNeedsHuman: {
		CommandWorkflowSuperseded: reduceWorkflowSuperseded,
	},
	WorkflowStateCompleted: {
		CommandWorkflowSuperseded: reduceWorkflowSuperseded,
	},
	WorkflowStateFailed: {
		CommandWorkflowSuperseded: reduceWorkflowSuperseded,
	},
	WorkflowStateSuperseded: {},
}

func ReduceWorkflow(snapshot WorkflowSnapshot, command CommandEnvelope) WorkflowDecision {
	commandKind := WorkflowCommandKind(command.CommandKind)
	reducers, ok := workflowReducers[snapshot.State]
	if !ok {
		return rejectWorkflow(snapshot.State, commandKind, "unsupported workflow state")
	}
	reducer, ok := reducers[commandKind]
	if !ok {
		return rejectWorkflow(snapshot.State, commandKind, "unsupported workflow command for current state")
	}
	return reducer(snapshot, command)
}

func reduceWorkflowStarted(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "workflow execution moved from queued to context collection",
			Events: []DomainEventDescriptor{{
				Kind: "workflow_started",
			}},
		},
		NextState:     WorkflowStateCollectingContext,
		ExpectedState: WorkflowStateQueued,
		AllowedNext:   []WorkflowStateKind{WorkflowStateCollectingContext},
	}
}

func reduceContextActionsQueued(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "context actions were queued and workflow is waiting on action execution",
			Events: []DomainEventDescriptor{{
				Kind: "workflow_context_actions_queued",
			}},
		},
		NextState:     WorkflowStateWaitingOnActions,
		ExpectedState: snapshot.State,
		AllowedNext:   []WorkflowStateKind{WorkflowStateWaitingOnActions},
	}
}

func reduceContextSkipped(snapshot WorkflowSnapshot, command CommandEnvelope) WorkflowDecision {
	effects := []EffectRequest{}
	reason := "no context actions were required and execution can start immediately"
	if !workflowUsesChildExecution(command.Payload) {
		effects = append(effects, EffectRequest{
			Kind:           EffectInvokeRunner,
			Status:         EffectQueued,
			IdempotencyKey: "invoke_runner",
		})
		reason = "no context actions were required and legacy runner execution can start immediately"
	}
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       reason,
			Events: []DomainEventDescriptor{{
				Kind: "workflow_context_skipped",
			}},
			Effects: effects,
		},
		NextState:     WorkflowStateExecuting,
		ExpectedState: snapshot.State,
		AllowedNext:   []WorkflowStateKind{WorkflowStateExecuting},
	}
}

func reduceContextCompleted(snapshot WorkflowSnapshot, command CommandEnvelope) WorkflowDecision {
	effects := []EffectRequest{}
	reason := "context collection is complete and execution can proceed"
	if !workflowUsesChildExecution(command.Payload) {
		effects = append(effects, EffectRequest{
			Kind:           EffectInvokeRunner,
			Status:         EffectQueued,
			IdempotencyKey: "invoke_runner",
		})
		reason = "context collection is complete and legacy runner execution can proceed"
	}
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       reason,
			Events: []DomainEventDescriptor{{
				Kind: "workflow_context_completed",
			}},
			Effects: effects,
		},
		NextState:     WorkflowStateExecuting,
		ExpectedState: snapshot.State,
		AllowedNext:   []WorkflowStateKind{WorkflowStateExecuting},
	}
}

func reduceExecutionCompleted(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return reduceExecutionCompletedWithReply(snapshot, false)
}

func reduceExecutionCompletedPartial(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return reduceExecutionCompletedWithReply(snapshot, true)
}

func reduceExecutionCompletedWithReply(snapshot WorkflowSnapshot, partial bool) WorkflowDecision {
	reason := "workflow execution completed and requested a reply side effect"
	eventKind := "workflow_execution_completed"
	if partial {
		reason = "workflow execution completed partially and requested a reply side effect"
		eventKind = "workflow_execution_completed_partial"
	}
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       reason,
			Events: []DomainEventDescriptor{{
				Kind: eventKind,
			}},
			Effects: []EffectRequest{{
				Kind:           EffectPostSlackReply,
				Status:         EffectQueued,
				IdempotencyKey: "post_reply",
			}},
		},
		NextState:     WorkflowStateReplyPending,
		ExpectedState: snapshot.State,
		AllowedNext:   []WorkflowStateKind{WorkflowStateReplyPending},
	}
}

func reduceExecutionCompletedNoReply(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return reduceExecutionCompletedWithoutReply(snapshot, false)
}

func reduceExecutionCompletedPartialNoReply(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return reduceExecutionCompletedWithoutReply(snapshot, true)
}

func reduceExecutionCompletedWithoutReply(snapshot WorkflowSnapshot, partial bool) WorkflowDecision {
	reason := "workflow execution finished without a reply side effect and workflow can terminate"
	eventKind := "workflow_execution_completed_no_reply"
	trigger := "workflow_execution_completed_no_reply"
	if partial {
		reason = "workflow execution finished partially with no reply side effect, and workflow can terminate"
		eventKind = "workflow_execution_completed_partial_no_reply"
		trigger = "workflow_execution_completed_partial_no_reply"
	}
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       reason,
			Events: []DomainEventDescriptor{{
				Kind: eventKind,
			}},
			Commands: workflowEvalCommands(snapshot.TraceID, trigger),
		},
		NextState:     WorkflowStateCompleted,
		ExpectedState: snapshot.State,
		AllowedNext:   []WorkflowStateKind{WorkflowStateCompleted},
	}
}

func reduceReplyPosted(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return reduceReplyPostedWithVerdict(snapshot, false)
}

func reduceReplyPostedPartial(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return reduceReplyPostedWithVerdict(snapshot, true)
}

func reduceReplyPostedWithVerdict(snapshot WorkflowSnapshot, partial bool) WorkflowDecision {
	reason := "reply side effect completed and workflow can terminate"
	eventKind := "workflow_reply_posted"
	trigger := "reply_posted"
	if partial {
		reason = "partial reply side effect completed and workflow can terminate"
		eventKind = "workflow_reply_posted_partial"
		trigger = "reply_posted_partial"
	}
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       reason,
			Events: []DomainEventDescriptor{{
				Kind: eventKind,
			}},
			Commands: workflowEvalCommands(snapshot.TraceID, trigger),
		},
		NextState:     WorkflowStateCompleted,
		ExpectedState: snapshot.State,
		AllowedNext:   []WorkflowStateKind{WorkflowStateCompleted},
	}
}

func reduceWorkflowNeedsHuman(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "workflow execution encountered a needs-human condition",
			Events: []DomainEventDescriptor{{
				Kind: "workflow_execution_needs_human",
			}},
			Commands: workflowEvalCommands(snapshot.TraceID, "workflow_execution_needs_human"),
		},
		NextState:     WorkflowStateNeedsHuman,
		ExpectedState: snapshot.State,
		AllowedNext:   []WorkflowStateKind{WorkflowStateNeedsHuman},
	}
}

func reduceWorkflowFailed(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "workflow execution encountered a terminal failure",
			Events: []DomainEventDescriptor{{
				Kind: "workflow_execution_failed",
			}},
			Commands: workflowEvalCommands(snapshot.TraceID, "workflow_execution_failed"),
		},
		NextState:     WorkflowStateFailed,
		ExpectedState: snapshot.State,
		AllowedNext:   []WorkflowStateKind{WorkflowStateFailed},
	}
}

func reduceWorkflowSuperseded(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "workflow was superseded by a newer line of evidence",
			Events: []DomainEventDescriptor{{
				Kind: "workflow_superseded",
			}},
		},
		NextState:     WorkflowStateSuperseded,
		ExpectedState: snapshot.State,
		AllowedNext:   []WorkflowStateKind{WorkflowStateSuperseded},
	}
}

func rejectWorkflow(state WorkflowStateKind, command WorkflowCommandKind, reason string) WorkflowDecision {
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionReject,
			Reason:       reason,
		},
		ExpectedState: state,
	}
}

func WorkflowExecutionCompletionCommand(completionVerdict string, hasReplyAction bool) WorkflowCommandKind {
	if hasReplyAction {
		if completionVerdict == "partial" {
			return CommandWorkflowExecutionCompletedPartial
		}
		return CommandWorkflowExecutionCompleted
	}
	if completionVerdict == "partial" {
		return CommandWorkflowExecutionCompletedPartialNoReply
	}
	return CommandWorkflowExecutionCompletedNoReply
}

func WorkflowReplyPostedCommand(completionVerdict string) WorkflowCommandKind {
	if completionVerdict == "partial" {
		return CommandReplyPostedPartial
	}
	return CommandReplyPosted
}

func WorkflowCommandLastVerdict(command WorkflowCommandKind) string {
	switch command {
	case CommandWorkflowExecutionCompletedPartial, CommandWorkflowExecutionCompletedPartialNoReply, CommandReplyPostedPartial:
		return "partial"
	default:
		return ""
	}
}

func workflowUsesChildExecution(payload map[string]any) bool {
	if payload == nil {
		return false
	}
	value, _ := payload["execution_strategy"].(string)
	return value == "read_heavy_slack_qna"
}

func workflowEvalCommands(traceID string, trigger string) []CommandDescriptor {
	if traceID == "" {
		return nil
	}
	return []CommandDescriptor{{
		MachineKind: MachineProblemLine,
		AggregateID: traceID,
		CommandKind: string(CommandProblemLineEvaluateTrace),
		Payload: map[string]any{
			"trigger": trigger,
		},
	}}
}
