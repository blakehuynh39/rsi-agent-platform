package companyknowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestRecordAndPublishWikiSourceWritesMarkdownManifestAndLedger(t *testing.T) {
	state := store.NewMemoryStore()
	root := t.TempDir()
	result, err := RecordAndPublishWikiSource(context.Background(), config.Config{CompanyWikiRoot: root}, state, store.CompanyWikiSourceRevisionInput{
		SourceType:        "notion_document",
		DocumentSourceKey: "notion:page:abc",
		SourceSessionKey:  "notion:page:abc",
		SourceRevision:    "rev-1",
		Workspace:         "rsi_company_knowledge",
		Environment:       "test",
		Title:             "Deploy Runbook",
		URL:               "https://notion.so/page-abc",
		Content:           "Roll forward after validation.\nCheck Argo before paging.",
		NativeLocator:     "notion:block:path",
	})
	if err != nil {
		t.Fatalf("RecordAndPublishWikiSource() error = %v", err)
	}
	if result.Skipped || result.Page.Page.Slug == "" || result.Audit.Status != store.CompanyWikiAuditStatusPublished {
		t.Fatalf("unexpected publish result: %+v", result)
	}
	body, err := os.ReadFile(filepath.Join(root, result.Page.Revision.Path))
	if err != nil {
		t.Fatalf("published markdown missing: %v", err)
	}
	for _, expected := range []string{"source_document_id", "source_revision_id", "chunk_id", "Roll forward after validation."} {
		if !strings.Contains(string(body), expected) {
			t.Fatalf("expected %q in markdown:\n%s", expected, string(body))
		}
	}
	manifest, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatalf("manifest missing: %v", err)
	}
	if !strings.Contains(string(manifest), result.Page.Revision.ID) || !strings.Contains(string(manifest), result.Page.Revision.BodySHA256) {
		t.Fatalf("manifest does not reference revision/sha:\n%s", string(manifest))
	}
	index, err := os.ReadFile(filepath.Join(root, "index.md"))
	if err != nil {
		t.Fatalf("index.md missing: %v", err)
	}
	if !strings.Contains(string(index), "[Deploy Runbook]("+result.Page.Revision.Path+")") || !strings.Contains(string(index), "source revisions") {
		t.Fatalf("index.md does not catalog published page:\n%s", string(index))
	}
	logBody, err := os.ReadFile(filepath.Join(root, "log.md"))
	if err != nil {
		t.Fatalf("log.md missing: %v", err)
	}
	if !strings.Contains(string(logBody), "## [") || !strings.Contains(string(logBody), "] ingest | Deploy Runbook") {
		t.Fatalf("log.md entry missing consistent prefix:\n%s", string(logBody))
	}
}

func TestReconcileWikiManifestDetectsAndRepairsDirectMutation(t *testing.T) {
	state := store.NewMemoryStore()
	root := t.TempDir()
	result, err := RecordAndPublishWikiSource(context.Background(), config.Config{CompanyWikiRoot: root}, state, store.CompanyWikiSourceRevisionInput{
		SourceType:        "slack_message",
		DocumentSourceKey: "slack:T:C:1",
		SourceSessionKey:  "slack:T:C:thread",
		SourceRevision:    "1",
		Title:             "Deploy Thread",
		Content:           "Use the deploy checklist.",
		NativeLocator:     "slack:C:1:1",
	})
	if err != nil {
		t.Fatalf("RecordAndPublishWikiSource() error = %v", err)
	}
	publishedPath := filepath.Join(root, result.Page.Revision.Path)
	if err := os.WriteFile(publishedPath, []byte("tampered"), 0o644); err != nil {
		t.Fatalf("tamper published file: %v", err)
	}
	check, err := ReconcileWikiManifest(context.Background(), config.Config{CompanyWikiRoot: root}, state, false)
	if err != nil {
		t.Fatalf("ReconcileWikiManifest(check) error = %v", err)
	}
	if check.OK || len(check.Warnings) != 1 || check.Warnings[0].Reason != "published_file_sha256_mismatch" {
		t.Fatalf("expected checksum warning, got %+v", check)
	}
	repaired, err := ReconcileWikiManifest(context.Background(), config.Config{CompanyWikiRoot: root}, state, true)
	if err != nil {
		t.Fatalf("ReconcileWikiManifest(repair) error = %v", err)
	}
	if !repaired.OK || len(repaired.Repaired) != 1 || len(repaired.Warnings) != 0 {
		t.Fatalf("expected repaired clean result, got %+v", repaired)
	}
	body, err := os.ReadFile(publishedPath)
	if err != nil {
		t.Fatalf("read repaired file: %v", err)
	}
	if !strings.Contains(string(body), "Use the deploy checklist.") {
		t.Fatalf("repaired file did not restore DB revision body:\n%s", string(body))
	}
}

func TestRecordWikiSourceRevisionDoesNotRequireWikiRoot(t *testing.T) {
	state := store.NewMemoryStore()
	recorded, err := RecordWikiSourceRevision(context.Background(), config.Config{}, state, store.CompanyWikiSourceRevisionInput{
		SourceType:        "notion_document",
		DocumentSourceKey: "notion:page:abc",
		SourceSessionKey:  "notion:page:abc",
		SourceRevision:    "rev-1",
		Title:             "Deploy Runbook",
		Content:           "Roll forward after validation.",
		NativeLocator:     "notion:block:path",
	})
	if err != nil {
		t.Fatalf("RecordWikiSourceRevision() error = %v", err)
	}
	if recorded.Skipped || recorded.Source.Document.ID == "" || recorded.Source.Revision.ID == "" {
		t.Fatalf("expected source ledger record without wiki root, got %+v", recorded)
	}
	published, err := PublishWikiSourceDocument(context.Background(), config.Config{}, state, recorded.Source)
	if err != nil {
		t.Fatalf("PublishWikiSourceDocument() error = %v", err)
	}
	if !published.Skipped || published.Reason != "company_wiki_root_not_configured" {
		t.Fatalf("publish result = %+v, want root-not-configured skip", published)
	}
}

func TestCleanRelativeWikiPathStripsRepeatedTraversal(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  string
	}{
		{input: "../../../pages/deploy.md", want: "pages/deploy.md"},
		{input: "/../../pages/deploy.md", want: "pages/deploy.md"},
		{input: "pages/../../../deploy.md", want: "deploy.md"},
		{input: "..", want: ""},
		{input: ".", want: ""},
	} {
		if got := cleanRelativeWikiPath(tc.input); got != tc.want {
			t.Fatalf("cleanRelativeWikiPath(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestWriteManifestFileConcurrentWritesKeepAllEntries(t *testing.T) {
	root := t.TempDir()
	const writers = 20
	var wg sync.WaitGroup
	errs := make(chan error, writers)
	for i := 0; i < writers; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- WriteManifestFile(
				root,
				fmt.Sprintf("page-%02d", i),
				fmt.Sprintf("rev-%02d", i),
				fmt.Sprintf("pages/page-%02d.md", i),
				fmt.Sprintf("sha-%02d", i),
				"compiler-test",
				time.Unix(int64(i), 0).UTC(),
			)
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("WriteManifestFile() concurrent error = %v", err)
		}
	}
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var decoded struct {
		Pages map[string]any `json:"pages"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode manifest: %v\n%s", err, raw)
	}
	if len(decoded.Pages) != writers {
		t.Fatalf("manifest pages = %d, want %d\n%s", len(decoded.Pages), writers, raw)
	}
}

func TestWikiSlugForSourceHandlesShortIDs(t *testing.T) {
	slug := WikiSlugForSource(store.CompanyWikiSourceDocument{
		ID:         "abc",
		SourceType: "notion_document",
		Title:      "Deploy Runbook",
	})
	if slug == "" || !strings.Contains(slug, "abc") {
		t.Fatalf("slug = %q, want short id suffix preserved", slug)
	}
}
