---
title: "Campaign Scheduling Feature (PR #431)"
type: "project"
slug: "projects/campaign-scheduling-feature-pr-431"
freshness: "2026-05-05T23:18:31Z"
tags:
  - "backend"
  - "campaign"
  - "pr"
  - "scheduling"
owners:
  - "U067QP5PD6J"
  - "U08V4SFU7LZ"
  - "U0927FP6HH9"
source_revision_ids:
  - "srcrev_0458cbd228243631cd4dc8a5bd1ddd8f"
  - "srcrev_826ef20d4ff51b0f65723c13644cce3b"
  - "srcrev_8d6c9b28a8f5839cf05cfd3041d40d4e"
  - "srcrev_9e921044b51a630e6b986b319be112c3"
  - "srcrev_9eaaab5318856c42f8a139f9a11749e3"
conflict_state: "none"
---

# Campaign Scheduling Feature (PR #431)

## Summary

PR #431 introduces the ability to schedule a campaign with a start date, after which it automatically appears on the App.

## Claims

- PR #431 allows creating a task with a start date; once the current time passes the start date, the campaign automatically shows up on the App. `claim:claim_4_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_826ef20d4ff51b0f65723c13644cce3b` `chunk_id=srcchunk_09ed64c7eb40d6d39094f6885b178479` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778022723.530409` `source_timestamp=2026-05-05T23:12:03Z`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_9e921044b51a630e6b986b319be112c3` `chunk_id=srcchunk_719b6f9a7cf1ab2683203462b77b0a38` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778022752.565549` `source_timestamp=2026-05-05T23:12:32Z`
- Potential race condition: if the campaign appears on the App before the on-chain collection is created, submissions might arrive before IP registration completes. `claim:claim_4_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_0458cbd228243631cd4dc8a5bd1ddd8f` `chunk_id=srcchunk_c7bb0b116560cffc6a80982f55cabfb4` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778023032.112119` `source_timestamp=2026-05-05T23:17:12Z`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_9eaaab5318856c42f8a139f9a11749e3` `chunk_id=srcchunk_a6ef643f88c8941e978b67f372217733` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778023051.935739` `source_timestamp=2026-05-05T23:17:46Z`
  - citation: `source_document_id=srcdoc_d15c947836e0a27e3d68a094607272ca` `source_revision_id=srcrev_8d6c9b28a8f5839cf05cfd3041d40d4e` `chunk_id=srcchunk_7878c58d2d27baf9f55a55d29f4d749f` `native_locator=slack:C0AL7EKNHDF:1777958270.714689:1778023111.604729` `source_timestamp=2026-05-05T23:18:31Z`

## Open Questions

- Handling of submissions before on-chain registration completes.
- How to mitigate race condition between campaign visibility and IP registration?

## Related Pages

- `ip-registration-auto-registration-issue`

## Sources

- `source_document_id`: `srcdoc_d15c947836e0a27e3d68a094607272ca`
- `source_revision_id`: `srcrev_eec3d3d009bf46a3831405890d417b81`
