---
title: "TaxBandits W-9/W-8 Integration"
type: "project"
slug: "projects/taxbandits-w9-w8-integration"
freshness: "2026-06-03T22:33:57Z"
tags:
  - "engineering"
  - "integration"
  - "tax"
owners:
  - "U04L0DD6B6F"
  - "U05A515NBFC"
  - "U06A5AQ1VD3"
  - "U06SA1M3RGE"
  - "U083MMT1771"
  - "U086FECSTP1"
  - "U08JPHTRTLN"
  - "U09QGMMUDPC"
source_revision_ids:
  - "srcrev_1f4777e50bc5e1ca1b9faecaf7f303f9"
  - "srcrev_3db1636d053027e5bade9177a34394d3"
  - "srcrev_41f9a895a29c351b3a02a62c0c870b28"
  - "srcrev_76b7693903b94fe559405c85a4189297"
  - "srcrev_9309f0001100d94f685267d33371b7f7"
  - "srcrev_b849fdf43febe8df5ebb6ceb28a4bd3d"
conflict_state: "none"
---

# TaxBandits W-9/W-8 Integration

## Summary

Integration of TaxBandits W-9/W-8 form drop-in for tax form collection.

## Claims

- The TaxBandits W-9/W-8 form drop-in was tested end-to-end in sandbox and is working. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_183ec13213b3b50403af64fc4dcc0b19` `source_revision_id=srcrev_1f4777e50bc5e1ca1b9faecaf7f303f9` `chunk_id=srcchunk_378b33aec88fe6e010296f19ad6b690e` `native_locator=slack:C0AL7EKNHDF:1780389743.764059:1780389743.764059` `source_timestamp=2026-06-02T08:42:23Z`
- An engineering handoff zip file (taxbandits-dropin-handoff.zip) was provided with an ENGINEERING-HANDOFF.md guide. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_183ec13213b3b50403af64fc4dcc0b19` `source_revision_id=srcrev_1f4777e50bc5e1ca1b9faecaf7f303f9` `chunk_id=srcchunk_378b33aec88fe6e010296f19ad6b690e` `native_locator=slack:C0AL7EKNHDF:1780389743.764059:1780389743.764059` `source_timestamp=2026-06-02T08:42:23Z`
- The iframe gives no reliable submit signal; completion must be detected via Status API or a webhook, server-side. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_183ec13213b3b50403af64fc4dcc0b19` `source_revision_id=srcrev_1f4777e50bc5e1ca1b9faecaf7f303f9` `chunk_id=srcchunk_378b33aec88fe6e010296f19ad6b690e` `native_locator=slack:C0AL7EKNHDF:1780389743.764059:1780389743.764059` `source_timestamp=2026-06-02T08:42:23Z`
- Origin allowlist requires HTTPS origins with a trailing slash; localhost http silently fails, necessitating an HTTPS tunnel for local testing. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_183ec13213b3b50403af64fc4dcc0b19` `source_revision_id=srcrev_1f4777e50bc5e1ca1b9faecaf7f303f9` `chunk_id=srcchunk_378b33aec88fe6e010296f19ad6b690e` `native_locator=slack:C0AL7EKNHDF:1780389743.764059:1780389743.764059` `source_timestamp=2026-06-02T08:42:23Z`
- Sandbox credentials are included inline; they must be rotated to production credentials before go-live, with Client Secret and User Token kept server-side only. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_183ec13213b3b50403af64fc4dcc0b19` `source_revision_id=srcrev_1f4777e50bc5e1ca1b9faecaf7f303f9` `chunk_id=srcchunk_378b33aec88fe6e010296f19ad6b690e` `native_locator=slack:C0AL7EKNHDF:1780389743.764059:1780389743.764059` `source_timestamp=2026-06-02T08:42:23Z`
- An integration plan exists at https://github.com/piplabs/numo-monorepo/blob/feat/taxbandits-w8ben-collection/docs/plans/active/2026-06-03-taxbandits-w8ben-collection.md. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_183ec13213b3b50403af64fc4dcc0b19` `source_revision_id=srcrev_76b7693903b94fe559405c85a4189297` `chunk_id=srcchunk_c00c6d9570685d7826b7a26878106b9f` `native_locator=slack:C0AL7EKNHDF:1780389743.764059:1780420043.547459` `source_timestamp=2026-06-02T17:07:23Z`
- A reviewer noted that the plan does not include the withholding rate from W-8 in payment calculations before paying out the user. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_183ec13213b3b50403af64fc4dcc0b19` `source_revision_id=srcrev_3db1636d053027e5bade9177a34394d3` `chunk_id=srcchunk_7b03ce054733e48a5f20a3aa8571ef02` `native_locator=slack:C0AL7EKNHDF:1780389743.764059:1780425052.665949` `source_timestamp=2026-06-02T18:30:52Z`
- It was proposed that the system must honor the W-8 form selections even if users make mistakes on treaty benefits, potentially reducing some payments. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_183ec13213b3b50403af64fc4dcc0b19` `source_revision_id=srcrev_b849fdf43febe8df5ebb6ceb28a4bd3d` `chunk_id=srcchunk_c1d88e034fb1c8db56aac630f0a1a4ca` `native_locator=slack:C0AL7EKNHDF:1780389743.764059:1780425480.206899` `source_timestamp=2026-06-02T21:15:02Z`
- TaxBandits API has scheduled maintenance on Mondays from 1 AM to 3 AM EST, requiring code to handle multi-hour downtimes. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_183ec13213b3b50403af64fc4dcc0b19` `source_revision_id=srcrev_9309f0001100d94f685267d33371b7f7` `chunk_id=srcchunk_c6d8e594edde328d75ee09f147423581` `native_locator=slack:C0AL7EKNHDF:1780389743.764059:1780525344.717899` `source_timestamp=2026-06-03T22:22:24Z`
- TaxBandits was provided the email ap@psdn.ai for any tax matter forwards. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_183ec13213b3b50403af64fc4dcc0b19` `source_revision_id=srcrev_41f9a895a29c351b3a02a62c0c870b28` `chunk_id=srcchunk_e9ccb2526c225c5767c3e25742ea5535` `native_locator=slack:C0AL7EKNHDF:1780389743.764059:1780526037.023539` `source_timestamp=2026-06-03T22:33:57Z`

## Open Questions

- How should the withholding rate from W-8 be integrated into payment calculations?
- How will the system handle the scheduled API downtime (Mondays 1 AM - 3 AM EST)?
- Should the system always honor W-8 form selections even if users make mistakes, potentially resulting in higher withholding?

## Sources

- `source_document_id`: `srcdoc_183ec13213b3b50403af64fc4dcc0b19`
- `source_revision_id`: `srcrev_fa0171b2873e60c690d6cd8746e101ca`
