---
title: "Numo Mobile Team Onboarding and Resource Provisioning"
type: "project"
slug: "projects/numo-mobile-team-onboarding"
freshness: "2026-06-01T16:40:32Z"
tags:
  - "access"
  - "accounts"
  - "mobile"
  - "numo"
  - "onboarding"
owners:
  - "@U04L0DD6B6F"
  - "@U07TNT9N4JC"
  - "@U08332YRB7W"
  - "@U0AG7GU6WBU"
source_revision_ids:
  - "srcrev_16c01c81f83fb3d30621730031d11210"
  - "srcrev_1acf6bc05b5bf866be264be5f352febe"
  - "srcrev_1c571706aa4e28b1ee1d1ce525578e5c"
  - "srcrev_1f2f3f884a0463077d993df8e84ba00b"
  - "srcrev_1f9998f7ff3726d8eb2b04575efd665d"
  - "srcrev_31f46c53aa0860aa148e8e3575c929bd"
  - "srcrev_88268e75515a0528f221107e1edb0193"
  - "srcrev_8fb10e83fee58210c5c714910d8d2810"
  - "srcrev_ab0e3cd5aeb3f169fb99da984f98d4b3"
  - "srcrev_b682c339ed37c532cd759a6a31c4de3a"
  - "srcrev_c2313ff37ed77f85fcc73220f410b9ac"
  - "srcrev_c8e38110ecd4e7d3ccbef95f0c65a7cc"
  - "srcrev_d6416646a2ea60949efbcbdc1c2b253c"
  - "srcrev_e78d85633996d111c3802692c5363b75"
  - "srcrev_f8c39f38f12f4d25dae8d564ef5af3bd"
conflict_state: "none"
---

# Numo Mobile Team Onboarding and Resource Provisioning

## Summary

Onboarding the Numo mobile team (Jed Lee and colleagues) to RSI infrastructure, including provisioning cloud, app store, marketing SDK accounts, email/Slack access, and defining billing processes. The initiative aims to unblock their development ASAP.

## Claims

- Numo mobile team requires cloud (AWS/GCP), app store accounts (Apple, Google Play), and third-party mobile marketing SDK accounts (Adjust, OneSignal, Meta Developer, Google Ads, Amplitude) as well as miscellaneous engineering tools (Infisical, Dagster). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_f8c39f38f12f4d25dae8d564ef5af3bd` `chunk_id=srcchunk_7827ce47473a0fe7b15db06b0dc99411` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780025189.061619` `source_timestamp=2026-05-29T03:26:29Z`
- Cloud and app store accounts can be provided from the RSI side; handling login and payments for other platforms needs to be decided. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_f8c39f38f12f4d25dae8d564ef5af3bd` `chunk_id=srcchunk_7827ce47473a0fe7b15db06b0dc99411` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780025189.061619` `source_timestamp=2026-05-29T03:26:29Z`
- The resources are needed ASAP to unblock Numo mobile development. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_f8c39f38f12f4d25dae8d564ef5af3bd` `chunk_id=srcchunk_7827ce47473a0fe7b15db06b0dc99411` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780025189.061619` `source_timestamp=2026-05-29T03:26:29Z`
- RSI functionality may need to be extended to multi-tenant for Poseidon slack access. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_8fb10e83fee58210c5c714910d8d2810` `chunk_id=srcchunk_2669461bf0a371532bae982032d666f1` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780031836.838159` `source_timestamp=2026-05-29T05:17:29Z`
- GCP is not required; AWS is acceptable for cloud infrastructure. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_1f2f3f884a0463077d993df8e84ba00b` `chunk_id=srcchunk_f1c7ae385750b574bd2d06ab8bd3fb9f` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780031984.232579` `source_timestamp=2026-05-29T05:19:44Z`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_b682c339ed37c532cd759a6a31c4de3a` `chunk_id=srcchunk_7c9cc548ddb98c5901ffb962695e5974` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780054936.183529` `source_timestamp=2026-05-29T11:42:16Z`
- Most mobile marketing SDK and miscellaneous engineering accounts are new and need to be created either by the company and team invited, or by the team with company billing. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_c2313ff37ed77f85fcc73220f410b9ac` `chunk_id=srcchunk_56a590d9110cfd958af6fc746d75251d` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780117782.769969` `source_timestamp=2026-05-30T05:09:42Z`
- For billing, creating a Brex card for the team or having them pay and get reimbursed was suggested. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_1acf6bc05b5bf866be264be5f352febe` `chunk_id=srcchunk_b86de635097d2324f06fc2c4e9556509` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780127947.018209` `source_timestamp=2026-05-30T07:59:07Z`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_88268e75515a0528f221107e1edb0193` `chunk_id=srcchunk_3230c628586805a326e76b029547e168` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780129971.028179` `source_timestamp=2026-05-30T08:32:51Z`
- Making RSI multi-tenant is currently difficult; instead, installing RSI for Poseidon with AWS/k8 permissions is possible but will not have cross-company domain knowledge. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_16c01c81f83fb3d30621730031d11210` `chunk_id=srcchunk_ae00f53c91bf26942f1b053d5ce7b7db` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780241554.077919` `source_timestamp=2026-05-31T15:32:34Z`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_d6416646a2ea60949efbcbdc1c2b253c` `chunk_id=srcchunk_39276c992da5392977597c369500e583` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780241575.408909` `source_timestamp=2026-05-31T15:32:55Z`
- Team member emails to be created under psdn.ai domain: jed@, mason@, elise@, sasha@, daniel.jung@, holmes@, june.son@ `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_31f46c53aa0860aa148e8e3575c929bd` `chunk_id=srcchunk_5f9ce25f914136b10539119257144c09` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780288173.227339` `source_timestamp=2026-06-01T04:29:33Z`
- After email creation, invite them to Poseidon Slack and create a channel #ext-story-kt in Story Slack. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_31f46c53aa0860aa148e8e3575c929bd` `chunk_id=srcchunk_5f9ce25f914136b10539119257144c09` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780288173.227339` `source_timestamp=2026-06-01T04:29:33Z`
- A high-level onboarding guide and a credentials/access checklist have been created on Notion for the Numo team. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_c8e38110ecd4e7d3ccbef95f0c65a7cc` `chunk_id=srcchunk_de074c6b6721122403ee2b0d1fbbb573` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780288561.949269` `source_timestamp=2026-06-01T04:37:29Z`
- Access to resources must be restricted; having a psdn.ai email grants access to public Poseidon Notion and Google docs/slides. `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_1c571706aa4e28b1ee1d1ce525578e5c` `chunk_id=srcchunk_a0fb044a78605a26d9cbcbdc65b352bd` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780127006.459599` `source_timestamp=2026-05-30T07:43:26Z`
- A shared email numo@psdn.ai is requested for sharing services through 1Password. `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_e78d85633996d111c3802692c5363b75` `chunk_id=srcchunk_2cae13d8584a0e22238ce7069cbdb834` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780322165.370259` `source_timestamp=2026-06-01T13:56:05Z`
- The onboarding AWS access, restricted Notion, and GitHub access are to be provisioned for the team. `claim:claim_1_14` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_1f9998f7ff3726d8eb2b04575efd665d` `chunk_id=srcchunk_272796c841503d9ff11300c5d4c89f28` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780332032.191859` `source_timestamp=2026-06-01T16:40:32Z`
- Emails are planned to be created by @U08332YRB7W and @U07TNT9N4JC, ideally before a Thursday KST meeting (Wednesday PST). `claim:claim_1_15` `confidence:1.00`
  - citation: `source_document_id=srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5` `source_revision_id=srcrev_ab0e3cd5aeb3f169fb99da984f98d4b3` `chunk_id=srcchunk_084e3487ecf18b89791f5532a8930c08` `native_locator=slack:C0AL7EKNHDF:1780025189.061619:1780331749.971829` `source_timestamp=2026-06-01T16:35:49Z`

## Open Questions

- How should login accounts and payments for third-party marketing SDKs and engineering tools be handled?
- Is a Brex card or reimbursement the finalized billing method?
- What specific restrictions will be placed on the team's access to Poseidon Notion and Google docs?
- Will RSI be installed for Poseidon or will multi-tenancy be developed later?

## Sources

- `source_document_id`: `srcdoc_f00a3d7cd3511e1e41d407b743f2b6d5`
- `source_revision_id`: `srcrev_7d4842f86aebdf60c8fe318c74712cf0`
