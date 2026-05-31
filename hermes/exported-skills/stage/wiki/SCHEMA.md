# Company Wiki Schema

This file is generated deterministically by Platform. LLM compiler calls may propose structured page data, claims, citations, conflicts, owners, and open questions, but they cannot change this schema.

## Page Roots

- `pages/` contains synthesized company wiki pages. These are the primary files Hermes should read first.
- `sources/` contains source evidence pages when evidence mode is enabled. These are audit material, not the canonical synthesis.
- `index.md` is the generated catalog.
- `log.md` is the generated publish and repair timeline.

## Synthesis Pages

Every factual bullet is rendered from a validated claim object. Every claim must cite existing source chunks from Platform Postgres. Conflict sections preserve cited disagreement rather than overwriting it.
