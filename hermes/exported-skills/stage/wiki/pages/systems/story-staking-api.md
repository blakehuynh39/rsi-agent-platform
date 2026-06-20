---
title: "Story Staking API"
type: "system"
slug: "systems/story-staking-api"
freshness: "2026-04-07T09:59:46Z"
tags:
  - "devnet"
  - "docker"
  - "service"
  - "staking-api"
owners:
  - "Staking Team"
source_revision_ids:
  - "srcrev_56c62523a5517e728304e174112321ff"
  - "srcrev_62cfb8254f935489a9c63fe04d62b239"
  - "srcrev_6fd7e0081cccc958e5c69f20ccbe1554"
conflict_state: "none"
---

# Story Staking API

## Summary

The Staking API service, deployed via Docker containers.  The Docker image is built from a public repository and pushed to a container registry, then pulled by an AWS EC2 instance.

## Claims

- The Staking API is deployed using a Docker image built and pushed via GitHub Actions. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_56c62523a5517e728304e174112321ff` `chunk_id=srcchunk_2548bd4d4cf2d9e4bea340c0295b5def` `native_locator=slack:C0547N89JUB:1775552711.299739:1775552711.299739` `source_timestamp=2026-04-07T09:08:46Z`
- The service is intended to be pulled from AWS ECR, but GHCR is used temporarily for devnet. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_62cfb8254f935489a9c63fe04d62b239` `chunk_id=srcchunk_7ba63361f3367c61b51b25e2e7932850` `native_locator=slack:C0547N89JUB:1775552711.299739:1775553994.937329` `source_timestamp=2026-04-07T09:26:34Z`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_6fd7e0081cccc958e5c69f20ccbe1554` `chunk_id=srcchunk_22df64c7e54593dd4c34dd9ffaf0fd9c` `native_locator=slack:C0547N89JUB:1775552711.299739:1775555986.382049` `source_timestamp=2026-04-07T09:59:46Z`
- The EC2 instance has no GitHub authentication, so private registry images cannot be pulled. `claim:claim_2_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_56c62523a5517e728304e174112321ff` `chunk_id=srcchunk_2548bd4d4cf2d9e4bea340c0295b5def` `native_locator=slack:C0547N89JUB:1775552711.299739:1775552711.299739` `source_timestamp=2026-04-07T09:08:46Z`

## Related Pages

- `use-public-ghcr-for-devnet-staking-api-deployment`

## Sources

- `source_document_id`: `srcdoc_b216aabd787018439adf2aac6a1128f0`
- `source_revision_id`: `srcrev_6fd7e0081cccc958e5c69f20ccbe1554`
