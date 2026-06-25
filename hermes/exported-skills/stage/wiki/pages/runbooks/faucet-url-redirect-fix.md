---
title: "Faucet URL Redirect Fix"
type: "runbook"
slug: "runbooks/faucet-url-redirect-fix"
freshness: "2026-02-03T07:10:14Z"
tags:
  - "faucet"
  - "incident"
  - "url-redirect"
owners:
  - "U07TNT9N4JC"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_0c10dd99f980c3f1b67909c00ef4abd4"
  - "srcrev_827686def66cf421e1a4f4dee41b37d4"
  - "srcrev_8bc009e2a484a41b12a7a0a6a1578874"
  - "srcrev_e7036f5ac42ee9af3dc271eeea4f55ef"
conflict_state: "none"
---

# Faucet URL Redirect Fix

## Summary

The faucet.story.foundation URL was incorrectly pointing to a GCP faucet instead of aeneid.faucet.story.foundation. The issue was caused by a disabled setting and resolved after cache clearing.

## Claims

- The URL https://faucet.story.foundation was redirecting to a GCP faucet instead of https://aeneid.faucet.story.foundation. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_827686def66cf421e1a4f4dee41b37d4` `chunk_id=srcchunk_43c67ba6a978a4e189892cc31a40b6da` `native_locator=slack:C0547N89JUB:1770085445.129679:1770085445.129679` `source_timestamp=2026-02-03T02:24:05Z`
- The deployment of the fix may have been overwritten. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_e7036f5ac42ee9af3dc271eeea4f55ef` `chunk_id=srcchunk_111ef5d2b87a4b7a47b0f41c3c1c0621` `native_locator=slack:C0547N89JUB:1770085445.129679:1770085582.939589` `source_timestamp=2026-02-03T02:26:22Z`
- The issue was caused by a disabled setting, which was then enabled to fix the redirection. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_0c10dd99f980c3f1b67909c00ef4abd4` `chunk_id=srcchunk_155fa2ebecb9ca4160c7d8eb4ae3e13a` `native_locator=slack:C0547N89JUB:1770085445.129679:1770102493.983209` `source_timestamp=2026-02-03T07:08:13Z`
- After clearing cache, the URL redirection worked correctly. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_8bc009e2a484a41b12a7a0a6a1578874` `chunk_id=srcchunk_76d78e772213ad57a4e850cb2d711f06` `native_locator=slack:C0547N89JUB:1770085445.129679:1770102614.464139` `source_timestamp=2026-02-03T07:10:14Z`

## Open Questions

- Is cache clearing required for all users when faucet URL changes?
- What caused the setting to be disabled?

## Sources

- `source_document_id`: `srcdoc_ab7d535c8f2378de39f329dc8e775452`
- `source_revision_id`: `srcrev_727e912a611b85c746c333ecc00aeecd`
