package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (p *PostgresStore) AcquireCompanyWikiCompilerLease(ctx context.Context, lockName string, holder string, ttl time.Duration) (func() error, bool, error) {
	_, _ = holder, ttl
	lockName = strings.TrimSpace(lockName)
	if lockName == "" {
		lockName = "company_wiki_compiler"
	}
	conn, err := p.db.Conn(ctx)
	if err != nil {
		return nil, false, err
	}
	var acquired bool
	if err := conn.QueryRowContext(ctx, `select pg_try_advisory_lock(hashtextextended($1, 0))`, lockName).Scan(&acquired); err != nil {
		_ = conn.Close()
		return nil, false, err
	}
	if !acquired {
		_ = conn.Close()
		return nil, false, nil
	}
	release := func() error {
		_, unlockErr := conn.ExecContext(context.Background(), `select pg_advisory_unlock(hashtextextended($1, 0))`, lockName)
		closeErr := conn.Close()
		if unlockErr != nil {
			return unlockErr
		}
		return closeErr
	}
	return release, true, nil
}

func (p *PostgresStore) BeginCompanyWikiCompilerRun(id string, metadata map[string]any) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("compiler run id is required")
	}
	rawMetadata, err := json.Marshal(nonNilMap(metadata))
	if err != nil {
		return err
	}
	_, err = p.db.Exec(`
insert into company_wiki_compiler_run (
  id, status, metadata, last_error, started_at
) values ($1,'running',$2::jsonb,'',now())
on conflict (id) do update
set status = 'running',
    metadata = coalesce(company_wiki_compiler_run.metadata, '{}'::jsonb) || excluded.metadata,
    last_error = '',
    completed_at = null`, id, rawMetadata)
	return err
}

func (p *PostgresStore) CompleteCompanyWikiCompilerRun(id string, status string, lastError string, metadata map[string]any) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	status = firstNonEmpty(strings.TrimSpace(status), CompanyWikiCompileStatusCompleted)
	if status != CompanyWikiCompileStatusCompleted && status != CompanyWikiCompileStatusFailed {
		status = CompanyWikiCompileStatusFailed
	}
	rawMetadata, err := json.Marshal(nonNilMap(metadata))
	if err != nil {
		return err
	}
	_, err = p.db.Exec(`
update company_wiki_compiler_run
set status = $2,
    metadata = coalesce(metadata, '{}'::jsonb) || $3::jsonb,
    last_error = $4,
    completed_at = now()
where id = $1`, id, status, rawMetadata, strings.TrimSpace(lastError))
	return err
}

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
	return CompanyWikiSourceRevisionResult{Document: doc, Revision: rev, Chunks: chunks, Inserted: inserted, Changed: inserted}, nil
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

func (p *PostgresStore) GetCompanyWikiSourceEvidence(sourceRevisionID string) (CompanyWikiSourceEvidence, bool, error) {
	revision, found, err := p.getCompanyWikiSourceRevision(sourceRevisionID)
	if err != nil || !found {
		return CompanyWikiSourceEvidence{}, found, err
	}
	document, found, err := p.getCompanyWikiSourceDocument(revision.DocumentID)
	if err != nil || !found {
		return CompanyWikiSourceEvidence{}, found, err
	}
	var chunks []CompanyWikiSourceChunk
	if document.SourceType == "slack_message" {
		chunks, err = p.ListCompanyWikiSourceChunks(document.ID)
		if err != nil {
			return CompanyWikiSourceEvidence{}, false, err
		}
	} else {
		chunks, err = p.companyWikiChunksForRevision(revision.ID)
		if err != nil {
			return CompanyWikiSourceEvidence{}, false, err
		}
	}
	return CompanyWikiSourceEvidence{Document: document, Revision: revision, Chunks: chunks}, true, nil
}

func (p *PostgresStore) ListCompanyWikiSourceRevisionIDsWithoutCompileItem(compilerVersion string, schemaVersion string, rendererVersion string, modelPolicyVersion string, limit int) ([]string, error) {
	compilerVersion = strings.TrimSpace(compilerVersion)
	schemaVersion = strings.TrimSpace(schemaVersion)
	rendererVersion = strings.TrimSpace(rendererVersion)
	modelPolicyVersion = strings.TrimSpace(modelPolicyVersion)
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := p.db.Query(`
select r.id
from company_source_revision r
join company_source_document d on d.id = r.document_id
where not exists (
  select 1
  from company_wiki_compile_item item
  where item.source_revision_id = r.id
    and item.compiler_version = $1
    and item.schema_version = $2
    and item.renderer_version = $3
    and item.model_policy_version = $4
)
order by
  case d.source_type
    when 'notion_document' then 0
    when 'slack_message' then 1
    else 2
  end,
  r.created_at asc,
  r.id asc
limit $5`, compilerVersion, schemaVersion, rendererVersion, modelPolicyVersion, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (p *PostgresStore) EnqueueCompanyWikiCompileItem(input CompanyWikiCompileItemInput) (CompanyWikiCompileItem, bool, error) {
	input = normalizeCompanyWikiCompileItemInput(input)
	key := companyWikiCompileItemKey(input.SourceRevisionID, input.CompilerVersion, input.SchemaVersion, input.RendererVersion, input.ModelPolicyVersion)
	id := CompanyWikiStableID("wikicompile", key)
	var inserted bool
	row := p.db.QueryRow(`
insert into company_wiki_compile_item (
  id, source_revision_id, compiler_version, schema_version, renderer_version,
  model_policy_version, input_hash, status, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,now(),now())
on conflict (source_revision_id, compiler_version, schema_version, renderer_version, model_policy_version)
do update set updated_at = company_wiki_compile_item.updated_at
returning id, source_revision_id, compiler_version, schema_version, renderer_version,
          model_policy_version, input_hash, status, lease_holder, lease_expires_at,
          attempt_count, last_attempt_id, last_error, created_at, updated_at,
          (xmax = 0) as inserted`,
		id, input.SourceRevisionID, input.CompilerVersion, input.SchemaVersion, input.RendererVersion,
		input.ModelPolicyVersion, input.InputHash, input.Status)
	item, err := scanCompanyWikiCompileItemWithInserted(row, &inserted)
	if err != nil {
		return CompanyWikiCompileItem{}, false, err
	}
	return item, inserted, nil
}

func (p *PostgresStore) ClaimCompanyWikiCompileItems(input CompanyWikiCompileClaimInput) ([]CompanyWikiCompileItem, error) {
	input = normalizeCompanyWikiCompileClaimInput(input)
	rows, err := p.db.Query(`
with picked as (
  select item.id
  from company_wiki_compile_item item
  join company_source_revision revision on revision.id = item.source_revision_id
  join company_source_document document on document.id = revision.document_id
  where item.compiler_version = $1
    and item.schema_version = $2
    and item.renderer_version = $3
    and item.model_policy_version = $4
    and item.status in ('pending', 'failed')
    and (item.status != 'failed' or item.attempt_count < $8)
    and (item.lease_expires_at is null or item.lease_expires_at < now() or item.lease_holder = $5)
  order by
    case document.source_type
      when 'notion_document' then 0
      when 'slack_message' then 1
      else 2
    end,
    item.updated_at asc,
    item.id asc
  limit $6
  for update of item skip locked
)
update company_wiki_compile_item item
set status = 'claimed',
    lease_holder = $5,
    lease_expires_at = now() + make_interval(secs => $7::int),
    attempt_count = attempt_count + 1,
    last_error = '',
    updated_at = now()
from picked
where item.id = picked.id
returning item.id, item.source_revision_id, item.compiler_version, item.schema_version,
          item.renderer_version, item.model_policy_version, item.input_hash, item.status,
          item.lease_holder, item.lease_expires_at, item.attempt_count,
          item.last_attempt_id, item.last_error, item.created_at, item.updated_at`,
		input.CompilerVersion, input.SchemaVersion, input.RendererVersion, input.ModelPolicyVersion,
		input.LeaseHolder, input.Limit, int(input.LeaseDuration.Seconds()), input.MaxAttempts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CompanyWikiCompileItem{}
	for rows.Next() {
		item, err := scanCompanyWikiCompileItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (p *PostgresStore) BeginCompanyWikiCompileAttempt(input CompanyWikiCompileAttemptInput) (CompanyWikiCompileAttempt, error) {
	input = normalizeCompanyWikiCompileAttemptInput(input)
	id := CompanyWikiStableID("wikiattempt", input.CompileItemID, input.CompilerRunID, input.ContextHash, time.Now().UTC().Format(time.RFC3339Nano))
	validationErrors, _ := json.Marshal(input.ValidationErrors)
	metadata, err := json.Marshal(nonNilMap(input.Metadata))
	if err != nil {
		return CompanyWikiCompileAttempt{}, err
	}
	tx, err := p.db.Begin()
	if err != nil {
		return CompanyWikiCompileAttempt{}, err
	}
	defer func() { _ = tx.Rollback() }()
	row := tx.QueryRow(`
insert into company_wiki_compile_attempt (
  id, compile_item_id, compiler_run_id, status, model, context_hash, output_hash,
  request_metadata_hash, response_metadata_hash, duration_millis,
  validation_errors, last_error, metadata, created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12,$13::jsonb,now())
returning id, compile_item_id, compiler_run_id, status, model, context_hash,
          output_hash, request_metadata_hash, response_metadata_hash, duration_millis,
          validation_errors, last_error, metadata, created_at, completed_at`,
		id, input.CompileItemID, input.CompilerRunID, input.Status, input.Model, input.ContextHash,
		input.OutputHash, input.RequestMetadataHash, input.ResponseMetadataHash,
		input.DurationMillis, validationErrors, input.LastError, metadata)
	attempt, err := scanCompanyWikiCompileAttempt(row)
	if err != nil {
		return CompanyWikiCompileAttempt{}, err
	}
	if _, err := tx.Exec(`update company_wiki_compile_item set last_attempt_id = $2, updated_at = now() where id = $1`, input.CompileItemID, attempt.ID); err != nil {
		return CompanyWikiCompileAttempt{}, err
	}
	if err := tx.Commit(); err != nil {
		return CompanyWikiCompileAttempt{}, err
	}
	return attempt, nil
}

func (p *PostgresStore) CompleteCompanyWikiCompileAttempt(attemptID string, status string, outputHash string, durationMillis int64, validationErrors []string, lastError string, metadata map[string]any) (CompanyWikiCompileAttempt, error) {
	validationErrorsRaw, _ := json.Marshal(validationErrors)
	metadataRaw, err := json.Marshal(nonNilMap(metadata))
	if err != nil {
		return CompanyWikiCompileAttempt{}, err
	}
	requestMetadataHash := stringFromAnyMap(metadata, "request_metadata_hash")
	responseMetadataHash := stringFromAnyMap(metadata, "response_metadata_hash")
	return scanCompanyWikiCompileAttempt(p.db.QueryRow(`
update company_wiki_compile_attempt
set status = $2,
    output_hash = $3,
    duration_millis = $4,
    validation_errors = $5::jsonb,
    last_error = $6,
    metadata = coalesce(metadata, '{}'::jsonb) || $7::jsonb,
    request_metadata_hash = coalesce(nullif($8, ''), request_metadata_hash),
    response_metadata_hash = coalesce(nullif($9, ''), response_metadata_hash),
    completed_at = now()
where id = $1
returning id, compile_item_id, compiler_run_id, status, model, context_hash,
          output_hash, request_metadata_hash, response_metadata_hash, duration_millis,
          validation_errors, last_error, metadata, created_at, completed_at`,
		strings.TrimSpace(attemptID), firstNonEmpty(strings.TrimSpace(status), CompanyWikiCompileStatusCompleted),
		strings.TrimSpace(outputHash), durationMillis, validationErrorsRaw, strings.TrimSpace(lastError), metadataRaw,
		requestMetadataHash, responseMetadataHash))
}

func scanCompanyWikiCompileItemWithInserted(row companyWikiRow, inserted *bool) (CompanyWikiCompileItem, error) {
	var item CompanyWikiCompileItem
	var leaseExpiresAt sql.NullTime
	if err := row.Scan(
		&item.ID, &item.SourceRevisionID, &item.CompilerVersion, &item.SchemaVersion,
		&item.RendererVersion, &item.ModelPolicyVersion, &item.InputHash, &item.Status,
		&item.LeaseHolder, &leaseExpiresAt, &item.AttemptCount, &item.LastAttemptID,
		&item.LastError, &item.CreatedAt, &item.UpdatedAt, inserted,
	); err != nil {
		return CompanyWikiCompileItem{}, err
	}
	if leaseExpiresAt.Valid {
		item.LeaseExpiresAt = leaseExpiresAt.Time
	}
	return item, nil
}

func (p *PostgresStore) UpsertCompanyWikiCompileTargets(compileItemID string, targets []CompanyWikiCompileTargetInput) ([]CompanyWikiCompileTarget, error) {
	compileItemID = strings.TrimSpace(compileItemID)
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	desired := map[string]struct{}{}
	out := []CompanyWikiCompileTarget{}
	for _, input := range targets {
		input.CompileItemID = compileItemID
		input = normalizeCompanyWikiCompileTargetInput(input)
		desired[input.TargetSlug] = struct{}{}
		id := CompanyWikiStableID("wikitarget", compileItemID, input.TargetSlug)
		item, err := scanCompanyWikiCompileTarget(tx.QueryRow(`
insert into company_wiki_compile_item_target (
  id, compile_item_id, target_slug, target_path, target_type, status,
  wiki_revision_id, idempotency_key, body_hash, last_error, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now(),now())
on conflict (compile_item_id, target_slug) do update
set target_path = excluded.target_path,
    target_type = excluded.target_type,
    status = case
      when company_wiki_compile_item_target.status = 'published'
       and company_wiki_compile_item_target.body_hash = excluded.body_hash
      then company_wiki_compile_item_target.status
      else excluded.status
    end,
    wiki_revision_id = coalesce(nullif(excluded.wiki_revision_id, ''), company_wiki_compile_item_target.wiki_revision_id),
    idempotency_key = excluded.idempotency_key,
    body_hash = excluded.body_hash,
    last_error = excluded.last_error,
    updated_at = now()
returning id, compile_item_id, target_slug, target_path, target_type, status,
          wiki_revision_id, idempotency_key, body_hash, last_error, created_at, updated_at`,
			id, input.CompileItemID, input.TargetSlug, input.TargetPath, input.TargetType,
			input.Status, input.WikiRevisionID, input.IdempotencyKey, input.BodyHash, input.LastError))
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	rows, err := tx.Query(`
select id, target_slug
from company_wiki_compile_item_target
where compile_item_id = $1`, compileItemID)
	if err != nil {
		return nil, err
	}
	type targetRef struct{ id, slug string }
	existing := []targetRef{}
	for rows.Next() {
		var item targetRef
		if err := rows.Scan(&item.id, &item.slug); err != nil {
			_ = rows.Close()
			return nil, err
		}
		existing = append(existing, item)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for _, item := range existing {
		if _, ok := desired[item.slug]; ok {
			continue
		}
		if _, err := tx.Exec(`
update company_wiki_compile_item_target
set status = 'superseded', updated_at = now()
where id = $1`, item.id); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (p *PostgresStore) UpdateCompanyWikiCompileTarget(input CompanyWikiCompileTargetInput) (CompanyWikiCompileTarget, error) {
	input = normalizeCompanyWikiCompileTargetInput(input)
	id := CompanyWikiStableID("wikitarget", input.CompileItemID, input.TargetSlug)
	return scanCompanyWikiCompileTarget(p.db.QueryRow(`
insert into company_wiki_compile_item_target (
  id, compile_item_id, target_slug, target_path, target_type, status,
  wiki_revision_id, idempotency_key, body_hash, last_error, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now(),now())
on conflict (compile_item_id, target_slug) do update
set target_path = coalesce(nullif(excluded.target_path, ''), company_wiki_compile_item_target.target_path),
    target_type = coalesce(nullif(excluded.target_type, ''), company_wiki_compile_item_target.target_type),
    status = excluded.status,
    wiki_revision_id = coalesce(nullif(excluded.wiki_revision_id, ''), company_wiki_compile_item_target.wiki_revision_id),
    idempotency_key = coalesce(nullif(excluded.idempotency_key, ''), company_wiki_compile_item_target.idempotency_key),
    body_hash = coalesce(nullif(excluded.body_hash, ''), company_wiki_compile_item_target.body_hash),
    last_error = excluded.last_error,
    updated_at = now()
returning id, compile_item_id, target_slug, target_path, target_type, status,
          wiki_revision_id, idempotency_key, body_hash, last_error, created_at, updated_at`,
		id, input.CompileItemID, input.TargetSlug, input.TargetPath, input.TargetType,
		input.Status, input.WikiRevisionID, input.IdempotencyKey, input.BodyHash, input.LastError))
}

func (p *PostgresStore) ListCompanyWikiCompileTargets(compileItemID string) ([]CompanyWikiCompileTarget, error) {
	rows, err := p.db.Query(`
select id, compile_item_id, target_slug, target_path, target_type, status,
       wiki_revision_id, idempotency_key, body_hash, last_error, created_at, updated_at
from company_wiki_compile_item_target
where compile_item_id = $1
order by target_slug asc`, strings.TrimSpace(compileItemID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CompanyWikiCompileTarget{}
	for rows.Next() {
		item, err := scanCompanyWikiCompileTarget(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (p *PostgresStore) CompleteCompanyWikiCompileItem(compileItemID string, status string, lastError string) (CompanyWikiCompileItem, error) {
	return scanCompanyWikiCompileItem(p.db.QueryRow(`
update company_wiki_compile_item
set status = $2,
    last_error = $3,
    lease_holder = case when $2 in ('completed', 'skipped') then '' else lease_holder end,
    lease_expires_at = case when $2 in ('completed', 'skipped') then null else lease_expires_at end,
    updated_at = now()
where id = $1
returning id, source_revision_id, compiler_version, schema_version, renderer_version,
          model_policy_version, input_hash, status, lease_holder, lease_expires_at,
          attempt_count, last_attempt_id, last_error, created_at, updated_at`,
		strings.TrimSpace(compileItemID), firstNonEmpty(strings.TrimSpace(status), CompanyWikiCompileStatusCompleted), strings.TrimSpace(lastError)))
}

func (p *PostgresStore) ReleaseCompanyWikiCompileItems(ids []string, leaseHolder string, reason string) (int, error) {
	leaseHolder = strings.TrimSpace(leaseHolder)
	if len(ids) == 0 || leaseHolder == "" {
		return 0, nil
	}
	tx, err := p.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	released := int64(0)
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		result, err := tx.Exec(`
update company_wiki_compile_item
set status = 'pending',
    lease_holder = '',
    lease_expires_at = null,
    attempt_count = greatest(attempt_count - 1, 0),
    last_error = $3,
    updated_at = now()
where id = $1
  and status = 'claimed'
  and lease_holder = $2`,
			id, leaseHolder, strings.TrimSpace(reason))
		if err != nil {
			return 0, err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		released += rows
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int(released), nil
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
	if _, err := tx.Exec(`delete from company_wiki_claim where wiki_revision_id = $1`, revisionID); err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	if _, err := tx.Exec(`delete from company_wiki_conflict where wiki_revision_id = $1`, revisionID); err != nil {
		return CompanyWikiPagePublishResult{}, err
	}
	citations := make([]CompanyWikiCitation, 0, len(input.Citations))
	citationIDsByClaim := map[string][]string{}
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
		citationIDsByClaim[inserted.ClaimKey] = append(citationIDsByClaim[inserted.ClaimKey], inserted.ID)
	}
	claims := make([]CompanyWikiClaim, 0, len(input.Claims))
	for _, claim := range input.Claims {
		if strings.TrimSpace(claim.ClaimKey) == "" || strings.TrimSpace(claim.ClaimText) == "" {
			continue
		}
		if claim.Confidence == 0 {
			claim.Confidence = 1
		}
		rawClaimMetadata, err := json.Marshal(nonNilMap(claim.Metadata))
		if err != nil {
			return CompanyWikiPagePublishResult{}, err
		}
		claimID := CompanyWikiStableID("wikiclaim", revisionID, claim.ClaimKey, claim.ClaimText)
		inserted, err := scanCompanyWikiClaim(tx.QueryRow(`
insert into company_wiki_claim (
  id, wiki_revision_id, claim_key, claim_text, confidence, metadata, created_at
) values ($1,$2,$3,$4,$5,$6::jsonb,now())
returning id, wiki_revision_id, claim_key, claim_text, confidence, metadata, created_at`,
			claimID, revisionID, strings.TrimSpace(claim.ClaimKey), strings.TrimSpace(claim.ClaimText), claim.Confidence, rawClaimMetadata))
		if err != nil {
			return CompanyWikiPagePublishResult{}, err
		}
		claims = append(claims, inserted)
	}
	conflicts := make([]CompanyWikiConflict, 0, len(input.Conflicts))
	for _, conflict := range input.Conflicts {
		if strings.TrimSpace(conflict.ClaimKey) == "" || strings.TrimSpace(conflict.Summary) == "" {
			continue
		}
		citationIDs := append([]string(nil), conflict.Citations...)
		if len(citationIDs) == 0 {
			citationIDs = append(citationIDs, citationIDsByClaim[conflict.ClaimKey]...)
		}
		rawCitationIDs, _ := json.Marshal(citationIDs)
		rawConflictMetadata, err := json.Marshal(nonNilMap(conflict.Metadata))
		if err != nil {
			return CompanyWikiPagePublishResult{}, err
		}
		conflictID := CompanyWikiStableID("wikiconflict", revisionID, conflict.ClaimKey, conflict.Summary)
		inserted, err := scanCompanyWikiConflict(tx.QueryRow(`
insert into company_wiki_conflict (
  id, wiki_revision_id, claim_key, summary, citation_ids, metadata, created_at
) values ($1,$2,$3,$4,$5::jsonb,$6::jsonb,now())
returning id, wiki_revision_id, claim_key, summary, citation_ids, metadata, created_at`,
			conflictID, revisionID, strings.TrimSpace(conflict.ClaimKey), strings.TrimSpace(conflict.Summary), rawCitationIDs, rawConflictMetadata))
		if err != nil {
			return CompanyWikiPagePublishResult{}, err
		}
		for _, citationID := range citationIDs {
			if strings.TrimSpace(citationID) == "" {
				continue
			}
			if _, err := tx.Exec(`
insert into company_wiki_conflict_citation (conflict_id, citation_id, created_at)
values ($1,$2,now())
on conflict (conflict_id, citation_id) do nothing`, conflictID, citationID); err != nil {
				return CompanyWikiPagePublishResult{}, err
			}
		}
		conflicts = append(conflicts, inserted)
	}
	if _, err := tx.Exec(`
insert into company_wiki_manifest (path, wiki_page_id, wiki_revision_id, sha256, compiler_run_id, generated_at, repair_status, last_repair_error)
values ($1,$2,$3,$4,$5,$6,'ok','')
on conflict (path) do update
set wiki_page_id = excluded.wiki_page_id,
    wiki_revision_id = excluded.wiki_revision_id,
    sha256 = excluded.sha256,
    compiler_run_id = excluded.compiler_run_id,
    generated_at = excluded.generated_at,
    repair_status = 'ok',
    last_repair_error = '',
    last_repaired_at = now()`,
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
	return CompanyWikiPagePublishResult{Page: read.Page, Revision: read.Revision, Citations: citations, Claims: claims, Conflicts: conflicts}, nil
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
       m.path, m.wiki_page_id, m.wiki_revision_id, m.sha256, m.compiler_run_id, m.generated_at,
       m.repair_status, m.last_repair_error, m.last_checked_at, m.last_repaired_at
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
	claims, err := p.companyWikiClaimsForRevision(read.Revision.ID)
	if err != nil {
		return CompanyWikiPageRead{}, false, err
	}
	read.Claims = claims
	conflicts, err := p.companyWikiConflictsForRevision(read.Revision.ID)
	if err != nil {
		return CompanyWikiPageRead{}, false, err
	}
	read.Conflicts = conflicts
	return read, true, nil
}

func (p *PostgresStore) ListCompanyWikiManifestEntries() ([]CompanyWikiManifestEntry, error) {
	rows, err := p.db.Query(`
select path, wiki_page_id, wiki_revision_id, sha256, compiler_run_id, generated_at,
       repair_status, last_repair_error, last_checked_at, last_repaired_at
from company_wiki_manifest
order by path asc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CompanyWikiManifestEntry{}
	for rows.Next() {
		var item CompanyWikiManifestEntry
		var lastChecked, lastRepaired sql.NullTime
		if err := rows.Scan(&item.Path, &item.WikiPageID, &item.WikiRevisionID, &item.SHA256, &item.CompilerRunID, &item.GeneratedAt, &item.RepairStatus, &item.LastRepairError, &lastChecked, &lastRepaired); err != nil {
			return nil, err
		}
		if lastChecked.Valid {
			item.LastCheckedAt = lastChecked.Time
		}
		if lastRepaired.Valid {
			item.LastRepairedAt = lastRepaired.Time
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
		metadata := cloneAnyMap(input.Metadata)
		if !input.ObservedAt.IsZero() {
			metadata["source_observed_at"] = input.ObservedAt.UTC().Format(time.RFC3339)
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
			Metadata:      metadata,
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

func (p *PostgresStore) companyWikiClaimsForRevision(revisionID string) ([]CompanyWikiClaim, error) {
	rows, err := p.db.Query(`
select id, wiki_revision_id, claim_key, claim_text, confidence, metadata, created_at
from company_wiki_claim
where wiki_revision_id = $1
order by claim_key asc, created_at asc`, strings.TrimSpace(revisionID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CompanyWikiClaim{}
	for rows.Next() {
		item, err := scanCompanyWikiClaim(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (p *PostgresStore) companyWikiConflictsForRevision(revisionID string) ([]CompanyWikiConflict, error) {
	rows, err := p.db.Query(`
select id, wiki_revision_id, claim_key, summary, citation_ids, metadata, created_at
from company_wiki_conflict
where wiki_revision_id = $1
order by claim_key asc, created_at asc`, strings.TrimSpace(revisionID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CompanyWikiConflict{}
	for rows.Next() {
		item, err := scanCompanyWikiConflict(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (p *PostgresStore) ListCompanyWikiCandidatePages(query CompanyWikiPageQuery) ([]CompanyWikiPageRead, error) {
	query.Query = strings.ToLower(strings.TrimSpace(query.Query))
	if query.Limit <= 0 || query.Limit > 50 {
		query.Limit = 10
	}
	excludeEvidence := query.ExcludeEvidence
	rows, err := p.db.Query(`
select p.id, p.slug, p.title, p.status, p.current_revision_id, p.metadata, p.created_at, p.updated_at,
       r.id, r.page_id, r.revision_number, r.compiler_run_id, r.title, r.body, r.body_sha256,
       r.path, r.source_revision_ids, r.metadata, r.published_at, r.created_at,
       m.path, m.wiki_page_id, m.wiki_revision_id, m.sha256, m.compiler_run_id, m.generated_at,
       m.repair_status, m.last_repair_error, m.last_checked_at, m.last_repaired_at
from company_wiki_page p
join company_wiki_revision r on r.id = p.current_revision_id
left join company_wiki_manifest m on m.wiki_revision_id = r.id
where ($1 = '' or lower(p.slug) like '%' || $1 || '%' escape '\' or lower(p.title) like '%' || $1 || '%' escape '\' or lower(r.path) like '%' || $1 || '%' escape '\' or lower(r.body) like '%' || $1 || '%' escape '\')
  and (not $2 or r.path not like 'sources/%')
  and (coalesce(array_length($3::text[], 1), 0) = 0 or p.metadata->>'type' = any($3::text[]) or r.metadata->>'type' = any($3::text[]))
  and (coalesce(array_length($4::text[], 1), 0) = 0 or p.metadata->'tags' ?| $4::text[] or r.metadata->'tags' ?| $4::text[])
order by r.path asc
limit $5`, escapeLikePattern(query.Query), excludeEvidence, textArrayLiteral(query.Types), textArrayLiteral(query.Tags), query.Limit)
	if err != nil {
		return nil, err
	}
	out := []CompanyWikiPageRead{}
	for rows.Next() {
		item, found, err := scanCompanyWikiPageRead(rows)
		if err != nil {
			return nil, err
		}
		if found {
			out = append(out, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for i := range out {
		read, found, err := p.GetCompanyWikiPage(out[i].Page.ID)
		if err != nil {
			return nil, err
		}
		if found {
			out[i] = read
		}
	}
	return out, nil
}

func (p *PostgresStore) UpdateCompanyWikiManifestRepair(path string, status string, lastError string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	status = firstNonEmpty(strings.TrimSpace(status), CompanyWikiManifestRepairOK)
	_, err := p.db.Exec(`
update company_wiki_manifest
set repair_status = $2,
    last_repair_error = $3,
    last_checked_at = now(),
    last_repaired_at = case when $2 = 'ok' then now() else last_repaired_at end
where path = $1`, path, status, strings.TrimSpace(lastError))
	return err
}

func textArrayLiteral(values []string) []string {
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
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
	if err := ValidateCompanyWikiCitationInputs(input.Citations); err != nil {
		return err
	}
	citationsByClaim := map[string]int{}
	for _, citation := range input.Citations {
		citationsByClaim[strings.TrimSpace(citation.ClaimKey)]++
	}
	for _, claim := range input.Claims {
		claimKey := strings.TrimSpace(claim.ClaimKey)
		if claimKey == "" {
			return errors.New("claim.claim_key is required")
		}
		if strings.TrimSpace(claim.ClaimText) == "" {
			return errors.New("claim.claim_text is required")
		}
		if citationsByClaim[claimKey] == 0 {
			return fmt.Errorf("claim %q must have at least one matching citation", claimKey)
		}
	}
	for _, conflict := range input.Conflicts {
		claimKey := strings.TrimSpace(conflict.ClaimKey)
		if claimKey == "" {
			return errors.New("conflict.claim_key is required")
		}
		if strings.TrimSpace(conflict.Summary) == "" {
			return errors.New("conflict.summary is required")
		}
		if len(conflict.Citations) == 0 && citationsByClaim[claimKey] == 0 {
			return fmt.Errorf("conflict %q must cite a claim with source citations", claimKey)
		}
	}
	return nil
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
