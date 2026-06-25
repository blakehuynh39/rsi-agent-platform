---
title: "Faucet Redirect to Aeneid Faucet"
type: "system"
slug: "systems/faucet-redirect-aeneid"
freshness: "2026-02-03T07:10:14Z"
tags:
  - "aeneid"
  - "deployment"
  - "faucet"
  - "redirect"
owners:
  - "@U07TNT9N4JC"
  - "@U09M2SPUTSL"
source_revision_ids:
  - "srcrev_0c10dd99f980c3f1b67909c00ef4abd4"
  - "srcrev_727e912a611b85c746c333ecc00aeecd"
  - "srcrev_793a4fc57879dfe63c9dddcad0e1e992"
  - "srcrev_827686def66cf421e1a4f4dee41b37d4"
  - "srcrev_8bc009e2a484a41b12a7a0a6a1578874"
  - "srcrev_e7036f5ac42ee9af3dc271eeea4f55ef"
conflict_state: "none"
---

# Faucet Redirect to Aeneid Faucet

## Summary

The faucet.story.foundation URL was incorrectly pointing to the GCP faucet instead of the Aeneid faucet. The issue was investigated and resolved after disabling a misconfiguration and clearing cache.

## Claims

- The URL faucet.story.foundation was pointing to the GCP faucet instead of the Aeneid faucet. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_827686def66cf421e1a4f4dee41b37d4` `chunk_id=srcchunk_43c67ba6a978a4e189892cc31a40b6da` `native_locator=slack:C0547N89JUB:1770085445.129679:1770085445.129679` `source_timestamp=2026-02-03T02:24:05Z`
- The deployment may have been overwritten, causing the incorrect redirect. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_e7036f5ac42ee9af3dc271eeea4f55ef` `chunk_id=srcchunk_111ef5d2b87a4b7a47b0f41c3c1c0621` `native_locator=slack:C0547N89JUB:1770085445.129679:1770085582.939589` `source_timestamp=2026-02-03T02:26:22Z`
- The issue persisted, and further investigation was requested. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_727e912a611b85c746c333ecc00aeecd` `chunk_id=srcchunk_e20ea05345d47688643f25505345f847` `native_locator=slack:C0547N89JUB:1770085445.129679:1770102058.092559` `source_timestamp=2026-02-03T07:00:58Z`
- The cause was identified as a disabled configuration that would be resolved shortly. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_0c10dd99f980c3f1b67909c00ef4abd4` `chunk_id=srcchunk_155fa2ebecb9ca4160c7d8eb4ae3e13a` `native_locator=slack:C0547N89JUB:1770085445.129679:1770102493.983209` `source_timestamp=2026-02-03T07:08:13Z`
- The fix was applied in a secret mode configuration. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_793a4fc57879dfe63c9dddcad0e1e992` `chunk_id=srcchunk_7213ae2759fbc2c6b34766478e57a278` `native_locator=slack:C0547N89JUB:1770085445.129679:1770102523.701989` `source_timestamp=2026-02-03T07:08:43Z`
- The redirect worked correctly after clearing the cache. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_ab7d535c8f2378de39f329dc8e775452` `source_revision_id=srcrev_8bc009e2a484a41b12a7a0a6a1578874` `chunk_id=srcchunk_76d78e772213ad57a4e850cb2d711f06` `native_locator=slack:C0547N89JUB:1770085445.129679:1770102614.464139` `source_timestamp=2026-02-03T07:10:14Z`

## Sources

- `source_document_id`: `srcdoc_ab7d535c8f2378de39f329dc8e775452`
- `source_revision_id`: `srcrev_827686def66cf421e1a4f4dee41b37d4`
