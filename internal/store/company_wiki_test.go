package store

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCompanyWikiMemoryStorePublishesAndReadsPage(t *testing.T) {
	state := NewMemoryStore()
	source, err := state.UpsertCompanyWikiSourceRevision(CompanyWikiSourceRevisionInput{
		SourceType:        "notion_document",
		DocumentSourceKey: "notion:page:abc",
		SourceSessionKey:  "notion:page:abc",
		SourceRevision:    "rev-1",
		Workspace:         "rsi_company_knowledge",
		Environment:       "test",
		Title:             "Deploy Runbook",
		Content:           strings.Repeat("roll forward after validation\n", 20),
		NativeLocator:     "notion:block:path",
	})
	if err != nil {
		t.Fatalf("UpsertCompanyWikiSourceRevision() error = %v", err)
	}
	if !source.Inserted || len(source.Chunks) == 0 {
		t.Fatalf("unexpected source result: %+v", source)
	}
	audit, err := state.BeginCompanyWikiAudit(CompanyWikiAuditInput{
		Mode:           CompanyWikiAuditModeApply,
		Actor:          "test",
		Reason:         "publish fixture",
		IdempotencyKey: "fixture-1",
		Slug:           "runbooks/deploy",
		Title:          "Deploy Runbook",
	})
	if err != nil {
		t.Fatalf("BeginCompanyWikiAudit() error = %v", err)
	}
	page, err := state.PublishCompanyWikiPage(CompanyWikiPagePublishInput{
		AuditID: audit.ID,
		Slug:    "runbooks/deploy",
		Title:   "Deploy Runbook",
		Body:    "---\ntitle: Deploy Runbook\n---\n# Deploy Runbook\n",
		Path:    "pages/runbooks/deploy.md",
		SHA256:  CompanyWikiSHA256("body"),
		Citations: []CompanyWikiCitationInput{{
			ClaimKey:         "deploy",
			SourceDocumentID: source.Document.ID,
			SourceRevisionID: source.Revision.ID,
			ChunkID:          source.Chunks[0].ID,
			NativeLocator:    source.Chunks[0].NativeLocator,
		}},
		SourceRevisionIDs: []string{source.Revision.ID},
	})
	if err != nil {
		t.Fatalf("PublishCompanyWikiPage() error = %v", err)
	}
	if _, err := state.CompleteCompanyWikiAudit(audit.ID, page.Revision.ID, page.Revision.Path, map[string]any{"sha256": page.Revision.BodySHA256}); err != nil {
		t.Fatalf("CompleteCompanyWikiAudit() error = %v", err)
	}
	found, ok, err := state.GetCompanyWikiPage("runbooks/deploy")
	if err != nil {
		t.Fatalf("GetCompanyWikiPage() error = %v", err)
	}
	if !ok || found.Page.ID != page.Page.ID || len(found.Citations) != 1 {
		t.Fatalf("unexpected page read ok=%t read=%+v", ok, found)
	}
	results, err := state.SearchCompanyWikiPages("deploy", 5)
	if err != nil {
		t.Fatalf("SearchCompanyWikiPages() error = %v", err)
	}
	if len(results) != 1 || results[0].Slug != "runbooks/deploy" {
		t.Fatalf("unexpected search results: %+v", results)
	}
}

func TestValidateCompanyWikiCitationInputsRequiresStableIDs(t *testing.T) {
	err := ValidateCompanyWikiCitationInputs([]CompanyWikiCitationInput{{
		ClaimKey:         "missing-chunk",
		SourceDocumentID: "srcdoc_1",
		SourceRevisionID: "srcrev_1",
	}})
	if err == nil {
		t.Fatal("expected missing chunk_id to fail validation")
	}
}

func TestSnippetForWikiSearchIsUTF8Safe(t *testing.T) {
	body := strings.Repeat("界", 600)
	snippet := snippetForWikiSearch(body)
	if !utf8.ValidString(snippet) {
		t.Fatalf("snippet is not valid UTF-8")
	}
	if got := len([]rune(snippet)); got != 500 {
		t.Fatalf("snippet runes = %d, want 500", got)
	}
}
