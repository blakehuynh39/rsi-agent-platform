package questionrun

type ReplyTarget struct {
	ChannelID string `json:"channel_id,omitempty"`
	ThreadTS  string `json:"thread_ts,omitempty"`
}

type SlackSurface struct {
	ChannelID string `json:"channel_id,omitempty"`
	ThreadTS  string `json:"thread_ts,omitempty"`
	Ref       string `json:"ref,omitempty"`
	Source    string `json:"source,omitempty"`
}

type PromptEntity struct {
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
}

type PromptEnvelope struct {
	ChannelID         string         `json:"channel_id,omitempty"`
	ChannelName       string         `json:"channel_name,omitempty"`
	ThreadTS          string         `json:"thread_ts,omitempty"`
	SenderUserID      string         `json:"sender_user_id,omitempty"`
	SenderDisplayName string         `json:"sender_display_name,omitempty"`
	RawText           string         `json:"raw_text,omitempty"`
	RenderedText      string         `json:"rendered_text,omitempty"`
	MentionedChannels []PromptEntity `json:"mentioned_channels,omitempty"`
	MentionedUsers    []PromptEntity `json:"mentioned_users,omitempty"`
	Permalink         string         `json:"permalink,omitempty"`
}

type InvestigationSpec struct {
	UserRequest       string         `json:"user_request"`
	ReplyTarget       ReplyTarget    `json:"reply_target"`
	Prompt            PromptEnvelope `json:"prompt_envelope,omitempty"`
	Repo              string         `json:"repo,omitempty"`
	ProjectKey        string         `json:"project_key,omitempty"`
	Since             string         `json:"since,omitempty"`
	Until             string         `json:"until,omitempty"`
	ReadSurfaces      []SlackSurface `json:"read_surfaces,omitempty"`
	AlignmentRequired bool           `json:"alignment_required,omitempty"`
	RetrievalBudget   int            `json:"retrieval_budget,omitempty"`
	AllowExpansion    bool           `json:"allow_expansion,omitempty"`
	WorkflowStrategy  string         `json:"workflow_strategy,omitempty"`
	GatherTaskType    string         `json:"gather_task_type,omitempty"`
	ReduceTaskType    string         `json:"reduce_task_type,omitempty"`
	ReductionTaskType string         `json:"reduction_task_type,omitempty"`
	ExpansionTaskType string         `json:"expansion_task_type,omitempty"`
}

type ToolCall struct {
	ToolName        string         `json:"tool_name"`
	ToolCallID      string         `json:"tool_call_id"`
	Request         map[string]any `json:"request,omitempty"`
	Summary         string         `json:"summary,omitempty"`
	Status          string         `json:"status,omitempty"`
	ProviderRef     string         `json:"provider_ref,omitempty"`
	RawArtifactRefs []string       `json:"raw_artifact_refs,omitempty"`
	StartedAt       string         `json:"started_at,omitempty"`
	CompletedAt     string         `json:"completed_at,omitempty"`
}

type EvidenceItem struct {
	Kind          string  `json:"kind"`
	Summary       string  `json:"summary"`
	FactOrSnippet string  `json:"fact_or_snippet,omitempty"`
	Snippet       string  `json:"snippet,omitempty"`
	SourceRef     string  `json:"source_ref,omitempty"`
	ToolName      string  `json:"tool_name,omitempty"`
	ChannelID     string  `json:"channel_id,omitempty"`
	ThreadTS      string  `json:"thread_ts,omitempty"`
	MessageTS     string  `json:"message_ts,omitempty"`
	Path          string  `json:"path,omitempty"`
	Repo          string  `json:"repo,omitempty"`
	Commit        string  `json:"commit,omitempty"`
	Permalink     string  `json:"permalink,omitempty"`
	Author        string  `json:"author,omitempty"`
	CapturedAt    string  `json:"captured_at,omitempty"`
	Score         float64 `json:"score,omitempty"`
}

type ProjectAlignmentLedger struct {
	ProjectKey         string         `json:"project_key,omitempty"`
	Summary            string         `json:"summary,omitempty"`
	RequiredOutcomes   []string       `json:"required_outcomes,omitempty"`
	Constraints        []string       `json:"constraints,omitempty"`
	OpenQuestions      []string       `json:"open_questions,omitempty"`
	Sources            []string       `json:"sources,omitempty"`
	EvidenceItems      []EvidenceItem `json:"evidence_items,omitempty"`
	Degraded           bool           `json:"degraded,omitempty"`
	DegradedReason     string         `json:"degraded_reason,omitempty"`
	KnowledgeEntryID   string         `json:"knowledge_entry_id,omitempty"`
	RefreshedAtRFC3339 string         `json:"refreshed_at,omitempty"`
}

type EvidenceLedger struct {
	InvestigationSpec    *InvestigationSpec      `json:"investigation_spec,omitempty"`
	UserRequest          string                  `json:"user_request"`
	ReplyTarget          ReplyTarget             `json:"reply_target"`
	Prompt               PromptEnvelope          `json:"prompt_envelope,omitempty"`
	Repo                 string                  `json:"repo,omitempty"`
	ProjectKey           string                  `json:"project_key,omitempty"`
	Since                string                  `json:"since,omitempty"`
	Until                string                  `json:"until,omitempty"`
	AlignmentRequired    bool                    `json:"alignment_required,omitempty"`
	AlignmentDegraded    bool                    `json:"alignment_degraded,omitempty"`
	AlignmentLedger      *ProjectAlignmentLedger `json:"alignment_ledger,omitempty"`
	ToolCalls            []ToolCall              `json:"tool_calls,omitempty"`
	EvidenceItems        []EvidenceItem          `json:"evidence_items,omitempty"`
	OpenQuestions        []string                `json:"open_questions,omitempty"`
	MissingEvidence      []string                `json:"missing_evidence,omitempty"`
	DraftReplyCandidates []string                `json:"draft_reply_candidates,omitempty"`
	TerminationReason    string                  `json:"termination_reason,omitempty"`
}

type EvidenceDelta struct {
	ToolCalls            []ToolCall     `json:"tool_calls,omitempty"`
	EvidenceItems        []EvidenceItem `json:"evidence_items,omitempty"`
	OpenQuestions        []string       `json:"open_questions,omitempty"`
	DraftReplyCandidates []string       `json:"draft_reply_candidates,omitempty"`
	InsufficiencyMarks   []string       `json:"insufficiency_markers,omitempty"`
	Confidence           float64        `json:"confidence,omitempty"`
}

type Result struct {
	ReplyMarkdown     string  `json:"reply_markdown"`
	Confidence        float64 `json:"confidence,omitempty"`
	CompletionVerdict string  `json:"completion_verdict,omitempty"`
	TerminationReason string  `json:"termination_reason,omitempty"`
	AlignmentDegraded bool    `json:"alignment_degraded,omitempty"`
	AlignmentNotice   string  `json:"alignment_notice,omitempty"`
}
