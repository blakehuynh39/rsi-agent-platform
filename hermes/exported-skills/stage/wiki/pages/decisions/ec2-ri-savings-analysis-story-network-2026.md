---
title: "EC2 Reserved Instance Savings Analysis for Story Network Nodes (2026)"
type: "decision"
slug: "decisions/ec2-ri-savings-analysis-story-network-2026"
freshness: "2026-03-27T22:50:36Z"
tags:
  - "aws"
  - "cost-optimization"
  - "ec2"
  - "reserved-instances"
  - "story-network"
owners: []
source_revision_ids:
  - "srcrev_0a9e4c2b1719b8af3c08a53bb951687d"
  - "srcrev_897a2c04925dbfef7a3fe0fcb6e6e267"
  - "srcrev_b52e93d79e8889216cc93cc2783f350c"
  - "srcrev_dfe6e4136c1ee5af323fc28f17523488"
  - "srcrev_e3fa2f0ccac30471362b6cc06f61ae5c"
conflict_state: "none"
---

# EC2 Reserved Instance Savings Analysis for Story Network Nodes (2026)

## Summary

Analysis of converting on-demand EC2 instances to 1-year All Upfront Reserved Instances for RPC, validator, and bootnode nodes across story-mainnet and story-testnet accounts, showing a combined annual savings of $56,236 with $90,308 upfront cost and 7.4-month payback period.

## Claims

- Converting all identified RPC nodes (19 instances) from on-demand to 1-year All Upfront Reserved Instances yields annual savings of $44,651 with upfront cost of $71,612 across story-mainnet and story-testnet. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_eaa816b44900604d601cafc4776e22eb` `source_revision_id=srcrev_b52e93d79e8889216cc93cc2783f350c` `chunk_id=srcchunk_a106c1cedd74ea38e98fc728780c80c9` `native_locator=slack:C0547N89JUB:1774648429.287299:1774648900.775379` `source_timestamp=2026-03-27T22:01:40Z`
- Converting all identified validator and bootnode nodes (10 instances) yields annual savings of $11,585 with upfront cost of $18,690, all located in us-east-1 only. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_eaa816b44900604d601cafc4776e22eb` `source_revision_id=srcrev_0a9e4c2b1719b8af3c08a53bb951687d` `chunk_id=srcchunk_66d08c3c84579373d120a552ac8e335b` `native_locator=slack:C0547N89JUB:1774648429.287299:1774649460.343439` `source_timestamp=2026-03-27T22:11:00Z`
- Combining all node types (RPC, validator, bootnode) gives total annual savings of approximately $56,236 and total upfront cost of approximately $90,308, yielding a net savings of $56K/year. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_eaa816b44900604d601cafc4776e22eb` `source_revision_id=srcrev_897a2c04925dbfef7a3fe0fcb6e6e267` `chunk_id=srcchunk_d64ef1ca5e658da7ea80834863a6fbee` `native_locator=slack:C0547N89JUB:1774648429.287299:1774649927.784749` `source_timestamp=2026-03-27T22:18:47Z`
- The payback period for the RI purchases is approximately 7.4 months across all node types. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_eaa816b44900604d601cafc4776e22eb` `source_revision_id=srcrev_e3fa2f0ccac30471362b6cc06f61ae5c` `chunk_id=srcchunk_dff284f6f28f767c4860093927cc9de8` `native_locator=slack:C0547N89JUB:1774648429.287299:1774648587.687019` `source_timestamp=2026-03-27T21:56:27Z`
- Discount rates vary slightly: RPC nodes receive 38.4% (ranging 38.3%–39.0% by region), and validator/bootnode nodes receive a consistent 38.3%. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_eaa816b44900604d601cafc4776e22eb` `source_revision_id=srcrev_e3fa2f0ccac30471362b6cc06f61ae5c` `chunk_id=srcchunk_dff284f6f28f767c4860093927cc9de8` `native_locator=slack:C0547N89JUB:1774648429.287299:1774648587.687019` `source_timestamp=2026-03-27T21:56:27Z`
  - citation: `source_document_id=srcdoc_eaa816b44900604d601cafc4776e22eb` `source_revision_id=srcrev_0a9e4c2b1719b8af3c08a53bb951687d` `chunk_id=srcchunk_66d08c3c84579373d120a552ac8e335b` `native_locator=slack:C0547N89JUB:1774648429.287299:1774649460.343439` `source_timestamp=2026-03-27T22:11:00Z`
- No validator or bootnode nodes were found in eu-central-1 or ap-northeast-2; all such nodes are in us-east-1. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_eaa816b44900604d601cafc4776e22eb` `source_revision_id=srcrev_0a9e4c2b1719b8af3c08a53bb951687d` `chunk_id=srcchunk_66d08c3c84579373d120a552ac8e335b` `native_locator=slack:C0547N89JUB:1774648429.287299:1774649460.343439` `source_timestamp=2026-03-27T22:11:00Z`
- A GitHub task (#443) was created to track execution of the RI purchases. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_eaa816b44900604d601cafc4776e22eb` `source_revision_id=srcrev_dfe6e4136c1ee5af323fc28f17523488` `chunk_id=srcchunk_4f3b3fc0014c92df1706ededdc152b2e` `native_locator=slack:C0547N89JUB:1774648429.287299:1774651836.570059` `source_timestamp=2026-03-27T22:50:36Z`

## Related Pages

- `aws-credits-usage-2026`

## Sources

- `source_document_id`: `srcdoc_eaa816b44900604d601cafc4776e22eb`
- `source_revision_id`: `srcrev_dfe6e4136c1ee5af323fc28f17523488`
