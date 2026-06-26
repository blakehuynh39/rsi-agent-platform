---
title: "Aeneid Portal SSL Fix Decision"
type: "decision"
slug: "decisions/aeneid-portal-ssl-fix"
freshness: "2026-02-10T23:23:36Z"
tags:
  - "aeneid"
  - "cloudflare"
  - "dns"
  - "ssl"
  - "vercel"
owners:
  - "U05A515NBFC"
  - "U08332YRB7W"
  - "U09M2SPUTSL"
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
  - "srcrev_a51e211308ccd800d2a4723b0399c2c8"
  - "srcrev_b06811d29850335772467ffa31a70b32"
  - "srcrev_cc2deee9e13ae814627252ad55972a70"
conflict_state: "none"
---

# Aeneid Portal SSL Fix Decision

## Summary

Decision to temporarily disable Cloudflare proxy for aeneid.portal.story.foundation due to SSL cipher mismatch caused by multi-level subdomain wildcard certificate limitations, with a long-term plan to move to a single-level subdomain.

## Claims

- A Cloudflare DNS addition PR (#120) was reviewed and approved by team members. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_cc2deee9e13ae814627252ad55972a70` `chunk_id=srcchunk_0b43df1203cdf90b542020fb1d31cfb3` `native_locator=slack:C0547N89JUB:1770276983.935799:1770276983.935799` `source_timestamp=2026-02-05T07:36:23Z`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_4d3cbfef24e356e3022c05bc3019fe29` `chunk_id=srcchunk_17ef52975556e35ca83b23e9262c5ffb` `native_locator=slack:C0547N89JUB:1770276983.935799:1770309862.514659` `source_timestamp=2026-02-05T16:44:22Z`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_7338f4aceee0dba3ec2c96f7c223ee14` `chunk_id=srcchunk_f2181d5fdae142eda3b89ca939ca7e6f` `native_locator=slack:C0547N89JUB:1770276983.935799:1770312855.530669` `source_timestamp=2026-02-05T17:34:15Z`
- The domain aeneid.portal.story.foundation encountered SSL errors due to cipher mismatch. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_7e0fdddf7fbfe857933e9fc1d02d6139` `chunk_id=srcchunk_5a6e654e196a75653d4a316764546338` `native_locator=slack:C0547N89JUB:1770276983.935799:1770630459.148909` `source_timestamp=2026-02-09T09:47:39Z`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_b06811d29850335772467ffa31a70b32` `chunk_id=srcchunk_3166b0494dd9698a202a2a92c4714ed6` `native_locator=slack:C0547N89JUB:1770276983.935799:1770658639.803209` `source_timestamp=2026-02-09T17:37:19Z`
- Vercel provisioned an SSL certificate for the domain when it was added. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_102ff3a2eccec294769617ff65570969` `chunk_id=srcchunk_8887d03bc1ff5686eb3caf57b3532a5d` `native_locator=slack:C0547N89JUB:1770276983.935799:1770630481.618159` `source_timestamp=2026-02-09T09:48:01Z`
- The cipher mismatch error occurred across multiple browsers and incognito mode. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_3e0f713e850fb05e4cc0b569fbd1d56f` `chunk_id=srcchunk_67402b6f391498082436658c3a8ac6b3` `native_locator=slack:C0547N89JUB:1770276983.935799:1770659156.323299` `source_timestamp=2026-02-09T17:45:56Z`
- The root cause is that Cloudflare does not support multi-level subdomain wildcard SSL certificates when the proxy is enabled. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_2292a94c388fd55a22b6ba248e877343` `chunk_id=srcchunk_6b141ef9696c657e11edc3a4533096b2` `native_locator=slack:C0547N89JUB:1770276983.935799:1770702491.145929` `source_timestamp=2026-02-10T05:48:11Z`
- Removing the proxied DNS record in Cloudflare resolved the SSL issue, but traffic then goes directly to Vercel, bypassing Cloudflare's WAF and DDoS protection. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_3270861c58e603f6cf8174ac534d2ff1` `chunk_id=srcchunk_d55960236be1e1072e74a552687caae9` `native_locator=slack:C0547N89JUB:1770276983.935799:1770702672.076819` `source_timestamp=2026-02-10T05:51:12Z`
- Two options were considered: 1) change the domain to a single-level subdomain (e.g., aeneid-portal.story.foundation) to retain Cloudflare protection, or 2) remove the Cloudflare proxy and keep the existing multi-level subdomain. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_16aa498a37413b2390172cefbd819bed` `chunk_id=srcchunk_d2d0ea6ee2cda94066e05c4641cd5f19` `native_locator=slack:C0547N89JUB:1770276983.935799:1770745633.616429` `source_timestamp=2026-02-10T17:47:13Z`
- The immediate decision was to go with option 2 (remove proxy) as a quick fix, with a plan to implement option 1 later after code changes. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_6ecebd9b5d28ad433f4ec5693313d0bf` `chunk_id=srcchunk_412ff0a75eea51165bdbcefd0269e025` `native_locator=slack:C0547N89JUB:1770276983.935799:1770760143.365469` `source_timestamp=2026-02-10T21:49:03Z`
- U09M2SPUTSL preferred to stick with option 2 (keep existing domain) for consistency with other subdomain variations. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_a51e211308ccd800d2a4723b0399c2c8` `chunk_id=srcchunk_6cf9ea9bfdc31ebf56970901fff85ab9` `native_locator=slack:C0547N89JUB:1770276983.935799:1770765816.665099` `source_timestamp=2026-02-10T23:23:36Z`

## Open Questions

- Why do other multi-level subdomains like staging.portal.story.foundation and canary.portal.story.foundation work without SSL issues?

## Sources

- `source_document_id`: `srcdoc_e5fe9c52e38e284597f96f38a8a3afbd`
- `source_revision_id`: `srcrev_bdd9eb17ed2bb6a1e5924ef3a59c18c0`
