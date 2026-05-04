---
title: "Slack thread C0AKH5SNGKH 1776304311.244769"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_151ae5311c913a12415df22d4f4b9b5f"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776304311.244769"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776304311.244769"
source_revision_ids:
  - "srcrev_216797a94f9ee1729886139c2e061d75"
  - "srcrev_2b4499a81abb85f95ec784b2f8116419"
  - "srcrev_3ae947a09f7aa0dc02bdf0d500fd527d"
  - "srcrev_4dae1f0656dd1e4796ea603fb6023ea9"
  - "srcrev_4e5e4870c726761a68066b680e591187"
  - "srcrev_5ef19461f26f59150cdc19933ee5aae4"
  - "srcrev_6d2a83be62a11503b116ae4d29345540"
  - "srcrev_80afd2d9ad6f11fb84403a3462621c2d"
  - "srcrev_91e0dcbfc86714c04eade13c48a47643"
  - "srcrev_a77cce31882d3dec961b9a30fe6a4252"
  - "srcrev_a836b7ef7426b3867daba57e51888168"
  - "srcrev_d474846c3429c59be8b3eb3690896a7b"
  - "srcrev_d4d08c92ef2eba12ec81611d2fe8983d"
  - "srcrev_e8c3b691e05dbb4ae37286dee7919461"
  - "srcrev_f0f37607597e4d2e93607f4622efb57e"
  - "srcrev_f7df39fe77c2141f2fcde245fa2d88d8"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1776304311.244769

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_f0f37607597e4d2e93607f4622efb57e`
- `chunk_id`: `srcchunk_954b6b535aa93fc4afd7069f3267661d`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776304311.244769`

<@U04L0DD6B6F> QA update:
1. Load test preparation
    ◦ Add admin fetch submission details endpoint
2. Perform initial round load test,  quick summary:
    ◦ Result: <https://www.notion.so/storyprotocol/STAGING-Numo-Load-Test-Result-20260415-344051299a54805c8a78e850575776cd?source=copy_link|[STAGING] Numo Load Test Result - 20260415> 
    ◦ Scenario
        ▪︎ 17 hours audio data (~1500 files)
        ▪︎ 100 Numo accounts,  100 concurrent Users
    ◦ Submission
        ▪︎ RPS: *19.05*
        ▪︎ Success Rate, *96.59%*.  53failed/1553uploads, failed reason: `confirm_upload failed: 422 {"error":"unprocessable","message":"uploaded artifact content_type does not match file contents"}"`  cc <@U0772SH7BRA> 
        ▪︎ Upload Latency,  *avg 1.934s*
    ◦ IP registration
        ▪︎ Success Rate, *19.47%.*    1208 pending, seems the job isn’t running cc <@U08V4SFU7LZ> 
        ▪︎ Registration Latency,  *avg 150s*

### Citation 2

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_d4d08c92ef2eba12ec81611d2fe8983d`
- `chunk_id`: `srcchunk_b1555ef746dd406ef82be66ba47ef3d7`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776304469.441159`

<@U08951K4SRY> what errors messages do you see?

### Citation 3

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_3ae947a09f7aa0dc02bdf0d500fd527d`
- `chunk_id`: `srcchunk_f20af78f0927fd817eb0dfaf0f04cb46`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776304702.684379`

<@U08V4SFU7LZ> still don’t know how to check the service logs, but the status in db is stuck at pending and the registration not progressing

### Citation 4

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_a836b7ef7426b3867daba57e51888168`
- `chunk_id`: `srcchunk_2f1159ad314c6e2895ce44e23f9bbfd7`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776305788.697199`

<@U08951K4SRY> where is the code for the load test located

### Citation 5

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_e8c3b691e05dbb4ae37286dee7919461`
- `chunk_id`: `srcchunk_73fbe5d61e98c2ced5e352806b85bca8`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776305862.096459`

<@U08951K4SRY> found the problem, it looks like the gas price was too high on aeneid for the hardcoded limit

### Citation 6

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_216797a94f9ee1729886139c2e061d75`
- `chunk_id`: `srcchunk_e14df8d72912a212d2d864c94ed7ab81`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776305896.492239`

will fix it

### Citation 7

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_f7df39fe77c2141f2fcde245fa2d88d8`
- `chunk_id`: `srcchunk_ac9f3afb7a8cea475489854a0932acad`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776305945.352269`

yes, we’ve talked about this, it's the same for psdn depin

### Citation 8

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_4e5e4870c726761a68066b680e591187`
- `chunk_id`: `srcchunk_b16a4f558774a8dc9b84ac3e9568efa4`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776306084.294099`

<@U0772SH7BRA> <https://github.com/piplabs/story-api-tests/tree/numo-testing/src/test/depin/numo_test/locustfiles>
I’m still debugging GHA,  it will be ready later

### Citation 9

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_6d2a83be62a11503b116ae4d29345540`
- `chunk_id`: `srcchunk_131babf55f1e41477c59bac84f3787de`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776356178.352579`

cc <@U04L0DD6B6F> ^

### Citation 10

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_5ef19461f26f59150cdc19933ee5aae4`
- `chunk_id`: `srcchunk_e347c0c817b7931aece0dd606a7944e9`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776370391.075019`

<@U08951K4SRY> for this load test it seems that the file you uploaded probably didn't match the declared type. Did you verify before upload ?

### Citation 11

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_2b4499a81abb85f95ec784b2f8116419`
- `chunk_id`: `srcchunk_70925b0a01b95a5d1c2ef1885eee8fb9`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776371107.739179`

All the files I’m using are .webm files, but some of them may not actually be that file type. I’ll check.

### Citation 12

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_a77cce31882d3dec961b9a30fe6a4252`
- `chunk_id`: `srcchunk_da3e96d0b4531cad279902a1e13d12bc`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776371476.381379`

<@U0772SH7BRA> 1 of the 30 files is actually a wav file, this should be the reason for 422 error. I’ll remove it and try again.

### Citation 13

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_4dae1f0656dd1e4796ea603fb6023ea9`
- `chunk_id`: `srcchunk_c8c500e4cb0f05a67a33b94ce9a5812e`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776371795.221989`

why are there different formats? is it due to World vs web app?

### Citation 14

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_d474846c3429c59be8b3eb3690896a7b`
- `chunk_id`: `srcchunk_997815a268439e71dfeef7544c4d4c37`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776371885.487079`

these are the pre prepared files we set up for api load test, not recorded from FE

### Citation 15

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_80afd2d9ad6f11fb84403a3462621c2d`
- `chunk_id`: `srcchunk_b288423b5bc782f2637f918efa96696b`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776371969.461329`

<@U04L0DD6B6F> but from the psdn experience, different devices or browsers may record different formats, I’ll keep an eye on that when testing compatibility

### Citation 16

- `source_document_id`: `srcdoc_151ae5311c913a12415df22d4f4b9b5f`
- `source_revision_id`: `srcrev_91e0dcbfc86714c04eade13c48a47643`
- `chunk_id`: `srcchunk_081e1180eca5a68fafe4e51278859ce8`
- `native_locator`: `slack:C0AKH5SNGKH:1776304311.244769:1776372013.444349`

i added more clear message on the backend, it will tell u detected file type vs declared type

