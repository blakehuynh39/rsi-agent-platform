package store

import (
	"database/sql"
	"strings"
	"time"

	"github.com/piplabs/rsi-agent-platform/internal/improvement"
)

func (p *PostgresStore) ListAttemptWorkspaces() []improvement.AttemptWorkspace {
	store, err := p.readStore()
	if err != nil {
		return []improvement.AttemptWorkspace{}
	}
	return store.ListAttemptWorkspaces()
}

func (p *PostgresStore) GetAttemptWorkspace(workspaceID string) (improvement.AttemptWorkspace, bool) {
	var item improvement.AttemptWorkspace
	err := p.withTx(func(tx *sql.Tx) error {
		var err error
		item, err = selectAttemptWorkspaceTx(tx, strings.TrimSpace(workspaceID), false)
		return err
	})
	if err != nil {
		return improvement.AttemptWorkspace{}, false
	}
	return normalizeAttemptWorkspace(item), true
}

func (p *PostgresStore) GetAttemptWorkspaceByAttempt(attemptID string) (improvement.AttemptWorkspace, bool) {
	var (
		item improvement.AttemptWorkspace
		ok   bool
	)
	err := p.withTx(func(tx *sql.Tx) error {
		var err error
		item, ok, err = selectAttemptWorkspaceByAttemptTx(tx, strings.TrimSpace(attemptID), false)
		return err
	})
	if err != nil || !ok {
		return improvement.AttemptWorkspace{}, false
	}
	return normalizeAttemptWorkspace(item), true
}

func (p *PostgresStore) upsertAttemptWorkspaceDirect(item improvement.AttemptWorkspace) (improvement.AttemptWorkspace, error) {
	item = normalizeAttemptWorkspace(item)
	now := time.Now().UTC()
	if item.ID == "" {
		item.ID = nextID("ws", 0)
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	err := p.withTx(func(tx *sql.Tx) error {
		temp := newSubsetStore()
		temp.attemptWorkspaces[item.ID] = item
		return persistAttemptWorkspaces(tx, temp)
	})
	if err != nil {
		return improvement.AttemptWorkspace{}, err
	}
	return item, nil
}

func loadAttemptWorkspaces(r sqlReader, store *MemoryStore) error {
	rows, err := r.Query(`select id, attempt_id, proposal_id, repo, base_ref, branch_name, namespace, job_name, pod_name, status, allowed_path_globs, head_sha, diff_summary, created_at, updated_at, expires_at from attempt_workspace order by created_at desc`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item improvement.AttemptWorkspace
		var namespace, jobName, podName, headSHA, diffSummary sql.NullString
		var allowed []byte
		var expiresAt sql.NullTime
		var status string
		if err := rows.Scan(&item.ID, &item.AttemptID, &item.ProposalID, &item.Repo, &item.BaseRef, &item.BranchName, &namespace, &jobName, &podName, &status, &allowed, &headSHA, &diffSummary, &item.CreatedAt, &item.UpdatedAt, &expiresAt); err != nil {
			return err
		}
		item.Namespace = namespace.String
		item.JobName = jobName.String
		item.PodName = podName.String
		item.Status = improvement.AttemptWorkspaceStatus(status)
		item.AllowedPathGlobs = decodeJSON(allowed, []string{})
		item.HeadSHA = headSHA.String
		item.DiffSummary = diffSummary.String
		if expiresAt.Valid {
			t := expiresAt.Time
			item.ExpiresAt = &t
		}
		store.attemptWorkspaces[item.ID] = normalizeAttemptWorkspace(item)
	}
	return rows.Err()
}

func persistAttemptWorkspaces(tx *sql.Tx, store *MemoryStore) error {
	keys := sortedMapKeys(store.attemptWorkspaces)
	for _, key := range keys {
		item := normalizeAttemptWorkspace(store.attemptWorkspaces[key])
		if _, err := tx.Exec(`insert into attempt_workspace (id, attempt_id, proposal_id, repo, base_ref, branch_name, namespace, job_name, pod_name, status, allowed_path_globs, head_sha, diff_summary, created_at, updated_at, expires_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12,$13,$14,$15,$16)
			on conflict (id) do update set
				attempt_id = excluded.attempt_id,
				proposal_id = excluded.proposal_id,
				repo = excluded.repo,
				base_ref = excluded.base_ref,
				branch_name = excluded.branch_name,
				namespace = excluded.namespace,
				job_name = excluded.job_name,
				pod_name = excluded.pod_name,
				status = excluded.status,
				allowed_path_globs = excluded.allowed_path_globs,
				head_sha = excluded.head_sha,
				diff_summary = excluded.diff_summary,
				created_at = excluded.created_at,
				updated_at = excluded.updated_at,
				expires_at = excluded.expires_at`,
			item.ID, item.AttemptID, item.ProposalID, item.Repo, item.BaseRef, item.BranchName, nullString(item.Namespace), nullString(item.JobName), nullString(item.PodName), string(item.Status), jsonString(item.AllowedPathGlobs), nullString(item.HeadSHA), item.DiffSummary, item.CreatedAt, item.UpdatedAt, nullTime(item.ExpiresAt),
		); err != nil {
			return err
		}
	}
	return nil
}

func selectAttemptWorkspaceTx(tx *sql.Tx, workspaceID string, forUpdate bool) (improvement.AttemptWorkspace, error) {
	query := `select id, attempt_id, proposal_id, repo, base_ref, branch_name, namespace, job_name, pod_name, status, allowed_path_globs, head_sha, diff_summary, created_at, updated_at, expires_at from attempt_workspace where id = $1`
	if forUpdate {
		query += ` for update`
	}
	var item improvement.AttemptWorkspace
	var namespace, jobName, podName, headSHA, diffSummary sql.NullString
	var allowed []byte
	var expiresAt sql.NullTime
	var status string
	if err := tx.QueryRow(query, strings.TrimSpace(workspaceID)).Scan(&item.ID, &item.AttemptID, &item.ProposalID, &item.Repo, &item.BaseRef, &item.BranchName, &namespace, &jobName, &podName, &status, &allowed, &headSHA, &diffSummary, &item.CreatedAt, &item.UpdatedAt, &expiresAt); err != nil {
		return improvement.AttemptWorkspace{}, err
	}
	item.Namespace = namespace.String
	item.JobName = jobName.String
	item.PodName = podName.String
	item.Status = improvement.AttemptWorkspaceStatus(status)
	item.AllowedPathGlobs = decodeJSON(allowed, []string{})
	item.HeadSHA = headSHA.String
	item.DiffSummary = diffSummary.String
	if expiresAt.Valid {
		t := expiresAt.Time
		item.ExpiresAt = &t
	}
	return normalizeAttemptWorkspace(item), nil
}

func selectAttemptWorkspaceByAttemptTx(tx *sql.Tx, attemptID string, forUpdate bool) (improvement.AttemptWorkspace, bool, error) {
	query := `select id from attempt_workspace where attempt_id = $1 order by created_at desc limit 1`
	if forUpdate {
		query += ` for update`
	}
	var workspaceID string
	if err := tx.QueryRow(query, strings.TrimSpace(attemptID)).Scan(&workspaceID); err != nil {
		if err == sql.ErrNoRows {
			return improvement.AttemptWorkspace{}, false, nil
		}
		return improvement.AttemptWorkspace{}, false, err
	}
	item, err := selectAttemptWorkspaceTx(tx, workspaceID, false)
	if err != nil {
		return improvement.AttemptWorkspace{}, false, err
	}
	return item, true, nil
}
