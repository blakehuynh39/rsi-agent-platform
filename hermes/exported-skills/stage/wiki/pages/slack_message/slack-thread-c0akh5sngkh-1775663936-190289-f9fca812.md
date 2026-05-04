---
title: "Slack thread C0AKH5SNGKH 1775663936.190289"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_9c2337c40a4a6bdf693177e4f9fca812"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1775663936.190289"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1775663936.190289"
source_revision_ids:
  - "srcrev_0d432c5559b85da2536f8c74b606975e"
  - "srcrev_2d94731e6c633b88624e4f5724533c0f"
  - "srcrev_3808dff888da67c1192f36ddc7fadc34"
  - "srcrev_551923ec46215adec3b11db6c9c5f64f"
  - "srcrev_6182d5f7be92bec78870b3682a731d86"
  - "srcrev_c6143fd221f86dc02cd5f9a1369ea5f7"
  - "srcrev_d767c7b92d257a7c1082abebccae3c90"
  - "srcrev_d92ddd707c50ee8de8dbe81ee988ea19"
  - "srcrev_fb02a6560464ea6af3524517999bae46"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1775663936.190289

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_9c2337c40a4a6bdf693177e4f9fca812`
- `source_revision_id`: `srcrev_c6143fd221f86dc02cd5f9a1369ea5f7`
- `chunk_id`: `srcchunk_67683b6ca7945c0c7edd25b1cc4e91ad`
- `native_locator`: `slack:C0AKH5SNGKH:1775663936.190289:1775663936.190289`

<@U05A515NBFC> I have time today to work on frontend task, what can I do to help while ur asleep

### Citation 2

- `source_document_id`: `srcdoc_9c2337c40a4a6bdf693177e4f9fca812`
- `source_revision_id`: `srcrev_3808dff888da67c1192f36ddc7fadc34`
- `chunk_id`: `srcchunk_90f9900763b72935e2760d0cdb8afb55`
- `native_locator`: `slack:C0AKH5SNGKH:1775663936.190289:1775665850.158069`

not sure if you've tested it with World app yet?

Otherwise, it kind of depends which endpoints will be ready next, could be adding the campaign scripts, i think the api request is commented out so it might work ootb

other outstanding things include the leaderboard and activity in the account and dashboard, you can sync with <@U0772SH7BRA> to see whats coming next

### Citation 3

- `source_document_id`: `srcdoc_9c2337c40a4a6bdf693177e4f9fca812`
- `source_revision_id`: `srcrev_d767c7b92d257a7c1082abebccae3c90`
- `chunk_id`: `srcchunk_633f62a7efb3cbc002ee877ef0900467`
- `native_locator`: `slack:C0AKH5SNGKH:1775663936.190289:1775665928.643479`

we've also had some branding and design guidance now, i've been working on the onboarding so i can do that part, but you could update the icons and text styles

<https://www.figma.com/design/x3DJKKDeQXAdks0lQCHFAN/Numo-Brand?node-id=1-942&amp;p=f&amp;m=dev>

### Citation 4

- `source_document_id`: `srcdoc_9c2337c40a4a6bdf693177e4f9fca812`
- `source_revision_id`: `srcrev_d92ddd707c50ee8de8dbe81ee988ea19`
- `chunk_id`: `srcchunk_e447b74fd4a15e9d73da4b615370a7fe`
- `native_locator`: `slack:C0AKH5SNGKH:1775663936.190289:1775666053.004789`

Slack file attachment:
- Numo – Brand fonts.zip (application/zip, zip, 1654445 bytes) https://storyprotocol.slack.com/files/U0ASDQKU3UL/F0ARKMERPPY/numo_____brand_fonts.zip

### Citation 5

- `source_document_id`: `srcdoc_9c2337c40a4a6bdf693177e4f9fca812`
- `source_revision_id`: `srcrev_0d432c5559b85da2536f8c74b606975e`
- `chunk_id`: `srcchunk_fcfddacada251f5db814e2728586df60`
- `native_locator`: `slack:C0AKH5SNGKH:1775663936.190289:1775674915.840469`

<@U08HVGL6LDR> can you look into the issue for triggering <https://staging-depin.storyprotocol.net/v1/me/complete-intro> after the onboarding, try sending a smaller image file to see if it will work for staging, right now the <http://staging-depin.storyprotocol.net|staging-depin.storyprotocol.net> nginx ingress needs `proxy-body-size` increased. looks like we're uploading ~5MB avatar images to `/v1/me/complete-intro` and nginx is truncating the body

### Citation 6

- `source_document_id`: `srcdoc_9c2337c40a4a6bdf693177e4f9fca812`
- `source_revision_id`: `srcrev_2d94731e6c633b88624e4f5724533c0f`
- `chunk_id`: `srcchunk_a4eb3a0e665006f2e3da18131f2ac2c6`
- `native_locator`: `slack:C0AKH5SNGKH:1775663936.190289:1775680402.510549`

<@U083MMT1771> I just tried this branch that sam was working on recently and the /complete-intro step worked: <https://numo-monorepo-24k03ytns-story-protocol.vercel.app/|numo-monorepo-24k03ytns-story-protocol.vercel.app>

### Citation 7

- `source_document_id`: `srcdoc_9c2337c40a4a6bdf693177e4f9fca812`
- `source_revision_id`: `srcrev_6182d5f7be92bec78870b3682a731d86`
- `chunk_id`: `srcchunk_0f99844598515e383bbf7d11f2945fd7`
- `native_locator`: `slack:C0AKH5SNGKH:1775663936.190289:1775680417.407819`

The staging is using mock API I believe

### Citation 8

- `source_document_id`: `srcdoc_9c2337c40a4a6bdf693177e4f9fca812`
- `source_revision_id`: `srcrev_551923ec46215adec3b11db6c9c5f64f`
- `chunk_id`: `srcchunk_82f4686a693b3fcb68d7c3a2e17791cf`
- `native_locator`: `slack:C0AKH5SNGKH:1775663936.190289:1775680474.902489`

Actually ignore that vercel link i sent because thats using mock api too

### Citation 9

- `source_document_id`: `srcdoc_9c2337c40a4a6bdf693177e4f9fca812`
- `source_revision_id`: `srcrev_fb02a6560464ea6af3524517999bae46`
- `chunk_id`: `srcchunk_525bd7caafefa0ab1847cca512947514`
- `native_locator`: `slack:C0AKH5SNGKH:1775663936.190289:1775680506.355539`

yeah looks like staging is calling the actual <https://staging-depin.storyprotocol.net/v1/me/complete-intro>

