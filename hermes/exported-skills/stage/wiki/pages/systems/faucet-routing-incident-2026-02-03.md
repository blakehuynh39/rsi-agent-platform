---
title: "Faucet Routing Incident (2026-02-03)"
type: "system"
slug: "systems/faucet-routing-incident-2026-02-03"
freshness: "2026-02-03T07:10:14Z"
tags:
  - "faucet"
  - "frontend"
  - "incident"
  - "routing"
owners:
  - "U07TNT9N4JC"
source_revision_ids:
  - "srcrev_0c10dd99f980c3f1b67909c00ef4abd4"
  - "srcrev_827686def66cf421e1a4f4dee41b37d4"
  - "srcrev_8bc009e2a484a41b12a7a0a6a1578874"
  - "srcrev_e7036f5ac42ee9af3dc271eeea4f55ef"
conflict_state: "none"
---

# Faucet Routing Incident (2026-02-03)

## Summary

On 2026-02-03, the public faucet URL https://faucet.story.foundation was routed to the legacy GCP endpoint instead of https://aeneid.faucet.story.foundation/. The issue was resolved by disabling a suspected component and clearing cache.

## Claims

- On 2026-02-03, the URL https://faucet.story.foundation was observed to be pointing to the legacy GCP faucet instead of the correct https://aeneid.faucet.story.foundation/. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_827686def66cf421e1a4f4dee41b37d4` `chunk_id=srcchunk_43c67ba6a978a4e189892cc31a40b6da` `native_locator=slack:C0547N89JUB:1770085445.129679:1770085445.129679` `source_timestamp=2026-02-03T02:24:05Z`
- The misconfiguration may have been due to a deployment that was overwritten. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_e7036f5ac42ee9af3dc271eeea4f55ef` `chunk_id=srcchunk_111ef5d2b87a4b7a47b0f41c3c1c0621` `native_locator=slack:C0547N89JUB:1770085445.129679:1770085582.939589` `source_timestamp=2026-02-03T02:26:22Z`
- The fix involved disabling a likely redirect component, after which the correct routing worked when cache was cleared. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_0c10dd99f980c3f1b67909c00ef4abd4` `chunk_id=srcchunk_155fa2ebecb9ca4160c7d8eb4ae3e13a` `native_locator=slack:C0547N89JUB:1770085445.129679:1770102493.983209` `source_timestamp=2026-02-03T07:08:13Z`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_8bc009e2a484a41b12a7a0a6a1578874` `chunk_id=srcchunk_76d78e772213ad57a4e850cb2d711f06` `native_locator=slack:C0547N89JUB:1770085445.129679:1770102614.464139` `source_timestamp=2026-02-03T07:10:14Z`

## Open Questions

- What specific component or redirect rule caused the misconfiguration?

## Sources

- `source_document_id`: `srcdoc_ab7d535c8f2378de39f329dc8e775452`
- `source_revision_id`: `srcrev_0c10dd99f980c3f1b67909c00ef4abd4`
