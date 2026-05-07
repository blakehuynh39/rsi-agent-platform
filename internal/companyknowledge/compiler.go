package companyknowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

type WikiCompilerResult struct {
	OK             bool     `json:"ok"`
	CompilerRunID  string   `json:"compiler_run_id"`
	Backfilled     int      `json:"backfilled"`
	Claimed        int      `json:"claimed"`
	PublishedPages int      `json:"published_pages"`
	FailedItems    []string `json:"failed_items,omitempty"`
	DeferredItems  int      `json:"deferred_items,omitempty"`
	StoppedReason  string   `json:"stopped_reason,omitempty"`
}

type WikiSynthesisClient interface {
	SynthesizeWiki(ctx context.Context, request WikiSynthesisRequest) (WikiSynthesisOutput, WikiSynthesisMetadata, error)
}

const (
	companyWikiCompilerValidationAttempts     = 2
	companyWikiCompilerProviderDecodeAttempts = 2
	companyWikiCompilerPerItemOverhead        = 15 * time.Second
)

type WikiSynthesisRequest struct {
	Model                    string                          `json:"model"`
	Source                   store.CompanyWikiSourceEvidence `json:"source"`
	Chunks                   []store.CompanyWikiSourceChunk  `json:"chunks"`
	CandidatePages           []store.CompanyWikiPageRead     `json:"candidate_pages,omitempty"`
	PreviousValidationErrors []string                        `json:"previous_validation_errors,omitempty"`
}

type WikiSynthesisMetadata struct {
	RequestMetadataHash  string
	ResponseMetadataHash string
}

type WikiSynthesisOutput struct {
	Pages []WikiSynthesisPage `json:"pages"`
}

type WikiSynthesisPage struct {
	Slug    string   `json:"slug"`
	Title   string   `json:"title"`
	Type    string   `json:"type"`
	Tags    []string `json:"tags"`
	Summary string   `json:"summary"`
	Owners  []string `json:"owners"`
	// Model outputs sometimes return structured freshness even though the
	// renderer derives canonical freshness from cited source timestamps.
	Freshness     any                     `json:"freshness"`
	Claims        []WikiSynthesisClaim    `json:"claims"`
	Conflicts     []WikiSynthesisConflict `json:"conflicts,omitempty"`
	OpenQuestions []string                `json:"open_questions,omitempty"`
	RelatedPages  []string                `json:"related_pages,omitempty"`
}

func (p *WikiSynthesisPage) UnmarshalJSON(raw []byte) error {
	var aux struct {
		Slug          string                  `json:"slug"`
		Title         string                  `json:"title"`
		Type          string                  `json:"type"`
		Tags          json.RawMessage         `json:"tags"`
		Summary       string                  `json:"summary"`
		Owners        json.RawMessage         `json:"owners"`
		Freshness     any                     `json:"freshness"`
		Claims        []WikiSynthesisClaim    `json:"claims"`
		Conflicts     []WikiSynthesisConflict `json:"conflicts"`
		OpenQuestions json.RawMessage         `json:"open_questions"`
		RelatedPages  json.RawMessage         `json:"related_pages"`
	}
	if err := json.Unmarshal(raw, &aux); err != nil {
		return err
	}
	tags, err := decodeWikiSynthesisStringList(aux.Tags)
	if err != nil {
		return fmt.Errorf("tags: %w", err)
	}
	owners, err := decodeWikiSynthesisStringList(aux.Owners)
	if err != nil {
		return fmt.Errorf("owners: %w", err)
	}
	openQuestions, err := decodeWikiSynthesisStringList(aux.OpenQuestions)
	if err != nil {
		return fmt.Errorf("open_questions: %w", err)
	}
	relatedPages, err := decodeWikiSynthesisStringList(aux.RelatedPages)
	if err != nil {
		return fmt.Errorf("related_pages: %w", err)
	}
	*p = WikiSynthesisPage{
		Slug:          aux.Slug,
		Title:         aux.Title,
		Type:          aux.Type,
		Tags:          tags,
		Summary:       aux.Summary,
		Owners:        owners,
		Freshness:     aux.Freshness,
		Claims:        aux.Claims,
		Conflicts:     aux.Conflicts,
		OpenQuestions: openQuestions,
		RelatedPages:  relatedPages,
	}
	return nil
}

type WikiSynthesisClaim struct {
	ClaimKey    string                           `json:"claim_key"`
	Text        string                           `json:"text"`
	Confidence  float64                          `json:"confidence"`
	Citations   []store.CompanyWikiCitationInput `json:"citations"`
	CitationIDs []string                         `json:"citation_ids,omitempty"`
}

func (c *WikiSynthesisClaim) UnmarshalJSON(raw []byte) error {
	var aux struct {
		ClaimKey    string          `json:"claim_key"`
		Text        string          `json:"text"`
		ClaimText   string          `json:"claim_text"`
		Claim       string          `json:"claim"`
		Statement   string          `json:"statement"`
		Body        string          `json:"body"`
		Summary     string          `json:"summary"`
		Confidence  float64         `json:"confidence"`
		Citations   json.RawMessage `json:"citations"`
		Citation    json.RawMessage `json:"citation"`
		CitationIDs []string        `json:"citation_ids"`
	}
	if err := json.Unmarshal(raw, &aux); err != nil {
		return err
	}
	citationsRaw := aux.Citations
	if !rawJSONValuePresent(citationsRaw) {
		citationsRaw = aux.Citation
	}
	citations, err := decodeWikiSynthesisCitations(citationsRaw)
	if err != nil {
		return err
	}
	*c = WikiSynthesisClaim{
		ClaimKey:    aux.ClaimKey,
		Text:        firstNonEmpty(aux.Text, aux.ClaimText, aux.Claim, aux.Statement, aux.Body, aux.Summary),
		Confidence:  aux.Confidence,
		Citations:   citations,
		CitationIDs: aux.CitationIDs,
	}
	return nil
}

type WikiSynthesisConflict struct {
	ClaimKey    string                           `json:"claim_key"`
	Summary     string                           `json:"summary"`
	Citations   []store.CompanyWikiCitationInput `json:"citations,omitempty"`
	CitationIDs []string                         `json:"citation_ids,omitempty"`
}

func (c *WikiSynthesisConflict) UnmarshalJSON(raw []byte) error {
	var aux struct {
		ClaimKey    string          `json:"claim_key"`
		Summary     string          `json:"summary"`
		Text        string          `json:"text"`
		Description string          `json:"description"`
		Citations   json.RawMessage `json:"citations"`
		Citation    json.RawMessage `json:"citation"`
		CitationIDs []string        `json:"citation_ids"`
	}
	if err := json.Unmarshal(raw, &aux); err != nil {
		return err
	}
	citationsRaw := aux.Citations
	if !rawJSONValuePresent(citationsRaw) {
		citationsRaw = aux.Citation
	}
	citations, err := decodeWikiSynthesisCitations(citationsRaw)
	if err != nil {
		return err
	}
	*c = WikiSynthesisConflict{
		ClaimKey:    aux.ClaimKey,
		Summary:     firstNonEmpty(aux.Summary, aux.Text, aux.Description),
		Citations:   citations,
		CitationIDs: aux.CitationIDs,
	}
	return nil
}

type OpenRouterWikiSynthesisClient struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

func NewOpenRouterWikiSynthesisClient(cfg config.Config) *OpenRouterWikiSynthesisClient {
	return &OpenRouterWikiSynthesisClient{
		BaseURL: strings.TrimRight(strings.TrimSpace(cfg.CompanyWikiCompilerOpenRouterBaseURL), "/"),
		APIKey:  strings.TrimSpace(cfg.CompanyWikiCompilerOpenRouterAPIKey),
		Client:  &http.Client{Timeout: cfg.CompanyWikiCompilerTimeout},
	}
}

func RunCompanyWikiCompiler(ctx context.Context, cfg config.Config, repo any, client WikiSynthesisClient) (result WikiCompilerResult, err error) {
	wikiStore, ok := repo.(store.CompanyWikiStore)
	if !ok {
		return WikiCompilerResult{}, errors.New("configured store does not support company wiki")
	}
	if client == nil {
		client = NewOpenRouterWikiSynthesisClient(cfg)
	}
	if cfg.CompanyWikiCompilerRunTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.CompanyWikiCompilerRunTimeout)
		defer cancel()
	}
	runID := "compiler_" + time.Now().UTC().Format("20060102T150405.000000000Z")
	result = WikiCompilerResult{OK: true, CompilerRunID: runID}
	leaseDuration := companyWikiCompilerLeaseDuration(cfg)
	claimLimit := companyWikiCompilerClaimLimit(cfg)
	log.Printf("company-wiki-compiler run_id=%s starting batch_limit=%d effective_claim_limit=%d item_timeout=%s run_timeout=%s shutdown_grace=%s lease=%s",
		runID,
		cfg.CompanyWikiCompilerBatchLimit,
		claimLimit,
		cfg.CompanyWikiCompilerTimeout,
		cfg.CompanyWikiCompilerRunTimeout,
		cfg.CompanyWikiCompilerShutdownGrace,
		leaseDuration,
	)
	release, acquired, err := acquireCompanyWikiCompilerLease(ctx, repo, runID, leaseDuration)
	if err != nil {
		return result, err
	}
	if !acquired {
		log.Printf("company-wiki-compiler run_id=%s skipped reason=lease_held", runID)
		return result, nil
	}
	defer func() { _ = release() }()
	leaseHolder := "company_wiki_compiler:" + runID
	if recorder, ok := repo.(companyWikiCompilerRunRecorder); ok {
		_ = recorder.BeginCompanyWikiCompilerRun(runID, map[string]any{
			"compiler_version":     CompanyWikiCompilerVersion,
			"schema_version":       CompanyWikiSchemaVersion,
			"renderer_version":     CompanyWikiRendererVersion,
			"model_policy_version": CompanyWikiModelPolicyVersion,
			"model":                cfg.CompanyWikiCompilerModel,
		})
		defer func() {
			status := store.CompanyWikiCompileStatusCompleted
			lastError := ""
			if err != nil || !result.OK {
				status = store.CompanyWikiCompileStatusFailed
				if err != nil {
					lastError = err.Error()
				} else if len(result.FailedItems) > 0 {
					lastError = strings.Join(result.FailedItems, ",")
				}
			}
			_ = recorder.CompleteCompanyWikiCompilerRun(runID, status, lastError, map[string]any{
				"backfilled":      result.Backfilled,
				"claimed":         result.Claimed,
				"published_pages": result.PublishedPages,
				"failed_items":    result.FailedItems,
				"deferred_items":  result.DeferredItems,
				"stopped_reason":  result.StoppedReason,
			})
		}()
	}
	if err := WriteSchemaFile(cfg.CompanyWikiRoot); err != nil {
		return result, err
	}
	backfilled, err := BackfillCompanyWikiCompileItems(ctx, cfg, wikiStore, claimLimit*5)
	if err != nil {
		return result, err
	}
	result.Backfilled = backfilled
	log.Printf("company-wiki-compiler run_id=%s backfilled=%d", runID, backfilled)
	items, err := wikiStore.ClaimCompanyWikiCompileItems(store.CompanyWikiCompileClaimInput{
		Limit:              claimLimit,
		LeaseHolder:        leaseHolder,
		LeaseDuration:      leaseDuration,
		CompilerVersion:    CompanyWikiCompilerVersion,
		SchemaVersion:      CompanyWikiSchemaVersion,
		RendererVersion:    CompanyWikiRendererVersion,
		ModelPolicyVersion: CompanyWikiModelPolicyVersion,
		MaxAttempts:        CompanyWikiCompileMaxAttemptCount,
	})
	if err != nil {
		return result, err
	}
	result.Claimed = len(items)
	log.Printf("company-wiki-compiler run_id=%s claimed=%d", runID, len(items))
	for idx, item := range items {
		if stopReason := companyWikiCompilerStopReason(ctx, cfg, idx); stopReason != "" {
			result.DeferredItems = len(items) - idx
			result.StoppedReason = stopReason
			if released, releaseErr := releaseCompanyWikiDeferredCompileItems(wikiStore, items[idx:], leaseHolder, stopReason); releaseErr != nil {
				log.Printf("company-wiki-compiler run_id=%s deferred_release_failed error=%q", runID, releaseErr.Error())
			} else if released > 0 {
				log.Printf("company-wiki-compiler run_id=%s deferred_released=%d", runID, released)
			}
			log.Printf("company-wiki-compiler run_id=%s stopping reason=%s deferred_items=%d", runID, stopReason, result.DeferredItems)
			break
		}
		log.Printf("company-wiki-compiler run_id=%s item=%s source_revision=%s status=started index=%d/%d",
			runID, item.ID, item.SourceRevisionID, idx+1, len(items))
		published, err := compileOneWikiItem(ctx, cfg, wikiStore, client, runID, item)
		if err != nil {
			result.OK = false
			result.FailedItems = append(result.FailedItems, item.ID)
			_, _ = wikiStore.CompleteCompanyWikiCompileItem(item.ID, store.CompanyWikiCompileStatusFailed, err.Error())
			log.Printf("company-wiki-compiler run_id=%s item=%s source_revision=%s status=failed error=%q", runID, item.ID, item.SourceRevisionID, err.Error())
			continue
		}
		result.PublishedPages += published
		log.Printf("company-wiki-compiler run_id=%s item=%s source_revision=%s status=completed published_pages=%d", runID, item.ID, item.SourceRevisionID, published)
	}
	if err := WriteIndexFile(cfg.CompanyWikiRoot, wikiStore); err != nil {
		return result, err
	}
	log.Printf("company-wiki-compiler run_id=%s completed ok=%t claimed=%d published_pages=%d failed_items=%d deferred_items=%d stopped_reason=%q",
		runID, result.OK, result.Claimed, result.PublishedPages, len(result.FailedItems), result.DeferredItems, result.StoppedReason)
	return result, nil
}

func companyWikiCompilerLeaseDuration(cfg config.Config) time.Duration {
	if cfg.CompanyWikiCompilerRunTimeout > 0 {
		return cfg.CompanyWikiCompilerRunTimeout + time.Minute
	}
	limit := cfg.CompanyWikiCompilerBatchLimit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return time.Duration(limit)*companyWikiCompilerPerItemBudget(cfg) + time.Minute
}

func companyWikiCompilerClaimLimit(cfg config.Config) int {
	limit := cfg.CompanyWikiCompilerBatchLimit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	runTimeout := cfg.CompanyWikiCompilerRunTimeout
	if runTimeout <= 0 {
		return limit
	}
	usable := runTimeout - companyWikiCompilerShutdownGrace(cfg)
	if usable <= 0 {
		return 1
	}
	perItem := companyWikiCompilerPerItemBudget(cfg)
	if perItem <= 0 {
		return limit
	}
	byBudget := int(usable / perItem)
	if byBudget < 1 {
		byBudget = 1
	}
	if byBudget < limit {
		return byBudget
	}
	return limit
}

func companyWikiCompilerPerItemBudget(cfg config.Config) time.Duration {
	timeout := cfg.CompanyWikiCompilerTimeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	return time.Duration(companyWikiCompilerValidationAttempts*companyWikiCompilerProviderDecodeAttempts)*timeout + companyWikiCompilerPerItemOverhead
}

func companyWikiCompilerShutdownGrace(cfg config.Config) time.Duration {
	if cfg.CompanyWikiCompilerShutdownGrace < 0 {
		return 0
	}
	if cfg.CompanyWikiCompilerShutdownGrace == 0 {
		return 30 * time.Second
	}
	return cfg.CompanyWikiCompilerShutdownGrace
}

func companyWikiCompilerStopReason(ctx context.Context, cfg config.Config, completed int) string {
	if err := ctx.Err(); err != nil {
		return err.Error()
	}
	deadline, ok := ctx.Deadline()
	if !ok || completed == 0 {
		return ""
	}
	remaining := time.Until(deadline) - companyWikiCompilerShutdownGrace(cfg)
	if remaining < companyWikiCompilerPerItemBudget(cfg) {
		return "run_budget_exhausted"
	}
	return ""
}

type companyWikiCompileItemReleaser interface {
	ReleaseCompanyWikiCompileItems(ids []string, leaseHolder string, reason string) (int, error)
}

func releaseCompanyWikiDeferredCompileItems(repo any, items []store.CompanyWikiCompileItem, leaseHolder string, reason string) (int, error) {
	releaser, ok := repo.(companyWikiCompileItemReleaser)
	if !ok || len(items) == 0 {
		return 0, nil
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.ID) != "" {
			ids = append(ids, item.ID)
		}
	}
	if len(ids) == 0 {
		return 0, nil
	}
	return releaser.ReleaseCompanyWikiCompileItems(ids, leaseHolder, reason)
}

type companyWikiCompilerLeaser interface {
	AcquireCompanyWikiCompilerLease(ctx context.Context, lockName string, holder string, ttl time.Duration) (func() error, bool, error)
}

type companyWikiCompilerRunRecorder interface {
	BeginCompanyWikiCompilerRun(id string, metadata map[string]any) error
	CompleteCompanyWikiCompilerRun(id string, status string, lastError string, metadata map[string]any) error
}

func acquireCompanyWikiCompilerLease(ctx context.Context, repo any, holder string, ttl time.Duration) (func() error, bool, error) {
	leaser, ok := repo.(companyWikiCompilerLeaser)
	if !ok {
		return func() error { return nil }, true, nil
	}
	return leaser.AcquireCompanyWikiCompilerLease(ctx, "company_wiki_compiler:"+CompanyWikiCompilerVersion, holder, ttl)
}

func compileOneWikiItem(ctx context.Context, cfg config.Config, wikiStore store.CompanyWikiStore, client WikiSynthesisClient, runID string, item store.CompanyWikiCompileItem) (int, error) {
	evidence, found, err := wikiStore.GetCompanyWikiSourceEvidence(item.SourceRevisionID)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, fmt.Errorf("source revision %s was not found", item.SourceRevisionID)
	}
	if err := validateCompilerSourcePolicy(evidence.Document, evidence.Revision); err != nil {
		_, _ = wikiStore.CompleteCompanyWikiCompileItem(item.ID, store.CompanyWikiCompileStatusSkipped, err.Error())
		return 0, nil
	}
	chunks := evidence.Chunks
	if cfg.CompanyWikiCompilerChunkLimit > 0 && len(chunks) > cfg.CompanyWikiCompilerChunkLimit {
		chunks = chunks[:cfg.CompanyWikiCompilerChunkLimit]
	}
	candidates, err := wikiStore.ListCompanyWikiCandidatePages(store.CompanyWikiPageQuery{
		Query:           evidence.Document.Title,
		Limit:           8,
		ExcludeEvidence: true,
	})
	if err != nil {
		return 0, err
	}
	contextHash := compilerContextHash(evidence, chunks, candidates)
	attempt, err := wikiStore.BeginCompanyWikiCompileAttempt(store.CompanyWikiCompileAttemptInput{
		CompileItemID: item.ID,
		CompilerRunID: runID,
		Status:        store.CompanyWikiCompileStatusClaimed,
		Model:         cfg.CompanyWikiCompilerModel,
		ContextHash:   contextHash,
	})
	if err != nil {
		return 0, err
	}
	start := time.Now()
	request := WikiSynthesisRequest{
		Model:          cfg.CompanyWikiCompilerModel,
		Source:         evidence,
		Chunks:         chunks,
		CandidatePages: candidates,
	}
	var output WikiSynthesisOutput
	var metadata WikiSynthesisMetadata
	var pages []WikiSynthesisPage
	var validationErrors []string
	var outputHash string
	var duration int64
	for synthAttempt := 1; synthAttempt <= 2; synthAttempt++ {
		output, metadata, err = client.SynthesizeWiki(ctx, request)
		duration = time.Since(start).Milliseconds()
		if err != nil {
			_, _ = wikiStore.CompleteCompanyWikiCompileAttempt(attempt.ID, store.CompanyWikiCompileStatusFailed, "", duration, nil, err.Error(), nil)
			return 0, err
		}
		output = preserveCandidateClaimsInSynthesisOutput(output, evidence, candidates)
		outputHash = store.CompanyWikiSHA256(mustMarshalString(output))
		pages, validationErrors = validateSynthesisOutput(evidence, candidates, output)
		if len(validationErrors) == 0 {
			break
		}
		request.PreviousValidationErrors = append([]string(nil), validationErrors...)
	}
	if len(validationErrors) > 0 {
		err := errors.New(strings.Join(validationErrors, "; "))
		_, _ = wikiStore.CompleteCompanyWikiCompileAttempt(attempt.ID, store.CompanyWikiCompileStatusFailed, outputHash, duration, validationErrors, err.Error(), map[string]any{
			"request_metadata_hash":  metadata.RequestMetadataHash,
			"response_metadata_hash": metadata.ResponseMetadataHash,
		})
		return 0, err
	}
	targetInputs := make([]store.CompanyWikiCompileTargetInput, 0, len(pages))
	rendered := map[string]renderedSynthesisPage{}
	revisionTimestamps := buildRevisionTimestampMap(evidence, candidates)
	for _, page := range pages {
		body, citations, claims, conflicts := renderSynthesisPageMarkdownWithCandidates(evidence, candidates, page)
		slug := synthesisSlug(page)
		freshness := synthesisFreshness(revisionTimestamps, page)
		bodyHash := store.CompanyWikiSHA256(body)
		targetPath := filepath.ToSlash(filepath.Join("pages", slug+".md"))
		idempotencyKey := strings.Join([]string{
			CompanyWikiCompilerVersion,
			CompanyWikiSchemaVersion,
			CompanyWikiRendererVersion,
			CompanyWikiModelPolicyVersion,
			item.SourceRevisionID,
			slug,
			bodyHash,
		}, ":")
		targetInputs = append(targetInputs, store.CompanyWikiCompileTargetInput{
			CompileItemID:  item.ID,
			TargetSlug:     slug,
			TargetPath:     targetPath,
			TargetType:     page.Type,
			Status:         store.CompanyWikiCompileTargetStatusPending,
			IdempotencyKey: idempotencyKey,
			BodyHash:       bodyHash,
		})
		rendered[slug] = renderedSynthesisPage{
			page: page, body: body, citations: citations, claims: claims, conflicts: conflicts,
			path: targetPath, bodyHash: bodyHash, idempotencyKey: idempotencyKey, freshness: freshness,
		}
	}
	targets, err := wikiStore.UpsertCompanyWikiCompileTargets(item.ID, targetInputs)
	if err != nil {
		_, _ = wikiStore.CompleteCompanyWikiCompileAttempt(attempt.ID, store.CompanyWikiCompileStatusFailed, "", duration, nil, err.Error(), nil)
		return 0, err
	}
	published := 0
	var firstPublishError error
	for _, target := range targets {
		renderedPage, ok := rendered[target.TargetSlug]
		if !ok {
			continue
		}
		if target.Status == store.CompanyWikiCompileTargetStatusPublished && target.BodyHash == renderedPage.bodyHash {
			continue
		}
		pageResult, err := publishSynthesisTarget(cfg, wikiStore, runID, renderedPage)
		if err != nil {
			_, _ = wikiStore.UpdateCompanyWikiCompileTarget(store.CompanyWikiCompileTargetInput{
				CompileItemID: item.ID,
				TargetSlug:    target.TargetSlug,
				TargetPath:    target.TargetPath,
				TargetType:    target.TargetType,
				Status:        store.CompanyWikiCompileTargetStatusFailed,
				BodyHash:      renderedPage.bodyHash,
				LastError:     err.Error(),
			})
			if firstPublishError == nil {
				firstPublishError = err
			}
			continue
		}
		if _, err := wikiStore.UpdateCompanyWikiCompileTarget(store.CompanyWikiCompileTargetInput{
			CompileItemID:  item.ID,
			TargetSlug:     target.TargetSlug,
			TargetPath:     target.TargetPath,
			TargetType:     target.TargetType,
			Status:         store.CompanyWikiCompileTargetStatusPublished,
			WikiRevisionID: pageResult.Revision.ID,
			IdempotencyKey: renderedPage.idempotencyKey,
			BodyHash:       renderedPage.bodyHash,
		}); err != nil {
			if firstPublishError == nil {
				firstPublishError = err
			}
			continue
		}
		published++
	}
	if firstPublishError != nil {
		_, _ = wikiStore.CompleteCompanyWikiCompileAttempt(attempt.ID, store.CompanyWikiCompileStatusFailed, "", duration, nil, firstPublishError.Error(), nil)
		return published, firstPublishError
	}
	if _, err := wikiStore.CompleteCompanyWikiCompileAttempt(attempt.ID, store.CompanyWikiCompileStatusCompleted, outputHash, duration, nil, "", map[string]any{
		"request_metadata_hash":  metadata.RequestMetadataHash,
		"response_metadata_hash": metadata.ResponseMetadataHash,
	}); err != nil {
		return published, err
	}
	finalTargets, err := wikiStore.ListCompanyWikiCompileTargets(item.ID)
	if err != nil {
		return published, err
	}
	for _, target := range finalTargets {
		if !compileTargetTerminal(target.Status) {
			return published, fmt.Errorf("compile target %s is not terminal: %s", target.TargetSlug, target.Status)
		}
	}
	if _, err := wikiStore.CompleteCompanyWikiCompileItem(item.ID, store.CompanyWikiCompileStatusCompleted, ""); err != nil {
		return published, err
	}
	return published, nil
}

func compileTargetTerminal(status string) bool {
	switch strings.TrimSpace(status) {
	case store.CompanyWikiCompileTargetStatusPublished, store.CompanyWikiCompileTargetStatusSkipped, store.CompanyWikiCompileTargetStatusSuperseded, store.CompanyWikiCompileTargetStatusFailed:
		return true
	default:
		return false
	}
}

type renderedSynthesisPage struct {
	page           WikiSynthesisPage
	body           string
	citations      []store.CompanyWikiCitationInput
	claims         []store.CompanyWikiClaimInput
	conflicts      []store.CompanyWikiConflictInput
	freshness      string
	path           string
	bodyHash       string
	idempotencyKey string
}

func publishSynthesisTarget(cfg config.Config, wikiStore store.CompanyWikiStore, runID string, rendered renderedSynthesisPage) (store.CompanyWikiPagePublishResult, error) {
	audit, err := wikiStore.BeginCompanyWikiAudit(store.CompanyWikiAuditInput{
		Mode:           store.CompanyWikiAuditModeCompiler,
		Actor:          "company_wiki_compiler",
		Reason:         "synthesis_page_compiled",
		IdempotencyKey: rendered.idempotencyKey,
		Slug:           synthesisSlug(rendered.page),
		Title:          rendered.page.Title,
		ProposedPath:   rendered.path,
		Metadata: map[string]any{
			"page_type":       rendered.page.Type,
			"compiler_run_id": runID,
		},
	})
	if err != nil {
		return store.CompanyWikiPagePublishResult{}, err
	}
	stage, sha, err := StageMarkdownFile(cfg.CompanyWikiRoot, rendered.path, rendered.body)
	if err != nil {
		_, _ = wikiStore.FailCompanyWikiAudit(audit.ID, err.Error(), map[string]any{"stage": "stage_markdown"})
		return store.CompanyWikiPagePublishResult{}, err
	}
	removeStage := true
	defer func() {
		if removeStage {
			_ = os.Remove(stage)
		}
	}()
	if sha != rendered.bodyHash {
		err := fmt.Errorf("staged body hash mismatch for %s", rendered.path)
		_, _ = wikiStore.FailCompanyWikiAudit(audit.ID, err.Error(), map[string]any{"stage": "stage_markdown"})
		return store.CompanyWikiPagePublishResult{}, err
	}
	page, err := wikiStore.PublishCompanyWikiPage(store.CompanyWikiPagePublishInput{
		AuditID:           audit.ID,
		Slug:              synthesisSlug(rendered.page),
		Title:             rendered.page.Title,
		Body:              rendered.body,
		Path:              rendered.path,
		SHA256:            sha,
		CompilerRunID:     runID,
		SourceRevisionIDs: uniqueCitationRevisionIDs(rendered.citations),
		Citations:         rendered.citations,
		Claims:            rendered.claims,
		Conflicts:         rendered.conflicts,
		Metadata: map[string]any{
			"type":      rendered.page.Type,
			"tags":      rendered.page.Tags,
			"owners":    rendered.page.Owners,
			"freshness": rendered.freshness,
		},
		PublishedAt: time.Now().UTC(),
	})
	if err != nil {
		_, _ = wikiStore.FailCompanyWikiAudit(audit.ID, err.Error(), map[string]any{"stage": "record_revision"})
		return store.CompanyWikiPagePublishResult{}, err
	}
	if err := CommitStagedMarkdownFile(stage, cfg.CompanyWikiRoot, rendered.path); err != nil {
		_ = wikiStore.UpdateCompanyWikiManifestRepair(rendered.path, store.CompanyWikiManifestRepairNeeded, err.Error())
		_, _ = wikiStore.FailCompanyWikiAudit(audit.ID, err.Error(), map[string]any{"stage": "commit_markdown", "repair_status": store.CompanyWikiManifestRepairNeeded})
		_ = AppendLogEntry(cfg.CompanyWikiRoot, WikiLogEntry{Action: "repair_needed", Title: rendered.page.Title, Slug: synthesisSlug(rendered.page), WikiRevisionID: page.Revision.ID, Summary: err.Error()})
		removeStage = false
		return store.CompanyWikiPagePublishResult{}, err
	}
	removeStage = false
	if err := WriteManifestFile(cfg.CompanyWikiRoot, page.Page.Slug, page.Revision.ID, rendered.path, sha, page.Revision.CompilerRunID, page.Revision.PublishedAt); err != nil {
		return store.CompanyWikiPagePublishResult{}, err
	}
	if _, err := wikiStore.CompleteCompanyWikiAudit(audit.ID, page.Revision.ID, rendered.path, map[string]any{"sha256": sha, "wiki_revision_id": page.Revision.ID}); err != nil {
		return store.CompanyWikiPagePublishResult{}, err
	}
	if err := AppendLogEntry(cfg.CompanyWikiRoot, WikiLogEntry{Action: "synthesis", Title: rendered.page.Title, Slug: page.Page.Slug, Status: "published", WikiRevisionID: page.Revision.ID, Summary: wikiOneLineSummary(rendered.body)}); err != nil {
		return store.CompanyWikiPagePublishResult{}, err
	}
	return page, nil
}

func (c *OpenRouterWikiSynthesisClient) SynthesizeWiki(ctx context.Context, request WikiSynthesisRequest) (WikiSynthesisOutput, WikiSynthesisMetadata, error) {
	if strings.TrimSpace(c.BaseURL) == "" || strings.TrimSpace(c.APIKey) == "" {
		return WikiSynthesisOutput{}, WikiSynthesisMetadata{}, errors.New("OpenRouter company wiki compiler is not configured")
	}
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	var lastErr error
	for attempt := 1; attempt <= 2; attempt++ {
		output, metadata, err := c.synthesizeWikiOnce(ctx, client, request, attempt, lastErr)
		if err == nil {
			return output, metadata, nil
		}
		if !isStructuredOutputDecodeError(err) {
			return WikiSynthesisOutput{}, WikiSynthesisMetadata{}, err
		}
		lastErr = err
	}
	return WikiSynthesisOutput{}, WikiSynthesisMetadata{}, lastErr
}

func (c *OpenRouterWikiSynthesisClient) synthesizeWikiOnce(ctx context.Context, client *http.Client, request WikiSynthesisRequest, attempt int, previousErr error) (WikiSynthesisOutput, WikiSynthesisMetadata, error) {
	messages := []map[string]string{
		{"role": "system", "content": synthesisSystemPrompt()},
		{"role": "user", "content": synthesisUserPrompt(request)},
	}
	if previousErr != nil {
		messages = append(messages, map[string]string{
			"role":    "user",
			"content": "The previous response was rejected before persistence: " + previousErr.Error() + ". Return one complete valid JSON object with a pages array, cited claim objects, and no markdown fences.",
		})
	}
	payload := map[string]any{
		"model":           request.Model,
		"messages":        messages,
		"temperature":     0,
		"response_format": map[string]string{"type": "json_object"},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return WikiSynthesisOutput{}, WikiSynthesisMetadata{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return WikiSynthesisOutput{}, WikiSynthesisMetadata{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return WikiSynthesisOutput{}, WikiSynthesisMetadata{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return WikiSynthesisOutput{}, WikiSynthesisMetadata{}, fmt.Errorf("OpenRouter compiler returned HTTP %d: %s", resp.StatusCode, truncateForCitation(string(body), 800))
	}
	var decoded struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage map[string]any `json:"usage"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return WikiSynthesisOutput{}, WikiSynthesisMetadata{}, err
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return WikiSynthesisOutput{}, WikiSynthesisMetadata{}, errors.New("OpenRouter compiler returned no content")
	}
	output, err := decodeWikiSynthesisContent(decoded.Choices[0].Message.Content)
	if err != nil {
		return WikiSynthesisOutput{}, WikiSynthesisMetadata{}, err
	}
	meta := map[string]any{"model": decoded.Model, "id": decoded.ID, "usage": decoded.Usage, "attempt": attempt}
	return output, WikiSynthesisMetadata{
		RequestMetadataHash:  store.CompanyWikiSHA256(request.Model + ":" + request.Source.Revision.ID),
		ResponseMetadataHash: store.CompanyWikiSHA256(mustMarshalString(meta)),
	}, nil
}
