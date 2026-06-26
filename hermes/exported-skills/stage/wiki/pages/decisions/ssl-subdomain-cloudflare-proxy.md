---
title: "SSL Issue for Multi-Level Subdomain with Cloudflare Proxy"
type: "decision"
slug: "decisions/ssl-subdomain-cloudflare-proxy"
freshness: "2026-02-10T23:23:36Z"
tags:
  - "cloudflare"
  - "dns"
  - "ssl"
  - "vercel"
owners: []
source_revision_ids:
  - "srcrev_102ff3a2eccec294769617ff65570969"
  - "srcrev_16aa498a37413b2390172cefbd819bed"
  - "srcrev_2292a94c388fd55a22b6ba248e877343"
  - "srcrev_3270861c58e603f6cf8174ac534d2ff1"
  - "srcrev_3e0f713e850fb05e4cc0b569fbd1d56f"
  - "srcrev_6ecebd9b5d28ad433f4ec5693313d0bf"
  - "srcrev_7e0fdddf7fbfe857933e9fc1d02d6139"
  - "srcrev_a51e211308ccd800d2a4723b0399c2c8"
  - "srcrev_b06811d29850335772467ffa31a70b32"
conflict_state: "none"
---

# SSL Issue for Multi-Level Subdomain with Cloudflare Proxy

## Summary

Decision to temporarily remove Cloudflare proxy for aeneid.portal.story.foundation to resolve SSL cipher mismatch, with plan to later migrate to a single-level subdomain to regain DDoS/WAF protection.

## Claims

- The domain aeneid.portal.story.foundation encountered SSL cipher mismatch errors. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_7e0fdddf7fbfe857933e9fc1d02d6139` `chunk_id=srcchunk_5a6e654e196a75653d4a316764546338` `native_locator=slack:C0547N89JUB:1770276983.935799:1770630459.148909` `source_timestamp=2026-02-09T09:47:39Z`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_b06811d29850335772467ffa31a70b32` `chunk_id=srcchunk_3166b0494dd9698a202a2a92c4714ed6` `native_locator=slack:C0547N89JUB:1770276983.935799:1770658639.803209` `source_timestamp=2026-02-09T17:37:19Z`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_3e0f713e850fb05e4cc0b569fbd1d56f` `chunk_id=srcchunk_67402b6f391498082436658c3a8ac6b3` `native_locator=slack:C0547N89JUB:1770276983.935799:1770659156.323299` `source_timestamp=2026-02-09T17:45:56Z`
- The SSL certificate was provisioned by Vercel when the domain was added. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_102ff3a2eccec294769617ff65570969` `chunk_id=srcchunk_8887d03bc1ff5686eb3caf57b3532a5d` `native_locator=slack:C0547N89JUB:1770276983.935799:1770630481.618159` `source_timestamp=2026-02-09T09:48:01Z`
- The issue is caused by multi-level subdomain wildcard SSL limitations when using Cloudflare's proxy. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_2292a94c388fd55a22b6ba248e877343` `chunk_id=srcchunk_6b141ef9696c657e11edc3a4533096b2` `native_locator=slack:C0547N89JUB:1770276983.935799:1770702491.145929` `source_timestamp=2026-02-10T05:48:11Z`
- Removing the proxied DNS record resolved the issue but means traffic hits Vercel directly, losing Cloudflare WAF and DDoS protection. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_3270861c58e603f6cf8174ac534d2ff1` `chunk_id=srcchunk_d55960236be1e1072e74a552687caae9` `native_locator=slack:C0547N89JUB:1770276983.935799:1770702672.076819` `source_timestamp=2026-02-10T05:51:12Z`
- Two options were considered: Option 1 – change domain to a single-level subdomain (e.g., aeneid-portal.story.foundation) to keep Cloudflare protection; Option 2 – remove the proxied DNS record and keep the multi-level domain. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_16aa498a37413b2390172cefbd819bed` `chunk_id=srcchunk_d2d0ea6ee2cda94066e05c4641cd5f19` `native_locator=slack:C0547N89JUB:1770276983.935799:1770745633.616429` `source_timestamp=2026-02-10T17:47:13Z`
- The immediate decision was to implement Option 2 as a quick fix, then later switch to Option 1 after code changes. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_6ecebd9b5d28ad433f4ec5693313d0bf` `chunk_id=srcchunk_412ff0a75eea51165bdbcefd0269e025` `native_locator=slack:C0547N89JUB:1770276983.935799:1770760143.365469` `source_timestamp=2026-02-10T21:49:03Z`
- Another stakeholder expressed preference for Option 2 to maintain domain consistency with other subdomains. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_a51e211308ccd800d2a4723b0399c2c8` `chunk_id=srcchunk_6cf9ea9bfdc31ebf56970901fff85ab9` `native_locator=slack:C0547N89JUB:1770276983.935799:1770765816.665099` `source_timestamp=2026-02-10T23:23:36Z`

## Open Questions

- Why do staging.portal.story.foundation and canary.portal.story.foundation work with Cloudflare proxy while aeneid.portal does not?

## Sources

- `source_document_id`: `srcdoc_e5fe9c52e38e284597f96f38a8a3afbd`
- `source_revision_id`: `srcrev_9e43736908eefd82b2d53cb75d900a3f`
