package runnerutil

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/events"
)

type Step struct {
	StepType     string               `json:"step_type"`
	Summary      string               `json:"summary"`
	EvidenceRefs []events.EvidenceRef `json:"evidence_refs"`
	Alternatives []string             `json:"alternatives"`
	Confidence   float64              `json:"confidence"`
	Decision     string               `json:"decision"`
}

type ProposedAction struct {
	Kind           string               `json:"kind"`
	TargetRef      string               `json:"target_ref,omitempty"`
	RequestPayload map[string]any       `json:"request_payload,omitempty"`
	ApprovalMode   string               `json:"approval_mode,omitempty"`
	IdempotencyKey string               `json:"idempotency_key,omitempty"`
	Rationale      string               `json:"rationale,omitempty"`
	EvidenceRefs   []events.EvidenceRef `json:"evidence_refs,omitempty"`
}

type KnowledgeDraft struct {
	Kind         string               `json:"kind"`
	ScopeType    string               `json:"scope_type"`
	ScopeID      string               `json:"scope_id,omitempty"`
	Title        string               `json:"title"`
	Summary      string               `json:"summary,omitempty"`
	Body         string               `json:"body,omitempty"`
	Confidence   float64              `json:"confidence,omitempty"`
	FreshUntil   string               `json:"fresh_until,omitempty"`
	EvidenceRefs []events.EvidenceRef `json:"evidence_refs,omitempty"`
}

type OutcomeHypothesis struct {
	OutcomeType         string `json:"outcome_type"`
	SuccessCondition    string `json:"success_condition"`
	MeasurementRef      string `json:"measurement_ref,omitempty"`
	ExpectedTimeHorizon string `json:"expected_time_horizon,omitempty"`
}

type StructuredOutput struct {
	ContextSummary    string              `json:"context_summary"`
	ReplyDraft        string              `json:"reply_draft"`
	FinalAnswer       string              `json:"final_answer"`
	Confidence        float64             `json:"confidence"`
	SelfCritique      string              `json:"self_critique"`
	VisibleReasoning  []Step              `json:"visible_reasoning"`
	ProposedActions   []ProposedAction    `json:"proposed_actions"`
	KnowledgeDrafts   []KnowledgeDraft    `json:"knowledge_drafts"`
	OutcomeHypotheses []OutcomeHypothesis `json:"outcome_hypotheses"`
}

func ParseStructuredOutput(resp clients.RunnerResponse) StructuredOutput {
	if raw, ok := resp.Raw["structured_output"]; ok {
		data, _ := json.Marshal(raw)
		var out StructuredOutput
		if err := json.Unmarshal(data, &out); err == nil {
			return out
		}
	}
	var out StructuredOutput
	if err := json.Unmarshal([]byte(resp.Message), &out); err == nil {
		return out
	}
	return StructuredOutput{
		FinalAnswer: resp.Message,
		Confidence:  0.5,
		VisibleReasoning: []Step{
			{
				StepType:   "fallback",
				Summary:    "Runner returned unstructured output; stored raw response as the visible answer.",
				Confidence: 0.5,
				Decision:   resp.Message,
			},
		},
		ProposedActions:   []ProposedAction{},
		KnowledgeDrafts:   []KnowledgeDraft{},
		OutcomeHypotheses: []OutcomeHypothesis{},
	}
}

func ToTraceReasoning(traceID string, workflowID string, output StructuredOutput, createdAt time.Time) []events.ReasoningStep {
	out := make([]events.ReasoningStep, 0, len(output.VisibleReasoning))
	for index, step := range output.VisibleReasoning {
		out = append(out, events.ReasoningStep{
			ID:           fmt.Sprintf("reason-runner-%d-%d", createdAt.UnixNano(), index),
			TraceID:      traceID,
			WorkflowID:   workflowID,
			StepType:     firstNonEmpty(step.StepType, "visible_reasoning"),
			Summary:      step.Summary,
			EvidenceRefs: normalizeEvidenceRefs(step.EvidenceRefs),
			Alternatives: normalizeStrings(step.Alternatives),
			Confidence:   step.Confidence,
			Decision:     step.Decision,
			CreatedAt:    createdAt,
		})
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func normalizeStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func normalizeEvidenceRefs(values []events.EvidenceRef) []events.EvidenceRef {
	if values == nil {
		return []events.EvidenceRef{}
	}
	return values
}
