---
title: "Archive/Delete Old GitHub Repositories Decision"
type: "decision"
slug: "decisions/archive-delete-old-github-repos"
freshness: "2026-01-30T01:26:58Z"
tags:
  - "cleanup"
  - "infrastructure"
  - "repository-management"
owners:
  - "U04KTUN5WFQ"
  - "U07TNT9N4JC"
  - "U080YAW205V"
source_revision_ids:
  - "srcrev_6860e0662fa5731ed0b390d06e655676"
  - "srcrev_cc7b315f4a5fea3edff6f8cb46187c3b"
  - "srcrev_e39bd8ff330c3922c77e789a5626eeee"
  - "srcrev_f3e5ce69b218c11e6a501b15def7a5a3"
conflict_state: "none"
---

# Archive/Delete Old GitHub Repositories Decision

## Summary

Decision to archive or delete GitHub repositories that haven't been committed to in over one year, based on a Slack discussion on 2026-01-30.

## Claims

- The repositories under review are: story-golden-image, iac-pro, terraform-modules, iac-max, iac-max-modules, story-terraform, provision-hardened-node, story-vpc-gcp. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cc0076ff8c6c099a331cdfbecba955b6` `source_revision_id=srcrev_e39bd8ff330c3922c77e789a5626eeee` `chunk_id=srcchunk_a0322a26f7ff7ef73793647fcedcb654` `native_locator=slack:C0547N89JUB:1769728055.775439:1769728055.775439` `source_timestamp=2026-01-29T23:07:35Z`
- It was suggested that repos with last commit 1+ years ago are safe to archive. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cc0076ff8c6c099a331cdfbecba955b6` `source_revision_id=srcrev_f3e5ce69b218c11e6a501b15def7a5a3` `chunk_id=srcchunk_b8a24f56e4620eb135985b5cc5e450c0` `native_locator=slack:C0547N89JUB:1769728055.775439:1769728347.702529` `source_timestamp=2026-01-29T23:12:27Z`
- User U080YAW205V confirmed it is okay to archive. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cc0076ff8c6c099a331cdfbecba955b6` `source_revision_id=srcrev_cc7b315f4a5fea3edff6f8cb46187c3b` `chunk_id=srcchunk_398c53b531baa96c567f99d04607aecd` `native_locator=slack:C0547N89JUB:1769728055.775439:1769733054.321389` `source_timestamp=2026-01-30T00:30:54Z`
- A contributor (likely U07TNT9N4JC) requested to drop all repos created by them. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_cc0076ff8c6c099a331cdfbecba955b6` `source_revision_id=srcrev_6860e0662fa5731ed0b390d06e655676` `chunk_id=srcchunk_0f7ba6a2f4de1414533d487ea9647315` `native_locator=slack:C0547N89JUB:1769728055.775439:1769736418.190519` `source_timestamp=2026-01-30T01:26:58Z`

## Open Questions

- Are the repositories now archived? Who will perform the archiving?

## Sources

- `source_document_id`: `srcdoc_cc0076ff8c6c099a331cdfbecba955b6`
- `source_revision_id`: `srcrev_cc7b315f4a5fea3edff6f8cb46187c3b`
