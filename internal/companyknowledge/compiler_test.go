package companyknowledge

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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

type cancelingWikiSynthesisClient struct {
	fakeWikiSynthesisClient
	cancel context.CancelFunc
}

func (f *cancelingWikiSynthesisClient) SynthesizeWiki(ctx context.Context, request WikiSynthesisRequest) (WikiSynthesisOutput, WikiSynthesisMetadata, error) {
	output, metadata, err := f.fakeWikiSynthesisClient.SynthesizeWiki(ctx, request)
	f.cancel()
	return output, metadata, err
}

type sequenceWikiSynthesisClient struct {
	requests []WikiSynthesisRequest
	outputs  []WikiSynthesisOutput
}

func (f *sequenceWikiSynthesisClient) SynthesizeWiki(ctx context.Context, request WikiSynthesisRequest) (WikiSynthesisOutput, WikiSynthesisMetadata, error) {
	_ = ctx
	f.requests = append(f.requests, request)
	index := len(f.requests) - 1
	if index >= len(f.outputs) {
		index = len(f.outputs) - 1
	}
	if index < 0 {
		return WikiSynthesisOutput{}, WikiSynthesisMetadata{RequestMetadataHash: "reqhash", ResponseMetadataHash: "resphash"}, nil
	}
	return f.outputs[index], WikiSynthesisMetadata{RequestMetadataHash: "reqhash", ResponseMetadataHash: "resphash"}, nil
}

func TestWikiSynthesisOutputAcceptsStructuredFreshness(t *testing.T) {
	raw := `{"pages":[{"slug":"status","title":"Status","type":"open_question","summary":"Status summary.","freshness":{"source_timestamp":"2026-04-28T06:23:43Z"},"claims":[{"claim_key":"status","text":"Status is available.","confidence":1,"citations":[]}]}]}`
	var output WikiSynthesisOutput
	if err := json.Unmarshal([]byte(raw), &output); err != nil {
		t.Fatalf("json.Unmarshal() should tolerate structured freshness: %v", err)
	}
	if len(output.Pages) != 1 || output.Pages[0].Title != "Status" {
		t.Fatalf("unexpected output: %+v", output)
	}
}

func TestWikiSynthesisOutputAcceptsFlexibleModelShapes(t *testing.T) {
	raw := `{"pages":[{"slug":"status","title":"Status","type":"project","tags":"rsi","summary":"Status summary.","owners":[{"name":"Platform"}],"open_questions":{"questions":[{"question":"Who owns rollout?"}]},"related_pages":{"pages":["systems/wiki"]},"claims":[{"claim_key":"status","claim_text":"Status is available.","confidence":1,"citations":null,"citation":{"source_document_id":"doc","source_revision_id":"rev","source_chunk_id":"chunk","quote":"Status"}}],"conflicts":[{"claim_key":"status","text":"Sources need review.","citations":null,"citation":{"source_document_id":"doc","source_revision_id":"rev","chunk_id":"chunk"}}]}]}`
	var output WikiSynthesisOutput
	if err := json.Unmarshal([]byte(raw), &output); err != nil {
		t.Fatalf("json.Unmarshal() should tolerate common model shape variations: %v", err)
	}
	page := output.Pages[0]
	if got := strings.Join(page.Tags, ","); got != "rsi" {
		t.Fatalf("tags = %q", got)
	}
	if got := strings.Join(page.Owners, ","); got != "Platform" {
		t.Fatalf("owners = %q", got)
	}
	if got := strings.Join(page.OpenQuestions, ","); got != "Who owns rollout?" {
		t.Fatalf("open_questions = %q", got)
	}
	if page.Claims[0].Text != "Status is available." || page.Claims[0].Citations[0].ChunkID != "chunk" {
		t.Fatalf("claim not normalized: %+v", page.Claims[0])
	}
	if page.Conflicts[0].Summary != "Sources need review." || page.Conflicts[0].Citations[0].ChunkID != "chunk" {
		t.Fatalf("conflict not normalized: %+v", page.Conflicts[0])
	}
}

func TestValidateSynthesisOutputDropsEmptyPlaceholderClaims(t *testing.T) {
	chunk := store.CompanyWikiSourceChunk{
		ID:         "chunk",
		DocumentID: "doc",
		RevisionID: "rev",
		Content:    "RSI Platform has a compiler.",
	}
	evidence := store.CompanyWikiSourceEvidence{
		Document: store.CompanyWikiSourceDocument{ID: "doc"},
		Revision: store.CompanyWikiSourceRevision{ID: "rev", DocumentID: "doc"},
		Chunks:   []store.CompanyWikiSourceChunk{chunk},
	}
	output := WikiSynthesisOutput{Pages: []WikiSynthesisPage{{
		Slug:    "rsi-platform",
		Title:   "RSI Platform",
		Type:    "project",
		Summary: "RSI Platform status.",
		Claims: []WikiSynthesisClaim{
			{},
			{
				ClaimKey: "compiler",
				Text:     "RSI Platform has a compiler.",
				Citations: []store.CompanyWikiCitationInput{{
					SourceDocumentID: "doc",
					SourceRevisionID: "rev",
					ChunkID:          "chunk",
				}},
			},
		},
	}}}
	pages, validationErrors := validateSynthesisOutput(evidence, nil, output)
	if len(validationErrors) > 0 {
		t.Fatalf("validateSynthesisOutput() errors = %+v", validationErrors)
	}
	if len(pages) != 1 || len(pages[0].Claims) != 1 || pages[0].Claims[0].ClaimKey != "compiler" {
		t.Fatalf("placeholder claim should be dropped, got %+v", pages)
	}
}

func TestValidateSynthesisOutputResolvesCitationIDs(t *testing.T) {
	chunk := store.CompanyWikiSourceChunk{
		ID:            "chunk",
		DocumentID:    "doc",
		RevisionID:    "rev",
		Content:       "Adaptive Trust Screening gates high-risk workflows.",
		NativeLocator: "notion:block:1",
	}
	evidence := store.CompanyWikiSourceEvidence{
		Document: store.CompanyWikiSourceDocument{ID: "doc"},
		Revision: store.CompanyWikiSourceRevision{ID: "rev", DocumentID: "doc"},
		Chunks:   []store.CompanyWikiSourceChunk{chunk},
	}
	output := WikiSynthesisOutput{Pages: []WikiSynthesisPage{{
		Slug:    "adaptive-trust-screening",
		Title:   "Adaptive Trust Screening",
		Type:    "system",
		Summary: "Adaptive Trust Screening status.",
		Claims: []WikiSynthesisClaim{{
			ClaimKey:    "adaptive_trust_gate",
			Text:        "Adaptive Trust Screening gates high-risk workflows.",
			CitationIDs: []string{"chunk"},
		}},
		Conflicts: []WikiSynthesisConflict{{
			ClaimKey:    "adaptive_trust_gate",
			Summary:     "Evidence should be reviewed during rollout.",
			CitationIDs: []string{"chunk"},
		}},
	}}}
	pages, validationErrors := validateSynthesisOutput(evidence, nil, output)
	if len(validationErrors) > 0 {
		t.Fatalf("validateSynthesisOutput() errors = %+v", validationErrors)
	}
	if len(pages) != 1 || len(pages[0].Claims) != 1 || len(pages[0].Claims[0].Citations) != 1 {
		t.Fatalf("claim citation_ids should resolve to citations, got %+v", pages)
	}
	if got := pages[0].Claims[0].Citations[0]; got.ChunkID != "chunk" || got.NativeLocator != "notion:block:1" {
		t.Fatalf("resolved claim citation = %+v", got)
	}
	if len(pages[0].Conflicts) != 1 || len(pages[0].Conflicts[0].Citations) != 1 {
		t.Fatalf("conflict citation_ids should resolve to citations, got %+v", pages[0].Conflicts)
	}
}

func TestOpenRouterWikiSynthesisClientRetriesMalformedJSON(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		content := `{"pages":[`
		if attempts == 2 {
			content = `{"pages":[{"slug":"status","title":"Status","type":"project","summary":"Status summary.","claims":[{"claim_key":"status","text":"Status is available.","confidence":1,"citations":[]}]}]}`
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":    "chatcmpl-test",
			"model": "test/model",
			"choices": []map[string]any{{
				"message": map[string]string{"content": content},
			}},
			"usage": map[string]any{"prompt_tokens": 1},
		})
	}))
	defer server.Close()
	client := &OpenRouterWikiSynthesisClient{BaseURL: server.URL, APIKey: "test", Client: server.Client()}
	output, _, err := client.SynthesizeWiki(context.Background(), WikiSynthesisRequest{
		Model: "test/model",
		Source: store.CompanyWikiSourceEvidence{
			Revision: store.CompanyWikiSourceRevision{ID: "rev"},
		},
	})
	if err != nil {
		t.Fatalf("SynthesizeWiki() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if len(output.Pages) != 1 || output.Pages[0].Title != "Status" {
		t.Fatalf("unexpected output: %+v", output)
	}
}

func TestOpenAICompatibleWikiSynthesisClientDeepSeekThinkingPayload(t *testing.T) {
	var captured map[string]any
	var capturedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":    "chatcmpl-test",
			"model": "deepseek-v4-pro",
			"choices": []map[string]any{{
				"message": map[string]string{"content": `{"pages":[{"slug":"status","title":"Status","type":"project","summary":"Status summary.","claims":[]}]}`},
			}},
			"usage": map[string]any{"prompt_tokens": 1},
		})
	}))
	defer server.Close()

	client := &OpenAICompatibleWikiSynthesisClient{
		Provider:        "deepseek",
		BaseURL:         server.URL + "/v1",
		APIKey:          "test",
		Thinking:        "enabled",
		ReasoningEffort: "xhigh",
		Client:          server.Client(),
	}
	_, _, err := client.SynthesizeWiki(context.Background(), WikiSynthesisRequest{
		Model: "deepseek-v4-pro",
		Source: store.CompanyWikiSourceEvidence{
			Revision: store.CompanyWikiSourceRevision{ID: "rev"},
		},
	})
	if err != nil {
		t.Fatalf("SynthesizeWiki() error = %v", err)
	}
	if capturedPath != "/v1/chat/completions" {
		t.Fatalf("path = %q, want /v1/chat/completions", capturedPath)
	}
	thinking, _ := captured["thinking"].(map[string]any)
	if thinking["type"] != "enabled" {
		t.Fatalf("thinking = %+v, want enabled", thinking)
	}
	if captured["reasoning_effort"] != "max" {
		t.Fatalf("reasoning_effort = %v, want max", captured["reasoning_effort"])
	}
	for _, key := range []string{"temperature", "top_p", "presence_penalty", "frequency_penalty"} {
		if _, ok := captured[key]; ok {
			t.Fatalf("payload unexpectedly included %s: %+v", key, captured)
		}
	}
}

func TestOpenAICompatibleWikiSynthesisClientDeepSeekDisabledThinkingPayload(t *testing.T) {
	var captured map[string]any
	var capturedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":    "chatcmpl-test",
			"model": "deepseek-v4-flash",
			"choices": []map[string]any{{
				"message": map[string]string{"content": `{"pages":[{"slug":"status","title":"Status","type":"project","summary":"Status summary.","claims":[]}]}`},
			}},
		})
	}))
	defer server.Close()

	client := &OpenAICompatibleWikiSynthesisClient{
		Provider:        "deepseek",
		BaseURL:         server.URL,
		APIKey:          "test",
		Thinking:        "disabled",
		ReasoningEffort: "none",
		Client:          server.Client(),
	}
	_, _, err := client.SynthesizeWiki(context.Background(), WikiSynthesisRequest{
		Model: "deepseek-v4-flash",
		Source: store.CompanyWikiSourceEvidence{
			Revision: store.CompanyWikiSourceRevision{ID: "rev"},
		},
	})
	if err != nil {
		t.Fatalf("SynthesizeWiki() error = %v", err)
	}
	if capturedPath != "/chat/completions" {
		t.Fatalf("path = %q, want /chat/completions", capturedPath)
	}
	thinking, _ := captured["thinking"].(map[string]any)
	if thinking["type"] != "disabled" {
		t.Fatalf("thinking = %+v, want disabled", thinking)
	}
	if _, ok := captured["reasoning_effort"]; ok {
		t.Fatalf("disabled thinking should omit reasoning_effort: %+v", captured)
	}
	if captured["temperature"] != float64(0) {
		t.Fatalf("temperature = %v, want 0 for disabled thinking", captured["temperature"])
	}
}

func TestRunCompanyWikiCompilerRetriesValidationFailure(t *testing.T) {
	state := store.NewMemoryStore()
	cfg := testWikiCompilerConfig(t.TempDir(), "off")
	recorded, err := RecordEnqueueAndMaybePublishWikiSource(context.Background(), cfg, state, testSlackSourceInput("1777831927.000000", "Adaptive Trust Screening", "Decision: Adaptive Trust Screening should gate high-risk workflows."))
	if err != nil {
		t.Fatalf("RecordEnqueueAndMaybePublishWikiSource() error = %v", err)
	}
	chunk := recorded.Source.Chunks[0]
	client := &sequenceWikiSynthesisClient{outputs: []WikiSynthesisOutput{
		{},
		{Pages: []WikiSynthesisPage{{
			Slug:    "adaptive-trust-screening",
			Title:   "Adaptive Trust Screening",
			Type:    "system",
			Summary: "Adaptive Trust Screening gates high-risk workflows.",
			Claims: []WikiSynthesisClaim{{
				ClaimKey: "adaptive_trust_gate",
				Text:     "Adaptive Trust Screening should gate high-risk workflows.",
				Citations: []store.CompanyWikiCitationInput{{
					SourceDocumentID: chunk.DocumentID,
					SourceRevisionID: chunk.RevisionID,
					ChunkID:          chunk.ID,
				}},
			}},
		}}},
	}}
	result, err := RunCompanyWikiCompiler(context.Background(), cfg, state, client)
	if err != nil {
		t.Fatalf("RunCompanyWikiCompiler() error = %v", err)
	}
	if !result.OK || result.PublishedPages != 1 {
		t.Fatalf("unexpected compiler result: %+v", result)
	}
	if len(client.requests) != 2 {
		t.Fatalf("requests = %d, want 2", len(client.requests))
	}
	if len(client.requests[1].PreviousValidationErrors) == 0 || !strings.Contains(client.requests[1].PreviousValidationErrors[0], "at least one page") {
		t.Fatalf("second request should include validation repair context: %+v", client.requests[1].PreviousValidationErrors)
	}
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

func TestRunCompanyWikiCompilerCapsClaimLimitByRunBudget(t *testing.T) {
	state := store.NewMemoryStore()
	cfg := testWikiCompilerConfig(t.TempDir(), "off")
	cfg.CompanyWikiCompilerBatchLimit = 10
	cfg.CompanyWikiCompilerTimeout = 2 * time.Minute
	cfg.CompanyWikiCompilerRunTimeout = 5 * time.Minute
	cfg.CompanyWikiCompilerShutdownGrace = 30 * time.Second
	for i, ts := range []string{"1777831927.000000", "1777831928.000000", "1777831929.000000"} {
		_, err := RecordEnqueueAndMaybePublishWikiSource(context.Background(), cfg, state, testSlackSourceInput(ts, "RSI Platform", "Decision: ship wiki compiler budget guard item "+string(rune('A'+i))+"."))
		if err != nil {
			t.Fatalf("record source %d: %v", i, err)
		}
	}
	client := &fakeWikiSynthesisClient{}
	result, err := RunCompanyWikiCompiler(context.Background(), cfg, state, client)
	if err != nil {
		t.Fatalf("RunCompanyWikiCompiler() error = %v", err)
	}
	if result.Claimed != 1 {
		t.Fatalf("Claimed = %d, want 1; result=%+v", result.Claimed, result)
	}
	if len(client.requests) != 1 {
		t.Fatalf("Synthesize requests = %d, want 1", len(client.requests))
	}
}

func TestCompanyWikiCompilerClaimLimitAccountsForRetries(t *testing.T) {
	got := companyWikiCompilerClaimLimit(config.Config{
		CompanyWikiCompilerBatchLimit:    10,
		CompanyWikiCompilerTimeout:       2 * time.Minute,
		CompanyWikiCompilerRunTimeout:    25 * time.Minute,
		CompanyWikiCompilerShutdownGrace: 30 * time.Second,
	})
	if got != 2 {
		t.Fatalf("claim limit = %d, want 2", got)
	}
}

func TestRunCompanyWikiCompilerReleasesDeferredClaims(t *testing.T) {
	state := store.NewMemoryStore()
	cfg := testWikiCompilerConfig(t.TempDir(), "off")
	cfg.CompanyWikiCompilerBatchLimit = 3
	cfg.CompanyWikiCompilerRunTimeout = 0
	for i, ts := range []string{"1777831930.000000", "1777831931.000000", "1777831932.000000"} {
		_, err := RecordEnqueueAndMaybePublishWikiSource(context.Background(), cfg, state, testSlackSourceInput(ts, "RSI Platform", "Decision: release deferred compiler claim item "+string(rune('A'+i))+"."))
		if err != nil {
			t.Fatalf("record source %d: %v", i, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	client := &cancelingWikiSynthesisClient{cancel: cancel}
	result, err := RunCompanyWikiCompiler(ctx, cfg, state, client)
	if err != nil {
		t.Fatalf("RunCompanyWikiCompiler() error = %v", err)
	}
	if result.DeferredItems != 2 || result.StoppedReason != context.Canceled.Error() {
		t.Fatalf("deferred result = %+v, want 2 deferred with context canceled", result)
	}
	if len(client.requests) != 1 {
		t.Fatalf("Synthesize requests = %d, want 1", len(client.requests))
	}
	reclaimed, err := state.ClaimCompanyWikiCompileItems(store.CompanyWikiCompileClaimInput{
		Limit:              10,
		LeaseHolder:        "test-reclaimer",
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
	if len(reclaimed) != 2 {
		t.Fatalf("reclaimed = %d, want 2; items=%+v", len(reclaimed), reclaimed)
	}
	for _, item := range reclaimed {
		if item.AttemptCount != 1 {
			t.Fatalf("reclaimed item %s attempt_count = %d, want 1 after release/reclaim", item.ID, item.AttemptCount)
		}
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

func TestRunCompanyWikiCompilerDoesNotPublishSourceShapedSynthesisPage(t *testing.T) {
	state := store.NewMemoryStore()
	root := t.TempDir()
	cfg := testWikiCompilerConfig(root, "off")
	recorded, err := RecordEnqueueAndMaybePublishWikiSource(context.Background(), cfg, state, testSlackSourceInput("1777831927.000000", "Slack thread C123", "Decision: add Sam to the admin allowlist."))
	if err != nil {
		t.Fatalf("RecordEnqueueAndMaybePublishWikiSource() error = %v", err)
	}
	chunk := recorded.Source.Chunks[0]
	client := &fakeWikiSynthesisClient{output: WikiSynthesisOutput{Pages: []WikiSynthesisPage{{
		Slug:    "slack_message/slack-thread-c123",
		Title:   "Slack thread C123",
		Type:    "decision",
		Summary: "Sam should be added to the admin allowlist.",
		Claims: []WikiSynthesisClaim{{
			ClaimKey:   "sam_admin_allowlist",
			Text:       "Sam should be added to the admin allowlist.",
			Confidence: 0.9,
			Citations: []store.CompanyWikiCitationInput{{
				SourceDocumentID: chunk.DocumentID,
				SourceRevisionID: chunk.RevisionID,
				ChunkID:          chunk.ID,
				NativeLocator:    chunk.NativeLocator,
				Quote:            "add Sam to the admin allowlist",
			}},
		}},
	}}}}
	result, err := RunCompanyWikiCompiler(context.Background(), cfg, state, client)
	if err != nil {
		t.Fatalf("RunCompanyWikiCompiler() error = %v", err)
	}
	if !result.OK || result.PublishedPages != 1 {
		t.Fatalf("unexpected compiler result: %+v", result)
	}
	if _, found, err := state.GetCompanyWikiPage("slack_message/slack-thread-c123"); err != nil || found {
		t.Fatalf("source-shaped synthesis page found=%t err=%v", found, err)
	}
	page, found, err := state.GetCompanyWikiPage("decisions/slack-thread-c123")
	if err != nil || !found {
		t.Fatalf("semantic synthesis page found=%t err=%v", found, err)
	}
	if page.Revision.Path != "pages/decisions/slack-thread-c123.md" {
		t.Fatalf("semantic synthesis path = %q", page.Revision.Path)
	}
	if _, err := os.Stat(filepath.Join(root, page.Revision.Path)); err != nil {
		t.Fatalf("expected semantic synthesis file: %v", err)
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

func TestPreserveCandidateClaimsKeepsMateriallyFresherSameKeyAndAllowsMissingFacts(t *testing.T) {
	oldTS := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	newTS := oldTS.Add(48 * time.Hour)
	evidence := store.CompanyWikiSourceEvidence{
		Document: store.CompanyWikiSourceDocument{ID: "slack-doc", SourceType: SlackMessageSourceType},
		Revision: store.CompanyWikiSourceRevision{
			ID:         "slack-old-revision",
			DocumentID: "slack-doc",
			ObservedAt: oldTS,
		},
	}
	candidate := store.CompanyWikiPageRead{
		Page: store.CompanyWikiPage{Slug: "projects/rsi-platform", Title: "RSI Platform"},
		Claims: []store.CompanyWikiClaim{{
			ClaimKey:   "project_status",
			ClaimText:  "Notion says RSI Platform is the project source of truth.",
			Confidence: 0.95,
			Metadata: map[string]any{"citation_refs": []map[string]string{{
				"source_revision_id": "notion-new-revision",
				"source_timestamp":   newTS.Format(time.RFC3339),
			}}},
		}},
		Citations: []store.CompanyWikiCitation{{
			ClaimKey:         "project_status",
			SourceDocumentID: "notion-doc",
			SourceRevisionID: "notion-new-revision",
			ChunkID:          "notion-chunk",
			NativeLocator:    "notion:block",
			Quote:            "project source of truth",
		}},
	}
	output := WikiSynthesisOutput{Pages: []WikiSynthesisPage{{
		Slug:    "rsi-platform",
		Title:   "RSI Platform",
		Type:    "project",
		Summary: "RSI Platform status.",
		Claims: []WikiSynthesisClaim{
			{
				ClaimKey:   "project_status",
				Text:       "Older Slack says RSI Platform status was still uncertain.",
				Confidence: 0.8,
				Citations: []store.CompanyWikiCitationInput{{
					SourceDocumentID: "slack-doc",
					SourceRevisionID: "slack-old-revision",
					ChunkID:          "slack-chunk",
				}},
			},
			{
				ClaimKey:   "slack_only_gap",
				Text:       "Older Slack captured a missing operational detail.",
				Confidence: 0.8,
				Citations: []store.CompanyWikiCitationInput{{
					SourceDocumentID: "slack-doc",
					SourceRevisionID: "slack-old-revision",
					ChunkID:          "slack-chunk",
				}},
			},
		},
	}}}

	got := preserveCandidateClaimsInSynthesisOutput(output, evidence, []store.CompanyWikiPageRead{candidate})
	if len(got.Pages) != 1 || len(got.Pages[0].Claims) != 2 {
		t.Fatalf("unexpected preserved output: %+v", got)
	}
	if got.Pages[0].Claims[0].Text != candidate.Claims[0].ClaimText {
		t.Fatalf("materially fresher candidate claim should win same claim key, got %q", got.Pages[0].Claims[0].Text)
	}
	if got.Pages[0].Claims[1].ClaimKey != "slack_only_gap" {
		t.Fatalf("older Slack-only missing fact should remain, claims=%+v", got.Pages[0].Claims)
	}
	if freshness := synthesisFreshness(buildRevisionTimestampMap(evidence, []store.CompanyWikiPageRead{candidate}), got.Pages[0]); freshness != newTS.Format(time.RFC3339) {
		t.Fatalf("freshness should use newest cited source timestamp, got %q want %q", freshness, newTS.Format(time.RFC3339))
	}
}

func TestPreserveCandidateClaimsAllowsCloseTimestampModelJudgment(t *testing.T) {
	oldTS := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	newTS := oldTS.Add(time.Hour)
	evidence := store.CompanyWikiSourceEvidence{
		Document: store.CompanyWikiSourceDocument{ID: "slack-doc", SourceType: SlackMessageSourceType},
		Revision: store.CompanyWikiSourceRevision{
			ID:         "slack-close-revision",
			DocumentID: "slack-doc",
			ObservedAt: oldTS,
		},
	}
	candidate := store.CompanyWikiPageRead{
		Page: store.CompanyWikiPage{Slug: "projects/rsi-platform"},
		Claims: []store.CompanyWikiClaim{{
			ClaimKey:  "project_status",
			ClaimText: "Near-time candidate status.",
			Metadata: map[string]any{"citation_refs": []map[string]string{{
				"source_revision_id": "notion-close-revision",
				"source_timestamp":   newTS.Format(time.RFC3339),
			}}},
		}},
		Citations: []store.CompanyWikiCitation{{
			ClaimKey:         "project_status",
			SourceDocumentID: "notion-doc",
			SourceRevisionID: "notion-close-revision",
			ChunkID:          "notion-chunk",
		}},
	}
	output := WikiSynthesisOutput{Pages: []WikiSynthesisPage{{
		Slug:    "rsi-platform",
		Title:   "RSI Platform",
		Type:    "project",
		Summary: "RSI Platform status.",
		Claims: []WikiSynthesisClaim{{
			ClaimKey: "project_status",
			Text:     "Near-time model judgment can supersede or conflict.",
			Citations: []store.CompanyWikiCitationInput{{
				SourceDocumentID: "slack-doc",
				SourceRevisionID: "slack-close-revision",
				ChunkID:          "slack-chunk",
			}},
		}},
	}}}
	got := preserveCandidateClaimsInSynthesisOutput(output, evidence, []store.CompanyWikiPageRead{candidate})
	if got.Pages[0].Claims[0].Text != "Near-time model judgment can supersede or conflict." {
		t.Fatalf("near-time evidence should be left to model conflict/supersession policy, got %+v", got.Pages[0].Claims[0])
	}
}

func TestRevisionTimestampMapIncludesAllSlackEvidenceChunks(t *testing.T) {
	rootTS := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	replyTS := rootTS.Add(2 * time.Hour)
	evidence := store.CompanyWikiSourceEvidence{
		Document: store.CompanyWikiSourceDocument{ID: "slack-doc", SourceType: SlackMessageSourceType},
		Revision: store.CompanyWikiSourceRevision{
			ID:         "slack-root-revision",
			DocumentID: "slack-doc",
			ObservedAt: rootTS,
		},
		Chunks: []store.CompanyWikiSourceChunk{
			{
				ID:         "root-chunk",
				DocumentID: "slack-doc",
				RevisionID: "slack-root-revision",
				Metadata:   map[string]any{"slack_ts": "1777636800.000000"},
			},
			{
				ID:         "reply-chunk",
				DocumentID: "slack-doc",
				RevisionID: "slack-reply-revision",
				Metadata:   map[string]any{"source_observed_at": replyTS.Format(time.RFC3339)},
			},
		},
	}
	timestamps := buildRevisionTimestampMap(evidence, nil)
	if got := timestamps["slack-root-revision"]; got != rootTS.Format(time.RFC3339) {
		t.Fatalf("root revision timestamp = %q, want %q", got, rootTS.Format(time.RFC3339))
	}
	if got := timestamps["slack-reply-revision"]; got != replyTS.Format(time.RFC3339) {
		t.Fatalf("reply revision timestamp = %q, want %q", got, replyTS.Format(time.RFC3339))
	}
	page := WikiSynthesisPage{Claims: []WikiSynthesisClaim{{
		ClaimKey: "reply_fact",
		Text:     "A reply carries the latest fact.",
		Citations: []store.CompanyWikiCitationInput{{
			SourceDocumentID: "slack-doc",
			SourceRevisionID: "slack-reply-revision",
			ChunkID:          "reply-chunk",
		}},
	}}}
	if freshness := synthesisFreshness(timestamps, page); freshness != replyTS.Format(time.RFC3339) {
		t.Fatalf("freshness should use cited reply revision timestamp, got %q want %q", freshness, replyTS.Format(time.RFC3339))
	}
}

func TestRevisionTimestampMapPrefersNotionLastEditedOverCrawlObservedAt(t *testing.T) {
	lastEdited := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	crawled := lastEdited.Add(48 * time.Hour)
	evidence := store.CompanyWikiSourceEvidence{
		Document: store.CompanyWikiSourceDocument{ID: "notion-doc", SourceType: NotionDocumentSourceType},
		Revision: store.CompanyWikiSourceRevision{
			ID:         "notion-revision",
			DocumentID: "notion-doc",
			ObservedAt: crawled,
			Metadata: map[string]any{
				"last_edited_time": lastEdited.Format(time.RFC3339),
				"created_time":     lastEdited.Add(-24 * time.Hour).Format(time.RFC3339),
			},
		},
		Chunks: []store.CompanyWikiSourceChunk{{
			ID:         "notion-chunk",
			DocumentID: "notion-doc",
			RevisionID: "notion-revision",
			Metadata: map[string]any{
				"source_observed_at": crawled.Format(time.RFC3339),
				"last_edited_time":   lastEdited.Format(time.RFC3339),
			},
		}},
	}

	timestamps := buildRevisionTimestampMap(evidence, nil)
	if got := timestamps["notion-revision"]; got != lastEdited.Format(time.RFC3339) {
		t.Fatalf("notion timestamp = %q, want source last edited %q", got, lastEdited.Format(time.RFC3339))
	}
}

func TestPreserveCandidateClaimsDoesNotOverwriteWhenModelTimestampUnknown(t *testing.T) {
	candidateTS := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	evidence := store.CompanyWikiSourceEvidence{
		Document: store.CompanyWikiSourceDocument{ID: "slack-doc", SourceType: SlackMessageSourceType},
		Revision: store.CompanyWikiSourceRevision{
			ID:         "different-revision",
			DocumentID: "slack-doc",
			ObservedAt: candidateTS.Add(-48 * time.Hour),
		},
		Chunks: []store.CompanyWikiSourceChunk{{
			ID:         "unknown-ts-chunk",
			DocumentID: "slack-doc",
			RevisionID: "slack-unknown-revision",
		}},
	}
	candidate := store.CompanyWikiPageRead{
		Page: store.CompanyWikiPage{Slug: "projects/rsi-platform"},
		Claims: []store.CompanyWikiClaim{{
			ClaimKey:  "project_status",
			ClaimText: "Candidate status has a timestamp.",
			Metadata: map[string]any{"citation_refs": []map[string]string{{
				"source_revision_id": "notion-new-revision",
				"source_timestamp":   candidateTS.Format(time.RFC3339),
			}}},
		}},
		Citations: []store.CompanyWikiCitation{{
			ClaimKey:         "project_status",
			SourceDocumentID: "notion-doc",
			SourceRevisionID: "notion-new-revision",
			ChunkID:          "notion-chunk",
		}},
	}
	output := WikiSynthesisOutput{Pages: []WikiSynthesisPage{{
		Slug:    "rsi-platform",
		Title:   "RSI Platform",
		Type:    "project",
		Summary: "RSI Platform status.",
		Claims: []WikiSynthesisClaim{{
			ClaimKey: "project_status",
			Text:     "Model synthesis with an unknown source timestamp should not be deterministically overwritten.",
			Citations: []store.CompanyWikiCitationInput{{
				SourceDocumentID: "slack-doc",
				SourceRevisionID: "slack-unknown-revision",
				ChunkID:          "unknown-ts-chunk",
			}},
		}},
	}}}
	got := preserveCandidateClaimsInSynthesisOutput(output, evidence, []store.CompanyWikiPageRead{candidate})
	if got.Pages[0].Claims[0].Text != output.Pages[0].Claims[0].Text {
		t.Fatalf("unknown model timestamp should avoid deterministic overwrite, got %+v", got.Pages[0].Claims[0])
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
	want := 10*(4*2*time.Minute+15*time.Second) + time.Minute
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

func TestSynthesisSlugNormalizesSourceShapedLLMRoots(t *testing.T) {
	cases := []struct {
		name string
		page WikiSynthesisPage
		want string
	}{
		{
			name: "slack message root becomes semantic decision root",
			page: WikiSynthesisPage{Slug: "slack_message/slack-thread-c123", Type: "decision"},
			want: "decisions/slack-thread-c123",
		},
		{
			name: "notion document root becomes semantic project root",
			page: WikiSynthesisPage{Slug: "notion_document/rsi-platform-plan", Type: "project"},
			want: "projects/rsi-platform-plan",
		},
		{
			name: "sources root is stripped",
			page: WikiSynthesisPage{Slug: "sources/slack/admin-allowlist", Type: "policy"},
			want: "policies/admin-allowlist",
		},
		{
			name: "existing semantic root is preserved",
			page: WikiSynthesisPage{Slug: "projects/rsi-platform", Type: "project"},
			want: "projects/rsi-platform",
		},
		{
			name: "plain title gets semantic root",
			page: WikiSynthesisPage{Title: "RSI Platform", Type: "project"},
			want: "projects/rsi-platform",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := synthesisSlug(tc.page); got != tc.want {
				t.Fatalf("synthesisSlug() = %q, want %q", got, tc.want)
			}
		})
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
