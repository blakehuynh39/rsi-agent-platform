package dbread

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"

	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

const (
	LambdaPurposeValidate = "validate"
	LambdaPurposeExecute  = "execute"
)

type LambdaJob struct {
	Purpose               string                         `json:"purpose"`
	Target                string                         `json:"target"`
	RequestID             string                         `json:"request_id"`
	SQL                   string                         `json:"sql"`
	SQLSHA256             string                         `json:"sql_sha256"`
	LeaseToken            string                         `json:"lease_token"`
	LeaseGeneration       int                            `json:"lease_generation"`
	Caps                  storepkg.DBReadCaps            `json:"caps"`
	Redaction             storepkg.DBReadRedactionPolicy `json:"redaction,omitempty"`
	ExpiresAt             time.Time                      `json:"expires_at"`
	ApprovedBySlackUserID string                         `json:"approved_by_slack_user_id,omitempty"`
	ApprovedAt            *time.Time                     `json:"approved_at,omitempty"`
}

type LambdaResult struct {
	Status       string         `json:"status"`
	Validation   *SQLValidation `json:"validation,omitempty"`
	Result       *DBResult      `json:"result,omitempty"`
	ErrorCode    string         `json:"error_code,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
}

type LambdaInvokeAPI interface {
	Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error)
}

type LambdaInvoker struct {
	client LambdaInvokeAPI
}

func NewLambdaInvoker(ctx context.Context) (*LambdaInvoker, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &LambdaInvoker{client: lambda.NewFromConfig(cfg)}, nil
}

func NewLambdaInvokerWithClient(client LambdaInvokeAPI) *LambdaInvoker {
	return &LambdaInvoker{client: client}
}

func (i *LambdaInvoker) Validate(ctx context.Context, target Target, lease storepkg.DBReadLease) (SQLValidation, error) {
	job := NewLambdaJob(LambdaPurposeValidate, lease)
	result, err := i.invoke(ctx, target, job)
	if err != nil {
		return SQLValidation{SQLSHA256: lease.Request.SQLSHA256, ErrorCode: "lambda_invoke_failed", Message: err.Error()}, err
	}
	if result.Validation != nil {
		return *result.Validation, nil
	}
	if result.Status == "succeeded" {
		return SQLValidation{OK: true, SQLSHA256: lease.Request.SQLSHA256, Preview: truncateSQL(lease.Request.SQL, 1800)}, nil
	}
	return SQLValidation{SQLSHA256: lease.Request.SQLSHA256, ErrorCode: firstNonEmpty(result.ErrorCode, "lambda_validation_failed"), Message: result.ErrorMessage}, nil
}

func (i *LambdaInvoker) Execute(ctx context.Context, target Target, lease storepkg.DBReadLease) (DBResult, error) {
	job := NewLambdaJob(LambdaPurposeExecute, lease)
	result, err := i.invoke(ctx, target, job)
	if err != nil {
		return DBResult{}, err
	}
	if result.Result != nil && result.Status == "succeeded" {
		return *result.Result, nil
	}
	return DBResult{}, fmt.Errorf("%s", firstNonEmpty(result.ErrorMessage, result.ErrorCode, "lambda execution failed"))
}

func NewLambdaJob(purpose string, lease storepkg.DBReadLease) LambdaJob {
	request := lease.Request
	return LambdaJob{
		Purpose:               purpose,
		Target:                request.Target,
		RequestID:             request.ID,
		SQL:                   request.SQL,
		SQLSHA256:             request.SQLSHA256,
		LeaseToken:            lease.Token,
		LeaseGeneration:       request.LeaseGeneration,
		Caps:                  request.Caps,
		Redaction:             request.Redaction,
		ExpiresAt:             request.ExpiresAt,
		ApprovedBySlackUserID: request.ApprovedBySlackUserID,
		ApprovedAt:            request.ApprovedAt,
	}
}

func (i *LambdaInvoker) invoke(ctx context.Context, target Target, job LambdaJob) (LambdaResult, error) {
	if i == nil || i.client == nil {
		return LambdaResult{}, fmt.Errorf("lambda invoker is not configured")
	}
	functionName := strings.TrimSpace(target.LambdaFunction)
	if functionName == "" {
		return LambdaResult{}, fmt.Errorf("target %q has no lambda_function_name", target.ID)
	}
	payload, err := json.Marshal(job)
	if err != nil {
		return LambdaResult{}, err
	}
	input := &lambda.InvokeInput{
		FunctionName: aws.String(functionName),
		Payload:      payload,
	}
	if strings.TrimSpace(target.LambdaQualifier) != "" {
		input.Qualifier = aws.String(strings.TrimSpace(target.LambdaQualifier))
	}
	out, err := i.client.Invoke(ctx, input)
	if err != nil {
		return LambdaResult{}, err
	}
	if out.FunctionError != nil && strings.TrimSpace(*out.FunctionError) != "" {
		return LambdaResult{}, fmt.Errorf("lambda function error: %s", strings.TrimSpace(*out.FunctionError))
	}
	var result LambdaResult
	if err := json.Unmarshal(out.Payload, &result); err != nil {
		return LambdaResult{}, err
	}
	if result.Status == "" {
		result.Status = "failed"
	}
	return result, nil
}

func HandleLambdaJob(ctx context.Context, registry Registry, job LambdaJob) (LambdaResult, error) {
	if err := validateLambdaJobEnvelope(job); err != nil {
		return LambdaResult{Status: "failed", ErrorCode: "invalid_envelope", ErrorMessage: err.Error()}, nil
	}
	target, ok := registry.Target(job.Target)
	if !ok {
		return LambdaResult{Status: "failed", ErrorCode: "unknown_target", ErrorMessage: "target is not configured"}, nil
	}
	if target.ID != job.Target {
		return LambdaResult{Status: "failed", ErrorCode: "target_mismatch", ErrorMessage: "target mismatch"}, nil
	}
	if strings.TrimSpace(target.DSN) == "" && strings.TrimSpace(target.DSNSecretARN) == "" {
		return LambdaResult{Status: "failed", ErrorCode: "target_dsn_missing", ErrorMessage: "target DSN is not configured"}, nil
	}
	request := storepkg.DBReadRequest{
		ID:                    job.RequestID,
		Target:                job.Target,
		SQL:                   job.SQL,
		SQLSHA256:             job.SQLSHA256,
		Caps:                  job.Caps,
		Redaction:             job.Redaction,
		ExpiresAt:             job.ExpiresAt,
		ApprovedBySlackUserID: job.ApprovedBySlackUserID,
		ApprovedAt:            job.ApprovedAt,
	}
	switch strings.TrimSpace(job.Purpose) {
	case LambdaPurposeValidate:
		validation := ValidateAgainstTarget(ctx, target, job.SQL)
		if validation.SQLSHA256 != job.SQLSHA256 {
			validation.OK = false
			validation.ErrorCode = "sql_hash_mismatch"
			validation.Message = "SQL hash mismatch"
		}
		if validation.OK {
			return LambdaResult{Status: "succeeded", Validation: &validation}, nil
		}
		return LambdaResult{Status: "failed", Validation: &validation, ErrorCode: validation.ErrorCode, ErrorMessage: validation.Message}, nil
	case LambdaPurposeExecute:
		if strings.TrimSpace(job.ApprovedBySlackUserID) == "" || job.ApprovedAt == nil {
			return LambdaResult{Status: "failed", ErrorCode: "approval_required", ErrorMessage: "execution requires approval metadata"}, nil
		}
		result, err := ExecuteRead(ctx, target, request)
		if err != nil {
			return LambdaResult{Status: "failed", ErrorCode: "execution_failed", ErrorMessage: sanitizeError(err)}, nil
		}
		return LambdaResult{Status: "succeeded", Result: &result}, nil
	default:
		return LambdaResult{Status: "failed", ErrorCode: "invalid_purpose", ErrorMessage: "purpose must be validate or execute"}, nil
	}
}

func validateLambdaJobEnvelope(job LambdaJob) error {
	if strings.TrimSpace(job.Purpose) != LambdaPurposeValidate && strings.TrimSpace(job.Purpose) != LambdaPurposeExecute {
		return fmt.Errorf("purpose must be validate or execute")
	}
	if strings.TrimSpace(job.Target) == "" {
		return fmt.Errorf("target is required")
	}
	if strings.TrimSpace(job.RequestID) == "" {
		return fmt.Errorf("request_id is required")
	}
	if strings.TrimSpace(job.SQL) == "" {
		return fmt.Errorf("sql is required")
	}
	if strings.TrimSpace(job.SQLSHA256) == "" {
		return fmt.Errorf("sql_sha256 is required")
	}
	if sha256String(job.SQL) != strings.TrimSpace(job.SQLSHA256) {
		return fmt.Errorf("sql_sha256 does not match sql")
	}
	if strings.TrimSpace(job.LeaseToken) == "" {
		return fmt.Errorf("lease_token is required")
	}
	if job.LeaseGeneration <= 0 {
		return fmt.Errorf("lease_generation is required")
	}
	if job.ExpiresAt.IsZero() || time.Now().UTC().After(job.ExpiresAt) {
		return fmt.Errorf("job is expired")
	}
	return nil
}

func sha256String(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return "sha256:" + hex.EncodeToString(sum[:])
}
