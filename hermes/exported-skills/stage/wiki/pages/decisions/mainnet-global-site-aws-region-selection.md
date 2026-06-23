---
title: "Mainnet Global Site AWS Region Selection"
type: "decision"
slug: "decisions/mainnet-global-site-aws-region-selection"
freshness: "2026-01-12T06:39:32Z"
tags:
  - "aws"
  - "infrastructure"
  - "korea"
  - "mainnet"
  - "region"
owners:
  - "subteam:S083BDZ4FTM"
source_revision_ids:
  - "srcrev_1971cd67e573d8eda3ce0711c96de633"
  - "srcrev_2ce23dfb5795ca142f916e276fca92cc"
  - "srcrev_3a81ecc354aedc3a1679a3f9559e7733"
  - "srcrev_748e976321a3fffdaedb290df62737db"
  - "srcrev_7bfdf312d84af6cd3cf95cf7763afa69"
conflict_state: "none"
---

# Mainnet Global Site AWS Region Selection

## Summary

Decision to use the Korean AWS region (ap-northeast-2) for the mainnet global site, superseding the initial plan to use Tokyo and Frankfurt, driven by company strategy to support Korean customers and builders.

## Claims

- Initial plan was to use ap-northeast-1 (Tokyo) and eu-central-1 (Frankfurt) for mainnet global site. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c34b356992b7550798ac0422d0e76bc9` `source_revision_id=srcrev_2ce23dfb5795ca142f916e276fca92cc` `chunk_id=srcchunk_0c5e294d03fbc274528cfb37ee2c8669` `native_locator=slack:C0547N89JUB:1768197873.577249:1768197873.577249` `source_timestamp=2026-01-12T06:04:33Z`
- Tokyo (ap-northeast-1) chosen for 4 AZs (highest in APAC) and 5 Direct Connect locations including Osaka. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c34b356992b7550798ac0422d0e76bc9` `source_revision_id=srcrev_2ce23dfb5795ca142f916e276fca92cc` `chunk_id=srcchunk_0c5e294d03fbc274528cfb37ee2c8669` `native_locator=slack:C0547N89JUB:1768197873.577249:1768197873.577249` `source_timestamp=2026-01-12T06:04:33Z`
- Frankfurt (eu-central-1) chosen for 15 Direct Connect locations (highest in EU), covering major cities. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c34b356992b7550798ac0422d0e76bc9` `source_revision_id=srcrev_2ce23dfb5795ca142f916e276fca92cc` `chunk_id=srcchunk_0c5e294d03fbc274528cfb37ee2c8669` `native_locator=slack:C0547N89JUB:1768197873.577249:1768197873.577249` `source_timestamp=2026-01-12T06:04:33Z`
- Korean region (ap-northeast-2) was raised as an alternative due to company strategy to support Korea and expected growth in Korean customers and builders. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c34b356992b7550798ac0422d0e76bc9` `source_revision_id=srcrev_748e976321a3fffdaedb290df62737db` `chunk_id=srcchunk_56760eb72c28c51facd5a1d8581f4177` `native_locator=slack:C0547N89JUB:1768197873.577249:1768199789.504249` `source_timestamp=2026-01-12T06:36:29Z`
- ap-northeast-1 gets new AWS features first and has more AZs and Direct Connect options than ap-northeast-2, but Korean dApps could justify ap-northeast-2. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c34b356992b7550798ac0422d0e76bc9` `source_revision_id=srcrev_7bfdf312d84af6cd3cf95cf7763afa69` `chunk_id=srcchunk_9dde7c4ac76fa3d00f7b30c0a76fc611` `native_locator=slack:C0547N89JUB:1768197873.577249:1768199550.358119` `source_timestamp=2026-01-12T06:32:30Z`
- Since only instances are needed, early access to new AWS features is not critical. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c34b356992b7550798ac0422d0e76bc9` `source_revision_id=srcrev_1971cd67e573d8eda3ce0711c96de633` `chunk_id=srcchunk_16751156936056d2da49fe43d0739d1d` `native_locator=slack:C0547N89JUB:1768197873.577249:1768199902.523669` `source_timestamp=2026-01-12T06:38:22Z`
- Final decision: use the Korean region (ap-northeast-2) for the mainnet global site. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c34b356992b7550798ac0422d0e76bc9` `source_revision_id=srcrev_3a81ecc354aedc3a1679a3f9559e7733` `chunk_id=srcchunk_8fc3c53a39938341e805cd91866eddc9` `native_locator=slack:C0547N89JUB:1768197873.577249:1768199972.021149` `source_timestamp=2026-01-12T06:39:32Z`

## Sources

- `source_document_id`: `srcdoc_c34b356992b7550798ac0422d0e76bc9`
- `source_revision_id`: `srcrev_748e976321a3fffdaedb290df62737db`
