---
title: "Vietnamese Font Fallback for Diacritic Rendering"
type: "decision"
slug: "decisions/vietnamese-font-fallback"
freshness: "2026-05-07T03:31:30Z"
tags:
  - "a11y"
  - "css"
  - "font"
  - "i18n"
  - "vietnamese"
owners:
  - "U05A515NBFC"
  - "U0772SH7BRA"
  - "U0ASDQKU3UL"
  - "U0AU3DWLVE2"
source_revision_ids:
  - "srcrev_47fa86429551b5e84111e40787d5bf18"
  - "srcrev_4ce4c185f57f82d3806e8872aac5b717"
  - "srcrev_54995778971c2e2d8fce8e6155d02dd5"
  - "srcrev_5f75cd3cc551ea91c4394f624d247978"
  - "srcrev_87c66b8e93aed982c154dc714e4c8f20"
  - "srcrev_ad32c5597489537467b4802932374a38"
  - "srcrev_d0ff532e1e39f2e6984d629814ad077b"
conflict_state: "none"
---

# Vietnamese Font Fallback for Diacritic Rendering

## Summary

Vietnamese text rendered incorrectly due to missing stacked diacritic support in STK Bureau Sans/Serif fonts. A CSS override was deployed using Spectral as the fallback font for lang=vi, fixing the issue via PR #243/244.

## Claims

- STK Bureau Sans and Serif fonts lack proper glyphs for Vietnamese stacked diacritics, causing incorrect rendering of accents. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_00248047054029d90d30928191afb4d7` `source_revision_id=srcrev_47fa86429551b5e84111e40787d5bf18` `chunk_id=srcchunk_dbfdec582a18d8b819a22c63a03af2fb` `native_locator=slack:C0AL7EKNHDF:1778118005.007429:1778118525.520189` `source_timestamp=2026-05-07T01:48:45Z`
- The issue was reproducible: Image 1 showed misaligned diacritics with brand fonts, Image 2 showed correct rendering with system serif/sans-serif fonts. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_00248047054029d90d30928191afb4d7` `source_revision_id=srcrev_47fa86429551b5e84111e40787d5bf18` `chunk_id=srcchunk_dbfdec582a18d8b819a22c63a03af2fb` `native_locator=slack:C0AL7EKNHDF:1778118005.007429:1778118525.520189` `source_timestamp=2026-05-07T01:48:45Z`
- Two frontend solutions were proposed: (A) CSS override by locale targeting html[lang='vi'], (B) @font-face unicode-range to restrict brand fonts to Latin characters only. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_00248047054029d90d30928191afb4d7` `source_revision_id=srcrev_47fa86429551b5e84111e40787d5bf18` `chunk_id=srcchunk_dbfdec582a18d8b819a22c63a03af2fb` `native_locator=slack:C0AL7EKNHDF:1778118005.007429:1778118525.520189` `source_timestamp=2026-05-07T01:48:45Z`
- A PR (pull request) was created to implement the Vietnamese fallback font, initially using the CSS override approach. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_00248047054029d90d30928191afb4d7` `source_revision_id=srcrev_5f75cd3cc551ea91c4394f624d247978` `chunk_id=srcchunk_f989ab897cfa666a9104d4bad828be92` `native_locator=slack:C0AL7EKNHDF:1778118005.007429:1778121515.871959` `source_timestamp=2026-05-07T02:38:35Z`
- Spectral font was suggested as an alternative fallback because it looks closer to the brand style and is free. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_00248047054029d90d30928191afb4d7` `source_revision_id=srcrev_54995778971c2e2d8fce8e6155d02dd5` `chunk_id=srcchunk_299bce835b37461c6bcdf0bc89563279` `native_locator=slack:C0AL7EKNHDF:1778118005.007429:1778121674.568839` `source_timestamp=2026-05-07T02:41:14Z`
- The PR was updated to use Spectral as the Vietnamese fallback font, and preview deployments confirmed the fix. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_00248047054029d90d30928191afb4d7` `source_revision_id=srcrev_4ce4c185f57f82d3806e8872aac5b717` `chunk_id=srcchunk_9316d02396af2b10fc87503498f207d9` `native_locator=slack:C0AL7EKNHDF:1778118005.007429:1778121791.321909` `source_timestamp=2026-05-07T02:43:11Z`
  - citation: `source_document_id=srcdoc_00248047054029d90d30928191afb4d7` `source_revision_id=srcrev_d0ff532e1e39f2e6984d629814ad077b` `chunk_id=srcchunk_e07a43dcc7ae4d08ab30a05a89589db4` `native_locator=slack:C0AL7EKNHDF:1778118005.007429:1778123707.948469` `source_timestamp=2026-05-07T03:15:07Z`
- The fix was deployed to production after confirming it is a safe CSS-only change. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_00248047054029d90d30928191afb4d7` `source_revision_id=srcrev_ad32c5597489537467b4802932374a38` `chunk_id=srcchunk_9055bc33f947716bf52594d7ee6a7a33` `native_locator=slack:C0AL7EKNHDF:1778118005.007429:1778124397.922679` `source_timestamp=2026-05-07T03:26:37Z`
  - citation: `source_document_id=srcdoc_00248047054029d90d30928191afb4d7` `source_revision_id=srcrev_87c66b8e93aed982c154dc714e4c8f20` `chunk_id=srcchunk_6e26ed17955eb3e55f15d8db10c3225e` `native_locator=slack:C0AL7EKNHDF:1778118005.007429:1778124690.935139` `source_timestamp=2026-05-07T03:31:30Z`

## Open Questions

- Foundry response about adding Vietnamese glyphs to STK Bureau fonts is pending.

## Related Pages

- `internationalization-i18n-strategy`
- `typography-design-system`

## Sources

- `source_document_id`: `srcdoc_00248047054029d90d30928191afb4d7`
- `source_revision_id`: `srcrev_87c66b8e93aed982c154dc714e4c8f20`
