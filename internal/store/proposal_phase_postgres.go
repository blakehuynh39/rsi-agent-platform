package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/queue"
)

func persistProposalAttemptPhaseMutationTx(tx *sql.Tx, store *MemoryStore, result proposalAttemptPhaseMutationResult) error {
	if result.CurrentWorkItem != "" {
		item, ok := store.workItems[result.CurrentWorkItem]
		if !ok {
			return fmt.Errorf("proposal phase work item %s not found in loaded store", result.CurrentWorkItem)
		}
		if err := replaceWorkItemScope(tx, item); err != nil {
			return err
		}
	}
	if result.CurrentOperation != "" {
		item, ok := store.operations[result.CurrentOperation]
		if !ok {
			return fmt.Errorf("proposal phase operation %s not found in loaded store", result.CurrentOperation)
		}
		if err := replaceOperationScope(tx, item); err != nil {
			return err
		}
	}
	if result.ProposalID != "" {
		if err := replaceProposalScope(tx, store, result.ProposalID); err != nil {
			return err
		}
	}
	if result.AttemptID != "" {
		if item, ok := store.changeAttempts[result.AttemptID]; ok {
			if err := replaceChangeAttemptScope(tx, item); err != nil {
				return err
			}
		}
	}
	if result.CandidateKey != "" {
		if err := replaceCandidateScope(tx, store, result.CandidateKey); err != nil {
			return err
		}
	}
	if result.WorkspaceID != "" {
		if item, ok := store.attemptWorkspaces[result.WorkspaceID]; ok {
			if err := replaceAttemptWorkspaceScope(tx, item); err != nil {
				return err
			}
		}
	}
	if result.RepoJobTouched && result.ProposalID != "" {
		if err := replaceRepoChangeJobScope(tx, store, result.ProposalID); err != nil {
			return err
		}
	}
	if result.TraceID != "" {
		if err := replaceTraceScope(tx, store, result.TraceID); err != nil {
			return err
		}
	}
	if result.NextOperation != "" {
		item, ok := store.operations[result.NextOperation]
		if !ok {
			return fmt.Errorf("next proposal phase operation %s not found in loaded store", result.NextOperation)
		}
		if err := replaceOperationScope(tx, item); err != nil {
			return err
		}
	}
	if result.NextWorkItem != "" {
		item, ok := store.workItems[result.NextWorkItem]
		if !ok {
			return fmt.Errorf("next proposal phase work item %s not found in loaded store", result.NextWorkItem)
		}
		if err := replaceWorkItemScope(tx, item); err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStore) AdvanceProposalAttemptPhase(req ProposalAttemptPhaseAdvance) error {
	proposalID := strings.TrimSpace(req.ProposalID)
	if req.Proposal != nil {
		proposalID = firstNonEmpty(proposalID, req.Proposal.ID)
	}
	if proposalID == "" {
		return fmt.Errorf("proposal_id is required for proposal phase advance")
	}
	return p.withProposalLockedStoreTx(proposalID, func(tx *sql.Tx, store *MemoryStore) error {
		result, err := store.advanceProposalAttemptPhaseLocked(req)
		if err != nil {
			return err
		}
		return persistProposalAttemptPhaseMutationTx(tx, store, result)
	})
}

func (p *PostgresStore) DeferProposalAttemptPhase(req ProposalAttemptPhaseDefer) error {
	proposalID := strings.TrimSpace(req.ProposalID)
	if req.Proposal != nil {
		proposalID = firstNonEmpty(proposalID, req.Proposal.ID)
	}
	if proposalID == "" {
		return fmt.Errorf("proposal_id is required for proposal phase defer")
	}
	return p.withProposalLockedStoreTx(proposalID, func(tx *sql.Tx, store *MemoryStore) error {
		result, err := store.deferProposalAttemptPhaseLocked(req)
		if err != nil {
			return err
		}
		return persistProposalAttemptPhaseMutationTx(tx, store, result)
	})
}

func (p *PostgresStore) FailProposalAttemptPhase(req ProposalAttemptPhaseFailure) error {
	proposalID := strings.TrimSpace(req.ProposalID)
	if req.Proposal != nil {
		proposalID = firstNonEmpty(proposalID, req.Proposal.ID)
	}
	if proposalID == "" {
		return fmt.Errorf("proposal_id is required for proposal phase failure")
	}
	return p.withProposalLockedStoreTx(proposalID, func(tx *sql.Tx, store *MemoryStore) error {
		result, err := store.failProposalAttemptPhaseLocked(req)
		if err != nil {
			return err
		}
		return persistProposalAttemptPhaseMutationTx(tx, store, result)
	})
}

func (p *PostgresStore) ReconcileProposalAttemptPhase(proposalID string, requestedBy string) (item queue.WorkItem, queued bool, err error) {
	proposalID = strings.TrimSpace(proposalID)
	if proposalID == "" {
		return queue.WorkItem{}, false, fmt.Errorf("proposal_id is required for proposal phase reconcile")
	}
	err = p.withProposalLockedStoreTx(proposalID, func(tx *sql.Tx, store *MemoryStore) error {
		var reconcileErr error
		item, queued, reconcileErr = store.reconcileProposalAttemptPhaseLocked(proposalID, requestedBy)
		if reconcileErr != nil {
			return reconcileErr
		}
		if !queued {
			return nil
		}
		if err := replaceWorkItemScope(tx, item); err != nil {
			return err
		}
		if item.OperationID != "" {
			op, ok := store.operations[item.OperationID]
			if !ok {
				return fmt.Errorf("reconciled proposal phase operation %s not found in loaded store", item.OperationID)
			}
			if err := replaceOperationScope(tx, op); err != nil {
				return err
			}
		}
		return nil
	})
	return
}
