---
title: "Numo Campaign Health Dashboard"
type: "system"
slug: "systems/numo-campaign-health-dashboard"
freshness: "2026-05-06T17:01:47Z"
tags:
  - "campaign-health"
  - "github-actions"
  - "numo"
owners: []
source_revision_ids:
  - "srcrev_22157a60e89311c4f59c0ab9444251ee"
  - "srcrev_87429680e949dfeca474d20f4ba8e77f"
  - "srcrev_a70daafb6b1dc9fd837d2fc2f783b9ba"
conflict_state: "none"
---

# Numo Campaign Health Dashboard

## Summary

The Numo Campaign Health dashboard provides morning digest statistics and campaign performance data. A request is pending to expand displayed campaigns from top 3 to top 5 or all campaigns.

## Claims

- The Numo morning digest on 2026-05-06 reported 39,981 submissions in the last 24 hours and 171,986 pending review, with $0.72 validated and $20,684.16 unvalidated. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_78439a9f6296b8d2e6aceb7bbce61b63` `source_revision_id=srcrev_a70daafb6b1dc9fd837d2fc2f783b9ba` `chunk_id=srcchunk_97c9d45f78266e74a4f5075e00b03d46` `native_locator=slack:C0AL7EKNHDF:1778085069.709049:1778085069.709049` `source_timestamp=2026-05-06T16:31:09Z`
- A team member requested that the Campaign Health data be updated to include the top 5 campaigns, because the 4th performing campaign (Tamil) is not showing data currently. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_78439a9f6296b8d2e6aceb7bbce61b63` `source_revision_id=srcrev_87429680e949dfeca474d20f4ba8e77f` `chunk_id=srcchunk_5d4686b6f3b9ee8d4750c377a6fa5edd` `native_locator=slack:C0AL7EKNHDF:1778085069.709049:1778085801.544369` `source_timestamp=2026-05-06T16:43:21Z`
- A follow-up request asked to modify the Numo Digest GitHub action to include all campaigns instead of just the top 3. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_78439a9f6296b8d2e6aceb7bbce61b63` `source_revision_id=srcrev_22157a60e89311c4f59c0ab9444251ee` `chunk_id=srcchunk_7cbee62220612cb20158b1de61d2f624` `native_locator=slack:C0AL7EKNHDF:1778085069.709049:1778086897.360229` `source_timestamp=2026-05-06T17:01:47Z`

## Open Questions

- Should the Campaign Health display be changed to top 5 or all campaigns? Who will implement the GitHub action change?

## Sources

- `source_document_id`: `srcdoc_78439a9f6296b8d2e6aceb7bbce61b63`
- `source_revision_id`: `srcrev_22157a60e89311c4f59c0ab9444251ee`
