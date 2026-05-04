package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (p *PostgresStore) UpsertCompanyWikiSourceRevision(input CompanyWikiSourceRevisionInput) (CompanyWikiSourceRevisionResult, error) {
	if err := validateCompanyWikiSourceRevisionInput(input); err != nil {
		return CompanyWikiSourceRevisionResult{}, err
	}
	input = normalizeCompanyWikiSourceRevisionInput(input)
	documentID := CompanyWikiStableID("srcdoc", input.SourceType, input.DocumentSourceKey)
	revisionID := CompanyWikiStableID("srcrev", documentID, input.SourceRevision)
	contentSHA := CompanyWikiSHA256(input.Content)
	chunks := companyWikiChunksFromInput(documentID, revisionID, input)
	rawMetadata, err := json.Marshal(nonNilMap(input.Metadata))
	if err != nil {
		return CompanyWikiSourceRevisionResult{}, err
	}
	tx, err := p.db.Begin()
	if err != nil {
		return CompanyWikiSourceRevisionResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC()
	_, err = tx.Exec(`
insert into company_source_document (
  id, source_type, source_key, source_session_key, workspace, environment,
  title, url, status, current_revision_id, metadata, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,'active',$9,$10::jsonb,$11,$11)
on conflict (source_type, source_key) do update
set source_session_key = excluded.source_session_key,
    workspace = excluded.workspace,
    environment = excluded.environment,
    title = excluded.title,
    url = excluded.url,
    status = 'active',
    current_revision_id = excluded.current_revision_id,
    metadata = coalesce(company_source_document.metadata, '{}'::jsonb) || excluded.metadata,
    updated_at = excluded.updated_at`,
		documentID, input.SourceType, input.DocumentSourceKey, input.SourceSessionKey, input.Workspace,
		input.Environment, input.Title, input.URL, revisionID, rawMetadata, now)
	if err != nil {
		return CompanyWikiSourceRevisionResult{}, err
	}

	inserted := false
	if err := tx.QueryRow(`
insert into company_source_revision (
  id, document_id, source_revision, content_sha256, title, url, metadata, observed_at, created_at
) values ($1,$2,$3,$4,$5,$6,$7::jsonb,$8,$9)
on conflict (document_id, source_revision) do nothing
returning true`,
		revisionID, documentID, input.SourceRevision, contentSHA, input.Title, input.URL, rawMetadata, input.ObservedAt, now).Scan(&inserted); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return CompanyWikiSourceRevisionResult{}, err
		}
		inserted = false
	}
	if inserted {
		for _, chunk := range chunks {
			rawChunkMetadata, err := json.Marshal(nonNilMap(chunk.Metadata))
			if err != nil {
				return CompanyWikiSourceRevisionResult{}, err
			}
			if _, err := tx.Exec(`
insert into company_source_chunk (
  id, document_id, revision_id, chunk_index, chunk_kind, content, content_sha256,
  native_locator, token_estimate, metadata, created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11)
on conflict (revision_id, chunk_index) do update
set content = excluded.content,
    content_sha256 = excluded.content_sha256,
    native_locator = excluded.native_locator,
    token_estimate = excluded.token_estimate,
    metadata = excluded.metadata`,
				chunk.ID, chunk.DocumentID, chunk.RevisionID, chunk.ChunkIndex, chunk.ChunkKind,
				chunk.Content, chunk.ContentSHA256, chunk.NativeLocator, chunk.TokenEstimate,
				rawChunkMetadata, now); err != nil {
				return CompanyWikiSourceRevisionResult{}, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return CompanyWikiSourceRevisionResult{}, err
	}
	doc, found, err := p.getCompanyWikiSourceDocument(documentID)
	if err != nil {
		return CompanyWikiSourceRevisionResult{}, err
	}
	if !found {
		return CompanyWikiSourceRevisionResult{}, sql.ErrNoRows
	}
	rev, found, err := p.getCompanyWikiSourceRevision(revisionID)
	if err != nil {
		return CompanyWikiSourceRevisionResult{}, err
	}
	if !found {
		return CompanyWikiSourceRevisionResult{}, sql.ErrNoRows
	}
	if !inserted {
		chunks, err = p.companyWikiChunksForRevision(revisionID)
		if err != nil {
			return CompanyWikiSourceRevisionResult{}, err
		}
	}
	return CompanyWikiSourceRevisionResult{Document: doc, Revision: rev, Chunks: chunks, Inserted: inserted}, nil
}

func (p *PostgresStore) ListCompanyWikiSourceChunks(documentID string) ([]CompanyWikiSourceChunk, error) {
	rows, err := p.db.Query(`
select id, document_id, revision_id, chunk_index, chunk_kind, content, content_sha256,
       native_locator, token_estimate, metadata, created_at
from company_source_chunk
where document_id = $1
order by created_at asc, revision_id asc, chunk_index asc`, strings.TrimSpace(documentID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCompanyWikiChunks(rows)
}

func (p *PostgresStore) BeginCompanyWikiAudit(input CompanyWikiAuditInput) (CompanyWikiAuditRecord, error) {
	input = normalizeCompanyWikiAuditInput(input)
	if input.ID == "" {
		input.ID = CompanyWikiStableID("wikiaudit", input.Mode, input.IdempotencyKey, input.Slug, input.Reason)
	}
	rawMetadata, err := json.Marshal(nonNilMap(input.Metadata))
	if err != nil {
		return CompanyWikiAuditRecord{}, err
	}
	return scanCompanyWikiAudit(p.db.QueryRow(`
insert into company_wiki_write_audit (
  id, mode, status, actor, reason, idempotency_key, page_id, slug, title,
  proposed_path, metadata, created_at, updated_at
) values ($1,$2,'intent',$3,$4,$5,$6,$7,$8,$9,$10::jsonb,now(),now())
on conflict (mode, idempotency_key) where idempotency_key <> '' do update
set updated_at = now()
returning id, mode, status, actor, reason, idempotency_key, page_id, wiki_revision_id,
          slug, title, proposed_path, published_path, metadata, last_error, created_at, updated_at`,
		input.ID, input.Mode, input.Actor, input.Reason, input.IdempotencyKey, input.PageID,
		input.Slug, input.Title, input.ProposedPath, rawMetadata))
}

func (p *PostgresStore) CompleteCompanyWikiAudit(auditID string, wikiRevisionID string, publishedPath string, metadata map[string]any) (CompanyWikiAuditRecord, error) {
	raw, err := json.Marshal(nonNilMap(metadata))
	if err != nil {
		return CompanyWikiAuditRecord{}, err
	}
	return scanCompanyWikiAudit(p.db.QueryRow(`
update company_wiki_write_audit
set status = 'published',
    wiki_revision_id = $2,
    published_path = $3,
    metadata = coalesce(metadata, '{}'::jsonb) || $4::jsonb,
    last_error = '',
    updated_at = now()
where id = $1
returning id, mode, status, actor, reason, idempotency_key, page_id, wiki_revision_id,
          slug, title, proposed_path, published_path, metadata, last_error, created_at, updated_at`,
		strings.TrimSpace(auditID), strings.TrimSpace(wikiRevisionID), strings.TrimSpace(publishedPath), raw))
}

func (p *PostgresStore) FailCompanyWikiAudit(auditID string, lastError string, metadata map[string]any) (CompanyWikiAuditRecord, error) {
	raw, err := json.Marshal(nonNilMap(metadata))
	if err != nil {
		return CompanyWikiAuditRecord{}, err
	}
	return scanCompanyWikiAudit(p.db.QueryRow(`
update company_wiki_write_audit
set status = 'failed',
    metadata = coalesce(metadata, '{}'::jsonb) || $3::jsonb,
    last_error = $2,
    updated_at = now()
where id = $1
returning id, mode, status, actor, reason, idempotency_key, page_id, wiki_revision_id,
          slug, title, proposed_path, published_path, metadata, last_error, created_at, updated_at`,
		strings.TrimSpace(auditID), strings.TrimSpace(lastError), raw))
}

func (p *PostgresStore) PublishCompanyWikiPage(input CompanyWikiPagePublishInput) (CompanyWikiPagePublishResult, error) {
	if err := validateCompanyWikiPublishInput(input); err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	input.Slug = NormalizeCompanyWikiSlug(input.Slug)
	if input.PageID == "" {
		input.PageID = CompanyWikiStableID("wikipage", input.Slug)
	}
	if input.PublishedAt.IsZero() {
		input.PublishedAt = time.Now().UTC()
	}
	revisionID := CompanyWikiStableID("wikirev", input.PageID, input.SHA256)
	sourceRevisionIDsRaw, err := json.Marshal(input.SourceRevisionIDs)
	if err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	rawMetadata, err := json.Marshal(nonNilMap(input.Metadata))
	if err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	tx, err := p.db.Begin()
	if err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`
insert into company_wiki_page (id, slug, title, status, current_revision_id, metadata, created_at, updated_at)
values ($1,$2,$3,'published',$4,$5::jsonb,now(),now())
on conflict (slug) do nothing`,
		input.PageID, input.Slug, input.Title, revisionID, rawMetadata); err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	if _, err := tx.Exec(`select id from company_wiki_page where id = $1 for update`, input.PageID); err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	var revisionNumber int
	if err := tx.QueryRow(`select coalesce(max(revision_number), 0) + 1 from company_wiki_revision where page_id = $1`, input.PageID).Scan(&revisionNumber); err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	if _, err := tx.Exec(`
update company_wiki_page
set title = $2,
    status = 'published',
    current_revision_id = $3,
    metadata = coalesce(metadata, '{}'::jsonb) || $4::jsonb,
    updated_at = now()
where id = $1`,
		input.PageID, input.Title, revisionID, rawMetadata); err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	if _, err := tx.Exec(`
insert into company_wiki_revision (
  id, page_id, revision_number, compiler_run_id, title, body, body_sha256,
  path, source_revision_ids, metadata, published_at, created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10::jsonb,$11,now())
on conflict (page_id, body_sha256) do nothing`,
		revisionID, input.PageID, revisionNumber, input.CompilerRunID, input.Title, input.Body,
		input.SHA256, input.Path, sourceRevisionIDsRaw, rawMetadata, input.PublishedAt); err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	if _, err := tx.Exec(`
update company_wiki_page
set current_revision_id = $2, updated_at = now()
where id = $1`, input.PageID, revisionID); err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	if _, err := tx.Exec(`delete from company_wiki_citation where wiki_revision_id = $1`, revisionID); err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	citations := make([]CompanyWikiCitation, 0, len(input.Citations))
	for _, citation := range input.Citations {
		citationID := CompanyWikiStableID("wikicit", revisionID, citation.SourceDocumentID, citation.SourceRevisionID, citation.ChunkID, citation.ClaimKey)
		row := tx.QueryRow(`
insert into company_wiki_citation (
  id, wiki_revision_id, claim_key, source_document_id, source_revision_id,
  chunk_id, native_locator, quote, created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,now())
returning id, wiki_revision_id, claim_key, source_document_id, source_revision_id,
          chunk_id, native_locator, quote, created_at`,
			citationID, revisionID, citation.ClaimKey, citation.SourceDocumentID,
			citation.SourceRevisionID, citation.ChunkID, citation.NativeLocator, citation.Quote)
		inserted, err := scanCompanyWikiCitation(row)
		if err != nil {
			return CompanyWikiPagePublishResult{}, err
		}
		citations = append(citations, inserted)
	}
	if _, err := tx.Exec(`
insert into company_wiki_manifest (path, wiki_page_id, wiki_revision_id, sha256, compiler_run_id, generated_at)
values ($1,$2,$3,$4,$5,$6)
on conflict (path) do update
set wiki_page_id = excluded.wiki_page_id,
    wiki_revision_id = excluded.wiki_revision_id,
    sha256 = excluded.sha256,
    compiler_run_id = excluded.compiler_run_id,
    generated_at = excluded.generated_at`,
		input.Path, input.PageID, revisionID, input.SHA256, input.CompilerRunID, input.PublishedAt); err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	read, found, err := p.GetCompanyWikiPage(input.Slug)
	if err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	if !found {
		return CompanyWikiPagePublishResult{}, sql.ErrNoRows
	}
	return CompanyWikiPagePublishResult{Page: read.Page, Revision: read.Revision, Citations: citations}, nil
}

func (p *PostgresStore) SearchCompanyWikiPages(query string, limit int) ([]CompanyWikiSearchResult, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	query = escapeLikePattern(query)
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	rows, err := p.db.Query(`
select p.id, p.slug, p.title, r.path, r.id, r.body_sha256,
       left(regexp_replace(r.body, '\s+', ' ', 'g'), 500) as snippet,
       r.published_at
from company_wiki_page p
join company_wiki_revision r on r.id = p.current_revision_id
where $1 = ''
   or lower(p.slug) like '%' || $1 || '%' escape '\'
   or lower(p.title) like '%' || $1 || '%' escape '\'
   or lower(r.body) like '%' || $1 || '%' escape '\'
order by r.published_at desc
limit $2`, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CompanyWikiSearchResult{}
	for rows.Next() {
		var item CompanyWikiSearchResult
		if err := rows.Scan(&item.PageID, &item.Slug, &item.Title, &item.Path, &item.WikiRevisionID, &item.SHA256, &item.Snippet, &item.PublishedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (p *PostgresStore) GetCompanyWikiPage(ref string) (CompanyWikiPageRead, bool, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return CompanyWikiPageRead{}, false, nil
	}
	row := p.db.QueryRow(`
select p.id, p.slug, p.title, p.status, p.current_revision_id, p.metadata, p.created_at, p.updated_at,
       r.id, r.page_id, r.revision_number, r.compiler_run_id, r.title, r.body, r.body_sha256,
       r.path, r.source_revision_ids, r.metadata, r.published_at, r.created_at,
       m.path, m.wiki_page_id, m.wiki_revision_id, m.sha256, m.compiler_run_id, m.generated_at
from company_wiki_page p
join company_wiki_revision r on r.id = p.current_revision_id
left join company_wiki_manifest m on m.wiki_revision_id = r.id
where p.id = $1 or p.slug = $1
limit 1`, ref)
	read, found, err := scanCompanyWikiPageRead(row)
	if err != nil || !found {
		return read, found, err
	}
	citations, err := p.companyWikiCitationsForRevision(read.Revision.ID)
	if err != nil {
		return CompanyWikiPageRead{}, false, err
	}
	read.Citations = citations
	return read, true, nil
}

func (p *PostgresStore) ListCompanyWikiManifestEntries() ([]CompanyWikiManifestEntry, error) {
	rows, err := p.db.Query(`
select path, wiki_page_id, wiki_revision_id, sha256, compiler_run_id, generated_at
from company_wiki_manifest
order by path asc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CompanyWikiManifestEntry{}
	for rows.Next() {
		var item CompanyWikiManifestEntry
		if err := rows.Scan(&item.Path, &item.WikiPageID, &item.WikiRevisionID, &item.SHA256, &item.CompilerRunID, &item.GeneratedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func validateCompanyWikiSourceRevisionInput(input CompanyWikiSourceRevisionInput) error {
	if strings.TrimSpace(input.SourceType) == "" {
		return errors.New("source_type is required")
	}
	if strings.TrimSpace(input.DocumentSourceKey) == "" && strings.TrimSpace(input.SourceKey) == "" {
		return errors.New("document_source_key or source_key is required")
	}
	if strings.TrimSpace(input.SourceRevision) == "" {
		return errors.New("source_revision is required")
	}
	if strings.TrimSpace(input.Content) == "" {
		return errors.New("content is required")
	}
	return nil
}

func normalizeCompanyWikiSourceRevisionInput(input CompanyWikiSourceRevisionInput) CompanyWikiSourceRevisionInput {
	input.SourceType = strings.TrimSpace(input.SourceType)
	input.DocumentSourceKey = strings.TrimSpace(input.DocumentSourceKey)
	input.SourceKey = strings.TrimSpace(input.SourceKey)
	if input.DocumentSourceKey == "" {
		input.DocumentSourceKey = input.SourceKey
	}
	if input.SourceKey == "" {
		input.SourceKey = input.DocumentSourceKey
	}
	input.SourceSessionKey = strings.TrimSpace(input.SourceSessionKey)
	input.Workspace = strings.TrimSpace(input.Workspace)
	input.Environment = strings.TrimSpace(input.Environment)
	input.Title = strings.TrimSpace(input.Title)
	input.URL = strings.TrimSpace(input.URL)
	input.SourceRevision = strings.TrimSpace(input.SourceRevision)
	input.NativeLocator = strings.TrimSpace(input.NativeLocator)
	if input.ObservedAt.IsZero() {
		input.ObservedAt = time.Now().UTC()
	}
	input.Metadata = cloneAnyMap(input.Metadata)
	return input
}

func companyWikiChunksFromInput(documentID string, revisionID string, input CompanyWikiSourceRevisionInput) []CompanyWikiSourceChunk {
	parts := ChunkCompanyWikiText(input.Content, 6000)
	chunks := make([]CompanyWikiSourceChunk, 0, len(parts))
	for idx, content := range parts {
		chunkSHA := CompanyWikiSHA256(content)
		locator := input.NativeLocator
		if len(parts) > 1 {
			locator = fmt.Sprintf("%s#chunk-%d", locator, idx+1)
		}
		chunks = append(chunks, CompanyWikiSourceChunk{
			ID:            CompanyWikiStableID("srcchunk", revisionID, fmt.Sprintf("%d", idx)),
			DocumentID:    documentID,
			RevisionID:    revisionID,
			ChunkIndex:    idx,
			ChunkKind:     "text",
			Content:       content,
			ContentSHA256: chunkSHA,
			NativeLocator: locator,
			TokenEstimate: estimateCompanyWikiTokens(content),
			Metadata:      cloneAnyMap(input.Metadata),
			CreatedAt:     time.Now().UTC(),
		})
	}
	return chunks
}

func estimateCompanyWikiTokens(text string) int {
	runes := len([]rune(text))
	if runes == 0 {
		return 0
	}
	return (runes + 3) / 4
}

func (p *PostgresStore) getCompanyWikiSourceDocument(documentID string) (CompanyWikiSourceDocument, bool, error) {
	return scanCompanyWikiSourceDocument(p.db.QueryRow(`
select id, source_type, source_key, source_session_key, workspace, environment, title, url,
       status, current_revision_id, metadata, created_at, updated_at
from company_source_document
where id = $1`, strings.TrimSpace(documentID)))
}

func (p *PostgresStore) getCompanyWikiSourceRevision(revisionID string) (CompanyWikiSourceRevision, bool, error) {
	return scanCompanyWikiSourceRevision(p.db.QueryRow(`
select id, document_id, source_revision, content_sha256, title, url, metadata, observed_at, created_at
from company_source_revision
where id = $1`, strings.TrimSpace(revisionID)))
}

func (p *PostgresStore) companyWikiChunksForRevision(revisionID string) ([]CompanyWikiSourceChunk, error) {
	rows, err := p.db.Query(`
select id, document_id, revision_id, chunk_index, chunk_kind, content, content_sha256,
       native_locator, token_estimate, metadata, created_at
from company_source_chunk
where revision_id = $1
order by chunk_index asc`, strings.TrimSpace(revisionID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCompanyWikiChunks(rows)
}

func (p *PostgresStore) companyWikiCitationsForRevision(revisionID string) ([]CompanyWikiCitation, error) {
	rows, err := p.db.Query(`
select id, wiki_revision_id, claim_key, source_document_id, source_revision_id,
       chunk_id, native_locator, quote, created_at
from company_wiki_citation
where wiki_revision_id = $1
order by created_at asc`, strings.TrimSpace(revisionID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CompanyWikiCitation{}
	for rows.Next() {
		item, err := scanCompanyWikiCitationScanner(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func validateCompanyWikiPublishInput(input CompanyWikiPagePublishInput) error {
	if strings.TrimSpace(input.Slug) == "" {
		return errors.New("slug is required")
	}
	if strings.TrimSpace(input.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(input.Body) == "" {
		return errors.New("body is required")
	}
	if strings.TrimSpace(input.Path) == "" {
		return errors.New("path is required")
	}
	if strings.TrimSpace(input.SHA256) == "" {
		return errors.New("sha256 is required")
	}
	return ValidateCompanyWikiCitationInputs(input.Citations)
}

func normalizeCompanyWikiAuditInput(input CompanyWikiAuditInput) CompanyWikiAuditInput {
	input.Mode = strings.TrimSpace(input.Mode)
	if input.Mode == "" {
		input.Mode = CompanyWikiAuditModeApply
	}
	input.Actor = strings.TrimSpace(input.Actor)
	input.Reason = strings.TrimSpace(input.Reason)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	input.PageID = strings.TrimSpace(input.PageID)
	input.Slug = NormalizeCompanyWikiSlug(input.Slug)
	input.Title = strings.TrimSpace(input.Title)
	input.ProposedPath = strings.TrimSpace(input.ProposedPath)
	input.Metadata = cloneAnyMap(input.Metadata)
	return input
}

func escapeLikePattern(pattern string) string {
	pattern = strings.ReplaceAll(pattern, "\\", "\\\\")
	pattern = strings.ReplaceAll(pattern, "%", "\\%")
	pattern = strings.ReplaceAll(pattern, "_", "\\_")
	return pattern
}
