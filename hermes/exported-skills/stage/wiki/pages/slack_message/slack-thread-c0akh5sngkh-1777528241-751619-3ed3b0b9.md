---
title: "Slack thread C0AKH5SNGKH 1777528241.751619"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_0ec0c6bb5d177e1b88e498253ed3b0b9"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1777528241.751619"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1777528241.751619"
source_revision_ids:
  - "srcrev_352e345df2ca3f50cd5c22329c6758b5"
  - "srcrev_f6c41256ef2d65fd4616550f25f4791d"
  - "srcrev_fe16e38cb93624d4e2acb5d786754dc0"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1777528241.751619

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_0ec0c6bb5d177e1b88e498253ed3b0b9`
- `source_revision_id`: `srcrev_fe16e38cb93624d4e2acb5d786754dc0`
- `chunk_id`: `srcchunk_78db1155171a579e8c5d0b4c6c13f74f`
- `native_locator`: `slack:C0AKH5SNGKH:1777528241.751619:1777528241.751619`

There's a gap between the number of submissions and the number of attempted IP registrations. I see 2 potential causes so far:

1. Metrics calculation issue: 
    a. if users start submissions but don't reach upload finalization, IP registration is not initiated. Counting such submissions in the total could cause the gap
    b. attempted IP registrations count is a bit late 
2. Registration worker is late to pick up submissions. I doubt this is the case since the same worker handled much bigger loads during tests, but who knows. Let's see if tomorrow the gap closes when it's night in India and less submissions come

### Citation 2

- `source_document_id`: `srcdoc_0ec0c6bb5d177e1b88e498253ed3b0b9`
- `source_revision_id`: `srcrev_352e345df2ca3f50cd5c22329c6758b5`
- `chunk_id`: `srcchunk_06c1d22ed85c1bf447b287911197a235`
- `native_locator`: `slack:C0AKH5SNGKH:1777528241.751619:1777530911.185339`

good call -

1. we shouldn't store any incomplete submissions. as in it only gets saved to the db when its finished uploaded. i hope that is the case now
2. how can we figure this out, ideally it should be simple:
# of submissions = successful IP registrations + failed IP registrations

### Citation 3

- `source_document_id`: `srcdoc_0ec0c6bb5d177e1b88e498253ed3b0b9`
- `source_revision_id`: `srcrev_f6c41256ef2d65fd4616550f25f4791d`
- `chunk_id`: `srcchunk_4f8deaaf544984597270216a9db91f48`
- `native_locator`: `slack:C0AKH5SNGKH:1777528241.751619:1777569512.686799`

Agree those numbers must be equal

