# Pre-Cloned Workspace Review Pattern

When a PR review workspace already exists at `/tmp/<repo>-review/` with the PR branch
checked out, skip `gh pr diff` and read files directly from disk. This is faster than
remote diffing and gives richer context.

## Entry Point

The workspace is injected by RSI context. Verify the branch:
```bash
cd /tmp/<repo>-review/
git branch --show-current
git log --oneline -1
```

If the branch is stale, pull:
```bash
git fetch origin pull/<N>/head:pr-<N> && git checkout pr-<N>
```

## Review Approach

### 1. Map the change surface
Read every changed file in full using `read_file`. Don't stop at diffs — full-file reads
reveal imports, surrounding logic, and type definitions that a diff would hide.

### 2. Trace the dependency graph
For each changed module, identify its consumers and dependencies:
- Use `search_files` with the imported name (e.g., `isTaxFormPendingVerification`) to find
  all callers in the repo
- Verify every consumer uses the changed API correctly
- Check that imports come from the right module (not a stale re-export)

### 3. Type union / Set membership analysis
When a change modifies Set membership (which statuses belong in which category):
1. Read the full type definition (e.g., `TaxFormStatus` union)
2. Identify ALL sets that reference those values (e.g., `TAX_IN_PROGRESS`,
   `TAX_PENDING_VERIFICATION`, `TAX_INCOMPLETE_IN_PROGRESS`, `TAX_TERMINAL`)
3. Trace every function that uses each set — both the rewritten helpers and downstream
   consumers (components, hooks, other helpers) that call those helpers
4. Verify no status falls through all guards without handling, and no two categories
   overlap except intentionally (e.g., `TAX_IN_PROGRESS` is a superset of
   `TAX_PENDING_VERIFICATION`)
5. Check cross-cutting concerns: copy text functions, label formatters, UI gating
   conditions, test assertions, and documentation

### 4. Cross-repo alignment
When the PR pairs with a backend change:
- Verify status/gating semantics match between FE sets and BE API contracts
- Check that new error codes (e.g., 409) are handled in the FE
- Read the API reference docs to confirm the described behavior matches the code

## Pitfalls

- **Reading only the diff**: A diff shows what changed, not the surrounding context needed
  to verify correctness. Always read full files for the core logic.
- **Missing cross-file consumers**: A helper function change may affect components in
  distant directories. Use `search_files` to find all references.
- **Stale workspace**: The pre-cloned repo may be outdated. Verify with `git log` before
  reviewing.
