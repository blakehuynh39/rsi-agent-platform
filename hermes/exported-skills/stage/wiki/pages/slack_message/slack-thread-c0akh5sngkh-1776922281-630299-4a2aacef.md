---
title: "Slack thread C0AKH5SNGKH 1776922281.630299"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_d18683739ae0c8ef7546c1b14a2aacef"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776922281.630299"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776922281.630299"
source_revision_ids:
  - "srcrev_50d84364c83a8b8213fee64d961ffb09"
  - "srcrev_6835ae6be703228cfc6eda49f90b852a"
  - "srcrev_6e277f6c4eb6ecb35666b9c771f47847"
  - "srcrev_728bde639f94e1ded664a88b01a99bd2"
  - "srcrev_aa4b209e0b1fe0e0f8e4afdf436383c9"
  - "srcrev_e44401b6003cb50b346abbf9057a56fc"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1776922281.630299

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_d18683739ae0c8ef7546c1b14a2aacef`
- `source_revision_id`: `srcrev_6e277f6c4eb6ecb35666b9c771f47847`
- `chunk_id`: `srcchunk_ee81605f11699b70d57f43b22102f713`
- `native_locator`: `slack:C0AKH5SNGKH:1776922281.630299:1776922281.630299`

<@U08951K4SRY> did we verify that the file has changed from .webm to .wav?

### Citation 2

- `source_document_id`: `srcdoc_d18683739ae0c8ef7546c1b14a2aacef`
- `source_revision_id`: `srcrev_e44401b6003cb50b346abbf9057a56fc`
- `chunk_id`: `srcchunk_7d5490ac4cf46289a17b39ff6ebd4332`
- `native_locator`: `slack:C0AKH5SNGKH:1776922281.630299:1776922629.705349`

yes, I checked it is .wav file now. but I’ll need to check on mobile as well, I remember it used to be in .mp4 format sometime

### Citation 3

- `source_document_id`: `srcdoc_d18683739ae0c8ef7546c1b14a2aacef`
- `source_revision_id`: `srcrev_728bde639f94e1ded664a88b01a99bd2`
- `chunk_id`: `srcchunk_64bb393782db3db3b3f9e71c9401bffc`
- `native_locator`: `slack:C0AKH5SNGKH:1776922281.630299:1776922656.923879`

okay thanks, worth it to validate that sound is actually coming through too :slightly_smiling_face:

### Citation 4

- `source_document_id`: `srcdoc_d18683739ae0c8ef7546c1b14a2aacef`
- `source_revision_id`: `srcrev_6835ae6be703228cfc6eda49f90b852a`
- `chunk_id`: `srcchunk_2ee06f669a97e6094e685f4eca80ce2f`
- `native_locator`: `slack:C0AKH5SNGKH:1776922281.630299:1776923524.618439`

btw <@U08951K4SRY> for load testing submissions - do we just call it locally using jwt? wondering if users plan on doing this, will they be able to? (I realized looking at the admin dashboard that if they are allowed to do this we won't have any castle info on them at all)

<@U0772SH7BRA> what's the rate limit for each user jwt?

### Citation 5

- `source_document_id`: `srcdoc_d18683739ae0c8ef7546c1b14a2aacef`
- `source_revision_id`: `srcrev_aa4b209e0b1fe0e0f8e4afdf436383c9`
- `chunk_id`: `srcchunk_83972b5fe5be88f098a01cddbeb3550a`
- `native_locator`: `slack:C0AKH5SNGKH:1776922281.630299:1776923966.595129`

yeah, we just call locally or on GHA, using jwt.   i think when we enable turnstile or castle, it can be detected?
here's what i tested with sasi in psdn s2
```it should block non-user requests to backend.
like requests from headless browsers like puppeteer, selenium automations
scripts to call api
anything without user click through the browser```

### Citation 6

- `source_document_id`: `srcdoc_d18683739ae0c8ef7546c1b14a2aacef`
- `source_revision_id`: `srcrev_50d84364c83a8b8213fee64d961ffb09`
- `chunk_id`: `srcchunk_7232ffd396d1511f968157234974d46f`
- `native_locator`: `slack:C0AKH5SNGKH:1776922281.630299:1776924313.129229`

<@U04L0DD6B6F> for .wav format,  just checked on web, mobile web(ios, android), miniapp(ios, android),  all good and audios can by played

