package companyknowledge

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

type fakeWikiSynthesisClient struct {
	requests []WikiSynthesisRequest
	output   WikiSynthesisOutput
	err      error
}

func (f *fakeWikiSynthesisClient) SynthesizeWiki(ctx context.Context, request WikiSynthesisRequest) (WikiSynthesisOutput, WikiSynthesisMetadata, error) {
	_ = ctx
	f.requests = append(f.requests, request)
	if f.err != nil {
		return WikiSynthesisOutput{}, WikiSynthesisMetadata{}, f.err
	}
	if len(f.output.Pages) > 0 {
		return f.output, WikiSynthesisMetadata{RequestMetadataHash: "reqhash", ResponseMetadataHash: "resphash"}, nil
	}
	chunk := request.Chunks[0]
	return WikiSynthesisOutput{Pages: []WikiSynthesisPage{{
		Slug:      "rsi-platform",
		Title:     "RSI Platform",
		Type:      "project",
		Tags:      []string{"platform", "rsi"},
		Summary:   "RSI Platform status is synthesized from company evidence.",
		Owners:    []string{"Platform"},
		Freshness: chunk.CreatedAt.Format(time.RFC3339),
		Claims: []WikiSynthesisClaim{{
			ClaimKey:   "platform_status",
			Text:       "RSI Platform has a deployed company wiki compiler path.",
			Confidence: 0.92,
			Citations: []store.CompanyWikiCitationInput{{
				SourceDocumentID: chunk.DocumentID,
				SourceRevisionID: chunk.RevisionID,
				ChunkID:          chunk.ID,
				NativeLocator:    chunk.NativeLocator,
				Quote:            "company wiki compiler path",
			}},
		}},
	}}}, WikiSynthesisMetadata{RequestMetadataHash: "reqhash", ResponseMetadataHash: "resphash"}, nil
}

func TestRunCompanyWikiCompilerSynthesizesPageFromLedger(t *testing.T) {
	state := store.NewMemoryStore()
	root := t.TempDir()
	cfg := config.Config{
		CompanyWikiRoot:               root,
		CompanyWikiSynthesisEnabled:   true,
		CompanyWikiSourcePageMode:     "evidence",
		CompanyWikiCompilerModel:      "test/model",
		CompanyWikiCompilerBatchLimit: 10,
		CompanyWikiCompilerChunkLimit: 10,
		CompanyWikiCompilerTimeout:    time.Minute,
	}
	recorded, err := RecordEnqueueAndMaybePublishWikiSource(context.Background(), cfg, state, store.CompanyWikiSourceRevisionInput{
		SourceType:        SlackMessageSourceType,
		DocumentSourceKey: "slack:T123:C123:1777831927.000000",
		SourceSessionKey:  "slack:T123:C123:thread",
		SourceRevision:    "rev-1",
		Workspace:         "T123",
		Environment:       "test",
		Title:             "RSI Platform thread",
		URL:               "https://slack.example/archives/C123/p1777831927000000",
		Content:           "Decision: ship the company wiki compiler path for RSI Platform.",
		NativeLocator:     "slack:C123:1777831927.000000",
		Metadata: map[string]any{
			"source":           "slack",
			"channel_id":       "C123",
			"channel_type":     "public_channel",
			"channel_private":  false,
			"channel_im":       false,
			"mirror_discovery": "joined_public",
			"mirror_allowed":   true,
			"mirror_denied":    false,
		},
		ObservedAt: time.Unix(1777831927, 0).UTC(),
	})
	if err != nil {
		t.Fatalf("RecordEnqueueAndMaybePublishWikiSource() error = %v", err)
	}
	if !recorded.Source.Changed {
		t.Fatalf("expected changed source revision, got %+v", recorded.Source)
	}
	result, err := RunCompanyWikiCompiler(context.Background(), cfg, state, &fakeWikiSynthesisClient{})
	if err != nil {
		t.Fatalf("RunCompanyWikiCompiler() error = %v", err)
	}
	if !result.OK || result.Claimed != 1 || result.PublishedPages != 1 {
		t.Fatalf("unexpected compiler result: %+v", result)
	}
	for _, path := range []string{
		"SCHEMA.md",
		"index.md",
		"log.md",
		recorded.Page.Revision.Path,
		"pages/projects/rsi-platform.md",
	} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}
	if !strings.HasPrefix(recorded.Page.Revision.Path, "sources/slack/") {
		t.Fatalf("evidence page path = %q, want sources/slack/", recorded.Page.Revision.Path)
	}
	if strings.Contains(recorded.Page.Revision.Path, "sources/slack/slack_message/") {
		t.Fatalf("evidence page path has redundant source type nesting: %q", recorded.Page.Revision.Path)
	}
	page, found, err := state.GetCompanyWikiPage("projects/rsi-platform")
	if err != nil {
		t.Fatalf("GetCompanyWikiPage() error = %v", err)
	}
	if !found {
		t.Fatal("synthesis page not found")
	}
	if len(page.Claims) != 1 || len(page.Citations) != 1 {
		t.Fatalf("expected persisted claims/citations, got claims=%+v citations=%+v", page.Claims, page.Citations)
	}
	body := page.Revision.Body
	if !strings.Contains(body, "## Claims") || !strings.Contains(body, "`claim:platform_status`") {
		t.Fatalf("synthesis body missing rendered claim:\n%s", body)
	}
	indexRaw, err := os.ReadFile(filepath.Join(root, "index.md"))
	if err != nil {
		t.Fatalf("read index.md: %v", err)
	}
	index := string(indexRaw)
	projectIndex := strings.Index(index, "## project")
	evidenceIndex := strings.Index(index, "## slack_message")
	if projectIndex < 0 || evidenceIndex < 0 || projectIndex > evidenceIndex {
		t.Fatalf("index should list synthesis before evidence:\n%s", index)
	}
}

func TestRunCompanyWikiCompilerSkipsDeniedSlackEvidenceBeforeLLM(t *testing.T) {
	state := store.NewMemoryStore()
	cfg := config.Config{
		CompanyWikiRoot:               t.TempDir(),
		CompanyWikiSynthesisEnabled:   true,
		CompanyWikiSourcePageMode:     "off",
		CompanyWikiCompilerModel:      "test/model",
		CompanyWikiCompilerBatchLimit: 10,
		CompanyWikiCompilerChunkLimit: 10,
		CompanyWikiCompilerTimeout:    time.Minute,
	}
	_, err := RecordEnqueueAndMaybePublishWikiSource(context.Background(), cfg, state, store.CompanyWikiSourceRevisionInput{
		SourceType:        SlackMessageSourceType,
		DocumentSourceKey: "slack:T123:CSECRET:1777831927.000000",
		SourceSessionKey:  "slack:T123:CSECRET:thread",
		SourceRevision:    "rev-1",
		Workspace:         "T123",
		Environment:       "test",
		Title:             "Denied thread",
		Content:           "Private decision.",
		NativeLocator:     "slack:CSECRET:1777831927.000000",
		Metadata: map[string]any{
			"source":          "slack",
			"channel_id":      "CSECRET",
			"channel_private": true,
			"mirror_denied":   true,
			"mirror_allowed":  false,
		},
	})
	if err != nil {
		t.Fatalf("RecordEnqueueAndMaybePublishWikiSource() error = %v", err)
	}
	client := &fakeWikiSynthesisClient{}
	result, err := RunCompanyWikiCompiler(context.Background(), cfg, state, client)
	if err != nil {
		t.Fatalf("RunCompanyWikiCompiler() error = %v", err)
	}
	if !result.OK || result.Claimed != 1 || result.PublishedPages != 0 {
		t.Fatalf("unexpected compiler result for denied source: %+v", result)
	}
	if len(client.requests) != 0 {
		t.Fatalf("denied source should be rejected before LLM context creation, got %d requests", len(client.requests))
	}
}

func TestRunCompanyWikiCompilerBackfillsExistingSourceLedger(t *testing.T) {
	state := store.NewMemoryStore()
	cfg := testWikiCompilerConfig(t.TempDir(), "off")
	_, err := RecordWikiSourceRevision(context.Background(), cfg, state, testSlackSourceInput("1777831927.000000", "RSI Platform", "Decision: ship the wiki compiler."))
	if err != nil {
		t.Fatalf("RecordWikiSourceRevision() error = %v", err)
	}
	result, err := RunCompanyWikiCompiler(context.Background(), cfg, state, &fakeWikiSynthesisClient{})
	if err != nil {
		t.Fatalf("RunCompanyWikiCompiler() error = %v", err)
	}
	if !result.OK || result.Backfilled != 1 || result.Claimed != 1 || result.PublishedPages != 1 {
		t.Fatalf("unexpected compiler result: %+v", result)
	}
}

func TestRunCompanyWikiCompilerRecordsItemFailuresWithoutFatalError(t *testing.T) {
	state := store.NewMemoryStore()
	cfg := testWikiCompilerConfig(t.TempDir(), "off")
	recorded, err := RecordEnqueueAndMaybePublishWikiSource(context.Background(), cfg, state, testSlackSourceInput("1777831927.000000", "RSI Platform", "Decision: ship the wiki compiler."))
	if err != nil {
		t.Fatalf("RecordEnqueueAndMaybePublishWikiSource() error = %v", err)
	}
	result, err := RunCompanyWikiCompiler(context.Background(), cfg, state, &fakeWikiSynthesisClient{err: errors.New("provider temporarily rate limited")})
	if err != nil {
		t.Fatalf("RunCompanyWikiCompiler() should not return fatal error for item failure: %v", err)
	}
	if result.OK || len(result.FailedItems) != 1 || result.FailedItems[0] == "" {
		t.Fatalf("unexpected compiler result: %+v", result)
	}
	items, err := state.ClaimCompanyWikiCompileItems(store.CompanyWikiCompileClaimInput{
		Limit:              1,
		LeaseHolder:        "company_wiki_compiler:" + result.CompilerRunID,
		LeaseDuration:      time.Minute,
		CompilerVersion:    CompanyWikiCompilerVersion,
		SchemaVersion:      CompanyWikiSchemaVersion,
		RendererVersion:    CompanyWikiRendererVersion,
		ModelPolicyVersion: CompanyWikiModelPolicyVersion,
		MaxAttempts:        CompanyWikiCompileMaxAttemptCount,
	})
	if err != nil {
		t.Fatalf("ClaimCompanyWikiCompileItems() error = %v", err)
	}
	if len(items) != 1 || items[0].SourceRevisionID != recorded.Source.Revision.ID {
		t.Fatalf("failed item should remain retryable, got %+v", items)
	}
}

func TestRunCompanyWikiCompilerPreservesExistingCandidateClaims(t *testing.T) {
	state := store.NewMemoryStore()
	cfg := testWikiCompilerConfig(t.TempDir(), "off")
	if _, err := RecordEnqueueAndMaybePublishWikiSource(context.Background(), cfg, state, testSlackSourceInput("1777831927.000000", "RSI Platform", "RSI Platform already has a wiki compiler path.")); err != nil {
		t.Fatalf("record first source: %v", err)
	}
	if _, err := RunCompanyWikiCompiler(context.Background(), cfg, state, &fakeWikiSynthesisClient{}); err != nil {
		t.Fatalf("first compiler run: %v", err)
	}

	recorded, err := RecordEnqueueAndMaybePublishWikiSource(context.Background(), cfg, state, testSlackSourceInput("1777832927.000000", "RSI Platform", "RSI Platform compiler ownership belongs to Platform."))
	if err != nil {
		t.Fatalf("record second source: %v", err)
	}
	chunk := recorded.Source.Chunks[0]
	client := &fakeWikiSynthesisClient{output: WikiSynthesisOutput{Pages: []WikiSynthesisPage{{
		Slug:    "rsi-platform",
		Title:   "RSI Platform",
		Type:    "project",
		Summary: "RSI Platform status is synthesized from current and prior evidence.",
		Claims: []WikiSynthesisClaim{{
			ClaimKey:   "platform_owner",
			Text:       "RSI Platform compiler ownership belongs to Platform.",
			Confidence: 0.9,
			Citations: []store.CompanyWikiCitationInput{{
				SourceDocumentID: chunk.DocumentID,
				SourceRevisionID: chunk.RevisionID,
				ChunkID:          chunk.ID,
				NativeLocator:    chunk.NativeLocator,
				Quote:            "compiler ownership belongs to Platform",
			}},
		}},
	}}}}
	result, err := RunCompanyWikiCompiler(context.Background(), cfg, state, client)
	if err != nil {
		t.Fatalf("second compiler run: %v", err)
	}
	if !result.OK || result.PublishedPages != 1 {
		t.Fatalf("unexpected compiler result: %+v", result)
	}
	page, found, err := state.GetCompanyWikiPage("projects/rsi-platform")
	if err != nil || !found {
		t.Fatalf("GetCompanyWikiPage() found=%t err=%v", found, err)
	}
	if len(page.Claims) != 2 {
		t.Fatalf("expected preserved and new claims, got %+v", page.Claims)
	}
	for _, want := range []string{"`claim:platform_status`", "`claim:platform_owner`"} {
		if !strings.Contains(page.Revision.Body, want) {
			t.Fatalf("body missing %s:\n%s", want, page.Revision.Body)
		}
	}
}

func TestRunCompanyWikiCompilerRendersConflictCitationsAndTimestamps(t *testing.T) {
	state := store.NewMemoryStore()
	cfg := testWikiCompilerConfig(t.TempDir(), "off")
	recorded, err := RecordEnqueueAndMaybePublishWikiSource(context.Background(), cfg, state, testSlackSourceInput("1777833927.000000", "RSI Platform", "RSI Platform launch date is disputed."))
	if err != nil {
		t.Fatalf("record source: %v", err)
	}
	chunk := recorded.Source.Chunks[0]
	citation := store.CompanyWikiCitationInput{
		SourceDocumentID: chunk.DocumentID,
		SourceRevisionID: chunk.RevisionID,
		ChunkID:          chunk.ID,
		NativeLocator:    chunk.NativeLocator,
		Quote:            "launch date is disputed",
	}
	client := &fakeWikiSynthesisClient{output: WikiSynthesisOutput{Pages: []WikiSynthesisPage{{
		Slug:    "rsi-platform",
		Title:   "RSI Platform",
		Type:    "project",
		Summary: "RSI Platform has disputed launch timing evidence.",
		Claims: []WikiSynthesisClaim{{
			ClaimKey:   "launch_date",
			Text:       "RSI Platform launch date is disputed.",
			Confidence: 0.6,
			Citations:  []store.CompanyWikiCitationInput{citation},
		}},
		Conflicts: []WikiSynthesisConflict{{
			ClaimKey:  "launch_date",
			Summary:   "Sources disagree on the launch date.",
			Citations: []store.CompanyWikiCitationInput{citation},
		}},
	}}}}
	if _, err := RunCompanyWikiCompiler(context.Background(), cfg, state, client); err != nil {
		t.Fatalf("RunCompanyWikiCompiler() error = %v", err)
	}
	page, found, err := state.GetCompanyWikiPage("projects/rsi-platform")
	if err != nil || !found {
		t.Fatalf("GetCompanyWikiPage() found=%t err=%v", found, err)
	}
	for _, want := range []string{"## Conflicts", "conflict citation:", "source_timestamp=2026-05-03T"} {
		if !strings.Contains(page.Revision.Body, want) {
			t.Fatalf("conflict body missing %q:\n%s", want, page.Revision.Body)
		}
	}
	if len(page.Conflicts) != 1 || len(page.Conflicts[0].Metadata) == 0 {
		t.Fatalf("expected persisted conflict metadata, got %+v", page.Conflicts)
	}
}

func TestCompanyWikiCompilerLeaseDurationCoversSequentialBatch(t *testing.T) {
	got := companyWikiCompilerLeaseDuration(config.Config{
		CompanyWikiCompilerBatchLimit: 10,
		CompanyWikiCompilerTimeout:    2 * time.Minute,
	})
	want := 21 * time.Minute
	if got != want {
		t.Fatalf("lease duration = %s, want %s", got, want)
	}
}

func TestWikiIndexCategoryRankKeepsManualAfterSynthesis(t *testing.T) {
	if wikiIndexCategoryRank("manual", "pages/manual/foo.md") <= wikiIndexCategoryRank("concept", "pages/concepts/foo.md") {
		t.Fatal("manual pages should sort after synthesis concept pages")
	}
	if wikiIndexCategoryRank("manual", "pages/manual/foo.md") >= wikiIndexCategoryRank("slack_message", "sources/slack/foo.md") {
		t.Fatal("manual pages should still sort before evidence pages")
	}
}

func testWikiCompilerConfig(root string, sourcePageMode string) config.Config {
	return config.Config{
		CompanyWikiRoot:               root,
		CompanyWikiSynthesisEnabled:   true,
		CompanyWikiSourcePageMode:     sourcePageMode,
		CompanyWikiCompilerModel:      "test/model",
		CompanyWikiCompilerBatchLimit: 10,
		CompanyWikiCompilerChunkLimit: 10,
		CompanyWikiCompilerTimeout:    time.Minute,
	}
}

func testSlackSourceInput(ts string, title string, content string) store.CompanyWikiSourceRevisionInput {
	return store.CompanyWikiSourceRevisionInput{
		SourceType:        SlackMessageSourceType,
		DocumentSourceKey: "slack:T123:C123:" + ts,
		SourceSessionKey:  "slack:T123:C123:thread",
		SourceRevision:    "rev-" + ts,
		Workspace:         "T123",
		Environment:       "test",
		Title:             title,
		URL:               "https://slack.example/archives/C123/p" + strings.ReplaceAll(ts, ".", ""),
		Content:           content,
		NativeLocator:     "slack:C123:" + ts,
		Metadata: map[string]any{
			"source":           "slack",
			"channel_id":       "C123",
			"channel_type":     "public_channel",
			"channel_private":  false,
			"channel_im":       false,
			"mirror_discovery": "joined_public",
			"mirror_allowed":   true,
			"mirror_denied":    false,
		},
		ObservedAt: time.Unix(1777833927, 0).UTC(),
	}
}
