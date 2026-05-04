---
title: "Slack thread C0AKH5SNGKH 1776887228.858889"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_c6dc668f1566abf8e5b59e61da5ac455"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776887228.858889"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776887228.858889"
source_revision_ids:
  - "srcrev_147bcfd436c318957c8358c50fa262eb"
  - "srcrev_17357408b95208f6f58437361f779a9e"
  - "srcrev_1d0df046b84caa3a017c39556f1fce21"
  - "srcrev_215776d6bd0dfa980e8fcc8d3efc9752"
  - "srcrev_45b10b6daf671924a62eb29b6f8c4180"
  - "srcrev_5e37e32c49b434f5f963ec0df8a54b08"
  - "srcrev_80b10e3a3f8f99c4875be42ea27022ea"
  - "srcrev_962b2d1aad3df9a506c7fa669dc68ae8"
  - "srcrev_a1c83abf6f3b4005a32fe3b93feca400"
  - "srcrev_a6dae5d61a95faca3e8f183c9b61070a"
  - "srcrev_b77e0487dd8ec871de74db9e43fbb883"
  - "srcrev_c273312a0cbac2ee3a126751dfcfcd4e"
  - "srcrev_d2e886d8cc151cfe80baf1babbe43332"
  - "srcrev_d57f69616935ff92d2ed61926c91e47a"
  - "srcrev_ecc6c6f9d7e80d2af85aa7d004f34205"
  - "srcrev_f54f4693e395b8588dd8f16e7dc4af8d"
  - "srcrev_f868e599d7e78f5accc88afcf282c0f2"
  - "srcrev_fb41fcd083e6d858951433c6708022f4"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1776887228.858889

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_1d0df046b84caa3a017c39556f1fce21`
- `chunk_id`: `srcchunk_4a96a866f6cf9796e032d9493ea8dbe7`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1777008824.951269`

cc <@U04L0DD6B6F>
our rps looks better than psdn s1.
last time used R2, and with 500 concurrency the rps was ~13.
there should have some rate limit previously, and db often became a bottleneck at that time. we haven’t run into any db issues this time

### Citation 2

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_a1c83abf6f3b4005a32fe3b93feca400`
- `chunk_id`: `srcchunk_2e28f7c53965f5f6017a03305e4e19bf`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1776887228.858889`

<@U0772SH7BRA> <@U04L0DD6B6F> I ran a test with 500 concurrent users, the upload success rate was ~98%. When the RPS reached ~50, the following errors started to occur
<https://grafana.ops.storyprotocol.net/d/depin-backend-api/depin-backend-e28094-api-overview?orgId=1&amp;from=now-2h&amp;to=now&amp;timezone=browser&amp;var-environment=$__all&amp;refresh=30s|grafana metrics>
```
64times:  RuntimeError("PUT (presigned upload to object storage): HTTPSConnectionPool(host='<http://depin-backend-staging-private-media.s3.us-east-1.amazonaws.com|depin-backend-staging-private-media.s3.us-east-1.amazonaws.com>', port=443): Max retries exceeded with url: xxxx (Caused by SSLError(SSLZeroReturnError(6, 'TLS/SSL connection has been closed (EOF) (_ssl.c:1129)')))")

31times RuntimeError("POST /v1/admin/scripts: HTTPSConnectionPool(host='<http://staging-depin.storyprotocol.net|staging-depin.storyprotocol.net>', port=443): Read timed out. (read timeout=8)")

25times RuntimeError("GET /v1/scripts/next: HTTPSConnectionPool(host='<http://staging-depin.storyprotocol.net|staging-depin.storyprotocol.net>', port=443): Read timed out. (read timeout=8)")

11times RuntimeError("POST /v1/submissions/initiate-upload (presigned URL): HTTPSConnectionPool(host='<http://staging-depin.storyprotocol.net|staging-depin.storyprotocol.net>', port=443): Read timed out. (read timeout=8)")

3times RuntimeError("POST /v1/submissions/.../complete (confirm upload): HTTPSConnectionPool(host='<http://staging-depin.storyprotocol.net|staging-depin.storyprotocol.net>', port=443): Read timed out. (read timeout=8)")```

### Citation 3

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_d57f69616935ff92d2ed61926c91e47a`
- `chunk_id`: `srcchunk_1b8d1678d959566832967466a3fb429a`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1776887373.603049`

<@U08V4SFU7LZ> registration jobs seem to have stopped..  could it be the wallet balance is insufficient?

### Citation 4

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_215776d6bd0dfa980e8fcc8d3efc9752`
- `chunk_id`: `srcchunk_6119042cd2427b9286cd3695509bf108`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1776888077.752689`

<@U0772SH7BRA> is this on AWS side?

### Citation 5

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_c273312a0cbac2ee3a126751dfcfcd4e`
- `chunk_id`: `srcchunk_47c2f44cd96f66e13d183021a5c57a0f`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1776888119.122969`

opening a fix for the above. It's probably because staging isn't configured same as prod. I'm scaling it up so it's similar for load test

### Citation 6

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_17357408b95208f6f58437361f779a9e`
- `chunk_id`: `srcchunk_093a808cdcb2c1e828c8315ba81fff8f`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1776889345.407399`

<@U08951K4SRY> yeah the wallets ran out of gas

### Citation 7

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_147bcfd436c318957c8358c50fa262eb`
- `chunk_id`: `srcchunk_660f01b0ec3a62f9387062e4d8c7fcf5`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1776889900.105269`

<@U08951K4SRY> just funded the wallets, jobs should resume now

### Citation 8

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_d2e886d8cc151cfe80baf1babbe43332`
- `chunk_id`: `srcchunk_5711ba00841078bfc5c734bec62fe05d`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1777007744.204149`

<@U0772SH7BRA>  re-run 500 users again,  ~7500 uploads,  rps ~48.   still have few errors
```14 times,   RuntimeError("POST /v1/admin/scripts: HTTPSConnectionPool(host='<http://staging-depin.storyprotocol.net|staging-depin.storyprotocol.net>', port=443): Read timed out. (read timeout=8)")

17 times,  RuntimeError("GET /v1/scripts/next: HTTPSConnectionPool(host='<http://staging-depin.storyprotocol.net|staging-depin.storyprotocol.net>', port=443): Read timed out. (read timeout=8)")

69 times,  RuntimeError("PUT (presigned upload to object storage): HTTPSConnectionPool(host='<http://depin-backend-staging-private-media.s3.us-east-1.amazonaws.com|depin-backend-staging-private-media.s3.us-east-1.amazonaws.com>', port=443): Max retries exceeded with url: /users/xxx (Caused by SSLError(SSLZeroReturnError(6, 'TLS/SSL connection has been closed (EOF) (_ssl.c:1129)')))")

2 times, RuntimeError("PUT (presigned upload to object storage): HTTPSConnectionPool(host='<http://depin-backend-staging-private-media.s3.us-east-1.amazonaws.com|depin-backend-staging-private-media.s3.us-east-1.amazonaws.com>', port=443): Read timed out. (read timeout=300)")```

### Citation 9

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_fb41fcd083e6d858951433c6708022f4`
- `chunk_id`: `srcchunk_9640742ed96f9de1edc32e47ae76adaa`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1777008916.992499`

This is dope

### Citation 10

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_962b2d1aad3df9a506c7fa669dc68ae8`
- `chunk_id`: `srcchunk_5caf344815cded9806c090a5e14b0bba`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1777009137.860459`

but based on last time’s experience, if the launch gets a lot of traffic, we may still get the same errors with from the load test :joy:
do we have some plans,  like maybe quickly scaling up to handle the peak

### Citation 11

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_45b10b6daf671924a62eb29b6f8c4180`
- `chunk_id`: `srcchunk_07636c5d10ec6515dda2fbcc8a46aa51`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1777010993.177479`

<@U08951K4SRY> can you point me to the repo with the load test code again

### Citation 12

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_f868e599d7e78f5accc88afcf282c0f2`
- `chunk_id`: `srcchunk_498ee76e415e35ca2f3c51b7fe00e159`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1777011087.665989`

<@U0772SH7BRA> <https://github.com/piplabs/story-api-tests> , but I ran it locally

### Citation 13

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_b77e0487dd8ec871de74db9e43fbb883`
- `chunk_id`: `srcchunk_7e8bab7b7068673c9ee1c602427aebfb`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1777011402.586669`

<@U0772SH7BRA> yeah something to be mindful of we might face into scalability problem. Don't underestimate the spam and the bots.

### Citation 14

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_5e37e32c49b434f5f963ec0df8a54b08`
- `chunk_id`: `srcchunk_9a8a3c58ffd3025a9ad3860496721cb5`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1777011999.981769`

Royce here are my recs for the next run so we have a clearer picture
```Use 500 distinct JWTs, or label it as a 100-user backend test.
Ramp slower, e.g. -u 500 -r 50, not -r 500.
Set LOAD_TEST_PRE_CREATE_SCRIPT=0 and pre-seed scripts before the test.
Add retry/backoff for S3 PUT failures and GET exception timeouts.
Do not retry POST /v1/admin/scripts unless the test sends a unique import_key so retries are idempotent.```

### Citation 15

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_a6dae5d61a95faca3e8f183c9b61070a`
- `chunk_id`: `srcchunk_ea468f38a78681928d8a540ebf5069a7`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1777012060.005549`

The test creates a script before every upload by default. We already have enough scripts on both staging and backend

### Citation 16

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_80b10e3a3f8f99c4875be42ea27022ea`
- `chunk_id`: `srcchunk_e62fc214f1b40c0fb3e908b1958226e9`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1777012224.511519`

i'll optimize it as the suggestions. The only thing is that getting 500 JWTs is quite difficult, don’t really have an easy way to bypass the frontend to obtain them

### Citation 17

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_f54f4693e395b8588dd8f16e7dc4af8d`
- `chunk_id`: `srcchunk_1ec7b348341b37c541842d01750460a4`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1777012647.025869`

there should be some easy way or we can add a staging only route to mint JWT for testing

### Citation 18

- `source_document_id`: `srcdoc_c6dc668f1566abf8e5b59e61da5ac455`
- `source_revision_id`: `srcrev_ecc6c6f9d7e80d2af85aa7d004f34205`
- `chunk_id`: `srcchunk_350bffa203e0f92c60768399402fa6c8`
- `native_locator`: `slack:C0AKH5SNGKH:1776887228.858889:1777012660.560429`

you should do some research on dynamic xyz to see whether they let you script this

