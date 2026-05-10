package transition

import (
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/action"
)

type ActionExecutionCommandKind string

const (
	CommandActionQueue   ActionExecutionCommandKind = "action_queued"
	CommandActionStart   ActionExecutionCommandKind = "action_started"
	CommandActionSucceed ActionExecutionCommandKind = "action_succeeded"
	CommandActionBlock   ActionExecutionCommandKind = "action_blocked"
	CommandActionFail    ActionExecutionCommandKind = "action_failed"
	CommandActionCancel  ActionExecutionCommandKind = "action_canceled"
)

type ActionExecutionSnapshot struct {
	State action.Status `json:"state"`
	Kind  action.Kind   `json:"kind,omitempty"`
}

type ActionExecutionDecision struct {
	TransitionDecision
	NextState     action.Status   `json:"next_state,omitempty"`
	ExpectedState action.Status   `json:"expected_state,omitempty"`
	AllowedNext   []action.Status `json:"allowed_next,omitempty"`
}

type actionExecutionReducer func(snapshot ActionExecutionSnapshot, command CommandEnvelope) ActionExecutionDecision

var actionExecutionReducers = map[action.Status]map[ActionExecutionCommandKind]actionExecutionReducer{
	"": {
		CommandActionQueue: reduceActionQueue,
	},
	action.StatusDrafted: {
		CommandActionQueue:  reduceActionQueueNoop,
		CommandActionStart:  reduceActionStart,
		CommandActionCancel: reduceActionCancel,
	},
	action.StatusApproved: {
		CommandActionQueue:  reduceActionQueueNoop,
		CommandActionStart:  reduceActionStart,
		CommandActionCancel: reduceActionCancel,
	},
	action.StatusQueued: {
		CommandActionQueue:  reduceActionQueueNoop,
		CommandActionStart:  reduceActionStart,
		CommandActionCancel: reduceActionCancel,
	},
	action.StatusExecuting: {
		CommandActionQueue:   reduceActionQueueNoop,
		CommandActionSucceed: reduceActionSucceeded,
		CommandActionBlock:   reduceActionBlocked,
		CommandActionFail:    reduceActionFailed,
		CommandActionCancel:  reduceActionCancel,
	},
	action.StatusSucceeded: {
		CommandActionQueue: reduceActionQueueNoop,
	},
	action.StatusBlocked: {
		CommandActionQueue: reduceActionQueueNoop,
	},
	action.StatusFailed: {
		CommandActionQueue: reduceActionQueueNoop,
	},
	action.StatusCanceled: {
		CommandActionQueue: reduceActionQueueNoop,
	},
	action.StatusSuperseded: {
		CommandActionQueue: reduceActionQueueNoop,
	},
}

func ReduceActionExecution(snapshot ActionExecutionSnapshot, command CommandEnvelope) ActionExecutionDecision {
	commandKind := ActionExecutionCommandKind(command.CommandKind)
	reducers, ok := actionExecutionReducers[snapshot.State]
	if !ok {
		return rejectActionExecution(snapshot.State, commandKind, "unsupported action state")
	}
	reducer, ok := reducers[commandKind]
	if !ok {
		return rejectActionExecution(snapshot.State, commandKind, "unsupported action command for current state")
	}
	return reducer(snapshot, command)
}

func reduceActionStart(snapshot ActionExecutionSnapshot, _ CommandEnvelope) ActionExecutionDecision {
	return ActionExecutionDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "action execution started",
			Events: []DomainEventDescriptor{{
				Kind: "action_execution_started",
			}},
		},
		NextState:     action.StatusExecuting,
		ExpectedState: snapshot.State,
		AllowedNext:   []action.Status{action.StatusExecuting},
	}
}

func reduceActionQueue(snapshot ActionExecutionSnapshot, command CommandEnvelope) ActionExecutionDecision {
	effects := []EffectRequest{}
	switch action.Kind(strings.TrimSpace(asString(command.Payload["kind"]))) {
	case action.KindToolRead, action.KindSlackPost, action.KindSlackReport:
		effects = append(effects, EffectRequest{
			Kind:           EffectInvokeAction,
			Status:         EffectQueued,
			IdempotencyKey: "invoke_action",
		})
	}
	return ActionExecutionDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "action execution queued",
			Events: []DomainEventDescriptor{{
				Kind: "action_execution_queued",
			}},
			Effects: effects,
		},
		NextState:     action.StatusQueued,
		ExpectedState: snapshot.State,
		AllowedNext:   []action.Status{action.StatusQueued},
	}
}

func reduceActionQueueNoop(snapshot ActionExecutionSnapshot, _ CommandEnvelope) ActionExecutionDecision {
	return ActionExecutionDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionNoop,
			Reason:       "action execution already queued",
		},
		NextState:     snapshot.State,
		ExpectedState: snapshot.State,
	}
}

func reduceActionSucceeded(snapshot ActionExecutionSnapshot, _ CommandEnvelope) ActionExecutionDecision {
	return ActionExecutionDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "action execution succeeded",
			Events: []DomainEventDescriptor{{
				Kind: "action_execution_succeeded",
			}},
		},
		NextState:     action.StatusSucceeded,
		ExpectedState: snapshot.State,
		AllowedNext:   []action.Status{action.StatusSucceeded},
	}
}

func reduceActionBlocked(snapshot ActionExecutionSnapshot, _ CommandEnvelope) ActionExecutionDecision {
	return ActionExecutionDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "action execution blocked",
			Events: []DomainEventDescriptor{{
				Kind: "action_execution_blocked",
			}},
		},
		NextState:     action.StatusBlocked,
		ExpectedState: snapshot.State,
		AllowedNext:   []action.Status{action.StatusBlocked},
	}
}

func reduceActionFailed(snapshot ActionExecutionSnapshot, _ CommandEnvelope) ActionExecutionDecision {
	return ActionExecutionDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "action execution failed",
			Events: []DomainEventDescriptor{{
				Kind: "action_execution_failed",
			}},
		},
		NextState:     action.StatusFailed,
		ExpectedState: snapshot.State,
		AllowedNext:   []action.Status{action.StatusFailed},
	}
}

func reduceActionCancel(snapshot ActionExecutionSnapshot, _ CommandEnvelope) ActionExecutionDecision {
	return ActionExecutionDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "action execution canceled",
			Events: []DomainEventDescriptor{{
				Kind: "action_execution_canceled",
			}},
		},
		NextState:     action.StatusCanceled,
		ExpectedState: snapshot.State,
		AllowedNext:   []action.Status{action.StatusCanceled},
	}
}

func rejectActionExecution(state action.Status, command ActionExecutionCommandKind, reason string) ActionExecutionDecision {
	return ActionExecutionDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionReject,
			Reason:       reason,
		},
		ExpectedState: state,
	}
}

func asString(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}
