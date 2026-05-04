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

	CompanyWikiMinimumSchemaVersion int64 = 33
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
	Metadata          map[string]any             `json:"metadata,omitempty"`
	PublishedAt       time.Time                  `json:"published_at,omitempty"`
}

type CompanyWikiPagePublishResult struct {
	Page      CompanyWikiPage       `json:"page"`
	Revision  CompanyWikiRevision   `json:"revision"`
	Citations []CompanyWikiCitation `json:"citations,omitempty"`
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
	Path           string    `json:"path"`
	WikiPageID     string    `json:"wiki_page_id"`
	WikiRevisionID string    `json:"wiki_revision_id"`
	SHA256         string    `json:"sha256"`
	CompilerRunID  string    `json:"compiler_run_id,omitempty"`
	GeneratedAt    time.Time `json:"generated_at,omitempty"`
}

type CompanyWikiSearchResult struct {
	PageID         string    `json:"page_id"`
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Path           string    `json:"path"`
	WikiRevisionID string    `json:"wiki_revision_id"`
	SHA256         string    `json:"sha256"`
	Snippet        string    `json:"snippet,omitempty"`
	PublishedAt    time.Time `json:"published_at,omitempty"`
}

type CompanyWikiPageRead struct {
	Page      CompanyWikiPage          `json:"page"`
	Revision  CompanyWikiRevision      `json:"revision"`
	Citations []CompanyWikiCitation    `json:"citations,omitempty"`
	Manifest  CompanyWikiManifestEntry `json:"manifest"`
}

type CompanyWikiStore interface {
	UpsertCompanyWikiSourceRevision(input CompanyWikiSourceRevisionInput) (CompanyWikiSourceRevisionResult, error)
	ListCompanyWikiSourceChunks(documentID string) ([]CompanyWikiSourceChunk, error)
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
