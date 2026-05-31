---
title: "CDN Management"
type: "runbook"
slug: "runbooks/cdn-management"
freshness: "2024-10-21T22:07:00Z"
tags:
  - "assets"
  - "cdn"
  - "cloudflare"
  - "s3"
owners:
  - "devops@storyprotocol.xyz"
source_revision_ids:
  - "srcrev_973155374c5ebf8ca378a66a78ac09af"
conflict_state: "none"
---

# CDN Management

## Summary

Guide for managing the CDN that fronts public S3 assets using Cloudflare.

## Claims

- Production CDN domain is cdn.sp-assets.net, backed by the S3 bucket story-services-prod. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-1) `source_document_id=srcdoc_4e28ac60379ab61837ad34267e29cdaa` `source_revision_id=srcrev_973155374c5ebf8ca378a66a78ac09af` `chunk_id=srcchunk_4cc8e674689ed258a669386722199431` `native_locator=https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-1` `source_timestamp=2024-10-21T22:07:00Z`
- Staging CDN domain is staging-cdn.sp-assets.net, backed by the S3 bucket story-services-staging. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-1) `source_document_id=srcdoc_4e28ac60379ab61837ad34267e29cdaa` `source_revision_id=srcrev_973155374c5ebf8ca378a66a78ac09af` `chunk_id=srcchunk_4cc8e674689ed258a669386722199431` `native_locator=https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-1` `source_timestamp=2024-10-21T22:07:00Z`
- Cloudflare login credentials: username devops@storyprotocol.xyz, password shared via LastPass (ask specific users for access). `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-1) `source_document_id=srcdoc_4e28ac60379ab61837ad34267e29cdaa` `source_revision_id=srcrev_973155374c5ebf8ca378a66a78ac09af` `chunk_id=srcchunk_4cc8e674689ed258a669386722199431` `native_locator=https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-1` `source_timestamp=2024-10-21T22:07:00Z`
- To upload assets, visit the corresponding S3 bucket for the environment. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-1) `source_document_id=srcdoc_4e28ac60379ab61837ad34267e29cdaa` `source_revision_id=srcrev_973155374c5ebf8ca378a66a78ac09af` `chunk_id=srcchunk_4cc8e674689ed258a669386722199431` `native_locator=https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-1` `source_timestamp=2024-10-21T22:07:00Z`
- Add a DNS target in Cloudflare for the subdomain, pointing to the S3 website endpoint $SUBDOMAIN.s3.$REGION.amazonaws.com. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-2) `source_document_id=srcdoc_4e28ac60379ab61837ad34267e29cdaa` `source_revision_id=srcrev_973155374c5ebf8ca378a66a78ac09af` `chunk_id=srcchunk_f677a2f0d16bcca4da385d10c3034989` `native_locator=https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-2` `source_timestamp=2024-10-21T22:07:00Z`
- Attach an S3 bucket policy that allows Cloudflare IPs to perform s3:GetObject on the bucket. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-3) `source_document_id=srcdoc_4e28ac60379ab61837ad34267e29cdaa` `source_revision_id=srcrev_973155374c5ebf8ca378a66a78ac09af` `chunk_id=srcchunk_c76553319204d032de516ebf16632c04` `native_locator=https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-3` `source_timestamp=2024-10-21T22:07:00Z`
- Add a CORS policy to the bucket allowing GET requests from all origins. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-3) `source_document_id=srcdoc_4e28ac60379ab61837ad34267e29cdaa` `source_revision_id=srcrev_973155374c5ebf8ca378a66a78ac09af` `chunk_id=srcchunk_c76553319204d032de516ebf16632c04` `native_locator=https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9#chunk-3` `source_timestamp=2024-10-21T22:07:00Z`

## Sources

- `source_document_id`: `srcdoc_4e28ac60379ab61837ad34267e29cdaa`
- `source_revision_id`: `srcrev_973155374c5ebf8ca378a66a78ac09af`
- `source_url`: [Notion source](https://www.notion.so/CDN-Management-e42897fdb4204fd18e8cb25aac6f85a9)
