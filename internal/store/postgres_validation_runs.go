package store

import (
	"database/sql"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
)

func (p *PostgresStore) ListValidationRuns() []improvement.ValidationRun {
	store, err := p.readStore()
	if err != nil {
		return nil
	}
	return store.ListValidationRuns()
}

func (p *PostgresStore) RecordValidationRun(run improvement.ValidationRun) (item improvement.ValidationRun, err error) {
	run = normalizeValidationRun(run)
	now := time.Now().UTC()
	if run.ID == "" {
		run.ID = nextID("validation", 0)
	}
	if run.CreatedAt.IsZero() {
		run.CreatedAt = now
	}
	if run.UpdatedAt.IsZero() || run.UpdatedAt.Before(run.CreatedAt) {
		run.UpdatedAt = run.CreatedAt
	}
	err = p.withTx(func(tx *sql.Tx) error {
		temp := newSubsetStore()
		temp.validationRuns[run.ID] = run
		if err := persistValidationRuns(tx, temp); err != nil {
			return err
		}
		item = run
		return nil
	})
	return
}

func loadValidationRuns(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, proposal_id, attempt_id, conversation_id, case_id, origin_trace_id, workspace_id, operation_id, generation, repo, branch_name, command, status, sandbox_namespace, sandbox_job_name, sandbox_pod_name, validation_ref, error_message, log_artifact_id, created_at, updated_at from validation_run order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item improvement.ValidationRun
		var attemptID, conversationID, caseID, originTraceID, workspaceID, operationID, repo, branchName, command, sandboxNamespace, sandboxJobName, sandboxPodName, validationRef, errorMessage, logArtifactID sql.NullString
		var generation sql.NullInt64
		var status string
		if err := rows.Scan(
			&item.ID,
			&item.ProposalID,
			&attemptID,
			&conversationID,
			&caseID,
			&originTraceID,
			&workspaceID,
			&operationID,
			&generation,
			&repo,
			&branchName,
			&command,
			&status,
			&sandboxNamespace,
			&sandboxJobName,
			&sandboxPodName,
			&validationRef,
			&errorMessage,
			&logArtifactID,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return err
		}
		item.AttemptID = attemptID.String
		item.ConversationID = conversationID.String
		item.CaseID = caseID.String
		item.OriginTraceID = originTraceID.String
		item.WorkspaceID = workspaceID.String
		item.OperationID = operationID.String
		if generation.Valid {
			item.Generation = int(generation.Int64)
		}
		item.Repo = repo.String
		item.BranchName = branchName.String
		item.Command = command.String
		item.Status = improvement.ValidationRunStatus(status)
		item.SandboxNamespace = sandboxNamespace.String
		item.SandboxJobName = sandboxJobName.String
		item.SandboxPodName = sandboxPodName.String
		item.ValidationRef = validationRef.String
		item.ErrorMessage = errorMessage.String
		item.LogArtifactID = logArtifactID.String
		store.validationRuns[item.ID] = normalizeValidationRun(item)
	}
	return rows.Err()
}

func persistValidationRuns(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.validationRuns)
	for _, key := range keys {
		item := normalizeValidationRun(store.validationRuns[key])
		if _, err := tx.Exec(`insert into validation_run (id, proposal_id, attempt_id, conversation_id, case_id, origin_trace_id, workspace_id, operation_id, generation, repo, branch_name, command, status, sandbox_namespace, sandbox_job_name, sandbox_pod_name, validation_ref, error_message, log_artifact_id, created_at, updated_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
			on conflict (id) do update set
				proposal_id = excluded.proposal_id,
				attempt_id = excluded.attempt_id,
				conversation_id = excluded.conversation_id,
				case_id = excluded.case_id,
				origin_trace_id = excluded.origin_trace_id,
				workspace_id = excluded.workspace_id,
				operation_id = excluded.operation_id,
				generation = excluded.generation,
				repo = excluded.repo,
				branch_name = excluded.branch_name,
				command = excluded.command,
				status = excluded.status,
				sandbox_namespace = excluded.sandbox_namespace,
				sandbox_job_name = excluded.sandbox_job_name,
				sandbox_pod_name = excluded.sandbox_pod_name,
				validation_ref = excluded.validation_ref,
				error_message = excluded.error_message,
				log_artifact_id = excluded.log_artifact_id,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at`,
			item.ID,
			item.ProposalID,
			firstNonEmpty(item.AttemptID),
			nullString(item.ConversationID),
			nullString(item.CaseID),
			nullString(item.OriginTraceID),
			nullString(item.WorkspaceID),
			nullString(item.OperationID),
			nullInt64(int64(item.Generation)),
			nullString(item.Repo),
			nullString(item.BranchName),
			nullString(item.Command),
			string(item.Status),
			nullString(item.SandboxNamespace),
			nullString(item.SandboxJobName),
			nullString(item.SandboxPodName),
			nullString(item.ValidationRef),
			nullString(item.ErrorMessage),
			nullString(item.LogArtifactID),
			item.CreatedAt,
			item.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}
