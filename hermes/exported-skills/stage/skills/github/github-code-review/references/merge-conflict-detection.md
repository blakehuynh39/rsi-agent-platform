# Merge Conflict Detection During PR Review

## When to Apply

A PR is marked `mergeable: CONFLICTING` in `gh pr view` metadata. This means the PR branch has diverged from the base branch and cannot be merged as-is. As a reviewer, you must determine **why** it conflicts and **what code would be affected** by a merge resolution.

## The Danger: Silent Code Drops

The most dangerous merge conflict is when the **base branch** (main) has gained new code since the PR was branched, and the PR branch doesn't include it. If merged without proper conflict resolution, the new code is **silently deleted**.

This happens when:
- Another PR was merged to main after this PR was branched
- The new code on main touches the same file(s) as the PR
- Git's merge algorithm can't reconcile the two changes automatically

## Detection Workflow

### Step 1: Check the PR metadata

```bash
gh pr view N --json mergeable,baseRefName,headRefName --jq '{mergeable, base: .baseRefName, head: .headRefName}'
```

If `mergeable` is `CONFLICTING`, proceed to Step 2.

### Step 2: Find the merge base

```bash
git fetch origin
MERGE_BASE=$(git merge-base origin/main pr-N)
```

### Step 3: Compare PR branch vs main (not just merge-base)

The critical comparison is **PR branch vs the tip of main**, not PR branch vs merge-base:

```bash
# What main has that PR branch is MISSING (the danger zone)
git diff pr-N..origin/main -- runner/rsi_runner/pr_review_gate.py

# What PR branch has that main is MISSING (the actual PR changes)
git diff origin/main..pr-N -- runner/rsi_runner/pr_review_gate.py
```

The first diff shows code that exists on main but NOT on the PR branch. If this includes entire functions, constants, or guard logic, merging without rebase would drop them.

### Step 4: Verify the merge-base context

```bash
# Does the merge base already have the feature merged to main?
git show $MERGE_BASE:path/to/file.py | grep -c "feature_indicator"
git show origin/main:path/to/file.py | grep -c "feature_indicator"
```

If `origin/main` has substantially more code than the merge base, the PR was branched before that code was added.

### Step 5: Read the full file from both branches

```bash
# Main version (the ground truth)
git show origin/main:path/to/file.py

# PR version (what merging would produce)
git show pr-N:path/to/file.py
```

Compare line counts, function presence, and structural differences. Anything in main but not in the PR branch is at risk.

## PITFALLS

**PITFALL: Trusting the GitHub PR diff alone.** The PR diff shows `pr-N` vs `merge-base`. If main has advanced beyond the merge base, the PR diff looks clean but the merge would be destructive. Always compare against the TIP of main, not the merge base.

**PITFALL: Assuming ALL conflicts are safe rebases.** A conflict on a single line can mask a much larger divergence. A PR adding 5 lines to a function might conflict because main renamed the function — but the real issue is that main also added 60 lines of new guard code that the PR doesn't know about.

**PITFALL: Using `git diff main...pr-N` (triple-dot).** Triple-dot diff shows `merge-base..pr-N`, not `main..pr-N`. For merge conflict detection, you need the double-dot comparison against the tip of main.

## Example: PR #1416 (Cache normalization)

- PR branched from `063c9948`, which was before PR #1421 (workspace guard) merged at `6dc6ae9d`
- `git show pr-1416:runner/rsi_runner/pr_review_gate.py | grep -c "_PR_REVIEW_MUTATING"` → 0
- `git show origin/main:runner/rsi_runner/pr_review_gate.py | grep -c "_PR_REVIEW_MUTATING"` → 4
- **Finding**: PR would silently drop all workspace guard constants and the `_pr_review_workspace_mutation_block()` function
- **Verdict**: CRITICAL — must rebase onto main before merging

## Remediation Instructions for the Author

When flagging a conflict-driven code drop, give the author specific instructions:

```
The PR must be rebased onto origin/main (specifically commit <sha> or later).
After rebase, verify that the workspace guard constants and functions from
PR #<N> are preserved alongside the new cached set.
```

## Working with Sparse/Fresh Clones

When reviewing a PR from a fresh clone (not an existing repo checkout), avoid `--filter=blob:none --sparse` unless you also handle sparse checkout:

```bash
# Good: full clone (slower but complete)
git clone https://github.com/owner/repo.git
cd repo && git fetch origin pull/N/head:pr-N && git checkout pr-N

# Alternative: if sparse clone is needed for speed, read files via git show
git clone --filter=blob:none --sparse https://github.com/owner/repo.git
cd repo
git show pr-N:path/to/file.py  # reads WITHOUT checkout, bypasses sparse filter
```

**PITFALL: `git sparse-checkout disable` on large repos can time out** (30+ seconds for repos with thousands of files). Use `git show <ref>:<path>` to read individual files without expanding the working tree.
