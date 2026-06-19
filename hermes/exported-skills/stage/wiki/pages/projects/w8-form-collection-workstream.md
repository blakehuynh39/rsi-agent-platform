---
title: "W8 Form Collection Workstream"
type: "project"
slug: "projects/w8-form-collection-workstream"
freshness: "2026-05-22T20:05:02Z"
tags:
  - "legal"
  - "onboarding"
  - "payments"
  - "tax"
owners:
  - "U09QGMMUDPC"
  - "U0AQZPN6ZQV"
source_revision_ids:
  - "srcrev_c1faffaf4d0adc9ccf177cb949c2cfb2"
  - "srcrev_c94f64e6f1ce6da708f1ffa2c4958752"
  - "srcrev_ca86e56d2094f2a34672628e8b1a346a"
  - "srcrev_e8bf2824e835de90706d93511a46edc6"
  - "srcrev_f22f4cdbb82b27ca1950f2b175e972aa"
conflict_state: "none"
---

# W8 Form Collection Workstream

## Summary

Workstream to ensure compliant collection of W8/W9 tax forms for payment processing, addressing legal risks and Stripe integration delays. Exploring alternatives like TaxBit and other products.

## Claims

- W8 form collection is legally sensitive; incorrect forms could lead to legal trouble if audited. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7a72e1479b8249da986cf8c356640fad` `source_revision_id=srcrev_e8bf2824e835de90706d93511a46edc6` `chunk_id=srcchunk_ffcef26c930b5d695a170f3df0a85be1` `native_locator=slack:C0AL7EKNHDF:1779477709.232499:1779477709.232499` `source_timestamp=2026-05-22T19:21:49Z`
- A dedicated workstream is needed to track W8BEN on Stripe, with daily updates and meeting invitations to U09QGMMUDPC and other stakeholders. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7a72e1479b8249da986cf8c356640fad` `source_revision_id=srcrev_e8bf2824e835de90706d93511a46edc6` `chunk_id=srcchunk_ffcef26c930b5d695a170f3df0a85be1` `native_locator=slack:C0AL7EKNHDF:1779477709.232499:1779477709.232499` `source_timestamp=2026-05-22T19:21:49Z`
- U0AQZPN6ZQV is appointed as the point person to coordinate and work cross-functionally with legal to unblock payments. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7a72e1479b8249da986cf8c356640fad` `source_revision_id=srcrev_e8bf2824e835de90706d93511a46edc6` `chunk_id=srcchunk_ffcef26c930b5d695a170f3df0a85be1` `native_locator=slack:C0AL7EKNHDF:1779477709.232499:1779477709.232499` `source_timestamp=2026-05-22T19:21:49Z`
- Payments rollout is dependent on proper completion of W8 form collection. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7a72e1479b8249da986cf8c356640fad` `source_revision_id=srcrev_e8bf2824e835de90706d93511a46edc6` `chunk_id=srcchunk_ffcef26c930b5d695a170f3df0a85be1` `native_locator=slack:C0AL7EKNHDF:1779477709.232499:1779477709.232499` `source_timestamp=2026-05-22T19:21:49Z`
- Stripe's W8/W9 is a private beta and their teams are new to it, with a fractured sales process (multiple contact transfers). `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7a72e1479b8249da986cf8c356640fad` `source_revision_id=srcrev_c94f64e6f1ce6da708f1ffa2c4958752` `chunk_id=srcchunk_0cfaa69dc64b70b518bf918d595f215f` `native_locator=slack:C0AL7EKNHDF:1779477709.232499:1779478197.703859` `source_timestamp=2026-05-22T19:29:57Z`
- Stable coin payouts process also slow, involving multiple Stripe departments that don't coordinate, and onboarding on private beta products is per-case approval. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7a72e1479b8249da986cf8c356640fad` `source_revision_id=srcrev_f22f4cdbb82b27ca1950f2b175e972aa` `chunk_id=srcchunk_69c72aa99a135c526e46fd43879eba09` `native_locator=slack:C0AL7EKNHDF:1779477709.232499:1779478381.648209` `source_timestamp=2026-05-22T19:33:01Z`
- TaxBit was contacted as a temporary alternative; quoted $10,000 for a few hundred users; negotiation to reduce to cents per user initiated. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7a72e1479b8249da986cf8c356640fad` `source_revision_id=srcrev_c1faffaf4d0adc9ccf177cb949c2cfb2` `chunk_id=srcchunk_eb31d885b469a07718bb0c1293f40aaf` `native_locator=slack:C0AL7EKNHDF:1779477709.232499:1779478732.949899` `source_timestamp=2026-05-22T19:38:52Z`
- Legal requirements for payments include KYC/tax forms; while waiting for Stripe, explore alternatives: KLED, Mercor, Handshake by simulating onboarding/withdrawal for India/foreign country. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_7a72e1479b8249da986cf8c356640fad` `source_revision_id=srcrev_ca86e56d2094f2a34672628e8b1a346a` `chunk_id=srcchunk_edcbe782c85a6ee38d7f1b88a58b7fa3` `native_locator=slack:C0AL7EKNHDF:1779477709.232499:1779480302.052459` `source_timestamp=2026-05-22T20:05:02Z`

## Related Pages

- `stripe-w8-w9-private-beta`
- `taxbit-alternative`
- `w8-form-collection-open-issues`

## Sources

- `source_document_id`: `srcdoc_7a72e1479b8249da986cf8c356640fad`
- `source_revision_id`: `srcrev_ca86e56d2094f2a34672628e8b1a346a`
