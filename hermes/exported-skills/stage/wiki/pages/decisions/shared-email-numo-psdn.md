---
title: "Numo Shared Email (numo@psdn.ai)"
type: "decision"
slug: "decisions/shared-email-numo-psdn"
freshness: "2026-06-01T18:37:49Z"
tags:
  - "distribution-list"
  - "email"
  - "numo"
  - "onboarding"
  - "services"
owners:
  - "U04L0DD6B6F"
  - "U067QP5PD6J"
source_revision_ids:
  - "srcrev_2cc4434ca7d6f29cd81d4ab0f4ffd84d"
  - "srcrev_2ea15a700cb44b3f4f3a52f5eb3f25bd"
  - "srcrev_31d431c35b21736079307a1643f8a0c6"
  - "srcrev_3f64291d165de4565e4bf5d7c3bae295"
  - "srcrev_4c736bbe3a11c9e01c69f005430d4dce"
  - "srcrev_645f2bd828bdd24e8fffd5e3d2b6419e"
  - "srcrev_6b3b48f2d5aabb3ec087b90bb2fbc9a6"
  - "srcrev_6be7750d98188f16f4c4d1907b119109"
  - "srcrev_6d929d59fe8366511f8f02e5b3a2d797"
  - "srcrev_83cbc41ed8d34d77956c06bb9b6b36a3"
  - "srcrev_90437eb700da2789b3d3c072138fa16e"
  - "srcrev_ae4b7cb90a877716530cf600b23be773"
  - "srcrev_b1596e84d41672e132ccf87693f0b027"
  - "srcrev_d9b5466a0f5fd52f82ce396c803bdd9c"
conflict_state: "none"
---

# Numo Shared Email (numo@psdn.ai)

## Summary

The team decided to create and use numo@psdn.ai as a shared email for registering and managing Numo-specific service accounts. A Google Group distribution list was created and managed by designated members. The email has been invited to multiple services, with some services deemed unnecessary and adjustments made (e.g., Castle.io user limits).

## Claims

- The email address numo@psdn.ai was requested to be created and used for registering Numo-specific services (e.g., Apple/Android store accounts, CloudFlare, mobile SDK accounts). `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_2ea15a700cb44b3f4f3a52f5eb3f25bd` `chunk_id=srcchunk_291a33d6c2de3e6c79e12f75c8ec9309` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780195325.479139` `source_timestamp=2026-05-31T02:42:05Z`
- The requester initially lacked permission to create the email. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_6be7750d98188f16f4c4d1907b119109` `chunk_id=srcchunk_5d36102d6f799834f70bf95d1b1509da` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780280039.719429` `source_timestamp=2026-06-01T02:13:59Z`
- A distribution list (Google Group) was suggested as the approach instead of a standard email account. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_31d431c35b21736079307a1643f8a0c6` `chunk_id=srcchunk_b423e362ee6f47642c7b5d83344831c0` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780331133.693629` `source_timestamp=2026-06-01T16:25:33Z`
- The Google Group numo@psdn.ai was created, with managers U067QP5PD6J and U04L0DD6B6F. Management URL: https://groups.google.com/a/psdn.ai/g/numo. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_2cc4434ca7d6f29cd81d4ab0f4ffd84d` `chunk_id=srcchunk_8e7e4e940e4373d9e7fca5f8885bef5f` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780331322.625789` `source_timestamp=2026-06-01T16:28:42Z`
- The email numo@psdn.ai should be invited to the following services: CloudFlare, Dynamic, Elevenlabs API, Sarvam, Castle.io, LinkedIn, Stripe, Resend, Temporal, Beehiv, Alchemy RPC. A Notion checklist tracks progress: https://www.notion.so/storyprotocol/Jed-Korean-Team-Onboarding-Checklist-Credential-access-372051299a5480dc8322d95db444fd91 `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_b1596e84d41672e132ccf87693f0b027` `chunk_id=srcchunk_39ab9ae23a5ec7f61053654b1bce6f57` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780331952.993489` `source_timestamp=2026-06-01T16:39:12Z`
- Temporal is not needed for numo@psdn.ai. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_3f64291d165de4565e4bf5d7c3bae295` `chunk_id=srcchunk_d19efc7772fcd9809bc2eacfc4f9b9f2` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780333054.605029` `source_timestamp=2026-06-01T16:57:34Z`
- Sarvam is not currently being used for Indic language analysis, and therefore not needed for numo@psdn.ai. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_90437eb700da2789b3d3c072138fa16e` `chunk_id=srcchunk_3937a3fc133e11c1c28dacdf20e47570` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780333835.960169` `source_timestamp=2026-06-01T17:10:35Z`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_645f2bd828bdd24e8fffd5e3d2b6419e` `chunk_id=srcchunk_635f6df5cf7a38ab24b752e8b133777d` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780336333.993129` `source_timestamp=2026-06-01T17:52:13Z`
- Castle.io has a plan limit of 5 users. To add numo@psdn.ai, the team decided to remove an existing user (suggested: Blake) and invited numo@psdn.ai as owner. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_6d929d59fe8366511f8f02e5b3a2d797` `chunk_id=srcchunk_cccd01621434c0bde4801e3e0b045ce5` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780336500.173849` `source_timestamp=2026-06-01T17:55:00Z`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_ae4b7cb90a877716530cf600b23be773` `chunk_id=srcchunk_148ad17abf7dbd08ad5e3a2cbea18a11` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780336560.127669` `source_timestamp=2026-06-01T17:56:00Z`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_6b3b48f2d5aabb3ec087b90bb2fbc9a6` `chunk_id=srcchunk_a6e18e87fadf02710e1549844322cc97` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780338921.326669` `source_timestamp=2026-06-01T18:35:21Z`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_83cbc41ed8d34d77956c06bb9b6b36a3` `chunk_id=srcchunk_93d30baccc258340e7781d3dbfd8efed` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780338933.713019` `source_timestamp=2026-06-01T18:35:33Z`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_4c736bbe3a11c9e01c69f005430d4dce` `chunk_id=srcchunk_c915d19b18eb846bb3bb256b8ced732e` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780339069.575659` `source_timestamp=2026-06-01T18:37:49Z`
- Access to numo@psdn.ai can be shared via 1Password so team members can log in. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_106d05d53a0f147f5bbdcc2a304aa295` `source_revision_id=srcrev_d9b5466a0f5fd52f82ce396c803bdd9c` `chunk_id=srcchunk_41581b29f8dd782d3ec7ad7830960a3c` `native_locator=slack:C0AL7EKNHDF:1780195325.479139:1780338963.649169` `source_timestamp=2026-06-01T18:36:03Z`

## Open Questions

- Has the numo@psdn.ai email been invited to all listed services? The Notion checklist (https://www.notion.so/storyprotocol/Jed-Korean-Team-Onboarding-Checklist-Credential-access-372051299a5480dc8322d95db444fd91) should be checked for current status.

## Sources

- `source_document_id`: `srcdoc_106d05d53a0f147f5bbdcc2a304aa295`
- `source_revision_id`: `srcrev_4c736bbe3a11c9e01c69f005430d4dce`
