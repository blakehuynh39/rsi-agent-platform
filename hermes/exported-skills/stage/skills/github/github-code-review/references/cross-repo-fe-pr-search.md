# Cross-Repo FE PR Search Checklist

When a depin-backend PR claims to have a linked FE PR on numo-monorepo but no URL is provided, run through this checklist to confirm whether it exists.

## Step 1: Direct head-branch search (most precise)

```bash
gh pr list --repo piplabs/numo-monorepo --state all --head <branch-name> \
  --json number,title,state,headRefName,baseRefName,url
```

If this returns a result, the PR exists and you have its URL.

**PITFALL:** `gh search prs --head <branch>` may also work but has a narrower `--json` field set. `gh pr list --head` is more reliable for getting full PR metadata.

## Step 2: Topic/title search (broader, catches renames)

```bash
gh pr list --repo piplabs/numo-monorepo --search "<keyword>" \
  --json number,title,state,headRefName,url
```

Use the functional keyword from the feature (e.g., "resume", "multiplier").

**PITFALL:** `gh pr list --search` matches against title AND body, so merged/closed PRs that mention the word in their description will appear. Filter results by `headRefName` to find the actual matching branch.

## Step 3: Body-content search (catch PRs with different titles)

```bash
gh search prs --repo piplabs/numo-monorepo --match body "<keyword>" \
  --json number,title,state,url
```

Checks PR descriptions for the feature keyword. This catches PRs where the title diverged from the branch name.

## Step 4: Full open PR scan (last resort)

```bash
gh pr list --repo piplabs/numo-monorepo --state open --limit 50 \
  --json number,title,headRefName,url
```

Scan manually for any PR with a matching feature or a branch prefix that resembles the expected one.

## Verdict

- **PR found:** Link it in the review and proceed with cross-repo alignment checks.
- **No PR found after all 4 steps, but BE PR claims one exists:** Flag as HIGH — `missing-cross-repo-pair`. The BE PR description is inaccurate.
- **No PR found, BE PR does NOT claim a linked FE PR, and the changes don't touch API surface:** No gap to flag.
- **No PR found, BE PR does NOT claim a linked FE PR, but the changes DO add routes/schemas:** Flag as HIGH — `missing-cross-repo-pair` per the cross-repo contract.
