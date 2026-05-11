---
title: "Public Testnet Explorer Service Outage"
type: "system"
slug: "systems/public-testnet-explorer-outage"
freshness: "2024-10-01T23:58:00Z"
tags:
  - "explorer"
  - "incident"
  - "post-mortem"
  - "testnet"
owners:
  - "Andy"
  - "Boris"
  - "Leeren"
  - "Ze"
source_revision_ids:
  - "srcrev_cf3d4aaff6a66d98abea7deec8423759"
conflict_state: "none"
---

# Public Testnet Explorer Service Outage

## Summary

The public-testnet blockchain explorer at https://testnet.storyscan.xyz/ went offline due to insufficient computational resources and a failure in docker-compose operations. The incident was resolved by scaling up the instance and upgrading EBS to io2. The outage lasted from 12:36 PM to 2:05 PM, with no impact on mainnet.

## Claims

- At 12:36 PM, the public-testnet blockchain explorer service at https://testnet.storyscan.xyz/ went offline. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-1) `source_document_id=srcdoc_6aa4e39fef46cdaf506a2100ff9cc758` `source_revision_id=srcrev_cf3d4aaff6a66d98abea7deec8423759` `chunk_id=srcchunk_99c1c93fb32cf97457c751158b415341` `native_locator=https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-1` `source_timestamp=2024-10-01T23:58:00Z`
- The typical command, docker-compose down, failed to bring all necessary services offline, and then the services could not be recovered by using docker-compose up -d. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-1) `source_document_id=srcdoc_6aa4e39fef46cdaf506a2100ff9cc758` `source_revision_id=srcrev_cf3d4aaff6a66d98abea7deec8423759` `chunk_id=srcchunk_99c1c93fb32cf97457c751158b415341` `native_locator=https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-1` `source_timestamp=2024-10-01T23:58:00Z`
- Root cause: insufficient computational resources; the instance lacked the resources needed to start all services simultaneously. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-1) `source_document_id=srcdoc_6aa4e39fef46cdaf506a2100ff9cc758` `source_revision_id=srcrev_cf3d4aaff6a66d98abea7deec8423759` `chunk_id=srcchunk_99c1c93fb32cf97457c751158b415341` `native_locator=https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-1` `source_timestamp=2024-10-01T23:58:00Z`
- Resolution included increasing server capacity and upgrading the EBS type to io2 (20,000 IOPS). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-1) `source_document_id=srcdoc_6aa4e39fef46cdaf506a2100ff9cc758` `source_revision_id=srcrev_cf3d4aaff6a66d98abea7deec8423759` `chunk_id=srcchunk_99c1c93fb32cf97457c751158b415341` `native_locator=https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-1` `source_timestamp=2024-10-01T23:58:00Z`
- Services were fully restored and the blockchain explorer was back online at 2:05 PM. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-1) `source_document_id=srcdoc_6aa4e39fef46cdaf506a2100ff9cc758` `source_revision_id=srcrev_cf3d4aaff6a66d98abea7deec8423759` `chunk_id=srcchunk_99c1c93fb32cf97457c751158b415341` `native_locator=https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-1` `source_timestamp=2024-10-01T23:58:00Z`
- The incident was isolated to the testnet explorer, with no impact on mainnet or production services. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-2) `source_document_id=srcdoc_6aa4e39fef46cdaf506a2100ff9cc758` `source_revision_id=srcrev_cf3d4aaff6a66d98abea7deec8423759` `chunk_id=srcchunk_aa572a208b52f28c871be054bc6907cf` `native_locator=https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-2` `source_timestamp=2024-10-01T23:58:00Z`
- Prevention steps include deeper analysis of core metrics scraped from the server using node-exporter and testing changes in the internal explorer before pushing to the public testnet explorer. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-2) `source_document_id=srcdoc_6aa4e39fef46cdaf506a2100ff9cc758` `source_revision_id=srcrev_cf3d4aaff6a66d98abea7deec8423759` `chunk_id=srcchunk_aa572a208b52f28c871be054bc6907cf` `native_locator=https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-2` `source_timestamp=2024-10-01T23:58:00Z`
- Incident response team: Boris, Ze, Leeren, and Andy were involved in the resolution. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-2) `source_document_id=srcdoc_6aa4e39fef46cdaf506a2100ff9cc758` `source_revision_id=srcrev_cf3d4aaff6a66d98abea7deec8423759` `chunk_id=srcchunk_aa572a208b52f28c871be054bc6907cf` `native_locator=https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-2` `source_timestamp=2024-10-01T23:58:00Z`
- Regular updates were provided to the internal team throughout the incident, with a Google meeting set up to track progress. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-2) `source_document_id=srcdoc_6aa4e39fef46cdaf506a2100ff9cc758` `source_revision_id=srcrev_cf3d4aaff6a66d98abea7deec8423759` `chunk_id=srcchunk_aa572a208b52f28c871be054bc6907cf` `native_locator=https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6#chunk-2` `source_timestamp=2024-10-01T23:58:00Z`

## Open Questions

- What specific metrics from node-exporter need to be analyzed to prevent future resource issues?

## Sources

- `source_document_id`: `srcdoc_6aa4e39fef46cdaf506a2100ff9cc758`
- `source_revision_id`: `srcrev_cf3d4aaff6a66d98abea7deec8423759`
- `source_url`: [Notion source](https://www.notion.so/Post-Mortem-public-testnet-Explorer-Service-Outage-112051299a548008ac9bc0caf43e1ae6)
