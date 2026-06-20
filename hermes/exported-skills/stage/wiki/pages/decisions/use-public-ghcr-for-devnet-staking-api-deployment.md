---
title: "Use Public GitHub Container Registry for Devnet Staking API Deployment"
type: "decision"
slug: "decisions/use-public-ghcr-for-devnet-staking-api-deployment"
freshness: "2026-04-07T09:59:46Z"
tags:
  - "container-registry"
  - "deployment"
  - "devnet"
  - "ghcr"
  - "staking-api"
owners:
  - "Staking Team"
  - "Woojin"
source_revision_ids:
  - "srcrev_10a6b95e9d56e4ef69d1f0d879281367"
  - "srcrev_10ce5f5e6719d599d3d5e6c3fdbc4e12"
  - "srcrev_1c0f3e4de5f63f92395884894cac16bb"
  - "srcrev_4de9652173657e55657248d14b8dbae7"
  - "srcrev_56c62523a5517e728304e174112321ff"
  - "srcrev_62cfb8254f935489a9c63fe04d62b239"
  - "srcrev_6fd7e0081cccc958e5c69f20ccbe1554"
conflict_state: "none"
---

# Use Public GitHub Container Registry for Devnet Staking API Deployment

## Summary

Temporarily make the story-staking-api package public on GitHub Container Registry to unblock devnet deployment while waiting for AWS ECR access.

## Claims

- The Staking API deployment uses a Docker image pushed to GitHub Container Registry (ghcr.io/storyprotocol/story-staking-api). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_56c62523a5517e728304e174112321ff` `chunk_id=srcchunk_2548bd4d4cf2d9e4bea340c0295b5def` `native_locator=slack:C0547N89JUB:1775552711.299739:1775552711.299739` `source_timestamp=2026-04-07T09:08:46Z`
- The EC2 instance performing the deployment lacks GitHub authentication credentials, causing a permission error when pulling a private image. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_56c62523a5517e728304e174112321ff` `chunk_id=srcchunk_2548bd4d4cf2d9e4bea340c0295b5def` `native_locator=slack:C0547N89JUB:1775552711.299739:1775552711.299739` `source_timestamp=2026-04-07T09:08:46Z`
- The health check continuously returns HTTP 000 because the container never starts. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_56c62523a5517e728304e174112321ff` `chunk_id=srcchunk_2548bd4d4cf2d9e4bea340c0295b5def` `native_locator=slack:C0547N89JUB:1775552711.299739:1775552711.299739` `source_timestamp=2026-04-07T09:08:46Z`
- Setting the GHCR package to Public allows the EC2 instance to pull the image without authentication. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_56c62523a5517e728304e174112321ff` `chunk_id=srcchunk_2548bd4d4cf2d9e4bea340c0295b5def` `native_locator=slack:C0547N89JUB:1775552711.299739:1775552711.299739` `source_timestamp=2026-04-07T09:08:46Z`
- Changing the package visibility to Public requires organization owner access and cannot be done via API. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_1c0f3e4de5f63f92395884894cac16bb` `chunk_id=srcchunk_a8ef9642754746f0b550b13505164bab` `native_locator=slack:C0547N89JUB:1775552711.299739:1775553378.805599` `source_timestamp=2026-04-07T09:16:18Z`
- The team already uses AWS ECR as the primary registry for this service. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_62cfb8254f935489a9c63fe04d62b239` `chunk_id=srcchunk_7ba63361f3367c61b51b25e2e7932850` `native_locator=slack:C0547N89JUB:1775552711.299739:1775553994.937329` `source_timestamp=2026-04-07T09:26:34Z`
- AWS access (role) is not yet available, so GHCR is being used temporarily for devnet deployment. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_10a6b95e9d56e4ef69d1f0d879281367` `chunk_id=srcchunk_c3b6a85a64eaed6e57f83cbb4743afb2` `native_locator=slack:C0547N89JUB:1775552711.299739:1775555977.444179` `source_timestamp=2026-04-07T09:59:37Z`
- The temporary plan is to use GHCR public for devnet, then switch to ECR once AWS role is provisioned. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_6fd7e0081cccc958e5c69f20ccbe1554` `chunk_id=srcchunk_22df64c7e54593dd4c34dd9ffaf0fd9c` `native_locator=slack:C0547N89JUB:1775552711.299739:1775555986.382049` `source_timestamp=2026-04-07T09:59:46Z`
- Shen already has an image on GHCR for the link. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_10ce5f5e6719d599d3d5e6c3fdbc4e12` `chunk_id=srcchunk_bdaa58fb47a0896c97dd9ddaefc19fa0` `native_locator=slack:C0547N89JUB:1775552711.299739:1775554414.293689` `source_timestamp=2026-04-07T09:33:34Z`
- There was a concern about whether making the image public is acceptable. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_b216aabd787018439adf2aac6a1128f0` `source_revision_id=srcrev_4de9652173657e55657248d14b8dbae7` `chunk_id=srcchunk_3b3401aa61a49c660f2e94b34c9eafde` `native_locator=slack:C0547N89JUB:1775552711.299739:1775554498.052899` `source_timestamp=2026-04-07T09:34:58Z`

## Open Questions

- When will the AWS role be provisioned for ECR access?

## Related Pages

- `story-staking-api`

## Sources

- `source_document_id`: `srcdoc_b216aabd787018439adf2aac6a1128f0`
- `source_revision_id`: `srcrev_6fd7e0081cccc958e5c69f20ccbe1554`
