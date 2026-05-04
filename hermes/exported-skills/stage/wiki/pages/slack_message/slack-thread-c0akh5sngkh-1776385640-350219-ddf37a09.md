---
title: "Slack thread C0AKH5SNGKH 1776385640.350219"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_a9af65ecff866c46152ac869ddf37a09"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776385640.350219"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1776385640.350219"
source_revision_ids:
  - "srcrev_245b9c5ae757d6f1f6834c0c6b0befd0"
  - "srcrev_52652ec09a505d89832c4c1dbd5d95ce"
  - "srcrev_7998342295ad9528108d54024124e95a"
  - "srcrev_8ddc1bf806c6ac6842f0758b7064b736"
  - "srcrev_968f584471a3cd44c0a7ef5607bdb77a"
  - "srcrev_b67b31d6766b1a8f9a61c3e3a0396602"
  - "srcrev_ec6d6e07590ad3ff4d83ae6ad38b9c16"
  - "srcrev_f6cf5018546addfa3bf59ddf592fb373"
  - "srcrev_f81467f0fcfad33f0b93dd92380cf329"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1776385640.350219

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_a9af65ecff866c46152ac869ddf37a09`
- `source_revision_id`: `srcrev_968f584471a3cd44c0a7ef5607bdb77a`
- `chunk_id`: `srcchunk_268b948df88aca89438ef2896bf3dcb7`
- `native_locator`: `slack:C0AKH5SNGKH:1776385640.350219:1776385640.350219`

<@U0772SH7BRA> after a few iterations <@U08951K4SRY> and I saw that gas price spikes much faster on Aeneid than on Mainnet, so I'll make the gas limit only enforced on Mainnet. here's the PR, can you approve it if all good to unblock testing?
<https://github.com/piplabs/depin-backend/pull/148>

### Citation 2

- `source_document_id`: `srcdoc_a9af65ecff866c46152ac869ddf37a09`
- `source_revision_id`: `srcrev_b67b31d6766b1a8f9a61c3e3a0396602`
- `chunk_id`: `srcchunk_7b73393ebcbb929618bfa5d387c3eeee`
- `native_locator`: `slack:C0AKH5SNGKH:1776385640.350219:1776385958.068159`

<@U08951K4SRY> you should be able to retry now

### Citation 3

- `source_document_id`: `srcdoc_a9af65ecff866c46152ac869ddf37a09`
- `source_revision_id`: `srcrev_8ddc1bf806c6ac6842f0758b7064b736`
- `chunk_id`: `srcchunk_7833ea3a854e54b9ff4641dbeff484d3`
- `native_locator`: `slack:C0AKH5SNGKH:1776385640.350219:1776387361.050379`

<@U08V4SFU7LZ> can you share more info on why that is? Also what’s worse case scenario if this happens on mainnet

### Citation 4

- `source_document_id`: `srcdoc_a9af65ecff866c46152ac869ddf37a09`
- `source_revision_id`: `srcrev_f81467f0fcfad33f0b93dd92380cf329`
- `chunk_id`: `srcchunk_2dff9ee9759f122bd47d31f08e13906f`
- `native_locator`: `slack:C0AKH5SNGKH:1776385640.350219:1776387577.779359`

<@U04L0DD6B6F> I saw today that on Aeneid gas price spiked beyond 18 gwei, whereas in the old `depin-api` repo the hardcoded gas price limit was much lower
I asked Seb and he confirmed that gas prices on testnet tend to be way higher, though I'm unsure about the exact reason why

Worse case what'd happen is jobs get deferred until the gas price drops, and if it doesn't happen we can raise the limit

The `depin-api` repo used 0.065 gwei as a limit, for now I hardcoded 0.1 gwei to give us some margin

### Citation 5

- `source_document_id`: `srcdoc_a9af65ecff866c46152ac869ddf37a09`
- `source_revision_id`: `srcrev_7998342295ad9528108d54024124e95a`
- `chunk_id`: `srcchunk_cf5b5156874c41ac94e568c0764e90f0`
- `native_locator`: `slack:C0AKH5SNGKH:1776385640.350219:1776387626.590209`

can we use this as opportunity to test the state machine for gas bumping

### Citation 6

- `source_document_id`: `srcdoc_a9af65ecff866c46152ac869ddf37a09`
- `source_revision_id`: `srcrev_245b9c5ae757d6f1f6834c0c6b0befd0`
- `chunk_id`: `srcchunk_469ee1ef9e919e81ed7ed4358213bff2`
- `native_locator`: `slack:C0AKH5SNGKH:1776385640.350219:1776387633.371099`

just make the limit super high or something

### Citation 7

- `source_document_id`: `srcdoc_a9af65ecff866c46152ac869ddf37a09`
- `source_revision_id`: `srcrev_ec6d6e07590ad3ff4d83ae6ad38b9c16`
- `chunk_id`: `srcchunk_aeec465570e5d74c19d6ec6e7afef717`
- `native_locator`: `slack:C0AKH5SNGKH:1776385640.350219:1776387679.557369`

There's no limit on testnet so we can definitely try it
<@U08951K4SRY> do you want to try?

### Citation 8

- `source_document_id`: `srcdoc_a9af65ecff866c46152ac869ddf37a09`
- `source_revision_id`: `srcrev_f6cf5018546addfa3bf59ddf592fb373`
- `chunk_id`: `srcchunk_112450e32fe8ca6ba78c8f17daabed69`
- `native_locator`: `slack:C0AKH5SNGKH:1776385640.350219:1776387969.500739`

<@U04L0DD6B6F> I can also remove the gas limit entirely if we don't mind paying high gas fees during a spike as long as IPs are registered?

### Citation 9

- `source_document_id`: `srcdoc_a9af65ecff866c46152ac869ddf37a09`
- `source_revision_id`: `srcrev_52652ec09a505d89832c4c1dbd5d95ce`
- `chunk_id`: `srcchunk_b80d117d4b7a6b3ec898f6e27e0fe5b9`
- `native_locator`: `slack:C0AKH5SNGKH:1776385640.350219:1776388023.649139`

I think it should just queue it n retry until gas price falls below our configured limit

