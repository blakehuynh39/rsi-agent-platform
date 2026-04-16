package transition

type WorkflowStateKind string

const (
	WorkflowStateQueued            WorkflowStateKind = "queued"
	WorkflowStateCollectingContext WorkflowStateKind = "collecting_context"
	WorkflowStateWaitingOnActions  WorkflowStateKind = "waiting_on_actions"
	WorkflowStateReasoning         WorkflowStateKind = "reasoning"
	WorkflowStateReplyPending      WorkflowStateKind = "reply_pending"
	WorkflowStateNeedsHuman        WorkflowStateKind = "needs_human"
	WorkflowStateCompleted         WorkflowStateKind = "completed"
	WorkflowStateFailed            WorkflowStateKind = "failed"
	WorkflowStateSuperseded        WorkflowStateKind = "superseded"
)

type WorkflowCommandKind string

const (
	CommandWorkflowStarted        WorkflowCommandKind = "workflow_started"
	CommandContextActionsQueued   WorkflowCommandKind = "context_actions_queued"
	CommandContextSkipped         WorkflowCommandKind = "context_skipped"
	CommandContextCompleted       WorkflowCommandKind = "context_completed"
	CommandRunnerCompleted        WorkflowCommandKind = "runner_completed"
	CommandRunnerCompletedNoReply WorkflowCommandKind = "runner_completed_no_reply"
	CommandReplyPosted            WorkflowCommandKind = "reply_posted"
	CommandWorkflowBlocked        WorkflowCommandKind = "workflow_blocked"
	CommandWorkflowFailed         WorkflowCommandKind = "workflow_failed"
	CommandWorkflowSuperseded     WorkflowCommandKind = "workflow_superseded"
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
		CommandWorkflowStarted:    reduceWorkflowStarted,
		CommandWorkflowFailed:     reduceWorkflowFailed,
		CommandWorkflowBlocked:    reduceWorkflowBlocked,
		CommandWorkflowSuperseded: reduceWorkflowSuperseded,
	},
	WorkflowStateCollectingContext: {
		CommandContextActionsQueued: reduceContextActionsQueued,
		CommandContextSkipped:       reduceContextSkipped,
		CommandWorkflowFailed:       reduceWorkflowFailed,
		CommandWorkflowBlocked:      reduceWorkflowBlocked,
		CommandWorkflowSuperseded:   reduceWorkflowSuperseded,
	},
	WorkflowStateWaitingOnActions: {
		CommandContextCompleted:   reduceContextCompleted,
		CommandWorkflowFailed:     reduceWorkflowFailed,
		CommandWorkflowBlocked:    reduceWorkflowBlocked,
		CommandWorkflowSuperseded: reduceWorkflowSuperseded,
	},
	WorkflowStateReasoning: {
		CommandRunnerCompleted:        reduceRunnerCompleted,
		CommandRunnerCompletedNoReply: reduceRunnerCompletedNoReply,
		CommandWorkflowFailed:         reduceWorkflowFailed,
		CommandWorkflowBlocked:        reduceWorkflowBlocked,
		CommandWorkflowSuperseded:     reduceWorkflowSuperseded,
	},
	WorkflowStateReplyPending: {
		CommandReplyPosted:        reduceReplyPosted,
		CommandWorkflowFailed:     reduceWorkflowFailed,
		CommandWorkflowBlocked:    reduceWorkflowBlocked,
		CommandWorkflowSuperseded: reduceWorkflowSuperseded,
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

func reduceContextSkipped(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "no context actions were required, runner can start immediately",
			Events: []DomainEventDescriptor{{
				Kind: "workflow_context_skipped",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectInvokeRunner,
				Status:         EffectQueued,
				IdempotencyKey: "invoke_runner",
			}},
		},
		NextState:     WorkflowStateReasoning,
		ExpectedState: snapshot.State,
		AllowedNext:   []WorkflowStateKind{WorkflowStateReasoning},
	}
}

func reduceContextCompleted(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "context collection is complete and the runner can reason over the gathered evidence",
			Events: []DomainEventDescriptor{{
				Kind: "workflow_context_completed",
			}},
			Effects: []EffectRequest{{
				Kind:           EffectInvokeRunner,
				Status:         EffectQueued,
				IdempotencyKey: "invoke_runner",
			}},
		},
		NextState:     WorkflowStateReasoning,
		ExpectedState: snapshot.State,
		AllowedNext:   []WorkflowStateKind{WorkflowStateReasoning},
	}
}

func reduceRunnerCompleted(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "runner finished and requested a reply side effect",
			Events: []DomainEventDescriptor{{
				Kind: "workflow_runner_completed",
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

func reduceRunnerCompletedNoReply(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "runner finished without a reply side effect and workflow can terminate",
			Events: []DomainEventDescriptor{{
				Kind: "workflow_runner_completed_no_reply",
			}},
			Commands: workflowEvalCommands(snapshot.TraceID, "runner_completed_no_reply"),
		},
		NextState:     WorkflowStateCompleted,
		ExpectedState: snapshot.State,
		AllowedNext:   []WorkflowStateKind{WorkflowStateCompleted},
	}
}

func reduceReplyPosted(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "reply side effect completed and workflow can terminate",
			Events: []DomainEventDescriptor{{
				Kind: "workflow_reply_posted",
			}},
			Commands: workflowEvalCommands(snapshot.TraceID, "reply_posted"),
		},
		NextState:     WorkflowStateCompleted,
		ExpectedState: snapshot.State,
		AllowedNext:   []WorkflowStateKind{WorkflowStateCompleted},
	}
}

func reduceWorkflowBlocked(snapshot WorkflowSnapshot, _ CommandEnvelope) WorkflowDecision {
	return WorkflowDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "workflow encountered a needs-human condition",
			Events: []DomainEventDescriptor{{
				Kind: "workflow_blocked",
			}},
			Commands: workflowEvalCommands(snapshot.TraceID, "workflow_blocked"),
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
			Reason:       "workflow encountered a terminal failure",
			Events: []DomainEventDescriptor{{
				Kind: "workflow_failed",
			}},
			Commands: workflowEvalCommands(snapshot.TraceID, "workflow_failed"),
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
