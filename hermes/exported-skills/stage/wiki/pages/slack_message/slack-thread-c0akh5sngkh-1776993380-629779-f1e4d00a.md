---
title: "Slack thread C0AKH5SNGKH 1776993380.629779"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_72c2b6cca610dd918caca22df1e4d00a"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776993380.629779"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776993380.629779"
source_revision_ids:
  - "srcrev_0d3ce89fa4f3c9930f0a6fb487c2d0fd"
  - "srcrev_1fd188569c971135f78d82a5ba4f8012"
  - "srcrev_37f7e3c03017c6a2a007009ce5ee1499"
  - "srcrev_3cfc6850527fde135027d5cb54237745"
  - "srcrev_3ee20b2a88638c786c856b91510e4a83"
  - "srcrev_580a0aecf97af9f2b2a98ebd7403dcba"
  - "srcrev_7fa69e662eba449379fd6c8302b73264"
  - "srcrev_a25be4a26d311ab7e4841f3e2dde3e4a"
  - "srcrev_b508b723c9373e18addacca176f6aa81"
  - "srcrev_e012c5e735486d947a4440ad847af1cf"
  - "srcrev_e8e6e377747f59cc5305b6a780ad95e5"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1776993380.629779

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_72c2b6cca610dd918caca22df1e4d00a`
- `source_revision_id`: `srcrev_e012c5e735486d947a4440ad847af1cf`
- `chunk_id`: `srcchunk_c4e83579fbb4b60623f6f3c2d394f5d6`
- `native_locator`: `slack:C0AKH5SNGKH:1776993380.629779:1776993380.629779`

hi <@U08332YRB7W> can you help to reach out to the castle team for them to enable search for us
```We're on the Pro plan but GET <https://api.castle.io/v1/events>
returns 503:
  service_unavailable
  "Castle Search API is not enabled for this account;
   cannot fetch live event history"
Please enable the Search / Events API for our Production environment.```

### Citation 2

- `source_document_id`: `srcdoc_72c2b6cca610dd918caca22df1e4d00a`
- `source_revision_id`: `srcrev_a25be4a26d311ab7e4841f3e2dde3e4a`
- `chunk_id`: `srcchunk_328a98433cbe46358d18a478c442e8a0`
- `native_locator`: `slack:C0AKH5SNGKH:1776993380.629779:1776993525.416129`

:white_check_mark: Task created: <https://github.com/piplabs/lion-team-sync/issues/627|#627> — tracking this request.

### Citation 3

- `source_document_id`: `srcdoc_72c2b6cca610dd918caca22df1e4d00a`
- `source_revision_id`: `srcrev_1fd188569c971135f78d82a5ba4f8012`
- `chunk_id`: `srcchunk_c4a202d3b50c0cca09248b1827a54133`
- `native_locator`: `slack:C0AKH5SNGKH:1776993380.629779:1776995411.185909`

I've reached out to them via support chat but all I see is this. I'll keep an eye and let you know once I get a response.

### Citation 4

- `source_document_id`: `srcdoc_72c2b6cca610dd918caca22df1e4d00a`
- `source_revision_id`: `srcrev_580a0aecf97af9f2b2a98ebd7403dcba`
- `chunk_id`: `srcchunk_7c02dfbc1781a232e078e43a1397e31a`
- `native_locator`: `slack:C0AKH5SNGKH:1776993380.629779:1776995452.047669`

sounds good ty! i think they're in EU so lets see when they can get back

### Citation 5

- `source_document_id`: `srcdoc_72c2b6cca610dd918caca22df1e4d00a`
- `source_revision_id`: `srcrev_3ee20b2a88638c786c856b91510e4a83`
- `chunk_id`: `srcchunk_ec9d1c0d741e3277371e98459347da24`
- `native_locator`: `slack:C0AKH5SNGKH:1776993380.629779:1777141169.681239`

<@U083MMT1771> can you share a screenshot of the search feature where you are trying to access it? The support is asking for it

### Citation 6

- `source_document_id`: `srcdoc_72c2b6cca610dd918caca22df1e4d00a`
- `source_revision_id`: `srcrev_3cfc6850527fde135027d5cb54237745`
- `chunk_id`: `srcchunk_e0156cea35b941294d3edb529268c3eb`
- `native_locator`: `slack:C0AKH5SNGKH:1776993380.629779:1777141656.156709`

let them know thay we have integrated with their Risk API for a handful of events ($login, $referral_claim, $submission_initiate, etc.). we persist the verdict + composite risk score in our own DB so our admin dashboard can list a user's history but we want to enable the Events API (POST /v1/events/query, POST /v1/events/group, GET /v1/events/schema), (same as what they use on our castle dashboard(castle owned) The Search API just exposes that capability programmatically so we can build our own UI on top of it.

### Citation 7

- `source_document_id`: `srcdoc_72c2b6cca610dd918caca22df1e4d00a`
- `source_revision_id`: `srcrev_37f7e3c03017c6a2a007009ce5ee1499`
- `chunk_id`: `srcchunk_bc1469ae69b71dab6aba45869b6384db`
- `native_locator`: `slack:C0AKH5SNGKH:1776993380.629779:1777393382.813739`

<@U083MMT1771> got the response.

```Hello Vinod,

Sorry for the late response. The Events API is available on our Enterprise plan, not on our Pro plan. What are you trying to achieve? Happy to think about  the best approach.```

### Citation 8

- `source_document_id`: `srcdoc_72c2b6cca610dd918caca22df1e4d00a`
- `source_revision_id`: `srcrev_7fa69e662eba449379fd6c8302b73264`
- `chunk_id`: `srcchunk_e07b5ad56af0eb00e6f0cdefa2feb036`
- `native_locator`: `slack:C0AKH5SNGKH:1776993380.629779:1777393398.975229`

:sweat_smile:

### Citation 9

- `source_document_id`: `srcdoc_72c2b6cca610dd918caca22df1e4d00a`
- `source_revision_id`: `srcrev_0d3ce89fa4f3c9930f0a6fb487c2d0fd`
- `chunk_id`: `srcchunk_5faface30a88e15ad5dbbb8667eb2a1d`
- `native_locator`: `slack:C0AKH5SNGKH:1776993380.629779:1777393415.187259`

its ok we logged it on the backend for now

### Citation 10

- `source_document_id`: `srcdoc_72c2b6cca610dd918caca22df1e4d00a`
- `source_revision_id`: `srcrev_b508b723c9373e18addacca176f6aa81`
- `chunk_id`: `srcchunk_a0b120cfa13ea9edb88161d6feeba87e`
- `native_locator`: `slack:C0AKH5SNGKH:1776993380.629779:1777393424.188759`

or should we get the enterprise plan?

### Citation 11

- `source_document_id`: `srcdoc_72c2b6cca610dd918caca22df1e4d00a`
- `source_revision_id`: `srcrev_e8e6e377747f59cc5305b6a780ad95e5`
- `chunk_id`: `srcchunk_f169b3b7d503958569cb7753f5038eb3`
- `native_locator`: `slack:C0AKH5SNGKH:1776993380.629779:1777394375.896509`

Not sure how much it costs. For enterprise we might need to talk to their sales.

