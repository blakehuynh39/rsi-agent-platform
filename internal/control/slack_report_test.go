package control

import (
	"encoding/base64"
	"strings"
	"testing"

	slackapi "github.com/slack-go/slack"

	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestSlackMessageBlockAndAttachmentArgsDecode(t *testing.T) {
	args := map[string]any{
		"blocks": []any{map[string]any{
			"type": "section",
			"text": map[string]any{"type": "mrkdwn", "text": "*hello*"},
		}},
		"attachments": []any{map[string]any{"fallback": "fallback", "text": "attachment"}},
	}
	blocks, err := slackBlocksArg(args, "blocks")
	if err != nil {
		t.Fatalf("decode blocks: %v", err)
	}
	if len(blocks.BlockSet) != 1 {
		t.Fatalf("expected one block, got %d", len(blocks.BlockSet))
	}
	if _, ok := blocks.BlockSet[0].(*slackapi.SectionBlock); !ok {
		t.Fatalf("expected section block, got %T", blocks.BlockSet[0])
	}
	attachments, err := slackAttachmentsArg(args, "attachments")
	if err != nil {
		t.Fatalf("decode attachments: %v", err)
	}
	if len(attachments) != 1 || attachments[0].Text != "attachment" {
		t.Fatalf("unexpected attachments: %#v", attachments)
	}
}

func TestSlackUploadParamsSupportsContentBase64AndFileSize(t *testing.T) {
	params, err := slackUploadParams(map[string]any{
		"channel_id":     "C123",
		"content_base64": base64.StdEncoding.EncodeToString([]byte("hello")),
		"filename":       "hello.txt",
	}, "")
	if err != nil {
		t.Fatalf("slackUploadParams: %v", err)
	}
	if params.Content != "hello" {
		t.Fatalf("decoded content = %q", params.Content)
	}
	if params.FileSize != 5 {
		t.Fatalf("file size = %d, want 5", params.FileSize)
	}
}

func TestSlackUploadParamsNormalizesFileURLArtifactRefs(t *testing.T) {
	if got := slackUploadLocalPath("file:///tmp/report.csv"); got != "/tmp/report.csv" {
		t.Fatalf("normalized path = %q", got)
	}
}

func TestBuildSlackReportPlanRendersOneSmallTableInline(t *testing.T) {
	plan, err := buildSlackReportPlan(map[string]any{
		"channel_id":            "C123",
		"report_schema_version": 1,
		"summary":               "Campaign summary",
		"tables": []any{map[string]any{
			"title": "Campaigns",
			"columns": []any{
				map[string]any{"key": "campaign", "label": "Campaign"},
				map[string]any{"key": "submissions", "label": "Submissions", "align": "right"},
			},
			"rows": []any{
				map[string]any{"campaign": "Vietnamese", "submissions": 12815},
			},
		}},
	}, "")
	if err != nil {
		t.Fatalf("buildSlackReportPlan: %v", err)
	}
	if len(plan.Uploads) != 0 {
		t.Fatalf("small inline report should not upload files: %#v", plan.Uploads)
	}
	foundTable := false
	for _, block := range plan.Blocks {
		if _, ok := block.(*slackapi.TableBlock); ok {
			foundTable = true
		}
	}
	if !foundTable {
		t.Fatalf("expected Slack table block in %#v", plan.Blocks)
	}
}

func TestBuildSlackReportPlanFallsBackToCSVForMultipleTables(t *testing.T) {
	table := map[string]any{
		"columns": []any{
			map[string]any{"key": "country", "label": "Country"},
			map[string]any{"key": "pct", "label": "%"},
		},
		"rows": []any{map[string]any{"country": "VN", "pct": 89.2}},
	}
	plan, err := buildSlackReportPlan(map[string]any{
		"channel_id":            "C123",
		"report_schema_version": 1,
		"summary":               "Country breakdown",
		"tables":                []any{table, table},
	}, "")
	if err != nil {
		t.Fatalf("buildSlackReportPlan: %v", err)
	}
	if len(plan.Uploads) != 2 {
		t.Fatalf("expected two CSV uploads, got %d", len(plan.Uploads))
	}
	planned, _ := plan.Manifest["uploads"].([]map[string]interface{})
	if len(planned) != 2 {
		t.Fatalf("expected two planned uploads in manifest, got %#v", plan.Manifest["uploads"])
	}
	if !strings.Contains(plan.Uploads[0].Params.Content, "Country,%") {
		t.Fatalf("expected CSV content, got %q", plan.Uploads[0].Params.Content)
	}
}

func TestBuildSlackReportDraftManifestDoesNotStatArtifactPaths(t *testing.T) {
	manifest, err := buildSlackReportDraftManifest(map[string]any{
		"channel_id":            "C123",
		"report_schema_version": 1,
		"summary":               "Report with generated artifact",
		"files": []any{
			map[string]any{"artifact_ref": "/definitely/not/present/report.csv", "filename": "report.csv"},
		},
	}, "")
	if err != nil {
		t.Fatalf("buildSlackReportDraftManifest should not stat artifact paths: %v", err)
	}
	if manifest["planned_uploads"] != 1 {
		t.Fatalf("planned uploads = %#v, want 1", manifest["planned_uploads"])
	}
}

func TestContainsMarkdownPipeTableOutsideFence(t *testing.T) {
	if !containsMarkdownPipeTableOutsideFence("| A | B |\n|---|---|\n| 1 | 2 |") {
		t.Fatal("expected raw pipe table to be detected")
	}
	if containsMarkdownPipeTableOutsideFence("```md\n| A | B |\n|---|---|\n```") {
		t.Fatal("fenced table should be ignored")
	}
}

func TestCloneJSONMapClonesNestedSlices(t *testing.T) {
	original := map[string]any{
		"report": map[string]any{
			"tables": []any{
				map[string]any{
					"rows": []any{
						[]any{map[string]any{"value": "old"}},
					},
				},
			},
		},
	}

	cloned := storepkg.CloneJSONMap(original)
	cloned["report"].(map[string]any)["tables"].([]any)[0].(map[string]any)["rows"].([]any)[0].([]any)[0].(map[string]any)["value"] = "new"

	got := original["report"].(map[string]any)["tables"].([]any)[0].(map[string]any)["rows"].([]any)[0].([]any)[0].(map[string]any)["value"]
	if got != "old" {
		t.Fatalf("original nested value mutated to %q", got)
	}
}

func TestCloneJSONMapPreservesNil(t *testing.T) {
	if got := storepkg.CloneJSONMap(nil); got != nil {
		t.Fatalf("CloneJSONMap(nil) = %#v, want nil", got)
	}
}

func TestSlackReportResumeHelpersReusePostedMainAndUploadedFiles(t *testing.T) {
	result := map[string]any{
		"render_manifest": map[string]any{
			"main_message": map[string]any{
				"status":     "posted",
				"channel_id": "C123",
				"ts":         "123.456",
				"source_ref": "slack:C123:123.456",
			},
			"uploads": []any{
				map[string]any{"id": "table-0", "status": "uploaded", "source_ref": "slack_file:F123"},
				map[string]any{"id": "table-1", "status": "failed", "error": "rate limited"},
			},
		},
	}
	if !slackReportResultHasPostedMain(result) {
		t.Fatal("expected posted main message to be detected")
	}
	uploads := slackReportPreviousUploadsByID(mapValue(result["render_manifest"]))
	if !slackReportUploadSucceeded(uploads["table-0"]) {
		t.Fatalf("expected table-0 upload to be reusable: %#v", uploads["table-0"])
	}
	if slackReportUploadSucceeded(uploads["table-1"]) {
		t.Fatalf("failed upload should not be reusable: %#v", uploads["table-1"])
	}
	if !slackReportResultHasUploadFailures(result) {
		t.Fatal("expected failed upload to be detected")
	}
}
