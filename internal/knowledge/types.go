package knowledge

import (
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/events"
)

type Tier string

const (
	TierWorking   Tier = "working"
	TierCanonical Tier = "canonical"
)

type Kind string

const (
	KindFact                Kind = "fact"
	KindPlaybook            Kind = "playbook"
	KindArchitecturePattern Kind = "architecture_pattern"
	KindRepoNote            Kind = "repo_note"
	KindIncidentRunbook     Kind = "incident_runbook"
)

type ScopeType string

const (
	ScopeGlobal       ScopeType = "global"
	ScopeRepo         ScopeType = "repo"
	ScopeService      ScopeType = "service"
	ScopeConversation ScopeType = "conversation"
	ScopeCase         ScopeType = "case"
)

type Status string

const (
	StatusDraft         Status = "draft"
	StatusReviewPending Status = "review_pending"
	StatusCanonical     Status = "canonical"
	StatusStale         Status = "stale"
	StatusContradicted  Status = "contradicted"
	StatusArchived      Status = "archived"
)

type SourceType string

const (
	SourceAgent    SourceType = "agent"
	SourceHuman    SourceType = "human"
	SourcePromoted SourceType = "promoted"
)

type Entry struct {
	ID                    string         `json:"id"`
	Tier                  Tier           `json:"tier"`
	Kind                  Kind           `json:"kind"`
	ScopeType             ScopeType      `json:"scope_type"`
	ScopeID               string         `json:"scope_id,omitempty"`
	Title                 string         `json:"title"`
	Summary               string         `json:"summary,omitempty"`
	Body                  string         `json:"body,omitempty"`
	StructuredFacts       map[string]any `json:"structured_facts,omitempty"`
	Status                Status         `json:"status"`
	Confidence            float64        `json:"confidence,omitempty"`
	FreshUntil            *time.Time     `json:"fresh_until,omitempty"`
	SourceType            SourceType     `json:"source_type"`
	SupersedesEntryID     string         `json:"supersedes_entry_id,omitempty"`
	ContradictedByEntryID string         `json:"contradicted_by_entry_id,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

type EvidenceLink struct {
	KnowledgeEntryID string             `json:"knowledge_entry_id"`
	EvidenceType     string             `json:"evidence_type"`
	EvidenceID       string             `json:"evidence_id"`
	RelevanceSummary string             `json:"relevance_summary,omitempty"`
	EvidenceRef      events.EvidenceRef `json:"evidence_ref"`
}

type Review struct {
	ID               string    `json:"id"`
	KnowledgeEntryID string    `json:"knowledge_entry_id"`
	Decision         string    `json:"decision"`
	ReviewerID       string    `json:"reviewer_id"`
	Rationale        string    `json:"rationale,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type Draft struct {
	Kind         Kind                 `json:"kind"`
	ScopeType    ScopeType            `json:"scope_type"`
	ScopeID      string               `json:"scope_id,omitempty"`
	Title        string               `json:"title"`
	Summary      string               `json:"summary,omitempty"`
	Body         string               `json:"body,omitempty"`
	Confidence   float64              `json:"confidence,omitempty"`
	FreshUntil   string               `json:"fresh_until,omitempty"`
	EvidenceRefs []events.EvidenceRef `json:"evidence_refs,omitempty"`
}
