---
title: "Story Orchestration Service Error Catalog"
type: "runbook"
slug: "runbooks/story-orchestration-service-errors"
freshness: "2026-05-30T05:07:19Z"
tags:
  - "aggregation"
  - "database"
  - "error"
  - "story-orchestration-service"
owners: []
source_revision_ids:
  - "srcrev_078012d6bf004c9bf41a379c9cc01703"
  - "srcrev_2373a038a5b933db0de8fc000aa2770c"
  - "srcrev_3820b01d575299701a0894fe812898e6"
  - "srcrev_452228c6884cb4dc8e16b65ed20bd437"
  - "srcrev_48b7dd575a78fb52e2cebe455db25ad5"
  - "srcrev_5eae208e55c0a78d02ec6ec4d406553a"
  - "srcrev_8ca3cce31936672de2fc53830f92cdb7"
  - "srcrev_b5bbc6d2ca9a574316905f624b62d460"
  - "srcrev_db6e842a6f596165a1c372d971fd092a"
  - "srcrev_deae808adc1aeb6f5b0482b706dee7b7"
  - "srcrev_f7a1351902c991b994a1576b15628b90"
conflict_state: "none"
---

# Story Orchestration Service Error Catalog

## Summary

Collection of errors observed in the story-orchestration-service, including nil pointer dereferences, context cancellations, database constraint violations, and aggregation failures for various IP-related sub-services.

## Claims

- The story-orchestration-service encountered a runtime error: invalid memory address or nil pointer dereference. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d98173eca656bf5435ac44d48403fb0` `source_revision_id=srcrev_db6e842a6f596165a1c372d971fd092a` `chunk_id=srcchunk_67c6aecf38ded26757e7cbf86cead4ca` `native_locator=slack:C08BWTULNPP:1770903185.138749` `source_timestamp=2026-02-12T13:33:05Z`
- The story-orchestration-service encountered an aggregation error for nft_ownership. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d98173eca656bf5435ac44d48403fb0` `source_revision_id=srcrev_deae808adc1aeb6f5b0482b706dee7b7` `chunk_id=srcchunk_b022660efa21da63f5831bc96a4fef2d` `native_locator=slack:C08BWTULNPP:1771973672.136439` `source_timestamp=2026-02-24T22:54:32Z`
- The story-orchestration-service encountered an aggregation error for ip_licensing_enrichment. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d98173eca656bf5435ac44d48403fb0` `source_revision_id=srcrev_5eae208e55c0a78d02ec6ec4d406553a` `chunk_id=srcchunk_44437b64311f7de255379ee4f202201c` `native_locator=slack:C08BWTULNPP:1771976184.985769` `source_timestamp=2026-02-24T23:36:24Z`
- The story-orchestration-service encountered an aggregation error for ip_transactions. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d98173eca656bf5435ac44d48403fb0` `source_revision_id=srcrev_8ca3cce31936672de2fc53830f92cdb7` `chunk_id=srcchunk_1b71efe3238176201607b2fa327a6cab` `native_locator=slack:C08BWTULNPP:1771985122.941059` `source_timestamp=2026-02-25T02:05:22Z`
- The story-orchestration-service encountered a context canceled error. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d98173eca656bf5435ac44d48403fb0` `source_revision_id=srcrev_078012d6bf004c9bf41a379c9cc01703` `chunk_id=srcchunk_549ce15ef6f05e9d32d19aa80d0fd929` `native_locator=slack:C08BWTULNPP:1772180932.085359` `source_timestamp=2026-02-27T08:28:52Z`
- A message indicates expected errors will be replaced, suggesting a planned improvement. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d98173eca656bf5435ac44d48403fb0` `source_revision_id=srcrev_48b7dd575a78fb52e2cebe455db25ad5` `chunk_id=srcchunk_62429d0a3da8a1af61ba54802e84da09` `native_locator=slack:C08BWTULNPP:1772241055.836429` `source_timestamp=2026-02-28T01:10:55Z`
- The story-orchestration-service encountered an aggregation error for ip_infringement_processing. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d98173eca656bf5435ac44d48403fb0` `source_revision_id=srcrev_3820b01d575299701a0894fe812898e6` `chunk_id=srcchunk_f46e0a76a30f5de634ed46b2c98f01c1` `native_locator=slack:C08BWTULNPP:1772294659.456609` `source_timestamp=2026-02-28T16:04:19Z`
- The story-orchestration-service encountered a database error: null value in column "created_at" of relation "nodes" violates not-null constraint. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d98173eca656bf5435ac44d48403fb0` `source_revision_id=srcrev_452228c6884cb4dc8e16b65ed20bd437` `chunk_id=srcchunk_1e4a594de05cb04ff692e695f93cce40` `native_locator=slack:C08BWTULNPP:1772500292.953819` `source_timestamp=2026-03-03T01:11:32Z`
- The story-orchestration-service encountered a database error: null value in column "id" of relation "edges" violates not-null constraint. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d98173eca656bf5435ac44d48403fb0` `source_revision_id=srcrev_f7a1351902c991b994a1576b15628b90` `chunk_id=srcchunk_677039f44734ddc313b67c812817e42a` `native_locator=slack:C08BWTULNPP:1772501506.764919` `source_timestamp=2026-03-03T01:31:46Z`
- The story-orchestration-service encountered an aggregation error for ip_graph_aggregation. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d98173eca656bf5435ac44d48403fb0` `source_revision_id=srcrev_b5bbc6d2ca9a574316905f624b62d460` `chunk_id=srcchunk_7581b957581cc52a8a475a250ec57479` `native_locator=slack:C08BWTULNPP:1773873790.654569` `source_timestamp=2026-03-18T22:43:10Z`
- The story-orchestration-service encountered an aggregation error for ip_ownership. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_6d98173eca656bf5435ac44d48403fb0` `source_revision_id=srcrev_2373a038a5b933db0de8fc000aa2770c` `chunk_id=srcchunk_e92228eccc6d184560ff4401ec00afff` `native_locator=slack:C08BWTULNPP:1780117639.004389` `source_timestamp=2026-05-30T05:07:19Z`

## Open Questions

- What causes the nil pointer dereference in the story-orchestration-service?
- What is the planned replacement for expected errors?
- What is the root cause of the database not-null constraint violations (nodes.created_at, edges.id)?
- Why are aggregation errors occurring across multiple IP sub-services?

## Sources

- `source_document_id`: `srcdoc_6d98173eca656bf5435ac44d48403fb0`
- `source_revision_id`: `srcrev_078012d6bf004c9bf41a379c9cc01703`
