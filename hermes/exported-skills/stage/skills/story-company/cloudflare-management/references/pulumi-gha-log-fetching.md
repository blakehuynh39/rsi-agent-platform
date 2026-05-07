# Fetching Pulumi Workflow Logs from GitHub Actions

When a Pulumi run fails and you need to see the full error output.

## Standard approach (may return empty)

```bash
gh run view <RUN_ID> --repo piplabs/cloudflare --log
gh run view <RUN_ID> --repo piplabs/cloudflare --log-failed
gh run view <RUN_ID> --repo piplabs/cloudflare --job <JOB_ID> --log
```

**PITFALL:** These commands may return empty output even when logs exist. This was observed with Pulumi GHA logs (pulumi/actions@v6). Don't trust empty output — try the zip approach.

## Reliable approach: download zipped logs

```bash
# Download the logs as a zip file
gh api "repos/piplabs/cloudflare/actions/runs/<RUN_ID>/logs" > /tmp/run-logs.zip

# Extract with Python (unzip may not be installed in container environments)
python3 -c "
import zipfile
with zipfile.ZipFile('/tmp/run-logs.zip', 'r') as zf:
    for name in zf.namelist():
        print(f'=== {name} ===')
        print(zf.read(name).decode('utf-8', errors='replace'))
"
```

## Finding specific errors in large logs

```python
import zipfile

with zipfile.ZipFile("/tmp/run-logs.zip", "r") as zf:
    content = zf.read("0_Pulumi.txt").decode("utf-8", errors="replace")
    lines = content.split("\n")
    for i, line in enumerate(lines):
        if "error" in line.lower() or "fail" in line.lower():
            start = max(0, i-3)
            end = min(len(lines), i+4)
            for j in range(start, end):
                print(f"{j}: {lines[j]}")
            print("---")
```

## Useful gh commands for workflow investigation

```bash
# View run metadata
gh run view <RUN_ID> --repo piplabs/cloudflare --json name,status,conclusion,headBranch,displayTitle

# List recent runs on a branch
gh run list --repo piplabs/cloudflare --branch main --limit 10 --json databaseId,conclusion,displayTitle,createdAt

# List jobs in a run
gh run view <RUN_ID> --repo piplabs/cloudflare --json jobs --jq '.jobs[] | "\(.name) id=\(.databaseId) status=\(.status) conclusion=\(.conclusion)"'

# Check PR merge status
gh pr view <PR_NUMBER> --repo piplabs/cloudflare --json state,mergedAt,mergeCommit
```

## Session evidence

- Run #25461941289 (failed): `gh run view --log` returned empty. Downloaded zip via `gh api` to find error 20120.
- Run #25465998689 (success): confirmed via `gh run list` after PR #204 merge.
