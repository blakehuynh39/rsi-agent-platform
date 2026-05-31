package runnerutil

import (
	"crypto/sha1"
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
	Kind                 string   `json:"kind"`
	Title                string   `json:"title,omitempty"`
	ArtifactRefs         []string `json:"artifact_refs,omitempty"`
	DeliveryStatus       string   `json:"delivery_status,omitempty"`
	FailureReason        string   `json:"failure_reason,omitempty"`
	WorkspacePath        string   `json:"workspace_path,omitempty"`
	FileRef              string   `json:"file_ref,omitempty"`
	SizeBytes            int64    `json:"size_bytes,omitempty"`
	SHA256               string   `json:"sha256,omitempty"`
	CreatedByExecutionID string   `json:"created_by_execution_id,omitempty"`
	ShareStatus          string   `json:"share_status,omitempty"`
}

type ArtifactRenderBrief struct {
	Kind           string         `json:"kind"`
	Skill          string         `json:"skill,omitempty"`
	Title          string         `json:"title,omitempty"`
	RenderPrompt   string         `json:"render_prompt,omitempty"`
	Inputs         map[string]any `json:"inputs,omitempty"`
	OutputPathHint string         `json:"output_path_hint,omitempty"`
}

type ReplyDelivery struct {
	Status        string   `json:"send_status,omitempty"`
	ChannelID     string   `json:"channel_id,omitempty"`
	ThreadTS      string   `json:"thread_ts,omitempty"`
	Body          string   `json:"body,omitempty"`
	BodySHA1      string   `json:"body_sha1,omitempty"`
	BodyExcerpt   string   `json:"body_excerpt,omitempty"`
	ToolCallID    string   `json:"tool_call_id,omitempty"`
	ToolName      string   `json:"tool_name,omitempty"`
	ProviderRef   string   `json:"provider_ref,omitempty"`
	MessageLink   string   `json:"message_link,omitempty"`
	ArtifactRefs  []string `json:"artifact_refs,omitempty"`
	FailureReason string   `json:"failure_reason,omitempty"`
}

func (rd *ReplyDelivery) UnmarshalJSON(data []byte) error {
	type Alias ReplyDelivery
	aux := &struct {
		LegacyStatus string `json:"status,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(rd),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if rd.Status == "" && aux.LegacyStatus != "" {
		rd.Status = aux.LegacyStatus
	}
	return nil
}

type RetryAssessment struct {
	FailureClass             string   `json:"failure_class,omitempty"`
	FailureSummary           string   `json:"failure_summary,omitempty"`
	RetryDecision            string   `json:"retry_decision,omitempty"`
	MaterialHypothesisChange bool     `json:"material_hypothesis_change,omitempty"`
	ChangedFiles             []string `json:"changed_files,omitempty"`
}

type StructuredOutput struct {
	SessionTitle          string                `json:"session_title,omitempty"`
	ContextSummary        string                `json:"context_summary"`
	ReplyDraft            string                `json:"reply_draft"`
	FinalAnswer           string                `json:"final_answer"`
	Confidence            float64               `json:"confidence"`
	SelfCritique          string                `json:"self_critique"`
	VisibleReasoning      []Step                `json:"visible_reasoning"`
	ProposedActions       []ProposedAction      `json:"proposed_actions"`
	ReplyDelivery         ReplyDelivery         `json:"reply_delivery,omitempty"`
	KnowledgeDrafts       []KnowledgeDraft      `json:"knowledge_drafts"`
	OutcomeHypotheses     []OutcomeHypothesis   `json:"outcome_hypotheses"`
	ArtifactRenderBriefs  []ArtifactRenderBrief `json:"artifact_render_briefs,omitempty"`
	ProducedArtifacts     []ProducedArtifact    `json:"produced_artifacts"`
	ArtifactFailureReason string                `json:"artifact_failure_reason,omitempty"`
	ChangePlan            string                `json:"change_plan,omitempty"`
	RepoPatch             string                `json:"repo_patch,omitempty"`
	ValidationPlan        string                `json:"validation_plan,omitempty"`
	RetryAssessment       RetryAssessment       `json:"retry_assessment,omitempty"`
	HypothesisDelta       string                `json:"hypothesis_delta,omitempty"`
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

type ExecutionEnvelope struct {
	ContractVersion string             `json:"contract_version"`
	ExecutionID     string             `json:"execution_id,omitempty"`
	OperationID     string             `json:"operation_id,omitempty"`
	TraceID         string             `json:"trace_id,omitempty"`
	WorkflowID      string             `json:"workflow_id,omitempty"`
	SessionID       string             `json:"session_id,omitempty"`
	ExecutionIntent map[string]any     `json:"execution_intent,omitempty"`
	DeliveryPolicy  map[string]any     `json:"delivery_policy,omitempty"`
	WorkspacePolicy map[string]any     `json:"workspace_policy,omitempty"`
	ApprovalPolicy  map[string]any     `json:"approval_policy,omitempty"`
	ExecutionPlan   ExecutionPlan      `json:"execution_plan,omitempty"`
	PhaseRuns       []PhaseRun         `json:"phase_runs,omitempty"`
	LedgerEvents    []LedgerEvent      `json:"ledger_events,omitempty"`
	Artifacts       []ProducedArtifact `json:"artifacts,omitempty"`
	Deliveries      []ReplyDelivery    `json:"deliveries,omitempty"`
	MemoryEvents    []LedgerEvent      `json:"memory_events,omitempty"`
	Completion      EnvelopeCompletion `json:"completion,omitempty"`
	FinalResponse   string             `json:"final_response,omitempty"`
}

type ExecutionPlan struct {
	Planner string           `json:"planner,omitempty"`
	Mode    string           `json:"mode,omitempty"`
	Phases  []map[string]any `json:"phases,omitempty"`
}

type PhaseRun struct {
	PhaseID           string         `json:"phase_id,omitempty"`
	PhaseType         string         `json:"phase_type,omitempty"`
	Status            string         `json:"status,omitempty"`
	InputRefs         []string       `json:"input_refs,omitempty"`
	OutputRefs        []string       `json:"output_refs,omitempty"`
	CompletionVerdict string         `json:"completion_verdict,omitempty"`
	TerminationReason string         `json:"termination_reason,omitempty"`
	Failure           map[string]any `json:"failure,omitempty"`
}

type LedgerEvent struct {
	EventID        string         `json:"event_id,omitempty"`
	Kind           string         `json:"kind,omitempty"`
	PhaseID        string         `json:"phase_id,omitempty"`
	Status         string         `json:"status,omitempty"`
	Sequence       int            `json:"sequence,omitempty"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
	RecordedAt     string         `json:"recorded_at,omitempty"`
	Payload        map[string]any `json:"payload,omitempty"`
}

type EnvelopeCompletion struct {
	CompletionVerdict    string `json:"completion_verdict,omitempty"`
	TerminationReason    string `json:"termination_reason,omitempty"`
	Partial              bool   `json:"partial,omitempty"`
	MaxIterationsReached bool   `json:"max_iterations_reached,omitempty"`
	OK                   bool   `json:"ok,omitempty"`
}

func ParseExecutionEnvelope(resp clients.RunnerResponse) (ExecutionEnvelope, bool, error) {
	raw, ok := resp.Raw["execution_envelope"]
	if !ok || raw == nil {
		return ExecutionEnvelope{}, false, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return ExecutionEnvelope{}, true, fmt.Errorf("marshal runner execution_envelope: %w", err)
	}
	var out ExecutionEnvelope
	if err := json.Unmarshal(data, &out); err != nil {
		return ExecutionEnvelope{}, true, fmt.Errorf("parse runner execution_envelope: %w", err)
	}
	if strings.TrimSpace(out.ContractVersion) == "" {
		return ExecutionEnvelope{}, true, fmt.Errorf("parse runner execution_envelope: contract_version is required")
	}
	return out, true, nil
}

func ParseStructuredOutput(resp clients.RunnerResponse) (StructuredOutput, error) {
	if envelope, ok, err := ParseExecutionEnvelope(resp); ok || err != nil {
		if err != nil {
			return StructuredOutput{}, err
		}
		return StructuredOutputFromEnvelope(envelope, resp)
	}
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

func ExecutionLedgerEventsFromRunnerRaw(raw map[string]any, occurredAt time.Time) []events.ExecutionLedgerEvent {
	envelope, ok, err := ParseExecutionEnvelope(clients.RunnerResponse{Raw: raw})
	if err != nil || !ok {
		return nil
	}
	return ExecutionLedgerEventsFromEnvelope(envelope, occurredAt)
}

func ExecutionLedgerEventsFromEnvelope(envelope ExecutionEnvelope, occurredAt time.Time) []events.ExecutionLedgerEvent {
	executionID := strings.TrimSpace(envelope.ExecutionID)
	if executionID == "" || len(envelope.LedgerEvents) == 0 {
		return nil
	}
	out := make([]events.ExecutionLedgerEvent, 0, len(envelope.LedgerEvents))
	for index, item := range envelope.LedgerEvents {
		seq := item.Sequence
		if seq <= 0 {
			seq = index + 1
		}
		recordedAt := occurredAt
		if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(item.RecordedAt)); err == nil {
			recordedAt = parsed
		}
		payload := item.Payload
		if payload == nil {
			payload = map[string]any{}
		}
		id := strings.TrimSpace(item.EventID)
		if id == "" {
			sum := sha1.Sum([]byte(fmt.Sprintf("%s|%d|%s|%s", executionID, seq, item.Kind, item.PhaseID)))
			id = fmt.Sprintf("xled-%x", sum[:8])
		}
		out = append(out, events.ExecutionLedgerEvent{
			ID:             id,
			ExecutionID:    executionID,
			OperationID:    strings.TrimSpace(envelope.OperationID),
			TraceID:        strings.TrimSpace(envelope.TraceID),
			WorkflowID:     strings.TrimSpace(envelope.WorkflowID),
			PhaseID:        strings.TrimSpace(item.PhaseID),
			Kind:           strings.TrimSpace(item.Kind),
			Status:         strings.TrimSpace(item.Status),
			Seq:            seq,
			IdempotencyKey: strings.TrimSpace(item.IdempotencyKey),
			Payload:        payload,
			RecordedAt:     recordedAt,
		})
	}
	return out
}

func StructuredOutputFromEnvelope(envelope ExecutionEnvelope, resp clients.RunnerResponse) (StructuredOutput, error) {
	var out StructuredOutput
	if raw, ok := resp.Raw["structured_output"]; ok && raw != nil {
		data, err := json.Marshal(raw)
		if err != nil {
			return StructuredOutput{}, fmt.Errorf("marshal runner structured_output projection: %w", err)
		}
		if err := json.Unmarshal(data, &out); err != nil {
			return StructuredOutput{}, fmt.Errorf("parse runner structured_output projection: %w", err)
		}
	}
	if out.FinalAnswer == "" && envelope.FinalResponse != "" {
		out.FinalAnswer = envelope.FinalResponse
	}
	if len(out.ProducedArtifacts) == 0 && len(envelope.Artifacts) > 0 {
		out.ProducedArtifacts = append([]ProducedArtifact(nil), envelope.Artifacts...)
	}
	if out.ReplyDelivery.Status == "" && len(envelope.Deliveries) > 0 {
		out.ReplyDelivery = envelope.Deliveries[0]
	}
	if out.VisibleReasoning == nil {
		out.VisibleReasoning = []Step{}
	}
	if out.ProposedActions == nil {
		out.ProposedActions = []ProposedAction{}
	}
	if out.KnowledgeDrafts == nil {
		out.KnowledgeDrafts = []KnowledgeDraft{}
	}
	if out.OutcomeHypotheses == nil {
		out.OutcomeHypotheses = []OutcomeHypothesis{}
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
	out := make([]events.ReasoningStep, 0, len(output.VisibleReasoning)+1)
	if title := strings.TrimSpace(output.SessionTitle); title != "" {
		out = append(out, events.ReasoningStep{
			ID:         fmt.Sprintf("reason-session-title-%d", createdAt.UnixNano()),
			TraceID:    traceID,
			WorkflowID: workflowID,
			StepType:   "session_title",
			Summary:    title,
			Confidence: output.Confidence,
			Decision:   "set_session_title",
			CreatedAt:  createdAt,
		})
	}
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
