---
title: "L1 Tech Framework Evaluation"
type: "decision"
slug: "decisions/l1-tech-framework-evaluation"
freshness: "2024-05-04T07:20:00Z"
tags:
  - "blockchain"
  - "cosmos"
  - "evaluation"
  - "framework"
  - "layer-1"
  - "meter"
owners: []
source_revision_ids:
  - "srcrev_6110bfffb6e4c826bdf902daeb084d21"
conflict_state: "none"
---

# L1 Tech Framework Evaluation

## Summary

Evaluation of Layer 1 blockchain frameworks including EVMOS, Berachain, and Meter.io for potential adoption, comparing metrics such as extensibility, consensus, stability, and ecosystem.

## Claims

- The primary advantage of selecting meter.io as our framework is the potential to drive innovation in the Layer 1 space, using the Hotstuff 2 consensus algorithm. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d) `source_document_id=srcdoc_abc776d673caae53ff16c531d6248f0f` `source_revision_id=srcrev_6110bfffb6e4c826bdf902daeb084d21` `chunk_id=srcchunk_47f469100a9220f590437a80cd8668c2` `native_locator=https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d` `source_timestamp=2024-05-04T07:20:00Z`
- The main challenge in utilizing the meter.io framework is its lack of extensibility, with a monolithic codebase not designed for modular expansion. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d) `source_document_id=srcdoc_abc776d673caae53ff16c531d6248f0f` `source_revision_id=srcrev_6110bfffb6e4c826bdf902daeb084d21` `chunk_id=srcchunk_47f469100a9220f590437a80cd8668c2` `native_locator=https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d` `source_timestamp=2024-05-04T07:20:00Z`
- EVMOS offers well-established tooling and documentation within the mature Cosmos ecosystem, with a modular design suitable for extensions. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d) `source_document_id=srcdoc_abc776d673caae53ff16c531d6248f0f` `source_revision_id=srcrev_6110bfffb6e4c826bdf902daeb084d21` `chunk_id=srcchunk_47f469100a9220f590437a80cd8668c2` `native_locator=https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d` `source_timestamp=2024-05-04T07:20:00Z`
- The primary drawback of EVMOS is its reliance on a mature yet outdated BFT consensus algorithm, which is inefficient and lacks scalability. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d) `source_document_id=srcdoc_abc776d673caae53ff16c531d6248f0f` `source_revision_id=srcrev_6110bfffb6e4c826bdf902daeb084d21` `chunk_id=srcchunk_47f469100a9220f590437a80cd8668c2` `native_locator=https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d` `source_timestamp=2024-05-04T07:20:00Z`
- EVMOS uses Comet BFT consensus, Berachain uses Comet BFT, and Meter.io uses Hot Stuff 2 with BLS signature aggregation. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d) `source_document_id=srcdoc_abc776d673caae53ff16c531d6248f0f` `source_revision_id=srcrev_6110bfffb6e4c826bdf902daeb084d21` `chunk_id=srcchunk_47f469100a9220f590437a80cd8668c2` `native_locator=https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d` `source_timestamp=2024-05-04T07:20:00Z`
- Extensibility ratings: EVMOS 5, Berachain 5, Meter.io 3. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d) `source_document_id=srcdoc_abc776d673caae53ff16c531d6248f0f` `source_revision_id=srcrev_6110bfffb6e4c826bdf902daeb084d21` `chunk_id=srcchunk_47f469100a9220f590437a80cd8668c2` `native_locator=https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d` `source_timestamp=2024-05-04T07:20:00Z`
- Stability ratings: EVMOS 5, Berachain 3, Meter.io 4. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d) `source_document_id=srcdoc_abc776d673caae53ff16c531d6248f0f` `source_revision_id=srcrev_6110bfffb6e4c826bdf902daeb084d21` `chunk_id=srcchunk_47f469100a9220f590437a80cd8668c2` `native_locator=https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d` `source_timestamp=2024-05-04T07:20:00Z`
- Innovation ratings: EVMOS 3, Berachain 5, Meter.io 5. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d) `source_document_id=srcdoc_abc776d673caae53ff16c531d6248f0f` `source_revision_id=srcrev_6110bfffb6e4c826bdf902daeb084d21` `chunk_id=srcchunk_47f469100a9220f590437a80cd8668c2` `native_locator=https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d` `source_timestamp=2024-05-04T07:20:00Z`
- Ecosystem ratings: EVMOS 5, Berachain 4, Meter.io 4. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d) `source_document_id=srcdoc_abc776d673caae53ff16c531d6248f0f` `source_revision_id=srcrev_6110bfffb6e4c826bdf902daeb084d21` `chunk_id=srcchunk_47f469100a9220f590437a80cd8668c2` `native_locator=https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d` `source_timestamp=2024-05-04T07:20:00Z`
- EVMOS's biggest challenge is an old code base, Berachain's is BUSL limitation, and Meter.io's is not being widely used. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d) `source_document_id=srcdoc_abc776d673caae53ff16c531d6248f0f` `source_revision_id=srcrev_6110bfffb6e4c826bdf902daeb084d21` `chunk_id=srcchunk_47f469100a9220f590437a80cd8668c2` `native_locator=https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d` `source_timestamp=2024-05-04T07:20:00Z`
- EVMOS's biggest advantage is being a mature project, Berachain's is newer/cleaner architecture, and Meter.io's is being fully customizable. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d) `source_document_id=srcdoc_abc776d673caae53ff16c531d6248f0f` `source_revision_id=srcrev_6110bfffb6e4c826bdf902daeb084d21` `chunk_id=srcchunk_47f469100a9220f590437a80cd8668c2` `native_locator=https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d` `source_timestamp=2024-05-04T07:20:00Z`
- Meter.io documentation is minimal with direct engineering support. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d) `source_document_id=srcdoc_abc776d673caae53ff16c531d6248f0f` `source_revision_id=srcrev_6110bfffb6e4c826bdf902daeb084d21` `chunk_id=srcchunk_47f469100a9220f590437a80cd8668c2` `native_locator=https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d` `source_timestamp=2024-05-04T07:20:00Z`
- Collaborating with the Meter team to develop a new framework provides an opportunity to delve deeply into the networking and consensus layers. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d) `source_document_id=srcdoc_abc776d673caae53ff16c531d6248f0f` `source_revision_id=srcrev_6110bfffb6e4c826bdf902daeb084d21` `chunk_id=srcchunk_47f469100a9220f590437a80cd8668c2` `native_locator=https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d` `source_timestamp=2024-05-04T07:20:00Z`

## Open Questions

- Is Ethereum compatible wallet usable on Cosmos based blockchain?
- Will wallet connect work on Cosmos based blockchain?

## Sources

- `source_document_id`: `srcdoc_abc776d673caae53ff16c531d6248f0f`
- `source_revision_id`: `srcrev_6110bfffb6e4c826bdf902daeb084d21`
- `source_url`: [Notion source](https://www.notion.so/L1-tech-framework-47372914e2174f7c90cf8d3693f4fc6d)
