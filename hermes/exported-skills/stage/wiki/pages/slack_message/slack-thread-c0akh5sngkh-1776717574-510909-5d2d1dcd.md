---
title: "Slack thread C0AKH5SNGKH 1776717574.510909"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_e72070d8fa0bf0a17147aa8a5d2d1dcd"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776717574.510909"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776717574.510909"
source_revision_ids:
  - "srcrev_09ed05eb3b8332900252afd56591f419"
  - "srcrev_1552730a0abac9bcb2eac1867d8bc4a6"
  - "srcrev_226e352cc625d34abaf011491e5f3053"
  - "srcrev_2f40d0f6acfd4d4c50afb33a022c4e16"
  - "srcrev_4d86483a985c4d72e871eaea2d08cf21"
  - "srcrev_835ecfc9bd12d81091cb56b8b07ba360"
  - "srcrev_bb01dffce9b814d4ba4bc51a7a0b94b1"
  - "srcrev_dc91927243eda6bbcd15f34679f016ce"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1776717574.510909

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_e72070d8fa0bf0a17147aa8a5d2d1dcd`
- `source_revision_id`: `srcrev_1552730a0abac9bcb2eac1867d8bc4a6`
- `chunk_id`: `srcchunk_848b7afb192263b417e61531dad18d34`
- `native_locator`: `slack:C0AKH5SNGKH:1776717574.510909:1776717574.510909`

<@U083MMT1771> <@U0772SH7BRA> what's the use case of the `numo/depin-ip-registration` deployment versus `story/depin-ip-registration`? Could we use it as a pre-prod environment where the IPs are registered on mainnet?

### Citation 2

- `source_document_id`: `srcdoc_e72070d8fa0bf0a17147aa8a5d2d1dcd`
- `source_revision_id`: `srcrev_4d86483a985c4d72e871eaea2d08cf21`
- `chunk_id`: `srcchunk_7d53410b3ab498569dd70c41ba4947d6`
- `native_locator`: `slack:C0AKH5SNGKH:1776717574.510909:1776717900.666299`

you should use story one
<https://github.com/storyprotocol/story-deployments/commits/main/story/depin-ip-registration>

### Citation 3

- `source_document_id`: `srcdoc_e72070d8fa0bf0a17147aa8a5d2d1dcd`
- `source_revision_id`: `srcrev_bb01dffce9b814d4ba4bc51a7a0b94b1`
- `chunk_id`: `srcchunk_ca461c125299c09b9674b47e09d46f22`
- `native_locator`: `slack:C0AKH5SNGKH:1776717574.510909:1776717911.047209`

the bot been pushing updates there

### Citation 4

- `source_document_id`: `srcdoc_e72070d8fa0bf0a17147aa8a5d2d1dcd`
- `source_revision_id`: `srcrev_dc91927243eda6bbcd15f34679f016ce`
- `chunk_id`: `srcchunk_19dc7d7cabb68c7780444ca571ab3912`
- `native_locator`: `slack:C0AKH5SNGKH:1776717574.510909:1776718130.610839`

Could the bot push updates on both so that we can run tests on both testnet and mainnet?

### Citation 5

- `source_document_id`: `srcdoc_e72070d8fa0bf0a17147aa8a5d2d1dcd`
- `source_revision_id`: `srcrev_835ecfc9bd12d81091cb56b8b07ba360`
- `chunk_id`: `srcchunk_327712dee123e5557b1d811ffb33c38f`
- `native_locator`: `slack:C0AKH5SNGKH:1776717574.510909:1776718155.939719`

Or is there any reason why we don't want to push updates on the `numo` env

### Citation 6

- `source_document_id`: `srcdoc_e72070d8fa0bf0a17147aa8a5d2d1dcd`
- `source_revision_id`: `srcrev_2f40d0f6acfd4d4c50afb33a022c4e16`
- `chunk_id`: `srcchunk_755a2297635fae5dc2c4811d05e45474`
- `native_locator`: `slack:C0AKH5SNGKH:1776717574.510909:1776718255.450579`

the story one was made before, it has both staging and prod

### Citation 7

- `source_document_id`: `srcdoc_e72070d8fa0bf0a17147aa8a5d2d1dcd`
- `source_revision_id`: `srcrev_226e352cc625d34abaf011491e5f3053`
- `chunk_id`: `srcchunk_21fa64b0f26dbdf593e739c7df042e3e`
- `native_locator`: `slack:C0AKH5SNGKH:1776717574.510909:1776718336.865849`

you can migrate numo over if needed otherwise it'll take a lot of effort to migrate to numo

### Citation 8

- `source_document_id`: `srcdoc_e72070d8fa0bf0a17147aa8a5d2d1dcd`
- `source_revision_id`: `srcrev_09ed05eb3b8332900252afd56591f419`
- `chunk_id`: `srcchunk_7d417f2bdffdbb7b31a16b5686c890c9`
- `native_locator`: `slack:C0AKH5SNGKH:1776717574.510909:1776718343.760369`

the deployment on argo/k8 is reading from the story one

