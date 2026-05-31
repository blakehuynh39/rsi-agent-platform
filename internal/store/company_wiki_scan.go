package store

import (
	"database/sql"
	"encoding/json"
)

type companyWikiRow interface {
	Scan(dest ...any) error
}

func scanCompanyWikiSourceDocument(row companyWikiRow) (CompanyWikiSourceDocument, bool, error) {
	var item CompanyWikiSourceDocument
	var metadata []byte
	err := row.Scan(
		&item.ID, &item.SourceType, &item.SourceKey, &item.SourceSessionKey,
		&item.Workspace, &item.Environment, &item.Title, &item.URL, &item.Status,
		&item.CurrentRevisionID, &metadata, &item.CreatedAt, &item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return CompanyWikiSourceDocument{}, false, nil
	}
	if err != nil {
		return CompanyWikiSourceDocument{}, false, err
	}
	_ = json.Unmarshal(metadata, &item.Metadata)
	if item.Metadata == nil {
		item.Metadata = map[string]any{}
	}
	return item, true, nil
}

func scanCompanyWikiSourceRevision(row companyWikiRow) (CompanyWikiSourceRevision, bool, error) {
	var item CompanyWikiSourceRevision
	var metadata []byte
	err := row.Scan(
		&item.ID, &item.DocumentID, &item.SourceRevision, &item.ContentSHA256,
		&item.Title, &item.URL, &metadata, &item.ObservedAt, &item.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return CompanyWikiSourceRevision{}, false, nil
	}
	if err != nil {
		return CompanyWikiSourceRevision{}, false, err
	}
	_ = json.Unmarshal(metadata, &item.Metadata)
	if item.Metadata == nil {
		item.Metadata = map[string]any{}
	}
	return item, true, nil
}

func scanCompanyWikiChunks(rows *sql.Rows) ([]CompanyWikiSourceChunk, error) {
	out := []CompanyWikiSourceChunk{}
	for rows.Next() {
		var item CompanyWikiSourceChunk
		var metadata []byte
		if err := rows.Scan(
			&item.ID, &item.DocumentID, &item.RevisionID, &item.ChunkIndex, &item.ChunkKind,
			&item.Content, &item.ContentSHA256, &item.NativeLocator, &item.TokenEstimate,
			&metadata, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(metadata, &item.Metadata)
		if item.Metadata == nil {
			item.Metadata = map[string]any{}
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanCompanyWikiAudit(row companyWikiRow) (CompanyWikiAuditRecord, error) {
	var item CompanyWikiAuditRecord
	var metadata []byte
	if err := row.Scan(
		&item.ID, &item.Mode, &item.Status, &item.Actor, &item.Reason,
		&item.IdempotencyKey, &item.PageID, &item.WikiRevisionID, &item.Slug,
		&item.Title, &item.ProposedPath, &item.PublishedPath, &metadata,
		&item.LastError, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return CompanyWikiAuditRecord{}, err
	}
	_ = json.Unmarshal(metadata, &item.Metadata)
	if item.Metadata == nil {
		item.Metadata = map[string]any{}
	}
	return item, nil
}

func scanCompanyWikiCitation(row companyWikiRow) (CompanyWikiCitation, error) {
	return scanCompanyWikiCitationScanner(row)
}

func scanCompanyWikiCitationScanner(row companyWikiRow) (CompanyWikiCitation, error) {
	var item CompanyWikiCitation
	if err := row.Scan(
		&item.ID, &item.WikiRevisionID, &item.ClaimKey, &item.SourceDocumentID,
		&item.SourceRevisionID, &item.ChunkID, &item.NativeLocator, &item.Quote,
		&item.CreatedAt,
	); err != nil {
		return CompanyWikiCitation{}, err
	}
	return item, nil
}

func scanCompanyWikiPageRead(row companyWikiRow) (CompanyWikiPageRead, bool, error) {
	var read CompanyWikiPageRead
	var pageMetadata, revisionMetadata, sourceRevisionIDs []byte
	var manifestPath, manifestPageID, manifestRevisionID, manifestSHA, manifestCompilerRunID, manifestRepairStatus, manifestRepairError sql.NullString
	var manifestGeneratedAt, manifestLastCheckedAt, manifestLastRepairedAt sql.NullTime
	err := row.Scan(
		&read.Page.ID, &read.Page.Slug, &read.Page.Title, &read.Page.Status,
		&read.Page.CurrentRevisionID, &pageMetadata, &read.Page.CreatedAt, &read.Page.UpdatedAt,
		&read.Revision.ID, &read.Revision.PageID, &read.Revision.RevisionNumber,
		&read.Revision.CompilerRunID, &read.Revision.Title, &read.Revision.Body,
		&read.Revision.BodySHA256, &read.Revision.Path, &sourceRevisionIDs,
		&revisionMetadata, &read.Revision.PublishedAt, &read.Revision.CreatedAt,
		&manifestPath, &manifestPageID, &manifestRevisionID, &manifestSHA,
		&manifestCompilerRunID, &manifestGeneratedAt, &manifestRepairStatus,
		&manifestRepairError, &manifestLastCheckedAt, &manifestLastRepairedAt,
	)
	if err == sql.ErrNoRows {
		return CompanyWikiPageRead{}, false, nil
	}
	if err != nil {
		return CompanyWikiPageRead{}, false, err
	}
	_ = json.Unmarshal(pageMetadata, &read.Page.Metadata)
	_ = json.Unmarshal(revisionMetadata, &read.Revision.Metadata)
	_ = json.Unmarshal(sourceRevisionIDs, &read.Revision.SourceRevisionIDs)
	if read.Page.Metadata == nil {
		read.Page.Metadata = map[string]any{}
	}
	if read.Revision.Metadata == nil {
		read.Revision.Metadata = map[string]any{}
	}
	read.Manifest = CompanyWikiManifestEntry{
		Path:            manifestPath.String,
		WikiPageID:      manifestPageID.String,
		WikiRevisionID:  manifestRevisionID.String,
		SHA256:          manifestSHA.String,
		CompilerRunID:   manifestCompilerRunID.String,
		RepairStatus:    manifestRepairStatus.String,
		LastRepairError: manifestRepairError.String,
	}
	if manifestGeneratedAt.Valid {
		read.Manifest.GeneratedAt = manifestGeneratedAt.Time
	}
	if manifestLastCheckedAt.Valid {
		read.Manifest.LastCheckedAt = manifestLastCheckedAt.Time
	}
	if manifestLastRepairedAt.Valid {
		read.Manifest.LastRepairedAt = manifestLastRepairedAt.Time
	}
	return read, true, nil
}

func scanCompanyWikiClaim(row companyWikiRow) (CompanyWikiClaim, error) {
	var item CompanyWikiClaim
	var metadata []byte
	if err := row.Scan(&item.ID, &item.WikiRevisionID, &item.ClaimKey, &item.ClaimText, &item.Confidence, &metadata, &item.CreatedAt); err != nil {
		return CompanyWikiClaim{}, err
	}
	_ = json.Unmarshal(metadata, &item.Metadata)
	if item.Metadata == nil {
		item.Metadata = map[string]any{}
	}
	return item, nil
}

func scanCompanyWikiConflict(row companyWikiRow) (CompanyWikiConflict, error) {
	var item CompanyWikiConflict
	var citationIDs, metadata []byte
	if err := row.Scan(&item.ID, &item.WikiRevisionID, &item.ClaimKey, &item.Summary, &citationIDs, &metadata, &item.CreatedAt); err != nil {
		return CompanyWikiConflict{}, err
	}
	_ = json.Unmarshal(citationIDs, &item.CitationIDs)
	_ = json.Unmarshal(metadata, &item.Metadata)
	if item.Metadata == nil {
		item.Metadata = map[string]any{}
	}
	return item, nil
}

func scanCompanyWikiCompileItem(row companyWikiRow) (CompanyWikiCompileItem, error) {
	var item CompanyWikiCompileItem
	var leaseExpiresAt sql.NullTime
	if err := row.Scan(
		&item.ID, &item.SourceRevisionID, &item.CompilerVersion, &item.SchemaVersion,
		&item.RendererVersion, &item.ModelPolicyVersion, &item.InputHash, &item.Status,
		&item.LeaseHolder, &leaseExpiresAt, &item.AttemptCount, &item.LastAttemptID,
		&item.LastError, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return CompanyWikiCompileItem{}, err
	}
	if leaseExpiresAt.Valid {
		item.LeaseExpiresAt = leaseExpiresAt.Time
	}
	return item, nil
}

func scanCompanyWikiCompileAttempt(row companyWikiRow) (CompanyWikiCompileAttempt, error) {
	var item CompanyWikiCompileAttempt
	var validationErrors, metadata []byte
	var completedAt sql.NullTime
	if err := row.Scan(
		&item.ID, &item.CompileItemID, &item.CompilerRunID, &item.Status, &item.Model,
		&item.ContextHash, &item.OutputHash, &item.RequestMetadataHash,
		&item.ResponseMetadataHash, &item.DurationMillis, &validationErrors,
		&item.LastError, &metadata, &item.CreatedAt, &completedAt,
	); err != nil {
		return CompanyWikiCompileAttempt{}, err
	}
	_ = json.Unmarshal(validationErrors, &item.ValidationErrors)
	_ = json.Unmarshal(metadata, &item.Metadata)
	if item.Metadata == nil {
		item.Metadata = map[string]any{}
	}
	if completedAt.Valid {
		item.CompletedAt = completedAt.Time
	}
	return item, nil
}

func scanCompanyWikiCompileTarget(row companyWikiRow) (CompanyWikiCompileTarget, error) {
	var item CompanyWikiCompileTarget
	if err := row.Scan(
		&item.ID, &item.CompileItemID, &item.TargetSlug, &item.TargetPath, &item.TargetType,
		&item.Status, &item.WikiRevisionID, &item.IdempotencyKey, &item.BodyHash,
		&item.LastError, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return CompanyWikiCompileTarget{}, err
	}
	return item, nil
}
