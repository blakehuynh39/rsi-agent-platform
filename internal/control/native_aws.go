package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"

	"github.com/piplabs/rsi-agent-platform/internal/config"
)

var awsNativeReadTimeout = 30 * time.Second

type awsReadRequest struct {
	Account   string         `json:"account"`
	Region    string         `json:"region"`
	Service   string         `json:"service"`
	Operation string         `json:"operation"`
	Params    map[string]any `json:"params,omitempty"`
}

type awsNativeRunnerFunc func(ctx context.Context, cfg config.Config, req awsReadRequest) (any, error)

var awsNativeRunner awsNativeRunnerFunc = runAWSNativeRead

var awsSupportedReadServices = map[string]bool{
	"sts":         true,
	"rds":         true,
	"cloudwatch":  true,
	"logs":        true,
	"cloudtrail":  true,
	"ec2":         true,
	"eks":         true,
	"elbv2":       true,
	"autoscaling": true,
}

var awsForbiddenServices = map[string]bool{
	"secretsmanager": true,
	"ssm":            true,
	"kms":            true,
	"s3":             true,
	"s3api":          true,
	"ecr":            true,
	"iam":            true,
}

var awsForbiddenOperationTokens = []string{
	"secret",
	"decrypt",
	"object",
	"authorization-token",
	"password",
}

var awsAllowedReadOperationPrefixes = []string{
	"describe-",
	"list-",
	"get-",
	"lookup-",
	"head-",
	"search-",
	"batch-get-",
}

var awsBlockedReadOps = map[string]map[string]bool{
	"logs": {
		"filter-log-events": true,
		"get-log-events":    true,
		"get-log-record":    true,
		"get-query-results": true,
		"start-query":       true,
		"start-live-tail":   true,
	},
	"ec2": {
		"describe-instance-attribute":       true,
		"describe-launch-template-versions": true,
		"get-launch-template-data":          true,
		"get-console-output":                true,
		"get-console-screenshot":            true,
		"get-password-data":                 true,
		"get-serial-console-output":         true,
	},
	"autoscaling": {
		"describe-launch-configurations": true,
	},
	"sts": {
		"get-session-token":    true,
		"get-federation-token": true,
	},
}

func executeAWSNativeToolAction(ctx context.Context, cfg config.Config, input nativeToolActionRequest) (any, string, string, string, map[string]any, int, error) {
	req, err := awsReadRequestFromInput(cfg, input)
	if err != nil {
		return nil, "", "", "", nativeToolNoopEffect("invalid_aws_arguments"), http.StatusBadRequest, err
	}
	if validationErr, status := validateAWSReadRequest(req); validationErr != nil {
		return map[string]any{
			"account":   req.Account,
			"region":    req.Region,
			"service":   req.Service,
			"operation": req.Operation,
			"error":     validationErr.Error(),
		}, "", awsReadSourceRef(req), "", nativeToolNoopEffect("aws_policy_denied"), status, validationErr
	}
	callCtx, cancel := context.WithTimeout(ctx, awsNativeReadTimeout)
	defer cancel()
	output, runErr := awsNativeRunner(callCtx, cfg, req)
	if runErr != nil {
		if callCtx.Err() == context.DeadlineExceeded {
			return nil, "", awsReadSourceRef(req), "", nativeToolNoopEffect("aws_read_timeout"), http.StatusGatewayTimeout, errors.New("AWS read timed out")
		}
	}
	redacted := redactAWSJSON(output)
	limited := limitAWSOutput(redacted, cfg.AWSReadMaxOutputBytes)
	response := map[string]any{
		"account":   req.Account,
		"region":    req.Region,
		"service":   req.Service,
		"operation": req.Operation,
		"params":    redactAWSJSON(req.Params),
		"output":    limited,
	}
	if runErr != nil {
		response["error"] = runErr.Error()
		return response, "", awsReadSourceRef(req), "", nativeToolNoopEffect("aws_read_failed"), http.StatusBadGateway, fmt.Errorf("AWS read failed: %w", runErr)
	}
	return response, awsReadSummary(req, output), awsReadSourceRef(req), "", map[string]any{"status": "not_applicable"}, http.StatusOK, nil
}

func validateAWSReadPolicy(input nativeToolActionRequest) (error, int) {
	req, err := awsReadRequestFromInput(config.Config{AWSReadDefaultRegion: "us-east-1"}, input)
	if err != nil {
		return err, http.StatusBadRequest
	}
	return validateAWSReadRequest(req)
}

func validateAWSReadRequest(req awsReadRequest) (error, int) {
	if req.Service == "" {
		return errors.New("rsi_aws.read requires service"), http.StatusBadRequest
	}
	if req.Operation == "" {
		return errors.New("rsi_aws.read requires operation"), http.StatusBadRequest
	}
	if awsForbiddenServices[req.Service] {
		return fmt.Errorf("AWS service %s is blocked because it can expose secrets or raw private objects", req.Service), http.StatusForbidden
	}
	if !awsSupportedReadServices[req.Service] {
		return fmt.Errorf("AWS service %s is not available in the read-only diagnostic gateway", req.Service), http.StatusBadRequest
	}
	if req.Service == "sts" && req.Operation != "get-caller-identity" {
		return fmt.Errorf("AWS operation %s.%s is blocked because STS reads are limited to caller identity", req.Service, req.Operation), http.StatusForbidden
	}
	if awsBlockedReadOps[req.Service][req.Operation] {
		return fmt.Errorf("AWS operation %s.%s is blocked because it can expose raw log, console, or credential-bearing content", req.Service, req.Operation), http.StatusForbidden
	}
	for _, token := range awsForbiddenOperationTokens {
		if strings.Contains(req.Operation, token) {
			return fmt.Errorf("AWS operation %s is blocked by read-only secret-safety policy", req.Operation), http.StatusForbidden
		}
	}
	if !awsOperationHasReadPrefix(req.Operation) {
		return fmt.Errorf("AWS operation %s.%s is not a read-only diagnostic operation", req.Service, req.Operation), http.StatusBadRequest
	}
	if err := rejectSensitiveAWSParamKeys(req.Params); err != nil {
		return err, http.StatusForbidden
	}
	return nil, http.StatusOK
}

func awsReadRequestFromInput(cfg config.Config, input nativeToolActionRequest) (awsReadRequest, error) {
	account := normalizeAWSAccount(firstNonEmpty(stringArg(input.Arguments, "account"), stringArg(input.Arguments, "environment"), stringArg(input.Arguments, "env"), cfg.Environment))
	if account == "" {
		account = "stage"
	}
	region := strings.TrimSpace(firstNonEmpty(stringArg(input.Arguments, "region"), cfg.AWSReadDefaultRegion, "us-east-1"))
	service := normalizeAWSService(stringArg(input.Arguments, "service"))
	operation := normalizeAWSOperation(firstNonEmpty(stringArg(input.Arguments, "operation"), stringArg(input.Arguments, "op")))
	params := mapArg(input.Arguments, "params")
	if params == nil {
		params = map[string]any{}
	}
	if account != "stage" && account != "prod" {
		return awsReadRequest{}, fmt.Errorf("unsupported AWS account %q; use stage or prod", account)
	}
	return awsReadRequest{
		Account:   account,
		Region:    region,
		Service:   service,
		Operation: operation,
		Params:    params,
	}, nil
}

func runAWSNativeRead(ctx context.Context, cfg config.Config, req awsReadRequest) (any, error) {
	awsCfg, err := awsConfigForRead(ctx, cfg, req)
	if err != nil {
		return nil, err
	}
	client := awsReadServiceClient(awsCfg, req.Service)
	if client == nil {
		return nil, fmt.Errorf("AWS service %s is not implemented in the read-only diagnostic gateway", req.Service)
	}
	return callAWSReadOperation(ctx, client, req.Operation, req.Params)
}

func awsReadServiceClient(awsCfg aws.Config, service string) any {
	switch service {
	case "sts":
		return sts.NewFromConfig(awsCfg)
	case "rds":
		return rds.NewFromConfig(awsCfg)
	case "cloudwatch":
		return cloudwatch.NewFromConfig(awsCfg)
	case "logs":
		return cloudwatchlogs.NewFromConfig(awsCfg)
	case "cloudtrail":
		return cloudtrail.NewFromConfig(awsCfg)
	case "ec2":
		return ec2.NewFromConfig(awsCfg)
	case "eks":
		return eks.NewFromConfig(awsCfg)
	case "elbv2":
		return elasticloadbalancingv2.NewFromConfig(awsCfg)
	case "autoscaling":
		return autoscaling.NewFromConfig(awsCfg)
	}
	return nil
}

func callAWSReadOperation(ctx context.Context, client any, operation string, params map[string]any) (any, error) {
	methodName := awsOperationMethodName(operation)
	method := reflect.ValueOf(client).MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("AWS read operation %s is not implemented by service client", operation)
	}
	methodType := method.Type()
	if methodType.NumIn() < 2 {
		return nil, fmt.Errorf("AWS read operation %s has unsupported SDK method shape", operation)
	}
	inputType := methodType.In(1)
	if inputType.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("AWS read operation %s has unsupported input type", operation)
	}
	input := reflect.New(inputType.Elem())
	normalized := normalizeAWSParamValue(params)
	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, input.Interface()); err != nil {
		return nil, err
	}
	results := method.Call([]reflect.Value{reflect.ValueOf(ctx), input})
	if len(results) != 2 {
		return nil, fmt.Errorf("AWS read operation %s has unsupported SDK result shape", operation)
	}
	if !results[1].IsNil() {
		err, ok := results[1].Interface().(error)
		if !ok {
			return nil, fmt.Errorf("AWS read operation %s returned non-error failure value", operation)
		}
		return nil, err
	}
	return results[0].Interface(), nil
}

func awsOperationHasReadPrefix(operation string) bool {
	for _, prefix := range awsAllowedReadOperationPrefixes {
		if strings.HasPrefix(operation, prefix) {
			return true
		}
	}
	return false
}

func awsOperationMethodName(operation string) string {
	parts := strings.Split(operation, "-")
	var builder strings.Builder
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.EqualFold(part, "db") {
			builder.WriteString("DB")
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}
	return builder.String()
}

func awsConfigForRead(ctx context.Context, cfg config.Config, req awsReadRequest) (aws.Config, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(req.Region))
	if err != nil {
		return aws.Config{}, err
	}
	roleARN := awsRoleARNForAccount(cfg, req.Account)
	if roleARN == "" {
		if req.Account == "stage" {
			return awsCfg, nil
		}
		return aws.Config{}, fmt.Errorf("no role ARN configured for account %q", req.Account)
	}
	client := sts.NewFromConfig(awsCfg)
	assumed, err := client.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleARN),
		RoleSessionName: aws.String("rsi-native-aws-read-" + req.Account),
		DurationSeconds: aws.Int32(900),
	})
	if err != nil {
		return aws.Config{}, err
	}
	if assumed.Credentials == nil {
		return aws.Config{}, errors.New("assume-role response did not include credentials")
	}
	creds := assumed.Credentials
	awsCfg.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
		aws.ToString(creds.AccessKeyId),
		aws.ToString(creds.SecretAccessKey),
		aws.ToString(creds.SessionToken),
	))
	return awsCfg, nil
}

func awsRoleARNForAccount(cfg config.Config, account string) string {
	switch normalizeAWSAccount(account) {
	case "stage":
		return strings.TrimSpace(cfg.AWSReadStageRoleARN)
	case "prod":
		return strings.TrimSpace(cfg.AWSReadProdRoleARN)
	default:
		return ""
	}
}

func decodeAWSInput[T any](params map[string]any) (*T, error) {
	normalized := normalizeAWSParamValue(params)
	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func normalizeAWSParamValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[awsParamFieldName(key)] = normalizeAWSParamValue(item)
		}
		return out
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, normalizeAWSParamValue(item))
		}
		return out
	default:
		return value
	}
}

func awsParamFieldName(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return key
	}
	if !strings.ContainsAny(key, "-_ ") {
		return strings.ToUpper(key[:1]) + key[1:]
	}
	parts := regexp.MustCompile(`[-_\s]+`).Split(key, -1)
	var builder strings.Builder
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}
	return builder.String()
}

func normalizeAWSAccount(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "staging", "stage", "use1-stage":
		return "stage"
	case "production", "prod", "use1-prod":
		return "prod"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func normalizeAWSService(value string) string {
	service := strings.ToLower(strings.TrimSpace(value))
	service = strings.ReplaceAll(service, "_", "-")
	switch service {
	case "cloudwatchlogs", "cloud-watch-logs", "logs":
		return "logs"
	case "elasticloadbalancingv2", "elastic-load-balancing-v2", "elbv2":
		return "elbv2"
	case "auto-scaling", "autoscaling":
		return "autoscaling"
	default:
		return service
	}
}

func normalizeAWSOperation(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	var builder strings.Builder
	for index, char := range value {
		if index > 0 && char >= 'A' && char <= 'Z' {
			prev := rune(value[index-1])
			if prev != '-' && !(prev >= 'A' && prev <= 'Z') {
				builder.WriteRune('-')
			}
		}
		builder.WriteRune(char)
	}
	return strings.ToLower(builder.String())
}

func rejectSensitiveAWSParamKeys(value any) error {
	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			if awsSensitiveKey(key) {
				return fmt.Errorf("AWS read param %s is blocked by secret-safety policy", key)
			}
			if err := rejectSensitiveAWSParamKeys(item); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range typed {
			if err := rejectSensitiveAWSParamKeys(item); err != nil {
				return err
			}
		}
	}
	return nil
}

func redactAWSJSON(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			if awsSensitiveOutputKey(key) {
				out[key] = "[REDACTED]"
				continue
			}
			out[key] = redactAWSJSON(item)
		}
		return out
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, redactAWSJSON(item))
		}
		return out
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		if str, ok := typed.(string); ok {
			var embeddedJSON any
			if err := json.Unmarshal([]byte(str), &embeddedJSON); err == nil {
				switch embeddedJSON.(type) {
				case map[string]any, []any:
					redacted := redactAWSJSON(embeddedJSON)
					if redactedBytes, err := json.Marshal(redacted); err == nil {
						return string(redactedBytes)
					}
				}
			}
		}
		return typed
	case *ststypes.Credentials:
		return "[REDACTED]"
	default:
		var jsonValue any
		raw, err := json.Marshal(typed)
		if err == nil && json.Unmarshal(raw, &jsonValue) == nil {
			switch jsonValue.(type) {
			case map[string]any, []any:
				return redactAWSJSON(jsonValue)
			default:
				return jsonValue
			}
		}
		return typed
	}
}

func awsSensitiveOutputKey(key string) bool {
	return awsSensitiveKey(key)
}

func awsSensitiveKey(key string) bool {
	normalized := awsSensitiveKeyFragment(key)
	for _, token := range []string{"secret", "password", "privatekey", "accesskey", "sessiontoken", "authorization", "credential"} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func awsSensitiveKeyFragment(key string) string {
	var builder strings.Builder
	for _, char := range key {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			builder.WriteRune(unicode.ToLower(char))
		}
	}
	return builder.String()
}

func limitAWSOutput(value any, maxBytes int) any {
	if maxBytes <= 0 {
		maxBytes = 128000
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return value
	}
	if len(raw) <= maxBytes {
		var normalized any
		if json.Unmarshal(raw, &normalized) == nil {
			return normalized
		}
		return value
	}
	prefix := string(raw[:maxBytes])
	return map[string]any{
		"truncated":       true,
		"original_bytes":  len(raw),
		"returned_bytes":  maxBytes,
		"raw_json_prefix": prefix,
	}
}

func awsReadSourceRef(req awsReadRequest) string {
	return fmt.Sprintf("aws:%s:%s:%s:%s", req.Account, req.Region, req.Service, req.Operation)
}

func awsReadSummary(req awsReadRequest, output any) string {
	count := awsOutputCount(output)
	if count > 0 {
		return fmt.Sprintf("read AWS %s.%s in %s/%s (%d item(s))", req.Service, req.Operation, req.Account, req.Region, count)
	}
	return fmt.Sprintf("read AWS %s.%s in %s/%s", req.Service, req.Operation, req.Account, req.Region)
}

func awsOutputCount(output any) int {
	redacted := redactAWSJSON(output)
	switch typed := redacted.(type) {
	case map[string]any:
		for _, value := range typed {
			if items, ok := value.([]any); ok {
				return len(items)
			}
		}
	}
	return 0
}
