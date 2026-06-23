---
title: "Mainnet Global Site AWS Region Selection"
type: "decision"
slug: "decisions/mainnet-global-site-aws-region-selection"
freshness: "2026-01-12T06:39:32Z"
tags:
  - "ap-northeast-2"
  - "aws"
  - "eu-central-1"
  - "korea"
  - "mainnet"
  - "region"
owners: []
source_revision_ids:
  - "srcrev_1971cd67e573d8eda3ce0711c96de633"
  - "srcrev_2ce23dfb5795ca142f916e276fca92cc"
  - "srcrev_3a81ecc354aedc3a1679a3f9559e7733"
  - "srcrev_748e976321a3fffdaedb290df62737db"
  - "srcrev_7bfdf312d84af6cd3cf95cf7763afa69"
  - "srcrev_7f41aceb4031f634b4cd0303ed5d9c28"
conflict_state: "none"
---

# Mainnet Global Site AWS Region Selection

## Summary

Decision to deploy the mainnet global site in ap-northeast-2 (Seoul) and eu-central-1 (Frankfurt). The initial proposal included ap-northeast-1 (Tokyo) and eu-central-1, but after discussing the company's strategy to support Korean market, the Asia region was changed to ap-northeast-2.

## Claims

- Initial proposal was to use ap-northeast-1 (Tokyo) and eu-central-1 (Frankfurt) for the mainnet global site, citing ap-northeast-1's 4 Availability Zones and 5 Direct Connect locations (including Osaka), and eu-central-1's 15 Direct Connect locations across Europe. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c34b356992b7550798ac0422d0e76bc9` `source_revision_id=srcrev_2ce23dfb5795ca142f916e276fca92cc` `chunk_id=srcchunk_0c5e294d03fbc274528cfb37ee2c8669` `native_locator=slack:C0547N89JUB:1768197873.577249:1768197873.577249` `source_timestamp=2026-01-12T06:04:33Z`
- A team member asked whether the Korean region (ap-northeast-2) should be considered. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c34b356992b7550798ac0422d0e76bc9` `source_revision_id=srcrev_7f41aceb4031f634b4cd0303ed5d9c28` `chunk_id=srcchunk_358f725b2e1d1e271c0dfc4bd4433e05` `native_locator=slack:C0547N89JUB:1768197873.577249:1768199381.673079` `source_timestamp=2026-01-12T06:29:41Z`
- Consideration that ap-northeast-1 receives new AWS features earlier and has more AZs and Direct Connect options, but ap-northeast-2 could be prioritized for Korean decentralized applications. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c34b356992b7550798ac0422d0e76bc9` `source_revision_id=srcrev_7bfdf312d84af6cd3cf95cf7763afa69` `chunk_id=srcchunk_9dde7c4ac76fa3d00f7b30c0a76fc611` `native_locator=slack:C0547N89JUB:1768197873.577249:1768199550.358119` `source_timestamp=2026-01-12T06:32:30Z`
- Company strategy is to support the Korean market, with expectations of more Korean customers and builders. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c34b356992b7550798ac0422d0e76bc9` `source_revision_id=srcrev_748e976321a3fffdaedb290df62737db` `chunk_id=srcchunk_56760eb72c28c51facd5a1d8581f4177` `native_locator=slack:C0547N89JUB:1768197873.577249:1768199789.504249` `source_timestamp=2026-01-12T06:36:29Z`
- Requirement for only instances meant that the advantage of new AWS features in ap-northeast-1 was not a major factor. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c34b356992b7550798ac0422d0e76bc9` `source_revision_id=srcrev_1971cd67e573d8eda3ce0711c96de633` `chunk_id=srcchunk_16751156936056d2da49fe43d0739d1d` `native_locator=slack:C0547N89JUB:1768197873.577249:1768199902.523669` `source_timestamp=2026-01-12T06:38:22Z`
- Decision was made to use the Korean region (ap-northeast-2). `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c34b356992b7550798ac0422d0e76bc9` `source_revision_id=srcrev_3a81ecc354aedc3a1679a3f9559e7733` `chunk_id=srcchunk_8fc3c53a39938341e805cd91866eddc9` `native_locator=slack:C0547N89JUB:1768197873.577249:1768199972.021149` `source_timestamp=2026-01-12T06:39:32Z`

## Open Questions

- Is eu-central-1 (Frankfurt) still included in the deployment? The discussion only altered the Asia region and did not explicitly confirm the Europe region.

## Sources

- `source_document_id`: `srcdoc_c34b356992b7550798ac0422d0e76bc9`
- `source_revision_id`: `srcrev_7bfdf312d84af6cd3cf95cf7763afa69`
