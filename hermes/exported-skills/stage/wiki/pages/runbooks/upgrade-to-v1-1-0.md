---
title: "Upgrade to v1.1.0"
type: "runbook"
slug: "runbooks/upgrade-to-v1-1-0"
freshness: "2024-05-10T21:29:00Z"
tags:
  - "smart-contracts"
  - "upgrade"
  - "v1.1.0"
owners: []
source_revision_ids:
  - "srcrev_a0b01408d3b93286ebc387354612f711"
conflict_state: "none"
---

# Upgrade to v1.1.0

## Summary

Details of the smart contract changes in the upgrade from v1.0.0-rc.1 to v1.1.0, including new, modified, and removed contracts.

## Claims

- The upgrade diff is available at https://github.com/storyprotocol/protocol-core-v1/compare/v1.0.0-rc.1...v1.1.0. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Upgrades-to-v1-1-0-hackathon-7ad8a823c20440e48ea0135b38edddbd) `source_document_id=srcdoc_0049ca4b3163533f642645b3c5ac2e5b` `source_revision_id=srcrev_a0b01408d3b93286ebc387354612f711` `chunk_id=srcchunk_b22f56f4596c79aad53b82590a5ece1e` `native_locator=https://www.notion.so/Upgrades-to-v1-1-0-hackathon-7ad8a823c20440e48ea0135b38edddbd` `source_timestamp=2024-05-10T21:29:00Z`
- Contracts with changes or new contracts include IPAccountImpl.sol, IPAccountStorage.sol, LicenseToken.sol, AccessController.sol, DisputeModule.sol, ArbitrationPolicySP.sol, LicensingModule.sol, PILicenseTemplate.sol, LicensorApprovalChecker.sol (redeploy), RoyaltyModule.sol, IpRoyaltyVault.sol, RoyaltyPolicyLAP.sol, ProtocolPausableUpgradeable.sol, ProtocolPauseAdmin.sol, IPAccountRegistry.sol, IPAssetRegistry.sol, LicenseRegistry.sol, and ModuleRegistry.sol. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Upgrades-to-v1-1-0-hackathon-7ad8a823c20440e48ea0135b38edddbd) `source_document_id=srcdoc_0049ca4b3163533f642645b3c5ac2e5b` `source_revision_id=srcrev_a0b01408d3b93286ebc387354612f711` `chunk_id=srcchunk_b22f56f4596c79aad53b82590a5ece1e` `native_locator=https://www.notion.so/Upgrades-to-v1-1-0-hackathon-7ad8a823c20440e48ea0135b38edddbd` `source_timestamp=2024-05-10T21:29:00Z`
- Some contracts require initialAuthority to be set, including AccessController, DisputeModule, ArbitrationPolicySP, LicensingModule, and RoyaltyModule. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Upgrades-to-v1-1-0-hackathon-7ad8a823c20440e48ea0135b38edddbd) `source_document_id=srcdoc_0049ca4b3163533f642645b3c5ac2e5b` `source_revision_id=srcrev_a0b01408d3b93286ebc387354612f711` `chunk_id=srcchunk_b22f56f4596c79aad53b82590a5ece1e` `native_locator=https://www.notion.so/Upgrades-to-v1-1-0-hackathon-7ad8a823c20440e48ea0135b38edddbd` `source_timestamp=2024-05-10T21:29:00Z`
- Removed contracts include PILPolicyFrameworkManager.sol and Governance.sol. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Upgrades-to-v1-1-0-hackathon-7ad8a823c20440e48ea0135b38edddbd) `source_document_id=srcdoc_0049ca4b3163533f642645b3c5ac2e5b` `source_revision_id=srcrev_a0b01408d3b93286ebc387354612f711` `chunk_id=srcchunk_b22f56f4596c79aad53b82590a5ece1e` `native_locator=https://www.notion.so/Upgrades-to-v1-1-0-hackathon-7ad8a823c20440e48ea0135b38edddbd` `source_timestamp=2024-05-10T21:29:00Z`

## Sources

- `source_document_id`: `srcdoc_0049ca4b3163533f642645b3c5ac2e5b`
- `source_revision_id`: `srcrev_a0b01408d3b93286ebc387354612f711`
- `source_url`: [Notion source](https://www.notion.so/Upgrades-to-v1-1-0-hackathon-7ad8a823c20440e48ea0135b38edddbd)
