package dbread

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type secretsManagerAPI interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

func resolveTargetDSN(ctx context.Context, target Target) (Target, error) {
	if strings.TrimSpace(target.DSN) != "" || strings.TrimSpace(target.DSNSecretARN) == "" {
		return target, nil
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return target, err
	}
	return resolveTargetDSNWithClient(ctx, target, secretsmanager.NewFromConfig(cfg))
}

func resolveTargetDSNWithClient(ctx context.Context, target Target, client secretsManagerAPI) (Target, error) {
	if strings.TrimSpace(target.DSN) != "" || strings.TrimSpace(target.DSNSecretARN) == "" {
		return target, nil
	}
	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: aws.String(strings.TrimSpace(target.DSNSecretARN))})
	if err != nil {
		return target, err
	}
	secret := strings.TrimSpace(aws.ToString(out.SecretString))
	if secret == "" {
		return target, fmt.Errorf("DSN secret is empty")
	}
	dsn, err := dsnFromSecretString(secret)
	if err != nil {
		return target, err
	}
	target.DSN = dsn
	if strings.TrimSpace(target.DSN) == "" {
		return target, fmt.Errorf("DSN secret does not contain a DSN")
	}
	return target, nil
}

func dsnFromSecretString(secret string) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", nil
	}
	if strings.HasPrefix(secret, "{") {
		var document map[string]any
		if err := json.Unmarshal([]byte(secret), &document); err != nil {
			return "", fmt.Errorf("DSN secret JSON is invalid: %w", err)
		}
		for _, key := range []string{"dsn", "url", "connection_string", "DATABASE_URL"} {
			raw, ok := document[key]
			if !ok {
				continue
			}
			value, ok := raw.(string)
			if !ok {
				return "", fmt.Errorf("DSN secret field %q must be a string", key)
			}
			if value = strings.TrimSpace(value); value != "" {
				return value, nil
			}
		}
		return "", fmt.Errorf("DSN secret JSON does not contain dsn, url, connection_string, or DATABASE_URL")
	}
	return secret, nil
}
