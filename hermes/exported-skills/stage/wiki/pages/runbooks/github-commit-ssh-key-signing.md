---
title: "Runbook: Secure GitHub Commits with SSH Key Signing"
type: "runbook"
slug: "runbooks/github-commit-ssh-key-signing"
freshness: "2025-11-12T19:24:00Z"
tags:
  - "commit-signing"
  - "git"
  - "github"
  - "security"
  - "ssh"
owners:
  - "andy@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_daf5a184ef23a8f9a3eacc6eea0f9bd9"
conflict_state: "none"
---

# Runbook: Secure GitHub Commits with SSH Key Signing

## Summary

Step-by-step guide to set up SSH key signing for GitHub commits, including key generation, agent setup, GitHub configuration, and troubleshooting unverified commits.

## Claims

- Signed commits prove the authenticity of contributions, ensuring code comes from the real author and not an impersonator. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-1) `source_document_id=srcdoc_47cc174b4e5b17c11409d090ab688389` `source_revision_id=srcrev_daf5a184ef23a8f9a3eacc6eea0f9bd9` `chunk_id=srcchunk_b759ebdd3c20984d15bd9615261e2e58` `native_locator=https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-1` `source_timestamp=2025-11-12T19:24:00Z`
- Using separate SSH keys for authentication and commit signing is a best practice for security isolation, purpose separation, and ease of rotation. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-1) `source_document_id=srcdoc_47cc174b4e5b17c11409d090ab688389` `source_revision_id=srcrev_daf5a184ef23a8f9a3eacc6eea0f9bd9` `chunk_id=srcchunk_b759ebdd3c20984d15bd9615261e2e58` `native_locator=https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-1` `source_timestamp=2025-11-12T19:24:00Z`
- Git version 2.34.0 or higher is required to use SSH key signing. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-1) `source_document_id=srcdoc_47cc174b4e5b17c11409d090ab688389` `source_revision_id=srcrev_daf5a184ef23a8f9a3eacc6eea0f9bd9` `chunk_id=srcchunk_b759ebdd3c20984d15bd9615261e2e58` `native_locator=https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-1` `source_timestamp=2025-11-12T19:24:00Z`
- Generate a new Ed25519 SSH key for signing with the command: ssh-keygen -t ed25519 -C "andy@storyprotocol.xyz" -f ~/.ssh/id_ed25519_signing `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-1) `source_document_id=srcdoc_47cc174b4e5b17c11409d090ab688389` `source_revision_id=srcrev_daf5a184ef23a8f9a3eacc6eea0f9bd9` `chunk_id=srcchunk_b759ebdd3c20984d15bd9615261e2e58` `native_locator=https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-1` `source_timestamp=2025-11-12T19:24:00Z`
- Add the signing key to the SSH agent using: eval "$(ssh-agent -s)" and ssh-add ~/.ssh/id_ed25519_signing `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-1) `source_document_id=srcdoc_47cc174b4e5b17c11409d090ab688389` `source_revision_id=srcrev_daf5a184ef23a8f9a3eacc6eea0f9bd9` `chunk_id=srcchunk_b759ebdd3c20984d15bd9615261e2e58` `native_locator=https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-1` `source_timestamp=2025-11-12T19:24:00Z`
- Add the public signing key to GitHub at https://github.com/settings/ssh/new, selecting "Signing Key" as the Key Type. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-1) `source_document_id=srcdoc_47cc174b4e5b17c11409d090ab688389` `source_revision_id=srcrev_daf5a184ef23a8f9a3eacc6eea0f9bd9` `chunk_id=srcchunk_b759ebdd3c20984d15bd9615261e2e58` `native_locator=https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-1` `source_timestamp=2025-11-12T19:24:00Z`
- Commits show as "Unverified" on GitHub if the commit email does not match the primary email registered with the GitHub account. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-2) `source_document_id=srcdoc_47cc174b4e5b17c11409d090ab688389` `source_revision_id=srcrev_daf5a184ef23a8f9a3eacc6eea0f9bd9` `chunk_id=srcchunk_727079e2d484a8027be1ff900a566469` `native_locator=https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-2` `source_timestamp=2025-11-12T19:24:00Z`
- The "No principal matched" error when signing a commit can be resolved by ensuring the public key is added to GitHub as a Signing Key and the allowed_signers file is populated with the correct email and public key. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-2) `source_document_id=srcdoc_47cc174b4e5b17c11409d090ab688389` `source_revision_id=srcrev_daf5a184ef23a8f9a3eacc6eea0f9bd9` `chunk_id=srcchunk_727079e2d484a8027be1ff900a566469` `native_locator=https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-2` `source_timestamp=2025-11-12T19:24:00Z`
- Using the same SSH key for both authentication and signing is technically possible but not recommended due to security and key rotation concerns. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-2) `source_document_id=srcdoc_47cc174b4e5b17c11409d090ab688389` `source_revision_id=srcrev_daf5a184ef23a8f9a3eacc6eea0f9bd9` `chunk_id=srcchunk_727079e2d484a8027be1ff900a566469` `native_locator=https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f#chunk-2` `source_timestamp=2025-11-12T19:24:00Z`

## Sources

- `source_document_id`: `srcdoc_47cc174b4e5b17c11409d090ab688389`
- `source_revision_id`: `srcrev_daf5a184ef23a8f9a3eacc6eea0f9bd9`
- `source_url`: [Notion source](https://www.notion.so/HOW-TO-Secure-GitHub-Commits-with-SSH-Key-Signing-10e051299a548000ac77d42c8802c29f)
