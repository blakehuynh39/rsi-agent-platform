package transition

type SettingsCommandKind string

const CommandSettingsUpdate SettingsCommandKind = "settings_update"

type SettingsSnapshot struct {
	ActiveProposalCap int `json:"active_proposal_cap"`
}

type SettingsDecision struct {
	TransitionDecision
	NextCap int `json:"next_cap,omitempty"`
}

func ReduceSettings(snapshot SettingsSnapshot, command CommandEnvelope) SettingsDecision {
	if SettingsCommandKind(command.CommandKind) != CommandSettingsUpdate {
		return SettingsDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionReject,
				Reason:       "unsupported settings command",
			},
			NextCap: snapshot.ActiveProposalCap,
		}
	}
	rawCap, _ := command.Payload["active_proposal_cap"].(float64)
	if rawCap == 0 {
		if intCap, ok := command.Payload["active_proposal_cap"].(int); ok {
			rawCap = float64(intCap)
		}
	}
	nextCap := int(rawCap)
	if nextCap == snapshot.ActiveProposalCap {
		return SettingsDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionNoop,
				Reason:       "settings already match requested proposal cap",
			},
			NextCap: snapshot.ActiveProposalCap,
		}
	}
	return SettingsDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionAdvance,
			Reason:       "platform settings updated",
			Events: []DomainEventDescriptor{{
				Kind: "platform_settings_updated",
			}},
		},
		NextCap: nextCap,
	}
}
