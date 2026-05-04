---
title: "Slack thread C0AKH5SNGKH 1776802794.336749"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_20e42bebdbed024979a94bba34b9cfc7"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776802794.336749"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776802794.336749"
source_revision_ids:
  - "srcrev_0d920efaba0b61129bf96bc1953acb80"
  - "srcrev_2293b33e62f760dbee9ca81a45d376ee"
  - "srcrev_2ad34d46b245188e33cfcd489b43ba23"
  - "srcrev_440787f950e6c5946068bb03853dafe6"
  - "srcrev_4bf3ea4a4c57026ef572184aa744c67d"
  - "srcrev_5e3238ba58fad54552991e86eb6b4b80"
  - "srcrev_861d571e749e261a552cf983d58fd809"
  - "srcrev_c2e194259759496d23c8c4b3642b4e41"
  - "srcrev_f8efbd57d2e1a08cf7a7a933d835ed71"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1776802794.336749

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_20e42bebdbed024979a94bba34b9cfc7`
- `source_revision_id`: `srcrev_c2e194259759496d23c8c4b3642b4e41`
- `chunk_id`: `srcchunk_b99ccd630a772cea9da46ad25dc0c1c7`
- `native_locator`: `slack:C0AKH5SNGKH:1776802794.336749:1776802794.336749`

<@U0772SH7BRA> the upload is stuck

### Citation 2

- `source_document_id`: `srcdoc_20e42bebdbed024979a94bba34b9cfc7`
- `source_revision_id`: `srcrev_4bf3ea4a4c57026ef572184aa744c67d`
- `chunk_id`: `srcchunk_082de1eeaf73c5f6007d747d00115e4e`
- `native_locator`: `slack:C0AKH5SNGKH:1776802794.336749:1776802843.491579`

what's the error

### Citation 3

- `source_document_id`: `srcdoc_20e42bebdbed024979a94bba34b9cfc7`
- `source_revision_id`: `srcrev_0d920efaba0b61129bf96bc1953acb80`
- `chunk_id`: `srcchunk_c50b64f639b25d98cd52ab5872205d50`
- `native_locator`: `slack:C0AKH5SNGKH:1776802794.336749:1776802858.629719`

look at response tab

### Citation 4

- `source_document_id`: `srcdoc_20e42bebdbed024979a94bba34b9cfc7`
- `source_revision_id`: `srcrev_861d571e749e261a552cf983d58fd809`
- `chunk_id`: `srcchunk_5e418d34df34b9bc1f0f854b8baa8384`
- `native_locator`: `slack:C0AKH5SNGKH:1776802794.336749:1776802953.504599`

no response there, looks like a CORS error

### Citation 5

- `source_document_id`: `srcdoc_20e42bebdbed024979a94bba34b9cfc7`
- `source_revision_id`: `srcrev_440787f950e6c5946068bb03853dafe6`
- `chunk_id`: `srcchunk_8812cdcb1c32c30aaf1319547c215ac9`
- `native_locator`: `slack:C0AKH5SNGKH:1776802794.336749:1776803125.953289`

can you paste the whole curl request

### Citation 6

- `source_document_id`: `srcdoc_20e42bebdbed024979a94bba34b9cfc7`
- `source_revision_id`: `srcrev_2ad34d46b245188e33cfcd489b43ba23`
- `chunk_id`: `srcchunk_68ecd5c6532978de595636e91a300a10`
- `native_locator`: `slack:C0AKH5SNGKH:1776802794.336749:1776803243.467059`

```curl '<https://staging-depin.storyprotocol.net/v1/submissions/initiate-upload>' \
  -H 'sec-ch-ua-platform: "macOS"' \
  -H 'Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiI3NzgyZjQ2ZS00ZTQyLTQ1ODEtOGI0MS04OWM2ZTAzYmNjM2QiLCJyb2xlIjoidXNlciIsInVzZXJfaWQiOiI3NzgyZjQ2ZS00ZTQyLTQ1ODEtOGI0MS04OWM2ZTAzYmNjM2QiLCJkeW5hbWljX3VzZXJfaWQiOiI5ZTgwN2U2Yi0zZjU0LTRkYTQtYmNhMS0zNWVmMzNjY2I0OTQiLCJlbWFpbCI6InJveWNlemhhb2NhQGdtYWlsLmNvbSIsIndhbGxldF9hZGRyZXNzIjoiMHhENThjMDM4YTgwNTc5NkQyRTZmMzZlREE0NkE2MjIyOTc4MDFkQThlIiwiYXVkIjoiZGVwaW4tYXBpIiwiaWF0IjoxNzc2NzA3MTM3LCJleHAiOjE3NzkyOTkxMzcsImlzcyI6ImRlcGluLWJhY2tlbmQifQ.5rwZQHBoJeA05K9Ert2SQMjYC4EXubWtb7fThaIsTXg' \
  -H 'Referer: <https://numo-staging.vercel.app/>' \
  -H 'sec-ch-ua: "Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"' \
  -H 'CF-Turnstile-Token: XXXX.DUMMY.TOKEN.XXXX' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36' \
  -H 'Content-Type: application/json' \
  --data-raw '{"campaign_id":"5b9c8ea9-0021-4cd0-bc8b-ae0a07566508","file_name":"submission_1776802613086.webm","content_type":"audio/webm","declared_size_bytes":199626,"script_assignment_id":"8ad9483c-9edf-4d5c-8bf4-fe1f9f8bc8d0"}'```

### Citation 7

- `source_document_id`: `srcdoc_20e42bebdbed024979a94bba34b9cfc7`
- `source_revision_id`: `srcrev_f8efbd57d2e1a08cf7a7a933d835ed71`
- `chunk_id`: `srcchunk_e73fc155c99cc510a192ce64a72293bb`
- `native_locator`: `slack:C0AKH5SNGKH:1776802794.336749:1776803923.413759`

can u try again

### Citation 8

- `source_document_id`: `srcdoc_20e42bebdbed024979a94bba34b9cfc7`
- `source_revision_id`: `srcrev_5e3238ba58fad54552991e86eb6b4b80`
- `chunk_id`: `srcchunk_4539ff4f052f3427788fe33aa8f997e9`
- `native_locator`: `slack:C0AKH5SNGKH:1776802794.336749:1776803938.565349`

it's because of newly added turnstile token

### Citation 9

- `source_document_id`: `srcdoc_20e42bebdbed024979a94bba34b9cfc7`
- `source_revision_id`: `srcrev_2293b33e62f760dbee9ca81a45d376ee`
- `chunk_id`: `srcchunk_848f51c683d3c2740337bc27a8cfd4d4`
- `native_locator`: `slack:C0AKH5SNGKH:1776802794.336749:1776804062.220709`

working fine now

