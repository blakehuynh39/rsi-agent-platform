## RSI Platform Deployment Verification

Use these patterns to verify whether a specific commit or feature is deployed on the RSI agent platform staging environment.

### Kubernetes namespace
- `rsi-platform` — all RSI control-plane, improvement-plane, Hermes executors, Honcho, Notion MCP, skill exporter

### Commands

```bash
# 1. Get the latest commit on origin/main
cd /workspace/company/rsi-agent-platform
git fetch --all
git log origin/main --oneline -5

# 2. Check deployed image tags (extract commit hashes)
kubectl get deployments -n rsi-platform -o json | python3 -c "
import json,sys
data = json.load(sys.stdin)
for item in data.get('items',[]):
    name = item['metadata']['name']
    containers = item['spec']['template']['spec']['containers']
    for c in containers:
        img = c['image']
        if ':' in img:
            tag = img.split(':')[-1]
            print(f'{name}: {tag}')
"

# 3. Verify specific API endpoint (use ClusterIP, no auth needed internally)
curl -s http://172.20.190.168:8080/internal/company-wiki/index
```

### How to read the deployed commit hash

ECR image tags follow the pattern: `<component>-<commit_hash>`

Example: `control-plane-e1d01543410dfe214c8a4a6a26e700d388e37311`

The 40-char hex after the component name is the git commit SHA. Compare against `git log origin/main`.

### Company wiki system paths

The company wiki is served by the **control-plane** at ClusterIP 172.20.190.168:8080:

| Endpoint | Description |
|---|---|
| `/internal/company-wiki/index` | List all wiki pages |
| `/internal/company-wiki/search?query=X` | Search wiki pages |
| `/internal/company-wiki/pages/*` | Get a specific page by slug |
| `/internal/company-wiki/log?limit=N` | Recent edit log |
| `/internal/company-wiki/edits/propose` | Propose an edit (POST) |
| `/internal/company-wiki/edits/apply` | Apply an edit (POST) |
| `/internal/company-wiki/manifest/reconcile` | Reconcile wiki manifest (POST) |

### Code vs Content distinction

The company wiki has two layers:
- **Code layer:** Go API handlers, Postgres store, migrations, router — deployed via Docker images
- **Content layer:** Markdown files at `/workspace/company/wiki/` (index.md, log.md, page files) — populated by mirror ingestion from Notion/Slack sources

Both must be checked separately. A 404 on wiki API endpoints with "no such file or directory" means the code is deployed but content hasn't been populated yet.

### Common pitfalls

- **Stale local repo:** `git fetch --all` is essential — the deployed hash may only exist on `origin/main`, not in the local clone's refs
- **Git dubitable ownership:** In some environments, `git log` returns "fatal: detected dubious ownership". Fix: `git config --global --add safe.directory /path/to/repo`
- **Wiki empty shell:** The wiki code returns 200 on search (empty results) but 404 on index/pages when the file directory hasn't been seeded
- **Hermes wiki tools vs API:** The `wiki_index_get`/`wiki_search` Hermes tools hit the file-based system, not the HTTP API directly — both layers must be functional
