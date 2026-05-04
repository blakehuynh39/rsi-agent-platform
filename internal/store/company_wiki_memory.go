package store

import (
	"database/sql"
	"sort"
	"strings"
	"time"
)

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
	m.companyWikiManifest[input.Path] = CompanyWikiManifestEntry{
		Path:           input.Path,
		WikiPageID:     page.ID,
		WikiRevisionID: revisionID,
		SHA256:         input.SHA256,
		CompilerRunID:  input.CompilerRunID,
		GeneratedAt:    input.PublishedAt,
	}
	return CompanyWikiPagePublishResult{Page: cloneCompanyWikiPage(page), Revision: cloneCompanyWikiRevision(revision), Citations: citations}, nil
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
	citations := []CompanyWikiCitation{}
	for _, citation := range m.companyWikiCitations {
		if citation.WikiRevisionID == revision.ID {
			citations = append(citations, citation)
		}
	}
	sort.SliceStable(citations, func(i, j int) bool { return citations[i].ID < citations[j].ID })
	return CompanyWikiPageRead{
		Page:      cloneCompanyWikiPage(page),
		Revision:  cloneCompanyWikiRevision(revision),
		Citations: citations,
		Manifest:  m.companyWikiManifest[revision.Path],
	}, true, nil
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

func (m *MemoryStore) ensureCompanyWikiMemoryLocked() {
	if m.companyWikiSourceDocuments == nil {
		m.companyWikiSourceDocuments = map[string]CompanyWikiSourceDocument{}
		m.companyWikiSourceRevisions = map[string]CompanyWikiSourceRevision{}
		m.companyWikiSourceChunks = map[string]CompanyWikiSourceChunk{}
		m.companyWikiSourceChunksByRevision = map[string][]CompanyWikiSourceChunk{}
		m.companyWikiPages = map[string]CompanyWikiPage{}
		m.companyWikiPageBySlug = map[string]string{}
		m.companyWikiRevisions = map[string]CompanyWikiRevision{}
		m.companyWikiCitations = map[string]CompanyWikiCitation{}
		m.companyWikiManifest = map[string]CompanyWikiManifestEntry{}
		m.companyWikiAudits = map[string]CompanyWikiAuditRecord{}
		m.companyWikiAuditByIdempotency = map[string]string{}
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

func snippetForWikiSearch(body string) string {
	body = strings.Join(strings.Fields(body), " ")
	runes := []rune(body)
	if len(runes) > 500 {
		return string(runes[:500])
	}
	return body
}
