---
title: "Subnet Pipeline Attack Vectors"
type: "concept"
slug: "concepts/subnet-pipeline-attack-vectors"
freshness: "2025-10-30T06:07:00Z"
tags:
  - "attack-vectors"
  - "miner"
  - "security"
  - "subnet"
  - "validator"
owners: []
source_revision_ids:
  - "srcrev_af1e3375b5419c211d76c42e83682c8a"
conflict_state: "none"
---

# Subnet Pipeline Attack Vectors

## Summary

Documented attack vectors for the subnet pipeline, covering both miner and validator roles. Miner attacks include bypassing transcription by using the reference script and skipping transcription/translation to score with RNG or preference. Validator attacks include copying another validator's work with slight modifications.

## Claims

- A miner attack vector is to use the reference script's text as the output of the transcription instead of transcribing the given raw audio file, effectively scoring the reference script against itself and bypassing transcription logic. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Pipeline-Attack-Vectors-29a051299a5480928918c11d39eec6b9) `source_document_id=srcdoc_01e3beb9e649788a17f3251adcaf1154` `source_revision_id=srcrev_af1e3375b5419c211d76c42e83682c8a` `chunk_id=srcchunk_4f4ad4889b08f116fd4a77ed41693a7c` `native_locator=https://www.notion.so/Subnet-Pipeline-Attack-Vectors-29a051299a5480928918c11d39eec6b9` `source_timestamp=2025-10-30T06:07:00Z`
- A miner attack vector is to skip transcription and translation entirely and simply score the audio with RNG or preference. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Pipeline-Attack-Vectors-29a051299a5480928918c11d39eec6b9) `source_document_id=srcdoc_01e3beb9e649788a17f3251adcaf1154` `source_revision_id=srcrev_af1e3375b5419c211d76c42e83682c8a` `chunk_id=srcchunk_4f4ad4889b08f116fd4a77ed41693a7c` `native_locator=https://www.notion.so/Subnet-Pipeline-Attack-Vectors-29a051299a5480928918c11d39eec6b9` `source_timestamp=2025-10-30T06:07:00Z`
- A validator attack vector is for one validator to wait for another validator to complete and submit validation work, then use that work to complete its own assigned validation activity, potentially with slight modifications to obfuscate the copying. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Pipeline-Attack-Vectors-29a051299a5480928918c11d39eec6b9) `source_document_id=srcdoc_01e3beb9e649788a17f3251adcaf1154` `source_revision_id=srcrev_af1e3375b5419c211d76c42e83682c8a` `chunk_id=srcchunk_4f4ad4889b08f116fd4a77ed41693a7c` `native_locator=https://www.notion.so/Subnet-Pipeline-Attack-Vectors-29a051299a5480928918c11d39eec6b9` `source_timestamp=2025-10-30T06:07:00Z`
- The proposed mitigation for the miner transcription bypass attack is to configure activities so that in the first activity of a workflow, the miner does not know the reference script content. The first activity is to produce a transcript and an English translation, the second activity is scoring the translated transcript against the reference script, and the third activity is IP registration if the score is valid. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Subnet-Pipeline-Attack-Vectors-29a051299a5480928918c11d39eec6b9) `source_document_id=srcdoc_01e3beb9e649788a17f3251adcaf1154` `source_revision_id=srcrev_af1e3375b5419c211d76c42e83682c8a` `chunk_id=srcchunk_4f4ad4889b08f116fd4a77ed41693a7c` `native_locator=https://www.notion.so/Subnet-Pipeline-Attack-Vectors-29a051299a5480928918c11d39eec6b9` `source_timestamp=2025-10-30T06:07:00Z`

## Sources

- `source_document_id`: `srcdoc_01e3beb9e649788a17f3251adcaf1154`
- `source_revision_id`: `srcrev_af1e3375b5419c211d76c42e83682c8a`
- `source_url`: [Notion source](https://www.notion.so/Subnet-Pipeline-Attack-Vectors-29a051299a5480928918c11d39eec6b9)
