package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (p *PostgresStore) ClaimSourceMirrorRecord(record SourceMirrorRecord, lease time.Duration) (SourceMirrorClaimResult, error) {
	if err := validateSourceMirrorRecord(record); err != nil {
		return SourceMirrorClaimResult{}, err
	}
	if lease <= 0 {
		lease = 5 * time.Minute
	}
	tx, err := p.db.Begin()
	if err != nil {
		return SourceMirrorClaimResult{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	existing, found, err := scanSourceMirrorRecord(tx.QueryRow(`
select source_type, source_key, workspace, environment, source_session_key, honcho_workspace,
       honcho_session_id, honcho_message_id, source_revision, status, metadata, last_error,
       created_at, updated_at
from source_mirror_record
where source_type = $1 and source_key = $2
for update`, record.SourceType, record.SourceKey))
	if err != nil {
		return SourceMirrorClaimResult{}, err
	}
	now := time.Now().UTC()
	if !found {
		inserted, err := insertSourceMirrorRecord(tx, record, now)
		if err != nil {
			return SourceMirrorClaimResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return SourceMirrorClaimResult{}, err
		}
		return SourceMirrorClaimResult{Record: inserted, ShouldWrite: true, Reason: "new"}, nil
	}

	if existing.Status == SourceMirrorStatusComplete && existing.SourceRevision == record.SourceRevision && strings.TrimSpace(existing.HonchoMessageID) != "" {
		if err := tx.Commit(); err != nil {
			return SourceMirrorClaimResult{}, err
		}
		return SourceMirrorClaimResult{Record: existing, ShouldWrite: false, Reason: "already_complete"}, nil
	}

	if existing.Status == SourceMirrorStatusPending && now.Sub(existing.UpdatedAt) < lease {
		if err := tx.Commit(); err != nil {
			return SourceMirrorClaimResult{}, err
		}
		return SourceMirrorClaimResult{Record: existing, ShouldWrite: false, Reason: "leased"}, nil
	}

	record.Status = SourceMirrorStatusPending
	record.HonchoMessageID = ""
	record.LastError = ""
	updated, err := updateSourceMirrorClaim(tx, record, now)
	if err != nil {
		return SourceMirrorClaimResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return SourceMirrorClaimResult{}, err
	}
	reason := "retry"
	if existing.SourceRevision != record.SourceRevision {
		reason = "revision_changed"
	}
	return SourceMirrorClaimResult{Record: updated, ShouldWrite: true, Reason: reason}, nil
}

func (p *PostgresStore) CompleteSourceMirrorRecord(sourceType string, sourceKey string, honchoMessageID string, metadata map[string]any) (SourceMirrorRecord, error) {
	honchoMessageID = strings.TrimSpace(honchoMessageID)
	if honchoMessageID == "" {
		return SourceMirrorRecord{}, errors.New("honcho message id is required")
	}
	raw, err := json.Marshal(nonNilMap(metadata))
	if err != nil {
		return SourceMirrorRecord{}, err
	}
	record, found, err := scanSourceMirrorRecord(p.db.QueryRow(`
update source_mirror_record
set status = 'complete',
    honcho_message_id = $3,
    metadata = coalesce(metadata, '{}'::jsonb) || $4::jsonb,
    last_error = '',
    updated_at = now()
where source_type = $1 and source_key = $2
returning source_type, source_key, workspace, environment, source_session_key, honcho_workspace,
          honcho_session_id, honcho_message_id, source_revision, status, metadata, last_error,
          created_at, updated_at`, sourceType, sourceKey, honchoMessageID, raw))
	if err != nil {
		return SourceMirrorRecord{}, err
	}
	if !found {
		return SourceMirrorRecord{}, sql.ErrNoRows
	}
	return record, nil
}

func (p *PostgresStore) FailSourceMirrorRecord(sourceType string, sourceKey string, lastError string, metadata map[string]any) (SourceMirrorRecord, error) {
	raw, err := json.Marshal(nonNilMap(metadata))
	if err != nil {
		return SourceMirrorRecord{}, err
	}
	record, found, err := scanSourceMirrorRecord(p.db.QueryRow(`
update source_mirror_record
set status = 'failed',
    metadata = coalesce(metadata, '{}'::jsonb) || $4::jsonb,
    last_error = $3,
    updated_at = now()
where source_type = $1 and source_key = $2
returning source_type, source_key, workspace, environment, source_session_key, honcho_workspace,
          honcho_session_id, honcho_message_id, source_revision, status, metadata, last_error,
          created_at, updated_at`, sourceType, sourceKey, strings.TrimSpace(lastError), raw))
	if err != nil {
		return SourceMirrorRecord{}, err
	}
	if !found {
		return SourceMirrorRecord{}, sql.ErrNoRows
	}
	return record, nil
}

func (p *PostgresStore) GetSourceMirrorRecord(sourceType string, sourceKey string) (SourceMirrorRecord, bool, error) {
	return scanSourceMirrorRecord(p.db.QueryRow(`
select source_type, source_key, workspace, environment, source_session_key, honcho_workspace,
       honcho_session_id, honcho_message_id, source_revision, status, metadata, last_error,
       created_at, updated_at
from source_mirror_record
where source_type = $1 and source_key = $2`, sourceType, sourceKey))
}

type sourceMirrorQuerier interface {
	QueryRow(query string, args ...any) *sql.Row
}

func insertSourceMirrorRecord(tx sourceMirrorQuerier, record SourceMirrorRecord, now time.Time) (SourceMirrorRecord, error) {
	raw, err := json.Marshal(nonNilMap(record.Metadata))
	if err != nil {
		return SourceMirrorRecord{}, err
	}
	returning, found, err := scanSourceMirrorRecord(tx.QueryRow(`
insert into source_mirror_record (
  source_type, source_key, workspace, environment, source_session_key, honcho_workspace,
  honcho_session_id, honcho_message_id, source_revision, status, metadata, last_error,
  created_at, updated_at
) values ($1, $2, $3, $4, $5, $6, $7, '', $8, 'pending', $9::jsonb, '', $10, $10)
returning source_type, source_key, workspace, environment, source_session_key, honcho_workspace,
          honcho_session_id, honcho_message_id, source_revision, status, metadata, last_error,
          created_at, updated_at`,
		record.SourceType, record.SourceKey, record.Workspace, record.Environment, record.SourceSessionKey,
		record.HonchoWorkspace, record.HonchoSessionID, record.SourceRevision, raw, now))
	if err != nil {
		return SourceMirrorRecord{}, err
	}
	if !found {
		return SourceMirrorRecord{}, sql.ErrNoRows
	}
	return returning, nil
}

func updateSourceMirrorClaim(tx sourceMirrorQuerier, record SourceMirrorRecord, now time.Time) (SourceMirrorRecord, error) {
	raw, err := json.Marshal(nonNilMap(record.Metadata))
	if err != nil {
		return SourceMirrorRecord{}, err
	}
	returning, found, err := scanSourceMirrorRecord(tx.QueryRow(`
update source_mirror_record
set workspace = $3,
    environment = $4,
    source_session_key = $5,
    honcho_workspace = $6,
    honcho_session_id = $7,
    honcho_message_id = '',
    source_revision = $8,
    status = 'pending',
    metadata = $9::jsonb,
    last_error = '',
    updated_at = $10
where source_type = $1 and source_key = $2
returning source_type, source_key, workspace, environment, source_session_key, honcho_workspace,
          honcho_session_id, honcho_message_id, source_revision, status, metadata, last_error,
          created_at, updated_at`,
		record.SourceType, record.SourceKey, record.Workspace, record.Environment, record.SourceSessionKey,
		record.HonchoWorkspace, record.HonchoSessionID, record.SourceRevision, raw, now))
	if err != nil {
		return SourceMirrorRecord{}, err
	}
	if !found {
		return SourceMirrorRecord{}, sql.ErrNoRows
	}
	return returning, nil
}

func scanSourceMirrorRecord(row *sql.Row) (SourceMirrorRecord, bool, error) {
	var record SourceMirrorRecord
	var metadata []byte
	err := row.Scan(
		&record.SourceType,
		&record.SourceKey,
		&record.Workspace,
		&record.Environment,
		&record.SourceSessionKey,
		&record.HonchoWorkspace,
		&record.HonchoSessionID,
		&record.HonchoMessageID,
		&record.SourceRevision,
		&record.Status,
		&metadata,
		&record.LastError,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return SourceMirrorRecord{}, false, nil
	}
	if err != nil {
		return SourceMirrorRecord{}, false, err
	}
	record.Metadata = decodeJSON(metadata, map[string]any{})
	return record, true, nil
}

func validateSourceMirrorRecord(record SourceMirrorRecord) error {
	if strings.TrimSpace(record.SourceType) == "" {
		return errors.New("source type is required")
	}
	if strings.TrimSpace(record.SourceKey) == "" {
		return errors.New("source key is required")
	}
	if strings.TrimSpace(record.SourceSessionKey) == "" {
		return errors.New("source session key is required")
	}
	if strings.TrimSpace(record.HonchoWorkspace) == "" {
		return errors.New("honcho workspace is required")
	}
	if strings.TrimSpace(record.HonchoSessionID) == "" {
		return errors.New("honcho session id is required")
	}
	if strings.TrimSpace(record.SourceRevision) == "" {
		return errors.New("source revision is required")
	}
	switch strings.TrimSpace(record.Status) {
	case "", SourceMirrorStatusPending:
		return nil
	default:
		return fmt.Errorf("claim status must be pending, got %q", record.Status)
	}
}

func nonNilMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	return input
}
