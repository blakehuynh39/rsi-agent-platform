# Review Output Template

Use this as the structure for PR review summary comments. Copy and fill in the sections.

## For PR Summary Comment

```markdown
## Code Review Summary

**Verdict: [Approved ✅ | Changes Requested 🔴 | Reviewed 💬]** ([N] issues, [N] suggestions)

**PR:** #[number] — [title]
**Author:** @[username]
**Files changed:** [N] (+[additions] -[deletions])

### 🔴 Critical
<!-- Issues that MUST be fixed before merge -->
- **file.py:line** — [description]. Suggestion: [fix].

### ⚠️ Warnings
<!-- Issues that SHOULD be fixed, but not strictly blocking -->
- **file.py:line** — [description].

### 💡 Suggestions
<!-- Non-blocking improvements, style preferences, future considerations -->
- **file.py:line** — [description].

### ✅ Looks Good
<!-- Call out things done well — positive reinforcement -->
- [aspect that was done well]

---
*Reviewed by Hermes Agent*
```

## Severity Guide

| Level | Icon | When to use | Blocks merge? |
|-------|------|-------------|---------------|
| Critical | 🔴 | Security vulnerabilities, data loss risk, crashes, broken core functionality | **YES — always** |
| High | 🔴 | Bugs in core logic, data integrity risks, cross-repo misalignment with production risk, missing critical error handling, performance regressions that would impact prod | **YES — always** |
| Medium | 🟡 | Missing non-critical error handling, non-idempotent migrations, missing docs/runbook, no retention policy, stale references | At reviewer discretion |
| Low / Suggestion | 💡 | Style improvements, refactoring ideas, dead code, minor redundancy, future considerations | No |
| Looks Good | ✅ | Clean patterns, good test coverage, clear naming, smart design decisions | N/A |

## Verdict Decision

- **Approved ✅** — Zero CRITICAL and zero HIGH items. Only MEDIUM, LOW, or suggestions at most.
- **Changes Requested 🔴** — Any CRITICAL or HIGH item exists. These are always blocking.
- **Reviewed 💬** — Observations only (draft PRs, uncertain findings, informational).

**🚫 HARD RULE: Never approve a PR with CRITICAL or HIGH issues. No exceptions.**

## For Inline Comments

Prefix inline comments with the severity icon so they're scannable:

```
🔴 **Critical:** User input passed directly to SQL query — use parameterized queries to prevent injection.
```

```
⚠️ **Warning:** This error is silently swallowed. At minimum, log it.
```

```
💡 **Suggestion:** This could be simplified with a dict comprehension:
`{k: v for k, v in items if v is not None}`
```

```
✅ **Nice:** Good use of context manager here — ensures cleanup on exceptions.
```

## For Re-Review (Delta Review) Summary

When re-reviewing after fixes, use this structure instead of the full review format:

```markdown
## RE-REVIEW STATUS: [repo]#[N]

**Delta review of fix commits** <code>[\`sha1\`, \`sha2\`]</code> against previous findings.

### Finding 1 — 🟡 MEDIUM: [title]

**Status: [FIXED ✅ | PARTIAL 🟡 | NOT FIXED 🔴]**

**What was expected:** [expected fix]

**What was delivered:** [actual changes in fix commits]

**What's missing:** [remaining gap, if any]

**Evidence:** <code>path/to/file.rs:123-145</code> — [specific code or diff showing status]

### Finding 2 — ...

### Summary Table

| # | Finding | Status | Key Evidence |
|---|---------|--------|-------------|
| 1 | [title] | **FIXED** | [one-liner] |
| 2 | [title] | **PARTIAL** | [one-liner] |
...

### New Findings

[List any new issues introduced by the fix commits, or "None"]
```

Every finding status MUST cite exact code locations and git evidence (commit SHA, diff line, or file:line). Do not report "FIXED" without showing what changed. Do not report "NOT FIXED" without pointing to the specific lines that still have the issue.

## For Local (Pre-Push) Review

When reviewing locally before push, use the same structure but present it as a message to the user instead of a PR comment. Skip the PR metadata header and just start with the severity sections.
