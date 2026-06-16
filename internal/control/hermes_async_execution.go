package control

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/clients"
	"github.com/piplabs/rsi-agent-platform/internal/config"
	storepkg "github.com/piplabs/rsi-agent-platform/internal/store"
	"github.com/piplabs/rsi-agent-platform/internal/transition"
)

type hermesExecutionRecovery struct {
	response     clients.RunnerResponse
	stillRunning bool
	status       string
	runnerStatus clients.HermesExecutionStatus
}

func recoverHermesExecution(client *clients.RunnerClient, executionID string) (hermesExecutionRecovery, error) {
	status, err := client.HermesExecutionStatus(executionID)
	if err != nil {
		if strings.Contains(err.Error(), "returned 404") {
			return hermesExecutionRecovery{
				response: hermesExecutorRecoveryFailure(
					executionID,
					workflowFailureRunnerExecutorStatusUnavailable,
					"Hermes executor status was unavailable for a previously started execution; refusing to launch a duplicate run.",
					"",
				),
			}, nil
		}
		return hermesExecutionRecovery{}, fmt.Errorf("check Hermes execution status %s: %w", executionID, err)
	}
	return recoverHermesExecutionFromStatus(executionID, status), nil
}

func recoverHermesExecutionFromStatus(executionID string, status clients.HermesExecutionStatus) hermesExecutionRecovery {
	statusText := strings.ToLower(strings.TrimSpace(status.Status))
	if status.Result != nil {
		return hermesExecutionRecovery{response: *status.Result, status: statusText, runnerStatus: status}
	}
	switch statusText {
	case "running", "accepted", "starting", "finalizing", "cancel_requested", "cancelling":
		return hermesExecutionRecovery{stillRunning: true, status: statusText, runnerStatus: status}
	case "queued":
		return hermesExecutionRecovery{stillRunning: true, status: statusText, runnerStatus: status}
	case "completed", "failed", "cancelled", "orphaned":
		return hermesExecutionRecovery{
			response: hermesExecutorRecoveryFailureWithStatus(
				executionID,
				workflowFailureRunnerExecutorResultUnavailable,
				"Hermes executor reached a terminal state but did not expose a durable result; refusing to launch a duplicate run.",
				status,
			),
			status:       statusText,
			runnerStatus: status,
		}
	default:
		statusText := strings.TrimSpace(status.Status)
		return hermesExecutionRecovery{
			response: hermesExecutorRecoveryFailure(
				executionID,
				workflowFailureRunnerExecutorStatusUnrecognized,
				fmt.Sprintf("Hermes executor returned unrecognized status %q for a previously started execution; refusing to launch a duplicate run.", statusText),
				statusText,
			),
			status:       strings.ToLower(statusText),
			runnerStatus: status,
		}
	}
}

func recoverHermesExecutionResult(client *clients.RunnerClient, executionID string) (clients.RunnerResponse, bool, error) {
	recovery, err := recoverHermesExecution(client, executionID)
	if err != nil {
		return clients.RunnerResponse{}, false, err
	}
	return recovery.response, recovery.stillRunning, nil
}

func cancelSupersededHermesExecutions(cfg config.Config, store storepkg.Store, client *clients.RunnerClient, caseID string, currentTraceID string) {
	caseID = strings.TrimSpace(caseID)
	currentTraceID = strings.TrimSpace(currentTraceID)
	if caseID == "" || currentTraceID == "" {
		return
	}
	runtime := newWorkflowRuntimeCoordinator(cfg, store)
	for _, item := range store.ListActiveRunnerExecutions() {
		itemTraceID := strings.TrimSpace(item.TraceID)
		if strings.TrimSpace(item.CaseID) != caseID || itemTraceID == "" || itemTraceID == currentTraceID {
			continue
		}
		itemStatus := strings.ToLower(strings.TrimSpace(item.Status))
		if itemStatus == "cancelling" {
			continue
		}
		itemClient := newHermesExecutorPool(cfg, firstNonEmpty(item.Role, "prod"), client).clientForRecord(item)
		if itemClient == nil {
			log.Printf("control-plane cancel superseded Hermes execution=%s trace=%s skipped: no executor endpoint recorded", item.ExecutionID, item.TraceID)
			continue
		}
		status, err := itemClient.CancelHermesExecution(item.ExecutionID)
		if err != nil {
			update := storepkg.RunnerExecution{
				ExecutionID:     item.ExecutionID,
				Status:          "cancel_requested",
				CancelRequested: true,
				FailureClass:    firstNonEmpty(strings.TrimSpace(item.FailureClass), workflowFailureRunnerExecutionCancelled),
				UpdatedAt:       time.Now().UTC(),
			}
			expectedHolder := item.Holder
			if expectedHolder == "" {
				expectedHolder = storepkg.HolderCASExpectEmpty()
			}
			if _, recordErr := runtime.recordRunnerExecutionWithHolderCAS(update, expectedHolder, item.HeartbeatAt); recordErr != nil {
				log.Printf("control-plane cancel superseded Hermes execution=%s trace=%s mark_cancel_requested_error=%v", item.ExecutionID, item.TraceID, recordErr)
			}
			log.Printf("control-plane cancel superseded Hermes execution=%s trace=%s error=%v", item.ExecutionID, item.TraceID, err)
			continue
		}
		now := time.Now().UTC()
		statusText := strings.ToLower(firstNonEmpty(status.Status, "cancelling"))
		if !storepkg.RunnerExecutionStatusTerminal(statusText) {
			statusText = "cancelling"
		}
		update := storepkg.RunnerExecution{
			ExecutionID:     item.ExecutionID,
			Status:          statusText,
			CancelRequested: true,
			HeartbeatAt:     &now,
			UpdatedAt:       now,
		}
		if status.Result != nil {
			completedAt := now
			update.Result = runnerResponseMap(*status.Result)
			update.CompletedAt = &completedAt
			update.Status = "cancelled"
			update.FailureClass = workflowFailureRunnerExecutionCancelled
		} else if storepkg.RunnerExecutionStatusTerminal(update.Status) {
			completedAt := now
			update.CompletedAt = &completedAt
			if !strings.EqualFold(update.Status, "cancelled") {
				supersessionFailureClass := firstNonEmpty(strings.TrimSpace(item.FailureClass), workflowFailureRunnerExecutionCancelled)
				failure := clients.RunnerResponse{
					OK:       false,
					Message:  "Hermes executor reached a terminal state without a durable result during trace supersession.",
					Provider: "hermes-executor",
					Raw: map[string]any{
						"failure_class": supersessionFailureClass,
						"runner_diagnostics": map[string]any{
							"execution_id":           strings.TrimSpace(item.ExecutionID),
							"executor_status":        strings.TrimSpace(update.Status),
							"provider_error_message": "Hermes executor reached a terminal state without a durable result during trace supersession.",
							"result_failure_class":   workflowFailureRunnerExecutorResultUnavailable,
							"recovery_decision":      "fail_closed_no_duplicate_execution",
							"supersession_reason":    "trace_superseded",
							"superseding_trace_id":   strings.TrimSpace(currentTraceID),
						},
					},
				}
				update.Status = "failed"
				update.Result = runnerResponseMap(failure)
				update.FailureClass = supersessionFailureClass
			}
		}
		expectedHolder := item.Holder
		if expectedHolder == "" {
			expectedHolder = storepkg.HolderCASExpectEmpty()
		}
		_, err = runtime.recordRunnerExecutionWithHolderCAS(update, expectedHolder, item.HeartbeatAt)
		if err != nil {
			if errors.Is(err, storepkg.ErrHolderCASMismatch) {
				refreshed, exists := store.GetRunnerExecution(item.ExecutionID)
				if exists && !storepkg.RunnerExecutionStatusTerminal(refreshed.Status) {
					expectedHolder := refreshed.Holder
					if expectedHolder == "" {
						expectedHolder = storepkg.HolderCASExpectEmpty()
					}
					_, err = runtime.recordRunnerExecutionWithHolderCAS(update, expectedHolder, refreshed.HeartbeatAt)
					if err != nil {
						log.Printf("control-plane cancel superseded Hermes execution=%s trace=%s CAS retry failed: %v", item.ExecutionID, item.TraceID, err)
					}
				}
			} else {
				log.Printf("control-plane cancel superseded Hermes execution=%s trace=%s CAS failed: %v", item.ExecutionID, item.TraceID, err)
			}
		}
	}
}

func executeOrPollAsyncHermesExecution(cfg config.Config, store storepkg.Store, client *clients.RunnerClient, task clients.RunnerTask, effect transition.EffectExecution, role string, ctx workflowContext, startedAt time.Time) (clients.RunnerResponse, bool, error) {
	now := time.Now().UTC()
	runtime := newWorkflowRuntimeCoordinator(cfg, store)
	executorPool := newHermesExecutorPool(cfg, role, client)
	executionHolder := runnerExecutionHolder(task.ExecutionID)
	task = runnerTaskWithExecutionHolder(task, executionHolder)
	record, exists := store.GetRunnerExecution(task.ExecutionID)
	recordClient := executorPool.clientForRecord(record)
	if recordClient == nil {
		recordClient = client
	}
	if exists && record.CancelRequested && !storepkg.RunnerExecutionStatusTerminal(record.Status) {
		status, err := recordClient.CancelHermesExecution(task.ExecutionID)
		if err != nil {
			if cfg.HermesExecutionHeartbeatTimeout > 0 {
				failureNow := time.Now().UTC()
				referenceTime := runnerExecutionHeartbeatReferenceTime(record)
				if !referenceTime.IsZero() && failureNow.Sub(referenceTime) > cfg.HermesExecutionHeartbeatTimeout {
					failure := hermesExecutorRecoveryFailure(task.ExecutionID, workflowFailureRunnerExecutorStatusUnavailable, "Hermes executor heartbeat expired while cancelling async execution.", "heartbeat_expired")
					_, _ = runtime.recordRunnerExecution(storepkg.RunnerExecution{
						ExecutionID:  task.ExecutionID,
						Status:       "failed",
						Result:       runnerResponseMap(failure),
						FailureClass: workflowFailureRunnerExecutorStatusUnavailable,
						CompletedAt:  &failureNow,
						HeartbeatAt:  &failureNow,
						UpdatedAt:    failureNow,
					})
					return failure, false, nil
				}
			}
			cancelRetryAt := time.Now().UTC()
			_, _ = runtime.recordRunnerExecutionWithHolderCAS(storepkg.RunnerExecution{
				ExecutionID:     task.ExecutionID,
				Status:          "cancelling",
				CancelRequested: true,
				UpdatedAt:       cancelRetryAt,
			}, expectedRunnerExecutionHolder(record), record.HeartbeatAt)
			return clients.RunnerResponse{}, true, errHermesExecutionStillRunning
		}
		cancelCompletedAt := time.Now().UTC()
		if hermesStatusObservedByNonOwner(status) {
			_, _ = runtime.recordRunnerExecutionWithHolderCAS(storepkg.RunnerExecution{
				ExecutionID:     task.ExecutionID,
				Status:          "cancel_requested",
				CancelRequested: true,
				UpdatedAt:       cancelCompletedAt,
			}, expectedRunnerExecutionHolder(record), record.HeartbeatAt)
			return clients.RunnerResponse{}, true, nil
		}
		statusText := strings.ToLower(firstNonEmpty(status.Status, "cancelling"))
		if !storepkg.RunnerExecutionStatusTerminal(statusText) {
			statusText = "cancelling"
		}
		update := storepkg.RunnerExecution{
			ExecutionID:     task.ExecutionID,
			Status:          statusText,
			CancelRequested: true,
			HeartbeatAt:     &cancelCompletedAt,
			UpdatedAt:       cancelCompletedAt,
		}
		if status.Result != nil {
			completedAt := cancelCompletedAt
			update.Result = runnerResponseMap(*status.Result)
			update.CompletedAt = &completedAt
			update.Status = "cancelled"
			update.FailureClass = workflowFailureRunnerExecutionCancelled
			_, _ = runtime.recordRunnerExecution(update)
			return clients.RunnerResponse{}, false, &workflowFailureError{
				failure: workflowFailure{
					Class:   workflowFailureRunnerExecutionCancelled,
					Summary: "Execution completed after cancellation was requested and is not deliverable.",
				},
			}
		}
		if storepkg.RunnerExecutionStatusTerminal(update.Status) {
			completedAt := cancelCompletedAt
			update.CompletedAt = &completedAt
			if strings.EqualFold(update.Status, "cancelled") {
				_, _ = runtime.recordRunnerExecution(update)
				return clients.RunnerResponse{}, false, &workflowFailureError{
					failure: workflowFailure{
						Class:   workflowFailureRunnerExecutionCancelled,
						Summary: "Execution was cancelled as requested.",
					},
				}
			}
			failure := hermesExecutorRecoveryFailure(task.ExecutionID, workflowFailureRunnerExecutorResultUnavailable, "Hermes executor reached a terminal state without a durable result.", update.Status)
			update.Status = "failed"
			update.Result = runnerResponseMap(failure)
			update.FailureClass = workflowFailureRunnerExecutorResultUnavailable
			updated, err := runtime.recordRunnerExecution(update)
			if err != nil {
				return clients.RunnerResponse{}, false, err
			}
			if runnerExecutionResultNonDeliverable(updated) {
				return clients.RunnerResponse{}, false, runnerExecutionCancelledError("Execution was cancelled as requested.")
			}
			return failure, false, nil
		}
		update.Status = "cancelling"
		_, _ = runtime.recordRunnerExecution(update)
		return clients.RunnerResponse{}, true, nil
	}
	if exists && storepkg.RunnerExecutionStatusTerminal(record.Status) {
		if record.CancelRequested || strings.EqualFold(record.Status, "cancelled") {
			return clients.RunnerResponse{}, false, &workflowFailureError{
				failure: workflowFailure{
					Class:   workflowFailureRunnerExecutionCancelled,
					Summary: "Execution was cancelled or superseded as requested.",
				},
			}
		}
		if resp, ok := runnerResponseFromMap(record.Result); ok {
			return resp, false, nil
		}
		failure := hermesExecutorRecoveryFailure(task.ExecutionID, workflowFailureRunnerExecutorResultUnavailable, "Hermes executor reached a terminal state without a durable result.", record.Status)
		return failure, false, nil
	}
	if !exists || strings.EqualFold(record.Status, "queued") {
		startFailureReferenceTime := time.Time{}
		if exists {
			startFailureReferenceTime = runnerExecutionHeartbeatReferenceTime(record)
		}
		if !exists {
			recorded, err := runtime.recordRunnerExecution(storepkg.RunnerExecution{
				ExecutionID:    task.ExecutionID,
				OperationID:    effect.ID,
				WorkflowID:     ctx.workflow.ID,
				TraceID:        ctx.trace.Summary.TraceID,
				ConversationID: ctx.workflow.ConversationID,
				CaseID:         ctx.workflow.CaseID,
				Role:           role,
				Status:         "queued",
				Holder:         executionHolder,
				Task:           runnerTaskMap(task),
				HeartbeatAt:    &now,
				CreatedAt:      now,
				UpdatedAt:      now,
			})
			if err != nil {
				return clients.RunnerResponse{}, false, err
			}
			record = recorded
			exists = true
			if startFailureReferenceTime.IsZero() {
				startFailureReferenceTime = runnerExecutionHeartbeatReferenceTime(record)
			}
		} else if strings.TrimSpace(record.ExecutorBaseURL) != "" || len(cfg.HermesExecutorPoolURLs) == 0 {
			recordClient = executorPool.clientForRecord(record)
			if recordClient == nil {
				recordClient = client
			}
			recovery, err := recoverHermesExecution(recordClient, task.ExecutionID)
			if err != nil {
				attemptedBaseURL := ""
				if recordClient != nil {
					attemptedBaseURL = recordClient.BaseURL()
				}
				if fallbackStatus, fallbackErr := executorPool.statusFromAnyEndpoint(task.ExecutionID, record, attemptedBaseURL); fallbackErr == nil {
					recovery = recoverHermesExecutionFromStatus(task.ExecutionID, fallbackStatus)
					err = nil
				}
			}
			if err != nil {
				failureNow := time.Now().UTC()
				referenceTime := startFailureReferenceTime
				if referenceTime.IsZero() {
					referenceTime = runnerExecutionHeartbeatReferenceTime(record)
				}
				shouldFail := cfg.HermesExecutionHeartbeatTimeout > 0 && !referenceTime.IsZero() && failureNow.Sub(referenceTime) > cfg.HermesExecutionHeartbeatTimeout
				if shouldFail {
					failure := workflowFailureFromRunnerError(err)
					failureResp := clients.RunnerResponse{
						OK:      false,
						Message: failure.Summary,
						Raw:     map[string]any{"failure_class": failure.Class},
					}
					completedAt := failureNow
					_, _ = runtime.recordRunnerExecution(storepkg.RunnerExecution{
						ExecutionID:  task.ExecutionID,
						Status:       "failed",
						Result:       runnerResponseMap(failureResp),
						FailureClass: failure.Class,
						HeartbeatAt:  &failureNow,
						CompletedAt:  &completedAt,
						UpdatedAt:    failureNow,
					})
					return clients.RunnerResponse{}, false, &workflowFailureError{failure: failure}
				}
				return clients.RunnerResponse{}, true, errHermesExecutionStillRunning
			}
			recoveredAt := time.Now().UTC()
			if err == nil && recovery.stillRunning {
				enteringFinalizing := runnerExecutionEnteringFinalizing(recovery.status, record)
				if hermesStatusObservedByNonOwner(recovery.runnerStatus) && runnerExecutionHeartbeatExpired(cfg, record, recoveredAt) {
					failure := hermesExecutorRecoveryFailureWithStatus(task.ExecutionID, workflowFailureRunnerExecutorStatusUnavailable, "Hermes executor heartbeat expired while polling a non-owner executor status.", recovery.runnerStatus)
					completedAt := recoveredAt
					_, _ = runtime.recordRunnerExecution(storepkg.RunnerExecution{
						ExecutionID:  task.ExecutionID,
						Status:       "failed",
						Result:       runnerResponseMap(failure),
						FailureClass: workflowFailureRunnerExecutorStatusUnavailable,
						HeartbeatAt:  &recoveredAt,
						CompletedAt:  &completedAt,
						UpdatedAt:    recoveredAt,
					})
					return failure, false, nil
				}
				if (recovery.status == "queued" || recovery.status == "finalizing") && !enteringFinalizing && runnerExecutionHeartbeatExpired(cfg, record, recoveredAt) {
					failureClass, message := heartbeatExpiredFailureClassAndMessage(recovery.status)
					failure := hermesExecutorRecoveryFailure(task.ExecutionID, failureClass, message, "heartbeat_expired")
					completedAt := recoveredAt
					_, _ = runtime.recordRunnerExecution(storepkg.RunnerExecution{
						ExecutionID:  task.ExecutionID,
						Status:       "failed",
						Result:       runnerResponseMap(failure),
						FailureClass: failureClass,
						HeartbeatAt:  &recoveredAt,
						CompletedAt:  &completedAt,
						UpdatedAt:    recoveredAt,
					})
					return failure, false, nil
				}
				update := storepkg.RunnerExecution{
					ExecutionID: task.ExecutionID,
					Status:      firstNonEmpty(recovery.status, "running"),
					Holder:      executionHolder,
					UpdatedAt:   recoveredAt,
				}
				if runnerExecutionStatusRefreshesHeartbeatFromStatus(recovery.runnerStatus, record) {
					update.HeartbeatAt = &recoveredAt
				}
				_, _ = runtime.recordRunnerExecution(update)
				return clients.RunnerResponse{}, true, errHermesExecutionStillRunning
			}
			if err == nil && !recovery.stillRunning {
				completedAt := recoveredAt
				recordStatus := "completed"
				if !recovery.response.OK {
					recordStatus = "failed"
				}
				updated, err := runtime.recordRunnerExecution(storepkg.RunnerExecution{
					ExecutionID:  task.ExecutionID,
					Status:       recordStatus,
					Result:       runnerResponseMap(recovery.response),
					FailureClass: stringValue(recovery.response.Raw["failure_class"]),
					Holder:       executionHolder,
					HeartbeatAt:  &recoveredAt,
					CompletedAt:  &completedAt,
					UpdatedAt:    recoveredAt,
				})
				if err != nil {
					return clients.RunnerResponse{}, false, err
				}
				if runnerExecutionResultNonDeliverable(updated) {
					return clients.RunnerResponse{}, false, runnerExecutionCancelledError("Execution completed after cancellation was requested and is not deliverable.")
				}
				return recovery.response, false, nil
			}
		}
		status, endpoint, err := executorPool.startExecution(task)
		if err != nil {
			if errors.Is(err, errNoReadyHermesExecutorEndpoints) {
				waitAt := time.Now().UTC()
				_, _ = runtime.recordRunnerExecutionWithHolderCAS(storepkg.RunnerExecution{
					ExecutionID: task.ExecutionID,
					Status:      "queued",
					Holder:      executionHolder,
					HeartbeatAt: &waitAt,
					UpdatedAt:   waitAt,
				}, expectedRunnerExecutionHolder(record), record.HeartbeatAt)
				return clients.RunnerResponse{}, true, errHermesExecutionStillRunning
			}
			failureNow := time.Now().UTC()
			referenceTime := startFailureReferenceTime
			if referenceTime.IsZero() {
				referenceTime = runnerExecutionHeartbeatReferenceTime(record)
			}
			shouldFail := cfg.HermesExecutionHeartbeatTimeout > 0 && !referenceTime.IsZero() && failureNow.Sub(referenceTime) > cfg.HermesExecutionHeartbeatTimeout
			if shouldFail {
				failure := workflowFailureFromRunnerError(err)
				failureResp := clients.RunnerResponse{
					OK:      false,
					Message: failure.Summary,
					Raw:     map[string]any{"failure_class": failure.Class},
				}
				completedAt := failureNow
				_, _ = runtime.recordRunnerExecution(storepkg.RunnerExecution{
					ExecutionID:        task.ExecutionID,
					Status:             "failed",
					Result:             runnerResponseMap(failureResp),
					FailureClass:       failure.Class,
					ExecutorInstanceID: record.ExecutorInstanceID,
					ExecutorBaseURL:    record.ExecutorBaseURL,
					HeartbeatAt:        &failureNow,
					CompletedAt:        &completedAt,
					UpdatedAt:          failureNow,
				})
				return clients.RunnerResponse{}, false, &workflowFailureError{failure: failure}
			}
			return clients.RunnerResponse{}, true, errHermesExecutionStillRunning
		}
		startCompletedAt := time.Now().UTC()
		statusText := strings.ToLower(firstNonEmpty(status.Status, "accepted"))
		if status.Result != nil {
			completedAt := startCompletedAt
			recordStatus := "completed"
			if !status.Result.OK {
				recordStatus = "failed"
			}
			updated, err := runtime.recordRunnerExecutionWithHolderCAS(storepkg.RunnerExecution{
				ExecutionID:        task.ExecutionID,
				OperationID:        effect.ID,
				WorkflowID:         ctx.workflow.ID,
				TraceID:            ctx.trace.Summary.TraceID,
				ConversationID:     ctx.workflow.ConversationID,
				CaseID:             ctx.workflow.CaseID,
				Role:               role,
				ExecutorInstanceID: endpoint.instanceID,
				ExecutorBaseURL:    endpoint.baseURL,
				Status:             recordStatus,
				Holder:             executionHolder,
				Task:               runnerTaskMap(task),
				Result:             runnerResponseMap(*status.Result),
				FailureClass:       stringValue(status.Result.Raw["failure_class"]),
				HeartbeatAt:        &startCompletedAt,
				StartedAt:          &startedAt,
				CompletedAt:        &completedAt,
				CreatedAt:          now,
				UpdatedAt:          startCompletedAt,
			}, expectedRunnerExecutionHolder(record), record.HeartbeatAt)
			if err != nil {
				if errors.Is(err, storepkg.ErrHolderCASMismatch) {
					return clients.RunnerResponse{}, true, errHermesExecutionStillRunning
				}
				return clients.RunnerResponse{}, false, err
			}
			if runnerExecutionResultNonDeliverable(updated) {
				return clients.RunnerResponse{}, false, runnerExecutionCancelledError("Execution completed after cancellation was requested and is not deliverable.")
			}
			return *status.Result, false, nil
		}
		if _, err := runtime.recordRunnerExecutionWithHolderCAS(storepkg.RunnerExecution{
			ExecutionID:        task.ExecutionID,
			OperationID:        effect.ID,
			WorkflowID:         ctx.workflow.ID,
			TraceID:            ctx.trace.Summary.TraceID,
			ConversationID:     ctx.workflow.ConversationID,
			CaseID:             ctx.workflow.CaseID,
			Role:               role,
			ExecutorInstanceID: endpoint.instanceID,
			ExecutorBaseURL:    endpoint.baseURL,
			Status:             statusText,
			Holder:             executionHolder,
			Task:               runnerTaskMap(task),
			HeartbeatAt:        &startCompletedAt,
			StartedAt:          &startedAt,
			CreatedAt:          now,
			UpdatedAt:          startCompletedAt,
		}, expectedRunnerExecutionHolder(record), record.HeartbeatAt); err != nil {
			if errors.Is(err, storepkg.ErrHolderCASMismatch) {
				return clients.RunnerResponse{}, true, errHermesExecutionStillRunning
			}
			return clients.RunnerResponse{}, false, err
		}
		return clients.RunnerResponse{}, true, nil
	}
	recordClient = executorPool.clientForRecord(record)
	if recordClient == nil {
		recordClient = client
	}
	status, err := recordClient.HermesExecutionStatus(task.ExecutionID)
	if err != nil {
		attemptedBaseURL := ""
		if recordClient != nil {
			attemptedBaseURL = recordClient.BaseURL()
		}
		if fallbackStatus, fallbackErr := executorPool.statusFromAnyEndpoint(task.ExecutionID, record, attemptedBaseURL); fallbackErr == nil {
			status = fallbackStatus
			err = nil
		}
	}
	if err != nil {
		if strings.Contains(err.Error(), "returned 404") {
			failureNow := time.Now().UTC()
			failure := hermesExecutorRecoveryFailure(
				task.ExecutionID,
				workflowFailureRunnerExecutorStatusUnavailable,
				"Hermes executor status was unavailable for a previously started execution; refusing to launch a duplicate run.",
				"",
			)
			_, _ = runtime.recordRunnerExecution(storepkg.RunnerExecution{
				ExecutionID:  task.ExecutionID,
				Status:       "failed",
				Result:       runnerResponseMap(failure),
				FailureClass: workflowFailureRunnerExecutorStatusUnavailable,
				CompletedAt:  &failureNow,
				HeartbeatAt:  &failureNow,
				UpdatedAt:    failureNow,
			})
			return failure, false, nil
		}
		if cfg.HermesExecutionHeartbeatTimeout > 0 {
			failureNow := time.Now().UTC()
			referenceTime := runnerExecutionHeartbeatReferenceTime(record)
			if !referenceTime.IsZero() && failureNow.Sub(referenceTime) > cfg.HermesExecutionHeartbeatTimeout {
				failure := hermesExecutorRecoveryFailure(task.ExecutionID, workflowFailureRunnerExecutorStatusUnavailable, "Hermes executor heartbeat expired while polling async execution.", "heartbeat_expired")
				_, _ = runtime.recordRunnerExecution(storepkg.RunnerExecution{
					ExecutionID:  task.ExecutionID,
					Status:       "failed",
					Result:       runnerResponseMap(failure),
					FailureClass: workflowFailureRunnerExecutorStatusUnavailable,
					CompletedAt:  &failureNow,
					HeartbeatAt:  &failureNow,
					UpdatedAt:    failureNow,
				})
				return failure, false, nil
			}
		}
		return clients.RunnerResponse{}, true, errHermesExecutionStillRunning
	}
	statusCompletedAt := time.Now().UTC()
	statusText := strings.ToLower(firstNonEmpty(status.Status, "running"))
	if status.Result != nil {
		completedAt := statusCompletedAt
		recordStatus := "completed"
		if !status.Result.OK {
			recordStatus = "failed"
		}
		failureClass := stringValue(status.Result.Raw["failure_class"])
		if record.CancelRequested || runnerExecutionStatusCancellationPending(record.Status) {
			recordStatus = "cancelled"
			failureClass = workflowFailureRunnerExecutionCancelled
		}
		updated, err := runtime.recordRunnerExecution(storepkg.RunnerExecution{
			ExecutionID:  task.ExecutionID,
			Status:       recordStatus,
			Result:       runnerResponseMap(*status.Result),
			FailureClass: failureClass,
			HeartbeatAt:  &statusCompletedAt,
			CompletedAt:  &completedAt,
			UpdatedAt:    statusCompletedAt,
		})
		if err != nil {
			return clients.RunnerResponse{}, false, err
		}
		if record.CancelRequested || runnerExecutionStatusCancellationPending(record.Status) || runnerExecutionResultNonDeliverable(updated) {
			return clients.RunnerResponse{}, false, runnerExecutionCancelledError("Execution completed after cancellation was requested and is not deliverable.")
		}
		return *status.Result, false, nil
	}
	switch strings.ToLower(strings.TrimSpace(statusText)) {
	case "running", "accepted", "starting", "finalizing", "cancel_requested", "cancelling", "queued":
		enteringFinalizing := runnerExecutionEnteringFinalizing(statusText, record)
		if hermesStatusObservedByNonOwner(status) && runnerExecutionHeartbeatExpired(cfg, record, statusCompletedAt) {
			failure := hermesExecutorRecoveryFailureWithStatus(task.ExecutionID, workflowFailureRunnerExecutorStatusUnavailable, "Hermes executor heartbeat expired while polling a non-owner executor status.", status)
			_, _ = runtime.recordRunnerExecution(storepkg.RunnerExecution{
				ExecutionID:  task.ExecutionID,
				Status:       "failed",
				Result:       runnerResponseMap(failure),
				FailureClass: workflowFailureRunnerExecutorStatusUnavailable,
				HeartbeatAt:  &statusCompletedAt,
				CompletedAt:  &statusCompletedAt,
				UpdatedAt:    statusCompletedAt,
			})
			return failure, false, nil
		}
		if (statusText == "queued" || statusText == "finalizing") && !enteringFinalizing && runnerExecutionHeartbeatExpired(cfg, record, statusCompletedAt) {
			failureClass, message := heartbeatExpiredFailureClassAndMessage(statusText)
			failure := hermesExecutorRecoveryFailure(task.ExecutionID, failureClass, message, "heartbeat_expired")
			_, _ = runtime.recordRunnerExecution(storepkg.RunnerExecution{
				ExecutionID:  task.ExecutionID,
				Status:       "failed",
				Result:       runnerResponseMap(failure),
				FailureClass: failureClass,
				HeartbeatAt:  &statusCompletedAt,
				CompletedAt:  &statusCompletedAt,
				UpdatedAt:    statusCompletedAt,
			})
			return failure, false, nil
		}
		update := storepkg.RunnerExecution{
			ExecutionID: task.ExecutionID,
			Status:      statusText,
			UpdatedAt:   statusCompletedAt,
		}
		if statusText == "cancel_requested" || statusText == "cancelling" {
			update.CancelRequested = true
		}
		if runnerExecutionStatusRefreshesHeartbeatFromStatus(status, record) {
			update.HeartbeatAt = &statusCompletedAt
		}
		_, _ = runtime.recordRunnerExecution(update)
		return clients.RunnerResponse{}, true, nil
	case "completed", "failed", "cancelled", "orphaned":
		if strings.EqualFold(statusText, "cancelled") && record.CancelRequested {
			_, _ = runtime.recordRunnerExecution(storepkg.RunnerExecution{
				ExecutionID: task.ExecutionID,
				Status:      "cancelled",
				HeartbeatAt: &statusCompletedAt,
				CompletedAt: &statusCompletedAt,
				UpdatedAt:   statusCompletedAt,
			})
			return clients.RunnerResponse{}, false, &workflowFailureError{
				failure: workflowFailure{
					Class:   workflowFailureRunnerExecutionCancelled,
					Summary: "Execution was cancelled as requested.",
				},
			}
		}
		failure := hermesExecutorRecoveryFailureWithStatus(task.ExecutionID, workflowFailureRunnerExecutorResultUnavailable, "Hermes executor reached a terminal state without a durable result.", status)
		_, _ = runtime.recordRunnerExecution(storepkg.RunnerExecution{
			ExecutionID:  task.ExecutionID,
			Status:       "failed",
			Result:       runnerResponseMap(failure),
			FailureClass: workflowFailureRunnerExecutorResultUnavailable,
			HeartbeatAt:  &statusCompletedAt,
			CompletedAt:  &statusCompletedAt,
			UpdatedAt:    statusCompletedAt,
		})
		return failure, false, nil
	default:
		failure := hermesExecutorRecoveryFailure(task.ExecutionID, workflowFailureRunnerExecutorStatusUnrecognized, fmt.Sprintf("Hermes executor returned unrecognized async status %q.", statusText), statusText)
		_, _ = runtime.recordRunnerExecution(storepkg.RunnerExecution{
			ExecutionID:  task.ExecutionID,
			Status:       "failed",
			Result:       runnerResponseMap(failure),
			FailureClass: workflowFailureRunnerExecutorStatusUnrecognized,
			HeartbeatAt:  &statusCompletedAt,
			CompletedAt:  &statusCompletedAt,
			UpdatedAt:    statusCompletedAt,
		})
		return failure, false, nil
	}
}

func runnerTaskMap(task clients.RunnerTask) map[string]any {
	raw, _ := json.Marshal(task)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return out
}

func runnerExecutionHolder(executionID string) string {
	executionID = strings.TrimSpace(executionID)
	if executionID == "" {
		return "hermes-executor"
	}
	return "hermes-executor:" + executionID
}

func expectedRunnerExecutionHolder(record storepkg.RunnerExecution) string {
	holder := strings.TrimSpace(record.Holder)
	if holder == "" {
		return storepkg.HolderCASExpectEmpty()
	}
	return holder
}

func runnerExecutionHeartbeatExpired(cfg config.Config, record storepkg.RunnerExecution, now time.Time) bool {
	if cfg.HermesExecutionHeartbeatTimeout <= 0 {
		return false
	}
	referenceTime := runnerExecutionHeartbeatReferenceTime(record)
	return !referenceTime.IsZero() && now.Sub(referenceTime) > cfg.HermesExecutionHeartbeatTimeout
}

func runnerExecutionStatusRefreshesHeartbeat(status string, record storepkg.RunnerExecution) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "queued", "cancel_requested":
		return false
	case "finalizing":
		return runnerExecutionEnteringFinalizing(status, record) || record.HeartbeatAt == nil
	default:
		return true
	}
}

func runnerExecutionStatusRefreshesHeartbeatFromStatus(status clients.HermesExecutionStatus, record storepkg.RunnerExecution) bool {
	if hermesStatusObservedByNonOwner(status) {
		return false
	}
	return runnerExecutionStatusRefreshesHeartbeat(status.Status, record)
}

func hermesStatusObservedByNonOwner(status clients.HermesExecutionStatus) bool {
	if status.ExecutorOwnerMismatch {
		return true
	}
	owner := strings.TrimSpace(status.ExecutorInstanceID)
	current := strings.TrimSpace(status.CurrentExecutorInstanceID)
	return owner != "" && current != "" && owner != current
}

func runnerExecutionEnteringFinalizing(status string, record storepkg.RunnerExecution) bool {
	return strings.EqualFold(strings.TrimSpace(status), "finalizing") &&
		!strings.EqualFold(strings.TrimSpace(record.Status), "finalizing")
}

func runnerExecutionResultNonDeliverable(record storepkg.RunnerExecution) bool {
	return record.CancelRequested ||
		runnerExecutionStatusCancellationPending(record.Status) ||
		strings.EqualFold(strings.TrimSpace(record.Status), "cancelled") ||
		strings.EqualFold(strings.TrimSpace(record.FailureClass), workflowFailureRunnerExecutionCancelled)
}

func runnerExecutionCancelledError(summary string) *workflowFailureError {
	return &workflowFailureError{
		failure: workflowFailure{
			Class:   workflowFailureRunnerExecutionCancelled,
			Summary: strings.TrimSpace(summary),
		},
	}
}

func runnerTaskWithExecutionHolder(task clients.RunnerTask, holder string) clients.RunnerTask {
	holder = strings.TrimSpace(holder)
	if holder == "" {
		return task
	}
	intent := map[string]any{}
	for key, value := range task.ExecutionIntent {
		intent[key] = value
	}
	intent["runner_execution_holder"] = holder
	task.ExecutionIntent = intent
	return task
}

func runnerResponseMap(resp clients.RunnerResponse) map[string]any {
	raw, _ := json.Marshal(resp)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return out
}

func runnerResponseFromMap(value map[string]any) (clients.RunnerResponse, bool) {
	if len(value) == 0 {
		return clients.RunnerResponse{}, false
	}
	_, hasOK := value["ok"]
	_, hasMessage := value["message"]
	_, hasProvider := value["provider"]
	if !hasOK && !hasMessage && !hasProvider {
		return clients.RunnerResponse{}, false
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return clients.RunnerResponse{}, false
	}
	var out clients.RunnerResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return clients.RunnerResponse{}, false
	}
	if out.Raw == nil {
		out.Raw = map[string]any{}
	}
	return out, true
}

func hermesExecutorRecoveryFailure(executionID string, failureClass string, message string, status string) clients.RunnerResponse {
	return clients.RunnerResponse{
		OK:       false,
		Message:  message,
		Provider: "hermes-executor",
		Raw: map[string]any{
			"failure_class": failureClass,
			"runner_diagnostics": map[string]any{
				"execution_id":           strings.TrimSpace(executionID),
				"executor_status":        strings.TrimSpace(status),
				"provider_error_message": strings.TrimSpace(message),
				"recovery_decision":      "fail_closed_no_duplicate_execution",
			},
		},
	}
}

func hermesExecutorRecoveryFailureWithStatus(executionID string, failureClass string, message string, status clients.HermesExecutionStatus) clients.RunnerResponse {
	resp := hermesExecutorRecoveryFailure(executionID, failureClass, message, firstNonEmpty(status.Status, status.LastObservedStatus))
	diagnostics, _ := resp.Raw["runner_diagnostics"].(map[string]any)
	if diagnostics == nil {
		diagnostics = map[string]any{}
		resp.Raw["runner_diagnostics"] = diagnostics
	}
	addStringDiagnostic := func(key string, value string) {
		if strings.TrimSpace(value) != "" {
			diagnostics[key] = strings.TrimSpace(value)
		}
	}
	addFloatDiagnostic := func(key string, value float64) {
		if value != 0 {
			diagnostics[key] = value
		}
	}
	addIntDiagnostic := func(key string, value int64) {
		if value != 0 {
			diagnostics[key] = value
		}
	}
	addStringDiagnostic("executor_instance_id", status.ExecutorInstanceID)
	addStringDiagnostic("current_executor_instance_id", status.CurrentExecutorInstanceID)
	if hermesStatusObservedByNonOwner(status) {
		diagnostics["executor_owner_mismatch"] = true
	}
	addFloatDiagnostic("executor_started_at_unix", status.ExecutorStartedAtUnix)
	addFloatDiagnostic("current_executor_started_at_unix", status.CurrentExecutorStartedAtUnix)
	addStringDiagnostic("executor_message", status.Message)
	addStringDiagnostic("operation_id", status.OperationID)
	addStringDiagnostic("trace_id", status.TraceID)
	addStringDiagnostic("workflow_id", status.WorkflowID)
	addStringDiagnostic("phase", status.Phase)
	addStringDiagnostic("workspace_root", status.WorkspaceRoot)
	addStringDiagnostic("session_id", status.SessionID)
	addStringDiagnostic("last_observed_status", status.LastObservedStatus)
	addStringDiagnostic("last_observed_ledger_seq", status.LastObservedLedgerSeq)
	addStringDiagnostic("last_observed_event_type", status.LastObservedEventType)
	addStringDiagnostic("last_observed_event_status", status.LastObservedEventStatus)
	addStringDiagnostic("last_observed_phase", status.LastObservedPhase)
	addStringDiagnostic("last_observed_recorded_at", status.LastObservedRecordedAt)
	addFloatDiagnostic("last_observed_at_unix", status.LastObservedAtUnix)
	addStringDiagnostic("last_observed_invocation_id", status.LastObservedInvocationID)
	addStringDiagnostic("status_file_path", status.StatusFilePath)
	addFloatDiagnostic("status_file_mtime_unix", status.StatusFileMtimeUnix)
	addIntDiagnostic("status_file_size_bytes", status.StatusFileSizeBytes)
	return resp
}

func heartbeatExpiredFailureClassAndMessage(status string) (failureClass string, message string) {
	if status == "finalizing" {
		return "plugin_execution_envelope_missing", "Hermes executor heartbeat expired while finalizing the native execution envelope."
	}
	return workflowFailureRunnerExecutorStatusUnavailable, "Hermes executor remained queued past the heartbeat timeout; refusing to defer indefinitely."
}
