---
title: "WF Metadata Design"
type: "project"
slug: "projects/wf-metadata-design"
freshness: "2026-05-05T05:41:58Z"
tags:
  - "character-design"
  - "game-design"
  - "metadata"
  - "nft"
owners: []
source_revision_ids:
  - "srcrev_facecae5ec4c25990af94f8e56f7ff14"
conflict_state: "none"
---

# WF Metadata Design

## Summary

Design document for character metadata in WF, covering traits, attributes, and artifact-finding mechanics.

## Claims

- Artifact-finding tools trait class determines when characters receive Artifact Shard airdrops. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f) `source_document_id=srcdoc_7d8204ad78311a5b7975c02cadb04a77` `source_revision_id=srcrev_facecae5ec4c25990af94f8e56f7ff14` `chunk_id=srcchunk_865ce42aa537c24d860376004dc2ea6e` `native_locator=https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f` `source_timestamp=2026-05-05T05:41:58Z`
- Some artifact-finding tools find artifacts faster but yield smaller artifacts. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f) `source_document_id=srcdoc_7d8204ad78311a5b7975c02cadb04a77` `source_revision_id=srcrev_facecae5ec4c25990af94f8e56f7ff14` `chunk_id=srcchunk_865ce42aa537c24d860376004dc2ea6e` `native_locator=https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f` `source_timestamp=2026-05-05T05:41:58Z`
- Characters have a set number of points to spend on attributes (strength, dexterity, etc) resembling D&D, tied to NFT mechanics, and can be re-rolled. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f) `source_document_id=srcdoc_7d8204ad78311a5b7975c02cadb04a77` `source_revision_id=srcrev_facecae5ec4c25990af94f8e56f7ff14` `chunk_id=srcchunk_865ce42aa537c24d860376004dc2ea6e` `native_locator=https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f` `source_timestamp=2026-05-05T05:41:58Z`
- Families of traits may include Kinds from different home planets, or Wides/Wisps/Warps, and form Teams. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f) `source_document_id=srcdoc_7d8204ad78311a5b7975c02cadb04a77` `source_revision_id=srcrev_facecae5ec4c25990af94f8e56f7ff14` `chunk_id=srcchunk_865ce42aa537c24d860376004dc2ea6e` `native_locator=https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f` `source_timestamp=2026-05-05T05:41:58Z`
- 1/1s exist for SP team to use as profile pictures (PFPs). `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f) `source_document_id=srcdoc_7d8204ad78311a5b7975c02cadb04a77` `source_revision_id=srcrev_facecae5ec4c25990af94f8e56f7ff14` `chunk_id=srcchunk_865ce42aa537c24d860376004dc2ea6e` `native_locator=https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f` `source_timestamp=2026-05-05T05:41:58Z`
- Visual traits include background/location, aiming for coherent visual character and relatability during QA. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f) `source_document_id=srcdoc_7d8204ad78311a5b7975c02cadb04a77` `source_revision_id=srcrev_facecae5ec4c25990af94f8e56f7ff14` `chunk_id=srcchunk_865ce42aa537c24d860376004dc2ea6e` `native_locator=https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f` `source_timestamp=2026-05-05T05:41:58Z`
- Metadata traits include Home Planet, Tarot Cards, Role on Ship, Short Description, Inventory (Equipment, Artifacts Owned). `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f) `source_document_id=srcdoc_7d8204ad78311a5b7975c02cadb04a77` `source_revision_id=srcrev_facecae5ec4c25990af94f8e56f7ff14` `chunk_id=srcchunk_865ce42aa537c24d860376004dc2ea6e` `native_locator=https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f` `source_timestamp=2026-05-05T05:41:58Z`
- Traits are categorized between gamification and storytelling/identification purposes. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f) `source_document_id=srcdoc_7d8204ad78311a5b7975c02cadb04a77` `source_revision_id=srcrev_facecae5ec4c25990af94f8e56f7ff14` `chunk_id=srcchunk_865ce42aa537c24d860376004dc2ea6e` `native_locator=https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f` `source_timestamp=2026-05-05T05:41:58Z`

## Open Questions

- How will character attribute points be used in future game mechanics (e.g., D&D game or minigame)?
- How will the balance between gamification and storytelling/identification traits be decided?

## Sources

- `source_document_id`: `srcdoc_7d8204ad78311a5b7975c02cadb04a77`
- `source_revision_id`: `srcrev_facecae5ec4c25990af94f8e56f7ff14`
- `source_url`: [Notion source](https://www.notion.so/WF-metadata-WIP-c20ace479c064f7eb5ae7927d969626f)
