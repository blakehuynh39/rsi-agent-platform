---
title: "Remove storage-node from Poseidon MVP"
type: "decision"
slug: "decisions/remove-storage-node-from-poseidon-mvp"
freshness: "2026-01-29T00:48:02Z"
tags:
  - "decommissioning"
  - "mvp"
  - "poseidon"
  - "servers"
  - "storage"
owners:
  - "U04KTUN5WFQ"
  - "U07A7AUGL5V"
  - "U07TNT9N4JC"
  - "U09M2SPUTSL"
  - "U0A3GPWELDP"
source_revision_ids:
  - "srcrev_0abf426249dfcdce6bd071b62ead643d"
  - "srcrev_335c02070a40bd3bc7936d07c7c50c0f"
  - "srcrev_464363522cdcce75befaafbeb9f49691"
  - "srcrev_fe40006dce593d477bd51749ff898033"
conflict_state: "none"
---

# Remove storage-node from Poseidon MVP

## Summary

Decision to remove the storage-node server from the Poseidon MVP environment, deemed safe, to be executed at 19:00 PT on 2026-01-29.

## Claims

- A request was made to remove servers in Poseidon MVP. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a18cc883b1097027e3c8fb42306abbc8` `source_revision_id=srcrev_335c02070a40bd3bc7936d07c7c50c0f` `chunk_id=srcchunk_826f5f2d506a0d0e2dd63901f9597177` `native_locator=slack:C0547N89JUB:1769647476.851109:1769647476.851109` `source_timestamp=2026-01-29T00:46:55Z`
- It was deemed safe to remove the servers. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a18cc883b1097027e3c8fb42306abbc8` `source_revision_id=srcrev_464363522cdcce75befaafbeb9f49691` `chunk_id=srcchunk_851e61c0ba3a810c7409e4c09bd4c0ad` `native_locator=slack:C0547N89JUB:1769647476.851109:1769647537.997239` `source_timestamp=2026-01-29T00:45:37Z`
- The storage-node server is among the servers to be removed from Poseidon MVP. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a18cc883b1097027e3c8fb42306abbc8` `source_revision_id=srcrev_0abf426249dfcdce6bd071b62ead643d` `chunk_id=srcchunk_0f9a996399624da2abf7320769d6bc57` `native_locator=slack:C0547N89JUB:1769647476.851109:1769647600.128969` `source_timestamp=2026-01-29T00:46:40Z`
- The removal is scheduled for 19:00 PT on 2026-01-29. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a18cc883b1097027e3c8fb42306abbc8` `source_revision_id=srcrev_0abf426249dfcdce6bd071b62ead643d` `chunk_id=srcchunk_0f9a996399624da2abf7320769d6bc57` `native_locator=slack:C0547N89JUB:1769647476.851109:1769647600.128969` `source_timestamp=2026-01-29T00:46:40Z`
- Jinn acknowledged the removal plan. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a18cc883b1097027e3c8fb42306abbc8` `source_revision_id=srcrev_fe40006dce593d477bd51749ff898033` `chunk_id=srcchunk_5209e703a38a75fa30804b243f0de934` `native_locator=slack:C0547N89JUB:1769647476.851109:1769647682.242179` `source_timestamp=2026-01-29T00:48:02Z`

## Open Questions

- What was the complete list of servers to be removed? The message 'below servers' omitted the list.

## Related Pages

- `poseidon-mvp`

## Sources

- `source_document_id`: `srcdoc_a18cc883b1097027e3c8fb42306abbc8`
- `source_revision_id`: `srcrev_335c02070a40bd3bc7936d07c7c50c0f`
