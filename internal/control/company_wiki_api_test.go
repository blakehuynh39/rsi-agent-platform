package control

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

func testCompanyWikiMarkdown(title string, sourceRevisionID string, markdownBody string) string {
	return "---\ntitle: " + title + "\nsource_revision_ids:\n  - " + sourceRevisionID + "\n---\n" + markdownBody
}

func TestCompanyWikiEditApplyRequiresCitationsAndPublishes(t *testing.T) {
	state := store.NewMemoryStore()
	root := t.TempDir()
	source, err := state.UpsertCompanyWikiSourceRevision(store.CompanyWikiSourceRevisionInput{
		SourceType:        "slack_message",
		DocumentSourceKey: "slack:T:C:1",
		SourceSessionKey:  "slack:T:C:thread",
		SourceRevision:    "1",
		Title:             "Slack discussion",
		Content:           "Use the deploy checklist before rollout.",
		NativeLocator:     "slack:C:1:1",
	})
	if err != nil {
		t.Fatalf("UpsertCompanyWikiSourceRevision() error = %v", err)
	}
	_, status, err := companyWikiEditApply(context.Background(), config.Config{CompanyWikiRoot: root}, state, companyWikiEditApplyRequest{
		Actor:          "hermes",
		Reason:         "test",
		IdempotencyKey: "missing-citation",
		Slug:           "runbooks/deploy",
		Title:          "Deploy",
		Body:           testCompanyWikiMarkdown("Deploy", source.Revision.ID, "# Deploy\n"),
	})
	if status != http.StatusBadRequest || err == nil {
		t.Fatalf("missing citations status=%d err=%v, want bad request", status, err)
	}
	body := testCompanyWikiMarkdown("Deploy", source.Revision.ID, "# Deploy\n\nUse the checklist. [citation:src]\n")
	out, status, err := companyWikiEditApply(context.Background(), config.Config{CompanyWikiRoot: root}, state, companyWikiEditApplyRequest{
		Actor:          "hermes",
		Reason:         "test",
		IdempotencyKey: "apply-1",
		Slug:           "runbooks/deploy",
		Title:          "Deploy",
		Body:           body,
		Citations: []store.CompanyWikiCitationInput{{
			ClaimKey:         "deploy-checklist",
			SourceDocumentID: source.Document.ID,
			SourceRevisionID: source.Revision.ID,
			ChunkID:          source.Chunks[0].ID,
			NativeLocator:    source.Chunks[0].NativeLocator,
		}},
	})
	if err != nil || status != http.StatusCreated {
		t.Fatalf("companyWikiEditApply() status=%d err=%v out=%+v", status, err, out)
	}
	if out.Audit.Status != store.CompanyWikiAuditStatusPublished || out.Page == nil {
		t.Fatalf("unexpected apply response: %+v", out)
	}
	published, err := os.ReadFile(filepath.Join(root, "pages", "runbooks", "deploy.md"))
	if err != nil {
		t.Fatalf("published file missing: %v", err)
	}
	if !strings.Contains(string(published), "Use the checklist") {
		t.Fatalf("published file mismatch:\n%s", string(published))
	}
}

func TestCompanyWikiEditApplyValidatesFrontmatterSourceReferencesAndPrivacy(t *testing.T) {
	state := store.NewMemoryStore()
	root := t.TempDir()
	source, err := state.UpsertCompanyWikiSourceRevision(store.CompanyWikiSourceRevisionInput{
		SourceType:        "slack_message",
		DocumentSourceKey: "slack:T:C:1",
		SourceSessionKey:  "slack:T:C:thread",
		SourceRevision:    "1",
		Title:             "Slack discussion",
		Content:           "Use the deploy checklist before rollout.",
		NativeLocator:     "slack:C:1:1",
	})
	if err != nil {
		t.Fatalf("UpsertCompanyWikiSourceRevision() error = %v", err)
	}
	citation := store.CompanyWikiCitationInput{
		ClaimKey:         "deploy-checklist",
		SourceDocumentID: source.Document.ID,
		SourceRevisionID: source.Revision.ID,
		ChunkID:          source.Chunks[0].ID,
		NativeLocator:    source.Chunks[0].NativeLocator,
	}

	_, status, err := companyWikiEditApply(context.Background(), config.Config{CompanyWikiRoot: root}, state, companyWikiEditApplyRequest{
		Actor:          "hermes",
		Reason:         "test",
		IdempotencyKey: "missing-source-revision-frontmatter",
		Slug:           "runbooks/deploy",
		Title:          "Deploy",
		Body:           "---\ntitle: Deploy\n---\n# Deploy\n\nUse the checklist.\n",
		Citations:      []store.CompanyWikiCitationInput{citation},
	})
	if status != http.StatusBadRequest || err == nil || !strings.Contains(err.Error(), "source_revision_ids") {
		t.Fatalf("missing source_revision_ids status=%d err=%v, want bad request", status, err)
	}

	badCitation := citation
	badCitation.ChunkID = "srcchunk_missing"
	_, status, err = companyWikiEditApply(context.Background(), config.Config{CompanyWikiRoot: root}, state, companyWikiEditApplyRequest{
		Actor:          "hermes",
		Reason:         "test",
		IdempotencyKey: "bad-citation-ref",
		Slug:           "runbooks/deploy",
		Title:          "Deploy",
		Body:           testCompanyWikiMarkdown("Deploy", source.Revision.ID, "# Deploy\n\nUse the checklist.\n"),
		Citations:      []store.CompanyWikiCitationInput{badCitation},
	})
	if status != http.StatusBadRequest || err == nil || !strings.Contains(err.Error(), "citation chunk") {
		t.Fatalf("bad citation status=%d err=%v, want bad request", status, err)
	}

	_, status, err = companyWikiEditApply(context.Background(), config.Config{CompanyWikiRoot: root}, state, companyWikiEditApplyRequest{
		Actor:          "hermes",
		Reason:         "test",
		IdempotencyKey: "secret-body",
		Slug:           "runbooks/deploy",
		Title:          "Deploy",
		Body:           testCompanyWikiMarkdown("Deploy", source.Revision.ID, "# Deploy\n\nSLACK_BOT_TOKEN=xoxb-123456789-secret\n"),
		Citations:      []store.CompanyWikiCitationInput{citation},
	})
	if status != http.StatusBadRequest || err == nil || !strings.Contains(err.Error(), "privacy") {
		t.Fatalf("secret body status=%d err=%v, want bad request", status, err)
	}
}

func TestCompanyWikiPageGetAcceptsEncodedHierarchicalSlug(t *testing.T) {
	state := store.NewMemoryStore()
	root := t.TempDir()
	source, err := state.UpsertCompanyWikiSourceRevision(store.CompanyWikiSourceRevisionInput{
		SourceType:        "slack_message",
		DocumentSourceKey: "slack:T:C:1",
		SourceSessionKey:  "slack:T:C:thread",
		SourceRevision:    "1",
		Title:             "Slack discussion",
		Content:           "Use the deploy checklist before rollout.",
		NativeLocator:     "slack:C:1:1",
	})
	if err != nil {
		t.Fatalf("UpsertCompanyWikiSourceRevision() error = %v", err)
	}
	body := testCompanyWikiMarkdown("Deploy", source.Revision.ID, "# Deploy\n\nUse the checklist. [citation:src]\n")
	if _, status, err := companyWikiEditApply(context.Background(), config.Config{CompanyWikiRoot: root}, state, companyWikiEditApplyRequest{
		Actor:          "hermes",
		Reason:         "test",
		IdempotencyKey: "apply-encoded-get",
		Slug:           "runbooks/deploy",
		Title:          "Deploy",
		Body:           body,
		Citations: []store.CompanyWikiCitationInput{{
			ClaimKey:         "deploy-checklist",
			SourceDocumentID: source.Document.ID,
			SourceRevisionID: source.Revision.ID,
			ChunkID:          source.Chunks[0].ID,
			NativeLocator:    source.Chunks[0].NativeLocator,
		}},
	}); err != nil || status != http.StatusCreated {
		t.Fatalf("companyWikiEditApply() status=%d err=%v", status, err)
	}
	page, status, err := companyWikiPageGet(context.Background(), state, "runbooks%2Fdeploy")
	if err != nil || status != http.StatusOK {
		t.Fatalf("companyWikiPageGet(encoded) status=%d err=%v", status, err)
	}
	if page.Page.Slug != "runbooks/deploy" {
		t.Fatalf("slug = %q, want runbooks/deploy", page.Page.Slug)
	}
}
