---
title: "Onboarding Data Field Requirements"
type: "decision"
slug: "decisions/onboarding-data-field-requirements"
freshness: "2026-05-13T20:31:40Z"
tags:
  - "data-collection"
  - "mandatory-fields"
  - "onboarding"
owners: []
source_revision_ids:
  - "srcrev_f1bc349e0692657ad4577117832fd755"
conflict_state: "none"
---

# Onboarding Data Field Requirements

## Summary

Specifies which user profile fields are mandatory during onboarding. All fields except Occupation are required. Occupation will be collected later via CV upload and LinkedIn login.

## Claims

- During user onboarding, the fields Name, Gender, Languages spoken, Year of birth, and Nationality are all mandatory. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c24d224682e674bc62d30c9d418b8bed` `source_revision_id=srcrev_f1bc349e0692657ad4577117832fd755` `chunk_id=srcchunk_17ed6edb812dd4e99b689c8768f598c9` `native_locator=slack:C0AL7EKNHDF:1778703399.170799:1778704300.896039` `source_timestamp=2026-05-13T20:31:40Z`
- The Occupation field is not mandatory during onboarding; it will be collected later through separate tasks such as CV upload and LinkedIn login. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c24d224682e674bc62d30c9d418b8bed` `source_revision_id=srcrev_f1bc349e0692657ad4577117832fd755` `chunk_id=srcchunk_17ed6edb812dd4e99b689c8768f598c9` `native_locator=slack:C0AL7EKNHDF:1778703399.170799:1778704300.896039` `source_timestamp=2026-05-13T20:31:40Z`
- CV upload and LinkedIn login are planned onboarding features to collect Occupation details. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_c24d224682e674bc62d30c9d418b8bed` `source_revision_id=srcrev_f1bc349e0692657ad4577117832fd755` `chunk_id=srcchunk_17ed6edb812dd4e99b689c8768f598c9` `native_locator=slack:C0AL7EKNHDF:1778703399.170799:1778704300.896039` `source_timestamp=2026-05-13T20:31:40Z`

## Sources

- `source_document_id`: `srcdoc_c24d224682e674bc62d30c9d418b8bed`
- `source_revision_id`: `srcrev_f1bc349e0692657ad4577117832fd755`
