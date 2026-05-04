---
title: "Slack thread C0AKH5SNGKH 1775798436.158669"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1775798436.158669"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1775798436.158669"
source_revision_ids:
  - "srcrev_0a41657324977dd7fb08641da0a5a3bd"
  - "srcrev_0cfc37999bbd40a56fb113d26c76caa9"
  - "srcrev_1f73b81c6dfe3e4bde744e8656f040c4"
  - "srcrev_35fffef552fd95655444055a0401f5bd"
  - "srcrev_375c17ca293af66b9b81ae603b2a920e"
  - "srcrev_3d0021da7ffeff337ff13666c40e8b8c"
  - "srcrev_3de401c74cb496e110d605fdfeb0bdb0"
  - "srcrev_45e5fdce76a93e9c63c23cccc4cd2c94"
  - "srcrev_604d2a76b164c12ef4a22b878b0c8cde"
  - "srcrev_64f771b5f3417d2a7487a7002ebed0ec"
  - "srcrev_6b21ca2e4286087c85d07983e0b0ff9e"
  - "srcrev_850b49d1cba498298bd85c6e1b8bd56b"
  - "srcrev_8e378855fc7dad830a7aa90e98a86727"
  - "srcrev_a07a4a433e0b5bcfcbcc597c51afc807"
  - "srcrev_a5f3b5bf5be86b497c38f93818277432"
  - "srcrev_c468dbab1a8237763322ba522a5e834c"
  - "srcrev_d93dfc4bee6f41f025e5827e85d0d174"
  - "srcrev_d97aa54920cc140b58728c6574a7432f"
  - "srcrev_da74749fbe7a3fb4068a2e35fe015483"
  - "srcrev_df70db065b6183255637cd8b9990a9f5"
  - "srcrev_eb7f5d434098911282178679eadcde10"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1775798436.158669

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_6b21ca2e4286087c85d07983e0b0ff9e`
- `chunk_id`: `srcchunk_f82004269e4bcc430d5428dadfa58ec0`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775798436.158669`

Ok <@U0772SH7BRA> I think there is an error on the `/verify` route:

AI says this is the error:
```I checked the frontend, World docs, and our staging OpenAPI.

Frontend is calling POST /v1/world/verify and sending the raw IDKit completion.result unchanged.

World docs say the IDKit payload must be forwarded to POST <https://developer.world.org/api/v4/verify/{rp_id}> as-is, with no field remapping.

Our own OpenAPI for /v1/world/verify says the same:
"Raw IDKit result payload forwarded to World v4 verify without field remapping."

Current 422:
"merkle_root must contain exactly 32 bytes"

That suggests the /v1/world/verify implementation is still parsing/remapping legacy proof fields or coercing merkle_root/nullifier/proof instead of forwarding the raw IDKit JSON body unchanged.```

### Citation 2

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_df70db065b6183255637cd8b9990a9f5`
- `chunk_id`: `srcchunk_ef67c893b6325e5364303a4d4678313c`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775841948.406999`

<@U0772SH7BRA> were u able to take a look at this? i know its still early for u

### Citation 3

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_a5f3b5bf5be86b497c38f93818277432`
- `chunk_id`: `srcchunk_86a619737fdb050b372e4b5af2ee5e9d`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775843037.765069`

can u try again

### Citation 4

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_64f771b5f3417d2a7487a7002ebed0ec`
- `chunk_id`: `srcchunk_8022d723165031e5b589975ca73b1b45`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775843077.853219`

yep

### Citation 5

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_604d2a76b164c12ef4a22b878b0c8cde`
- `chunk_id`: `srcchunk_28fab9e4f49dd3dd34238ae51acf3b13`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775843814.628669`

im getting a different error now

### Citation 6

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_45e5fdce76a93e9c63c23cccc4cd2c94`
- `chunk_id`: `srcchunk_d8f25b78fcc4a98b929e28e97b34627a`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775843847.075469`

AI said this:
```Frontend is still getting through World login and the World verification UI successfully (green check appears).

But now the call to POST /v1/world/verify is failing as a network/load error in the browser, not as a normal JSON API error.

Frontend console now shows:
[World] Verification error: "Load Failed"

This means fetch() is rejecting before the frontend receives a usable HTTP response.

So this is different from the earlier 422 "merkle_root must contain exactly 32 bytes".

Please check /v1/world/verify for this request and confirm:
1. whether the request is reaching the backend at all
2. whether the backend is crashing/panicking before sending a response
3. whether CORS/preflight is allowed for this route
   - Origin: <https://ouch-film-discover.ngrok-free.dev>
   - Method: POST
   - Headers: Authorization, Content-Type
4. whether error responses on this route also include CORS headers
5. whether a proxy/gateway is closing the connection before the response is sent

From the frontend side, this is happening during fetch() itself, before response.ok handling.```

### Citation 7

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_35fffef552fd95655444055a0401f5bd`
- `chunk_id`: `srcchunk_8b9d349d266969922a2133a46c3fe2f9`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775843865.486999`

I'm just getting a "Load failed" message when it calls /verify

### Citation 8

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_3d0021da7ffeff337ff13666c40e8b8c`
- `chunk_id`: `srcchunk_b40f03450ae508a9635607827941c387`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775843868.081069`

can u give me both curl requests that u did to login then verify

### Citation 9

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_0a41657324977dd7fb08641da0a5a3bd`
- `chunk_id`: `srcchunk_50d72fd77c0664d4bea0d7b61c7ad1ad`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775843871.674509`

yup

### Citation 10

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_8e378855fc7dad830a7aa90e98a86727`
- `chunk_id`: `srcchunk_61b5ca2d2ed1f4a25f27b3f3b1a6601e`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775844264.867299`

4 steps:

Nonce
```curl '<https://staging-depin.storyprotocol.net/v1/auth/world/nonce>' \
  -X 'POST' \
  -H 'accept: */*' \
  -H 'accept-language: en-US,en;q=0.9' \
  -H 'content-type: application/json' \
  -H 'origin: <https://ouch-film-discover.ngrok-free.dev>' \
  -H 'priority: u=1, i' \
  -H 'referer: <https://ouch-film-discover.ngrok-free.dev/login>' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: cross-site' \
  -H 'user-agent: Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148' \
  -H 'content-length: 0'```
Exchange
```curl '<https://staging-depin.storyprotocol.net/v1/auth/world/exchange>' \
  -X 'POST' \
  -H 'accept: */*' \
  -H 'accept-language: en-US,en;q=0.9' \
  -H 'content-type: application/json' \
  -H 'origin: <https://ouch-film-discover.ngrok-free.dev>' \
  -H 'priority: u=1, i' \
  -H 'referer: <https://ouch-film-discover.ngrok-free.dev/login>' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: cross-site' \
  -H 'user-agent: Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148' \
  --data-raw '{
  "nonce": "2VBcstuy424D8qVCF",
  "payload": {
    "address": "0x1a71781b5a8b9cbc183f741b10beccb4d4c1de77",
    "message": "<https://ouch-film-discover.ngrok-free.dev> wants you to sign in with your Ethereum account:\n0x1a71781B5a8B9cBC183f741B10beCCB4d4c1De77\n\n\nURI: <https://ouch-film-discover.ngrok-free.dev/login>\nVersion: 1\nChain ID: 480\nNonce: 2VBcstuy424D8qVCF\nIssued At: 2026-04-10T18:02:36.827Z\nExpiration Time: 2026-04-17T18:02:36.826Z\nNot Before: 2026-04-09T18:02:36.826Z",
    "signature": "0x3afc239b44d4be9b7e1e86055712cd7d2e15670d65508fbdd2cf034f81021a094f4b4ced42150aef820725faad5093c795126367fbc4c0edfc3c531611c141b31b"
  }
}'```
Signature
```curl '<https://staging-depin.storyprotocol.net/v1/world/signature>' \
  -X 'POST' \
  -H 'accept: */*' \
  -H 'accept-language: en-US,en;q=0.9' \
  -H 'content-type: application/json' \
  -H 'origin: <https://ouch-film-discover.ngrok-free.dev>' \
  -H 'priority: u=1, i' \
  -H 'referer: <https://ouch-film-discover.ngrok-free.dev/login>' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: cross-site' \
  -H 'user-agent: Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148' \
  -H 'authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiI3NzZlOWIyYS1iNTg1LTQ0MjEtODViNC0zN2E4YjkwM2U0ZGIiLCJyb2xlIjoidXNlciIsInVzZXJfaWQiOiI3NzZlOWIyYS1iNTg1LTQ0MjEtODViNC0zN2E4YjkwM2U0ZGIiLCJkeW5hbWljX3VzZXJfaWQiOm51bGwsImVtYWlsIjpudWxsLCJ3YWxsZXRfYWRkcmVzcyI6IjB4MWE3MTc4MUI1YThCOWNCQzE4M2Y3NDFCMTBiZUNDQjRkNGMxRGU3NyIsImF1ZCI6ImRlcGluLWFwaSIsImlhdCI6MTc3NTg0NDE1NywiZXhwIjoxNzc1ODQ3NzU3LCJpc3MiOiJkZXBpbi1iYWNrZW5kIn0.6lV21aS5WcRYsB36nLAsoR5fvPZ1RxldZgzWvsWnrRI' \
  --data-raw '{
  "action": "verify-human"
}'```
Verify
```curl '<https://staging-depin.storyprotocol.net/v1/world/verify>' \
  -X 'POST' \
  -H 'accept: */*' \
  -H 'accept-language: en-US,en;q=0.9' \
  -H 'content-type: application/json' \
  -H 'origin: <https://ouch-film-discover.ngrok-free.dev>' \
  -H 'priority: u=1, i' \
  -H 'referer: <https://ouch-film-discover.ngrok-free.dev/login>' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: cross-site' \
  -H 'user-agent: Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148' \
  -H 'authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiI3NzZlOWIyYS1iNTg1LTQ0MjEtODViNC0zN2E4YjkwM2U0ZGIiLCJyb2xlIjoidXNlciIsInVzZXJfaWQiOiI3NzZlOWIyYS1iNTg1LTQ0MjEtODViNC0zN2E4YjkwM2U0ZGIiLCJkeW5hbWljX3VzZXJfaWQiOm51bGwsImVtYWlsIjpudWxsLCJ3YWxsZXRfYWRkcmVzcyI6IjB4MWE3MTc4MUI1YThCOWNCQzE4M2Y3NDFCMTBiZUNDQjRkNGMxRGU3NyIsImF1ZCI6ImRlcGluLWFwaSIsImlhdCI6MTc3NTg0NDE1NywiZXhwIjoxNzc1ODQ3NzU3LCJpc3MiOiJkZXBpbi1iYWNrZW5kIn0.6lV21aS5WcRYsB36nLAsoR5fvPZ1RxldZgzWvsWnrRI' \
  --data-raw '{
  "protocol_version": "3.0",
  "nonce": "0x00457c58c087b50ddd724052ae2ce739af492a3537b60d7a8be6076b4ba293c6",
  "action": "verify-human",
  "responses": [
    {
      "identifier": "device",
      "signal_hash": "0x00fd94eabf2e84d5a3c49a5b3bce3de869b8fea2ab04c7743e4800527d69e044",
      "proof": "0x0f83b064b45d60140036da5084121a32a3408103fbe790c0c72bf594f5dcb4dd26120fa5531ad7596487bd18a8679b0f5ef9dc01a5ffe1c79505352e2fe8f84b0a09e6c43ac61372d4a6880608a3fa7e902a2ba3e3e3fc4405b5b210558f6b622afc1e27aec3b8c8e9b7b1a1cf857656aa56036747d9d0dd12227936224f3593182b2ecfd06e9db2e8931f82c9cee98da607fa4b3f51937678a67ea21b9578522cf4a5c023a2c864cd8488f0b0b3647eed3e95285ca6ca95482c2dd7db4a37c50951ebfb96c2f6c9fa9b506fe9af2f034b4cc92a5ea40516baa07ed1aa80d2340e32f8624e459b109f9e4a253fda7134e669287c41e8c6ceacafe4a46a4f6ef2",
      "merkle_root": "0x955b8451e16ad29a6c77c1a41d683bf437e05fd735b134525237eb59f84717c",
      "nullifier": "0x07e99cff8ad789c4a08f8e6189dea0b6b69079b17e70f98b1e8acfb6c1f38f3a"
    }
  ],
  "environment": "production"
}'```

### Citation 11

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_375c17ca293af66b9b81ae603b2a920e`
- `chunk_id`: `srcchunk_dbbce2dc192cdd8e75a49694ba875567`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775846255.322469`

try again

### Citation 12

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_1f73b81c6dfe3e4bde744e8656f040c4`
- `chunk_id`: `srcchunk_702ab7ebc5fdc0dbe7a518c6a8eb1795`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775846560.551299`

beast <@U0772SH7BRA> it seems to be working. I log in and then its skipping verify step, maybe because I'm already verified

### Citation 13

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_eb7f5d434098911282178679eadcde10`
- `chunk_id`: `srcchunk_04921f42e58ecc85603e3333efe300ab`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775846590.519939`

i wonder if theres a way to invalidate user to test

### Citation 14

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_850b49d1cba498298bd85c6e1b8bd56b`
- `chunk_id`: `srcchunk_2c764b6d53f2450a16693fe465aaf286`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775846596.910799`

Probably because I tried ur curl locally

### Citation 15

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_c468dbab1a8237763322ba522a5e834c`
- `chunk_id`: `srcchunk_129dd7d5eea70311bb73aac91b0fb284`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775846605.865309`

ah yes

### Citation 16

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_d93dfc4bee6f41f025e5827e85d0d174`
- `chunk_id`: `srcchunk_dbf2146b10ff83d0ca1b49846250841e`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775847180.033049`

could u invalidate my user on the backend somehw

### Citation 17

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_da74749fbe7a3fb4068a2e35fe015483`
- `chunk_id`: `srcchunk_ae0413bc820a36a1a74fcca6e2a9e43b`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775847183.033049`

so i can test 1 more time

### Citation 18

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_d97aa54920cc140b58728c6574a7432f`
- `chunk_id`: `srcchunk_d21044415ec8e505cc9e883bce1ffa09`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775847219.346999`

k

### Citation 19

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_0cfc37999bbd40a56fb113d26c76caa9`
- `chunk_id`: `srcchunk_d3d56c2b0e9a0b84608a0b19f4506815`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775847415.065749`

try again

### Citation 20

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_3de401c74cb496e110d605fdfeb0bdb0`
- `chunk_id`: `srcchunk_ed68b176002d9b211de1e277ffad308c`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775847540.357119`

yessss its fking working

### Citation 21

- `source_document_id`: `srcdoc_6f05f8e43c03e102e02b43b9ef35e5ef`
- `source_revision_id`: `srcrev_a07a4a433e0b5bcfcbcc597c51afc807`
- `chunk_id`: `srcchunk_f5b211ab704b0cd4ce33d341771c2069`
- `native_locator`: `slack:C0AKH5SNGKH:1775798436.158669:1775847541.920789`

thanks blake

