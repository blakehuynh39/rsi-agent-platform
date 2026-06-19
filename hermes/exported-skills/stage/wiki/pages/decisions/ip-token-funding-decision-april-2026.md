---
title: "IP Token Funding for IP Registration Wallets (April 2026 Decision)"
type: "decision"
slug: "decisions/ip-token-funding-decision-april-2026"
freshness: "2026-04-23T00:17:20Z"
tags:
  - "decision"
  - "funding"
  - "ip-tokens"
  - "registration"
  - "story-foundation"
owners: []
source_revision_ids:
  - "srcrev_2be1f72063e6a9815fd1ef6fc9c3623f"
  - "srcrev_8aca88ea74e36f72790df1b9bd8f59ea"
  - "srcrev_9117382373286cd8dc517c2da7780e96"
  - "srcrev_b565f546b5d27322d445f799398cb1af"
  - "srcrev_dc4a07276cf3ea107baff47f9e20558d"
conflict_state: "none"
---

# IP Token Funding for IP Registration Wallets (April 2026 Decision)

## Summary

In April 2026, the team decided how to fund IP tokens needed for large‑scale IP registration. After considering personal Brex purchases, the final decision was to request the funds from the Story Foundation.

## Claims

- One IP token is estimated to support up to 10,000 registrations, pending confirmation from mainnet tests (as of 2026-04-22). `claim:ip-registrations-per-token` `confidence:1.00`
  - citation: `source_document_id=srcdoc_76f4cc2b05d657771e2b6784912e5d5c` `source_revision_id=srcrev_dc4a07276cf3ea107baff47f9e20558d` `chunk_id=srcchunk_6d97cce0a3a251fbba265dd9b6e9d209` `native_locator=slack:C0AL7EKNHDF:1776889464.626269:1776892570.110369` `source_timestamp=2026-04-22T21:21:45Z`
  - citation: `source_document_id=srcdoc_76f4cc2b05d657771e2b6784912e5d5c` `source_revision_id=srcrev_b565f546b5d27322d445f799398cb1af` `chunk_id=srcchunk_6fbddda29e7d77f965fc41ab5f6a4518` `native_locator=slack:C0AL7EKNHDF:1776889464.626269:1776892803.134799` `source_timestamp=2026-04-22T21:20:03Z`
- To support a goal of ~2.5 million registrations, approximately 250 IP tokens are required. `claim:needed-ip-tokens` `confidence:1.00`
  - citation: `source_document_id=srcdoc_76f4cc2b05d657771e2b6784912e5d5c` `source_revision_id=srcrev_dc4a07276cf3ea107baff47f9e20558d` `chunk_id=srcchunk_6d97cce0a3a251fbba265dd9b6e9d209` `native_locator=slack:C0AL7EKNHDF:1776889464.626269:1776892570.110369` `source_timestamp=2026-04-22T21:21:45Z`
- The estimated cost for 250 IP tokens is approximately $125 USD. `claim:estimated-cost` `confidence:1.00`
  - citation: `source_document_id=srcdoc_76f4cc2b05d657771e2b6784912e5d5c` `source_revision_id=srcrev_8aca88ea74e36f72790df1b9bd8f59ea` `chunk_id=srcchunk_124aded04daeedd38bb68b1789e22419` `native_locator=slack:C0AL7EKNHDF:1776889464.626269:1776892936.546769` `source_timestamp=2026-04-22T21:22:16Z`
- For amounts above 100 IP tokens, the formal process is to request funds from the Story Foundation, though it can be slow. `claim:standard-foundation-process` `confidence:1.00`
  - citation: `source_document_id=srcdoc_76f4cc2b05d657771e2b6784912e5d5c` `source_revision_id=srcrev_2be1f72063e6a9815fd1ef6fc9c3623f` `chunk_id=srcchunk_ee853e395e549675dc5d78e27c0ef43a` `native_locator=slack:C0AL7EKNHDF:1776889464.626269:1776889775.475809` `source_timestamp=2026-04-22T20:29:46Z`
- In the past, engineers have personally purchased IP tokens using Brex cards and transferred them to test accounts as a workaround when the foundation process was delayed. `claim:past-brex-workaround` `confidence:1.00`
  - citation: `source_document_id=srcdoc_76f4cc2b05d657771e2b6784912e5d5c` `source_revision_id=srcrev_2be1f72063e6a9815fd1ef6fc9c3623f` `chunk_id=srcchunk_ee853e395e549675dc5d78e27c0ef43a` `native_locator=slack:C0AL7EKNHDF:1776889464.626269:1776889775.475809` `source_timestamp=2026-04-22T20:29:46Z`
- The decision was made to request the IP token funds from the Story Foundation instead of using personal Brex purchases. `claim:final-decision` `confidence:1.00`
  - citation: `source_document_id=srcdoc_76f4cc2b05d657771e2b6784912e5d5c` `source_revision_id=srcrev_9117382373286cd8dc517c2da7780e96` `chunk_id=srcchunk_87343f1210f0e0e6677c1d22084be78c` `native_locator=slack:C0AL7EKNHDF:1776889464.626269:1776903440.699659` `source_timestamp=2026-04-23T00:17:20Z`

## Open Questions

- A mechanism to dynamically add IP to wallets when registration goals are exceeded is still needed.
- Confirmation of the 10,000 registrations per IP token figure is pending mainnet tests (as of 2026-04-22).
- The process for requesting funds from the Foundation needs to be formalized; initial conversation with the Foundation side to be kicked off.

## Sources

- `source_document_id`: `srcdoc_76f4cc2b05d657771e2b6784912e5d5c`
- `source_revision_id`: `srcrev_9117382373286cd8dc517c2da7780e96`
