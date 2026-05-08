---
title: "Deprecate IPAssetRenderer"
type: "decision"
slug: "decisions/decision-deprecate-ipassetrenderer"
freshness: "2024-03-15T22:39:00Z"
tags:
  - "deprecation"
  - "IPAssetRenderer"
  - "smart-contracts"
owners:
  - "user://6e49a49b-0756-434a-b0ff-5e6c7e7bfe20"
  - "user://ce2077ef-5025-403e-b8e2-2d9d1c2c7bbd"
  - "user://d5afbb5c-e4fa-4d48-a970-a1716d0c2a6b"
source_revision_ids:
  - "srcrev_aaccfab29dc2f1313449c27418c49745"
conflict_state: "none"
---

# Deprecate IPAssetRenderer

## Summary

Decision to deprecate the IPAssetRenderer contract. IPAsset is not an NFT, so SVG rendering is unnecessary. JSON metadata generation can be consolidated into existing metadata-related contracts.

## Claims

- IPAssetRenderer is deprecated. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/7-Use-of-IPAssetRenderer-777ad0dcc4374ee7a31068a21d30cdf4) `source_document_id=srcdoc_29b2c9d7ef24599127c7a705e7586c1e` `source_revision_id=srcrev_aaccfab29dc2f1313449c27418c49745` `chunk_id=srcchunk_2c13f1858006e69fc16c3b54116b7e21` `native_locator=https://www.notion.so/7-Use-of-IPAssetRenderer-777ad0dcc4374ee7a31068a21d30cdf4` `source_timestamp=2024-03-15T22:39:00Z`
- IPAsset is not an NFT, so rendering an SVG is not needed. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/7-Use-of-IPAssetRenderer-777ad0dcc4374ee7a31068a21d30cdf4) `source_document_id=srcdoc_29b2c9d7ef24599127c7a705e7586c1e` `source_revision_id=srcrev_aaccfab29dc2f1313449c27418c49745` `chunk_id=srcchunk_2c13f1858006e69fc16c3b54116b7e21` `native_locator=https://www.notion.so/7-Use-of-IPAssetRenderer-777ad0dcc4374ee7a31068a21d30cdf4` `source_timestamp=2024-03-15T22:39:00Z`
- JSON metadata generation should be handled by existing metadata-related contracts instead of a separate IPAssetRenderer. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/7-Use-of-IPAssetRenderer-777ad0dcc4374ee7a31068a21d30cdf4) `source_document_id=srcdoc_29b2c9d7ef24599127c7a705e7586c1e` `source_revision_id=srcrev_aaccfab29dc2f1313449c27418c49745` `chunk_id=srcchunk_2c13f1858006e69fc16c3b54116b7e21` `native_locator=https://www.notion.so/7-Use-of-IPAssetRenderer-777ad0dcc4374ee7a31068a21d30cdf4` `source_timestamp=2024-03-15T22:39:00Z`

## Sources

- `source_document_id`: `srcdoc_29b2c9d7ef24599127c7a705e7586c1e`
- `source_revision_id`: `srcrev_aaccfab29dc2f1313449c27418c49745`
- `source_url`: [Notion source](https://www.notion.so/7-Use-of-IPAssetRenderer-777ad0dcc4374ee7a31068a21d30cdf4)
