---
title: "Decision: Handling SSL for aeneid.portal.story.foundation"
type: "decision"
slug: "decisions/decision-aeneid-portal-ssl-proxy"
freshness: "2026-02-10T23:23:36Z"
tags:
  - "cloudflare"
  - "dns"
  - "ssl"
  - "vercel"
owners:
  - "slack:U05A515NBFC"
  - "slack:U08332YRB7W"
  - "slack:U09M2SPUTSL"
source_revision_ids:
  - "srcrev_102ff3a2eccec294769617ff65570969"
  - "srcrev_16aa498a37413b2390172cefbd819bed"
  - "srcrev_2292a94c388fd55a22b6ba248e877343"
  - "srcrev_3270861c58e603f6cf8174ac534d2ff1"
  - "srcrev_4d3cbfef24e356e3022c05bc3019fe29"
  - "srcrev_6ecebd9b5d28ad433f4ec5693313d0bf"
  - "srcrev_7e0fdddf7fbfe857933e9fc1d02d6139"
  - "srcrev_9e43736908eefd82b2d53cb75d900a3f"
  - "srcrev_a51e211308ccd800d2a4723b0399c2c8"
  - "srcrev_b06811d29850335772467ffa31a70b32"
  - "srcrev_c11342c7dffd3ecdb69c64f00a1ac845"
  - "srcrev_cc2deee9e13ae814627252ad55972a70"
conflict_state: "none"
---

# Decision: Handling SSL for aeneid.portal.story.foundation

## Summary

Resolving SSL cipher mismatch on aeneid.portal.story.foundation caused by Cloudflare multi-level subdomain limitation. Decision to remove Cloudflare proxy as quick fix, with plan to change domain to single-level subdomain later.

## Claims

- A Cloudflare DNS addition for aeneid.portal.story.foundation was requested and reviewed as part of PR #120. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_cc2deee9e13ae814627252ad55972a70` `chunk_id=srcchunk_0b43df1203cdf90b542020fb1d31cfb3` `native_locator=slack:C0547N89JUB:1770276983.935799:1770276983.935799` `source_timestamp=2026-02-05T07:36:23Z`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_4d3cbfef24e356e3022c05bc3019fe29` `chunk_id=srcchunk_17ef52975556e35ca83b23e9262c5ffb` `native_locator=slack:C0547N89JUB:1770276983.935799:1770309862.514659` `source_timestamp=2026-02-05T16:44:22Z`
- After deployment, the site aeneid.portal.story.foundation experienced SSL cipher mismatch error, rendering it inaccessible via HTTPS. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_7e0fdddf7fbfe857933e9fc1d02d6139` `chunk_id=srcchunk_5a6e654e196a75653d4a316764546338` `native_locator=slack:C0547N89JUB:1770276983.935799:1770630459.148909` `source_timestamp=2026-02-09T09:47:39Z`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_b06811d29850335772467ffa31a70b32` `chunk_id=srcchunk_3166b0494dd9698a202a2a92c4714ed6` `native_locator=slack:C0547N89JUB:1770276983.935799:1770658639.803209` `source_timestamp=2026-02-09T17:37:19Z`
- Vercel automatically provisioned an SSL certificate for the domain when it was added. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_102ff3a2eccec294769617ff65570969` `chunk_id=srcchunk_8887d03bc1ff5686eb3caf57b3532a5d` `native_locator=slack:C0547N89JUB:1770276983.935799:1770630481.618159` `source_timestamp=2026-02-09T09:48:01Z`
- The issue is likely due to Cloudflare's limitation with multi-level subdomain wildcard SSL certificates. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_2292a94c388fd55a22b6ba248e877343` `chunk_id=srcchunk_6b141ef9696c657e11edc3a4533096b2` `native_locator=slack:C0547N89JUB:1770276983.935799:1770702491.145929` `source_timestamp=2026-02-10T05:48:11Z`
- Removing the Cloudflare proxy DNS record resolved the SSL issue, but exposed the site directly to Vercel, losing Cloudflare WAF and DDoS protection. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_3270861c58e603f6cf8174ac534d2ff1` `chunk_id=srcchunk_d55960236be1e1072e74a552687caae9` `native_locator=slack:C0547N89JUB:1770276983.935799:1770702672.076819` `source_timestamp=2026-02-10T05:51:12Z`
- Two permanent solutions were proposed: Option 1) change the domain to aeneid-portal.story.foundation (single-level subdomain) to retain Cloudflare protection; Option 2) keep the multi-level domain with no Cloudflare proxy. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_16aa498a37413b2390172cefbd819bed` `chunk_id=srcchunk_d2d0ea6ee2cda94066e05c4641cd5f19` `native_locator=slack:C0547N89JUB:1770276983.935799:1770745633.616429` `source_timestamp=2026-02-10T17:47:13Z`
- Team members expressed a preference for Option 2 as a quick temporary fix, with the intention to later implement Option 1 as the permanent solution. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_6ecebd9b5d28ad433f4ec5693313d0bf` `chunk_id=srcchunk_412ff0a75eea51165bdbcefd0269e025` `native_locator=slack:C0547N89JUB:1770276983.935799:1770760143.365469` `source_timestamp=2026-02-10T21:49:03Z`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_a51e211308ccd800d2a4723b0399c2c8` `chunk_id=srcchunk_6cf9ea9bfdc31ebf56970901fff85ab9` `native_locator=slack:C0547N89JUB:1770276983.935799:1770765816.665099` `source_timestamp=2026-02-10T23:23:36Z`
- Other multi-level subdomains like staging.portal.story.foundation and canary.portal.story.foundation work with Cloudflare proxy because they were likely added before Cloudflare proxy was used. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_c11342c7dffd3ecdb69c64f00a1ac845` `chunk_id=srcchunk_c7eba6076951a44c3c66fdc127124014` `native_locator=slack:C0547N89JUB:1770276983.935799:1770688212.706459` `source_timestamp=2026-02-10T01:50:12Z`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_9e43736908eefd82b2d53cb75d900a3f` `chunk_id=srcchunk_47ed5c51d71e9a440b863e7297da17ca` `native_locator=slack:C0547N89JUB:1770276983.935799:1770688969.911039` `source_timestamp=2026-02-10T02:02:49Z`

## Open Questions

- Why do some multi-level subdomains work with Cloudflare proxy while aeneid fails?

## Related Pages

- `open-qa-multi-subdomain-ssl-discrepancy`

## Sources

- `source_document_id`: `srcdoc_e5fe9c52e38e284597f96f38a8a3afbd`
- `source_revision_id`: `srcrev_7e0fdddf7fbfe857933e9fc1d02d6139`
