---
title: "cdr-sdk GitHub Access Request (April 2026)"
type: "decision"
slug: "decisions/2026-04-13-cdr-sdk-github-access-request"
freshness: "2026-04-13T06:28:44Z"
tags:
  - "access-control"
  - "github"
  - "permissions"
  - "repo-transfer"
owners:
  - "piplabs"
source_revision_ids:
  - "srcrev_064c861d9ffc995677595b2a9b1c4c8e"
  - "srcrev_31d9a59cfd59fc5aa75d3fe2f19d5c32"
  - "srcrev_7297a2a6031cf2ed1b11c9e7f53eb48a"
  - "srcrev_84ac4509dd0281ad9365a53a2a0ca25b"
  - "srcrev_fb399532d760d40bee61e90badfffc79"
conflict_state: "none"
---

# cdr-sdk GitHub Access Request (April 2026)

## Summary

Slack discussion about transferring ownership and changing permissions for the cdr-sdk repo, resulting in a permission change for @U067QP5PD6J and the requester retracting the request due to misunderstanding.

## Claims

- A request was made to transfer ownership of the GitHub repository cdr-sdk (https://github.com/piplabs/cdr-sdk) to the infra team and downgrade user jdubpark's permissions from admin to write. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2c4f1a61a8fb1b806bf77b062b7da163` `source_revision_id=srcrev_7297a2a6031cf2ed1b11c9e7f53eb48a` `chunk_id=srcchunk_2fa38df5b3df8d7bd9b5e743b7b815c0` `native_locator=slack:C0547N89JUB:1776061486.872239:1776061486.872239` `source_timestamp=2026-04-13T06:24:46Z`
- The automated assistant indicated that repo transfer and org ownership changes require a human org admin and cannot be done via its API setup, but permission changes could be done via API if admin token scope is available. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2c4f1a61a8fb1b806bf77b062b7da163` `source_revision_id=srcrev_fb399532d760d40bee61e90badfffc79` `chunk_id=srcchunk_c029fb9ec073bbc3a514653dc1be08fe` `native_locator=slack:C0547N89JUB:1776061486.872239:1776061536.308919` `source_timestamp=2026-04-13T06:25:36Z`
- A user claimed to have changed the permission of user @U067QP5PD6J, and asked for clarification on what 'transfer ownership' means. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2c4f1a61a8fb1b806bf77b062b7da163` `source_revision_id=srcrev_31d9a59cfd59fc5aa75d3fe2f19d5c32` `chunk_id=srcchunk_48c9e485b525cb40bfa5685c4a773368` `native_locator=slack:C0547N89JUB:1776061486.872239:1776061677.679789` `source_timestamp=2026-04-13T06:27:57Z`
- The assistant clarified that transfer ownership means moving the repo to a different org or personal account, and noted that the permission change appeared to be handled, so no repo move was needed. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2c4f1a61a8fb1b806bf77b062b7da163` `source_revision_id=srcrev_84ac4509dd0281ad9365a53a2a0ca25b` `chunk_id=srcchunk_36e550b8df0f9c007140c08a77223573` `native_locator=slack:C0547N89JUB:1776061486.872239:1776061688.347989` `source_timestamp=2026-04-13T06:28:08Z`
- The original requester then stated they had misunderstood the situation and that everything was resolved. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_2c4f1a61a8fb1b806bf77b062b7da163` `source_revision_id=srcrev_064c861d9ffc995677595b2a9b1c4c8e` `chunk_id=srcchunk_16fb0e386edcfa74b214d4112f4677ff` `native_locator=slack:C0547N89JUB:1776061486.872239:1776061724.759559` `source_timestamp=2026-04-13T06:28:44Z`

## Open Questions

- Which user's permissions were actually changed? The thread indicates a change was made to @U067QP5PD6J, not jdubpark as originally requested, but the requester later said all is good, implying no issue.

## Sources

- `source_document_id`: `srcdoc_2c4f1a61a8fb1b806bf77b062b7da163`
- `source_revision_id`: `srcrev_064c861d9ffc995677595b2a9b1c4c8e`
