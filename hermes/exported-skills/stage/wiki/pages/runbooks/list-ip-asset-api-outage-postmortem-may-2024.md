---
title: "List IP Asset API Outage Postmortem (May 2024)"
type: "runbook"
slug: "runbooks/list-ip-asset-api-outage-postmortem-may-2024"
freshness: "2026-05-05T06:38:42Z"
tags:
  - "api"
  - "data-ingestion"
  - "incident"
  - "postmortem"
owners: []
source_revision_ids:
  - "srcrev_84e115e71511abaf9c4bf51d0ee5fc1d"
conflict_state: "none"
---

# List IP Asset API Outage Postmortem (May 2024)

## Summary

Postmortem for the List IP Asset API not returning the latest data on May 22-23, 2024. The root cause was data ingestion being turned off, resolved by Ruimin. Delays in response were due to lack of urgency, no formal oncall, and part-time offshore engineers. Improvements include e2e tests, formal oncall process, hiring a US backend engineer, and using an additional indexer.

## Claims

- The List IP Asset API stopped returning the latest data around Wednesday, May 22, 2024 9:05 PM. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-List-IP-Asset-API-didn-t-return-the-latest-data-a0c0220c5abb47eb8281948f443d81ca#chunk-1) `source_document_id=srcdoc_c417157a743192eb8d2b9a3ce3bfc18a` `source_revision_id=srcrev_84e115e71511abaf9c4bf51d0ee5fc1d` `chunk_id=srcchunk_7d3bca86515ce557fffeff8877cfdec0` `native_locator=https://www.notion.so/Postmortem-List-IP-Asset-API-didn-t-return-the-latest-data-a0c0220c5abb47eb8281948f443d81ca#chunk-1` `source_timestamp=2026-05-05T06:38:42Z`
- The issue was first reported by a user in the Discord dev-chat channel around Thursday, May 23, 5:01 PM. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-List-IP-Asset-API-didn-t-return-the-latest-data-a0c0220c5abb47eb8281948f443d81ca#chunk-1) `source_document_id=srcdoc_c417157a743192eb8d2b9a3ce3bfc18a` `source_revision_id=srcrev_84e115e71511abaf9c4bf51d0ee5fc1d` `chunk_id=srcchunk_7d3bca86515ce557fffeff8877cfdec0` `native_locator=https://www.notion.so/Postmortem-List-IP-Asset-API-didn-t-return-the-latest-data-a0c0220c5abb47eb8281948f443d81ca#chunk-1` `source_timestamp=2026-05-05T06:38:42Z`
- QA test by 57 blocks detected the issue around Thursday, May 23, 7:36 PM. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-List-IP-Asset-API-didn-t-return-the-latest-data-a0c0220c5abb47eb8281948f443d81ca#chunk-1) `source_document_id=srcdoc_c417157a743192eb8d2b9a3ce3bfc18a` `source_revision_id=srcrev_84e115e71511abaf9c4bf51d0ee5fc1d` `chunk_id=srcchunk_7d3bca86515ce557fffeff8877cfdec0` `native_locator=https://www.notion.so/Postmortem-List-IP-Asset-API-didn-t-return-the-latest-data-a0c0220c5abb47eb8281948f443d81ca#chunk-1` `source_timestamp=2026-05-05T06:38:42Z`
- Ze and Ruimin had a call and figured out the issue; Ruimin resolved it by turning on data ingestion. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-List-IP-Asset-API-didn-t-return-the-latest-data-a0c0220c5abb47eb8281948f443d81ca#chunk-2) `source_document_id=srcdoc_c417157a743192eb8d2b9a3ce3bfc18a` `source_revision_id=srcrev_84e115e71511abaf9c4bf51d0ee5fc1d` `chunk_id=srcchunk_19f2d3c7cbb3b851efb46bde8b368271` `native_locator=https://www.notion.so/Postmortem-List-IP-Asset-API-didn-t-return-the-latest-data-a0c0220c5abb47eb8281948f443d81ca#chunk-2` `source_timestamp=2026-05-05T06:38:42Z`
- The time from first reporting to starting resolution took too long because reports were deemed not reproducible, Slack messages lacked urgency, platform team was focused on L1 development, and API tasks were handled by part-time engineers in China who were not notified. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-List-IP-Asset-API-didn-t-return-the-latest-data-a0c0220c5abb47eb8281948f443d81ca#chunk-2) `source_document_id=srcdoc_c417157a743192eb8d2b9a3ce3bfc18a` `source_revision_id=srcrev_84e115e71511abaf9c4bf51d0ee5fc1d` `chunk_id=srcchunk_19f2d3c7cbb3b851efb46bde8b368271` `native_locator=https://www.notion.so/Postmortem-List-IP-Asset-API-didn-t-return-the-latest-data-a0c0220c5abb47eb8281948f443d81ca#chunk-2` `source_timestamp=2026-05-05T06:38:42Z`
- Improvements proposed: add e2e auto test, setup formal oncall process with phone numbers, hire full-time backend engineer in US (Blake), start working with another indexer Gold Sky. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Postmortem-List-IP-Asset-API-didn-t-return-the-latest-data-a0c0220c5abb47eb8281948f443d81ca#chunk-2) `source_document_id=srcdoc_c417157a743192eb8d2b9a3ce3bfc18a` `source_revision_id=srcrev_84e115e71511abaf9c4bf51d0ee5fc1d` `chunk_id=srcchunk_19f2d3c7cbb3b851efb46bde8b368271` `native_locator=https://www.notion.so/Postmortem-List-IP-Asset-API-didn-t-return-the-latest-data-a0c0220c5abb47eb8281948f443d81ca#chunk-2` `source_timestamp=2026-05-05T06:38:42Z`

## Sources

- `source_document_id`: `srcdoc_c417157a743192eb8d2b9a3ce3bfc18a`
- `source_revision_id`: `srcrev_84e115e71511abaf9c4bf51d0ee5fc1d`
- `source_url`: [Notion source](https://www.notion.so/Postmortem-List-IP-Asset-API-didn-t-return-the-latest-data-a0c0220c5abb47eb8281948f443d81ca)
