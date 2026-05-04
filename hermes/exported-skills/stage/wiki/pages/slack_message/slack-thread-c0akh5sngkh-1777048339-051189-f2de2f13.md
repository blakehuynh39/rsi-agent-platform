---
title: "Slack thread C0AKH5SNGKH 1777048339.051189"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_933ff03dd9081886dbc0006bf2de2f13"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1777048339.051189"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1777048339.051189"
source_revision_ids:
  - "srcrev_04e4ac12d0db7aabb5fc9f34f6335b32"
  - "srcrev_0ff023f5a35b3f220c0a57f5aac95966"
  - "srcrev_1ca69d42914db5c594bceca11674b1ec"
  - "srcrev_833d6962f3c7efec26aa8824b6696b6c"
  - "srcrev_95d59989b7a3b70e3d8ef15fa5b2cedd"
  - "srcrev_974805440f16eda292c8ef982f250c77"
  - "srcrev_b01105c2bc2de82bbd1300ae3bafb7f8"
  - "srcrev_bbeef83062d7061b7c8c1d827be198f3"
  - "srcrev_c9a7cfe6d34c70e50a6c342797455b66"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1777048339.051189

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_933ff03dd9081886dbc0006bf2de2f13`
- `source_revision_id`: `srcrev_974805440f16eda292c8ef982f250c77`
- `chunk_id`: `srcchunk_f51281a95e22c07e2929b41acef742f0`
- `native_locator`: `slack:C0AKH5SNGKH:1777048339.051189:1777048339.051189`

<@U04L0DD6B6F> <@U05A515NBFC> rn once a campain reaches its target, it shows negative available tasks.
How do we want to handle this when a campaign is completed? Should we mark the campaign as expired, or show it as no longer available?

### Citation 2

- `source_document_id`: `srcdoc_933ff03dd9081886dbc0006bf2de2f13`
- `source_revision_id`: `srcrev_0ff023f5a35b3f220c0a57f5aac95966`
- `chunk_id`: `srcchunk_286860cddb5c4d7379a971a50d132b24`
- `native_locator`: `slack:C0AKH5SNGKH:1777048339.051189:1777063725.274329`

<@U04L0DD6B6F> <@U067QP5PD6J> bump up this

### Citation 3

- `source_document_id`: `srcdoc_933ff03dd9081886dbc0006bf2de2f13`
- `source_revision_id`: `srcrev_833d6962f3c7efec26aa8824b6696b6c`
- `chunk_id`: `srcchunk_4078d79c80f050633a8df98ece87a48c`
- `native_locator`: `slack:C0AKH5SNGKH:1777048339.051189:1777063782.302059`

I think for initial phase, we should show but gray out with overlay and change the "start" button to "expired" gray button

### Citation 4

- `source_document_id`: `srcdoc_933ff03dd9081886dbc0006bf2de2f13`
- `source_revision_id`: `srcrev_95d59989b7a3b70e3d8ef15fa5b2cedd`
- `chunk_id`: `srcchunk_9bcb9897e41410dffb7be58857530508`
- `native_locator`: `slack:C0AKH5SNGKH:1777048339.051189:1777063839.308679`

and yeah negative task is likely backend issue? for frontend, we can just cap the value to the max submission so it always shows 0. But we would need to investigate why the system allows more submissions than max target

### Citation 5

- `source_document_id`: `srcdoc_933ff03dd9081886dbc0006bf2de2f13`
- `source_revision_id`: `srcrev_04e4ac12d0db7aabb5fc9f34f6335b32`
- `chunk_id`: `srcchunk_e2e476bbe7052381cafe437540057dfb`
- `native_locator`: `slack:C0AKH5SNGKH:1777048339.051189:1777063924.706879`

Sounds good, I’ve noted it down. We can look into it and fix this before launch

### Citation 6

- `source_document_id`: `srcdoc_933ff03dd9081886dbc0006bf2de2f13`
- `source_revision_id`: `srcrev_c9a7cfe6d34c70e50a6c342797455b66`
- `chunk_id`: `srcchunk_f43a4fe5f1a7a4b4b7b07fa7fd7eea85`
- `native_locator`: `slack:C0AKH5SNGKH:1777048339.051189:1777063996.762379`

<https://storyprotocol.slack.com/archives/C0ARB528XEZ/p1777063964018719>

### Citation 7

- `source_document_id`: `srcdoc_933ff03dd9081886dbc0006bf2de2f13`
- `source_revision_id`: `srcrev_b01105c2bc2de82bbd1300ae3bafb7f8`
- `chunk_id`: `srcchunk_ec929c0883dba6af1e96872d2fcb2ce7`
- `native_locator`: `slack:C0AKH5SNGKH:1777048339.051189:1777064032.881539`

but also are we rewarding submissions AFTER we already reach our goal?

### Citation 8

- `source_document_id`: `srcdoc_933ff03dd9081886dbc0006bf2de2f13`
- `source_revision_id`: `srcrev_1ca69d42914db5c594bceca11674b1ec`
- `chunk_id`: `srcchunk_2151ed567a5e42f9b36e171174eb84b0`
- `native_locator`: `slack:C0AKH5SNGKH:1777048339.051189:1777064040.013089`

what's the logic rn do you know?

### Citation 9

- `source_document_id`: `srcdoc_933ff03dd9081886dbc0006bf2de2f13`
- `source_revision_id`: `srcrev_bbeef83062d7061b7c8c1d827be198f3`
- `chunk_id`: `srcchunk_02c0ac6b4a9816e5918c7eec2939332d`
- `native_locator`: `slack:C0AKH5SNGKH:1777048339.051189:1777064336.387179`

if we allow submissions after campaign task amount is full, the code still rewards after reaching the goal.

