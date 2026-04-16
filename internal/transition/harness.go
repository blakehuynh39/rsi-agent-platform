package transition

import "github.com/piplabs/rsi-agent-platform/internal/harness"

type HarnessCommandKind string

const (
	CommandHarnessActivateOverlay HarnessCommandKind = "harness_activate_overlay"
	CommandHarnessBindSession     HarnessCommandKind = "harness_bind_session"
	CommandHarnessRecordExecution HarnessCommandKind = "harness_record_execution"
)

type HarnessSnapshot struct {
	OverlayStatus harness.OverlayStatus `json:"overlay_status,omitempty"`
	SessionBound  bool                  `json:"session_bound,omitempty"`
	ExecutionSeen bool                  `json:"execution_seen,omitempty"`
}

type HarnessDecision struct {
	TransitionDecision
	NextStatus harness.OverlayStatus `json:"next_status,omitempty"`
}

func ReduceHarness(snapshot HarnessSnapshot, command CommandEnvelope) HarnessDecision {
	switch HarnessCommandKind(command.CommandKind) {
	case CommandHarnessActivateOverlay:
		reason := "harness overlay activated and experiment recorded"
		if snapshot.OverlayStatus == harness.OverlayStatusActive {
			reason = "harness overlay activation refreshed and experiment recorded"
		}
		return HarnessDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       reason,
				Events: []DomainEventDescriptor{
					{Kind: "harness_overlay_activated"},
					{Kind: "harness_experiment_recorded"},
				},
			},
			NextStatus: harness.OverlayStatusActive,
		}
	case CommandHarnessBindSession:
		reason := "harness session binding recorded"
		eventKind := "harness_session_bound"
		if snapshot.SessionBound {
			reason = "harness session binding refreshed"
			eventKind = "harness_session_refreshed"
		}
		return HarnessDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       reason,
				Events: []DomainEventDescriptor{
					{Kind: eventKind},
				},
			},
			NextStatus: snapshot.OverlayStatus,
		}
	case CommandHarnessRecordExecution:
		reason := "harness execution recorded"
		if snapshot.ExecutionSeen {
			reason = "harness execution refreshed"
		}
		return HarnessDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       reason,
				Events: []DomainEventDescriptor{
					{Kind: "harness_execution_recorded"},
				},
			},
			NextStatus: snapshot.OverlayStatus,
		}
	default:
		return HarnessDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionReject,
				Reason:       "unsupported harness command for current state",
			},
			NextStatus: snapshot.OverlayStatus,
		}
	}
}
