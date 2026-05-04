---
title: "Slack thread C0AKH5SNGKH 1777057603.999419"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_4cd265c0f86d89c52b669714acd28d76"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1777057603.999419"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1777057603.999419"
source_revision_ids:
  - "srcrev_67e25cb73ffd6ba0bb389c9d39f8064f"
  - "srcrev_943e3b13480e14e913dbea4a8f121e36"
  - "srcrev_a66da2dd140cc24fc44b5ae6d4a16648"
  - "srcrev_ac8243bb5134a72e26fe32f502dd17c9"
  - "srcrev_b8f852d3dff161b91148686bcf028edc"
  - "srcrev_ca5da3d0d06669c79bfee21072a8120d"
  - "srcrev_e992682cc8f4e6d6d538609d540f9f64"
  - "srcrev_f7ea11178aaaf3b268d9a2093dd08f78"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1777057603.999419

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_4cd265c0f86d89c52b669714acd28d76`
- `source_revision_id`: `srcrev_67e25cb73ffd6ba0bb389c9d39f8064f`
- `chunk_id`: `srcchunk_cced56cc1cc329252f432de3c60de45d`
- `native_locator`: `slack:C0AKH5SNGKH:1777057603.999419:1777057603.999419`

<@U0772SH7BRA> I tried to clean up IP registration wallets and jobs on staging but I'm hitting 504 error, cloudfare might be the bottleneck. Any way to bypass the CDN or access the DB directly?

### Citation 2

- `source_document_id`: `srcdoc_4cd265c0f86d89c52b669714acd28d76`
- `source_revision_id`: `srcrev_b8f852d3dff161b91148686bcf028edc`
- `chunk_id`: `srcchunk_c023feff78d51c178c6311357700056f`
- `native_locator`: `slack:C0AKH5SNGKH:1777057603.999419:1777062539.946429`

cc <@U08332YRB7W> <@U07TNT9N4JC>

### Citation 3

- `source_document_id`: `srcdoc_4cd265c0f86d89c52b669714acd28d76`
- `source_revision_id`: `srcrev_ca5da3d0d06669c79bfee21072a8120d`
- `chunk_id`: `srcchunk_cba31215e2215d442832f5c642f80e73`
- `native_locator`: `slack:C0AKH5SNGKH:1777057603.999419:1777062679.654529`

That doesn't sound like cloudflare issue unless you were doing it via API.

### Citation 4

- `source_document_id`: `srcdoc_4cd265c0f86d89c52b669714acd28d76`
- `source_revision_id`: `srcrev_e992682cc8f4e6d6d538609d540f9f64`
- `chunk_id`: `srcchunk_f6d9f62fb126a55098ed69285eed202a`
- `native_locator`: `slack:C0AKH5SNGKH:1777057603.999419:1777062766.963409`

I was actually hitting the admin API endpoints

### Citation 5

- `source_document_id`: `srcdoc_4cd265c0f86d89c52b669714acd28d76`
- `source_revision_id`: `srcrev_a66da2dd140cc24fc44b5ae6d4a16648`
- `chunk_id`: `srcchunk_9efb0f7eefd1a5846ae095910905b216`
- `native_locator`: `slack:C0AKH5SNGKH:1777057603.999419:1777062840.317909`

I see. Do we even have the functionality built into API to remove/delete wallets?

### Citation 6

- `source_document_id`: `srcdoc_4cd265c0f86d89c52b669714acd28d76`
- `source_revision_id`: `srcrev_ac8243bb5134a72e26fe32f502dd17c9`
- `chunk_id`: `srcchunk_0aec75a168a8dec6c7e388a059beaeb7`
- `native_locator`: `slack:C0AKH5SNGKH:1777057603.999419:1777062884.729059`

Yeah
 POST <https://staging-depin.storyprotocol.net/v1/admin/ip-registration/wallets/wipe> to remove wallets + associated jobs
 POST <https://staging-depin.storyprotocol.net/v1/admin/ip-registration/wallets/delete> to remove only wallets

### Citation 7

- `source_document_id`: `srcdoc_4cd265c0f86d89c52b669714acd28d76`
- `source_revision_id`: `srcrev_943e3b13480e14e913dbea4a8f121e36`
- `chunk_id`: `srcchunk_6829a025678c44c14a9c430d8590d49b`
- `native_locator`: `slack:C0AKH5SNGKH:1777057603.999419:1777063022.493309`

Probably the service is down. hence you got 504

### Citation 8

- `source_document_id`: `srcdoc_4cd265c0f86d89c52b669714acd28d76`
- `source_revision_id`: `srcrev_f7ea11178aaaf3b268d9a2093dd08f78`
- `chunk_id`: `srcchunk_cf9b2a71012001b2d4b31b95c94a13b5`
- `native_locator`: `slack:C0AKH5SNGKH:1777057603.999419:1777063084.870179`

looks like the service is up and running: <https://argocd.ops.storyprotocol.net/applications/argocd/use1-stage-depin-backend?view=tree&amp;resource=>

