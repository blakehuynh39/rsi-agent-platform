package control

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/dbread"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
)

var errDBReadRequestExpired = errors.New("db read request expired")

func RunDBReadWorker(cfg config.Config, store storepkg.Store) error {
	if !cfg.DBReadEnabled {
		return fmt.Errorf("db-read-worker requires RSI_DB_READ_ENABLED=true")
	}
	registry, err := dbread.LoadRegistry(cfg.DBReadTargetsJSON)
	if err != nil {
		return err
	}
	holder := "db-read-worker-" + uuid.NewString()
	var slackAPI slackMessagePoster
	if cfg.SlackBotToken != "" {
		slackAPI = slack.New(cfg.SlackBotToken)
	}
	log.Printf("starting db-read-worker holder=%s targets=%v", holder, cfg.DBReadWorkerTargets)
	prioritizeValidation := true
	for {
		now := time.Now().UTC()
		expired, err := store.ExpirePendingDBReadRequests(now)
		if err != nil {
			log.Printf("db-read-worker expire error=%v", err)
		}
		for _, request := range expired {
			if err := updateDBReadSlackCard(context.Background(), slackAPI, request, "expired"); err != nil {
				log.Printf("db-read-worker expire slack update request=%s error=%v", request.ID, err)
			}
		}
		prioritizeValidation = !prioritizeValidation
		if prioritizeValidation {
			validationLease, ok, err := store.ClaimNextDBReadValidationRequest(holder, cfg.WorkItemLeaseDuration, now, cfg.DBReadWorkerTargets)
			if err != nil {
				log.Printf("db-read-worker validation claim error=%v", err)
				time.Sleep(cfg.WorkerPollInterval)
				continue
			}
			if ok {
				handleDBReadValidationLease(context.Background(), cfg, store, registry, slackAPI, validationLease)
				continue
			}
			lease, ok, err := store.ClaimNextDBReadRequest(holder, cfg.WorkItemLeaseDuration, now, cfg.DBReadWorkerTargets)
			if err != nil {
				log.Printf("db-read-worker claim error=%v", err)
				time.Sleep(cfg.WorkerPollInterval)
				continue
			}
			if !ok {
				time.Sleep(cfg.WorkerPollInterval)
				continue
			}
			handleDBReadLease(context.Background(), store, registry, slackAPI, lease)
		} else {
			lease, ok, err := store.ClaimNextDBReadRequest(holder, cfg.WorkItemLeaseDuration, now, cfg.DBReadWorkerTargets)
			if err != nil {
				log.Printf("db-read-worker claim error=%v", err)
				time.Sleep(cfg.WorkerPollInterval)
				continue
			}
			if ok {
				handleDBReadLease(context.Background(), store, registry, slackAPI, lease)
				continue
			}
			validationLease, ok, err := store.ClaimNextDBReadValidationRequest(holder, cfg.WorkItemLeaseDuration, now, cfg.DBReadWorkerTargets)
			if err != nil {
				log.Printf("db-read-worker validation claim error=%v", err)
				time.Sleep(cfg.WorkerPollInterval)
				continue
			}
			if !ok {
				time.Sleep(cfg.WorkerPollInterval)
				continue
			}
			handleDBReadValidationLease(context.Background(), cfg, store, registry, slackAPI, validationLease)
		}
	}
}

func handleDBReadValidationLease(ctx context.Context, cfg config.Config, store storepkg.Store, registry dbread.Registry, slackAPI slackMessagePoster, lease storepkg.DBReadLease) {
	request := lease.Request
	target, ok := registry.Target(request.Target)
	now := time.Now().UTC()
	if !ok {
		if _, err := appendDBReadValidation(store, request, storepkg.DBReadValidationStatusFailed, "target_lookup", "target is not configured", "unknown_target", now); err != nil {
			log.Printf("db-read-worker validation target lookup append request=%s error=%v", request.ID, err)
		}
		return
	}
	if request.ExpiresAt.Before(now) {
		_, _ = store.TransitionDBReadRequest(request.ID, storepkg.DBReadStateValidating, storepkg.DBReadStateExpired, func(item *storepkg.DBReadRequest) error {
			item.ErrorMessage = "validation expired before target-side prepare"
			return nil
		})
		return
	}
	validation := dbread.ValidateAgainstTarget(ctx, target, request.SQL)
	status := storepkg.DBReadValidationStatusFailed
	if validation.OK {
		status = storepkg.DBReadValidationStatusSucceeded
	}
	attempt, err := appendDBReadValidation(store, request, status, "target_prepare", validation.Message, validation.ErrorCode, now)
	if err != nil {
		log.Printf("db-read-worker validation append request=%s error=%v", request.ID, err)
		return
	}
	updated, _ := store.GetDBReadRequest(request.ID)
	if validation.OK && updated.ChannelID != "" {
		if err := postDBReadApprovalCard(ctx, cfg, store, slackAPI, updated, attempt, validation.Preview); err != nil {
			log.Printf("db-read-worker approval card request=%s error=%v", request.ID, err)
		}
	}
}

func appendDBReadValidation(store storepkg.Store, request storepkg.DBReadRequest, status storepkg.DBReadValidationStatus, stage string, message string, code string, now time.Time) (storepkg.DBReadValidationAttempt, error) {
	attempt := storepkg.NewDBReadValidationAttempt(request, status, stage, message, map[string]any{"error_code": code}, now)
	return store.AppendDBReadValidationAttempt(attempt)
}

func handleDBReadLease(ctx context.Context, store storepkg.Store, registry dbread.Registry, slackAPI slackMessagePoster, lease storepkg.DBReadLease) {
	request := lease.Request
	target, ok := registry.Target(request.Target)
	if !ok {
		recordDBReadExecutionFailure(store, request, lease.Token, "unknown_target", "target is not configured")
		return
	}
	now := time.Now().UTC()
	request, err := store.TransitionDBReadRequest(request.ID, storepkg.DBReadStateApproved, storepkg.DBReadStateExecuting, func(item *storepkg.DBReadRequest) error {
		if item.LeaseToken != lease.Token {
			return fmt.Errorf("db read lease token mismatch")
		}
		if item.ExpiresAt.Before(now) {
			return errDBReadRequestExpired
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errDBReadRequestExpired) {
			expired, expireErr := store.TransitionDBReadRequest(lease.Request.ID, storepkg.DBReadStateApproved, storepkg.DBReadStateExpired, func(item *storepkg.DBReadRequest) error {
				if item.LeaseToken != lease.Token {
					return fmt.Errorf("db read lease token mismatch")
				}
				item.ErrorMessage = errDBReadRequestExpired.Error()
				item.LeaseHolder = ""
				item.LeaseToken = ""
				item.LeaseExpiresAt = nil
				return nil
			})
			if expireErr != nil {
				log.Printf("db-read-worker expire approved request=%s error=%v", lease.Request.ID, expireErr)
			} else if slackErr := updateDBReadSlackCard(ctx, slackAPI, expired, "expired"); slackErr != nil {
				log.Printf("db-read-worker expire slack update request=%s error=%v", lease.Request.ID, slackErr)
			}
			return
		}
		recordDBReadExecutionFailure(store, lease.Request, lease.Token, "transition_failed", err.Error())
		return
	}
	if err := updateDBReadSlackCard(ctx, slackAPI, request, "executing"); err != nil {
		log.Printf("db-read-worker slack executing update request=%s error=%v", request.ID, err)
	}
	result, err := dbread.ExecuteRead(ctx, target, request)
	status := storepkg.DBReadExecutionStatusSucceeded
	errorCode := ""
	errorMessage := ""
	if err != nil {
		status = storepkg.DBReadExecutionStatusFailed
		errorCode = "execution_failed"
		errorMessage = err.Error()
	}
	execResult := storepkg.NewDBReadExecutionResult(request, status, result.Rows, time.Now().UTC())
	execResult.LeaseToken = lease.Token
	execResult.RowCount = result.RowCount
	execResult.Truncated = result.Truncated
	execResult.ErrorCode = errorCode
	execResult.ErrorMessage = errorMessage
	if _, err := store.AppendDBReadExecutionResult(execResult); err != nil {
		log.Printf("db-read-worker append execution result request=%s error=%v", request.ID, err)
		return
	}
	updated, _ := store.GetDBReadRequest(request.ID)
	if status == storepkg.DBReadExecutionStatusSucceeded {
		_ = updateDBReadSlackCard(ctx, slackAPI, updated, fmt.Sprintf("succeeded; rows=%d truncated=%t", result.RowCount, result.Truncated))
	} else {
		_ = updateDBReadSlackCard(ctx, slackAPI, updated, "failed: "+truncateSlackText(errorMessage, 800))
	}
}

func recordDBReadExecutionFailure(store storepkg.Store, request storepkg.DBReadRequest, leaseToken string, code string, message string) {
	now := time.Now().UTC()
	result := storepkg.NewDBReadExecutionResult(request, storepkg.DBReadExecutionStatusFailed, nil, now)
	result.LeaseToken = leaseToken
	result.ErrorCode = code
	result.ErrorMessage = message
	if _, err := store.AppendDBReadExecutionResult(result); err != nil {
		log.Printf("db-read-worker failure append request=%s error=%v", request.ID, err)
	}
}
