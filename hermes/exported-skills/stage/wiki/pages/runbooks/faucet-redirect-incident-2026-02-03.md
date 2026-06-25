---
title: "Faucet Redirect Incident - 2026-02-03"
type: "runbook"
slug: "runbooks/faucet-redirect-incident-2026-02-03"
freshness: "2026-02-03T07:10:14Z"
tags:
  - "aeneid"
  - "faucet"
  - "incident"
  - "redirect"
owners:
  - "U07TNT9N4JC"
  - "U09M2SPUTSL"
source_revision_ids:
  - "srcrev_0c10dd99f980c3f1b67909c00ef4abd4"
  - "srcrev_827686def66cf421e1a4f4dee41b37d4"
  - "srcrev_8bc009e2a484a41b12a7a0a6a1578874"
conflict_state: "none"
---

# Faucet Redirect Incident - 2026-02-03

## Summary

On 2026-02-03, it was discovered that the URL https://faucet.story.foundation was incorrectly directing users to the GCP faucet instead of the intended Aeneid faucet at https://aeneid.faucet.story.foundation/. The issue was resolved by disabling a configuration that caused the incorrect routing, followed by clearing client-side caches.

## Claims

- https://faucet.story.foundation was pointing to GCP faucet instead of Aeneid faucet. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_827686def66cf421e1a4f4dee41b37d4` `chunk_id=srcchunk_43c67ba6a978a4e189892cc31a40b6da` `native_locator=slack:C0547N89JUB:1770085445.129679:1770085445.129679` `source_timestamp=2026-02-03T02:24:05Z`
- The issue was resolved by disabling a setting that caused the incorrect routing. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_0c10dd99f980c3f1b67909c00ef4abd4` `chunk_id=srcchunk_155fa2ebecb9ca4160c7d8eb4ae3e13a` `native_locator=slack:C0547N89JUB:1770085445.129679:1770102493.983209` `source_timestamp=2026-02-03T07:08:13Z`
- After the fix, the redirect worked, but users needed to clear their browser cache to see the change. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_8bc009e2a484a41b12a7a0a6a1578874` `chunk_id=srcchunk_76d78e772213ad57a4e850cb2d711f06` `native_locator=slack:C0547N89JUB:1770085445.129679:1770102614.464139` `source_timestamp=2026-02-03T07:10:14Z`

## Open Questions

- What was disabled to fix the redirect?

## Sources

- `source_document_id`: `srcdoc_ab7d535c8f2378de39f329dc8e775452`
- `source_revision_id`: `srcrev_e7036f5ac42ee9af3dc271eeea4f55ef`
