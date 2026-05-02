package runnerutil

import (
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
)

func TestParseStructuredOutputPrefersExecutionEnvelope(t *testing.T) {
	resp := clients.RunnerResponse{
		OK:       true,
		Provider: "fake",
		Message:  "legacy message",
		Raw: map[string]any{
			"execution_envelope": map[string]any{
				"contract_version": "execution-envelope/v1",
				"final_response":   "Envelope answer",
				"artifacts": []any{
					map[string]any{
						"kind":           "diagram",
						"title":          "Architecture",
						"artifact_refs":  []any{"file:///workspace/company/artifacts/diagram.html"},
						"workspace_path": "/workspace/company/artifacts/diagram.html",
						"file_ref":       "file:///workspace/company/artifacts/diagram.html",
						"size_bytes":     42,
						"sha256":         "abc123",
					},
				},
				"deliveries": []any{
					map[string]any{
						"send_status":  "posted",
						"channel_id":   "C123",
						"thread_ts":    "171000001.000100",
						"body":         "Envelope answer",
						"provider_ref": "171000001.000101",
					},
				},
				"completion": map[string]any{
					"completion_verdict": "complete",
					"termination_reason": "normal_completion",
				},
			},
		},
	}

	out, err := ParseStructuredOutput(resp)
	if err != nil {
		t.Fatalf("ParseStructuredOutput() error = %v", err)
	}
	if out.FinalAnswer != "Envelope answer" {
		t.Fatalf("FinalAnswer = %q, want envelope answer", out.FinalAnswer)
	}
	if len(out.ProducedArtifacts) != 1 || out.ProducedArtifacts[0].WorkspacePath != "/workspace/company/artifacts/diagram.html" {
		t.Fatalf("ProducedArtifacts = %#v", out.ProducedArtifacts)
	}
	if out.ReplyDelivery.Status != "posted" || out.ReplyDelivery.ChannelID != "C123" {
		t.Fatalf("ReplyDelivery = %#v", out.ReplyDelivery)
	}
}

func TestParseExecutionEnvelopeRejectsMissingContractVersion(t *testing.T) {
	_, ok, err := ParseExecutionEnvelope(clients.RunnerResponse{
		Raw: map[string]any{
			"execution_envelope": map[string]any{"final_response": "missing version"},
		},
	})
	if !ok {
		t.Fatal("expected execution envelope to be present")
	}
	if err == nil {
		t.Fatal("expected missing contract_version error")
	}
}

func TestParseExecutionEnvelopeAcceptsV2WithoutLegacyLeaseFields(t *testing.T) {
	envelope, ok, err := ParseExecutionEnvelope(clients.RunnerResponse{
		Raw: map[string]any{
			"execution_envelope": map[string]any{
				"contract_version": "execution-envelope/v2",
				"execution_id":     "hexec-v2",
				"execution_plan": map[string]any{
					"phases": []any{map[string]any{"phase_id": "operate", "phase_type": "operate"}},
				},
				"phase_runs": []any{map[string]any{"phase_id": "operate", "phase_type": "operate", "status": "completed"}},
			},
		},
	})
	if err != nil || !ok {
		t.Fatalf("ParseExecutionEnvelope() ok=%v err=%v", ok, err)
	}
	if envelope.ContractVersion != "execution-envelope/v2" || envelope.ExecutionID != "hexec-v2" {
		t.Fatalf("unexpected envelope: %#v", envelope)
	}
}

func TestParseExecutionEnvelopeAcceptsLegacyV1FieldsAndIgnoresLeases(t *testing.T) {
	envelope, ok, err := ParseExecutionEnvelope(clients.RunnerResponse{
		Raw: map[string]any{
			"execution_envelope": map[string]any{
				"contract_version":  "execution-envelope/v1",
				"execution_id":      "hexec-v1",
				"capability_leases": []any{map[string]any{"capability": "artifact_write"}},
				"phase_runs": []any{
					map[string]any{"phase_id": "operate", "phase_type": "operate", "status": "completed", "required_leases": []any{"artifact_write"}},
				},
			},
		},
	})
	if err != nil || !ok {
		t.Fatalf("ParseExecutionEnvelope() ok=%v err=%v", ok, err)
	}
	if envelope.ContractVersion != "execution-envelope/v1" || envelope.ExecutionID != "hexec-v1" {
		t.Fatalf("unexpected envelope: %#v", envelope)
	}
	if len(envelope.PhaseRuns) != 1 || envelope.PhaseRuns[0].PhaseID != "operate" {
		t.Fatalf("unexpected phase runs: %#v", envelope.PhaseRuns)
	}
}

func TestParseStructuredOutputLegacyReplyDeliveryStatus(t *testing.T) {
	resp := clients.RunnerResponse{
		OK:       true,
		Provider: "fake",
		Message:  "legacy message",
		Raw: map[string]any{
			"structured_output": map[string]any{
				"final_answer": "Legacy answer",
				"reply_delivery": map[string]any{
					"status":       "posted",
					"channel_id":   "C456",
					"thread_ts":    "171000002.000200",
					"body":         "Legacy reply",
					"provider_ref": "171000002.000201",
				},
			},
		},
	}

	out, err := ParseStructuredOutput(resp)
	if err != nil {
		t.Fatalf("ParseStructuredOutput() error = %v", err)
	}
	if out.FinalAnswer != "Legacy answer" {
		t.Fatalf("FinalAnswer = %q, want legacy answer", out.FinalAnswer)
	}
	if out.ReplyDelivery.Status != "posted" {
		t.Fatalf("ReplyDelivery.Status = %q, want %q", out.ReplyDelivery.Status, "posted")
	}
	if out.ReplyDelivery.ChannelID != "C456" {
		t.Fatalf("ReplyDelivery.ChannelID = %q, want %q", out.ReplyDelivery.ChannelID, "C456")
	}
}

func TestExecutionLedgerEventsFromRunnerRaw(t *testing.T) {
	occurredAt := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	items := ExecutionLedgerEventsFromRunnerRaw(map[string]any{
		"execution_envelope": map[string]any{
			"contract_version": "execution-envelope/v1",
			"execution_id":     "hexec-123",
			"operation_id":     "eff-123",
			"trace_id":         "trace-123",
			"workflow_id":      "wf-123",
			"ledger_events": []any{
				map[string]any{
					"event_id":        "ledger-1",
					"kind":            "artifact.created",
					"phase_id":        "render",
					"status":          "completed",
					"sequence":        7,
					"idempotency_key": "idem-1",
					"payload": map[string]any{
						"file_ref": "file:///workspace/company/artifacts/a.html",
					},
					"recorded_at": "2026-04-24T12:01:02Z",
				},
			},
		},
	}, occurredAt)
	if len(items) != 1 {
		t.Fatalf("ledger events = %#v", items)
	}
	if items[0].ExecutionID != "hexec-123" || items[0].Seq != 7 || items[0].Kind != "artifact.created" {
		t.Fatalf("unexpected event: %#v", items[0])
	}
	if items[0].RecordedAt.Equal(occurredAt) {
		t.Fatalf("expected recorded_at from envelope, got fallback")
	}
}
