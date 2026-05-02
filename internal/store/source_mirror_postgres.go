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
       honcho_session_id, honcho_message_id, honcho_object_type, honcho_object_id,
       source_revision, status, metadata, last_error,
       created_at, updated_at
from source_mirror_record
where source_type = $1 and source_key = $2
for update`, record.SourceType, record.SourceKey))
	if err != nil {
		return SourceMirrorClaimResult{}, err
	}
	now := time.Now().UTC()
	if !found {
		inserted, insertedRecord, err := insertSourceMirrorRecord(tx, record, now)
		if err != nil {
			return SourceMirrorClaimResult{}, err
		}
		if insertedRecord {
			if err := tx.Commit(); err != nil {
				return SourceMirrorClaimResult{}, err
			}
			return SourceMirrorClaimResult{Record: inserted, ShouldWrite: true, Reason: "new"}, nil
		}
		existing, found, err = scanSourceMirrorRecord(tx.QueryRow(`
select source_type, source_key, workspace, environment, source_session_key, honcho_workspace,
       honcho_session_id, honcho_message_id, honcho_object_type, honcho_object_id,
       source_revision, status, metadata, last_error,
       created_at, updated_at
from source_mirror_record
where source_type = $1 and source_key = $2
for update`, record.SourceType, record.SourceKey))
		if err != nil {
			return SourceMirrorClaimResult{}, err
		}
		if !found {
			return SourceMirrorClaimResult{}, sql.ErrNoRows
		}
	}

	if existing.Status == SourceMirrorStatusComplete && existing.SourceRevision == record.SourceRevision && sourceMirrorRecordHasHonchoObject(existing) {
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
	record.HonchoObjectType = ""
	record.HonchoObjectID = ""
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
	return p.CompleteSourceMirrorObject(sourceType, sourceKey, "message", honchoMessageID, metadata)
}

func (p *PostgresStore) CompleteSourceMirrorObject(sourceType string, sourceKey string, honchoObjectType string, honchoObjectID string, metadata map[string]any) (SourceMirrorRecord, error) {
	honchoObjectType = strings.TrimSpace(honchoObjectType)
	honchoObjectID = strings.TrimSpace(honchoObjectID)
	if honchoObjectType == "" {
		return SourceMirrorRecord{}, errors.New("honcho object type is required")
	}
	if honchoObjectID == "" {
		return SourceMirrorRecord{}, errors.New("honcho object id is required")
	}
	honchoMessageID := ""
	if honchoObjectType == "message" {
		honchoMessageID = honchoObjectID
	}
	raw, err := json.Marshal(nonNilMap(metadata))
	if err != nil {
		return SourceMirrorRecord{}, err
	}
	record, found, err := scanSourceMirrorRecord(p.db.QueryRow(`
update source_mirror_record
set status = 'complete',
    honcho_message_id = $3,
    honcho_object_type = $4,
    honcho_object_id = $5,
    metadata = coalesce(metadata, '{}'::jsonb) || $6::jsonb,
    last_error = '',
    updated_at = now()
where source_type = $1 and source_key = $2
returning source_type, source_key, workspace, environment, source_session_key, honcho_workspace,
          honcho_session_id, honcho_message_id, honcho_object_type, honcho_object_id,
          source_revision, status, metadata, last_error,
          created_at, updated_at`, sourceType, sourceKey, honchoMessageID, honchoObjectType, honchoObjectID, raw))
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
          honcho_session_id, honcho_message_id, honcho_object_type, honcho_object_id,
          source_revision, status, metadata, last_error,
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
       honcho_session_id, honcho_message_id, honcho_object_type, honcho_object_id,
       source_revision, status, metadata, last_error,
       created_at, updated_at
from source_mirror_record
where source_type = $1 and source_key = $2`, sourceType, sourceKey))
}

func (p *PostgresStore) ListSourceMirrorRecords(sourceTypes []string, limit int) ([]SourceMirrorRecord, error) {
	args := []any{}
	query := `
select source_type, source_key, workspace, environment, source_session_key, honcho_workspace,
       honcho_session_id, honcho_message_id, honcho_object_type, honcho_object_id,
       source_revision, status, metadata, last_error,
       created_at, updated_at
from source_mirror_record`
	wanted := sourceMirrorTypeSet(sourceTypes)
	if len(wanted) > 0 {
		placeholders := make([]string, 0, len(wanted))
		for sourceType := range wanted {
			args = append(args, sourceType)
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
		}
		query += " where source_type in (" + strings.Join(placeholders, ",") + ")"
	}
	query += " order by updated_at desc"
	if limit > 0 {
		args = append(args, limit)
		query += fmt.Sprintf(" limit $%d", len(args))
	}
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []SourceMirrorRecord
	for rows.Next() {
		record, err := scanSourceMirrorRecordScanner(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

type sourceMirrorQuerier interface {
	QueryRow(query string, args ...any) *sql.Row
}

func insertSourceMirrorRecord(tx sourceMirrorQuerier, record SourceMirrorRecord, now time.Time) (SourceMirrorRecord, bool, error) {
	raw, err := json.Marshal(nonNilMap(record.Metadata))
	if err != nil {
		return SourceMirrorRecord{}, false, err
	}
	returning, found, err := scanSourceMirrorRecord(tx.QueryRow(`
insert into source_mirror_record (
  source_type, source_key, workspace, environment, source_session_key, honcho_workspace,
  honcho_session_id, honcho_message_id, honcho_object_type, honcho_object_id,
  source_revision, status, metadata, last_error,
  created_at, updated_at
) values ($1, $2, $3, $4, $5, $6, $7, '', '', '', $8, 'pending', $9::jsonb, '', $10, $10)
on conflict (source_type, source_key) do nothing
returning source_type, source_key, workspace, environment, source_session_key, honcho_workspace,
          honcho_session_id, honcho_message_id, honcho_object_type, honcho_object_id,
          source_revision, status, metadata, last_error,
          created_at, updated_at`,
		record.SourceType, record.SourceKey, record.Workspace, record.Environment, record.SourceSessionKey,
		record.HonchoWorkspace, record.HonchoSessionID, record.SourceRevision, raw, now))
	if err != nil {
		return SourceMirrorRecord{}, false, err
	}
	if !found {
		return SourceMirrorRecord{}, false, nil
	}
	return returning, true, nil
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
    honcho_object_type = '',
    honcho_object_id = '',
    source_revision = $8,
    status = 'pending',
    metadata = $9::jsonb,
    last_error = '',
    updated_at = $10
where source_type = $1 and source_key = $2
returning source_type, source_key, workspace, environment, source_session_key, honcho_workspace,
          honcho_session_id, honcho_message_id, honcho_object_type, honcho_object_id,
          source_revision, status, metadata, last_error,
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
	record, err := scanSourceMirrorRecordScanner(row)
	if errors.Is(err, sql.ErrNoRows) {
		return SourceMirrorRecord{}, false, nil
	}
	if err != nil {
		return SourceMirrorRecord{}, false, err
	}
	return record, true, nil
}

type sourceMirrorScanner interface {
	Scan(dest ...any) error
}

func scanSourceMirrorRecordScanner(row sourceMirrorScanner) (SourceMirrorRecord, error) {
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
		&record.HonchoObjectType,
		&record.HonchoObjectID,
		&record.SourceRevision,
		&record.Status,
		&metadata,
		&record.LastError,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		return SourceMirrorRecord{}, err
	}
	record.Metadata = decodeJSON(metadata, map[string]any{})
	return record, nil
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
