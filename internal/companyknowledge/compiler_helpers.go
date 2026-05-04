package companyknowledge

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/store"
)

const companyWikiSchemaBody = `# Company Wiki Schema

This file is generated deterministically by Platform. LLM compiler calls may propose structured page data, claims, citations, conflicts, owners, and open questions, but they cannot change this schema.

## Page Roots

- ` + "`pages/`" + ` contains synthesized company wiki pages. These are the primary files Hermes should read first.
- ` + "`sources/`" + ` contains source evidence pages when evidence mode is enabled. These are audit material, not the canonical synthesis.
- ` + "`index.md`" + ` is the generated catalog.
- ` + "`log.md`" + ` is the generated publish and repair timeline.

## Synthesis Pages

Every factual bullet is rendered from a validated claim object. Every claim must cite existing source chunks from Platform Postgres. Conflict sections preserve cited disagreement rather than overwriting it.
`

const companyWikiCloseSourceTimestampWindow = 24 * time.Hour

func WriteSchemaFile(root string) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil
	}
	_, err := PublishMarkdownFile(root, "SCHEMA.md", companyWikiSchemaBody)
	return err
}

func StageMarkdownFile(root string, relativePath string, body string) (string, string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", "", errors.New("company wiki root is required")
	}
	relativePath = cleanRelativeWikiPath(relativePath)
	if relativePath == "" {
		return "", "", errors.New("wiki relative path is required")
	}
	stageDir := filepath.Join(root, ".staging", filepath.Dir(relativePath))
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return "", "", err
	}
	stagePath := filepath.Join(stageDir, filepath.Base(relativePath)+fmt.Sprintf(".%d.tmp", time.Now().UnixNano()))
	if err := os.WriteFile(stagePath, []byte(body), 0o644); err != nil {
		return "", "", err
	}
	if err := fsyncFile(stagePath); err != nil {
		_ = os.Remove(stagePath)
		return "", "", err
	}
	return stagePath, store.CompanyWikiSHA256(body), nil
}

func CommitStagedMarkdownFile(stagePath string, root string, relativePath string) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return errors.New("company wiki root is required")
	}
	relativePath = cleanRelativeWikiPath(relativePath)
	if relativePath == "" {
		return errors.New("wiki relative path is required")
	}
	target := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	if err := os.Rename(stagePath, target); err != nil {
		return err
	}
	return fsyncDir(filepath.Dir(target))
}

func validateCompilerSourcePolicy(document store.CompanyWikiSourceDocument, revision store.CompanyWikiSourceRevision) error {
	return ValidateCompanyWikiSourcePolicy(document, revision)
}

func ValidateCompanyWikiSourcePolicy(document store.CompanyWikiSourceDocument, revision store.CompanyWikiSourceRevision) error {
	metadata := storeMergeMetadata(document.Metadata, revision.Metadata)
	switch strings.TrimSpace(document.SourceType) {
	case SlackMessageSourceType:
		if boolFromMetadata(metadata, "mirror_denied") {
			return fmt.Errorf("slack source %s is denied by mirror policy", document.SourceKey)
		}
		if boolFromMetadata(metadata, "channel_im") {
			return fmt.Errorf("slack source %s is a DM/IM and is not allowed for V1 wiki synthesis", document.SourceKey)
		}
		if boolFromMetadata(metadata, "channel_private") {
			return fmt.Errorf("slack source %s is private and is not allowed for V1 wiki synthesis", document.SourceKey)
		}
		if allowed, ok := metadata["mirror_allowed"]; ok && !boolFromAny(allowed) {
			return fmt.Errorf("slack source %s is not explicitly allowed by mirror policy", document.SourceKey)
		}
	case NotionDocumentSourceType:
		if allowed, ok := metadata["notion_allowlisted"]; ok && !boolFromAny(allowed) {
			return fmt.Errorf("notion source %s is outside the configured allowlist", document.SourceKey)
		}
	}
	return nil
}

func compilerContextHash(evidence store.CompanyWikiSourceEvidence, chunks []store.CompanyWikiSourceChunk, candidates []store.CompanyWikiPageRead) string {
	parts := []string{
		evidence.Document.ID,
		evidence.Revision.ID,
		evidence.Revision.ContentSHA256,
		CompanyWikiCompilerVersion,
		CompanyWikiSchemaVersion,
		CompanyWikiRendererVersion,
		CompanyWikiModelPolicyVersion,
	}
	for _, chunk := range chunks {
		parts = append(parts, chunk.ID, chunk.ContentSHA256)
	}
	candidatesCopy := make([]store.CompanyWikiPageRead, len(candidates))
	copy(candidatesCopy, candidates)
	sort.SliceStable(candidatesCopy, func(i, j int) bool { return candidatesCopy[i].Page.Slug < candidatesCopy[j].Page.Slug })
	for _, page := range candidatesCopy {
		parts = append(parts, page.Page.Slug, page.Revision.ID, page.Revision.BodySHA256)
	}
	return store.CompanyWikiSHA256(strings.Join(parts, "\x00"))
}

func mustMarshalString(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(raw)
}

func preserveCandidateClaimsInSynthesisOutput(output WikiSynthesisOutput, evidence store.CompanyWikiSourceEvidence, candidates []store.CompanyWikiPageRead) WikiSynthesisOutput {
	candidatesBySlug := map[string]store.CompanyWikiPageRead{}
	for _, candidate := range candidates {
		slug := store.NormalizeCompanyWikiSlug(candidate.Page.Slug)
		if slug != "" {
			candidatesBySlug[slug] = candidate
		}
	}
	revisionTimestamps := buildRevisionTimestampMap(evidence, candidates)
	for pageIndex := range output.Pages {
		page := output.Pages[pageIndex]
		candidate, ok := candidatesBySlug[synthesisSlug(page)]
		if !ok {
			continue
		}
		claimIndexes := map[string]int{}
		for claimIndex, claim := range page.Claims {
			if key := strings.TrimSpace(claim.ClaimKey); key != "" {
				claimIndexes[key] = claimIndex
			}
		}
		for _, claim := range candidate.Claims {
			claimKey := strings.TrimSpace(claim.ClaimKey)
			if claimKey == "" {
				continue
			}
			citations := candidateCitationInputsForClaim(candidate, claimKey)
			if len(citations) == 0 {
				continue
			}
			if existingIndex, ok := claimIndexes[claimKey]; ok {
				existingClaim := output.Pages[pageIndex].Claims[existingIndex]
				if candidateClaimMateriallyFresher(existingClaim.Citations, citations, revisionTimestamps) {
					output.Pages[pageIndex].Claims[existingIndex] = WikiSynthesisClaim{
						ClaimKey:   claimKey,
						Text:       claim.ClaimText,
						Confidence: claim.Confidence,
						Citations:  citations,
					}
				}
				continue
			}
			output.Pages[pageIndex].Claims = append(output.Pages[pageIndex].Claims, WikiSynthesisClaim{
				ClaimKey:   claimKey,
				Text:       claim.ClaimText,
				Confidence: claim.Confidence,
				Citations:  citations,
			})
			claimIndexes[claimKey] = len(output.Pages[pageIndex].Claims) - 1
		}
	}
	return output
}

func validateSynthesisOutput(evidence store.CompanyWikiSourceEvidence, candidates []store.CompanyWikiPageRead, output WikiSynthesisOutput) ([]WikiSynthesisPage, []string) {
	errorsOut := []string{}
	pages := make([]WikiSynthesisPage, 0, len(output.Pages))
	allowedCitations := allowedSynthesisCitations(evidence, candidates)
	if len(output.Pages) == 0 {
		return nil, []string{"synthesis output must include at least one page"}
	}
	for pageIndex, page := range output.Pages {
		page.Title = strings.TrimSpace(page.Title)
		page.Type = normalizeSynthesisPageType(page.Type)
		page.Slug = store.NormalizeCompanyWikiSlug(firstNonEmpty(page.Slug, page.Title))
		page.Summary = strings.TrimSpace(page.Summary)
		page.Tags = normalizeStringList(page.Tags)
		page.Owners = normalizeStringList(page.Owners)
		page.RelatedPages = normalizeStringList(page.RelatedPages)
		page.OpenQuestions = normalizeStringList(page.OpenQuestions)
		if page.Title == "" {
			errorsOut = append(errorsOut, fmt.Sprintf("page[%d].title is required", pageIndex))
		}
		if page.Summary == "" {
			errorsOut = append(errorsOut, fmt.Sprintf("page[%d].summary is required", pageIndex))
		}
		if len(page.Claims) == 0 {
			errorsOut = append(errorsOut, fmt.Sprintf("page[%d] must include at least one cited claim", pageIndex))
		}
		claimKeys := map[string]struct{}{}
		for claimIndex, claim := range page.Claims {
			claim.ClaimKey = strings.TrimSpace(claim.ClaimKey)
			claim.Text = strings.TrimSpace(claim.Text)
			if claim.ClaimKey == "" {
				claim.ClaimKey = fmt.Sprintf("claim_%d_%d", pageIndex+1, claimIndex+1)
			}
			if claim.Text == "" {
				errorsOut = append(errorsOut, fmt.Sprintf("page[%d].claims[%d].text is required", pageIndex, claimIndex))
			}
			if claim.Confidence <= 0 || claim.Confidence > 1 {
				claim.Confidence = 1
			}
			if len(claim.Citations) == 0 {
				errorsOut = append(errorsOut, fmt.Sprintf("page[%d].claims[%d] must cite at least one source chunk", pageIndex, claimIndex))
			}
			for citationIndex, citation := range claim.Citations {
				citation.SourceDocumentID = strings.TrimSpace(citation.SourceDocumentID)
				citation.SourceRevisionID = strings.TrimSpace(citation.SourceRevisionID)
				citation.ChunkID = strings.TrimSpace(citation.ChunkID)
				allowed, ok := allowedCitations[citationKey(citation)]
				if !ok {
					errorsOut = append(errorsOut, fmt.Sprintf("page[%d].claims[%d].citations[%d] references unknown chunk %q", pageIndex, claimIndex, citationIndex, citation.ChunkID))
					continue
				}
				citation.ClaimKey = claim.ClaimKey
				citation.NativeLocator = firstNonEmpty(citation.NativeLocator, allowed.NativeLocator)
				citation.Quote = firstNonEmpty(citation.Quote, allowed.Quote)
				claim.Citations[citationIndex] = citation
			}
			page.Claims[claimIndex] = claim
			claimKeys[claim.ClaimKey] = struct{}{}
		}
		for conflictIndex, conflict := range page.Conflicts {
			conflict.ClaimKey = strings.TrimSpace(conflict.ClaimKey)
			conflict.Summary = strings.TrimSpace(conflict.Summary)
			if conflict.ClaimKey == "" {
				errorsOut = append(errorsOut, fmt.Sprintf("page[%d].conflicts[%d].claim_key is required", pageIndex, conflictIndex))
			}
			if conflict.Summary == "" {
				errorsOut = append(errorsOut, fmt.Sprintf("page[%d].conflicts[%d].summary is required", pageIndex, conflictIndex))
			}
			if _, ok := claimKeys[conflict.ClaimKey]; conflict.ClaimKey != "" && !ok {
				errorsOut = append(errorsOut, fmt.Sprintf("page[%d].conflicts[%d] references unknown claim_key %q", pageIndex, conflictIndex, conflict.ClaimKey))
			}
			if len(conflict.Citations) == 0 && conflict.ClaimKey != "" {
				conflict.Citations = citationsForSynthesisClaim(page.Claims, conflict.ClaimKey)
			}
			for citationIndex, citation := range conflict.Citations {
				citation.SourceDocumentID = strings.TrimSpace(citation.SourceDocumentID)
				citation.SourceRevisionID = strings.TrimSpace(citation.SourceRevisionID)
				citation.ChunkID = strings.TrimSpace(citation.ChunkID)
				allowed, ok := allowedCitations[citationKey(citation)]
				if !ok {
					errorsOut = append(errorsOut, fmt.Sprintf("page[%d].conflicts[%d].citations[%d] references unknown chunk %q", pageIndex, conflictIndex, citationIndex, citation.ChunkID))
					continue
				}
				citation.ClaimKey = conflict.ClaimKey
				citation.NativeLocator = firstNonEmpty(citation.NativeLocator, allowed.NativeLocator)
				citation.Quote = firstNonEmpty(citation.Quote, allowed.Quote)
				conflict.Citations[citationIndex] = citation
			}
			page.Conflicts[conflictIndex] = conflict
		}
		if containsLikelySecret(page.Title + "\n" + page.Summary + "\n" + claimsText(page.Claims)) {
			errorsOut = append(errorsOut, fmt.Sprintf("page[%d] contains secret-like text", pageIndex))
		}
		pages = append(pages, page)
	}
	return pages, errorsOut
}

func allowedSynthesisCitations(evidence store.CompanyWikiSourceEvidence, candidates []store.CompanyWikiPageRead) map[string]store.CompanyWikiCitationInput {
	out := map[string]store.CompanyWikiCitationInput{}
	for _, chunk := range evidence.Chunks {
		citation := store.CompanyWikiCitationInput{
			SourceDocumentID: chunk.DocumentID,
			SourceRevisionID: chunk.RevisionID,
			ChunkID:          chunk.ID,
			NativeLocator:    chunk.NativeLocator,
			Quote:            truncateForCitation(chunk.Content, 260),
		}
		out[citationKey(citation)] = citation
	}
	for _, candidate := range candidates {
		for _, citation := range candidate.Citations {
			input := store.CompanyWikiCitationInput{
				ClaimKey:         citation.ClaimKey,
				SourceDocumentID: citation.SourceDocumentID,
				SourceRevisionID: citation.SourceRevisionID,
				ChunkID:          citation.ChunkID,
				NativeLocator:    citation.NativeLocator,
				Quote:            citation.Quote,
			}
			if strings.TrimSpace(input.SourceDocumentID) != "" && strings.TrimSpace(input.SourceRevisionID) != "" && strings.TrimSpace(input.ChunkID) != "" {
				out[citationKey(input)] = input
			}
		}
	}
	return out
}

func citationKey(citation store.CompanyWikiCitationInput) string {
	return strings.TrimSpace(citation.SourceDocumentID) + "\x00" + strings.TrimSpace(citation.SourceRevisionID) + "\x00" + strings.TrimSpace(citation.ChunkID)
}

func candidateCitationInputsForClaim(candidate store.CompanyWikiPageRead, claimKey string) []store.CompanyWikiCitationInput {
	claimKey = strings.TrimSpace(claimKey)
	out := []store.CompanyWikiCitationInput{}
	for _, citation := range candidate.Citations {
		if strings.TrimSpace(citation.ClaimKey) != claimKey {
			continue
		}
		out = append(out, store.CompanyWikiCitationInput{
			ClaimKey:         claimKey,
			SourceDocumentID: citation.SourceDocumentID,
			SourceRevisionID: citation.SourceRevisionID,
			ChunkID:          citation.ChunkID,
			NativeLocator:    citation.NativeLocator,
			Quote:            citation.Quote,
		})
	}
	return out
}

func citationsForSynthesisClaim(claims []WikiSynthesisClaim, claimKey string) []store.CompanyWikiCitationInput {
	claimKey = strings.TrimSpace(claimKey)
	for _, claim := range claims {
		if strings.TrimSpace(claim.ClaimKey) == claimKey {
			return append([]store.CompanyWikiCitationInput(nil), claim.Citations...)
		}
	}
	return nil
}

func RenderSynthesisPageMarkdown(evidence store.CompanyWikiSourceEvidence, page WikiSynthesisPage) (string, []store.CompanyWikiCitationInput, []store.CompanyWikiClaimInput, []store.CompanyWikiConflictInput) {
	return renderSynthesisPageMarkdownWithCandidates(evidence, nil, page)
}

func renderSynthesisPageMarkdownWithCandidates(evidence store.CompanyWikiSourceEvidence, candidates []store.CompanyWikiPageRead, page WikiSynthesisPage) (string, []store.CompanyWikiCitationInput, []store.CompanyWikiClaimInput, []store.CompanyWikiConflictInput) {
	slug := synthesisSlug(page)
	pageType := normalizeSynthesisPageType(page.Type)
	revisionTimestamps := buildRevisionTimestampMap(evidence, candidates)
	freshness := synthesisFreshness(revisionTimestamps, page)

	citations := []store.CompanyWikiCitationInput{}
	claims := []store.CompanyWikiClaimInput{}
	conflicts := []store.CompanyWikiConflictInput{}
	sourceRevisionIDs := uniqueCitationRevisionIDs(citationsForSynthesisPage(page))
	if len(sourceRevisionIDs) == 0 && strings.TrimSpace(evidence.Revision.ID) != "" {
		sourceRevisionIDs = []string{strings.TrimSpace(evidence.Revision.ID)}
	}

	var b strings.Builder
	b.WriteString("---\n")
	writeYAMLScalar(&b, "title", page.Title)
	writeYAMLScalar(&b, "type", pageType)
	writeYAMLScalar(&b, "slug", slug)
	writeYAMLScalar(&b, "freshness", freshness)
	tags := normalizeStringList(page.Tags)
	if len(tags) == 0 {
		b.WriteString("tags: []\n")
	} else {
		b.WriteString("tags:\n")
	}
	for _, tag := range tags {
		b.WriteString("  - ")
		b.WriteString(yamlQuote(tag))
		b.WriteString("\n")
	}
	owners := normalizeStringList(page.Owners)
	if len(owners) == 0 {
		b.WriteString("owners: []\n")
	} else {
		b.WriteString("owners:\n")
	}
	for _, owner := range owners {
		b.WriteString("  - ")
		b.WriteString(yamlQuote(owner))
		b.WriteString("\n")
	}
	b.WriteString("source_revision_ids:\n")
	for _, revisionID := range sourceRevisionIDs {
		b.WriteString("  - ")
		b.WriteString(yamlQuote(revisionID))
		b.WriteString("\n")
	}
	writeYAMLScalar(&b, "conflict_state", conflictState(page.Conflicts))
	b.WriteString("---\n\n")

	b.WriteString("# ")
	b.WriteString(strings.TrimSpace(page.Title))
	b.WriteString("\n\n")
	if strings.TrimSpace(page.Summary) != "" {
		b.WriteString("## Summary\n\n")
		b.WriteString(strings.TrimSpace(page.Summary))
		b.WriteString("\n\n")
	}
	if len(page.Claims) > 0 {
		b.WriteString("## Claims\n\n")
	}
	for _, claim := range page.Claims {
		claimKey := strings.TrimSpace(claim.ClaimKey)
		claims = append(claims, store.CompanyWikiClaimInput{
			ClaimKey:   claimKey,
			ClaimText:  strings.TrimSpace(claim.Text),
			Confidence: claim.Confidence,
			Metadata: map[string]any{
				"citation_refs":       citationRefsForMetadataWithMap(claim.Citations, revisionTimestamps),
				"source_revision_ids": uniqueCitationRevisionIDs(claim.Citations),
				"page_type":           pageType,
			},
		})
		b.WriteString("- ")
		b.WriteString(strings.TrimSpace(claim.Text))
		b.WriteString(" `claim:")
		b.WriteString(claimKey)
		b.WriteString("`")
		if claim.Confidence > 0 {
			b.WriteString(fmt.Sprintf(" `confidence:%.2f`", claim.Confidence))
		}
		b.WriteString("\n")
		for _, citation := range claim.Citations {
			citation.ClaimKey = claimKey
			citations = append(citations, citation)
			b.WriteString("  - citation: `source_document_id=")
			b.WriteString(citation.SourceDocumentID)
			b.WriteString("` `source_revision_id=")
			b.WriteString(citation.SourceRevisionID)
			b.WriteString("` `chunk_id=")
			b.WriteString(citation.ChunkID)
			b.WriteString("`")
			if strings.TrimSpace(citation.NativeLocator) != "" {
				b.WriteString(" `native_locator=")
				b.WriteString(strings.ReplaceAll(citation.NativeLocator, "`", "'"))
				b.WriteString("`")
			}
			b.WriteString(" `source_timestamp=")
			b.WriteString(synthesisCitationTimestampWithMap(revisionTimestamps, citation))
			b.WriteString("`")
			b.WriteString("\n")
		}
	}
	if len(page.Claims) > 0 {
		b.WriteString("\n")
	}
	if len(page.Conflicts) > 0 {
		b.WriteString("## Conflicts\n\n")
		for _, conflict := range page.Conflicts {
			conflictCitations := conflict.Citations
			if len(conflictCitations) == 0 {
				conflictCitations = citationsForSynthesisClaim(page.Claims, conflict.ClaimKey)
			}
			conflicts = append(conflicts, store.CompanyWikiConflictInput{
				ClaimKey:  conflict.ClaimKey,
				Summary:   conflict.Summary,
				Citations: append([]string(nil), conflict.CitationIDs...),
				Metadata: map[string]any{
					"citation_refs":       citationRefsForMetadataWithMap(conflictCitations, revisionTimestamps),
					"source_revision_ids": uniqueCitationRevisionIDs(conflictCitations),
					"page_type":           pageType,
				},
			})
			b.WriteString("- ")
			b.WriteString(strings.TrimSpace(conflict.Summary))
			b.WriteString(" `claim:")
			b.WriteString(strings.TrimSpace(conflict.ClaimKey))
			b.WriteString("`\n")
			for _, citation := range conflictCitations {
				b.WriteString("  - conflict citation: `source_document_id=")
				b.WriteString(citation.SourceDocumentID)
				b.WriteString("` `source_revision_id=")
				b.WriteString(citation.SourceRevisionID)
				b.WriteString("` `chunk_id=")
				b.WriteString(citation.ChunkID)
				b.WriteString("`")
				if strings.TrimSpace(citation.NativeLocator) != "" {
					b.WriteString(" `native_locator=")
					b.WriteString(strings.ReplaceAll(citation.NativeLocator, "`", "'"))
					b.WriteString("`")
				}
				b.WriteString(" `source_timestamp=")
				b.WriteString(synthesisCitationTimestampWithMap(revisionTimestamps, citation))
				b.WriteString("`")
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}
	if len(page.OpenQuestions) > 0 {
		b.WriteString("## Open Questions\n\n")
		for _, question := range page.OpenQuestions {
			b.WriteString("- ")
			b.WriteString(strings.TrimSpace(question))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	if len(page.RelatedPages) > 0 {
		b.WriteString("## Related Pages\n\n")
		for _, related := range page.RelatedPages {
			b.WriteString("- `")
			b.WriteString(store.NormalizeCompanyWikiSlug(related))
			b.WriteString("`\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("## Sources\n\n")
	b.WriteString("- `source_document_id`: `")
	b.WriteString(evidence.Document.ID)
	b.WriteString("`\n")
	b.WriteString("- `source_revision_id`: `")
	b.WriteString(evidence.Revision.ID)
	b.WriteString("`\n")
	if strings.TrimSpace(evidence.Document.URL) != "" {
		b.WriteString("- `source_url`: ")
		b.WriteString(evidence.Document.URL)
		b.WriteString("\n")
	}

	return b.String(), citations, claims, conflicts
}

func synthesisSlug(page WikiSynthesisPage) string {
	pageType := normalizeSynthesisPageType(page.Type)
	slug := store.NormalizeCompanyWikiSlug(firstNonEmpty(page.Slug, page.Title))
	root := synthesisPageTypeRoot(pageType)
	if strings.Contains(slug, "/") {
		parts := strings.Split(slug, "/")
		first := ""
		for _, part := range parts {
			if strings.TrimSpace(part) != "" {
				first = part
				break
			}
		}
		if isSynthesisPageRoot(first) {
			return slug
		}
		for i := len(parts) - 1; i >= 0; i-- {
			if strings.TrimSpace(parts[i]) != "" {
				slug = parts[i]
				break
			}
		}
	}
	return store.NormalizeCompanyWikiSlug(filepath.ToSlash(filepath.Join(root, slug)))
}

func isSynthesisPageRoot(value string) bool {
	switch strings.TrimSpace(value) {
	case "projects", "systems", "decisions", "runbooks", "policies", "people", "open-questions", "concepts":
		return true
	default:
		return false
	}
}

func synthesisFreshness(revisionTimestamps map[string]string, page WikiSynthesisPage) string {
	var newest time.Time
	for _, citation := range citationsForSynthesisPage(page) {
		ts, ok := sourceTimestampForRevision(revisionTimestamps, citation.SourceRevisionID)
		if !ok {
			continue
		}
		if newest.IsZero() || ts.After(newest) {
			newest = ts
		}
	}
	if !newest.IsZero() {
		return newest.Format(time.RFC3339)
	}
	return "unknown"
}

func candidateClaimMateriallyFresher(existingCitations []store.CompanyWikiCitationInput, candidateCitations []store.CompanyWikiCitationInput, revisionTimestamps map[string]string) bool {
	existingNewest, existingOK := newestCitationTimestamp(existingCitations, revisionTimestamps)
	candidateNewest, candidateOK := newestCitationTimestamp(candidateCitations, revisionTimestamps)
	if !candidateOK {
		return false
	}
	if !existingOK {
		return false
	}
	return candidateNewest.After(existingNewest.Add(companyWikiCloseSourceTimestampWindow))
}

func newestCitationTimestamp(citations []store.CompanyWikiCitationInput, revisionTimestamps map[string]string) (time.Time, bool) {
	var newest time.Time
	ok := false
	for _, citation := range citations {
		ts, found := sourceTimestampForRevision(revisionTimestamps, citation.SourceRevisionID)
		if !found {
			continue
		}
		if !ok || ts.After(newest) {
			newest = ts
			ok = true
		}
	}
	return newest, ok
}

func sourceTimestampForRevision(revisionTimestamps map[string]string, revisionID string) (time.Time, bool) {
	if revisionTimestamps == nil {
		return time.Time{}, false
	}
	raw := strings.TrimSpace(revisionTimestamps[strings.TrimSpace(revisionID)])
	if raw == "" || raw == "unknown" {
		return time.Time{}, false
	}
	if ts, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return ts, true
	}
	if ts, err := time.Parse(time.RFC3339, raw); err == nil {
		return ts, true
	}
	return time.Time{}, false
}

func citationsForSynthesisPage(page WikiSynthesisPage) []store.CompanyWikiCitationInput {
	out := []store.CompanyWikiCitationInput{}
	for _, claim := range page.Claims {
		out = append(out, claim.Citations...)
	}
	for _, conflict := range page.Conflicts {
		out = append(out, conflict.Citations...)
	}
	return out
}

func buildRevisionTimestampMap(evidence store.CompanyWikiSourceEvidence, candidates []store.CompanyWikiPageRead) map[string]string {
	m := make(map[string]string)
	currentRevID := strings.TrimSpace(evidence.Revision.ID)
	if currentRevID != "" {
		if !evidence.Revision.ObservedAt.IsZero() {
			m[currentRevID] = evidence.Revision.ObservedAt.Format(time.RFC3339)
		} else if !evidence.Revision.CreatedAt.IsZero() {
			m[currentRevID] = evidence.Revision.CreatedAt.Format(time.RFC3339)
		}
	}
	for _, chunk := range evidence.Chunks {
		revisionID := strings.TrimSpace(chunk.RevisionID)
		if revisionID == "" || m[revisionID] != "" {
			continue
		}
		if timestamp, ok := sourceTimestampForChunk(chunk); ok {
			m[revisionID] = timestamp.Format(time.RFC3339)
		}
	}
	for _, candidate := range candidates {
		for _, claim := range candidate.Claims {
			for _, refMap := range citationRefsFromMetadata(claim.Metadata) {
				revID := ""
				if v, ok := refMap["source_revision_id"]; ok {
					revID = strings.TrimSpace(v)
				}
				if revID == "" || m[revID] != "" {
					continue
				}
				if ts := strings.TrimSpace(refMap["source_timestamp"]); ts != "" && ts != "unknown" {
					m[revID] = ts
				}
			}
		}
		for _, conflict := range candidate.Conflicts {
			for _, refMap := range citationRefsFromMetadata(conflict.Metadata) {
				revID := ""
				if v, ok := refMap["source_revision_id"]; ok {
					revID = strings.TrimSpace(v)
				}
				if revID == "" || m[revID] != "" {
					continue
				}
				if ts := strings.TrimSpace(refMap["source_timestamp"]); ts != "" && ts != "unknown" {
					m[revID] = ts
				}
			}
		}
	}
	return m
}

func sourceTimestampForChunk(chunk store.CompanyWikiSourceChunk) (time.Time, bool) {
	for _, key := range []string{"source_observed_at", "observed_at", "last_edited_time", "created_time"} {
		if ts, ok := sourceTimestampFromAny(chunk.Metadata[key]); ok {
			return ts, true
		}
	}
	if ts, ok := sourceTimestampFromAny(chunk.Metadata["slack_ts"]); ok {
		return ts, true
	}
	return time.Time{}, false
}

func sourceTimestampFromAny(value any) (time.Time, bool) {
	switch typed := value.(type) {
	case time.Time:
		if typed.IsZero() {
			return time.Time{}, false
		}
		return typed.UTC(), true
	case string:
		text := strings.TrimSpace(typed)
		if text == "" || text == "unknown" {
			return time.Time{}, false
		}
		if ts := SlackTimestampToTime(text); !ts.IsZero() {
			return ts, true
		}
		if ts, err := time.Parse(time.RFC3339Nano, text); err == nil {
			return ts.UTC(), true
		}
		if ts, err := time.Parse(time.RFC3339, text); err == nil {
			return ts.UTC(), true
		}
	default:
		return time.Time{}, false
	}
	return time.Time{}, false
}

func citationRefsFromMetadata(metadata map[string]any) []map[string]string {
	if metadata == nil {
		return nil
	}
	switch refs := metadata["citation_refs"].(type) {
	case []map[string]string:
		out := make([]map[string]string, 0, len(refs))
		for _, ref := range refs {
			out = append(out, ref)
		}
		return out
	case []map[string]any:
		out := make([]map[string]string, 0, len(refs))
		for _, ref := range refs {
			out = append(out, stringifyCitationRef(ref))
		}
		return out
	case []interface{}:
		out := make([]map[string]string, 0, len(refs))
		for _, ref := range refs {
			if refMap, ok := ref.(map[string]interface{}); ok {
				out = append(out, stringifyCitationRef(refMap))
			}
		}
		return out
	default:
		return nil
	}
}

func stringifyCitationRef(ref map[string]any) map[string]string {
	out := map[string]string{}
	for key, value := range ref {
		if text, ok := value.(string); ok {
			out[key] = text
		}
	}
	return out
}

func synthesisCitationTimestampWithMap(revisionTimestamps map[string]string, citation store.CompanyWikiCitationInput) string {
	revID := strings.TrimSpace(citation.SourceRevisionID)
	if revID != "" {
		if ts, ok := revisionTimestamps[revID]; ok && ts != "" {
			return ts
		}
	}
	return "unknown"
}

func citationRefsForMetadataWithMap(citations []store.CompanyWikiCitationInput, revisionTimestamps map[string]string) []map[string]string {
	out := make([]map[string]string, 0, len(citations))
	for _, citation := range citations {
		out = append(out, map[string]string{
			"source_document_id": citation.SourceDocumentID,
			"source_revision_id": citation.SourceRevisionID,
			"chunk_id":           citation.ChunkID,
			"source_timestamp":   synthesisCitationTimestampWithMap(revisionTimestamps, citation),
		})
	}
	return out
}

func uniqueCitationRevisionIDs(citations []store.CompanyWikiCitationInput) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, citation := range citations {
		revisionID := strings.TrimSpace(citation.SourceRevisionID)
		if revisionID == "" {
			continue
		}
		if _, ok := seen[revisionID]; ok {
			continue
		}
		seen[revisionID] = struct{}{}
		out = append(out, revisionID)
	}
	sort.Strings(out)
	return out
}

func synthesisSystemPrompt() string {
	return strings.Join([]string{
		"You are the RSI company wiki compiler.",
		"Return JSON only. Do not include markdown fences.",
		"Compress Slack/Notion evidence into durable company wiki pages.",
		"Emit only structured pages with cited claim objects. Every factual claim must cite existing source chunk IDs or preserved candidate claim citations.",
		"Preserve conflicts explicitly instead of resolving unsupported disagreements.",
		"Freshness policy: newer cited source timestamps usually win for the same claim key. If sources are within 24 hours and disagree, mark the claim conflicted or superseded instead of silently overwriting.",
		"Older sources may add missing facts when they cover gaps that fresher candidate pages do not already cover.",
	}, "\n")
}

func synthesisUserPrompt(request WikiSynthesisRequest) string {
	type promptChunk struct {
		ID            string `json:"id"`
		RevisionID    string `json:"source_revision_id"`
		DocumentID    string `json:"source_document_id"`
		NativeLocator string `json:"native_locator,omitempty"`
		Content       string `json:"content"`
	}
	chunks := make([]promptChunk, 0, len(request.Chunks))
	for _, chunk := range request.Chunks {
		chunks = append(chunks, promptChunk{
			ID:            chunk.ID,
			RevisionID:    chunk.RevisionID,
			DocumentID:    chunk.DocumentID,
			NativeLocator: chunk.NativeLocator,
			Content:       truncateForCitation(chunk.Content, 3000),
		})
	}
	candidates := make([]map[string]any, 0, len(request.CandidatePages))
	for _, page := range request.CandidatePages {
		claims := make([]map[string]any, 0, len(page.Claims))
		for _, claim := range page.Claims {
			claims = append(claims, map[string]any{
				"claim_key":  claim.ClaimKey,
				"text":       claim.ClaimText,
				"confidence": claim.Confidence,
				"citations":  candidateCitationInputsForClaim(page, claim.ClaimKey),
			})
		}
		candidates = append(candidates, map[string]any{
			"slug":    page.Page.Slug,
			"title":   page.Page.Title,
			"type":    stringFromMap(page.Revision.Metadata, "type"),
			"summary": wikiOneLineSummary(page.Revision.Body),
			"claims":  claims,
		})
	}
	payload := map[string]any{
		"instructions": []string{
			"Return an object with a pages array.",
			"Each page needs slug, title, type, tags, summary, owners, freshness, claims, conflicts, open_questions, related_pages.",
			"Valid page types: project, system, decision, runbook, policy, person, concept, open_question.",
			"Use semantic wiki slugs. Do not return source-shaped roots such as slack_message, notion_document, sources, slack, or notion.",
			"Claim citations must use source_document_id, source_revision_id, chunk_id, native_locator, quote.",
			"Use only the provided source chunks and existing candidate claim citations as source evidence.",
			"When updating a candidate page, preserve existing claims unless a new cited claim explicitly supersedes them.",
		},
		"source": map[string]any{
			"document_id":        request.Source.Document.ID,
			"source_type":        request.Source.Document.SourceType,
			"source_key":         request.Source.Document.SourceKey,
			"source_revision_id": request.Source.Revision.ID,
			"title":              request.Source.Document.Title,
			"url":                request.Source.Document.URL,
			"observed_at":        request.Source.Revision.ObservedAt.Format(time.RFC3339),
		},
		"chunks":          chunks,
		"candidate_pages": candidates,
	}
	return mustMarshalString(payload)
}

func normalizeSynthesisPageType(value string) string {
	value = store.NormalizeCompanyWikiSlug(strings.ReplaceAll(value, "_", "-"))
	switch value {
	case "projects", "project":
		return "project"
	case "systems", "system":
		return "system"
	case "decisions", "decision":
		return "decision"
	case "runbooks", "runbook":
		return "runbook"
	case "policies", "policy":
		return "policy"
	case "people", "person":
		return "person"
	case "open-questions", "open-question":
		return "open_question"
	case "concepts", "concept":
		return "concept"
	default:
		return "concept"
	}
}

func synthesisPageTypeRoot(pageType string) string {
	switch normalizeSynthesisPageType(pageType) {
	case "project":
		return "projects"
	case "system":
		return "systems"
	case "decision":
		return "decisions"
	case "runbook":
		return "runbooks"
	case "policy":
		return "policies"
	case "person":
		return "people"
	case "open_question":
		return "open-questions"
	default:
		return "concepts"
	}
}

func normalizeStringList(values []string) []string {
	seen := map[string]string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; !ok {
			seen[key] = value
		}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]string, len(keys))
	for i, key := range keys {
		out[i] = seen[key]
	}
	return out
}

func conflictState(conflicts []WikiSynthesisConflict) string {
	if len(conflicts) == 0 {
		return "none"
	}
	return "disputed"
}

func storeMergeMetadata(values ...map[string]any) map[string]any {
	out := map[string]any{}
	for _, metadata := range values {
		for key, value := range metadata {
			out[key] = value
		}
	}
	return out
}

func boolFromMetadata(metadata map[string]any, key string) bool {
	if metadata == nil {
		return false
	}
	return boolFromAny(metadata[key])
}

func boolFromAny(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true") || strings.EqualFold(strings.TrimSpace(typed), "yes") || strings.TrimSpace(typed) == "1"
	default:
		return false
	}
}

func claimsText(claims []WikiSynthesisClaim) string {
	parts := make([]string, 0, len(claims))
	for _, claim := range claims {
		parts = append(parts, claim.Text)
	}
	return strings.Join(parts, "\n")
}

func containsLikelySecret(value string) bool {
	value = strings.ToLower(value)
	needles := []string{
		"sk-",
		"xoxb-",
		"xoxp-",
		"aws_secret_access_key",
		"-----begin private key-----",
		"authorization: bearer",
	}
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
