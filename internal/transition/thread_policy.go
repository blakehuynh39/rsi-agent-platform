package transition

import "github.com/piplabs/rsi-agent-platform/internal/policy"

type ThreadPolicyCommandKind string

const (
	CommandThreadMute   ThreadPolicyCommandKind = "thread_mute"
	CommandThreadResume ThreadPolicyCommandKind = "thread_resume"
)

type ThreadPolicySnapshot struct {
	State policy.ThreadState `json:"state"`
}

type ThreadPolicyDecision struct {
	TransitionDecision
	NextState policy.ThreadState `json:"next_state,omitempty"`
}

func ReduceThreadPolicy(snapshot ThreadPolicySnapshot, command CommandEnvelope) ThreadPolicyDecision {
	commandKind := ThreadPolicyCommandKind(command.CommandKind)
	switch snapshot.State {
	case policy.ThreadStateActive:
		switch commandKind {
		case CommandThreadMute:
			return ThreadPolicyDecision{
				TransitionDecision: TransitionDecision{
					DecisionKind: DecisionAdvance,
					Reason:       "thread policy moved to muted",
					Events: []DomainEventDescriptor{{
						Kind: "thread_policy_muted",
					}},
				},
				NextState: policy.ThreadStateMuted,
			}
		case CommandThreadResume:
			return ThreadPolicyDecision{
				TransitionDecision: TransitionDecision{
					DecisionKind: DecisionNoop,
					Reason:       "thread policy already active",
				},
				NextState: snapshot.State,
			}
		}
	case policy.ThreadStateMuted:
		switch commandKind {
		case CommandThreadMute:
			return ThreadPolicyDecision{
				TransitionDecision: TransitionDecision{
					DecisionKind: DecisionNoop,
					Reason:       "thread policy already muted",
				},
				NextState: snapshot.State,
			}
		case CommandThreadResume:
			return ThreadPolicyDecision{
				TransitionDecision: TransitionDecision{
					DecisionKind: DecisionAdvance,
					Reason:       "thread policy resumed to active",
					Events: []DomainEventDescriptor{{
						Kind: "thread_policy_resumed",
					}},
				},
				NextState: policy.ThreadStateActive,
			}
		}
	case policy.ThreadStateMuteUntilMention:
		switch commandKind {
		case CommandThreadMute:
			return ThreadPolicyDecision{
				TransitionDecision: TransitionDecision{
					DecisionKind: DecisionAdvance,
					Reason:       "thread policy moved from mention-only mute to full mute",
					Events: []DomainEventDescriptor{{
						Kind: "thread_policy_muted",
					}},
				},
				NextState: policy.ThreadStateMuted,
			}
		case CommandThreadResume:
			return ThreadPolicyDecision{
				TransitionDecision: TransitionDecision{
					DecisionKind: DecisionAdvance,
					Reason:       "thread policy resumed to active",
					Events: []DomainEventDescriptor{{
						Kind: "thread_policy_resumed",
					}},
				},
				NextState: policy.ThreadStateActive,
			}
		}
	case policy.ThreadStateClosed, policy.ThreadStateObserveOnly:
		return ThreadPolicyDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionReject,
				Reason:       "thread policy does not allow mute/resume from current state",
			},
			NextState: snapshot.State,
		}
	}
	return ThreadPolicyDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionReject,
			Reason:       "unsupported thread policy command for current state",
		},
		NextState: snapshot.State,
	}
}
