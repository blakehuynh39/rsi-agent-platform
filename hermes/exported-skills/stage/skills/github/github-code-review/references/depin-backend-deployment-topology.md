## depin-backend Deployment Topology

> Last verified: 2026-05-05. Update when replica counts, namespaces, or target branches change.

### Kubernetes (story namespace)

```
NAMESPACE   NAME                                          READY   REPLICAS
story       use1-stage-depin-backend                      2/2     2
story       use1-stage-depin-ip-registration-confirmer    3/3     3
story       use1-stage-depin-ip-registration-poller       2/2     2
story       use1-stage-depin-ip-registration-submitter    3/3     3
```

### Architecture pattern

- **N API pods** (currently 2) all connected to **one shared PostgreSQL** instance
- Each pod runs the same set of background jobs independently — no leader election
- Source of truth: PostgreSQL (all writes converge via idempotent SQL)
- In-memory state: per-pod, not shared (e.g., `HotPathCache`)

### Background jobs (all run in every pod)

| Job | Interval | Idempotency mechanism |
|-----|----------|----------------------|
| `multiplier_sweep` | 5 min | `UPDATE ... WHERE is_active = TRUE AND expires_at <= NOW()` |
| `hot_path_cache` | 120 sec | Read-only from DB into per-pod in-memory cache |
| `idempotency_cleanup` | 24 hrs | `DELETE WHERE created_at < now() - INTERVAL '7 days'` |
| `user_safety_signals_refresh` | 5 min | `INSERT ... ON CONFLICT DO UPDATE` (introduced in #422) |

### Cross-repo pairing

- Backend: `piplabs/depin-backend` → base branch `staging`
- Frontend: `piplabs/numo-monorepo` → base branch `develop` (changed from `main` May 2026)
- Merge order: BE merges to staging and deploys **before** FE merges to develop
- Matching branch prefixes (e.g., `feat/trust-safety-cluster-and-cadence`)

### Review implications

- Any new background job multiplies its DB cost by the replica count
- `pg_try_advisory_lock` is available but not currently used by any job — introducing it would be a new pattern
- CI validates via GitHub Actions (Rust Checks, Image Builds, Validate migrations, Wiz scanners) + Vercel (FE deploys)
