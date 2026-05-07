---
title: "Node Recovery Process"
type: "runbook"
slug: "runbooks/node-recovery-process"
freshness: "2024-07-24T20:31:00Z"
tags:
  - "alerting"
  - "monitoring"
  - "node"
  - "outage"
  - "recovery"
owners: []
source_revision_ids:
  - "srcrev_5fd31494a142d8e46161d1f0b2365116"
conflict_state: "none"
---

# Node Recovery Process

## Summary

Defines the process for discovering, mitigating, and root-causing node outages. Covers alerting via Grafana with Critical and High severity levels, initial alert types, mitigation strategies including runbooks, and root cause analysis.

## Claims

- Alerts will be set in Grafana based on certain metrics and sent to Slack or call the on-call person depending on severity. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67) `source_document_id=srcdoc_dc71b11618d582d5ff55e48568cffa5c` `source_revision_id=srcrev_5fd31494a142d8e46161d1f0b2365116` `chunk_id=srcchunk_bb5b283d86b2c05a7d6b124426e6428f` `native_locator=https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67` `source_timestamp=2024-07-24T20:31:00Z`
- Two severity levels are used: Critical (calls on-call and sends Slack alerts) and High (sends Slack alerts). Both must be actionable. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67) `source_document_id=srcdoc_dc71b11618d582d5ff55e48568cffa5c` `source_revision_id=srcrev_5fd31494a142d8e46161d1f0b2365116` `chunk_id=srcchunk_bb5b283d86b2c05a7d6b124426e6428f` `native_locator=https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67` `source_timestamp=2024-07-24T20:31:00Z`
- Initial alert types include: Network halt (block height not increasing in X seconds, Critical), Hardware resource over-utilization (CPU/memory/disk > X%, Critical/High), Node crash (health check, Critical), and Network fork (needs research). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67) `source_document_id=srcdoc_dc71b11618d582d5ff55e48568cffa5c` `source_revision_id=srcrev_5fd31494a142d8e46161d1f0b2365116` `chunk_id=srcchunk_bb5b283d86b2c05a7d6b124426e6428f` `native_locator=https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67` `source_timestamp=2024-07-24T20:31:00Z`
- Monitor screens in the Palo Alto office should display current network stats for daily familiarity, not just during outages or upgrades. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67) `source_document_id=srcdoc_dc71b11618d582d5ff55e48568cffa5c` `source_revision_id=srcrev_5fd31494a142d8e46161d1f0b2365116` `chunk_id=srcchunk_bb5b283d86b2c05a7d6b124426e6428f` `native_locator=https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67` `source_timestamp=2024-07-24T20:31:00Z`
- Outage mitigation may involve quick actions like falling back to a previous deployment or capturing logs and restarting a crashed/halted node before root cause analysis. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67) `source_document_id=srcdoc_dc71b11618d582d5ff55e48568cffa5c` `source_revision_id=srcrev_5fd31494a142d8e46161d1f0b2365116` `chunk_id=srcchunk_bb5b283d86b2c05a7d6b124426e6428f` `native_locator=https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67` `source_timestamp=2024-07-24T20:31:00Z`
- Runbooks will be created to capture mitigation and troubleshooting steps for specific alerts, initially based on hypotheses, tests, simulations, and learnings from other chains. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67) `source_document_id=srcdoc_dc71b11618d582d5ff55e48568cffa5c` `source_revision_id=srcrev_5fd31494a142d8e46161d1f0b2365116` `chunk_id=srcchunk_bb5b283d86b2c05a7d6b124426e6428f` `native_locator=https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67` `source_timestamp=2024-07-24T20:31:00Z`
- Common root causes of outages include introducing new code/configs (especially upgrades), running out of hardware resources, and existing bugs triggered in rare scenarios. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67) `source_document_id=srcdoc_dc71b11618d582d5ff55e48568cffa5c` `source_revision_id=srcrev_5fd31494a142d8e46161d1f0b2365116` `chunk_id=srcchunk_bb5b283d86b2c05a7d6b124426e6428f` `native_locator=https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67` `source_timestamp=2024-07-24T20:31:00Z`
- Grafana metrics dashboard collects node hardware-level and application-level metrics to help identify resource over-utilization, shutdowns, and software liveness. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67) `source_document_id=srcdoc_dc71b11618d582d5ff55e48568cffa5c` `source_revision_id=srcrev_5fd31494a142d8e46161d1f0b2365116` `chunk_id=srcchunk_bb5b283d86b2c05a7d6b124426e6428f` `native_locator=https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67` `source_timestamp=2024-07-24T20:31:00Z`

## Open Questions

- Can a network fork occur, and how should it be detected and alerted on?
- What are the exact runbook steps for each alert type?
- What specific X thresholds should be set for block height increase, CPU, memory, and disk utilization alerts?

## Sources

- `source_document_id`: `srcdoc_dc71b11618d582d5ff55e48568cffa5c`
- `source_revision_id`: `srcrev_5fd31494a142d8e46161d1f0b2365116`
- `source_url`: [Notion source](https://www.notion.so/Node-recovery-process-407db5499dec414982fb68b92cf9cb67)
