---
title: "Beta 1st Week Architectural Review"
type: "decision"
slug: "decisions/beta-1st-week-architectural-review"
freshness: "2024-12-03T05:20:00Z"
tags:
  - "architecture"
  - "beta"
  - "protocol"
owners: []
source_revision_ids:
  - "srcrev_ab7040e1dbc4a9f8c448ddd978a3d56d"
conflict_state: "none"
---

# Beta 1st Week Architectural Review

## Summary

Architectural review of the beta protocol covering alpha architecture recap, principles-first rearchitecting, and example use-cases. Key decisions include abandoning IPOrgs, restructuring IPAs into four components, and defining protocol authorization flavors.

## Claims

- The architectural review agenda included recap on Alpha architecture, reasons for IPOrgs, issues with IPAs, principles-first rearchitecting, and example use-cases. `claim:claim_beta_arch_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-1) `source_document_id=srcdoc_b2a3d7e7ee954d310f72d8ea09f6c6a5` `source_revision_id=srcrev_ab7040e1dbc4a9f8c448ddd978a3d56d` `chunk_id=srcchunk_6182825c47b3436cea041735436aae53` `native_locator=https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-1` `source_timestamp=2024-12-03T05:20:00Z`
- IPOrgs were intended to group IP assets for branding, functionality, and ACL, but were abandoned because they imposed a single organizational restriction that didn't fit the malleable nature of IPAs, and treating IPAs as sovereign citizens avoids binding IP logic to a specific org. `claim:claim_beta_arch_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-1) `source_document_id=srcdoc_b2a3d7e7ee954d310f72d8ea09f6c6a5` `source_revision_id=srcrev_ab7040e1dbc4a9f8c448ddd978a3d56d` `chunk_id=srcchunk_6182825c47b3436cea041735436aae53` `native_locator=https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-1` `source_timestamp=2024-12-03T05:20:00Z`
- In Alpha, IPOrgs were only metadata wrappers without functionality or access control logic; the team never implemented ACL restrictions on IPAs and lacked a deterministic identification scheme for IPAs. `claim:claim_beta_arch_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-1) `source_document_id=srcdoc_b2a3d7e7ee954d310f72d8ea09f6c6a5` `source_revision_id=srcrev_ab7040e1dbc4a9f8c448ddd978a3d56d` `chunk_id=srcchunk_6182825c47b3436cea041735436aae53` `native_locator=https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-1` `source_timestamp=2024-12-03T05:20:00Z`
- The protocol architecture was categorized into Frontends Contracts (e.g., SPG, periphery contracts) and Backends/Modules (core protocol state and changes, e.g., IPRegistry). `claim:claim_beta_arch_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-2) `source_document_id=srcdoc_b2a3d7e7ee954d310f72d8ea09f6c6a5` `source_revision_id=srcrev_ab7040e1dbc4a9f8c448ddd978a3d56d` `chunk_id=srcchunk_3ec4e4f572bb7346d22ba6bdda568e9b` `native_locator=https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-2` `source_timestamp=2024-12-03T05:20:00Z`
- Core protocol authorization flavors include internal module-to-module calls, external frontend-to-module calls, logic within a module, and additional authorization by an IPA owner. `claim:claim_beta_arch_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-2) `source_document_id=srcdoc_b2a3d7e7ee954d310f72d8ea09f6c6a5` `source_revision_id=srcrev_ab7040e1dbc4a9f8c448ddd978a3d56d` `chunk_id=srcchunk_3ec4e4f572bb7346d22ba6bdda568e9b` `native_locator=https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-2` `source_timestamp=2024-12-03T05:20:00Z`
- An idea for protocol ACL involved composing the protocol of Frontends/Gateways that declare dependent modules, and a standard interface for authorization. `claim:claim_beta_arch_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-2) `source_document_id=srcdoc_b2a3d7e7ee954d310f72d8ea09f6c6a5` `source_revision_id=srcrev_ab7040e1dbc4a9f8c448ddd978a3d56d` `chunk_id=srcchunk_3ec4e4f572bb7346d22ba6bdda568e9b` `native_locator=https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-2` `source_timestamp=2024-12-03T05:20:00Z`
- Restructuring IPAs involved separating IPAccount creation into a separate factory, and splitting IP into four components: the IP itself (deterministically identified by an NFT), the IP account (for ACL), the IP asset record (on IP Registry), and IP Metadata. `claim:claim_beta_arch_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-3) `source_document_id=srcdoc_b2a3d7e7ee954d310f72d8ea09f6c6a5` `source_revision_id=srcrev_ab7040e1dbc4a9f8c448ddd978a3d56d` `chunk_id=srcchunk_d583788f9c3667dc6c06301ef8067c2c` `native_locator=https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-3` `source_timestamp=2024-12-03T05:20:00Z`
- Example use-cases discussed included Frontend Enrollment, IP Registration, and Module Interaction (e.g., licensing). `claim:claim_beta_arch_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-3) `source_document_id=srcdoc_b2a3d7e7ee954d310f72d8ea09f6c6a5` `source_revision_id=srcrev_ab7040e1dbc4a9f8c448ddd978a3d56d` `chunk_id=srcchunk_d583788f9c3667dc6c06301ef8067c2c` `native_locator=https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-3` `source_timestamp=2024-12-03T05:20:00Z`
- The end-of-day summary noted alignment and TBD items. `claim:claim_beta_arch_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-3) `source_document_id=srcdoc_b2a3d7e7ee954d310f72d8ea09f6c6a5` `source_revision_id=srcrev_ab7040e1dbc4a9f8c448ddd978a3d56d` `chunk_id=srcchunk_d583788f9c3667dc6c06301ef8067c2c` `native_locator=https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a#chunk-3` `source_timestamp=2024-12-03T05:20:00Z`

## Open Questions

- TBD items from end-of-day summary remain unresolved.

## Related Pages

- `concepts/protocol-home`

## Sources

- `source_document_id`: `srcdoc_b2a3d7e7ee954d310f72d8ea09f6c6a5`
- `source_revision_id`: `srcrev_ab7040e1dbc4a9f8c448ddd978a3d56d`
- `source_url`: [Notion source](https://www.notion.so/Beta-1st-Week-Architectural-Review-325b40188cd649618c4ba4439d41151a)
