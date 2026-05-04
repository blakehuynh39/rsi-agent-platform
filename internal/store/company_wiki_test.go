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
	again, err := state.UpsertCompanyWikiSourceRevision(CompanyWikiSourceRevisionInput{
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
		t.Fatalf("second UpsertCompanyWikiSourceRevision() error = %v", err)
	}
	if again.Inserted || again.Changed {
		t.Fatalf("duplicate source revision should be a no-op, got %+v", again)
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

func TestCompanyWikiCompileItemVersioningTargetsAndCandidateFilters(t *testing.T) {
	state := NewMemoryStore()
	source, err := state.UpsertCompanyWikiSourceRevision(CompanyWikiSourceRevisionInput{
		SourceType:        "slack_message",
		DocumentSourceKey: "slack:T:C:1",
		SourceSessionKey:  "slack:T:C:thread",
		SourceRevision:    "rev-1",
		Title:             "Compiler thread",
		Content:           "Decision: update RSI Platform wiki.",
		Metadata:          map[string]any{"mirror_allowed": true, "channel_private": false},
	})
	if err != nil {
		t.Fatalf("UpsertCompanyWikiSourceRevision() error = %v", err)
	}
	item, inserted, err := state.EnqueueCompanyWikiCompileItem(CompanyWikiCompileItemInput{
		SourceRevisionID:   source.Revision.ID,
		CompilerVersion:    "compiler.v1",
		SchemaVersion:      "schema.v1",
		RendererVersion:    "renderer.v1",
		ModelPolicyVersion: "policy.v1",
		InputHash:          CompanyWikiSHA256(source.Revision.ID),
	})
	if err != nil || !inserted {
		t.Fatalf("EnqueueCompanyWikiCompileItem() item=%+v inserted=%t err=%v", item, inserted, err)
	}
	again, inserted, err := state.EnqueueCompanyWikiCompileItem(CompanyWikiCompileItemInput{
		SourceRevisionID:   source.Revision.ID,
		CompilerVersion:    "compiler.v1",
		SchemaVersion:      "schema.v1",
		RendererVersion:    "renderer.v1",
		ModelPolicyVersion: "policy.v1",
		InputHash:          "different-context-should-not-create-new-item",
	})
	if err != nil || inserted || again.ID != item.ID {
		t.Fatalf("duplicate compile item should be no-op, item=%+v inserted=%t err=%v", again, inserted, err)
	}
	claimed, err := state.ClaimCompanyWikiCompileItems(CompanyWikiCompileClaimInput{
		Limit:              5,
		LeaseHolder:        "test",
		CompilerVersion:    "compiler.v1",
		SchemaVersion:      "schema.v1",
		RendererVersion:    "renderer.v1",
		ModelPolicyVersion: "policy.v1",
		MaxAttempts:        5,
	})
	if err != nil || len(claimed) != 1 || claimed[0].ID != item.ID {
		t.Fatalf("ClaimCompanyWikiCompileItems() = %+v err=%v", claimed, err)
	}
	targets, err := state.UpsertCompanyWikiCompileTargets(item.ID, []CompanyWikiCompileTargetInput{
		{TargetSlug: "projects/rsi-platform", TargetPath: "pages/projects/rsi-platform.md", TargetType: "project", BodyHash: "hash-1"},
		{TargetSlug: "systems/hermes", TargetPath: "pages/systems/hermes.md", TargetType: "system", BodyHash: "hash-2"},
	})
	if err != nil || len(targets) != 2 {
		t.Fatalf("UpsertCompanyWikiCompileTargets() = %+v err=%v", targets, err)
	}
	targets, err = state.UpsertCompanyWikiCompileTargets(item.ID, []CompanyWikiCompileTargetInput{
		{TargetSlug: "projects/rsi-platform", TargetPath: "pages/projects/rsi-platform.md", TargetType: "project", BodyHash: "hash-1"},
	})
	if err != nil || len(targets) != 1 {
		t.Fatalf("second UpsertCompanyWikiCompileTargets() = %+v err=%v", targets, err)
	}
	allTargets, err := state.ListCompanyWikiCompileTargets(item.ID)
	if err != nil {
		t.Fatalf("ListCompanyWikiCompileTargets() error = %v", err)
	}
	statusBySlug := map[string]string{}
	for _, target := range allTargets {
		statusBySlug[target.TargetSlug] = target.Status
	}
	if statusBySlug["systems/hermes"] != CompanyWikiCompileTargetStatusSuperseded {
		t.Fatalf("stale target should be superseded, got %+v", allTargets)
	}
}

func TestCompanyWikiCompileItemsPrioritizeNotionBeforeSlack(t *testing.T) {
	state := NewMemoryStore()
	slackSource, err := state.UpsertCompanyWikiSourceRevision(CompanyWikiSourceRevisionInput{
		SourceType:        "slack_message",
		DocumentSourceKey: "slack:T:C:channel",
		SourceSessionKey:  "slack:T:C:channel",
		SourceRevision:    "slack-rev-1",
		Title:             "Slack backlog",
		Content:           "Older Slack project chatter.",
		Metadata:          map[string]any{"mirror_allowed": true, "channel_private": false},
	})
	if err != nil {
		t.Fatalf("insert slack source: %v", err)
	}
	notionSource, err := state.UpsertCompanyWikiSourceRevision(CompanyWikiSourceRevisionInput{
		SourceType:        "notion_document",
		DocumentSourceKey: "notion:project:abc",
		SourceSessionKey:  "notion:project:abc",
		SourceRevision:    "notion-rev-1",
		Title:             "Project Bootstrap SoT",
		Content:           "Notion has the project source of truth.",
		Metadata:          map[string]any{"notion_allowlisted": true},
	})
	if err != nil {
		t.Fatalf("insert notion source: %v", err)
	}
	for _, source := range []CompanyWikiSourceRevisionResult{slackSource, notionSource} {
		if _, _, err := state.EnqueueCompanyWikiCompileItem(CompanyWikiCompileItemInput{
			SourceRevisionID:   source.Revision.ID,
			CompilerVersion:    "compiler.v1",
			SchemaVersion:      "schema.v1",
			RendererVersion:    "renderer.v1",
			ModelPolicyVersion: "policy.v1",
			InputHash:          CompanyWikiSHA256(source.Revision.ID),
		}); err != nil {
			t.Fatalf("enqueue source %s: %v", source.Revision.ID, err)
		}
	}
	claimed, err := state.ClaimCompanyWikiCompileItems(CompanyWikiCompileClaimInput{
		Limit:              1,
		LeaseHolder:        "test",
		CompilerVersion:    "compiler.v1",
		SchemaVersion:      "schema.v1",
		RendererVersion:    "renderer.v1",
		ModelPolicyVersion: "policy.v1",
		MaxAttempts:        5,
	})
	if err != nil {
		t.Fatalf("ClaimCompanyWikiCompileItems() error = %v", err)
	}
	if len(claimed) != 1 || claimed[0].SourceRevisionID != notionSource.Revision.ID {
		t.Fatalf("expected Notion compile item first, got %+v want revision %s", claimed, notionSource.Revision.ID)
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
