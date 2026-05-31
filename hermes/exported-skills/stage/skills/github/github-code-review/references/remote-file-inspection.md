# Remote File Inspection Patterns

Techniques for inspecting files on a GitHub branch without a local checkout.
Essential for fast re-reviews and targeted fix verification.

## Read a file from a branch

```bash
gh api "repos/<owner>/<repo>/contents/<path>?ref=<branch>" --jq '.content' | base64 -d
```

The content is base64-encoded — always pipe through `base64 -d`. For binary/text autodetection, omit `--jq` and check `.encoding`.

## Search within a remote file

```bash
gh api "repos/<owner>/<repo>/contents/<path>?ref=<branch>" --jq '.content' | base64 -d | grep "pattern"
```

Useful for quick checks: "does this file still have `voice_phrase` fields?", "was `typeof` check added?"

## Inspect a specific commit's changes

```bash
# Which files were touched?
gh api "repos/<owner>/<repo>/commits/<sha>" --jq '.files[] | "\(.filename): +\(.additions) -\(.deletions)"'

# Full patch for one file in the commit
gh api "repos/<owner>/<repo>/commits/<sha>" --jq '.files[] | select(.filename == "path/to/file.ts") | .patch'
```

## Find fix commits after a review

```bash
# List last N commits on a PR
gh pr view <N> --json commits --jq '.commits[-5:] | .[] | {oid: .oid[0:8], message: .messageHeadline}'
```

Look for commits with names like "fix: address PR review" or "fix: round-2 review comments".

## Quick diff scan without checkout

```bash
# Count lines in PR diff
gh pr diff <N> | wc -l

# Check if a file was modified at all
gh pr diff <N> | grep -c "path/to/file"

# Find which locale files were updated
gh pr diff <N> | grep "diff --git.*locales"
```

## i18n/locale checklist

When reviewing locale changes:
1. Check which locales are active: `gh api .../lingui.config.ts?ref=<branch> | base64 -d | grep locales`
2. Check which locale files the PR actually modifies: `gh pr diff <N> | grep "diff --git.*locales"`
3. Only flag missing locales if they're in the active config AND not in the diff

## Pitfalls

- `base64 -d` on Windows: use `base64 --decode` or `certutil -decode`
- Large files (>1MB): GitHub's contents API returns `"size": <bytes>` — check before base64-decoding
- Binary files: skip `--jq '.content'` and check `.encoding` field instead
