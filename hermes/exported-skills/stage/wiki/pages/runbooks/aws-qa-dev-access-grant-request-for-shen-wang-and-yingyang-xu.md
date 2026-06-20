---
title: "AWS QA Dev Access Grant Request for shen.wang and yingyang.xu"
type: "runbook"
slug: "runbooks/aws-qa-dev-access-grant-request-for-shen-wang-and-yingyang-xu"
freshness: "2026-04-07T07:12:46Z"
tags:
  - "access-request"
  - "aws"
  - "iam"
  - "terraform"
owners:
  - "woojin.kim@piplabs.xyz"
  - "yao.wang@piplabs.xyz"
source_revision_ids:
  - "srcrev_0b920441152b200e896a0f4cdb300aab"
  - "srcrev_2021ea92e59fe4af27ccc91d90887cb8"
  - "srcrev_24d481d71443106a73f52d8987083c58"
  - "srcrev_2a67aca781edd86fb391da25dded94b3"
  - "srcrev_4d70ed1a54ee260d89475ee08261b135"
  - "srcrev_ef762e106cae9c948cbb4eed4bb57c20"
  - "srcrev_fbd873fe1cd20f27001fbf989c8a5803"
conflict_state: "none"
---

# AWS QA Dev Access Grant Request for shen.wang and yingyang.xu

## Summary

Yao Wang requested AWS-QA-Dev-Access role for shen.wang@piplabs.xyz and yingyang.xu@piplabs.xyz, assigned P1 priority. IT admin was offline. A Terraform PR (storyprotocol/AWS-Organization#24) was created adding the users to accounts/story-dev-ou/sso.tf. The diff was verified as additive with no destroys, but Terraform planning failed, possibly due to a trailing comma or other syntax issue. The IT admin needs to review and fix before merging.

## Claims

- Yao Wang requested AWS-QA-Dev-Access for shen.wang@piplabs.xyz and yingyang.xu@piplabs.xyz. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dacf64b9fe120ef26de07c2e06f5d632` `source_revision_id=srcrev_2a67aca781edd86fb391da25dded94b3` `chunk_id=srcchunk_1984b3de87443ad810d499d486eef108` `native_locator=slack:C0547N89JUB:1775545108.430009:1775545108.430009` `source_timestamp=2026-04-07T06:58:28Z`
- The request was prioritized as P1 (High / needed ASAP). `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dacf64b9fe120ef26de07c2e06f5d632` `source_revision_id=srcrev_4d70ed1a54ee260d89475ee08261b135` `chunk_id=srcchunk_b8dab53fc36c1f9534d7baa11315edd3` `native_locator=slack:C0547N89JUB:1775545108.430009:1775545190.433109` `source_timestamp=2026-04-07T06:59:50Z`
- IT admin was offline and expected back at 9:00 AM Los Angeles time. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dacf64b9fe120ef26de07c2e06f5d632` `source_revision_id=srcrev_0b920441152b200e896a0f4cdb300aab` `chunk_id=srcchunk_83b754fad38149f53949de47bb23fd35` `native_locator=slack:C0547N89JUB:1775545108.430009:1775545158.957969` `source_timestamp=2026-04-07T06:59:18Z`
- A Terraform PR (storyprotocol/AWS-Organization#24) was created to add shen.wang and yingyang.xu to the user_names list in accounts/story-dev-ou/sso.tf for the AWS-QA-Dev-Access permission set. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dacf64b9fe120ef26de07c2e06f5d632` `source_revision_id=srcrev_24d481d71443106a73f52d8987083c58` `chunk_id=srcchunk_57c5c903583c6efc288aafe497205462` `native_locator=slack:C0547N89JUB:1775545108.430009:1775545817.264449` `source_timestamp=2026-04-07T07:10:17Z`
- The PR diff is a pure additive change with no destroys or modifications to existing users; only two new user entries are appended. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dacf64b9fe120ef26de07c2e06f5d632` `source_revision_id=srcrev_fbd873fe1cd20f27001fbf989c8a5803` `chunk_id=srcchunk_f48465f7a53aa212414f8630412684b4` `native_locator=slack:C0547N89JUB:1775545108.430009:1775545929.508809` `source_timestamp=2026-04-07T07:11:52Z`
- Terraform planning on the PR failed with 'Planning failed' error; Yao Wang suggested a trailing comma might be the issue. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dacf64b9fe120ef26de07c2e06f5d632` `source_revision_id=srcrev_2021ea92e59fe4af27ccc91d90887cb8` `chunk_id=srcchunk_64be7b60f2e2fb5be6d92ac4bc122388` `native_locator=slack:C0547N89JUB:1775545108.430009:1775545957.108659` `source_timestamp=2026-04-07T07:12:37Z`
- The bot noted that HCL generally allows trailing commas, so the failure might be due to another syntax error or a JSON context where trailing commas are not allowed; the IT admin needs to review and fix. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_dacf64b9fe120ef26de07c2e06f5d632` `source_revision_id=srcrev_ef762e106cae9c948cbb4eed4bb57c20` `chunk_id=srcchunk_542d5374cd930bfcff759d64708642c4` `native_locator=slack:C0547N89JUB:1775545108.430009:1775545966.704479` `source_timestamp=2026-04-07T07:12:46Z`

## Open Questions

- Did the PR get merged and access granted?
- What was the final resolution for the Terraform plan failure?
- Who is the IT admin? What is their availability?

## Sources

- `source_document_id`: `srcdoc_dacf64b9fe120ef26de07c2e06f5d632`
- `source_revision_id`: `srcrev_dab1c38646b1edda750fb9756f86dedd`
