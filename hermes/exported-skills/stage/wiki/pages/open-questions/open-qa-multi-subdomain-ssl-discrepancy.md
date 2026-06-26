---
title: "Open Question: Why do some multi-level subdomains work with Cloudflare proxy?"
type: "open_question"
slug: "open-questions/open-qa-multi-subdomain-ssl-discrepancy"
freshness: "2026-02-10T23:23:39Z"
tags:
  - "cloudflare"
  - "dns"
  - "ssl"
owners: []
source_revision_ids:
  - "srcrev_726633655bd32252eb297b924c81aa1b"
  - "srcrev_9e43736908eefd82b2d53cb75d900a3f"
conflict_state: "none"
---

# Open Question: Why do some multi-level subdomains work with Cloudflare proxy?

## Summary

Understanding why subdomains like staging.portal.story.foundation work with Cloudflare proxy despite being multi-level, while aeneid.portal.story.foundation does not.

## Claims

- While aeneid.portal.story.foundation failed with Cloudflare proxy, other multi-level subdomains like staging.portal.story.foundation and canary.portal.story.foundation continue to work. `claim:claim_2_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_726633655bd32252eb297b924c81aa1b` `chunk_id=srcchunk_71643d8e3c02fac1174d22e0754583b0` `native_locator=slack:C0547N89JUB:1770276983.935799:1770765819.768759` `source_timestamp=2026-02-10T23:23:39Z`
- It is hypothesized that these other subdomains were added before Cloudflare proxy was enforced, avoiding the multi-level wildcard SSL limitation. `claim:claim_2_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_e5fe9c52e38e284597f96f38a8a3afbd` `source_revision_id=srcrev_9e43736908eefd82b2d53cb75d900a3f` `chunk_id=srcchunk_47ed5c51d71e9a440b863e7297da17ca` `native_locator=slack:C0547N89JUB:1770276983.935799:1770688969.911039` `source_timestamp=2026-02-10T02:02:49Z`

## Related Pages

- `decision-aeneid-portal-ssl-proxy`

## Sources

- `source_document_id`: `srcdoc_e5fe9c52e38e284597f96f38a8a3afbd`
- `source_revision_id`: `srcrev_7e0fdddf7fbfe857933e9fc1d02d6139`
