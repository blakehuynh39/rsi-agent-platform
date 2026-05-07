package store

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	CompanyWikiSourceStatusActive     = "active"
	CompanyWikiSourceStatusTombstoned = "tombstoned"

	CompanyWikiPageStatusPublished = "published"
	CompanyWikiPageStatusDraft     = "draft"

	CompanyWikiAuditStatusIntent    = "intent"
	CompanyWikiAuditStatusPublished = "published"
	CompanyWikiAuditStatusFailed    = "failed"

	CompanyWikiAuditModeCompiler = "compiler"
	CompanyWikiAuditModePropose  = "propose"
	CompanyWikiAuditModeApply    = "apply"

	CompanyWikiCompileStatusPending   = "pending"
	CompanyWikiCompileStatusClaimed   = "claimed"
	CompanyWikiCompileStatusCompleted = "completed"
	CompanyWikiCompileStatusFailed    = "failed"
	CompanyWikiCompileStatusSkipped   = "skipped"

	CompanyWikiCompileTargetStatusPending    = "pending"
	CompanyWikiCompileTargetStatusPublished  = "published"
	CompanyWikiCompileTargetStatusFailed     = "failed"
	CompanyWikiCompileTargetStatusSkipped    = "skipped"
	CompanyWikiCompileTargetStatusSuperseded = "superseded"

	CompanyWikiManifestRepairOK           = "ok"
	CompanyWikiManifestRepairNeeded       = "repair_needed"
	CompanyWikiManifestRepairFailed       = "repair_failed"
	CompanyWikiManifestRepairNotGenerated = "not_generated"

	CompanyWikiMinimumSchemaVersion int64 = 34
)

var companyWikiSlugUnsafe = regexp.MustCompile(`[^a-z0-9/_-]+`)

type CompanyWikiSourceRevisionInput struct {
	SourceType        string         `json:"source_type"`
	DocumentSourceKey string         `json:"document_source_key"`
	SourceKey         string         `json:"source_key"`
	SourceSessionKey  string         `json:"source_session_key"`
	Workspace         string         `json:"workspace"`
	Environment       string         `json:"environment"`
	Title             string         `json:"title"`
	URL               string         `json:"url"`
	SourceRevision    string         `json:"source_revision"`
	Content           string         `json:"content"`
	NativeLocator     string         `json:"native_locator"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	ObservedAt        time.Time      `json:"observed_at,omitempty"`
}

type CompanyWikiSourceRevisionResult struct {
	Document CompanyWikiSourceDocument `json:"document"`
	Revision CompanyWikiSourceRevision `json:"revision"`
	Chunks   []CompanyWikiSourceChunk  `json:"chunks,omitempty"`
	Inserted bool                      `json:"inserted"`
	Changed  bool                      `json:"changed"`
}

type CompanyWikiSourceDocument struct {
	ID                string         `json:"id"`
	SourceType        string         `json:"source_type"`
	SourceKey         string         `json:"source_key"`
	SourceSessionKey  string         `json:"source_session_key"`
	Workspace         string         `json:"workspace"`
	Environment       string         `json:"environment"`
	Title             string         `json:"title"`
	URL               string         `json:"url"`
	Status            string         `json:"status"`
	CurrentRevisionID string         `json:"current_revision_id,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	CreatedAt         time.Time      `json:"created_at,omitempty"`
	UpdatedAt         time.Time      `json:"updated_at,omitempty"`
}

type CompanyWikiSourceRevision struct {
	ID             string         `json:"id"`
	DocumentID     string         `json:"document_id"`
	SourceRevision string         `json:"source_revision"`
	ContentSHA256  string         `json:"content_sha256"`
	Title          string         `json:"title"`
	URL            string         `json:"url"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	ObservedAt     time.Time      `json:"observed_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at,omitempty"`
}

type CompanyWikiSourceChunk struct {
	ID            string         `json:"id"`
	DocumentID    string         `json:"document_id"`
	RevisionID    string         `json:"revision_id"`
	ChunkIndex    int            `json:"chunk_index"`
	ChunkKind     string         `json:"chunk_kind"`
	Content       string         `json:"content"`
	ContentSHA256 string         `json:"content_sha256"`
	NativeLocator string         `json:"native_locator"`
	TokenEstimate int            `json:"token_estimate"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedAt     time.Time      `json:"created_at,omitempty"`
}

type CompanyWikiCitationInput struct {
	ClaimKey         string `json:"claim_key,omitempty"`
	SourceDocumentID string `json:"source_document_id"`
	SourceRevisionID string `json:"source_revision_id"`
	ChunkID          string `json:"chunk_id"`
	NativeLocator    string `json:"native_locator,omitempty"`
	Quote            string `json:"quote,omitempty"`
}

type CompanyWikiCitation struct {
	ID               string    `json:"id"`
	WikiRevisionID   string    `json:"wiki_revision_id"`
	ClaimKey         string    `json:"claim_key,omitempty"`
	SourceDocumentID string    `json:"source_document_id"`
	SourceRevisionID string    `json:"source_revision_id"`
	ChunkID          string    `json:"chunk_id"`
	NativeLocator    string    `json:"native_locator,omitempty"`
	Quote            string    `json:"quote,omitempty"`
	CreatedAt        time.Time `json:"created_at,omitempty"`
}

type CompanyWikiClaimInput struct {
	ClaimKey   string         `json:"claim_key"`
	ClaimText  string         `json:"claim_text"`
	Confidence float64        `json:"confidence,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type CompanyWikiClaim struct {
	ID             string         `json:"id"`
	WikiRevisionID string         `json:"wiki_revision_id"`
	ClaimKey       string         `json:"claim_key"`
	ClaimText      string         `json:"claim_text"`
	Confidence     float64        `json:"confidence"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	CreatedAt      time.Time      `json:"created_at,omitempty"`
}

type CompanyWikiConflictInput struct {
	ClaimKey  string         `json:"claim_key"`
	Summary   string         `json:"summary"`
	Citations []string       `json:"citation_ids,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type CompanyWikiConflict struct {
	ID             string         `json:"id"`
	WikiRevisionID string         `json:"wiki_revision_id"`
	ClaimKey       string         `json:"claim_key"`
	Summary        string         `json:"summary"`
	CitationIDs    []string       `json:"citation_ids,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	CreatedAt      time.Time      `json:"created_at,omitempty"`
}

type CompanyWikiAuditInput struct {
	ID             string         `json:"id,omitempty"`
	Mode           string         `json:"mode"`
	Actor          string         `json:"actor"`
	Reason         string         `json:"reason"`
	IdempotencyKey string         `json:"idempotency_key"`
	PageID         string         `json:"page_id,omitempty"`
	Slug           string         `json:"slug,omitempty"`
	Title          string         `json:"title,omitempty"`
	ProposedPath   string         `json:"proposed_path,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

type CompanyWikiAuditRecord struct {
	ID             string         `json:"id"`
	Mode           string         `json:"mode"`
	Status         string         `json:"status"`
	Actor          string         `json:"actor"`
	Reason         string         `json:"reason"`
	IdempotencyKey string         `json:"idempotency_key"`
	PageID         string         `json:"page_id,omitempty"`
	WikiRevisionID string         `json:"wiki_revision_id,omitempty"`
	Slug           string         `json:"slug,omitempty"`
	Title          string         `json:"title,omitempty"`
	ProposedPath   string         `json:"proposed_path,omitempty"`
	PublishedPath  string         `json:"published_path,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	LastError      string         `json:"last_error,omitempty"`
	CreatedAt      time.Time      `json:"created_at,omitempty"`
	UpdatedAt      time.Time      `json:"updated_at,omitempty"`
}

type CompanyWikiPagePublishInput struct {
	AuditID           string                     `json:"audit_id,omitempty"`
	PageID            string                     `json:"page_id,omitempty"`
	Slug              string                     `json:"slug"`
	Title             string                     `json:"title"`
	Body              string                     `json:"body"`
	Path              string                     `json:"path"`
	SHA256            string                     `json:"sha256"`
	CompilerRunID     string                     `json:"compiler_run_id,omitempty"`
	SourceRevisionIDs []string                   `json:"source_revision_ids,omitempty"`
	Citations         []CompanyWikiCitationInput `json:"citations,omitempty"`
	Claims            []CompanyWikiClaimInput    `json:"claims,omitempty"`
	Conflicts         []CompanyWikiConflictInput `json:"conflicts,omitempty"`
	Metadata          map[string]any             `json:"metadata,omitempty"`
	PublishedAt       time.Time                  `json:"published_at,omitempty"`
}

type CompanyWikiPagePublishResult struct {
	Page      CompanyWikiPage       `json:"page"`
	Revision  CompanyWikiRevision   `json:"revision"`
	Citations []CompanyWikiCitation `json:"citations,omitempty"`
	Claims    []CompanyWikiClaim    `json:"claims,omitempty"`
	Conflicts []CompanyWikiConflict `json:"conflicts,omitempty"`
}

type CompanyWikiPage struct {
	ID                string         `json:"id"`
	Slug              string         `json:"slug"`
	Title             string         `json:"title"`
	Status            string         `json:"status"`
	CurrentRevisionID string         `json:"current_revision_id,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	CreatedAt         time.Time      `json:"created_at,omitempty"`
	UpdatedAt         time.Time      `json:"updated_at,omitempty"`
}

type CompanyWikiRevision struct {
	ID                string         `json:"id"`
	PageID            string         `json:"page_id"`
	RevisionNumber    int            `json:"revision_number"`
	CompilerRunID     string         `json:"compiler_run_id,omitempty"`
	Title             string         `json:"title"`
	Body              string         `json:"body,omitempty"`
	BodySHA256        string         `json:"body_sha256"`
	Path              string         `json:"path"`
	SourceRevisionIDs []string       `json:"source_revision_ids,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	PublishedAt       time.Time      `json:"published_at,omitempty"`
	CreatedAt         time.Time      `json:"created_at,omitempty"`
}

type CompanyWikiManifestEntry struct {
	Path            string    `json:"path"`
	WikiPageID      string    `json:"wiki_page_id"`
	WikiRevisionID  string    `json:"wiki_revision_id"`
	SHA256          string    `json:"sha256"`
	CompilerRunID   string    `json:"compiler_run_id,omitempty"`
	GeneratedAt     time.Time `json:"generated_at,omitempty"`
	RepairStatus    string    `json:"repair_status,omitempty"`
	LastRepairError string    `json:"last_repair_error,omitempty"`
	LastCheckedAt   time.Time `json:"last_checked_at,omitempty"`
	LastRepairedAt  time.Time `json:"last_repaired_at,omitempty"`
}

type CompanyWikiSearchResult struct {
	PageID         string    `json:"page_id"`
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Path           string    `json:"path"`
	WikiRevisionID string    `json:"wiki_revision_id"`
	SHA256         string    `json:"sha256"`
	Snippet        string    `json:"snippet,omitempty"`
	Freshness      string    `json:"freshness,omitempty"`
	PublishedAt    time.Time `json:"published_at,omitempty"`
}

type CompanyWikiPageRead struct {
	Page      CompanyWikiPage          `json:"page"`
	Revision  CompanyWikiRevision      `json:"revision"`
	Citations []CompanyWikiCitation    `json:"citations,omitempty"`
	Claims    []CompanyWikiClaim       `json:"claims,omitempty"`
	Conflicts []CompanyWikiConflict    `json:"conflicts,omitempty"`
	Manifest  CompanyWikiManifestEntry `json:"manifest"`
}

type CompanyWikiSourceEvidence struct {
	Document CompanyWikiSourceDocument `json:"document"`
	Revision CompanyWikiSourceRevision `json:"revision"`
	Chunks   []CompanyWikiSourceChunk  `json:"chunks,omitempty"`
}

type CompanyWikiCompileItemInput struct {
	SourceRevisionID   string    `json:"source_revision_id"`
	CompilerVersion    string    `json:"compiler_version"`
	SchemaVersion      string    `json:"schema_version"`
	RendererVersion    string    `json:"renderer_version"`
	ModelPolicyVersion string    `json:"model_policy_version"`
	InputHash          string    `json:"input_hash"`
	Status             string    `json:"status,omitempty"`
	CreatedAt          time.Time `json:"created_at,omitempty"`
}

type CompanyWikiCompileItem struct {
	ID                 string    `json:"id"`
	SourceRevisionID   string    `json:"source_revision_id"`
	CompilerVersion    string    `json:"compiler_version"`
	SchemaVersion      string    `json:"schema_version"`
	RendererVersion    string    `json:"renderer_version"`
	ModelPolicyVersion string    `json:"model_policy_version"`
	InputHash          string    `json:"input_hash"`
	Status             string    `json:"status"`
	LeaseHolder        string    `json:"lease_holder,omitempty"`
	LeaseExpiresAt     time.Time `json:"lease_expires_at,omitempty"`
	AttemptCount       int       `json:"attempt_count"`
	LastAttemptID      string    `json:"last_attempt_id,omitempty"`
	LastError          string    `json:"last_error,omitempty"`
	CreatedAt          time.Time `json:"created_at,omitempty"`
	UpdatedAt          time.Time `json:"updated_at,omitempty"`
}

type CompanyWikiCompileClaimInput struct {
	Limit              int
	LeaseHolder        string
	LeaseDuration      time.Duration
	CompilerVersion    string
	SchemaVersion      string
	RendererVersion    string
	ModelPolicyVersion string
	MaxAttempts        int
}

type CompanyWikiCompileAttemptInput struct {
	CompileItemID        string         `json:"compile_item_id"`
	CompilerRunID        string         `json:"compiler_run_id"`
	Status               string         `json:"status,omitempty"`
	Model                string         `json:"model"`
	ContextHash          string         `json:"context_hash"`
	OutputHash           string         `json:"output_hash,omitempty"`
	RequestMetadataHash  string         `json:"request_metadata_hash,omitempty"`
	ResponseMetadataHash string         `json:"response_metadata_hash,omitempty"`
	DurationMillis       int64          `json:"duration_millis,omitempty"`
	ValidationErrors     []string       `json:"validation_errors,omitempty"`
	LastError            string         `json:"last_error,omitempty"`
	Metadata             map[string]any `json:"metadata,omitempty"`
}

type CompanyWikiCompileAttempt struct {
	ID                   string         `json:"id"`
	CompileItemID        string         `json:"compile_item_id"`
	CompilerRunID        string         `json:"compiler_run_id"`
	Status               string         `json:"status"`
	Model                string         `json:"model"`
	ContextHash          string         `json:"context_hash"`
	OutputHash           string         `json:"output_hash,omitempty"`
	RequestMetadataHash  string         `json:"request_metadata_hash,omitempty"`
	ResponseMetadataHash string         `json:"response_metadata_hash,omitempty"`
	DurationMillis       int64          `json:"duration_millis,omitempty"`
	ValidationErrors     []string       `json:"validation_errors,omitempty"`
	LastError            string         `json:"last_error,omitempty"`
	Metadata             map[string]any `json:"metadata,omitempty"`
	CreatedAt            time.Time      `json:"created_at,omitempty"`
	CompletedAt          time.Time      `json:"completed_at,omitempty"`
}

type CompanyWikiCompileTargetInput struct {
	CompileItemID  string `json:"compile_item_id"`
	TargetSlug     string `json:"target_slug"`
	TargetPath     string `json:"target_path"`
	TargetType     string `json:"target_type"`
	Status         string `json:"status,omitempty"`
	WikiRevisionID string `json:"wiki_revision_id,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	BodyHash       string `json:"body_hash,omitempty"`
	LastError      string `json:"last_error,omitempty"`
}

type CompanyWikiCompileTarget struct {
	ID             string    `json:"id"`
	CompileItemID  string    `json:"compile_item_id"`
	TargetSlug     string    `json:"target_slug"`
	TargetPath     string    `json:"target_path"`
	TargetType     string    `json:"target_type"`
	Status         string    `json:"status"`
	WikiRevisionID string    `json:"wiki_revision_id,omitempty"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
	BodyHash       string    `json:"body_hash,omitempty"`
	LastError      string    `json:"last_error,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

type CompanyWikiPageQuery struct {
	Query             string
	Limit             int
	Types             []string
	Tags              []string
	SourceRevisionIDs []string
	ExcludeEvidence   bool
}

type CompanyWikiStore interface {
	UpsertCompanyWikiSourceRevision(input CompanyWikiSourceRevisionInput) (CompanyWikiSourceRevisionResult, error)
	ListCompanyWikiSourceChunks(documentID string) ([]CompanyWikiSourceChunk, error)
	GetCompanyWikiSourceEvidence(sourceRevisionID string) (CompanyWikiSourceEvidence, bool, error)
	ListCompanyWikiSourceRevisionIDsWithoutCompileItem(compilerVersion string, schemaVersion string, rendererVersion string, modelPolicyVersion string, limit int) ([]string, error)
	EnqueueCompanyWikiCompileItem(input CompanyWikiCompileItemInput) (CompanyWikiCompileItem, bool, error)
	ClaimCompanyWikiCompileItems(input CompanyWikiCompileClaimInput) ([]CompanyWikiCompileItem, error)
	BeginCompanyWikiCompileAttempt(input CompanyWikiCompileAttemptInput) (CompanyWikiCompileAttempt, error)
	CompleteCompanyWikiCompileAttempt(attemptID string, status string, outputHash string, durationMillis int64, validationErrors []string, lastError string, metadata map[string]any) (CompanyWikiCompileAttempt, error)
	UpsertCompanyWikiCompileTargets(compileItemID string, targets []CompanyWikiCompileTargetInput) ([]CompanyWikiCompileTarget, error)
	UpdateCompanyWikiCompileTarget(input CompanyWikiCompileTargetInput) (CompanyWikiCompileTarget, error)
	ListCompanyWikiCompileTargets(compileItemID string) ([]CompanyWikiCompileTarget, error)
	CompleteCompanyWikiCompileItem(compileItemID string, status string, lastError string) (CompanyWikiCompileItem, error)
	ListCompanyWikiCandidatePages(query CompanyWikiPageQuery) ([]CompanyWikiPageRead, error)
	UpdateCompanyWikiManifestRepair(path string, status string, lastError string) error
	BeginCompanyWikiAudit(input CompanyWikiAuditInput) (CompanyWikiAuditRecord, error)
	CompleteCompanyWikiAudit(auditID string, wikiRevisionID string, publishedPath string, metadata map[string]any) (CompanyWikiAuditRecord, error)
	FailCompanyWikiAudit(auditID string, lastError string, metadata map[string]any) (CompanyWikiAuditRecord, error)
	PublishCompanyWikiPage(input CompanyWikiPagePublishInput) (CompanyWikiPagePublishResult, error)
	SearchCompanyWikiPages(query string, limit int) ([]CompanyWikiSearchResult, error)
	GetCompanyWikiPage(ref string) (CompanyWikiPageRead, bool, error)
	ListCompanyWikiManifestEntries() ([]CompanyWikiManifestEntry, error)
}

func CompanyWikiStableID(prefix string, parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return strings.TrimSpace(prefix) + "_" + hex.EncodeToString(sum[:])[:32]
}

func CompanyWikiSHA256(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func NormalizeCompanyWikiSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "\\", "/")
	value = strings.Trim(value, "/")
	value = companyWikiSlugUnsafe.ReplaceAllString(value, "-")
	value = strings.ReplaceAll(value, "//", "/")
	value = strings.Trim(value, "-_/")
	if value == "" {
		return "untitled"
	}
	clean := filepath.Clean(value)
	clean = strings.TrimPrefix(clean, "../")
	clean = strings.Trim(clean, "/.")
	if clean == "" || clean == "." {
		return "untitled"
	}
	return clean
}

func ValidateCompanyWikiCitationInputs(citations []CompanyWikiCitationInput) error {
	if len(citations) == 0 {
		return errors.New("at least one citation is required")
	}
	for _, citation := range citations {
		if strings.TrimSpace(citation.SourceDocumentID) == "" {
			return errors.New("citation.source_document_id is required")
		}
		if strings.TrimSpace(citation.SourceRevisionID) == "" {
			return errors.New("citation.source_revision_id is required")
		}
		if strings.TrimSpace(citation.ChunkID) == "" {
			return errors.New("citation.chunk_id is required")
		}
	}
	return nil
}

func normalizeCompanyWikiCompileItemInput(input CompanyWikiCompileItemInput) CompanyWikiCompileItemInput {
	input.SourceRevisionID = strings.TrimSpace(input.SourceRevisionID)
	input.CompilerVersion = strings.TrimSpace(input.CompilerVersion)
	input.SchemaVersion = strings.TrimSpace(input.SchemaVersion)
	input.RendererVersion = strings.TrimSpace(input.RendererVersion)
	input.ModelPolicyVersion = strings.TrimSpace(input.ModelPolicyVersion)
	input.InputHash = strings.TrimSpace(input.InputHash)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = CompanyWikiCompileStatusPending
	}
	return input
}

func normalizeCompanyWikiCompileClaimInput(input CompanyWikiCompileClaimInput) CompanyWikiCompileClaimInput {
	input.LeaseHolder = strings.TrimSpace(input.LeaseHolder)
	if input.LeaseHolder == "" {
		input.LeaseHolder = "company_wiki_compiler"
	}
	if input.LeaseDuration <= 0 {
		input.LeaseDuration = 5 * time.Minute
	}
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 10
	}
	input.CompilerVersion = strings.TrimSpace(input.CompilerVersion)
	input.SchemaVersion = strings.TrimSpace(input.SchemaVersion)
	input.RendererVersion = strings.TrimSpace(input.RendererVersion)
	input.ModelPolicyVersion = strings.TrimSpace(input.ModelPolicyVersion)
	return input
}

func normalizeCompanyWikiCompileAttemptInput(input CompanyWikiCompileAttemptInput) CompanyWikiCompileAttemptInput {
	input.CompileItemID = strings.TrimSpace(input.CompileItemID)
	input.CompilerRunID = strings.TrimSpace(input.CompilerRunID)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = CompanyWikiCompileStatusClaimed
	}
	input.Model = strings.TrimSpace(input.Model)
	input.ContextHash = strings.TrimSpace(input.ContextHash)
	input.OutputHash = strings.TrimSpace(input.OutputHash)
	input.RequestMetadataHash = strings.TrimSpace(input.RequestMetadataHash)
	input.ResponseMetadataHash = strings.TrimSpace(input.ResponseMetadataHash)
	input.LastError = strings.TrimSpace(input.LastError)
	input.Metadata = cloneAnyMap(input.Metadata)
	return input
}

func normalizeCompanyWikiCompileTargetInput(input CompanyWikiCompileTargetInput) CompanyWikiCompileTargetInput {
	input.CompileItemID = strings.TrimSpace(input.CompileItemID)
	input.TargetSlug = NormalizeCompanyWikiSlug(input.TargetSlug)
	input.TargetPath = strings.TrimSpace(input.TargetPath)
	input.TargetType = strings.TrimSpace(input.TargetType)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = CompanyWikiCompileTargetStatusPending
	}
	input.WikiRevisionID = strings.TrimSpace(input.WikiRevisionID)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	input.BodyHash = strings.TrimSpace(input.BodyHash)
	input.LastError = strings.TrimSpace(input.LastError)
	return input
}

func companyWikiCompileItemKey(sourceRevisionID string, compilerVersion string, schemaVersion string, rendererVersion string, modelPolicyVersion string) string {
	return strings.Join([]string{
		strings.TrimSpace(sourceRevisionID),
		strings.TrimSpace(compilerVersion),
		strings.TrimSpace(schemaVersion),
		strings.TrimSpace(rendererVersion),
		strings.TrimSpace(modelPolicyVersion),
	}, "\x00")
}

func cloneCompanyWikiChunks(input []CompanyWikiSourceChunk) []CompanyWikiSourceChunk {
	out := make([]CompanyWikiSourceChunk, len(input))
	copy(out, input)
	for i := range out {
		out[i].Metadata = cloneAnyMap(out[i].Metadata)
	}
	return out
}

func sortCompanyWikiChunks(chunks []CompanyWikiSourceChunk) {
	sort.SliceStable(chunks, func(i, j int) bool {
		if chunks[i].RevisionID == chunks[j].RevisionID {
			return chunks[i].ChunkIndex < chunks[j].ChunkIndex
		}
		if chunks[i].CreatedAt.Equal(chunks[j].CreatedAt) {
			return chunks[i].RevisionID < chunks[j].RevisionID
		}
		return chunks[i].CreatedAt.Before(chunks[j].CreatedAt)
	})
}

func ChunkCompanyWikiText(text string, maxRunes int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if maxRunes <= 0 {
		maxRunes = 6000
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return []string{text}
	}
	out := []string{}
	for start := 0; start < len(runes); {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		if end < len(runes) {
			for i := end; i > start+maxRunes/2; i-- {
				if runes[i-1] == '\n' || runes[i-1] == '.' {
					end = i
					break
				}
			}
		}
		out = append(out, strings.TrimSpace(string(runes[start:end])))
		start = end
	}
	return out
}

func companyWikiFreshness(metadata map[string]any, body string) string {
	for _, key := range []string{"freshness", "source_freshness", "source_last_edited_at"} {
		value := stringFromAnyMap(metadata, key)
		if value != "" && value != "unknown" {
			return value
		}
	}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "---" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.TrimSpace(key) != "freshness" {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if value != "unknown" {
			return value
		}
		return ""
	}
	return ""
}
