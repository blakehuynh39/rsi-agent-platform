# Code Fix Workflow — Production Issue to PR

Reference for the full cycle: a Grafana/Thanos observation of anomalous behavior → source-code investigation → fix → branch + PR. Uses depin-backend as the canonical example, but the pattern applies to any repo in the piplabs ecosystem.

## Trigger

User asks "is it bad code?", "can we fix this?", "open a PR to fix X" — i.e., they want a production observation turned into a code fix.

## Workflow

### 1. Correlate Observation with Deploy Timeline

Before touching code, confirm which deploy introduced the issue:

```bash
# 1a. Check Thanos for pod creation times (epoch → human-readable)
kube_pod_created{pod=~"use1-prod-depin-backend.*"}

# 1b. Find the deploy PR — look for "staging > main" merges near the pod creation time
gh pr list --repo piplabs/depin-backend --state merged --limit 10 --search "merged:>=YYYY-MM-DD"

# 1c. For the deploy PR's description, find which feature PRs it promoted
gh pr view <N> --repo piplabs/depin-backend --json title,body
```

### 2. Clone and Inspect the Feature PR's Code

```bash
cd /tmp && gh repo clone piplabs/depin-backend depin-backend-fix
cd depin-backend-fix

# View the feature PR to understand what was changed
gh pr view <feature-pr> --json title,body,files
```

### 3. Identify the Allocating Code Path

Focus on patterns that allocate large memory in Rust:
- `fetch_all` on unbounded queries (loads entire result set into one `Vec`)
- Large `HashMap` or `BTreeMap` builds from `collect()`
- Missing `shrink_to_fit()` or capacity pre-allocation
- Intermediate `Vec` that coexists with final structures during construction

For depin-backend specifically, the `hot_path_cache` (refreshed every 60s) and `user_safety_signals_refresh` (every 5 min) background jobs are common culprits.

### 4. Write the Fix

#### Pattern: `fetch_all` → streaming `fetch`

**Before** (double-peaks memory — intermediate Vec + final structures):
```rust
let rows = sqlx::query_as::<_, Row>(...)
    .fetch_all(db)
    .await?;

let mut entries = Vec::with_capacity(rows.len());
for row in rows {
    entries.push(transform(row));
}
```

**After** (single-peak — only final structures):
```rust
use futures::StreamExt;

let mut rows = sqlx::query_as::<_, Row>(...)
    .fetch(db);

let mut entries = Vec::new();
while let Some(row) = rows.next().await {
    let row = row?;
    entries.push(transform(row));
}
```

#### Dependency Addition

If `futures` isn't already in `Cargo.toml`:
```diff
 sqlx = { version = "0.8.6", features = [...] }
+futures = "0.3"
 subtle = "2.6.1"
```

`futures-core` is already a transitive dependency via sqlx — this just exposes `StreamExt`.

### 5. Create Branch and PR

Conventions for depin-backend:
- **Branch prefix**: `fix/` for fixes, `feat/` for features, `perf/` for performance
- **Base branch**: `staging` (NOT `main`)
- **GitHub account**: The `rsi-platform-bot[bot]` account is authenticated via `GH_TOKEN`

```bash
cd /tmp/depin-backend-fix
git checkout -b fix/stream-hot-path-cache-to-reduce-memory
git add <changed files>
git config user.email "rsi-platform-bot@storyprotocol.net"
git config user.name "RSI Platform Bot"
git commit -m "perf(area): brief description of what changed"
git push origin fix/stream-hot-path-cache-to-reduce-memory

gh pr create \
  --base staging \
  --title "perf(area): brief title" \
  --body '## Summary
...
## Problem
...
## Fix
...
## Impact
...'
```

### 6. Validate PR

After opening, confirm CI passes:
```bash
gh pr checks <N> --repo piplabs/depin-backend
```

Expected checks: Rust Checks, Image Builds, Validate migrations, Wiz scanners (Data, IaC, SAST, Secret, Vulnerability).

## Pitfalls

### Wrong base branch
`gh pr create` defaults to `main`. Override with `--base staging` for depin-backend. For numo-monorepo, use `--base develop`.

### Commit message shell escaping
Backtick-heavy commit messages will break in shell. Use single quotes for the commit message body and avoid nested backticks, or use `git commit -F <file>` with a temp file.

### RSI bot permissions
The `rsi-platform-bot[bot]` token can create branches and PRs but may not have merge permissions. PRs need human approval + merge.

### Testing without cargo
The RSI staging environment may not have `cargo` installed. Verify code correctness by visual inspection or by checking if the pattern matches existing code in the repo. Formatting and clippy will be caught by CI.
