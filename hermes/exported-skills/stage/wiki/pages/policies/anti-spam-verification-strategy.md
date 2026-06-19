---
title: "Anti-Spam Verification Strategy"
type: "policy"
slug: "policies/anti-spam-verification-strategy"
freshness: "2026-05-04T23:55:12Z"
tags:
  - "anti-fraud"
  - "mobile-app"
  - "spam"
  - "verification"
owners: []
source_revision_ids:
  - "srcrev_095604412a2c2e1f6a630046c8872a81"
  - "srcrev_5c88d210872098e97389dfe782dffaf7"
  - "srcrev_68cebdbd292efa116e067a6d4b74337c"
  - "srcrev_8158321b56ac2b4e0141d1ab970b7602"
  - "srcrev_84195b5b5fd2ea87c5aac5366b6f1c3d"
  - "srcrev_86b93dc2397a29ae2ec8da9d9dee6b3e"
  - "srcrev_985391f211ea38d6b03881f1f81b35ef"
  - "srcrev_e56ff86f8ee15ea14ca8b7e3c346ed90"
conflict_state: "none"
---

# Anti-Spam Verification Strategy

## Summary

Strategy to combat spam submissions, using light touch methods (voice fingerprinting, language test) and escalating to payment block, informed by Pi Network and WeChat approaches, currently implemented via admin dashboard with future integration of Elevenlabs voice verification.

## Claims

- Some users are submitting random English speech to Tengulu submissions that should be in a specific language. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2651109b4aa27a3b604a26c102fd7ebc` `source_revision_id=srcrev_84195b5b5fd2ea87c5aac5366b6f1c3d` `chunk_id=srcchunk_dfab70e421d04bd04e32b2564e1b7e11` `native_locator=slack:C0AL7EKNHDF:1777738278.018759:1777738278.018759` `source_timestamp=2026-05-02T16:11:18Z`
- We need to figure out a verification solution before launching the mobile app and enabling payments, as a priority to save money by not paying infringing users. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2651109b4aa27a3b604a26c102fd7ebc` `source_revision_id=srcrev_985391f211ea38d6b03881f1f81b35ef` `chunk_id=srcchunk_f409d6b93e2645da9415ffdc457fee8a` `native_locator=slack:C0AL7EKNHDF:1777738278.018759:1777739031.182209` `source_timestamp=2026-05-02T16:24:14Z`
- A suggestion is to do a real-time local voice fingerprint check with a 30-second language test, rejecting those who fail. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2651109b4aa27a3b604a26c102fd7ebc` `source_revision_id=srcrev_86b93dc2397a29ae2ec8da9d9dee6b3e` `chunk_id=srcchunk_aa0ca6388e1b30a6c2572d7c7915e0af` `native_locator=slack:C0AL7EKNHDF:1777738278.018759:1777843800.587089` `source_timestamp=2026-05-03T21:30:00Z`
- Pi Network used a trust graph with vouching rewards and pyramid verification with hand-selected verifiers, but eventually launched without verifying everyone, leaving unverified users without rewards. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2651109b4aa27a3b604a26c102fd7ebc` `source_revision_id=srcrev_5c88d210872098e97389dfe782dffaf7` `chunk_id=srcchunk_cd25ff410f21e44b4258ed9ffa680319` `native_locator=slack:C0AL7EKNHDF:1777738278.018759:1777898833.473719` `source_timestamp=2026-05-04T12:47:13Z`
- WeChat has complex KYC mechanisms to prevent spammers, with free accounts having limited access and full KYC required to bind a bank card for payments. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2651109b4aa27a3b604a26c102fd7ebc` `source_revision_id=srcrev_68cebdbd292efa116e067a6d4b74337c` `chunk_id=srcchunk_f282b450b3b1c142f0c8b977d5f10438` `native_locator=slack:C0AL7EKNHDF:1777738278.018759:1777910058.152539` `source_timestamp=2026-05-04T15:54:18Z`
- The team has added a risk section under users with data points for spammer triage, and plans to associate device IDs when the mobile app launches. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2651109b4aa27a3b604a26c102fd7ebc` `source_revision_id=srcrev_095604412a2c2e1f6a630046c8872a81` `chunk_id=srcchunk_c2c100a6d250c9612d5c494c08796028` `native_locator=slack:C0AL7EKNHDF:1777738278.018759:1777910341.164939` `source_timestamp=2026-05-04T15:59:01Z`
- Chinese farming operations may have ways to circumvent app-based verification; valid phone numbers via WeChat/Ali KYC are central to identity fraud. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2651109b4aa27a3b604a26c102fd7ebc` `source_revision_id=srcrev_e56ff86f8ee15ea14ca8b7e3c346ed90` `chunk_id=srcchunk_91c1dab2e04493b1045c6a5ed29f4834` `native_locator=slack:C0AL7EKNHDF:1777738278.018759:1777938606.979779` `source_timestamp=2026-05-04T23:50:06Z`
- Current plan is to monitor via admin dashboard and add light validation from Elevenlabs at low cost (~$0.22/hour, ~$20 for current users), blocking withdrawal but not submission, with appeal option. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2651109b4aa27a3b604a26c102fd7ebc` `source_revision_id=srcrev_8158321b56ac2b4e0141d1ab970b7602` `chunk_id=srcchunk_5be644607f6ea91d87fd050916595e30` `native_locator=slack:C0AL7EKNHDF:1777738278.018759:1777938912.112099` `source_timestamp=2026-05-04T23:55:12Z`

## Open Questions

- What is the optimal balance between friction and fraud prevention for user onboarding?

## Sources

- `source_document_id`: `srcdoc_2651109b4aa27a3b604a26c102fd7ebc`
- `source_revision_id`: `srcrev_8158321b56ac2b4e0141d1ab970b7602`
