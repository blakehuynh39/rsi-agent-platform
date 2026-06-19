---
title: "RSI Platform v2 Roadmap and Post-Release Priorities"
type: "project"
slug: "projects/rsi-platform-v2-roadmap"
freshness: "2026-04-30T13:08:06Z"
tags:
  - "priorities"
  - "roadmap"
  - "v2"
owners: []
source_revision_ids:
  - "srcrev_31c5342a2db8d723556ea415a86562bd"
  - "srcrev_99e82180725312551e0b0ca472469ff4"
  - "srcrev_cee641840701a738263f356febb5553f"
  - "srcrev_feabee125644229afb882adb6df553ea"
conflict_state: "none"
---

# RSI Platform v2 Roadmap and Post-Release Priorities

## Summary

Post-V1 release roadmap and priorities including seedphrase voice matching, multi-language support, world launch, native apps, payment integration, non-audio tasks, and analytics improvements.

## Claims

- The team celebrated the v1 release but acknowledged it is just the beginning and parallel workstreams are required. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_cee641840701a738263f356febb5553f` `chunk_id=srcchunk_7da0f15e10779f83638982f0e3280d3f` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777539361.184219` `source_timestamp=2026-04-30T08:56:01Z`
- [P0.5] Add back Seedphrase for voice matching, requiring transcript generation and re-recording every X submissions or Y days. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_cee641840701a738263f356febb5553f` `chunk_id=srcchunk_7da0f15e10779f83638982f0e3280d3f` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777539361.184219` `source_timestamp=2026-04-30T08:56:01Z`
- [P0.5] Add English, Vietnamese, Korean, Filipino voice support to increase user base; Poseidon vetting required. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_cee641840701a738263f356febb5553f` `chunk_id=srcchunk_7da0f15e10779f83638982f0e3280d3f` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777539361.184219` `source_timestamp=2026-04-30T08:56:01Z`
- [P0.5] World launch planned as a quick follow-up; Royce to confirm QA. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_cee641840701a738263f356febb5553f` `chunk_id=srcchunk_7da0f15e10779f83638982f0e3280d3f` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777539361.184219` `source_timestamp=2026-04-30T08:56:01Z`
- [P1] Native mobile apps to be launched on iOS and Android. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_cee641840701a738263f356febb5553f` `chunk_id=srcchunk_7da0f15e10779f83638982f0e3280d3f` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777539361.184219` `source_timestamp=2026-04-30T08:56:01Z`
- [P1] Payment provider implementation: ~90% of KLED pays in crypto stablecoins; fiat option to use easiest provider (Stripe/Whop/Paypal). `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_cee641840701a738263f356febb5553f` `chunk_id=srcchunk_7da0f15e10779f83638982f0e3280d3f` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777539361.184219` `source_timestamp=2026-04-30T08:56:01Z`
- [P1] Non-audio tasks to begin with Resume upload (possibly H2A); a list of tasks and design help for card images needed. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_cee641840701a738263f356febb5553f` `chunk_id=srcchunk_7da0f15e10779f83638982f0e3280d3f` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777539361.184219` `source_timestamp=2026-04-30T08:56:01Z`
- [P2] Customer Support feedback provider implementation is planned. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_cee641840701a738263f356febb5553f` `chunk_id=srcchunk_7da0f15e10779f83638982f0e3280d3f` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777539361.184219` `source_timestamp=2026-04-30T08:56:01Z`
- Ongoing: Marketing and ads targeting. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_cee641840701a738263f356febb5553f` `chunk_id=srcchunk_7da0f15e10779f83638982f0e3280d3f` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777539361.184219` `source_timestamp=2026-04-30T08:56:01Z`
- Ongoing: Work with Poseidon on other data needs; design help required for card images. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_cee641840701a738263f356febb5553f` `chunk_id=srcchunk_7da0f15e10779f83638982f0e3280d3f` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777539361.184219` `source_timestamp=2026-04-30T08:56:01Z`
- Ongoing: Monitoring traffic/users to adjust rewards and flag/ban malicious users. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_cee641840701a738263f356febb5553f` `chunk_id=srcchunk_7da0f15e10779f83638982f0e3280d3f` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777539361.184219` `source_timestamp=2026-04-30T08:56:01Z`
- Ongoing: Gamifying quests for multiplier bonus; may need design help for badges/icons. `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_cee641840701a738263f356febb5553f` `chunk_id=srcchunk_7da0f15e10779f83638982f0e3280d3f` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777539361.184219` `source_timestamp=2026-04-30T08:56:01Z`
- A wallet team member will explore decentralized/stablecoin native payment solutions for crypto payouts instead of Stripe. `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_feabee125644229afb882adb6df553ea` `chunk_id=srcchunk_51c3c75e16177627c6783399ede63ee5` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777551310.303389` `source_timestamp=2026-04-30T12:15:10Z`
- Combining Google Analytics properties is important and should be done ASAP. `claim:claim_1_14` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_99e82180725312551e0b0ca472469ff4` `chunk_id=srcchunk_8104d2e0d72902c998aab376f55f00a8` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777554434.089829` `source_timestamp=2026-04-30T13:07:14Z`
- A better analytics solution than GA (Posthog or Amplitude) is needed for mobile apps and is worth the cost. `claim:claim_1_15` `confidence:1.00`
  - citation: `source_document_id=srcdoc_688d4add0189f9211b47ac6faad4777b` `source_revision_id=srcrev_31c5342a2db8d723556ea415a86562bd` `chunk_id=srcchunk_95a4e0cf0cd51568231fe552a35ab615` `native_locator=slack:C0AL7EKNHDF:1777539361.184219:1777554486.871499` `source_timestamp=2026-04-30T13:08:06Z`

## Open Questions

- Complete list of non-audio tasks to surface and their design requirements.
- Design requirements for card images, badges, and icons across multiple workstreams.
- Exact timing and checklist for world launch (Royce to confirm QA).

## Sources

- `source_document_id`: `srcdoc_688d4add0189f9211b47ac6faad4777b`
- `source_revision_id`: `srcrev_31c5342a2db8d723556ea415a86562bd`
