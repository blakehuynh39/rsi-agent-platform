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
	StepType     string   `json:"step_type"`
	Summary      string   `json:"summary"`
	Alternatives []string `json:"alternatives"`
	Confidence   float64  `json:"confidence"`
	Decision     string   `json:"decision"`
}

type StructuredOutput struct {
	ContextSummary   string  `json:"context_summary"`
	ReplyDraft       string  `json:"reply_draft"`
	FinalAnswer      string  `json:"final_answer"`
	Confidence       float64 `json:"confidence"`
	SelfCritique     string  `json:"self_critique"`
	VisibleReasoning []Step  `json:"visible_reasoning"`
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
