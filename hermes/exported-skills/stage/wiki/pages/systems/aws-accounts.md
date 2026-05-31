---
title: "Our AWS Accounts"
type: "system"
slug: "systems/aws-accounts"
freshness: "2023-12-27T22:11:00Z"
tags:
  - "accounts"
  - "aws"
  - "infrastructure"
owners: []
source_revision_ids:
  - "srcrev_8d000ae59ef74843c889c77108c4748f"
conflict_state: "none"
---

# Our AWS Accounts

## Summary

Details of the AWS accounts used by the organization, including account IDs, purposes, regions, and deprecation status.

## Claims

- The login portal for AWS accounts is https://story.awsapps.com/start#/. `claim:claim_aws_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb) `source_document_id=srcdoc_a8d2b9caf43ddf476f3369297641cc19` `source_revision_id=srcrev_8d000ae59ef74843c889c77108c4748f` `chunk_id=srcchunk_6620cfd16a67f0c824f6f737ad25a7fe` `native_locator=https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb` `source_timestamp=2023-12-27T22:11:00Z`
- Account ID 087635269473 (story-cicd) is not used yet. `claim:claim_aws_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb) `source_document_id=srcdoc_a8d2b9caf43ddf476f3369297641cc19` `source_revision_id=srcrev_8d000ae59ef74843c889c77108c4748f` `chunk_id=srcchunk_6620cfd16a67f0c824f6f737ad25a7fe` `native_locator=https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb` `source_timestamp=2023-12-27T22:11:00Z`
- Account ID 145440314588 (story-management) is used for IAM and user management. `claim:claim_aws_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb) `source_document_id=srcdoc_a8d2b9caf43ddf476f3369297641cc19` `source_revision_id=srcrev_8d000ae59ef74843c889c77108c4748f` `chunk_id=srcchunk_6620cfd16a67f0c824f6f737ad25a7fe` `native_locator=https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb` `source_timestamp=2023-12-27T22:11:00Z`
- Account ID 478656756051 (story-services-staging) uses us-west-2 for the staging environment. `claim:claim_aws_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb) `source_document_id=srcdoc_a8d2b9caf43ddf476f3369297641cc19` `source_revision_id=srcrev_8d000ae59ef74843c889c77108c4748f` `chunk_id=srcchunk_6620cfd16a67f0c824f6f737ad25a7fe` `native_locator=https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb` `source_timestamp=2023-12-27T22:11:00Z`
- Account ID 243963068353 (story-services-production) uses us-west-2 for the production environment, while us-east-1 is a deprecated production environment and us-east-2 is a deprecated staging environment. `claim:claim_aws_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb) `source_document_id=srcdoc_a8d2b9caf43ddf476f3369297641cc19` `source_revision_id=srcrev_8d000ae59ef74843c889c77108c4748f` `chunk_id=srcchunk_6620cfd16a67f0c824f6f737ad25a7fe` `native_locator=https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb` `source_timestamp=2023-12-27T22:11:00Z`
- Account ID 120007995527 (story-shared-services) is not used yet. `claim:claim_aws_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb) `source_document_id=srcdoc_a8d2b9caf43ddf476f3369297641cc19` `source_revision_id=srcrev_8d000ae59ef74843c889c77108c4748f` `chunk_id=srcchunk_6620cfd16a67f0c824f6f737ad25a7fe` `native_locator=https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb` `source_timestamp=2023-12-27T22:11:00Z`
- The deprecated production EKS environment #243963068353/us-east-1 has not been used. `claim:claim_aws_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb) `source_document_id=srcdoc_a8d2b9caf43ddf476f3369297641cc19` `source_revision_id=srcrev_8d000ae59ef74843c889c77108c4748f` `chunk_id=srcchunk_6620cfd16a67f0c824f6f737ad25a7fe` `native_locator=https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb` `source_timestamp=2023-12-27T22:11:00Z`
- The deprecated staging EKS environment #243963068353/us-east-2 had the API running for Hackathon but is being moved to staging environment #478656756051/us-west-2, which has its TF codes and K8S manifests synced to the IAC repo. `claim:claim_aws_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb) `source_document_id=srcdoc_a8d2b9caf43ddf476f3369297641cc19` `source_revision_id=srcrev_8d000ae59ef74843c889c77108c4748f` `chunk_id=srcchunk_6620cfd16a67f0c824f6f737ad25a7fe` `native_locator=https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb` `source_timestamp=2023-12-27T22:11:00Z`
- Region us-west-2 was chosen because AWS account #243963068353 already had existing infrastructure in us-east-1 and us-east-2, and us-west-2 will work in both #478656756051 and #243963068353 as the target states for staging and production respectively. `claim:claim_aws_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb) `source_document_id=srcdoc_a8d2b9caf43ddf476f3369297641cc19` `source_revision_id=srcrev_8d000ae59ef74843c889c77108c4748f` `chunk_id=srcchunk_6620cfd16a67f0c824f6f737ad25a7fe` `native_locator=https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb` `source_timestamp=2023-12-27T22:11:00Z`
- The move to us-west-2 is to have the IAC repo synced to the actual AWS environment infra for easier and more consistent maintenance. `claim:claim_aws_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb) `source_document_id=srcdoc_a8d2b9caf43ddf476f3369297641cc19` `source_revision_id=srcrev_8d000ae59ef74843c889c77108c4748f` `chunk_id=srcchunk_6620cfd16a67f0c824f6f737ad25a7fe` `native_locator=https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb` `source_timestamp=2023-12-27T22:11:00Z`

## Related Pages

- `concepts/devex`

## Sources

- `source_document_id`: `srcdoc_a8d2b9caf43ddf476f3369297641cc19`
- `source_revision_id`: `srcrev_8d000ae59ef74843c889c77108c4748f`
- `source_url`: [Notion source](https://www.notion.so/Our-AWS-Accounts-025ab8a684b34a0a84cdda283c9a95bb)
