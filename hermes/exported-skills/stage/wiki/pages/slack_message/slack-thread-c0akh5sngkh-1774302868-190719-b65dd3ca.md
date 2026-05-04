---
title: "Slack thread C0AKH5SNGKH 1774302868.190719"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1774302868.190719"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1774302868.190719"
source_revision_ids:
  - "srcrev_334d99d986cccde5924ba8b4e7a579ae"
  - "srcrev_34ed52a9fdbecf7a2123ecf569133ad4"
  - "srcrev_3da9626a11c9d54efe130bac14aa8440"
  - "srcrev_73f7117fc86a725017e0e5af3c6919f6"
  - "srcrev_7b1ffba1bdec26b81883d7a25b170239"
  - "srcrev_a06ece44508438aa19d1a6c02f80709b"
  - "srcrev_a8d6b8b46dea4677e4f321c278dafe4c"
  - "srcrev_b98781cbbe49ec92662b663ae3a8f140"
  - "srcrev_caa45183e7c63e4a23da262ae7262c59"
  - "srcrev_df53cee0315f82179d433ecbbc8563d4"
  - "srcrev_e50e9892b9f1f1767ab4c7f2fba7cd77"
  - "srcrev_f3fdc20ad4e6341310b1c5278a2912f7"
  - "srcrev_f985bf951a757f61637465db129a7a29"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1774302868.190719

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca`
- `source_revision_id`: `srcrev_334d99d986cccde5924ba8b4e7a579ae`
- `chunk_id`: `srcchunk_2b476e716cafbcc35c376f28ec4414f1`
- `native_locator`: `slack:C0AKH5SNGKH:1774302868.190719:1774302868.190719`

can yall start to list out the Todos in the individual apps
• Scout (Jobs) <@U08HVGL6LDR> / <@U04L0DD6B6F>  (<@U0772SH7BRA> for relevant APIs)
• Aura  <@U05A515NBFC> / <@U067QP5PD6J> 
• DePIN Admin  <@U04L0DD6B6F> 
• DePIN Unfiied Backend <@U0772SH7BRA> / <@U08V4SFU7LZ> / <@U0AC11JV8AX>

### Citation 2

- `source_document_id`: `srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca`
- `source_revision_id`: `srcrev_a06ece44508438aa19d1a6c02f80709b`
- `chunk_id`: `srcchunk_7a91a5694b99bf593404dc4d251afa67`
- `native_locator`: `slack:C0AKH5SNGKH:1774302868.190719:1774306349.628329`

is there any functionalities you see that we currently don't have ? Once aligned I think we can build it out more but the MVP architecture is already mostly in place

### Citation 3

- `source_document_id`: `srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca`
- `source_revision_id`: `srcrev_f3fdc20ad4e6341310b1c5278a2912f7`
- `chunk_id`: `srcchunk_1233330afb2d9368e5ba8b0ce9c45ed3`
- `native_locator`: `slack:C0AKH5SNGKH:1774302868.190719:1774306367.525049`

the leftover work is mostly productionizing

### Citation 4

- `source_document_id`: `srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca`
- `source_revision_id`: `srcrev_f985bf951a757f61637465db129a7a29`
- `chunk_id`: `srcchunk_7357c5d3d202bb9a8bfacea532d51bd6`
- `native_locator`: `slack:C0AKH5SNGKH:1774302868.190719:1774306375.056339`

observability/loggings etc

### Citation 5

- `source_document_id`: `srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca`
- `source_revision_id`: `srcrev_34ed52a9fdbecf7a2123ecf569133ad4`
- `chunk_id`: `srcchunk_7cf6474db451158cef10c711423cb890`
- `native_locator`: `slack:C0AKH5SNGKH:1774302868.190719:1774306522.783359`

some highlevel stuff i'm thinking
• notification / email (resend)
• job search API with <@U08V4SFU7LZ>’s api?
    ◦ is this already in OpenAPI?

### Citation 6

- `source_document_id`: `srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca`
- `source_revision_id`: `srcrev_73f7117fc86a725017e0e5af3c6919f6`
- `chunk_id`: `srcchunk_09bbf4694263426df80cb6b0770fd9c0`
- `native_locator`: `slack:C0AKH5SNGKH:1774302868.190719:1774306865.561419`

I think this is a different service. <@U08V4SFU7LZ> u can hook this up so it writes to our DB, to make it queryable to our job helper backend instead

### Citation 7

- `source_document_id`: `srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca`
- `source_revision_id`: `srcrev_df53cee0315f82179d433ecbbc8563d4`
- `chunk_id`: `srcchunk_f867c81e9c6b0b68ae0f8d7b8fb712b5`
- `native_locator`: `slack:C0AKH5SNGKH:1774302868.190719:1774306891.029599`

the cron job can also live inside our job helper service too

### Citation 8

- `source_document_id`: `srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca`
- `source_revision_id`: `srcrev_7b1ffba1bdec26b81883d7a25b170239`
- `chunk_id`: `srcchunk_0782f2ac43b2bdbbeac1b1d5e37755bf`
- `native_locator`: `slack:C0AKH5SNGKH:1774302868.190719:1774306896.592059`

<@U04L0DD6B6F> I don't have an API endpoint yet, the job data is stored in the scraper github repo for now

### Citation 9

- `source_document_id`: `srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca`
- `source_revision_id`: `srcrev_caa45183e7c63e4a23da262ae7262c59`
- `chunk_id`: `srcchunk_f78accd9f84cb0e5811f899cb7c38130`
- `native_locator`: `slack:C0AKH5SNGKH:1774302868.190719:1774306965.088889`

Is there a way for me to pull job recommendations for a given user into the frontend? 

Or are we not doing per user yet? If not, maybe just general recommendations for a certain role?

### Citation 10

- `source_document_id`: `srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca`
- `source_revision_id`: `srcrev_a8d6b8b46dea4677e4f321c278dafe4c`
- `chunk_id`: `srcchunk_fd83871ac657089cb464fddde9e2e246`
- `native_locator`: `slack:C0AKH5SNGKH:1774302868.190719:1774307002.498029`

job match making recommendation is an algorithm problem, so we can work on that next if desired

### Citation 11

- `source_document_id`: `srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca`
- `source_revision_id`: `srcrev_3da9626a11c9d54efe130bac14aa8440`
- `chunk_id`: `srcchunk_bd2be77014675db2d5b4e15add30dfe5`
- `native_locator`: `slack:C0AKH5SNGKH:1774302868.190719:1774307009.696909`

otherwise it'll just be a job search

### Citation 12

- `source_document_id`: `srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca`
- `source_revision_id`: `srcrev_b98781cbbe49ec92662b663ae3a8f140`
- `chunk_id`: `srcchunk_006891978acf4bd87f391d396161750a`
- `native_locator`: `slack:C0AKH5SNGKH:1774302868.190719:1774307016.456459`

<@U08V4SFU7LZ> can you work with <@U0772SH7BRA> on integrating that this week? can we add to the todo to track?

### Citation 13

- `source_document_id`: `srcdoc_e4d9aa23eb0aba0432dca4fcb65dd3ca`
- `source_revision_id`: `srcrev_e50e9892b9f1f1767ab4c7f2fba7cd77`
- `chunk_id`: `srcchunk_54292f33cf9224e1de9d845020884b55`
- `native_locator`: `slack:C0AKH5SNGKH:1774302868.190719:1774307022.176019`

yes i think that'll be next

