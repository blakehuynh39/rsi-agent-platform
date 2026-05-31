package transition

type IngressStateKind string

const (
	IngressStatePending      IngressStateKind = "pending"
	IngressStateMaterialized IngressStateKind = "materialized"
)

type IngressCommandKind string

const (
	CommandIngressRecordEvent IngressCommandKind = "ingress_record_event"
	CommandIngressRecordSlack IngressCommandKind = "ingress_record_slack"
)

type IngressSnapshot struct {
	State IngressStateKind `json:"state,omitempty"`
}

type IngressDecision struct {
	TransitionDecision
	NextState IngressStateKind `json:"next_state,omitempty"`
}

func ReduceIngress(snapshot IngressSnapshot, command CommandEnvelope) IngressDecision {
	current := snapshot.State
	if current == "" {
		current = IngressStatePending
	}
	switch IngressCommandKind(command.CommandKind) {
	case CommandIngressRecordEvent:
		return IngressDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "external event normalized through ingress command path",
				Events: []DomainEventDescriptor{{
					Kind: "ingress_event_recorded",
				}},
			},
			NextState: IngressStateMaterialized,
		}
	case CommandIngressRecordSlack:
		return IngressDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "slack envelope normalized through ingress command path",
				Events: []DomainEventDescriptor{{
					Kind: "ingress_slack_recorded",
				}},
			},
			NextState: IngressStateMaterialized,
		}
	default:
		return IngressDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionReject,
				Reason:       "unsupported ingress command",
			},
			NextState: current,
		}
	}
}
