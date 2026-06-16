# Release PR Review Output Template

For `develop → main` (or `staging → main`) release PRs. Code already vetted on source branch; focus on cherry-pick integrity, CI, security, and cross-repo alignment.

## Template

```
## 📋 PR Review: <org>/<repo>#<N> — Release: <source> → <target>

**Review type**: Release PR (cherry-pick integrity + full Section 3 review)
**Policy**: Allen's develop→main always-approve unless new issues introduced

---

### 1. Cherry-Pick Integrity (Section 5e)

| Check | Result | Evidence |
|---|---|---|
| Content identity vs source | ✅ PASS / 🔴 FAIL | `git diff <target>...origin/<source> --stat` = N files, +A/-D — matches PR |
| Only expected files changed | ✅ PASS / 🔴 FAIL | N files, all from expected PRs; no extra files or merge artifacts |
| Commit messages reference origin PRs | ✅ PASS / 🔴 FAIL | Each commit references original PR number e.g. `(#407)` |
| Single-commit per feature | ✅ PASS / 🔴 FAIL | N commits = N features; `git log <target>..origin/<source>` confirms |

**Cherry-pick verdict**: ...

---

### 2. CI Status

| Check | Status |
|---|---|
| <check name> | ✅ SUCCESS / ❌ FAILURE |

**CI verdict**: ...

---

### 3. Security Quick-Scan

| Check | Result |
|---|---|
| Secrets in code | ✅ None found / 🔴 <count> found |
| XSS vectors | ✅ None found / 🔴 <count> found |
| Debug artifacts | ✅ None found / 🔴 <count> found |
| Merge conflict markers | ✅ None found / 🔴 <count> found |

---

### 4. Cross-Repo Alignment: <be-repo>#<N> ↔ <fe-repo>#<N>

| Check | Result | Detail |
|---|---|---|
| Route paths match | ✅ / 🔴 | FE calls `GET /v1/...` matches BE route definition |
| Type alignment | ✅ / 🔴 | Field names and types match across repos |
| Logic alignment | ✅ / 🔴 | Status handling, error codes match |
| PR cross-reference | ✅ / 🔴 | Both PR bodies link to each other |
| Merge order | ✅ / 🔴 | Correct merge order (BE first) |

---

### 5. Section 3 Review Summary

**Correctness**: ✅ / 🔴 findings — <summary>
**Code Quality**: ✅ / 🔴 findings — <summary>
**Testing**: ✅ / 🔴 findings — <summary>
**Performance**: ✅ / 🔴 findings — <summary>
**Documentation**: ✅ / 🔴 findings — <summary>

---

### 6. RSI_PR_REVIEW_VERDICT

```json
RSI_PR_REVIEW_VERDICT {"pr_number":<N>,"approval_safe":true,"blocking_findings":0,"verdict":"approve"}
```

**Verdict**: **APPROVE / REQUEST_CHANGES** — <one-line summary>
```

## Sections to include

1. **Cherry-Pick Integrity** — diff verification, commit traceability
2. **CI Status** — all checks, Vercel previews, Wiz scanners
3. **Security Quick-Scan** — secrets, XSS, debug statements, merge markers
4. **Cross-Repo Alignment** — if paired with a backend PR, verify routes/types/logic match
5. **Section 3 Review** — Correctness, Code Quality, Testing, Performance, Documentation
6. **Verdict** — RSI_PR_REVIEW_VERDICT JSON + one-line summary

## When to block a release PR

- 🔴 Cherry-pick integrity failure (non-empty diff showing unintended changes)
- 🔴 CI failure on a required check
- 🔴 Secrets or XSS vectors found
- 🔴 Cross-repo misalignment (route path mismatch, missing field, logic divergence)
- 🔴 Merge conflict that dropped or corrupted code
