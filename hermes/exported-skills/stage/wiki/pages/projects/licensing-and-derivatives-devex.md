---
title: "Licensing and Derivatives DevEx"
type: "project"
slug: "projects/licensing-and-derivatives-devex"
freshness: "2026-05-05T06:39:00Z"
tags:
  - "derivatives"
  - "developer-experience"
  - "ip-asset-registry"
  - "licensing"
  - "lnft"
owners: []
source_revision_ids:
  - "srcrev_0c9917762c404d012527a8a06cd892df"
conflict_state: "none"
---

# Licensing and Derivatives DevEx

## Summary

Exploration of developer experience improvements for licensing and derivative registration in Story Protocol, including skipping LNFT minting in certain cases, default non-commercial remix policies, direct parent linking from IPAssetRegistry, and optional IPAccount deployments.

## Claims

- If the caller is a licensor registering a derivative of their own work, a method could allow registration and linking without minting/burning a license, just pointing to the parent policy. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-1) `source_document_id=srcdoc_b4f3eeb81317e4d0378414d4f885916f` `source_revision_id=srcrev_0c9917762c404d012527a8a06cd892df` `chunk_id=srcchunk_904984e1fb8921a8eefbb4f592f6ed38` `native_locator=https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-1` `source_timestamp=2026-05-05T06:39:00Z`
- If the original IPA owner has called addPolicy in the licensor and the registrant passes the checks, the derivative could be registered without minting and linking as separate steps. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-1) `source_document_id=srcdoc_b4f3eeb81317e4d0378414d4f885916f` `source_revision_id=srcrev_0c9917762c404d012527a8a06cd892df` `chunk_id=srcchunk_904984e1fb8921a8eefbb4f592f6ed38` `native_locator=https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-1` `source_timestamp=2026-05-05T06:39:00Z`
- A default Non-Commercial Social Remixing policy should be automatically attached to each IPA, allowing permissionless derivation for non-commercial use, with IP holders able to opt out. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-2) `source_document_id=srcdoc_b4f3eeb81317e4d0378414d4f885916f` `source_revision_id=srcrev_0c9917762c404d012527a8a06cd892df` `chunk_id=srcchunk_2537efe04eefe5a43da1956676b6afa1` `native_locator=https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-2` `source_timestamp=2026-05-05T06:39:00Z`
- In the real world, anyone can make a derivative regardless of licensing terms; it only becomes a problem upon commercialization. Default allow-listing only to cases the IP owner has foreseen is very limiting. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-2) `source_document_id=srcdoc_b4f3eeb81317e4d0378414d4f885916f` `source_revision_id=srcrev_0c9917762c404d012527a8a06cd892df` `chunk_id=srcchunk_2537efe04eefe5a43da1956676b6afa1` `native_locator=https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-2` `source_timestamp=2026-05-05T06:39:00Z`
- Support variable price hooks for all policies and frameworks, so changing price does not require a new policy. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-3) `source_document_id=srcdoc_b4f3eeb81317e4d0378414d4f885916f` `source_revision_id=srcrev_0c9917762c404d012527a8a06cd892df` `chunk_id=srcchunk_6d4fe6e7daa82ff47097e7c5d8634a91` `native_locator=https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-3` `source_timestamp=2026-05-05T06:39:00Z`
- Allow other contracts to sell or mint licenses (1combo use case). `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-3) `source_document_id=srcdoc_b4f3eeb81317e4d0378414d4f885916f` `source_revision_id=srcrev_0c9917762c404d012527a8a06cd892df` `chunk_id=srcchunk_6d4fe6e7daa82ff47097e7c5d8634a91` `native_locator=https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-3` `source_timestamp=2026-05-05T06:39:00Z`
- Move parent/child relationship out of LicensingModule by allowing IPAssetRegistry.registerDerivative to call LicensingModule.linkToParent, simplifying the user flow. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-3) `source_document_id=srcdoc_b4f3eeb81317e4d0378414d4f885916f` `source_revision_id=srcrev_0c9917762c404d012527a8a06cd892df` `chunk_id=srcchunk_6d4fe6e7daa82ff47097e7c5d8634a91` `native_locator=https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-3` `source_timestamp=2026-05-05T06:39:00Z`
- IPAssetRegistry should remain simple since it is not upgradeable; this is what RegistrationModule does now. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-3) `source_document_id=srcdoc_b4f3eeb81317e4d0378414d4f885916f` `source_revision_id=srcrev_0c9917762c404d012527a8a06cd892df` `chunk_id=srcchunk_6d4fe6e7daa82ff47097e7c5d8634a91` `native_locator=https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-3` `source_timestamp=2026-05-05T06:39:00Z`
- Consider more permanent LNFTs that grant infinite derivative rights or N derivatives (burned after spending its limit). `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-3) `source_document_id=srcdoc_b4f3eeb81317e4d0378414d4f885916f` `source_revision_id=srcrev_0c9917762c404d012527a8a06cd892df` `chunk_id=srcchunk_6d4fe6e7daa82ff47097e7c5d8634a91` `native_locator=https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-3` `source_timestamp=2026-05-05T06:39:00Z`
- A templateId should be stored separately so users can find templates like 'Commercial Social Remixing' (template 3) and configure values, resulting in Policy N, Template 3. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-4) `source_document_id=srcdoc_b4f3eeb81317e4d0378414d4f885916f` `source_revision_id=srcrev_0c9917762c404d012527a8a06cd892df` `chunk_id=srcchunk_100e8f3024b957ffae65ae4515400bb0` `native_locator=https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-4` `source_timestamp=2026-05-05T06:39:00Z`
- Explore not requiring mandatory IPAccount deployments when NFTs to be registered are already owned by an existing IPAccount, allowing them to be controlled from that account and optionally deployed independently later. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-4) `source_document_id=srcdoc_b4f3eeb81317e4d0378414d4f885916f` `source_revision_id=srcrev_0c9917762c404d012527a8a06cd892df` `chunk_id=srcchunk_100e8f3024b957ffae65ae4515400bb0` `native_locator=https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5#chunk-4` `source_timestamp=2026-05-05T06:39:00Z`

## Open Questions

- How to balance IPAssetRegistry simplicity (non-upgradeable) with adding derivative registration logic?
- Should the default Non-Commercial Social Remix policy be opt-out or opt-in? Discussion with Ben Sternberg needed.
- What are the full implications of permanent LNFTs with infinite or limited derivative rights?

## Sources

- `source_document_id`: `srcdoc_b4f3eeb81317e4d0378414d4f885916f`
- `source_revision_id`: `srcrev_0c9917762c404d012527a8a06cd892df`
- `source_url`: [Notion source](https://www.notion.so/Licensing-and-Derivatives-DevEx-552614d481fc4ee8a5a071a8bd5287b5)
