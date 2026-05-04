package store

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
)

func (m *MemoryStore) AcquireCompanyWikiCompilerLease(ctx context.Context, lockName string, holder string, ttl time.Duration) (func() error, bool, error) {
	_, _, _, _ = ctx, lockName, holder, ttl
	return func() error { return nil }, true, nil
}

func (m *MemoryStore) BeginCompanyWikiCompilerRun(id string, metadata map[string]any) error {
	_, _ = id, metadata
	return nil
}

func (m *MemoryStore) CompleteCompanyWikiCompilerRun(id string, status string, lastError string, metadata map[string]any) error {
	_, _, _, _ = id, status, lastError, metadata
	return nil
}

func (m *MemoryStore) UpsertCompanyWikiSourceRevision(input CompanyWikiSourceRevisionInput) (CompanyWikiSourceRevisionResult, error) {
	if err := validateCompanyWikiSourceRevisionInput(input); err != nil {
		return CompanyWikiSourceRevisionResult{}, err
	}
	input = normalizeCompanyWikiSourceRevisionInput(input)
	documentID := CompanyWikiStableID("srcdoc", input.SourceType, input.DocumentSourceKey)
	revisionID := CompanyWikiStableID("srcrev", documentID, input.SourceRevision)
	chunks := companyWikiChunksFromInput(documentID, revisionID, input)
	now := time.Now().UTC()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureCompanyWikiMemoryLocked()
	doc, existed := m.companyWikiSourceDocuments[documentID]
	if !existed {
		doc = CompanyWikiSourceDocument{
			ID:         documentID,
			SourceType: input.SourceType,
			SourceKey:  input.DocumentSourceKey,
			CreatedAt:  now,
		}
	}
	doc.SourceSessionKey = input.SourceSessionKey
	doc.Workspace = input.Workspace
	doc.Environment = input.Environment
	doc.Title = input.Title
	doc.URL = input.URL
	doc.Status = CompanyWikiSourceStatusActive
	doc.CurrentRevisionID = revisionID
	doc.Metadata = mergeAnyMaps(doc.Metadata, input.Metadata)
	doc.UpdatedAt = now
	m.companyWikiSourceDocuments[documentID] = doc

	_, revisionExisted := m.companyWikiSourceRevisions[revisionID]
	revision := CompanyWikiSourceRevision{
		ID:             revisionID,
		DocumentID:     documentID,
		SourceRevision: input.SourceRevision,
		ContentSHA256:  CompanyWikiSHA256(input.Content),
		Title:          input.Title,
		URL:            input.URL,
		Metadata:       cloneAnyMap(input.Metadata),
		ObservedAt:     input.ObservedAt,
		CreatedAt:      now,
	}
	if revisionExisted {
		revision = m.companyWikiSourceRevisions[revisionID]
		chunks = m.companyWikiSourceChunksByRevision[revisionID]
	} else {
		m.companyWikiSourceRevisions[revisionID] = revision
		m.companyWikiSourceChunksByRevision[revisionID] = cloneCompanyWikiChunks(chunks)
		for _, chunk := range chunks {
			m.companyWikiSourceChunks[chunk.ID] = chunk
		}
	}
	return CompanyWikiSourceRevisionResult{
		Document: cloneCompanyWikiSourceDocument(doc),
		Revision: cloneCompanyWikiSourceRevision(revision),
		Chunks:   cloneCompanyWikiChunks(chunks),
		Inserted: !revisionExisted,
		Changed:  !revisionExisted,
	}, nil
}

func (m *MemoryStore) ListCompanyWikiSourceChunks(documentID string) ([]CompanyWikiSourceChunk, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []CompanyWikiSourceChunk{}
	for _, chunk := range m.companyWikiSourceChunks {
		if chunk.DocumentID == strings.TrimSpace(documentID) {
			out = append(out, chunk)
		}
	}
	sortCompanyWikiChunks(out)
	return cloneCompanyWikiChunks(out), nil
}

func (m *MemoryStore) GetCompanyWikiSourceEvidence(sourceRevisionID string) (CompanyWikiSourceEvidence, bool, error) {
	sourceRevisionID = strings.TrimSpace(sourceRevisionID)
	if sourceRevisionID == "" {
		return CompanyWikiSourceEvidence{}, false, nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	revision, ok := m.companyWikiSourceRevisions[sourceRevisionID]
	if !ok {
		return CompanyWikiSourceEvidence{}, false, nil
	}
	document, ok := m.companyWikiSourceDocuments[revision.DocumentID]
	if !ok {
		return CompanyWikiSourceEvidence{}, false, nil
	}
	chunks := cloneCompanyWikiChunks(m.companyWikiSourceChunksByRevision[sourceRevisionID])
	sortCompanyWikiChunks(chunks)
	return CompanyWikiSourceEvidence{
		Document: cloneCompanyWikiSourceDocument(document),
		Revision: cloneCompanyWikiSourceRevision(revision),
		Chunks:   chunks,
	}, true, nil
}

func (m *MemoryStore) ListCompanyWikiSourceRevisionIDsWithoutCompileItem(compilerVersion string, schemaVersion string, rendererVersion string, modelPolicyVersion string, limit int) ([]string, error) {
	compilerVersion = strings.TrimSpace(compilerVersion)
	schemaVersion = strings.TrimSpace(schemaVersion)
	rendererVersion = strings.TrimSpace(rendererVersion)
	modelPolicyVersion = strings.TrimSpace(modelPolicyVersion)
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, 0, len(m.companyWikiSourceRevisions))
	for id, revision := range m.companyWikiSourceRevisions {
		key := companyWikiCompileItemKey(revision.ID, compilerVersion, schemaVersion, rendererVersion, modelPolicyVersion)
		if m.companyWikiCompileItemsByKey[key] != "" {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if len(ids) > limit {
		ids = ids[:limit]
	}
	return ids, nil
}

func (m *MemoryStore) EnqueueCompanyWikiCompileItem(input CompanyWikiCompileItemInput) (CompanyWikiCompileItem, bool, error) {
	input = normalizeCompanyWikiCompileItemInput(input)
	now := time.Now().UTC()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureCompanyWikiMemoryLocked()
	key := companyWikiCompileItemKey(input.SourceRevisionID, input.CompilerVersion, input.SchemaVersion, input.RendererVersion, input.ModelPolicyVersion)
	if existingID := m.companyWikiCompileItemsByKey[key]; existingID != "" {
		return cloneCompanyWikiCompileItem(m.companyWikiCompileItems[existingID]), false, nil
	}
	item := CompanyWikiCompileItem{
		ID:                 CompanyWikiStableID("wikicompile", key),
		SourceRevisionID:   input.SourceRevisionID,
		CompilerVersion:    input.CompilerVersion,
		SchemaVersion:      input.SchemaVersion,
		RendererVersion:    input.RendererVersion,
		ModelPolicyVersion: input.ModelPolicyVersion,
		InputHash:          input.InputHash,
		Status:             firstNonEmpty(input.Status, CompanyWikiCompileStatusPending),
		CreatedAt:          input.CreatedAt,
		UpdatedAt:          now,
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	m.companyWikiCompileItems[item.ID] = item
	m.companyWikiCompileItemsByKey[key] = item.ID
	return cloneCompanyWikiCompileItem(item), true, nil
}

func (m *MemoryStore) ClaimCompanyWikiCompileItems(input CompanyWikiCompileClaimInput) ([]CompanyWikiCompileItem, error) {
	input = normalizeCompanyWikiCompileClaimInput(input)
	now := time.Now().UTC()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureCompanyWikiMemoryLocked()
	items := make([]CompanyWikiCompileItem, 0, input.Limit)
	keys := make([]string, 0, len(m.companyWikiCompileItems))
	for id := range m.companyWikiCompileItems {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	for _, id := range keys {
		item := m.companyWikiCompileItems[id]
		if item.CompilerVersion != input.CompilerVersion || item.SchemaVersion != input.SchemaVersion || item.RendererVersion != input.RendererVersion || item.ModelPolicyVersion != input.ModelPolicyVersion {
			continue
		}
		if item.Status != CompanyWikiCompileStatusPending && item.Status != CompanyWikiCompileStatusFailed {
			continue
		}
		if item.Status == CompanyWikiCompileStatusFailed && item.AttemptCount >= input.MaxAttempts {
			continue
		}
		if !item.LeaseExpiresAt.IsZero() && item.LeaseExpiresAt.After(now) && item.LeaseHolder != input.LeaseHolder {
			continue
		}
		item.Status = CompanyWikiCompileStatusClaimed
		item.LeaseHolder = input.LeaseHolder
		item.LeaseExpiresAt = now.Add(input.LeaseDuration)
		item.AttemptCount++
		item.UpdatedAt = now
		m.companyWikiCompileItems[id] = item
		items = append(items, cloneCompanyWikiCompileItem(item))
		if len(items) >= input.Limit {
			break
		}
	}
	return items, nil
}

func (m *MemoryStore) BeginCompanyWikiCompileAttempt(input CompanyWikiCompileAttemptInput) (CompanyWikiCompileAttempt, error) {
	input = normalizeCompanyWikiCompileAttemptInput(input)
	now := time.Now().UTC()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureCompanyWikiMemoryLocked()
	attempt := CompanyWikiCompileAttempt{
		ID:                   CompanyWikiStableID("wikiattempt", input.CompileItemID, input.CompilerRunID, input.ContextHash, now.Format(time.RFC3339Nano)),
		CompileItemID:        input.CompileItemID,
		CompilerRunID:        input.CompilerRunID,
		Status:               firstNonEmpty(input.Status, CompanyWikiCompileStatusClaimed),
		Model:                input.Model,
		ContextHash:          input.ContextHash,
		OutputHash:           input.OutputHash,
		RequestMetadataHash:  input.RequestMetadataHash,
		ResponseMetadataHash: input.ResponseMetadataHash,
		DurationMillis:       input.DurationMillis,
		ValidationErrors:     append([]string(nil), input.ValidationErrors...),
		LastError:            input.LastError,
		Metadata:             cloneAnyMap(input.Metadata),
		CreatedAt:            now,
	}
	m.companyWikiCompileAttempts[attempt.ID] = attempt
	item := m.companyWikiCompileItems[input.CompileItemID]
	item.LastAttemptID = attempt.ID
	item.UpdatedAt = now
	m.companyWikiCompileItems[item.ID] = item
	return cloneCompanyWikiCompileAttempt(attempt), nil
}

func (m *MemoryStore) CompleteCompanyWikiCompileAttempt(attemptID string, status string, outputHash string, durationMillis int64, validationErrors []string, lastError string, metadata map[string]any) (CompanyWikiCompileAttempt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	attempt, ok := m.companyWikiCompileAttempts[strings.TrimSpace(attemptID)]
	if !ok {
		return CompanyWikiCompileAttempt{}, sql.ErrNoRows
	}
	attempt.Status = firstNonEmpty(strings.TrimSpace(status), CompanyWikiCompileStatusCompleted)
	attempt.OutputHash = strings.TrimSpace(outputHash)
	if requestHash := stringFromAnyMap(metadata, "request_metadata_hash"); requestHash != "" {
		attempt.RequestMetadataHash = requestHash
	}
	if responseHash := stringFromAnyMap(metadata, "response_metadata_hash"); responseHash != "" {
		attempt.ResponseMetadataHash = responseHash
	}
	attempt.DurationMillis = durationMillis
	attempt.ValidationErrors = append([]string(nil), validationErrors...)
	attempt.LastError = strings.TrimSpace(lastError)
	attempt.Metadata = mergeAnyMaps(attempt.Metadata, metadata)
	attempt.CompletedAt = time.Now().UTC()
	m.companyWikiCompileAttempts[attempt.ID] = attempt
	return cloneCompanyWikiCompileAttempt(attempt), nil
}

func (m *MemoryStore) UpsertCompanyWikiCompileTargets(compileItemID string, targets []CompanyWikiCompileTargetInput) ([]CompanyWikiCompileTarget, error) {
	compileItemID = strings.TrimSpace(compileItemID)
	now := time.Now().UTC()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureCompanyWikiMemoryLocked()
	desired := map[string]struct{}{}
	out := []CompanyWikiCompileTarget{}
	for _, input := range targets {
		input.CompileItemID = compileItemID
		input = normalizeCompanyWikiCompileTargetInput(input)
		desired[input.TargetSlug] = struct{}{}
		id := CompanyWikiStableID("wikitarget", compileItemID, input.TargetSlug)
		item := m.companyWikiCompileTargets[id]
		if item.ID == "" {
			item = CompanyWikiCompileTarget{ID: id, CompileItemID: compileItemID, CreatedAt: now}
		}
		item.TargetSlug = input.TargetSlug
		item.TargetPath = input.TargetPath
		item.TargetType = input.TargetType
		newBodyHash := strings.TrimSpace(input.BodyHash)
		if item.Status == CompanyWikiCompileTargetStatusPublished && item.BodyHash == newBodyHash {
			// Preserve published status if body hasn't changed
		} else {
			item.Status = firstNonEmpty(input.Status, item.Status, CompanyWikiCompileTargetStatusPending)
		}
		if input.WikiRevisionID != "" {
			item.WikiRevisionID = strings.TrimSpace(input.WikiRevisionID)
		}
		item.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
		item.BodyHash = newBodyHash
		item.LastError = strings.TrimSpace(input.LastError)
		item.UpdatedAt = now
		m.companyWikiCompileTargets[id] = item
		out = append(out, cloneCompanyWikiCompileTarget(item))
	}
	for id, item := range m.companyWikiCompileTargets {
		if item.CompileItemID != compileItemID {
			continue
		}
		if _, ok := desired[item.TargetSlug]; !ok {
			item.Status = CompanyWikiCompileTargetStatusSuperseded
			item.UpdatedAt = now
			m.companyWikiCompileTargets[id] = item
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].TargetSlug < out[j].TargetSlug })
	return out, nil
}

func (m *MemoryStore) UpdateCompanyWikiCompileTarget(input CompanyWikiCompileTargetInput) (CompanyWikiCompileTarget, error) {
	input = normalizeCompanyWikiCompileTargetInput(input)
	id := CompanyWikiStableID("wikitarget", input.CompileItemID, input.TargetSlug)
	m.mu.Lock()
	defer m.mu.Unlock()
	item := m.companyWikiCompileTargets[id]
	if item.ID == "" {
		item = CompanyWikiCompileTarget{ID: id, CompileItemID: input.CompileItemID, TargetSlug: input.TargetSlug, TargetPath: input.TargetPath, TargetType: input.TargetType, CreatedAt: time.Now().UTC()}
	}
	item.Status = firstNonEmpty(input.Status, item.Status, CompanyWikiCompileTargetStatusPending)
	item.WikiRevisionID = strings.TrimSpace(input.WikiRevisionID)
	item.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	item.BodyHash = strings.TrimSpace(input.BodyHash)
	item.LastError = strings.TrimSpace(input.LastError)
	item.UpdatedAt = time.Now().UTC()
	m.companyWikiCompileTargets[id] = item
	return cloneCompanyWikiCompileTarget(item), nil
}

func (m *MemoryStore) ListCompanyWikiCompileTargets(compileItemID string) ([]CompanyWikiCompileTarget, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []CompanyWikiCompileTarget{}
	for _, item := range m.companyWikiCompileTargets {
		if item.CompileItemID == strings.TrimSpace(compileItemID) {
			out = append(out, cloneCompanyWikiCompileTarget(item))
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].TargetSlug < out[j].TargetSlug })
	return out, nil
}

func (m *MemoryStore) CompleteCompanyWikiCompileItem(compileItemID string, status string, lastError string) (CompanyWikiCompileItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.companyWikiCompileItems[strings.TrimSpace(compileItemID)]
	if !ok {
		return CompanyWikiCompileItem{}, sql.ErrNoRows
	}
	item.Status = firstNonEmpty(strings.TrimSpace(status), CompanyWikiCompileStatusCompleted)
	item.LastError = strings.TrimSpace(lastError)
	if item.Status == CompanyWikiCompileStatusCompleted || item.Status == CompanyWikiCompileStatusSkipped {
		item.LeaseHolder = ""
		item.LeaseExpiresAt = time.Time{}
	}
	item.UpdatedAt = time.Now().UTC()
	m.companyWikiCompileItems[item.ID] = item
	return cloneCompanyWikiCompileItem(item), nil
}

func (m *MemoryStore) BeginCompanyWikiAudit(input CompanyWikiAuditInput) (CompanyWikiAuditRecord, error) {
	input = normalizeCompanyWikiAuditInput(input)
	if input.ID == "" {
		input.ID = CompanyWikiStableID("wikiaudit", input.Mode, input.IdempotencyKey, input.Slug, input.Reason)
	}
	now := time.Now().UTC()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureCompanyWikiMemoryLocked()
	if input.IdempotencyKey != "" {
		key := input.Mode + "\x00" + input.IdempotencyKey
		if existingID := m.companyWikiAuditByIdempotency[key]; existingID != "" {
			return cloneCompanyWikiAudit(m.companyWikiAudits[existingID]), nil
		}
		m.companyWikiAuditByIdempotency[key] = input.ID
	}
	record := CompanyWikiAuditRecord{
		ID:             input.ID,
		Mode:           input.Mode,
		Status:         CompanyWikiAuditStatusIntent,
		Actor:          input.Actor,
		Reason:         input.Reason,
		IdempotencyKey: input.IdempotencyKey,
		PageID:         input.PageID,
		Slug:           input.Slug,
		Title:          input.Title,
		ProposedPath:   input.ProposedPath,
		Metadata:       cloneAnyMap(input.Metadata),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	m.companyWikiAudits[record.ID] = record
	return cloneCompanyWikiAudit(record), nil
}

func (m *MemoryStore) CompleteCompanyWikiAudit(auditID string, wikiRevisionID string, publishedPath string, metadata map[string]any) (CompanyWikiAuditRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	record, ok := m.companyWikiAudits[strings.TrimSpace(auditID)]
	if !ok {
		return CompanyWikiAuditRecord{}, sql.ErrNoRows
	}
	record.Status = CompanyWikiAuditStatusPublished
	record.WikiRevisionID = strings.TrimSpace(wikiRevisionID)
	record.PublishedPath = strings.TrimSpace(publishedPath)
	record.Metadata = mergeAnyMaps(record.Metadata, metadata)
	record.LastError = ""
	record.UpdatedAt = time.Now().UTC()
	m.companyWikiAudits[record.ID] = record
	return cloneCompanyWikiAudit(record), nil
}

func (m *MemoryStore) FailCompanyWikiAudit(auditID string, lastError string, metadata map[string]any) (CompanyWikiAuditRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	record, ok := m.companyWikiAudits[strings.TrimSpace(auditID)]
	if !ok {
		return CompanyWikiAuditRecord{}, sql.ErrNoRows
	}
	record.Status = CompanyWikiAuditStatusFailed
	record.LastError = strings.TrimSpace(lastError)
	record.Metadata = mergeAnyMaps(record.Metadata, metadata)
	record.UpdatedAt = time.Now().UTC()
	m.companyWikiAudits[record.ID] = record
	return cloneCompanyWikiAudit(record), nil
}

func (m *MemoryStore) PublishCompanyWikiPage(input CompanyWikiPagePublishInput) (CompanyWikiPagePublishResult, error) {
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
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureCompanyWikiMemoryLocked()
	page := m.companyWikiPages[input.PageID]
	if page.ID == "" {
		page = CompanyWikiPage{ID: input.PageID, Slug: input.Slug, CreatedAt: input.PublishedAt}
	}
	page.Title = input.Title
	page.Status = CompanyWikiPageStatusPublished
	page.CurrentRevisionID = revisionID
	page.Metadata = mergeAnyMaps(page.Metadata, input.Metadata)
	page.UpdatedAt = input.PublishedAt
	m.companyWikiPages[page.ID] = page
	m.companyWikiPageBySlug[page.Slug] = page.ID

	revisionNumber := 1
	for _, rev := range m.companyWikiRevisions {
		if rev.PageID == page.ID && rev.RevisionNumber >= revisionNumber {
			revisionNumber = rev.RevisionNumber + 1
		}
	}
	revision := CompanyWikiRevision{
		ID:                revisionID,
		PageID:            page.ID,
		RevisionNumber:    revisionNumber,
		CompilerRunID:     input.CompilerRunID,
		Title:             input.Title,
		Body:              input.Body,
		BodySHA256:        input.SHA256,
		Path:              input.Path,
		SourceRevisionIDs: append([]string(nil), input.SourceRevisionIDs...),
		Metadata:          cloneAnyMap(input.Metadata),
		PublishedAt:       input.PublishedAt,
		CreatedAt:         input.PublishedAt,
	}
	if existing := m.companyWikiRevisions[revisionID]; existing.ID != "" {
		revision = existing
	}
	m.companyWikiRevisions[revisionID] = revision
	citations := make([]CompanyWikiCitation, 0, len(input.Citations))
	for _, citation := range input.Citations {
		item := CompanyWikiCitation{
			ID:               CompanyWikiStableID("wikicit", revisionID, citation.SourceDocumentID, citation.SourceRevisionID, citation.ChunkID, citation.ClaimKey),
			WikiRevisionID:   revisionID,
			ClaimKey:         citation.ClaimKey,
			SourceDocumentID: citation.SourceDocumentID,
			SourceRevisionID: citation.SourceRevisionID,
			ChunkID:          citation.ChunkID,
			NativeLocator:    citation.NativeLocator,
			Quote:            citation.Quote,
			CreatedAt:        input.PublishedAt,
		}
		m.companyWikiCitations[item.ID] = item
		citations = append(citations, item)
	}
	claims := make([]CompanyWikiClaim, 0, len(input.Claims))
	for _, claim := range input.Claims {
		item := CompanyWikiClaim{
			ID:             CompanyWikiStableID("wikiclaim", revisionID, claim.ClaimKey, claim.ClaimText),
			WikiRevisionID: revisionID,
			ClaimKey:       strings.TrimSpace(claim.ClaimKey),
			ClaimText:      strings.TrimSpace(claim.ClaimText),
			Confidence:     claim.Confidence,
			Metadata:       cloneAnyMap(claim.Metadata),
			CreatedAt:      input.PublishedAt,
		}
		if item.Confidence == 0 {
			item.Confidence = 1
		}
		m.companyWikiClaims[item.ID] = item
		claims = append(claims, cloneCompanyWikiClaim(item))
	}
	conflicts := make([]CompanyWikiConflict, 0, len(input.Conflicts))
	for _, conflict := range input.Conflicts {
		item := CompanyWikiConflict{
			ID:             CompanyWikiStableID("wikiconflict", revisionID, conflict.ClaimKey, conflict.Summary),
			WikiRevisionID: revisionID,
			ClaimKey:       strings.TrimSpace(conflict.ClaimKey),
			Summary:        strings.TrimSpace(conflict.Summary),
			CitationIDs:    append([]string(nil), conflict.Citations...),
			Metadata:       cloneAnyMap(conflict.Metadata),
			CreatedAt:      input.PublishedAt,
		}
		m.companyWikiConflicts[item.ID] = item
		conflicts = append(conflicts, cloneCompanyWikiConflict(item))
	}
	m.companyWikiManifest[input.Path] = CompanyWikiManifestEntry{
		Path:           input.Path,
		WikiPageID:     page.ID,
		WikiRevisionID: revisionID,
		SHA256:         input.SHA256,
		CompilerRunID:  input.CompilerRunID,
		GeneratedAt:    input.PublishedAt,
		RepairStatus:   CompanyWikiManifestRepairOK,
	}
	return CompanyWikiPagePublishResult{Page: cloneCompanyWikiPage(page), Revision: cloneCompanyWikiRevision(revision), Citations: citations, Claims: claims, Conflicts: conflicts}, nil
}

func (m *MemoryStore) SearchCompanyWikiPages(query string, limit int) ([]CompanyWikiSearchResult, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []CompanyWikiSearchResult{}
	for _, page := range m.companyWikiPages {
		rev := m.companyWikiRevisions[page.CurrentRevisionID]
		if query != "" && !strings.Contains(strings.ToLower(page.Slug+" "+page.Title+" "+rev.Body), query) {
			continue
		}
		out = append(out, CompanyWikiSearchResult{
			PageID: page.ID, Slug: page.Slug, Title: page.Title, Path: rev.Path,
			WikiRevisionID: rev.ID, SHA256: rev.BodySHA256, Snippet: snippetForWikiSearch(rev.Body), PublishedAt: rev.PublishedAt,
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PublishedAt.After(out[j].PublishedAt) })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *MemoryStore) GetCompanyWikiPage(ref string) (CompanyWikiPageRead, bool, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return CompanyWikiPageRead{}, false, nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	page, ok := m.companyWikiPages[ref]
	if !ok {
		if pageID := m.companyWikiPageBySlug[ref]; pageID != "" {
			page, ok = m.companyWikiPages[pageID]
		}
	}
	if !ok {
		return CompanyWikiPageRead{}, false, nil
	}
	revision := m.companyWikiRevisions[page.CurrentRevisionID]
	return m.companyWikiPageReadLocked(page, revision), true, nil
}

func (m *MemoryStore) companyWikiPageReadLocked(page CompanyWikiPage, revision CompanyWikiRevision) CompanyWikiPageRead {
	citations := []CompanyWikiCitation{}
	for _, citation := range m.companyWikiCitations {
		if citation.WikiRevisionID == revision.ID {
			citations = append(citations, citation)
		}
	}
	sort.SliceStable(citations, func(i, j int) bool { return citations[i].ID < citations[j].ID })
	claims := []CompanyWikiClaim{}
	for _, claim := range m.companyWikiClaims {
		if claim.WikiRevisionID == revision.ID {
			claims = append(claims, cloneCompanyWikiClaim(claim))
		}
	}
	sort.SliceStable(claims, func(i, j int) bool { return claims[i].ID < claims[j].ID })
	conflicts := []CompanyWikiConflict{}
	for _, conflict := range m.companyWikiConflicts {
		if conflict.WikiRevisionID == revision.ID {
			conflicts = append(conflicts, cloneCompanyWikiConflict(conflict))
		}
	}
	sort.SliceStable(conflicts, func(i, j int) bool { return conflicts[i].ID < conflicts[j].ID })
	return CompanyWikiPageRead{
		Page:      cloneCompanyWikiPage(page),
		Revision:  cloneCompanyWikiRevision(revision),
		Citations: citations,
		Claims:    claims,
		Conflicts: conflicts,
		Manifest:  m.companyWikiManifest[revision.Path],
	}
}

func (m *MemoryStore) ListCompanyWikiManifestEntries() ([]CompanyWikiManifestEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]CompanyWikiManifestEntry, 0, len(m.companyWikiManifest))
	for _, item := range m.companyWikiManifest {
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func (m *MemoryStore) ListCompanyWikiCandidatePages(query CompanyWikiPageQuery) ([]CompanyWikiPageRead, error) {
	query.Query = strings.ToLower(strings.TrimSpace(query.Query))
	if query.Limit <= 0 || query.Limit > 50 {
		query.Limit = 10
	}
	types := stringSet(query.Types)
	tags := stringSet(query.Tags)
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []CompanyWikiPageRead{}
	for _, page := range m.companyWikiPages {
		revision := m.companyWikiRevisions[page.CurrentRevisionID]
		if query.ExcludeEvidence && strings.HasPrefix(revision.Path, "sources/") {
			continue
		}
		pageType := strings.TrimSpace(stringFromAnyMap(page.Metadata, "type"))
		if pageType == "" {
			pageType = strings.TrimSpace(stringFromAnyMap(revision.Metadata, "type"))
		}
		if len(types) > 0 {
			if _, ok := types[pageType]; !ok {
				continue
			}
		}
		if len(tags) > 0 && !anyStringOverlap(tags, stringsFromAnyMap(page.Metadata, "tags"), stringsFromAnyMap(revision.Metadata, "tags")) {
			continue
		}
		if len(query.SourceRevisionIDs) > 0 && !anySliceOverlap(stringSet(query.SourceRevisionIDs), revision.SourceRevisionIDs) {
			continue
		}
		haystack := strings.ToLower(page.Slug + " " + page.Title + " " + revision.Path + " " + revision.Body)
		if query.Query != "" && !strings.Contains(haystack, query.Query) {
			continue
		}
		out = append(out, m.companyWikiPageReadLocked(page, revision))
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Revision.Path < out[j].Revision.Path })
	if len(out) > query.Limit {
		out = out[:query.Limit]
	}
	return out, nil
}

func (m *MemoryStore) UpdateCompanyWikiManifestRepair(path string, status string, lastError string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	entry := m.companyWikiManifest[path]
	if entry.Path == "" {
		entry.Path = path
	}
	entry.RepairStatus = firstNonEmpty(strings.TrimSpace(status), CompanyWikiManifestRepairOK)
	entry.LastRepairError = strings.TrimSpace(lastError)
	entry.LastCheckedAt = time.Now().UTC()
	if entry.RepairStatus == CompanyWikiManifestRepairOK {
		entry.LastRepairedAt = entry.LastCheckedAt
	}
	m.companyWikiManifest[path] = entry
	return nil
}

func (m *MemoryStore) ensureCompanyWikiMemoryLocked() {
	if m.companyWikiSourceDocuments == nil {
		m.companyWikiSourceDocuments = map[string]CompanyWikiSourceDocument{}
	}
	if m.companyWikiSourceRevisions == nil {
		m.companyWikiSourceRevisions = map[string]CompanyWikiSourceRevision{}
	}
	if m.companyWikiSourceChunks == nil {
		m.companyWikiSourceChunks = map[string]CompanyWikiSourceChunk{}
	}
	if m.companyWikiSourceChunksByRevision == nil {
		m.companyWikiSourceChunksByRevision = map[string][]CompanyWikiSourceChunk{}
	}
	if m.companyWikiPages == nil {
		m.companyWikiPages = map[string]CompanyWikiPage{}
	}
	if m.companyWikiPageBySlug == nil {
		m.companyWikiPageBySlug = map[string]string{}
	}
	if m.companyWikiRevisions == nil {
		m.companyWikiRevisions = map[string]CompanyWikiRevision{}
	}
	if m.companyWikiCitations == nil {
		m.companyWikiCitations = map[string]CompanyWikiCitation{}
	}
	if m.companyWikiClaims == nil {
		m.companyWikiClaims = map[string]CompanyWikiClaim{}
	}
	if m.companyWikiConflicts == nil {
		m.companyWikiConflicts = map[string]CompanyWikiConflict{}
	}
	if m.companyWikiManifest == nil {
		m.companyWikiManifest = map[string]CompanyWikiManifestEntry{}
	}
	if m.companyWikiAudits == nil {
		m.companyWikiAudits = map[string]CompanyWikiAuditRecord{}
	}
	if m.companyWikiAuditByIdempotency == nil {
		m.companyWikiAuditByIdempotency = map[string]string{}
	}
	if m.companyWikiCompileItems == nil {
		m.companyWikiCompileItems = map[string]CompanyWikiCompileItem{}
	}
	if m.companyWikiCompileItemsByKey == nil {
		m.companyWikiCompileItemsByKey = map[string]string{}
	}
	if m.companyWikiCompileAttempts == nil {
		m.companyWikiCompileAttempts = map[string]CompanyWikiCompileAttempt{}
	}
	if m.companyWikiCompileTargets == nil {
		m.companyWikiCompileTargets = map[string]CompanyWikiCompileTarget{}
	}
}

func cloneCompanyWikiSourceDocument(item CompanyWikiSourceDocument) CompanyWikiSourceDocument {
	item.Metadata = cloneAnyMap(item.Metadata)
	return item
}

func cloneCompanyWikiSourceRevision(item CompanyWikiSourceRevision) CompanyWikiSourceRevision {
	item.Metadata = cloneAnyMap(item.Metadata)
	return item
}

func cloneCompanyWikiPage(item CompanyWikiPage) CompanyWikiPage {
	item.Metadata = cloneAnyMap(item.Metadata)
	return item
}

func cloneCompanyWikiRevision(item CompanyWikiRevision) CompanyWikiRevision {
	item.Metadata = cloneAnyMap(item.Metadata)
	item.SourceRevisionIDs = append([]string(nil), item.SourceRevisionIDs...)
	return item
}

func cloneCompanyWikiAudit(item CompanyWikiAuditRecord) CompanyWikiAuditRecord {
	item.Metadata = cloneAnyMap(item.Metadata)
	return item
}

func cloneCompanyWikiClaim(item CompanyWikiClaim) CompanyWikiClaim {
	item.Metadata = cloneAnyMap(item.Metadata)
	return item
}

func cloneCompanyWikiConflict(item CompanyWikiConflict) CompanyWikiConflict {
	item.CitationIDs = append([]string(nil), item.CitationIDs...)
	item.Metadata = cloneAnyMap(item.Metadata)
	return item
}

func cloneCompanyWikiCompileItem(item CompanyWikiCompileItem) CompanyWikiCompileItem {
	return item
}

func cloneCompanyWikiCompileAttempt(item CompanyWikiCompileAttempt) CompanyWikiCompileAttempt {
	item.ValidationErrors = append([]string(nil), item.ValidationErrors...)
	item.Metadata = cloneAnyMap(item.Metadata)
	return item
}

func cloneCompanyWikiCompileTarget(item CompanyWikiCompileTarget) CompanyWikiCompileTarget {
	return item
}

func snippetForWikiSearch(body string) string {
	body = strings.Join(strings.Fields(body), " ")
	runes := []rune(body)
	if len(runes) > 500 {
		return string(runes[:500])
	}
	return body
}

func stringSet(values []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out[value] = struct{}{}
		}
	}
	return out
}

func anySliceOverlap(set map[string]struct{}, values []string) bool {
	for _, value := range values {
		if _, ok := set[strings.TrimSpace(value)]; ok {
			return true
		}
	}
	return false
}

func anyStringOverlap(set map[string]struct{}, groups ...[]string) bool {
	for _, group := range groups {
		if anySliceOverlap(set, group) {
			return true
		}
	}
	return false
}

func stringFromAnyMap(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	if raw, ok := values[key].(string); ok {
		return strings.TrimSpace(raw)
	}
	return ""
}

func stringsFromAnyMap(values map[string]any, key string) []string {
	raw, ok := values[key]
	if !ok {
		return nil
	}
	switch value := raw.(type) {
	case []string:
		return append([]string(nil), value...)
	case []any:
		out := []string{}
		for _, item := range value {
			if text := strings.TrimSpace(fmt.Sprint(item)); text != "" {
				out = append(out, text)
			}
		}
		return out
	case string:
		if strings.TrimSpace(value) == "" {
			return nil
		}
		return []string{strings.TrimSpace(value)}
	default:
		return nil
	}
}
