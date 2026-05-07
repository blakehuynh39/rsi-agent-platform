package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/piplabs/rsi-agent-platform/internal/dbread"
)

func main() {
	registry, err := dbread.LoadRegistry("")
	if err != nil {
		log.Fatalf("load db read registry: %v", err)
	}
	lambda.Start(func(ctx context.Context, job dbread.LambdaJob) (dbread.LambdaResult, error) {
		return dbread.HandleLambdaJob(ctx, registry, job)
	})
}
