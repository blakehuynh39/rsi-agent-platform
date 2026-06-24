---
title: "Internal Devnet Infrastructure"
type: "project"
slug: "projects/internal-devnet-infrastructure"
freshness: "2026-02-05T06:54:31Z"
tags:
  - "aws"
  - "chain-id"
  - "devnet"
  - "eip"
  - "s3"
owners:
  - "U079ZJ48D62"
  - "U07A7AUGL5V"
  - "U07KLPN0JN6"
  - "U07TNT9N4JC"
source_revision_ids:
  - "srcrev_1858b846f00450bac02fc003b3e2f767"
  - "srcrev_29c556f034026e746cbaa92a140b0546"
  - "srcrev_40997ca502215e2186332dabde0c0154"
  - "srcrev_5883c3872d8ff2b81ecc529bcb728635"
  - "srcrev_5b4f1e5b2a1b0f6158bb13484a6d75e0"
  - "srcrev_639ff73952aa6fc4f6a1982e9e2c805b"
  - "srcrev_6b714646e7b11fd62a9623a3d781b0a5"
  - "srcrev_707d50cba12544ee70259580ebcad9a0"
  - "srcrev_7c00ae2acc246936432c5c6852f6949a"
  - "srcrev_808fd64c2fcae438b9846d6f2f038657"
  - "srcrev_81f173f8abcf3f9a689014153dadfa68"
  - "srcrev_9362bf0af67e059fa21cdb50686fb1cc"
  - "srcrev_a1f8a6d26baa556608ae23a56f754653"
  - "srcrev_a4032f3bd6481994ad6fdd2a08f01173"
  - "srcrev_b51311e46dd1d6635b6cfef071df9de3"
  - "srcrev_baa5fcbb060db93fd07901bf7d209453"
  - "srcrev_f88bc3655844af0681f96f0b41df1d8c"
conflict_state: "none"
---

# Internal Devnet Infrastructure

## Summary

Management and configuration details for the internal Story Protocol devnet on AWS, including binary storage, chain ID transitions, and access permissions.

## Claims

- Deprecation of the old internal-devnet is under discussion; decision deferred to next week. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_29c556f034026e746cbaa92a140b0546` `chunk_id=srcchunk_d83d068aaa46c6e1b61e9bf592f9f2bf` `native_locator=slack:C0547N89JUB:1769579086.050139:1769579086.050139` `source_timestamp=2026-01-28T05:44:46Z`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_81f173f8abcf3f9a689014153dadfa68` `chunk_id=srcchunk_12de9ccefb8c996e709afa103aef2cea` `native_locator=slack:C0547N89JUB:1769579086.050139:1769582022.856699` `source_timestamp=2026-01-28T06:33:42Z`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_9362bf0af67e059fa21cdb50686fb1cc` `chunk_id=srcchunk_c1f41e328c8a9689465ddd07397f0a17` `native_locator=slack:C0547N89JUB:1769579086.050139:1769582050.869729` `source_timestamp=2026-01-28T06:34:10Z`
- Devnet reset workflow leverages S3 binaries; the default bucket is story-devnet-binaries in us-east-1. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_6b714646e7b11fd62a9623a3d781b0a5` `chunk_id=srcchunk_d4a00986f6fa27e17017461efe6411b5` `native_locator=slack:C0547N89JUB:1769579086.050139:1770177603.905659` `source_timestamp=2026-02-04T04:00:17Z`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_707d50cba12544ee70259580ebcad9a0` `chunk_id=srcchunk_bd1f54c8e1a3df5a202608660882680c` `native_locator=slack:C0547N89JUB:1769579086.050139:1770177671.763229` `source_timestamp=2026-02-04T04:01:11Z`
- IAM permission issues with the S3 bucket were resolved by adjusting access permissions. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_f88bc3655844af0681f96f0b41df1d8c` `chunk_id=srcchunk_f40b2f058da43c5682bce793bd08bdba` `native_locator=slack:C0547N89JUB:1769579086.050139:1770192367.639279` `source_timestamp=2026-02-04T08:06:07Z`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_b51311e46dd1d6635b6cfef071df9de3` `chunk_id=srcchunk_7e643704900bd2b4cf194f6e11f7e20e` `native_locator=slack:C0547N89JUB:1769579086.050139:1770197000.646059` `source_timestamp=2026-02-04T09:23:20Z`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_40997ca502215e2186332dabde0c0154` `chunk_id=srcchunk_ee8beccaf44b02372640b64768f61a06` `native_locator=slack:C0547N89JUB:1769579086.050139:1770197087.962469` `source_timestamp=2026-02-04T09:24:47Z`
- The AWS devnet now uses story-localnet as its chain-id (previously internal-devnet-1). `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_1858b846f00450bac02fc003b3e2f767` `chunk_id=srcchunk_47a09b977f0539d03ee4fd0e63d7c59c` `native_locator=slack:C0547N89JUB:1769579086.050139:1770196471.033219` `source_timestamp=2026-02-04T09:14:31Z`
- There is an ongoing discussion about adding internal-devnet-1 to the story upgrade file (netconf/chains.go); no decision made due to concerns about exposing internal devnet info to external members. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_a4032f3bd6481994ad6fdd2a08f01173` `chunk_id=srcchunk_1974517f3b06afd7b40632d67dbf0cd9` `native_locator=slack:C0547N89JUB:1769579086.050139:1770196661.180909` `source_timestamp=2026-02-04T09:17:41Z`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_7c00ae2acc246936432c5c6852f6949a` `chunk_id=srcchunk_757667cabc6c5f2628bed6ccd58de2f0` `native_locator=slack:C0547N89JUB:1769579086.050139:1770196688.679959` `source_timestamp=2026-02-04T09:18:08Z`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_5883c3872d8ff2b81ecc529bcb728635` `chunk_id=srcchunk_50d56f9d71ff13728251ada297396b27` `native_locator=slack:C0547N89JUB:1769579086.050139:1770196950.959279` `source_timestamp=2026-02-04T09:22:30Z`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_a1f8a6d26baa556608ae23a56f754653` `chunk_id=srcchunk_c10f723b1d10f191cf5b591b7eff7c6a` `native_locator=slack:C0547N89JUB:1769579086.050139:1770264258.876459` `source_timestamp=2026-02-05T04:04:18Z`
- EIP quota increase for L1 testing validators was requested and is under review by AWS. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_5b4f1e5b2a1b0f6158bb13484a6d75e0` `chunk_id=srcchunk_e598060fb6cc7942504dae3f86d45ba8` `native_locator=slack:C0547N89JUB:1769579086.050139:1770273545.713809` `source_timestamp=2026-02-05T06:39:05Z`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_639ff73952aa6fc4f6a1982e9e2c805b` `chunk_id=srcchunk_99ca78a7ccf0c6fc8f3249c2e2e04b3a` `native_locator=slack:C0547N89JUB:1769579086.050139:1770273603.495029` `source_timestamp=2026-02-05T06:40:03Z`
- To adjust a devnet chain-id, modify the genesis file chain-id and ensure upgrade heights are present in netconf/upgrades.go. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_808fd64c2fcae438b9846d6f2f038657` `chunk_id=srcchunk_5d7aaaf311ccb0ed52036ea1d7e0b245` `native_locator=slack:C0547N89JUB:1769579086.050139:1770273731.241329` `source_timestamp=2026-02-05T06:42:11Z`
  - citation: `source_document_id=srcdoc_294f2516b3e9908895b36e310eae179d` `source_revision_id=srcrev_baa5fcbb060db93fd07901bf7d209453` `chunk_id=srcchunk_4d8f62641434bd80a12f5bfd326c1512` `native_locator=slack:C0547N89JUB:1769579086.050139:1770274471.585089` `source_timestamp=2026-02-05T06:54:31Z`

## Open Questions

- Is using story-localnet as chain-id safe from potential sync interference with external localnets?
- Should internal-devnet-1 be added to the public story upgrade file?
- Will the old internal-devnet be fully deprecated after the decision next week?

## Sources

- `source_document_id`: `srcdoc_294f2516b3e9908895b36e310eae179d`
- `source_revision_id`: `srcrev_7c00ae2acc246936432c5c6852f6949a`
