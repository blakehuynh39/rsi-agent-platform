package transition

type RuntimeDiagnosisState string

const (
	RuntimeDiagnosisStateQueued        RuntimeDiagnosisState = "queued"
	RuntimeDiagnosisStateInvestigating RuntimeDiagnosisState = "investigating"
	RuntimeDiagnosisStateGrounded      RuntimeDiagnosisState = "grounded"
	RuntimeDiagnosisStateNeedsEvidence RuntimeDiagnosisState = "needs_evidence"
	RuntimeDiagnosisStatePromoted      RuntimeDiagnosisState = "promoted"
	RuntimeDiagnosisStateClosed        RuntimeDiagnosisState = "closed"
)

type RuntimeDiagnosisCommandKind string

const (
	CommandRuntimeDiagnosisQueue         RuntimeDiagnosisCommandKind = "runtime_diagnosis_queue"
	CommandRuntimeDiagnosisRunnerStarted RuntimeDiagnosisCommandKind = "runtime_diagnosis_runner_started"
	CommandRuntimeDiagnosisRecordResult  RuntimeDiagnosisCommandKind = "runtime_diagnosis_record_result"
	CommandRuntimeDiagnosisMarkPromoted  RuntimeDiagnosisCommandKind = "runtime_diagnosis_mark_promoted"
	CommandRuntimeDiagnosisClose         RuntimeDiagnosisCommandKind = "runtime_diagnosis_close"
)

type RuntimeDiagnosisSnapshot struct {
	State  RuntimeDiagnosisState `json:"state,omitempty"`
	Exists bool                  `json:"exists,omitempty"`
}

type RuntimeDiagnosisDecision struct {
	TransitionDecision
	NextState RuntimeDiagnosisState `json:"next_state,omitempty"`
}

func ReduceRuntimeDiagnosis(snapshot RuntimeDiagnosisSnapshot, command CommandEnvelope) RuntimeDiagnosisDecision {
	current := snapshot.State
	if current == "" {
		current = RuntimeDiagnosisStateQueued
	}
	switch RuntimeDiagnosisCommandKind(command.CommandKind) {
	case CommandRuntimeDiagnosisQueue:
		if current == RuntimeDiagnosisStateInvestigating {
			return noopRuntimeDiagnosis(current, "runtime diagnosis already investigating")
		}
		return RuntimeDiagnosisDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "runtime diagnosis queued for governed investigation",
				Events: []DomainEventDescriptor{{
					Kind: "runtime_diagnosis_queued",
				}},
				Effects: []EffectRequest{{
					Kind:           EffectInvokeRunner,
					Status:         EffectQueued,
					IdempotencyKey: command.CommandID,
				}},
			},
			NextState: RuntimeDiagnosisStateQueued,
		}
	case CommandRuntimeDiagnosisRunnerStarted:
		return RuntimeDiagnosisDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "runtime diagnosis runner investigation started",
				Events: []DomainEventDescriptor{{
					Kind: "runtime_diagnosis_runner_started",
				}},
			},
			NextState: RuntimeDiagnosisStateInvestigating,
		}
	case CommandRuntimeDiagnosisRecordResult:
		status := RuntimeDiagnosisState(commandPayloadString(command, "status"))
		switch status {
		case RuntimeDiagnosisStateGrounded, RuntimeDiagnosisStateNeedsEvidence, RuntimeDiagnosisStateClosed:
		default:
			return rejectRuntimeDiagnosis(current, "runtime diagnosis result requires grounded, needs_evidence, or closed status")
		}
		return RuntimeDiagnosisDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "runtime diagnosis result recorded",
				Events: []DomainEventDescriptor{{
					Kind: "runtime_diagnosis_result_recorded",
				}},
			},
			NextState: status,
		}
	case CommandRuntimeDiagnosisMarkPromoted:
		return RuntimeDiagnosisDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "runtime diagnosis promoted into a concrete proposal",
				Events: []DomainEventDescriptor{{
					Kind: "runtime_diagnosis_promoted",
				}},
			},
			NextState: RuntimeDiagnosisStatePromoted,
		}
	case CommandRuntimeDiagnosisClose:
		return RuntimeDiagnosisDecision{
			TransitionDecision: TransitionDecision{
				DecisionKind: DecisionAdvance,
				Reason:       "runtime diagnosis closed",
				Events: []DomainEventDescriptor{{
					Kind: "runtime_diagnosis_closed",
				}},
			},
			NextState: RuntimeDiagnosisStateClosed,
		}
	default:
		return rejectRuntimeDiagnosis(current, "unsupported runtime diagnosis command")
	}
}

func noopRuntimeDiagnosis(state RuntimeDiagnosisState, reason string) RuntimeDiagnosisDecision {
	return RuntimeDiagnosisDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionNoop,
			Reason:       reason,
		},
		NextState: state,
	}
}

func rejectRuntimeDiagnosis(state RuntimeDiagnosisState, reason string) RuntimeDiagnosisDecision {
	return RuntimeDiagnosisDecision{
		TransitionDecision: TransitionDecision{
			DecisionKind: DecisionReject,
			Reason:       reason,
		},
		NextState: state,
	}
}

func commandPayloadString(command CommandEnvelope, key string) string {
	if command.Payload == nil {
		return ""
	}
	if value, ok := command.Payload[key].(string); ok {
		return value
	}
	return ""
}
