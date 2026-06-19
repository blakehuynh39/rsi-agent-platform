---
title: "Fraud Detection and Country-Based Access Control Strategy"
type: "decision"
slug: "decisions/fraud-detection-and-country-based-access-control-strategy"
freshness: "2026-05-04T21:17:31Z"
tags:
  - "country-restrictions"
  - "fraud"
  - "geolocation"
  - "shadow-ban"
  - "spam"
  - "voice-campaigns"
owners: []
source_revision_ids:
  - "srcrev_1d703eba4008370be95208f699d30874"
  - "srcrev_3c23dc020bc79c154e934ce52243839e"
  - "srcrev_68eb6f96d9521b013a415603cbb30b1d"
  - "srcrev_7c8257f121087c72b4f5717e8a852687"
  - "srcrev_8fbf6e0d277f5d4908bfb5730cc90bbf"
  - "srcrev_99ecca78dc430c660105221d16041108"
  - "srcrev_e32c23b176623a128655844b9e825039"
conflict_state: "none"
---

# Fraud Detection and Country-Based Access Control Strategy

## Summary

Discussion on implementing fraud detection and country-based access controls to prevent spam and farming in data collection campaigns. Topics include IP banning, shadow banning, country whitelisting for voice campaigns, and cross-checking user country, IP, and language.

## Claims

- IP banning was considered as a method to combat farming/spam, with plans for an announcement backed by data. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_93d6da8ebd97c57f0690e397ffcd48ae` `source_revision_id=srcrev_1d703eba4008370be95208f699d30874` `chunk_id=srcchunk_7198c6db7b21e3895e6ee7df2bf58d78` `native_locator=slack:C0AL7EKNHDF:1777924305.712219:1777926888.700219` `source_timestamp=2026-05-04T20:34:58Z`
  - citation: `source_document_id=srcdoc_93d6da8ebd97c57f0690e397ffcd48ae` `source_revision_id=srcrev_7c8257f121087c72b4f5717e8a852687` `chunk_id=srcchunk_d807eb3777c289b114c447368e5f7002` `native_locator=slack:C0AL7EKNHDF:1777924305.712219:1777927028.314029` `source_timestamp=2026-05-04T20:37:08Z`
- For voice campaigns, restricting submissions to specific countries per campaign was suggested to improve data quality over quantity. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_93d6da8ebd97c57f0690e397ffcd48ae` `source_revision_id=srcrev_68eb6f96d9521b013a415603cbb30b1d` `chunk_id=srcchunk_dd555f73c2516dbec56512043674de77` `native_locator=slack:C0AL7EKNHDF:1777924305.712219:1777929011.759189` `source_timestamp=2026-05-04T21:10:11Z`
- Indic dialect campaigns should accept from India and low-wage countries with high Indian populations; Indonesian campaigns from Indonesia; Malaysian campaigns from Malaysia. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_93d6da8ebd97c57f0690e397ffcd48ae` `source_revision_id=srcrev_68eb6f96d9521b013a415603cbb30b1d` `chunk_id=srcchunk_dd555f73c2516dbec56512043674de77` `native_locator=slack:C0AL7EKNHDF:1777924305.712219:1777929011.759189` `source_timestamp=2026-05-04T21:10:11Z`
- A one-click request button to open up new countries per campaign was suggested. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_93d6da8ebd97c57f0690e397ffcd48ae` `source_revision_id=srcrev_68eb6f96d9521b013a415603cbb30b1d` `chunk_id=srcchunk_dd555f73c2516dbec56512043674de77` `native_locator=slack:C0AL7EKNHDF:1777924305.712219:1777929011.759189` `source_timestamp=2026-05-04T21:10:11Z`
- Silently shadow banning users from high-spam countries (e.g., Nigeria) was proposed: accept submissions but discard in backend and flag accounts as 'pending account review' to prevent payouts, without public announcement. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_93d6da8ebd97c57f0690e397ffcd48ae` `source_revision_id=srcrev_3c23dc020bc79c154e934ce52243839e` `chunk_id=srcchunk_73b6e614f722b8f288ecd3a3a1627a98` `native_locator=slack:C0AL7EKNHDF:1777924305.712219:1777927218.377239` `source_timestamp=2026-05-04T20:40:18Z`
- An alternative proposal is to internally flag users and block withdrawal when payments are live. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_93d6da8ebd97c57f0690e397ffcd48ae` `source_revision_id=srcrev_8fbf6e0d277f5d4908bfb5730cc90bbf` `chunk_id=srcchunk_dc5ce76e7cf9ab1a9eca0a66ba7e3030` `native_locator=slack:C0AL7EKNHDF:1777924305.712219:1777927443.129599` `source_timestamp=2026-05-04T20:44:03Z`
- A country:language map was being built in the backend (GitHub PR #412) to support fraud detection. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_93d6da8ebd97c57f0690e397ffcd48ae` `source_revision_id=srcrev_e32c23b176623a128655844b9e825039` `chunk_id=srcchunk_4d1f254c9dceb9a2c824bdacd3ee781b` `native_locator=slack:C0AL7EKNHDF:1777924305.712219:1777929347.745269` `source_timestamp=2026-05-04T21:15:47Z`
- Fraud detection will cross-check user-selected country, IP address origin country, and language; any mismatch is a red flag. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_93d6da8ebd97c57f0690e397ffcd48ae` `source_revision_id=srcrev_99ecca78dc430c660105221d16041108` `chunk_id=srcchunk_9c5cb1414571e04644a711c39948fb45` `native_locator=slack:C0AL7EKNHDF:1777924305.712219:1777929443.873439` `source_timestamp=2026-05-04T21:17:31Z`

## Open Questions

- Feasibility of IP banning from a technical perspective
- How to handle campaigns without geographic association (e.g., videos)
- Public announcement of IP bans or shadow bans

## Sources

- `source_document_id`: `srcdoc_93d6da8ebd97c57f0690e397ffcd48ae`
- `source_revision_id`: `srcrev_99ecca78dc430c660105221d16041108`
