---
title: "Internal Devnet Migration"
type: "decision"
slug: "decisions/internal-devnet-migration"
freshness: "2026-02-05T06:54:31Z"
tags:
  - "chain-id"
  - "devnet"
  - "internal-devnet"
  - "s3"
owners:
  - "Jin"
  - "Yao"
  - "Zak"
  - "Zaki"
source_revision_ids:
  - "srcrev_1858b846f00450bac02fc003b3e2f767"
  - "srcrev_29c556f034026e746cbaa92a140b0546"
  - "srcrev_5b4f1e5b2a1b0f6158bb13484a6d75e0"
  - "srcrev_639ff73952aa6fc4f6a1982e9e2c805b"
  - "srcrev_6b714646e7b11fd62a9623a3d781b0a5"
  - "srcrev_707d50cba12544ee70259580ebcad9a0"
  - "srcrev_808fd64c2fcae438b9846d6f2f038657"
  - "srcrev_81f173f8abcf3f9a689014153dadfa68"
  - "srcrev_9362bf0af67e059fa21cdb50686fb1cc"
  - "srcrev_a1f8a6d26baa556608ae23a56f754653"
  - "srcrev_a4032f3bd6481994ad6fdd2a08f01173"
  - "srcrev_a7482203826988aebd39d24d88aa8ac5"
  - "srcrev_baa5fcbb060db93fd07901bf7d209453"
  - "srcrev_d22cc4be6f9085b65f82fe3e64024a2e"
  - "srcrev_f88bc3655844af0681f96f0b41df1d8c"
conflict_state: "none"
---

# Internal Devnet Migration

## Summary

Discussion about migrating internal devnet to new setup, chain-id changes, and S3 binary usage.

## Claims

- Yao asked if the new internal-devnet is working and inquired about deprecating the old one. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_29c556f034026e746cbaa92a140b0546` `chunk_id=srcchunk_d83d068aaa46c6e1b61e9bf592f9f2bf` `native_locator=slack:C0547N89JUB:1769579086.050139:1769579086.050139` `source_timestamp=2026-01-28T05:44:46Z`
- Jin responded that the decision should be deferred until next week because they still need the current environment. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_81f173f8abcf3f9a689014153dadfa68` `chunk_id=srcchunk_12de9ccefb8c996e709afa103aef2cea` `native_locator=slack:C0547N89JUB:1769579086.050139:1769582022.856699` `source_timestamp=2026-01-28T06:33:42Z`
- Yao agreed to wait until next week. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_9362bf0af67e059fa21cdb50686fb1cc` `chunk_id=srcchunk_c1f41e328c8a9689465ddd07397f0a17` `native_locator=slack:C0547N89JUB:1769579086.050139:1769582050.869729` `source_timestamp=2026-01-28T06:34:10Z`
- Later, Yao requested S3 binaries location for faster devnet reset. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_6b714646e7b11fd62a9623a3d781b0a5` `chunk_id=srcchunk_d4a00986f6fa27e17017461efe6411b5` `native_locator=slack:C0547N89JUB:1769579086.050139:1770177603.905659` `source_timestamp=2026-02-04T04:00:17Z`
- Jin provided a link to the story-devnet-binaries S3 bucket. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_707d50cba12544ee70259580ebcad9a0` `chunk_id=srcchunk_bd1f54c8e1a3df5a202608660882680c` `native_locator=slack:C0547N89JUB:1769579086.050139:1770177671.763229` `source_timestamp=2026-02-04T04:01:11Z`
- Yao encountered IAM permission error when uploading binaries. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_f88bc3655844af0681f96f0b41df1d8c` `chunk_id=srcchunk_f40b2f058da43c5682bce793bd08bdba` `native_locator=slack:C0547N89JUB:1769579086.050139:1770192367.639279` `source_timestamp=2026-02-04T08:06:07Z`
- Yao noted that the chain-id in devnet config is now story-localnet instead of internal-devnet-1. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_1858b846f00450bac02fc003b3e2f767` `chunk_id=srcchunk_47a09b977f0539d03ee4fd0e63d7c59c` `native_locator=slack:C0547N89JUB:1769579086.050139:1770196471.033219` `source_timestamp=2026-02-04T09:14:31Z`
- Jin asked if adding internal-devnet-1 to story upgrade file, and Yao agreed it could be added or use story-localnet chain-id for tests. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_a4032f3bd6481994ad6fdd2a08f01173` `chunk_id=srcchunk_1974517f3b06afd7b40632d67dbf0cd9` `native_locator=slack:C0547N89JUB:1769579086.050139:1770196661.180909` `source_timestamp=2026-02-04T09:17:41Z`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_d22cc4be6f9085b65f82fe3e64024a2e` `chunk_id=srcchunk_c74adb4f4e87cc596b703f7849d1494e` `native_locator=slack:C0547N89JUB:1769579086.050139:1770196730.275129` `source_timestamp=2026-02-04T09:18:50Z`
- Zak raised concern about exposing internal devnet info by adding to story upgrade file. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_a1f8a6d26baa556608ae23a56f754653` `chunk_id=srcchunk_c10f723b1d10f191cf5b591b7eff7c6a` `native_locator=slack:C0547N89JUB:1769579086.050139:1770264258.876459` `source_timestamp=2026-02-05T04:04:18Z`
- Jin requested EIP quota increase for L1 testing on AWS. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_5b4f1e5b2a1b0f6158bb13484a6d75e0` `chunk_id=srcchunk_e598060fb6cc7942504dae3f86d45ba8` `native_locator=slack:C0547N89JUB:1769579086.050139:1770273545.713809` `source_timestamp=2026-02-05T06:39:05Z`
- Yao clarified that EIP quota increase is reviewed by AWS. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_639ff73952aa6fc4f6a1982e9e2c805b` `chunk_id=srcchunk_99ca78a7ccf0c6fc8f3249c2e2e04b3a` `native_locator=slack:C0547N89JUB:1769579086.050139:1770273603.495029` `source_timestamp=2026-02-05T06:40:03Z`
- Jin is going to try using previous internal-devnet genesis files and only changing chain-id. `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_808fd64c2fcae438b9846d6f2f038657` `chunk_id=srcchunk_5d7aaaf311ccb0ed52036ea1d7e0b245` `native_locator=slack:C0547N89JUB:1769579086.050139:1770273731.241329` `source_timestamp=2026-02-05T06:42:11Z`
- Yao responded that only changing that is needed, and confirmed by Zaki. `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_a7482203826988aebd39d24d88aa8ac5` `chunk_id=srcchunk_4f219b4d49be9c6dad256fb969d66163` `native_locator=slack:C0547N89JUB:1769579086.050139:1770273898.785269` `source_timestamp=2026-02-05T06:44:58Z`
- Zaki added to ensure upgrade heights are specified in netconf/upgrades.go. `claim:claim_1_14` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_baa5fcbb060db93fd07901bf7d209453` `chunk_id=srcchunk_4d8f62641434bd80a12f5bfd326c1512` `native_locator=slack:C0547N89JUB:1769579086.050139:1770274471.585089` `source_timestamp=2026-02-05T06:54:31Z`

## Open Questions

- Should internal-devnet-1 be added to the public story upgrade file despite exposure concerns?

## Sources

- `source_document_id`: `srcdoc_294f2516b3e9908895b36e310eae179d`
- `source_revision_id`: `srcrev_275122a2c93e16ba77cf955657855581`
