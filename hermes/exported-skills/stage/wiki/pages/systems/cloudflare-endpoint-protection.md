---
title: "Cloudflare Endpoint Protection"
type: "system"
slug: "systems/cloudflare-endpoint-protection"
freshness: "2025-08-21T21:26:00Z"
tags:
  - "cloudflare"
  - "poseidon"
  - "psdn"
  - "rate-limiting"
  - "security"
owners: []
source_revision_ids:
  - "srcrev_edb61495a4eb031a827fdbce14e44002"
conflict_state: "none"
---

# Cloudflare Endpoint Protection

## Summary

Configuration for Cloudflare protection of PSDN frontend and backend endpoints, including origin hardening, rate limiting rules, and bot management.

## Claims

- Frontend production URL is https://app.psdn.ai/login. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`
- Frontend staging URL is https://staging.psdn.ai/. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`
- Backend production URL is https://poseidon-depin-server.storyapis.com. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`
- Backend staging URL is https://poseidon-mvp-media-platform.storyapis.com/. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`
- Origin hardening includes enabling proxy, with mTLS and GCP firewall restriction to Cloudflare IP ranges under review. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`
- Frontend rate limiting: 1000 requests per minute per IP for production triggers managed challenge. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`
- Frontend rate limiting: 400 requests per minute per IP for staging triggers managed challenge. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`
- Backend rate limiting for POST /files* is 20 requests per minute per IP, blocking for 600 seconds. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`
- Backend rate limiting for POST /users/me/* is 30 requests per minute per IP, blocking for 600 seconds. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`
- Backend rate limiting for POST /users/world-id/* is 10 requests per minute per IP, blocking for 600 seconds. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`
- Backend rate limiting for GET /scripts/next is 60 requests per hour per IP, blocking for 3600 seconds. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`
- Backend rate limiting for multipart/form-data requests is 15 requests per minute per IP, blocking for 600 seconds. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`
- Backend protection includes bot management/SBFM skips for specific API flows and origin hardening with AOP and IP allowlisting. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`
- Frontend uses Cloudflare caching for static assets and optional Turnstile. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334) `source_document_id=srcdoc_bca909f2d31f8e57a1cf96dee5f04170` `source_revision_id=srcrev_edb61495a4eb031a827fdbce14e44002` `chunk_id=srcchunk_2a8804005b811d2d90de00967376dc6c` `native_locator=https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334` `source_timestamp=2025-08-21T21:26:00Z`

## Open Questions

- Should GCP firewall be restricted to only Cloudflare IP ranges?
- Should mTLS be enabled for origin hardening?
- Should Turnstile be implemented for frontend?

## Sources

- `source_document_id`: `srcdoc_bca909f2d31f8e57a1cf96dee5f04170`
- `source_revision_id`: `srcrev_edb61495a4eb031a827fdbce14e44002`
- `source_url`: [Notion source](https://www.notion.so/Endpoints-Protection-on-CloudFlare-255051299a5480929666cb29de7c8334)
