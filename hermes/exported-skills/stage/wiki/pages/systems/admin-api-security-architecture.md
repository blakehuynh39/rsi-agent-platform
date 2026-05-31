---
title: "Admin API Security Architecture"
type: "system"
slug: "systems/admin-api-security-architecture"
freshness: "2023-03-30T04:08:00Z"
tags:
  - "admin-api"
  - "authentication"
  - "kms"
  - "secret-manager"
  - "security"
owners: []
source_revision_ids:
  - "srcrev_8c9ea9ad6e41260fbb9682e600cfe069"
conflict_state: "none"
---

# Admin API Security Architecture

## Summary

Describes the authentication mechanism for Admin APIs used by Streamer and Admin Server to access API Server's Admin APIs. It uses a combination of Secret Manager and KMS to encrypt a shared message, providing double protection and simplicity.

## Claims

- Both Streamer and Admin Server get access to API Server’s Admin APIs by adding an encrypted message to the request’s header. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Admin-API-Security-Architecture-7094af0ac31944ac967a6c9fa1d2db2b) `source_document_id=srcdoc_25add0388afa844b4c33f7821502f8b2` `source_revision_id=srcrev_8c9ea9ad6e41260fbb9682e600cfe069` `chunk_id=srcchunk_9ff6cf8f0f09918f9e04dcc905fb6269` `native_locator=https://www.notion.so/Admin-API-Security-Architecture-7094af0ac31944ac967a6c9fa1d2db2b` `source_timestamp=2023-03-30T04:08:00Z`
- To obtain the encrypted message, both services need to call secret manager at startup to get the original unencrypted message. Then when there is a trigger to call an admin API, they will call KMS to encrypt the unencrypted message and send the message with the API request. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Admin-API-Security-Architecture-7094af0ac31944ac967a6c9fa1d2db2b) `source_document_id=srcdoc_25add0388afa844b4c33f7821502f8b2` `source_revision_id=srcrev_8c9ea9ad6e41260fbb9682e600cfe069` `chunk_id=srcchunk_9ff6cf8f0f09918f9e04dcc905fb6269` `native_locator=https://www.notion.so/Admin-API-Security-Architecture-7094af0ac31944ac967a6c9fa1d2db2b` `source_timestamp=2023-03-30T04:08:00Z`
- On API Server side, the authentication process is: 1. At service startup, API Server calls Secret Manager to obtain the unencrypted message. 2. When an admin API request comes in, the API Server checks the auth header to get the encrypted message if it’s present. 3. API server then calls KMS to decrypt the message. 4. If decryption succeeds, API server compares the decrypted message with the unencrypted message fetched from secret manager to make sure they match. 5. Any error in the above steps will fail the authentication. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Admin-API-Security-Architecture-7094af0ac31944ac967a6c9fa1d2db2b) `source_document_id=srcdoc_25add0388afa844b4c33f7821502f8b2` `source_revision_id=srcrev_8c9ea9ad6e41260fbb9682e600cfe069` `chunk_id=srcchunk_9ff6cf8f0f09918f9e04dcc905fb6269` `native_locator=https://www.notion.so/Admin-API-Security-Architecture-7094af0ac31944ac967a6c9fa1d2db2b` `source_timestamp=2023-03-30T04:08:00Z`
- The architecture provides double protection: an attacker needs to breach both secret manager and KMS to create the encrypted messages. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Admin-API-Security-Architecture-7094af0ac31944ac967a6c9fa1d2db2b) `source_document_id=srcdoc_25add0388afa844b4c33f7821502f8b2` `source_revision_id=srcrev_8c9ea9ad6e41260fbb9682e600cfe069` `chunk_id=srcchunk_9ff6cf8f0f09918f9e04dcc905fb6269` `native_locator=https://www.notion.so/Admin-API-Security-Architecture-7094af0ac31944ac967a6c9fa1d2db2b` `source_timestamp=2023-03-30T04:08:00Z`
- The architecture uses a simple message structure and symmetrical key for encrypt/decrypt, as not a lot of data is needed for authentication. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Admin-API-Security-Architecture-7094af0ac31944ac967a6c9fa1d2db2b) `source_document_id=srcdoc_25add0388afa844b4c33f7821502f8b2` `source_revision_id=srcrev_8c9ea9ad6e41260fbb9682e600cfe069` `chunk_id=srcchunk_9ff6cf8f0f09918f9e04dcc905fb6269` `native_locator=https://www.notion.so/Admin-API-Security-Architecture-7094af0ac31944ac967a6c9fa1d2db2b` `source_timestamp=2023-03-30T04:08:00Z`

## Sources

- `source_document_id`: `srcdoc_25add0388afa844b4c33f7821502f8b2`
- `source_revision_id`: `srcrev_8c9ea9ad6e41260fbb9682e600cfe069`
- `source_url`: [Notion source](https://www.notion.so/Admin-API-Security-Architecture-7094af0ac31944ac967a6c9fa1d2db2b)
