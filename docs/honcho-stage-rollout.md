# Honcho Stage Rollout

Use this runbook for the stage cutover from managed Honcho credentials to the
self-hosted Honcho service in the RSI platform deployment.

## Rollout Order

1. Merge and deploy the RSI platform image set that includes the Honcho baseline
   migration and Honcho image build.
2. Confirm the `PreSync` `improvement-plane --mode migrate` job completes.
3. Verify the shared stage Postgres contains:
   - `pg_extension.extname = 'vector'`
   - schema `honcho`
   - Honcho tables such as `honcho.messages`
4. Deploy the stage chart update that adds:
   - `honcho-api`
   - `honcho-deriver`
   - `honcho-redis`
   - the `standard-rwo` compatibility `StorageClass` backed by EBS `gp3`
5. Wait for `honcho-api` and `honcho-deriver` to become ready.
6. Roll the four runner roles onto the updated config with
   `RSI_HONCHO_BASE_URL=http://use1-stage-rsi-agent-platform-honcho-api:8000`.
7. Validate `/runtimez` on each runner shows:
   - `honcho_configured=true`
   - `honcho_available=true`
   - `persistence_enabled=true`

## Acceptance Checks

- `honcho-api` `/health` returns success inside the cluster.
- Redis responds to `PING`.
- Runner PVCs bind through the chart-managed `standard-rwo` compatibility alias.
- A first prompt stores memory, and a later prompt recalls it through Honcho.
- Restarting a runner preserves local Hermes session continuity from the role PVC.
- Existing stage acceptance checks in
  [`docs/persistence-hardening-stage-acceptance.md`](./persistence-hardening-stage-acceptance.md)
  still pass.

## Notes

- The stage chart owns the `standard-rwo` compatibility `StorageClass` because
  the cluster only exposes `gp3` natively and older runner PVCs still reference
  `standard-rwo`.
- The main CD workflow verifies that each target ECR repository exists before
  pushing images, including `rsi-agent-platform-honcho`.

## Rollback

1. Revert runner config away from `RSI_HONCHO_BASE_URL`.
2. Roll runners first so they stop depending on the in-cluster Honcho service.
3. Scale down `honcho-api`, `honcho-deriver`, and `honcho-redis`.
4. Leave the `honcho` schema and `vector` extension in place; do not attempt
   destructive rollback of shared database objects during stage cutback.
