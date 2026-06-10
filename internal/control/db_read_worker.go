package control

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"github.com/piplabs/rsi-agent-platform/internal/dbread"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
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
	lambdaInvoker, err := dbread.NewLambdaInvoker(context.Background())
	if err != nil {
		log.Printf("db-read-worker lambda invoker unavailable error=%v", err)
	}
	holder := "db-read-worker-" + uuid.NewString()
	slackAPI := dbReadSlackAPI(cfg)
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
			markDBReadExternalToolOutcome(cfg, store, request, storepkg.ExternalToolOutcomeExpired, "DB read request expired")
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
				handleDBReadValidationLease(context.Background(), cfg, store, registry, slackAPI, lambdaInvoker, validationLease)
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
			handleDBReadLease(context.Background(), cfg, store, registry, slackAPI, lambdaInvoker, lease)
		} else {
			lease, ok, err := store.ClaimNextDBReadRequest(holder, cfg.WorkItemLeaseDuration, now, cfg.DBReadWorkerTargets)
			if err != nil {
				log.Printf("db-read-worker claim error=%v", err)
				time.Sleep(cfg.WorkerPollInterval)
				continue
			}
			if ok {
				handleDBReadLease(context.Background(), cfg, store, registry, slackAPI, lambdaInvoker, lease)
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
			handleDBReadValidationLease(context.Background(), cfg, store, registry, slackAPI, lambdaInvoker, validationLease)
		}
	}
}

func dbReadSlackAPI(cfg config.Config) slackMessagePoster {
	if cfg.SlackBotToken == "" {
		return nil
	}
	return slack.New(cfg.SlackBotToken)
}

func handleDBReadValidationLease(ctx context.Context, cfg config.Config, store storepkg.Store, registry dbread.Registry, slackAPI slackMessagePoster, lambdaInvoker *dbread.LambdaInvoker, lease storepkg.DBReadLease) {
	request := lease.Request
	target, ok := registry.Target(request.Target)
	now := time.Now().UTC()
	if !ok {
		if _, err := appendDBReadValidation(store, request, storepkg.DBReadValidationStatusFailed, "target_lookup", "target is not configured", "unknown_target", now); err != nil {
			log.Printf("db-read-worker validation target lookup append request=%s error=%v", request.ID, err)
		}
		if updated, found := store.GetDBReadRequest(request.ID); found {
			markDBReadExternalToolOutcome(cfg, store, updated, storepkg.ExternalToolOutcomeFailed, "target is not configured")
		}
		return
	}
	if request.ExpiresAt.Before(now) {
		expired, err := store.TransitionDBReadRequest(request.ID, storepkg.DBReadStateValidating, storepkg.DBReadStateExpired, func(item *storepkg.DBReadRequest) error {
			item.ErrorMessage = "validation expired before target-side prepare"
			return nil
		})
		if err != nil {
			log.Printf("db-read-worker validation expiry transition request=%s error=%v", request.ID, err)
			if latest, found := store.GetDBReadRequest(request.ID); found {
				markDBReadExternalToolOutcome(cfg, store, latest, storepkg.ExternalToolOutcomeFailed, "validation expiry transition failed: "+err.Error())
			} else {
				markDBReadExternalToolOutcome(cfg, store, request, storepkg.ExternalToolOutcomeFailed, "validation expiry transition failed: "+err.Error())
			}
			return
		}
		markDBReadExternalToolOutcome(cfg, store, expired, storepkg.ExternalToolOutcomeExpired, "validation expired before target-side prepare")
		return
	}
	var validation dbread.SQLValidation
	if target.ExecutionBoundary() == "aws_lambda" {
		if lambdaInvoker == nil {
			validation = dbread.SQLValidation{SQLSHA256: request.SQLSHA256, ErrorCode: "lambda_unavailable", Message: "Lambda invoker is not configured"}
		} else {
			lambdaValidation, err := lambdaInvoker.Validate(ctx, target, lease)
			validation = lambdaValidation
			if err != nil {
				log.Printf("db-read-worker lambda validation request=%s target=%s error=%v", request.ID, request.Target, err)
			}
		}
	} else {
		validation = dbread.ValidateAgainstTarget(ctx, target, request.SQL)
	}
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
	if !validation.OK {
		markDBReadExternalToolOutcome(cfg, store, updated, storepkg.ExternalToolOutcomeFailed, firstNonEmpty(validation.Message, validation.ErrorCode, "target-side validation failed"))
		return
	}
	if cfg.DBReadAutoApprove {
		if err := autoApproveDBReadRequest(ctx, store, slackAPI, updated, attempt, validation.Preview); err != nil {
			log.Printf("db-read-worker auto-approve request=%s error=%v", request.ID, err)
		}
		return
	}
	if updated.ChannelID != "" {
		if err := postDBReadApprovalCard(ctx, cfg, store, slackAPI, updated, attempt, validation.Preview); err != nil {
			log.Printf("db-read-worker approval card request=%s error=%v", request.ID, err)
		}
	}
}

// autoApproveDBReadRequest moves a validated request straight to approved
// without waiting for a Slack approver. SQL safety is still guaranteed by the
// read-only AST validation and target-side prepare that ran before this point.
// A buttonless audit card is posted to Slack so the result still lands in the
// originating thread.
func autoApproveDBReadRequest(ctx context.Context, store storepkg.Store, slackAPI slackMessagePoster, request storepkg.DBReadRequest, attempt storepkg.DBReadValidationAttempt, preview string) error {
	if request.ChannelID != "" && slackAPI != nil {
		if err := postDBReadAuditCard(ctx, store, slackAPI, request, attempt, preview); err != nil {
			log.Printf("db-read-worker audit card request=%s error=%v", request.ID, err)
		}
		if latest, ok := store.GetDBReadRequest(request.ID); ok {
			request = latest
		}
	}
	now := time.Now().UTC()
	updated, err := store.TransitionDBReadRequest(request.ID, storepkg.DBReadStatePendingApproval, storepkg.DBReadStateApproved, func(item *storepkg.DBReadRequest) error {
		item.ApprovedAt = &now
		return nil
	})
	if err != nil {
		return err
	}
	if pause, ok := store.GetExternalToolPauseByDBReadRequestID(updated.ID); ok {
		_, _ = store.UpdateExternalToolPause(pause.ID, func(item *storepkg.ExternalToolPause) error {
			item.ApprovalStatus = storepkg.ExternalToolApprovalApproved
			item.ApprovalRef = "auto:read_only_validated"
			return nil
		})
	}
	return updateDBReadSlackCard(ctx, slackAPI, updated, "auto-approved (validated read-only); queued for execution")
}

func appendDBReadValidation(store storepkg.Store, request storepkg.DBReadRequest, status storepkg.DBReadValidationStatus, stage string, message string, code string, now time.Time) (storepkg.DBReadValidationAttempt, error) {
	attempt := storepkg.NewDBReadValidationAttempt(request, status, stage, message, map[string]any{"error_code": code}, now)
	return store.AppendDBReadValidationAttempt(attempt)
}

func handleDBReadLease(ctx context.Context, cfg config.Config, store storepkg.Store, registry dbread.Registry, slackAPI slackMessagePoster, lambdaInvoker *dbread.LambdaInvoker, lease storepkg.DBReadLease) {
	request := lease.Request
	target, ok := registry.Target(request.Target)
	if !ok {
		recordDBReadExecutionFailure(store, request, lease.Token, "unknown_target", "target is not configured")
		if updated, found := store.GetDBReadRequest(request.ID); found {
			markDBReadExternalToolOutcome(cfg, store, updated, storepkg.ExternalToolOutcomeFailed, "target is not configured")
		}
		return
	}
	now := time.Now().UTC()
	if request.ExpiresAt.Before(now) {
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
			if latest, found := store.GetDBReadRequest(lease.Request.ID); found {
				markDBReadExternalToolOutcome(cfg, store, latest, storepkg.ExternalToolOutcomeFailed, "expiry transition failed: "+expireErr.Error())
			} else {
				markDBReadExternalToolOutcome(cfg, store, lease.Request, storepkg.ExternalToolOutcomeFailed, "expiry transition failed: "+expireErr.Error())
			}
			return
		}
		if slackErr := updateDBReadSlackCard(ctx, slackAPI, expired, "expired"); slackErr != nil {
			log.Printf("db-read-worker expire slack update request=%s error=%v", lease.Request.ID, slackErr)
		}
		markDBReadExternalToolOutcome(cfg, store, expired, storepkg.ExternalToolOutcomeExpired, errDBReadRequestExpired.Error())
		return
	}
	ready, reason, permanent := dbReadExternalPauseReadyForExecution(store, request)
	if !ready {
		if permanent {
			log.Printf("db-read-worker permanent pause readiness failure request=%s reason=%s", request.ID, reason)
			recordDBReadExecutionFailure(store, request, lease.Token, "pause_readiness_failed", reason)
			if updated, found := store.GetDBReadRequest(request.ID); found {
				markDBReadExternalToolOutcome(cfg, store, updated, storepkg.ExternalToolOutcomeFailed, reason)
			}
			return
		}
		log.Printf("db-read-worker waiting for external tool pause before execution request=%s reason=%s", request.ID, reason)
		return
	}
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
				if latest, found := store.GetDBReadRequest(lease.Request.ID); found {
					markDBReadExternalToolOutcome(cfg, store, latest, storepkg.ExternalToolOutcomeFailed, "expiry transition failed: "+expireErr.Error())
				} else {
					markDBReadExternalToolOutcome(cfg, store, lease.Request, storepkg.ExternalToolOutcomeFailed, "expiry transition failed: "+expireErr.Error())
				}
				return
			}
			if slackErr := updateDBReadSlackCard(ctx, slackAPI, expired, "expired"); slackErr != nil {
				log.Printf("db-read-worker expire slack update request=%s error=%v", lease.Request.ID, slackErr)
			}
			markDBReadExternalToolOutcome(cfg, store, expired, storepkg.ExternalToolOutcomeExpired, errDBReadRequestExpired.Error())
			return
		}
		recordDBReadExecutionFailure(store, lease.Request, lease.Token, "transition_failed", err.Error())
		if updated, found := store.GetDBReadRequest(lease.Request.ID); found {
			markDBReadExternalToolOutcome(cfg, store, updated, storepkg.ExternalToolOutcomeFailed, err.Error())
		}
		return
	}
	if err := updateDBReadSlackCard(ctx, slackAPI, request, "executing"); err != nil {
		log.Printf("db-read-worker slack executing update request=%s error=%v", request.ID, err)
	}
	var result dbread.DBResult
	if target.ExecutionBoundary() == "aws_lambda" {
		if lambdaInvoker == nil {
			err = fmt.Errorf("Lambda invoker is not configured")
		} else {
			result, err = lambdaInvoker.Execute(ctx, target, storepkg.DBReadLease{Request: request, Token: lease.Token})
		}
	} else {
		result, err = dbread.ExecuteRead(ctx, target, request)
	}
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
		markDBReadExternalToolOutcome(cfg, store, updated, storepkg.ExternalToolOutcomeSucceeded, "")
	} else {
		_ = updateDBReadSlackCard(ctx, slackAPI, updated, "failed: "+truncateSlackText(errorMessage, 800))
		markDBReadExternalToolOutcome(cfg, store, updated, storepkg.ExternalToolOutcomeFailed, errorMessage)
	}
}

func dbReadExternalPauseReadyForExecution(store storepkg.Store, request storepkg.DBReadRequest) (ready bool, reason string, permanent bool) {
	pause, ok := store.GetExternalToolPauseByDBReadRequestID(request.ID)
	if !ok {
		return false, "missing external tool pause", false
	}
	if strings.TrimSpace(pause.TransportToolName) != "db_read_query" ||
		strings.TrimSpace(pause.ToolCallID) == "" ||
		strings.TrimSpace(pause.HermesSessionID) == "" {
		return false, "external tool pause identity is incomplete", true
	}
	if strings.TrimSpace(pause.DBReadRequestID) != request.ID || strings.TrimSpace(pause.SQLSHA256) != request.SQLSHA256 {
		return false, "external tool pause request/hash mismatch", true
	}
	if len(pause.PendingAssistantMessage) == 0 || len(pause.TranscriptSnapshot) == 0 {
		return false, "external tool pause transcript is not committed", false
	}
	workflow, ok := findWorkflow(store.ListWorkflows(), pause.WorkflowID)
	if !ok {
		return false, "workflow not found", true
	}
	if workflow.Status != string(transition.WorkflowStateWaitingExternalTool) {
		return false, "workflow is not waiting on external tool", false
	}
	return true, "", false
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
