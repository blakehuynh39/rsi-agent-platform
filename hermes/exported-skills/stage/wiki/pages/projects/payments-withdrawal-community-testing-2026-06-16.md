---
title: "Payments Withdrawal Community Testing (2026-06-16)"
type: "project"
slug: "projects/payments-withdrawal-community-testing-2026-06-16"
freshness: "2026-06-16T17:02:27Z"
tags:
  - "payments"
  - "stripe"
  - "tax-forms"
  - "testing"
  - "withdrawal"
owners:
  - "payments-team"
source_revision_ids:
  - "srcrev_04d6cdbc0e2416eb77553a0a5bc5d747"
  - "srcrev_33f6977d4c0f7c63094376c3fd6e8089"
  - "srcrev_4b35274e2d1772b8575ff0fa1104d328"
  - "srcrev_86743c1ad11e4aafedbc61f7439bcdae"
  - "srcrev_98e6cccd12b11b5e5fc96f2765bbc519"
  - "srcrev_aa2bf468a12224d0193e8c642cfab5c4"
  - "srcrev_e830078f65f72231ea14a8dcf887727c"
  - "srcrev_eef31c1526760c885046edb18465cd80"
  - "srcrev_f85a1a2424b6d8dde2cbd4806cd9c497"
conflict_state: "none"
---

# Payments Withdrawal Community Testing (2026-06-16)

## Summary

Community withdrawal testing conducted on June 16, 2026, involving testers from USA, Vietnam, Indonesia, Pakistan, and Malaysia. Main friction was unclear tax form approval status. Payment statuses varied: Indonesia posted, Pakistan initiated but not posted, Vietnam not initiated.

## Claims

- Testing group split into two: ASAP test with tax form working and later test after new UX changes. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_f85a1a2424b6d8dde2cbd4806cd9c497` `chunk_id=srcchunk_fb317acbcf6811c37e6632b81b1936b8` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781504198.641459` `source_timestamp=2026-06-15T06:16:38Z`
- USA tester waited 24+ hours for W9 approval but could withdraw. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_f85a1a2424b6d8dde2cbd4806cd9c497` `chunk_id=srcchunk_fb317acbcf6811c37e6632b81b1936b8` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781504198.641459` `source_timestamp=2026-06-15T06:16:38Z`
- Vietnam tester waited long for W-8BEN approval (unsure time) but could withdraw. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_f85a1a2424b6d8dde2cbd4806cd9c497` `chunk_id=srcchunk_fb317acbcf6811c37e6632b81b1936b8` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781504198.641459` `source_timestamp=2026-06-15T06:16:38Z`
- Indonesia tester had W-8BEN submitted and approved within 30 minutes, enabling quick withdrawal. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_f85a1a2424b6d8dde2cbd4806cd9c497` `chunk_id=srcchunk_fb317acbcf6811c37e6632b81b1936b8` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781504198.641459` `source_timestamp=2026-06-15T06:16:38Z`
- Malaysia tester acknowledged but not tested. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_f85a1a2424b6d8dde2cbd4806cd9c497` `chunk_id=srcchunk_fb317acbcf6811c37e6632b81b1936b8` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781504198.641459` `source_timestamp=2026-06-15T06:16:38Z`
- Pakistan tester waited 12+ hours for W-8BEN approval but could withdraw. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_f85a1a2424b6d8dde2cbd4806cd9c497` `chunk_id=srcchunk_fb317acbcf6811c37e6632b81b1936b8` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781504198.641459` `source_timestamp=2026-06-15T06:16:38Z`
- Biggest friction point was uncertainty about tax form approval status; suggested status on page or email update. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_f85a1a2424b6d8dde2cbd4806cd9c497` `chunk_id=srcchunk_fb317acbcf6811c37e6632b81b1936b8` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781504198.641459` `source_timestamp=2026-06-15T06:16:38Z`
- Indonesia payment posted about 26 hours ago, need user confirmation of receipt. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_04d6cdbc0e2416eb77553a0a5bc5d747` `chunk_id=srcchunk_e0e23439b5a62be0a3c0bb2f5cc27dae` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781508790.722449` `source_timestamp=2026-06-15T07:33:10Z`
- Vietnam payment not initiated; approval may be pending manually. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_04d6cdbc0e2416eb77553a0a5bc5d747` `chunk_id=srcchunk_e0e23439b5a62be0a3c0bb2f5cc27dae` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781508790.722449` `source_timestamp=2026-06-15T07:33:10Z`
- Pakistan payment initiated but not posted; user received Stripe payment email about 1h50min ago. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_4b35274e2d1772b8575ff0fa1104d328` `chunk_id=srcchunk_0055179d4139aeb901faade0ccddff3c` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781538570.717129` `source_timestamp=2026-06-15T15:49:30Z`
- Pakistan user eventually received 6,959 PKR. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_aa2bf468a12224d0193e8c642cfab5c4` `chunk_id=srcchunk_d5583e65e8c69a6cbd7215c902e91168` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781629347.456089` `source_timestamp=2026-06-16T17:02:27Z`
- Philippines tester requested to be added for next test after UX changes. `claim:claim_1_12` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_eef31c1526760c885046edb18465cd80` `chunk_id=srcchunk_af2404dee73a2fd679788b05a2a7e167` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781545045.308779` `source_timestamp=2026-06-15T17:37:25Z`
- UX bugs fixed; staging tests completed; awaiting deployment to production. `claim:claim_1_13` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_33f6977d4c0f7c63094376c3fd6e8089` `chunk_id=srcchunk_a1179a7b4c10ed3a87e61d7a346b4693` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781568228.637619` `source_timestamp=2026-06-16T00:03:48Z`
- Need to coordinate Poseidon validation results propagation to Numo. `claim:claim_1_14` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_e830078f65f72231ea14a8dcf887727c` `chunk_id=srcchunk_d5818bc7e9fd1043e7d9ce83b6d7733f` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781569301.571989` `source_timestamp=2026-06-16T00:21:41Z`
- Stablecoin updates and other withdrawal solutions required ASAP; timeline requested. `claim:claim_1_15` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_86743c1ad11e4aafedbc61f7439bcdae` `chunk_id=srcchunk_3a39af5ea237a021852512d5e0fedb23` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781571993.160809` `source_timestamp=2026-06-16T01:09:01Z`
- All payments-related tasks aimed to be completed by end of month (EOM, ~2 weeks from June 16, 2026). `claim:claim_1_16` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_86743c1ad11e4aafedbc61f7439bcdae` `chunk_id=srcchunk_3a39af5ea237a021852512d5e0fedb23` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781571993.160809` `source_timestamp=2026-06-16T01:09:01Z`
- User notification emails for payments might need to be built; a batch cron job considered. `claim:claim_1_17` `confidence:1.00`
  - citation: `source_document_id=srcdoc_a9ea744764ef6595215cc1eeb90da0e9` `source_revision_id=srcrev_98e6cccd12b11b5e5fc96f2765bbc519` `chunk_id=srcchunk_0ca122a96cabb6b36535977bcc9d89e0` `native_locator=slack:C0AL7EKNHDF:1781504198.641459:1781545706.245039` `source_timestamp=2026-06-15T17:48:26Z`

## Open Questions

- Has Indonesia tester confirmed receiving the payment in their bank?
- Has Vietnam payment been initiated after approval?
- Is Philippines tester added and funded for the next test round?
- What is the timeline for stablecoin updates?
- When will user notification emails be implemented?

## Sources

- `source_document_id`: `srcdoc_a9ea744764ef6595215cc1eeb90da0e9`
- `source_revision_id`: `srcrev_6c608b707dc4bd16ac61ddcf8d055c7b`
