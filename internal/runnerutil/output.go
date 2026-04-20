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

type ProducedArtifact struct {
	Kind           string   `json:"kind"`
	Title          string   `json:"title,omitempty"`
	ArtifactRefs   []string `json:"artifact_refs,omitempty"`
	DeliveryStatus string   `json:"delivery_status,omitempty"`
	FailureReason  string   `json:"failure_reason,omitempty"`
}

type RetryAssessment struct {
	FailureClass             string   `json:"failure_class,omitempty"`
	FailureSummary           string   `json:"failure_summary,omitempty"`
	RetryDecision            string   `json:"retry_decision,omitempty"`
	MaterialHypothesisChange bool     `json:"material_hypothesis_change,omitempty"`
	ChangedFiles             []string `json:"changed_files,omitempty"`
}

type StructuredOutput struct {
	ContextSummary        string              `json:"context_summary"`
	ReplyDraft            string              `json:"reply_draft"`
	FinalAnswer           string              `json:"final_answer"`
	Confidence            float64             `json:"confidence"`
	SelfCritique          string              `json:"self_critique"`
	VisibleReasoning      []Step              `json:"visible_reasoning"`
	ProposedActions       []ProposedAction    `json:"proposed_actions"`
	KnowledgeDrafts       []KnowledgeDraft    `json:"knowledge_drafts"`
	OutcomeHypotheses     []OutcomeHypothesis `json:"outcome_hypotheses"`
	ProducedArtifacts     []ProducedArtifact  `json:"produced_artifacts"`
	ArtifactFailureReason string              `json:"artifact_failure_reason,omitempty"`
	ChangePlan            string              `json:"change_plan,omitempty"`
	RepoPatch             string              `json:"repo_patch,omitempty"`
	ValidationPlan        string              `json:"validation_plan,omitempty"`
	RetryAssessment       RetryAssessment     `json:"retry_assessment,omitempty"`
	HypothesisDelta       string              `json:"hypothesis_delta,omitempty"`
}

type RuntimeDiagnosisOutput struct {
	Status          string   `json:"status"`
	Subsystem       string   `json:"subsystem"`
	FailureMode     string   `json:"failure_mode"`
	Summary         string   `json:"summary"`
	EvidenceRefs    []string `json:"evidence_refs"`
	MissingEvidence []string `json:"missing_evidence"`
	RecommendedFix  string   `json:"recommended_fix"`
	TargetSurface   string   `json:"target_surface"`
	ValidationPlan  string   `json:"validation_plan"`
}

func ParseStructuredOutput(resp clients.RunnerResponse) (StructuredOutput, error) {
	raw, ok := resp.Raw["structured_output"]
	if !ok {
		return StructuredOutput{}, fmt.Errorf("runner response missing structured_output")
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return StructuredOutput{}, fmt.Errorf("marshal runner structured_output: %w", err)
	}
	var out StructuredOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return StructuredOutput{}, fmt.Errorf("parse runner structured_output: %w", err)
	}
	return out, nil
}

func ParseRuntimeDiagnosisOutput(resp clients.RunnerResponse) (RuntimeDiagnosisOutput, error) {
	raw, ok := resp.Raw["structured_output"]
	if !ok {
		return RuntimeDiagnosisOutput{}, fmt.Errorf("runner response missing structured_output")
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return RuntimeDiagnosisOutput{}, fmt.Errorf("marshal runner structured_output: %w", err)
	}
	var out RuntimeDiagnosisOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return RuntimeDiagnosisOutput{}, fmt.Errorf("parse runtime diagnosis structured_output: %w", err)
	}
	return out, nil
}

func ToTraceReasoning(traceID string, workflowID string, output StructuredOutput, createdAt time.Time) []events.ReasoningStep {
	out := make([]events.ReasoningStep, 0, len(output.VisibleReasoning))
	for index, step := range output.VisibleReasoning {
		stepType := strings.TrimSpace(step.StepType)
		if stepType == "" {
			stepType = "visible_reasoning"
		}
		out = append(out, events.ReasoningStep{
			ID:           fmt.Sprintf("reason-runner-%d-%d", createdAt.UnixNano(), index),
			TraceID:      traceID,
			WorkflowID:   workflowID,
			StepType:     stepType,
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
