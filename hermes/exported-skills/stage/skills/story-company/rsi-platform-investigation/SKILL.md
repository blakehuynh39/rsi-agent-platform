---
name: rsi-platform-investigation
description: "Investigate RSI agent platform deployment status, wiki system health, sync/mirror pipelines, and internal API surface. Covers verifying code-vs-deployed parity, debugging wiki sync gaps, and tracing cronjob/mirror behavior."
tags:
  - rsi-platform
  - wiki
  - kubernetes
  - deployment
  - debugging
  - mirror
  - sync
  - investigation
triggers:
  - "has X been deployed"
  - "is the wiki working"
  - "check deployment status"
  - "wiki sync bug"
  - "rsi platform investigation"
  - "check if commit X is deployed"
  - "mirror not working"
---

# RSI Platform Investigation

Verifying deployment status, debugging wiki sync gaps, and tracing mirror/pipeline behavior on the RSI agent platform.

## 1. Verify deployment status (code vs cluster)

**Goal:** Confirm a git commit is actually running in the cluster.

```bash
# Get deployed image tags
kubectl get deployments -n rsi-platform -o json | python3 -c "
import json,sys
data = json.load(sys.stdin)
for item in data.get('items',[]):
    name = item['metadata']['name']
    for c in item['spec']['template']['spec']['containers']:
        img = c['image']
        if ':' in img:
            tag = img.split(':')[-1]
            print(f'{name}: {tag}')
"

# Compare with latest git commit
cd /workspace/company/rsi-agent-platform
git fetch --all
git log origin/main --oneline -5
```

**PITFALL:** The local clone may be stale or on a different branch. Always `git fetch --all` first and use `origin/main`, not local `HEAD`. The deployed tag format is `<component>-<full_40char_sha>`. If `git show <short_sha>` fails with "ambiguous argument", the commit may not be in the local repo — fetch first.

**PITFALL: Git dubious ownership.** The workspace repos may trigger:
```
fatal: detected dubious ownership in repository at '/workspace/company/rsi-agent-platform'
```
Fix: `git config --global --add safe.directory /workspace/company/rsi-agent-platform`

## 2. Internal API access pattern

The RSI platform services are accessible via ClusterIP within the cluster. Use these stable IPs:

| Service | ClusterIP | Path prefix | Has wiki routes? |
|---|---|---|---|
| control-plane | 172.20.190.168:8080 | `/internal/company-wiki/*` | **Yes** |
| improvement-plane | 172.20.234.73:8080 | `/api/*` | **No** (returns HTML) |

**PITFALL:** The improvement-plane returns HTML (its web app shell) for the same wiki paths. Always use the control-plane for wiki API calls. The improvement-plane's `/api/` prefix works for improvement cases, not wiki.

Wiki API endpoints on control-plane:
```
GET  /internal/company-wiki/index          → returns wiki index (or error if empty)
GET  /internal/company-wiki/log?limit=N    → returns change log
GET  /internal/company-wiki/search?query=X → semantic search
GET  /internal/company-wiki/pages/*        → read a specific page by slug
POST /internal/company-wiki/manifest/reconcile?repair=true  → reconcile file system vs DB
POST /internal/company-wiki/edits/propose  → propose a new page edit
POST /internal/company-wiki/edits/apply    → apply and publish an edit
```

## 3. Wiki sync pipeline debugging

### Architecture

The wiki sync pipeline couples to the Slack/Notion mirror ingestion:
```
Slack/Notion source → mirror cronjob → IngestMessage → shouldPublishWikiSource?
  → wikiBatch.record() → RecordWikiSourceRevision() → UpsertCompanyWikiSourceRevision()
  → wikiBatch.publish() → PublishWikiSourceDocument() → WriteMarkdownFile()
```

### Common failure: empty wiki after deployment (no backfill)

**Symptom:** Wiki code is deployed, all pods green, but `/internal/company-wiki/index` returns `"open /workspace/company/wiki/index.md: no such file or directory"` and manifest reconcile returns `{"checked": 0}`.

**Root cause:** Wiki source creation is triggered only during NEW mirror message ingestion. If the mirrors already backfilled all historical content BEFORE the wiki code was deployed, they enter incremental mode and ingest zero new messages — so zero wiki sources are ever created.

**Diagnostic checks:**
```bash
# 1. Check if mirror cronjobs are running
kubectl get cronjobs -n rsi-platform

# 2. Check latest mirror job logs
kubectl logs -n rsi-platform jobs/<job-name> --tail=30

# 3. If messages=0 threads=0 → mirrors are in incremental mode with no new content
#    Look for: "mode=incremental complete messages=0 threads=0 backfill_complete=true"

# 4. Check wiki database
curl -s "http://172.20.190.168:8080/internal/company-wiki/manifest/reconcile"
# {"checked": 0} → no wiki source documents exist

# 5. Check control-plane logs for wiki activity
kubectl logs -n rsi-platform deploy/use1-stage-rsi-agent-platform-control-plane --tail=200 | grep -i wiki
```

**Fix options:**
- **Option A (quick):** Reset mirror checkpoints to force re-ingestion (creates duplicate Honcho messages)
- **Option B (correct):** Implement a wiki-backfill mode that reads existing Honcho sessions and creates wiki sources without re-ingesting
- **Option C (manual):** Use `POST /internal/company-wiki/edits/apply` to manually publish pages

### Source recording logic

The `RecordWikiSourceRevision` function checks:
1. `repo` implements `CompanyWikiStore` → if not: skipped (`store_not_company_wiki_capable`)
2. `cfg.CompanyWikiRoot` is set → if empty: skipped (`company_wiki_root_not_configured`)
3. `UpsertCompanyWikiSourceRevision` → if already exists: skipped (`revision_already_exists`)

The resulting wiki `CompanyWikiSourceRevisionResult` has an `Inserted: bool` field. Only NEW revisions (not previously upserted) proceed to publishing.

## 4. Git inspection patterns for deployed commits

```bash
# Show files changed in a specific commit (even if not at HEAD)
git show <commit>:<path>       # show file at that commit
git show <commit> --stat       # summary of changes
git show <commit> --name-only  # just filenames

# Search code across the entire branch
git grep "pattern" origin/main -- "path/"

# Check what branch the commit is on
git branch -a --contains <commit>

# Compare local HEAD vs origin/main
git rev-parse HEAD --short=7
git rev-parse origin/main --short=7
```

## 5. Kubernetes log inspection patterns

```bash
# Recent logs from a deployment
kubectl logs -n rsi-platform deploy/<name> --tail=200

# Filter for specific patterns
kubectl logs ... | grep -iE "wiki|mirror|sync|error"

# Get logs from a completed cronjob
kubectl logs -n rsi-platform jobs/<job-name> --tail=50

# List recent jobs
kubectl get jobs -n rsi-platform --sort-by=.metadata.creationTimestamp

# Check cronjob schedule
kubectl get cronjobs -n rsi-platform
```

## 6. Honcho wiki tools (Hermes-side)

The Hermes `wiki_*` tools (`wiki_index_get`, `wiki_log_get`, `wiki_search`, `wiki_page_get`) read from the file system at `/workspace/company/wiki/`. If the wiki hasn't been populated (no sources ever published), these tools return 404 errors. This is normal for a freshly-deployed wiki that hasn't received any source content yet — the tools will work once the first wiki page is published.

## 7. GitHub API for repo inspection

```bash
# List recent commits (no local clone needed)
gh api repos/piplabs/rsi-agent-platform/commits --jq '.[0:10].[] | "\(.sha[0:7]) \(.commit.message | split("\n")[0])"'

# Check repo permissions
gh api repos/piplabs/rsi-agent-platform --jq '.permissions'
```

### Diagnosing "repo not found" (GitHub App scope)

When `gh api /repos/<owner>/<repo>` returns HTTP 404 for a repo you expect to exist, the most common cause is that the GitHub App hasn't been installed on that repo. Check the installation scope:

```bash
# List all repos the GitHub App token can access
gh api /installation/repositories --jq '.repositories[] | {full_name, private}'
```

This returns the exact installation scope — if a repo isn't listed here, the token has no access to it. The 404 is GitHub's way of hiding private repos the token can't see (even if they exist).

**PITFALL:** A 404 can mean either "repo doesn't exist" OR "repo is private and this token isn't authorized." The `/installation/repositories` call is the only reliable way to distinguish them — if the repo isn't in the list, ask the repo owner to install the GitHub App on it.

**PITFALL:** For GitHub App tokens, the `/user` and `/app/installations` endpoints are typically inaccessible (require JWT, not installation token). Stick to `/installation/repositories` for scope diagnosis.
