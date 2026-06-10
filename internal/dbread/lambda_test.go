package dbread

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"

	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

type mockLambdaClient struct {
	input *lambda.InvokeInput
	out   *lambda.InvokeOutput
	err   error
}

func (m *mockLambdaClient) Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error) {
	m.input = params
	return m.out, m.err
}

func TestPublicSourcesHideExecutionBoundary(t *testing.T) {
	registry, err := LoadRegistry(`{"targets":[{"id":"depin-prod","placement":"prod","lambda_function_name":"rsi-db-read-prod"},{"id":"depin-stage","placement":"stage","dsn_env":"STAGE_DSN"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	sources := registry.PublicSources()
	raw, _ := json.Marshal(sources)
	if strings.Contains(string(raw), "lambda") || strings.Contains(string(raw), "execution") {
		t.Fatalf("public sources should not expose execution boundary: %s", raw)
	}
}

func TestLambdaInvokerSendsStableJobEnvelope(t *testing.T) {
	payload, _ := json.Marshal(LambdaResult{
		Status:     "succeeded",
		Validation: &SQLValidation{OK: true, SQLSHA256: "sha256:822ae07d4783158bc1912bb623e5107cc9002d519e1143a9c200ed6ee18b6d0f"},
	})
	client := &mockLambdaClient{out: &lambda.InvokeOutput{Payload: payload}}
	invoker := NewLambdaInvokerWithClient(client)
	request := storepkg.DBReadRequest{
		ID:              "dbread_1",
		Target:          "depin-prod",
		SQL:             "select 1",
		SQLSHA256:       "sha256:822ae07d4783158bc1912bb623e5107cc9002d519e1143a9c200ed6ee18b6d0f",
		Caps:            storepkg.DBReadCaps{MaxRows: 10, MaxBytes: 1024, TimeoutSeconds: 5},
		ExpiresAt:       time.Now().UTC().Add(time.Hour),
		LeaseGeneration: 2,
	}
	_, err := invoker.Validate(context.Background(), Target{ID: "depin-prod", LambdaFunction: "rsi-db-read-prod"}, storepkg.DBReadLease{Request: request, Token: "lease-token"})
	if err != nil {
		t.Fatal(err)
	}
	if got := aws.ToString(client.input.FunctionName); got != "rsi-db-read-prod" {
		t.Fatalf("function name = %q", got)
	}
	var job LambdaJob
	if err := json.Unmarshal(client.input.Payload, &job); err != nil {
		t.Fatal(err)
	}
	if job.Purpose != LambdaPurposeValidate || job.Target != "depin-prod" || job.LeaseToken != "lease-token" || job.LeaseGeneration != 2 {
		t.Fatalf("unexpected job envelope: %+v", job)
	}
}

func TestNewLambdaJobAttributesAutoApprovalOnlyForApprovedExecute(t *testing.T) {
	now := time.Now().UTC()
	autoApproved := storepkg.DBReadRequest{
		ID:         "dbread_1",
		Target:     "depin-prod",
		SQL:        "select 1",
		SQLSHA256:  "sha256:abc",
		ApprovedAt: &now,
	}
	if got := NewLambdaJob(LambdaPurposeExecute, storepkg.DBReadLease{Request: autoApproved, Token: "lease"}).ApprovedBySlackUserID; got != AutoApprovalRef {
		t.Fatalf("execute job approver = %q, want %q", got, AutoApprovalRef)
	}
	if got := NewLambdaJob(LambdaPurposeValidate, storepkg.DBReadLease{Request: autoApproved, Token: "lease"}).ApprovedBySlackUserID; got != "" {
		t.Fatalf("validate job must not carry approval attribution, got %q", got)
	}

	manual := autoApproved
	manual.ApprovedBySlackUserID = "UADMIN"
	if got := NewLambdaJob(LambdaPurposeExecute, storepkg.DBReadLease{Request: manual, Token: "lease"}).ApprovedBySlackUserID; got != "UADMIN" {
		t.Fatalf("manual approver must be preserved, got %q", got)
	}

	unapproved := autoApproved
	unapproved.ApprovedAt = nil
	if got := NewLambdaJob(LambdaPurposeExecute, storepkg.DBReadLease{Request: unapproved, Token: "lease"}).ApprovedBySlackUserID; got != "" {
		t.Fatalf("unapproved request must not be attributed, got %q", got)
	}
}

func TestLambdaInvokerExecuteSendsAutoApprovalMetadata(t *testing.T) {
	payload, _ := json.Marshal(LambdaResult{Status: "succeeded", Result: &DBResult{RowCount: 1}})
	client := &mockLambdaClient{out: &lambda.InvokeOutput{Payload: payload}}
	invoker := NewLambdaInvokerWithClient(client)
	now := time.Now().UTC()
	request := storepkg.DBReadRequest{
		ID:              "dbread_1",
		Target:          "depin-prod",
		SQL:             "select 1",
		SQLSHA256:       "sha256:822ae07d4783158bc1912bb623e5107cc9002d519e1143a9c200ed6ee18b6d0f",
		Caps:            storepkg.DBReadCaps{MaxRows: 10, MaxBytes: 1024, TimeoutSeconds: 5},
		ExpiresAt:       time.Now().UTC().Add(time.Hour),
		LeaseGeneration: 2,
		ApprovedAt:      &now,
	}
	if _, err := invoker.Execute(context.Background(), Target{ID: "depin-prod", LambdaFunction: "rsi-db-read-prod"}, storepkg.DBReadLease{Request: request, Token: "lease-token"}); err != nil {
		t.Fatal(err)
	}
	var job LambdaJob
	if err := json.Unmarshal(client.input.Payload, &job); err != nil {
		t.Fatal(err)
	}
	if job.ApprovedBySlackUserID != AutoApprovalRef {
		t.Fatalf("execute payload approver = %q, want %q", job.ApprovedBySlackUserID, AutoApprovalRef)
	}
	if job.ApprovedAt == nil {
		t.Fatal("execute payload must carry approved_at")
	}
}

func TestLambdaHandlerAcceptsAutoApprovalAttribution(t *testing.T) {
	now := time.Now().UTC()
	sql := "select 1"
	result, err := HandleLambdaJob(context.Background(), Registry{Targets: map[string]Target{"depin-prod": {ID: "depin-prod", DSN: "postgres://127.0.0.1:1/refused", Caps: storepkg.DBReadCaps{TimeoutSeconds: 1}}}}, LambdaJob{
		Purpose:               LambdaPurposeExecute,
		Target:                "depin-prod",
		RequestID:             "dbread_1",
		SQL:                   sql,
		SQLSHA256:             sha256String(sql),
		LeaseToken:            "lease-token",
		LeaseGeneration:       1,
		ExpiresAt:             time.Now().UTC().Add(time.Hour),
		ApprovedBySlackUserID: AutoApprovalRef,
		ApprovedAt:            &now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ErrorCode == "approval_required" {
		t.Fatalf("auto-approval attribution must pass the approval gate, got %+v", result)
	}
}

func TestLambdaHandlerRequiresApprovalForExecution(t *testing.T) {
	result, err := HandleLambdaJob(context.Background(), Registry{Targets: map[string]Target{"depin-prod": {ID: "depin-prod", DSN: "postgres://example"}}}, LambdaJob{
		Purpose:         LambdaPurposeExecute,
		Target:          "depin-prod",
		RequestID:       "dbread_1",
		SQL:             "select 1",
		SQLSHA256:       "sha256:822ae07d4783158bc1912bb623e5107cc9002d519e1143a9c200ed6ee18b6d0f",
		LeaseToken:      "lease-token",
		LeaseGeneration: 1,
		ExpiresAt:       time.Now().UTC().Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "failed" || result.ErrorCode != "approval_required" {
		t.Fatalf("expected approval_required failure, got %+v", result)
	}
}
