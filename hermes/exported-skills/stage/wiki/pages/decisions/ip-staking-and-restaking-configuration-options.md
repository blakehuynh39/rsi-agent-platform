---
title: "$IP Staking and Restaking Configuration Options"
type: "decision"
slug: "decisions/ip-staking-and-restaking-configuration-options"
freshness: "2024-05-01T22:59:00Z"
tags:
  - "ip-token"
  - "restaking"
  - "rip"
  - "sip"
  - "staking"
  - "tokenomics"
owners: []
source_revision_ids:
  - "srcrev_d939d41133a6cd766e6c958781d83c74"
conflict_state: "none"
---

# $IP Staking and Restaking Configuration Options

## Summary

Explores possible configurations for $IP staking into validators ($IP → $sIP) and IPA pools ($IP → $LIP or $sIP → $rIP), including trade-offs around restaking independence, slashing mechanisms, and capital efficiency.

## Claims

- Validator staking converts $IP to $sIP, while IPA staking can convert $IP to $LIP and/or $sIP to $rIP. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072) `source_document_id=srcdoc_adbd74d9450f6ad1c22930e0ffd3942b` `source_revision_id=srcrev_d939d41133a6cd766e6c958781d83c74` `chunk_id=srcchunk_a62ea0b9a7a40769421c6aee5290c213` `native_locator=https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072` `source_timestamp=2024-05-01T22:59:00Z`
- Key design choices include whether to keep validator staking and IPA pool staking independent or allow restaking, what percentage of $sIP can be restaked, whether IPA pool slashing or wealth transfer occurs, and whether $rIP exists as a receipt token. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072) `source_document_id=srcdoc_adbd74d9450f6ad1c22930e0ffd3942b` `source_revision_id=srcrev_d939d41133a6cd766e6c958781d83c74` `chunk_id=srcchunk_a62ea0b9a7a40769421c6aee5290c213` `native_locator=https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072` `source_timestamp=2024-05-01T22:59:00Z`
- Option 1 (keep independent) avoids double-slashing but may lose capital efficiency if a significant percentage of stakers want to restake on IPA pools. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072) `source_document_id=srcdoc_adbd74d9450f6ad1c22930e0ffd3942b` `source_revision_id=srcrev_d939d41133a6cd766e6c958781d83c74` `chunk_id=srcchunk_a62ea0b9a7a40769421c6aee5290c213` `native_locator=https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072` `source_timestamp=2024-05-01T22:59:00Z`
- Option 2 (IP + sIP + sIP slashing + rIP + 100% restaking allowed) provides maximum capital efficiency but introduces double-slashing risk. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072) `source_document_id=srcdoc_adbd74d9450f6ad1c22930e0ffd3942b` `source_revision_id=srcrev_d939d41133a6cd766e6c958781d83c74` `chunk_id=srcchunk_a62ea0b9a7a40769421c6aee5290c213` `native_locator=https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072` `source_timestamp=2024-05-01T22:59:00Z`
- Option 3 (IP + sIP + sIP non-slashing via wealth transfer to other IPA stakers + rIP + 100% restaking allowed) offers maximum capital efficiency without double-slashing risk but could make IPA pool staking riskless and require (3,3) game theory. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072) `source_document_id=srcdoc_adbd74d9450f6ad1c22930e0ffd3942b` `source_revision_id=srcrev_d939d41133a6cd766e6c958781d83c74` `chunk_id=srcchunk_a62ea0b9a7a40769421c6aee5290c213` `native_locator=https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072` `source_timestamp=2024-05-01T22:59:00Z`
- Option 4 (IP + sIP + sIP non-slashing via wealth transfer to SP treasury + rIP + 100% restaking allowed) provides maximum capital efficiency and no double-slashing risk but concentrates $IP in the treasury over time. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072) `source_document_id=srcdoc_adbd74d9450f6ad1c22930e0ffd3942b` `source_revision_id=srcrev_d939d41133a6cd766e6c958781d83c74` `chunk_id=srcchunk_a62ea0b9a7a40769421c6aee5290c213` `native_locator=https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072` `source_timestamp=2024-05-01T22:59:00Z`
- Option 5 (IP + sIP + sIP slashing + rIP + less than 100% restaking allowed) mitigates double-slashing risk by limiting the portion of $IP that can be restaked. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072) `source_document_id=srcdoc_adbd74d9450f6ad1c22930e0ffd3942b` `source_revision_id=srcrev_d939d41133a6cd766e6c958781d83c74` `chunk_id=srcchunk_a62ea0b9a7a40769421c6aee5290c213` `native_locator=https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072` `source_timestamp=2024-05-01T22:59:00Z`
- Ethereum staking rate is approximately 26% (32M/122M) and restaking rate is approximately 3% (4M/122M), both growing. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072) `source_document_id=srcdoc_adbd74d9450f6ad1c22930e0ffd3942b` `source_revision_id=srcrev_d939d41133a6cd766e6c958781d83c74` `chunk_id=srcchunk_a62ea0b9a7a40769421c6aee5290c213` `native_locator=https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072` `source_timestamp=2024-05-01T22:59:00Z`

## Open Questions

- How to redistribute treasury-concentrated $IP in option 4?
- Should $rIP exist as a receipt token?
- Should IPA pool use slashing or wealth transfer?
- What percentage of $sIP should be allowed for restaking?

## Sources

- `source_document_id`: `srcdoc_adbd74d9450f6ad1c22930e0ffd3942b`
- `source_revision_id`: `srcrev_d939d41133a6cd766e6c958781d83c74`
- `source_url`: [Notion source](https://www.notion.so/Possible-configurations-for-the-IP-sIP-rIP-and-trade-offs-3c4b3abbeac64a5ea7143a7a1f267072)
