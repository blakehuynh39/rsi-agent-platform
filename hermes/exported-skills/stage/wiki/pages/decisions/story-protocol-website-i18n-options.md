---
title: "Story Protocol Website i18n Options"
type: "decision"
slug: "decisions/story-protocol-website-i18n-options"
freshness: "2026-05-05T06:26:22Z"
tags:
  - "contentful"
  - "i18n"
  - "translation"
  - "website"
  - "weglot"
owners: []
source_revision_ids:
  - "srcrev_02911d8b009b8b21a9e41253e7b9b122"
conflict_state: "none"
---

# Story Protocol Website i18n Options

## Summary

Evaluation of two approaches for internationalizing the Story Protocol website: a custom solution using Contentful's built-in locale support, and an automated no-code solution using Weglot. The Contentful approach offers more control but is limited by plan tiers and requires significant dev work. Weglot is simpler and cheaper but lacks a native Contentful integration.

## Claims

- Contentful's free plan allows only 2 locales, the Basic plan at $300/month allows 4 locales, and beyond that requires a custom negotiated plan. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Website-i18n-16f9f6989062425380f6ec2fb0303382) `source_document_id=srcdoc_fc231b3d3a25a28a3aaf4dcd7d82676d` `source_revision_id=srcrev_02911d8b009b8b21a9e41253e7b9b122` `chunk_id=srcchunk_ab44f04d26e536aff1fb48eea8fe96f1` `native_locator=https://www.notion.so/Story-Protocol-Website-i18n-16f9f6989062425380f6ec2fb0303382` `source_timestamp=2026-05-05T06:26:22Z`
- Implementing i18n via Contentful requires dev work to restructure the directory for locale-prefixed URLs (e.g., storyprotocol.xyz/en/) and manually handling default language detection and preference saving. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Website-i18n-16f9f6989062425380f6ec2fb0303382) `source_document_id=srcdoc_fc231b3d3a25a28a3aaf4dcd7d82676d` `source_revision_id=srcrev_02911d8b009b8b21a9e41253e7b9b122` `chunk_id=srcchunk_ab44f04d26e536aff1fb48eea8fe96f1` `native_locator=https://www.notion.so/Story-Protocol-Website-i18n-16f9f6989062425380f6ec2fb0303382` `source_timestamp=2026-05-05T06:26:22Z`
- Weglot is an automated no/low-code translation solution priced at $87/month for 5 languages, but it has no integration for Contentful. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Website-i18n-16f9f6989062425380f6ec2fb0303382) `source_document_id=srcdoc_fc231b3d3a25a28a3aaf4dcd7d82676d` `source_revision_id=srcrev_02911d8b009b8b21a9e41253e7b9b122` `chunk_id=srcchunk_ab44f04d26e536aff1fb48eea8fe96f1` `native_locator=https://www.notion.so/Story-Protocol-Website-i18n-16f9f6989062425380f6ec2fb0303382` `source_timestamp=2026-05-05T06:26:22Z`
- With the Contentful approach, content must be translated and updated manually through the Contentful CMS. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Website-i18n-16f9f6989062425380f6ec2fb0303382) `source_document_id=srcdoc_fc231b3d3a25a28a3aaf4dcd7d82676d` `source_revision_id=srcrev_02911d8b009b8b21a9e41253e7b9b122` `chunk_id=srcchunk_ab44f04d26e536aff1fb48eea8fe96f1` `native_locator=https://www.notion.so/Story-Protocol-Website-i18n-16f9f6989062425380f6ec2fb0303382` `source_timestamp=2026-05-05T06:26:22Z`

## Open Questions

- How many target languages are required, and does that fit within Contentful's plan limits or Weglot's pricing tiers?
- Which approach (Contentful custom i18n vs Weglot) will be selected for the Story Protocol website?
- Will the lack of a native Weglot-Contentful integration be a blocker, or can it be worked around with a script include?

## Sources

- `source_document_id`: `srcdoc_fc231b3d3a25a28a3aaf4dcd7d82676d`
- `source_revision_id`: `srcrev_02911d8b009b8b21a9e41253e7b9b122`
- `source_url`: [Notion source](https://www.notion.so/Story-Protocol-Website-i18n-16f9f6989062425380f6ec2fb0303382)
