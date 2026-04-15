package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/operation"
	"github.com/piplabs/rsi-agent-platform/internal/queue"
)

func payloadHash(payload map[string]interface{}) string {
	body := strings.TrimSpace(jsonString(payload))
	if body == "" || body == "{}" {
		return ""
	}
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}

func ensureOperationWorkItemTx(tx *sql.Tx, op operation.Execution, item queue.WorkItem) (operation.Execution, queue.WorkItem, bool, error) {
	now := time.Now().UTC()
	if op.CreatedAt.IsZero() {
		op.CreatedAt = now
	}
	if op.UpdatedAt.IsZero() || op.UpdatedAt.Before(op.CreatedAt) {
		op.UpdatedAt = op.CreatedAt
	}
	if op.Status == "" {
		op.Status = operation.StatusQueued
	}
	if op.Queue == "" {
		op.Queue = item.Queue
	}
	if op.PayloadHash == "" {
		op.PayloadHash = payloadHash(item.Payload)
	}
	op, err := normalizeOperationExecution(op)
	if err != nil {
		return operation.Execution{}, queue.WorkItem{}, false, err
	}
	createdOp, _, err := getOrCreateOperationTx(tx, op)
	if err != nil {
		return operation.Execution{}, queue.WorkItem{}, false, err
	}
	if createdOp.Status == operation.StatusFailed && item.Status == queue.WorkQueued {
		row := tx.QueryRow(`
			update operation_execution
			set status = $2,
				holder = '',
				last_error = '',
				payload_hash = $3,
				updated_at = $4,
				completed_at = null
			where id = $1
			returning `+operationSelectColumns(),
			createdOp.ID,
			string(operation.StatusQueued),
			firstNonEmpty(op.PayloadHash, createdOp.PayloadHash),
			now,
		)
		createdOp, err = scanOperation(row)
		if err != nil {
			return operation.Execution{}, queue.WorkItem{}, false, err
		}
	}
	item.OperationID = createdOp.ID
	existing, found, err := findExistingWorkItemByOperation(tx, createdOp.ID)
	if err != nil {
		return operation.Execution{}, queue.WorkItem{}, false, err
	}
	if found {
		if item.Status == queue.WorkQueued && shouldRequeueOperationForWorkItem(createdOp, existing.Status, false) {
			createdOp, err = requeueOperationTx(tx, createdOp.ID, "")
			if err != nil {
				return operation.Execution{}, queue.WorkItem{}, false, err
			}
		}
		if item.Status == queue.WorkQueued && (existing.Status == queue.WorkFailed || existing.Status == queue.WorkCanceled) {
			row := tx.QueryRow(`
				update work_item
				set status = $2,
					payload = $3::jsonb,
					lease_owner = null,
					lease_expires_at = null,
					last_error = '',
					updated_at = $4,
					completed_at = null
				where id = $1
				returning `+workItemSelectColumns(),
				existing.ID,
				string(queue.WorkQueued),
				jsonString(item.Payload),
				now,
			)
			existing, err = scanWorkItem(row)
			if err != nil {
				return operation.Execution{}, queue.WorkItem{}, false, err
			}
		}
		return createdOp, existing, false, nil
	}
	if item.Status == queue.WorkQueued && shouldRequeueOperationForMissingWorkItem(createdOp, false) {
		createdOp, err = requeueOperationTx(tx, createdOp.ID, "")
		if err != nil {
			return operation.Execution{}, queue.WorkItem{}, false, err
		}
	}
	createdItem, err := enqueueWorkItemTx(tx, item)
	if err != nil {
		return operation.Execution{}, queue.WorkItem{}, false, err
	}
	return createdOp, createdItem, true, nil
}
