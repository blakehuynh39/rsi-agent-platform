package runnerutil

import (
	"testing"

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
