---
title: "License Token: ERC-721 vs ERC-1155"
type: "decision"
slug: "decisions/license-token-standard-evaluation"
freshness: "2026-05-05T06:32:50Z"
tags:
  - "license-token"
  - "smart-contracts"
  - "story-protocol"
owners: []
source_revision_ids:
  - "srcrev_92980056bda4f24e2653dad215ed9735"
conflict_state: "none"
---

# License Token: ERC-721 vs ERC-1155

## Summary

Analysis of using ERC-721 vs ERC-1155 for License Tokens in Story Protocol, including benefits, downsides, and required changes.

## Claims

- Story Protocol currently uses ERC-1155 for minting License NFTs. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/12-License-Token-721-vs-1155-614d837aa3df45ad82e2023d57cd1d45) `source_document_id=srcdoc_6936a283155a22b535ea280c9ab13ec7` `source_revision_id=srcrev_92980056bda4f24e2653dad215ed9735` `chunk_id=srcchunk_cadc550ff8d23aba66cd639a0de220ff` `native_locator=https://www.notion.so/12-License-Token-721-vs-1155-614d837aa3df45ad82e2023d57cd1d45` `source_timestamp=2026-05-05T06:32:50Z`
- ERC-1155 allows multiple minters of identical license parameters to share the same token ID, maintaining fungibility. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/12-License-Token-721-vs-1155-614d837aa3df45ad82e2023d57cd1d45) `source_document_id=srcdoc_6936a283155a22b535ea280c9ab13ec7` `source_revision_id=srcrev_92980056bda4f24e2653dad215ed9735` `chunk_id=srcchunk_cadc550ff8d23aba66cd639a0de220ff` `native_locator=https://www.notion.so/12-License-Token-721-vs-1155-614d837aa3df45ad82e2023d57cd1d45` `source_timestamp=2026-05-05T06:32:50Z`
- If ERC-721 were used, each mint would produce a unique token ID, even for identical license parameters, breaking fungibility. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/12-License-Token-721-vs-1155-614d837aa3df45ad82e2023d57cd1d45) `source_document_id=srcdoc_6936a283155a22b535ea280c9ab13ec7` `source_revision_id=srcrev_92980056bda4f24e2653dad215ed9735` `chunk_id=srcchunk_cadc550ff8d23aba66cd639a0de220ff` `native_locator=https://www.notion.so/12-License-Token-721-vs-1155-614d837aa3df45ad82e2023d57cd1d45` `source_timestamp=2026-05-05T06:32:50Z`
- ERC-1155 has worse marketplace and app support than ERC-721, due to legacy adoption and image-based NFT dominance. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/12-License-Token-721-vs-1155-614d837aa3df45ad82e2023d57cd1d45) `source_document_id=srcdoc_6936a283155a22b535ea280c9ab13ec7` `source_revision_id=srcrev_92980056bda4f24e2653dad215ed9735` `chunk_id=srcchunk_cadc550ff8d23aba66cd639a0de220ff` `native_locator=https://www.notion.so/12-License-Token-721-vs-1155-614d837aa3df45ad82e2023d57cd1d45` `source_timestamp=2026-05-05T06:32:50Z`
- License parameters consist of policyId, licensorIpId, and transferable, which are hashed to form a unique license identifier. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/12-License-Token-721-vs-1155-614d837aa3df45ad82e2023d57cd1d45) `source_document_id=srcdoc_6936a283155a22b535ea280c9ab13ec7` `source_revision_id=srcrev_92980056bda4f24e2653dad215ed9735` `chunk_id=srcchunk_cadc550ff8d23aba66cd639a0de220ff` `native_locator=https://www.notion.so/12-License-Token-721-vs-1155-614d837aa3df45ad82e2023d57cd1d45` `source_timestamp=2026-05-05T06:32:50Z`

## Sources

- `source_document_id`: `srcdoc_6936a283155a22b535ea280c9ab13ec7`
- `source_revision_id`: `srcrev_92980056bda4f24e2653dad215ed9735`
- `source_url`: [Notion source](https://www.notion.so/12-License-Token-721-vs-1155-614d837aa3df45ad82e2023d57cd1d45)
