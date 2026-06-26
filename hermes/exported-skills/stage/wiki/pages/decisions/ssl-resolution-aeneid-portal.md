---
title: "SSL Resolution for aeneid.portal.story.foundation"
type: "decision"
slug: "decisions/ssl-resolution-aeneid-portal"
freshness: "2026-02-10T23:23:36Z"
tags:
  - "cloudflare"
  - "dns"
  - "ssl"
  - "story-foundation"
  - "vercel"
owners: []
source_revision_ids:
  - "srcrev_102ff3a2eccec294769617ff65570969"
  - "srcrev_16aa498a37413b2390172cefbd819bed"
  - "srcrev_2292a94c388fd55a22b6ba248e877343"
  - "srcrev_3270861c58e603f6cf8174ac534d2ff1"
  - "srcrev_3e0f713e850fb05e4cc0b569fbd1d56f"
  - "srcrev_4d3cbfef24e356e3022c05bc3019fe29"
  - "srcrev_6ecebd9b5d28ad433f4ec5693313d0bf"
  - "srcrev_7338f4aceee0dba3ec2c96f7c223ee14"
  - "srcrev_7e0fdddf7fbfe857933e9fc1d02d6139"
  - "srcrev_9e43736908eefd82b2d53cb75d900a3f"
  - "srcrev_a51e211308ccd800d2a4723b0399c2c8"
  - "srcrev_b06811d29850335772467ffa31a70b32"
  - "srcrev_bdd9eb17ed2bb6a1e5924ef3a59c18c0"
  - "srcrev_c11342c7dffd3ecdb69c64f00a1ac845"
  - "srcrev_cc2deee9e13ae814627252ad55972a70"
conflict_state: "none"
---

# SSL Resolution for aeneid.portal.story.foundation

## Summary

Decision to temporarily disable Cloudflare proxy for aeneid.portal.story.foundation to resolve SSL cipher mismatch caused by multi-level subdomain wildcard certificate limitation. Long-term plan to migrate to single-level subdomain.

## Claims

- A pull request (#120) was created in piplabs/cloudflare to add DNS configuration for aeneid.portal.story.foundation. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_cc2deee9e13ae814627252ad55972a70` `chunk_id=srcchunk_0b43df1203cdf90b542020fb1d31cfb3` `native_locator=slack:C0547N89JUB:1770276983.935799:1770276983.935799` `source_timestamp=2026-02-05T07:36:23Z`
- The pull request was approved. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_4d3cbfef24e356e3022c05bc3019fe29` `chunk_id=srcchunk_17ef52975556e35ca83b23e9262c5ffb` `native_locator=slack:C0547N89JUB:1770276983.935799:1770309862.514659` `source_timestamp=2026-02-05T16:44:22Z`
- The pull request was merged. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_7338f4aceee0dba3ec2c96f7c223ee14` `chunk_id=srcchunk_f2181d5fdae142eda3b89ca939ca7e6f` `native_locator=slack:C0547N89JUB:1770276983.935799:1770312855.530669` `source_timestamp=2026-02-05T17:34:15Z`
- After deployment, accessing aeneid.portal.story.foundation resulted in an SSL error. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_7e0fdddf7fbfe857933e9fc1d02d6139` `chunk_id=srcchunk_5a6e654e196a75653d4a316764546338` `native_locator=slack:C0547N89JUB:1770276983.935799:1770630459.148909` `source_timestamp=2026-02-09T09:47:39Z`
- Vercel automatically provisioned an SSL certificate for the domain. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_102ff3a2eccec294769617ff65570969` `chunk_id=srcchunk_8887d03bc1ff5686eb3caf57b3532a5d` `native_locator=slack:C0547N89JUB:1770276983.935799:1770630481.618159` `source_timestamp=2026-02-09T09:48:01Z`
- The SSL error was identified as a cipher mismatch between client and server. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_b06811d29850335772467ffa31a70b32` `chunk_id=srcchunk_3166b0494dd9698a202a2a92c4714ed6` `native_locator=slack:C0547N89JUB:1770276983.935799:1770658639.803209` `source_timestamp=2026-02-09T17:37:19Z`
- The error occurred across multiple browsers and incognito mode, indicating a server-side issue. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_3e0f713e850fb05e4cc0b569fbd1d56f` `chunk_id=srcchunk_67402b6f391498082436658c3a8ac6b3` `native_locator=slack:C0547N89JUB:1770276983.935799:1770659156.323299` `source_timestamp=2026-02-09T17:45:56Z`
- The site is hosted on Vercel, and the SSL certificate is provisioned by Vercel, but the DNS is behind Cloudflare proxy. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_bdd9eb17ed2bb6a1e5924ef3a59c18c0` `chunk_id=srcchunk_a832bf5794e86b59281dad43630e4881` `native_locator=slack:C0547N89JUB:1770276983.935799:1770660093.633799` `source_timestamp=2026-02-09T18:01:33Z`
- Other story.foundation subdomains (e.g., staging.portal.story.foundation, canary.portal.story.foundation) work, but aeneid.portal.story.foundation does not, possibly because its DNS record was added while Cloudflare proxy was enabled. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_c11342c7dffd3ecdb69c64f00a1ac845` `chunk_id=srcchunk_c7eba6076951a44c3c66fdc127124014` `native_locator=slack:C0547N89JUB:1770276983.935799:1770688212.706459` `source_timestamp=2026-02-10T01:50:12Z`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_9e43736908eefd82b2d53cb75d900a3f` `chunk_id=srcchunk_47ed5c51d71e9a440b863e7297da17ca` `native_locator=slack:C0547N89JUB:1770276983.935799:1770688969.911039` `source_timestamp=2026-02-10T02:02:49Z`
- The root cause is a Cloudflare limitation with wildcard SSL certificates on multi-level subdomains (e.g., *.story.foundation does not cover aeneid.portal.story.foundation). `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_2292a94c388fd55a22b6ba248e877343` `chunk_id=srcchunk_6b141ef9696c657e11edc3a4533096b2` `native_locator=slack:C0547N89JUB:1770276983.935799:1770702491.145929` `source_timestamp=2026-02-10T05:48:11Z`
- Option 1: Change the domain to a single-level subdomain (e.g., aeneid-portal.story.foundation) to restore Cloudflare WAF and DDoS protection. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_3270861c58e603f6cf8174ac534d2ff1` `chunk_id=srcchunk_d55960236be1e1072e74a552687caae9` `native_locator=slack:C0547N89JUB:1770276983.935799:1770702672.076819` `source_timestamp=2026-02-10T05:51:12Z`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_16aa498a37413b2390172cefbd819bed` `chunk_id=srcchunk_d2d0ea6ee2cda94066e05c4641cd5f19` `native_locator=slack:C0547N89JUB:1770276983.935799:1770745633.616429` `source_timestamp=2026-02-10T17:47:13Z`
- Option 2: Remove the Cloudflare proxied DNS record for aeneid.portal.story.foundation, losing WAF and DDoS protection but keeping the domain unchanged. `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_3270861c58e603f6cf8174ac534d2ff1` `chunk_id=srcchunk_d55960236be1e1072e74a552687caae9` `native_locator=slack:C0547N89JUB:1770276983.935799:1770702672.076819` `source_timestamp=2026-02-10T05:51:12Z`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_16aa498a37413b2390172cefbd819bed` `chunk_id=srcchunk_d2d0ea6ee2cda94066e05c4641cd5f19` `native_locator=slack:C0547N89JUB:1770276983.935799:1770745633.616429` `source_timestamp=2026-02-10T17:47:13Z`
- Decision: Initially implement option 2 as a quick fix; later migrate to option 1 when code changes are ready to support the domain change. `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_6ecebd9b5d28ad433f4ec5693313d0bf` `chunk_id=srcchunk_412ff0a75eea51165bdbcefd0269e025` `native_locator=slack:C0547N89JUB:1770276983.935799:1770760143.365469` `source_timestamp=2026-02-10T21:49:03Z`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_a51e211308ccd800d2a4723b0399c2c8` `chunk_id=srcchunk_6cf9ea9bfdc31ebf56970901fff85ab9` `native_locator=slack:C0547N89JUB:1770276983.935799:1770765816.665099` `source_timestamp=2026-02-10T23:23:36Z`

## Open Questions

- Why do other multi-level subdomains like staging.portal.story.foundation and canary.portal.story.foundation work even with Cloudflare proxy?

## Sources

- `source_document_id`: `srcdoc_e5fe9c52e38e284597f96f38a8a3afbd`
- `source_revision_id`: `srcrev_c11342c7dffd3ecdb69c64f00a1ac845`
