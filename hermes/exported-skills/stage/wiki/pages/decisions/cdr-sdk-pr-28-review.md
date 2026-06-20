---
title: "CDR SDK PR #28 Review"
type: "decision"
slug: "decisions/cdr-sdk-pr-28-review"
freshness: "2026-04-08T01:43:37Z"
tags:
  - "attestation"
  - "cdr-sdk"
  - "code-review"
  - "dkg"
  - "security"
owners:
  - "U07KLPN0JN6"
  - "U0AKJV8710S"
source_revision_ids:
  - "srcrev_1b9a4ece426fdc42319fa8621e475ac7"
  - "srcrev_6b29ba9239e53a54e19118c6af87ff2a"
  - "srcrev_d305eafe6fbc8783f18fe3562fd0fb1b"
  - "srcrev_e4af6cef29c4d9970d572374289146eb"
conflict_state: "none"
---

# CDR SDK PR #28 Review

## Summary

Automated review of PR #28 for piplabs/cdr-sdk identified key issues: DKG validator key lookback window may drop old validators, `verifyAttestation` is a no‑op, `skipHashCheck` needs prod guard, and WASM hash trust model requires documentation.

## Claims

- A review was requested for PR #28 in piplabs/cdr-sdk. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d4129d63bff44db2d4e179d29e6a4ec4` `source_revision_id=srcrev_1b9a4ece426fdc42319fa8621e475ac7` `chunk_id=srcchunk_74dc7d34ce2c6a7cac0427aaeca72ce9` `native_locator=slack:C0547N89JUB:1775611774.693859:1775611774.693859` `source_timestamp=2026-04-08T01:29:34Z`
- User U07KLPN0JN6 requested to be added as a reviewer for the PR review and was added. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d4129d63bff44db2d4e179d29e6a4ec4` `source_revision_id=srcrev_6b29ba9239e53a54e19118c6af87ff2a` `chunk_id=srcchunk_fec023242065b7f125ce292944458e47` `native_locator=slack:C0547N89JUB:1775611774.693859:1775611921.629329` `source_timestamp=2026-04-08T01:32:01Z`
  - citation: `source_document_id=srcdoc_d4129d63bff44db2d4e179d29e6a4ec4` `source_revision_id=srcrev_e4af6cef29c4d9970d572374289146eb` `chunk_id=srcchunk_583ae60d1180867bde39ae4ccaf8bea1` `native_locator=slack:C0547N89JUB:1775611774.693859:1775612506.751789` `source_timestamp=2026-04-08T01:41:46Z`
- Automated review posted for PR #28 with a TL;DR and detailed findings. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d4129d63bff44db2d4e179d29e6a4ec4` `source_revision_id=srcrev_d305eafe6fbc8783f18fe3562fd0fb1b` `chunk_id=srcchunk_c1c13d13ddcc2fac54b614d0af9d354c` `native_locator=slack:C0547N89JUB:1775611774.693859:1775612617.035209` `source_timestamp=2026-04-08T01:43:37Z`
- The 302,400‑block DKG validator key lookback window (~7 days) may cause validators registered earlier to be dropped as ‘unknown validator’ without a fallback. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d4129d63bff44db2d4e179d29e6a4ec4` `source_revision_id=srcrev_d305eafe6fbc8783f18fe3562fd0fb1b` `chunk_id=srcchunk_c1c13d13ddcc2fac54b614d0af9d354c` `native_locator=slack:C0547N89JUB:1775611774.693859:1775612617.035209` `source_timestamp=2026-04-08T01:43:37Z`
- attestation.ts exports verifyAttestation() but it is a no‑op that always returns valid: true, which is a footgun and should throw NotImplementedError or not be exported. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d4129d63bff44db2d4e179d29e6a4ec4` `source_revision_id=srcrev_d305eafe6fbc8783f18fe3562fd0fb1b` `chunk_id=srcchunk_c1c13d13ddcc2fac54b614d0af9d354c` `native_locator=slack:C0547N89JUB:1775611774.693859:1775612617.035209` `source_timestamp=2026-04-08T01:43:37Z`
- skipHashCheck bypass in the PR needs a production guard to prevent misuse. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d4129d63bff44db2d4e179d29e6a4ec4` `source_revision_id=srcrev_d305eafe6fbc8783f18fe3562fd0fb1b` `chunk_id=srcchunk_c1c13d13ddcc2fac54b614d0af9d354c` `native_locator=slack:C0547N89JUB:1775611774.693859:1775612617.035209` `source_timestamp=2026-04-08T01:43:37Z`
- The colocated WASM hash trust model must be documented to clarify security assumptions. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d4129d63bff44db2d4e179d29e6a4ec4` `source_revision_id=srcrev_d305eafe6fbc8783f18fe3562fd0fb1b` `chunk_id=srcchunk_c1c13d13ddcc2fac54b614d0af9d354c` `native_locator=slack:C0547N89JUB:1775611774.693859:1775612617.035209` `source_timestamp=2026-04-08T01:43:37Z`

## Open Questions

- Is skipHashCheck intended only for development, and how will it be guarded in production?
- Is the 7‑day DKG validator key lookback window intentional? Is there a fallback for older validators?
- Should verifyAttestation() be removed or throw NotImplementedError?
- Where is the colocated WASM hash trust model documented?

## Sources

- `source_document_id`: `srcdoc_d4129d63bff44db2d4e179d29e6a4ec4`
- `source_revision_id`: `srcrev_d305eafe6fbc8783f18fe3562fd0fb1b`
