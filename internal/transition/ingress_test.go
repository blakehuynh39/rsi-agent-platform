package transition

import "testing"

func TestReduceIngressEventAdvances(t *testing.T) {
	decision := ReduceIngress(IngressSnapshot{}, CommandEnvelope{
		MachineKind: MachineIngress,
		CommandKind: string(CommandIngressRecordEvent),
		CommandID:   "cmd-ingress-event",
	})
	if decision.DecisionKind != DecisionAdvance {
		t.Fatalf("expected advance, got %+v", decision)
	}
	if decision.NextState != IngressStateMaterialized {
		t.Fatalf("expected materialized state, got %s", decision.NextState)
	}
}

func TestReduceIngressRejectsUnknownCommand(t *testing.T) {
	decision := ReduceIngress(IngressSnapshot{}, CommandEnvelope{
		MachineKind: MachineIngress,
		CommandKind: "unsupported",
		CommandID:   "cmd-ingress-unsupported",
	})
	if decision.DecisionKind != DecisionReject {
		t.Fatalf("expected reject, got %+v", decision)
	}
}
