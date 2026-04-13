# Persistence Hardening Stage Acceptance

Use this runbook after the three app PRs and the deployment migration-job PR are merged.

## Preconditions

- Stage deployments are on the updated images.
- The stage chart includes the `PreSync` migration job.
- `improvement-plane --mode migrate` has completed successfully in Argo.
- Slack and GitHub stage credentials are present.

## Checks

1. Verify schema gating:
   - `curl -sf https://staging-rsi-platform.storyprotocol.net/readyz | jq`
   - Confirm `schema_state` is `current`.
   - Confirm `schema_current_version == schema_expected_version`.
2. Verify runtime metadata:
   - `curl -sf https://staging-rsi-platform.storyprotocol.net/api/meta | jq`
   - Confirm schema versions are exposed for debugging.
3. Trigger a Slack DM question in the allowed stage bot surface.
4. In improvement-plane, confirm the DM produced:
   - one `Conversation`
   - one `Case`
   - one terminal `Trace`
5. Force or observe a terminal failed trace for the recursive-loop path.
6. Verify eval persistence:
   - confirm an `EvalRun` exists for the failed trace
   - confirm at least one `ImprovementCandidate` exists
7. Verify promotion:
   - wait for cron or run the promoter once
   - confirm a proposal appears when the active slot cap allows it
8. Verify approval path:
   - approve the proposal
   - confirm a `RepoChangeJob` and sandbox work item are created
   - confirm the PR path can progress without schema-write errors

## Failure triage

- `readyz` not ready:
  - schema mismatch or DB connection problem; inspect migration job logs first
- eval missing for terminal trace:
  - inspect `improvement-plane` worker logs and `eval_queue` work items
- candidate missing after eval:
  - inspect `trace_event` failure metadata and `improvement_candidate` rows
- proposal not promoted:
  - inspect proposal slot occupancy and cron logs
- repo-change path stalls:
  - inspect `repo_change_job`, sandbox work items, and PR attempt records
