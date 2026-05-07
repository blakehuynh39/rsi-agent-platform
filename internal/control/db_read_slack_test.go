package control

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

func TestDBReadApprovalBlocksShowExactSQLAndApproverScope(t *testing.T) {
	sql := "SELECT language_code, COUNT(*) FROM scripts GROUP BY language_code"
	request := storepkg.DBReadRequest{
		ID:        "dbread_1",
		Target:    "depin-prod",
		Purpose:   "query",
		SQL:       sql,
		SQLSHA256: "sha256:1234567890abcdef1234567890abcdef",
		Requester: "user:U123",
		ExpiresAt: time.Date(2026, 5, 7, 20, 0, 0, 0, time.UTC),
		Caps:      storepkg.DBReadCaps{MaxRows: 20, MaxBytes: 4096, TimeoutSeconds: 10},
	}
	attempt := storepkg.DBReadValidationAttempt{ID: "dbreadval_1"}
	raw, err := json.Marshal(dbReadApprovalBlocks(request, attempt, sql))
	if err != nil {
		t.Fatal(err)
	}
	body := strings.ReplaceAll(string(raw), "\\u003c", "<")
	body = strings.ReplaceAll(body, "\\u003e", ">")
	for _, want := range []string{
		"Approve DB read?",
		"Only authorized approvers can approve or deny",
		"Exact SQL to run",
		sql,
		"<@U123> via Hermes",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("approval blocks missing %q in %s", want, body)
		}
	}
}

func TestPostDBReadApprovalCardIsIdempotentWhenSlackMessageExists(t *testing.T) {
	request := storepkg.DBReadRequest{
		ID:                    "dbread_1",
		Target:                "depin-prod",
		Purpose:               "query",
		SQL:                   "SELECT count(*) FROM scripts",
		SQLSHA256:             "sha256:1234567890abcdef1234567890abcdef",
		Requester:             "user:U123",
		ChannelID:             "C123",
		ThreadTS:              "171000001.000100",
		SlackMessageChannelID: "C123",
		SlackMessageTS:        "171000001.000200",
		ExpiresAt:             time.Date(2026, 5, 7, 20, 0, 0, 0, time.UTC),
		Caps:                  storepkg.DBReadCaps{MaxRows: 20, MaxBytes: 4096, TimeoutSeconds: 10},
	}
	poster := &fakeSlackPoster{}
	err := postDBReadApprovalCard(context.Background(), config.Config{}, storepkg.NewMemoryStore(), poster, request, storepkg.DBReadValidationAttempt{ID: "dbreadval_1"}, request.SQL)
	if err != nil {
		t.Fatal(err)
	}
	if len(poster.calls) != 0 {
		t.Fatalf("expected existing Slack approval card to be left untouched, got %d calls", len(poster.calls))
	}
}

func TestDBReadResultUpdateFormatsSampleAsTableAndRemovesButtons(t *testing.T) {
	request := storepkg.DBReadRequest{
		ID:                         "dbread_1",
		Target:                     "depin-prod",
		Purpose:                    "query",
		SQL:                        "SELECT language_code, COUNT(*) AS transcript_count FROM scripts GROUP BY language_code",
		SQLSHA256:                  "sha256:1234567890abcdef1234567890abcdef",
		Requester:                  "user:U123",
		CurrentValidationAttemptID: "dbreadval_1",
		ApprovedBySlackUserID:      "UADMIN",
		SlackMessageChannelID:      "C123",
		SlackMessageTS:             "171000001.000200",
		ResultSample: []map[string]string{
			{"language_code": "hi", "transcript_count": "350000"},
			{"language_code": "bn", "transcript_count": "350000"},
		},
	}
	poster := &fakeSlackPoster{}
	if err := updateDBReadSlackCard(context.Background(), poster, request, "succeeded; rows=2 truncated=false"); err != nil {
		t.Fatal(err)
	}
	if len(poster.calls) != 1 {
		t.Fatalf("expected one Slack update, got %d", len(poster.calls))
	}
	body := poster.calls[0].values.Encode()
	body = strings.ReplaceAll(body, "%3C", "<")
	body = strings.ReplaceAll(body, "%3E", ">")
	for _, want := range []string{
		"language_code",
		"transcript_count",
		"350000",
		"Approved+by",
		"UADMIN",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("result card missing %q in %s", want, body)
		}
	}
	if strings.Contains(body, dbReadSlackApproveAction) || strings.Contains(body, dbReadSlackDenyAction) {
		t.Fatalf("finalized result update should not keep approval actions: %s", body)
	}
}
