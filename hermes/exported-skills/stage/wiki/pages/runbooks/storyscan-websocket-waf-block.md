---
title: "Storyscan WebSocket access blocked by Cloudflare WAF"
type: "runbook"
slug: "runbooks/storyscan-websocket-waf-block"
freshness: "2026-04-10T16:50:09Z"
tags:
  - "blockscout"
  - "cloudflare"
  - "storyscan"
  - "websocket"
owners: []
source_revision_ids:
  - "srcrev_1aa67ef38028768ee9dc09129552d659"
  - "srcrev_2cdba42ec97ebd8265a6495c90e2072d"
  - "srcrev_5989fdccba2e9251610a27c99c4c508a"
  - "srcrev_8db985308c061f1858266591a6f345ef"
conflict_state: "none"
---

# Storyscan WebSocket access blocked by Cloudflare WAF

## Summary

Investigation and resolution steps for when WebSocket connections to storyscan are blocked by Cloudflare WAF, affecting homepage block/transaction list loading.

## Claims

- WebSocket connections to storyscan (aeneid and mainnet) are blocked by Cloudflare WAF, returning HTTP 403 on /socket/websocket. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1de4cc738cd560ec3a6f515178d377b9` `source_revision_id=srcrev_8db985308c061f1858266591a6f345ef` `chunk_id=srcchunk_38e1fdad94e1410027c92c60a33f3b68` `native_locator=slack:C0547N89JUB:1775810814.554249:1775810814.554249` `source_timestamp=2026-04-10T08:46:54Z`
- Regular API works fine; the WebSocket block causes homepage blocks/transactions lists to not load. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1de4cc738cd560ec3a6f515178d377b9` `source_revision_id=srcrev_8db985308c061f1858266591a6f345ef` `chunk_id=srcchunk_38e1fdad94e1410027c92c60a33f3b68` `native_locator=slack:C0547N89JUB:1775810814.554249:1775810814.554249` `source_timestamp=2026-04-10T08:46:54Z`
- The affected origin IP is 129.227.79.246 (Zenlayer Singapore, AS21859). `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1de4cc738cd560ec3a6f515178d377b9` `source_revision_id=srcrev_8db985308c061f1858266591a6f345ef` `chunk_id=srcchunk_38e1fdad94e1410027c92c60a33f3b68` `native_locator=slack:C0547N89JUB:1775810814.554249:1775810814.554249` `source_timestamp=2026-04-10T08:46:54Z`
- Storyscan is a Blockscout-hosted instance; WAF/Cloudflare rules for storyscan domains are managed by the Blockscout team, not RSI infrastructure. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1de4cc738cd560ec3a6f515178d377b9` `source_revision_id=srcrev_1aa67ef38028768ee9dc09129552d659` `chunk_id=srcchunk_7d3d4dd69a614bb17dd146b5b31300c5` `native_locator=slack:C0547N89JUB:1775810814.554249:1775814231.683319` `source_timestamp=2026-04-10T09:43:51Z`
  - citation: `source_document_id=srcdoc_1de4cc738cd560ec3a6f515178d377b9` `source_revision_id=srcrev_5989fdccba2e9251610a27c99c4c508a` `chunk_id=srcchunk_854496072a7ed4a3b05cf3d3a70192d6` `native_locator=slack:C0547N89JUB:1775810814.554249:1775814247.433669` `source_timestamp=2026-04-10T09:44:07Z`
- To resolve the WebSocket block, the affected party (Yao) must raise the issue directly with the Blockscout team to allow WebSocket traffic for the reported IP range. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_1de4cc738cd560ec3a6f515178d377b9` `source_revision_id=srcrev_5989fdccba2e9251610a27c99c4c508a` `chunk_id=srcchunk_854496072a7ed4a3b05cf3d3a70192d6` `native_locator=slack:C0547N89JUB:1775810814.554249:1775814247.433669` `source_timestamp=2026-04-10T09:44:07Z`
  - citation: `source_document_id=srcdoc_1de4cc738cd560ec3a6f515178d377b9` `source_revision_id=srcrev_2cdba42ec97ebd8265a6495c90e2072d` `chunk_id=srcchunk_6d109e45a4353befeee2500743ab4c0f` `native_locator=slack:C0547N89JUB:1775810814.554249:1775839809.924989` `source_timestamp=2026-04-10T16:50:09Z`

## Open Questions

- Has the Blockscout team been contacted, and is the WebSocket access now allowed from the affected IP?

## Sources

- `source_document_id`: `srcdoc_1de4cc738cd560ec3a6f515178d377b9`
- `source_revision_id`: `srcrev_2cdba42ec97ebd8265a6495c90e2072d`
