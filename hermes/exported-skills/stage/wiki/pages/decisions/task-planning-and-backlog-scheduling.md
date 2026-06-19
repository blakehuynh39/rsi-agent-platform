---
title: "Task Planning and Backlog Scheduling"
type: "decision"
slug: "decisions/task-planning-and-backlog-scheduling"
freshness: "2026-05-06T17:13:32Z"
tags:
  - "admin-portal"
  - "backlog"
  - "calendar"
  - "marketing-access"
  - "standup"
  - "tasks"
owners:
  - "U04L0DD6B6F"
  - "U04L0DD71TM"
  - "U06A5AQ1VD3"
  - "U0772SH7BRA"
  - "U0866HDK755"
  - "U086FECSTP1"
  - "U0883L0RBRR"
  - "U0AQZPN6ZQV"
  - "U0AU3DWLVE2"
source_revision_ids:
  - "srcrev_389f4ddc96addb62a513d247b1d58e6f"
  - "srcrev_5333fb03eb6628f4e29b691ba30bdbed"
  - "srcrev_5cdd10ae1d5dfb40578f2f73c26e2fa4"
  - "srcrev_6491695b6e5ab6f5b4f842d60e2c5733"
  - "srcrev_6758f0b44230d3cbe9f93b4b27b0a32c"
  - "srcrev_9389be5adda456c64cb8cba65870e513"
  - "srcrev_ef61f9b44110b25927e24d0049c608cd"
  - "srcrev_f3775e787e98e20b90669e1a64e30a79"
conflict_state: "none"
---

# Task Planning and Backlog Scheduling

## Summary

Team discussion to address confusion around task scheduling by creating a structured backlog and a calendar tool in the admin portal for visualizing rollout. Tasks will be repopulated based on tech feasibility, and the tool will be walked through at standup. Marketing requests access but rollout awaits sensitivity configuration.

## Claims

- There is confusion around tasks, specifically what tasks are next or upcoming and what process for scheduling those. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d1b51300541356139029602a8cea4bdc` `source_revision_id=srcrev_5333fb03eb6628f4e29b691ba30bdbed` `chunk_id=srcchunk_eec509c8feb6aa1f1c58cf34ce52318b` `native_locator=slack:C0AL7EKNHDF:1778082109.181859:1778082109.181859` `source_timestamp=2026-05-06T15:48:37Z`
- It would be beneficial to go through the task backlog at standup and create a scheduled task plan. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d1b51300541356139029602a8cea4bdc` `source_revision_id=srcrev_5333fb03eb6628f4e29b691ba30bdbed` `chunk_id=srcchunk_eec509c8feb6aa1f1c58cf34ce52318b` `native_locator=slack:C0AL7EKNHDF:1778082109.181859:1778082109.181859` `source_timestamp=2026-05-06T15:48:37Z`
- A preliminary backlog was created, but the decision on what to launch when should be made by the team. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d1b51300541356139029602a8cea4bdc` `source_revision_id=srcrev_5333fb03eb6628f4e29b691ba30bdbed` `chunk_id=srcchunk_eec509c8feb6aa1f1c58cf34ce52318b` `native_locator=slack:C0AL7EKNHDF:1778082109.181859:1778082109.181859` `source_timestamp=2026-05-06T15:48:37Z`
- A Notion page titled 'Numo Backlogs' exists and was shared with the team. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d1b51300541356139029602a8cea4bdc` `source_revision_id=srcrev_6758f0b44230d3cbe9f93b4b27b0a32c` `chunk_id=srcchunk_6783720f049a1a9776924c9373fc4574` `native_locator=slack:C0AL7EKNHDF:1778082109.181859:1778082126.237759` `source_timestamp=2026-05-06T15:42:06Z`
- The team plans to meet and repopulate the backlog by end of day based on tech feasibility. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d1b51300541356139029602a8cea4bdc` `source_revision_id=srcrev_f3775e787e98e20b90669e1a64e30a79` `chunk_id=srcchunk_d9c3fc0a03b21b82f5f07c5a52db7aad` `native_locator=slack:C0AL7EKNHDF:1778082109.181859:1778082831.796589` `source_timestamp=2026-05-06T15:53:51Z`
- It's been unclear what the roadmap is regarding when/which countries/what modalities and why. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d1b51300541356139029602a8cea4bdc` `source_revision_id=srcrev_5cdd10ae1d5dfb40578f2f73c26e2fa4` `chunk_id=srcchunk_668fe16e88c73d165c22fc94000b90bb` `native_locator=slack:C0AL7EKNHDF:1778082109.181859:1778083149.091669` `source_timestamp=2026-05-06T15:59:09Z`
- A calendar tool in the admin portal has been prepared to visualize weekly, monthly, Gantt chart task rollout. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d1b51300541356139029602a8cea4bdc` `source_revision_id=srcrev_9389be5adda456c64cb8cba65870e513` `chunk_id=srcchunk_09c964a8f635811f0ac75e3c6937d783` `native_locator=slack:C0AL7EKNHDF:1778082109.181859:1778087146.533499` `source_timestamp=2026-05-06T17:05:46Z`
- The calendar tool will help schedule tasks in advance for automatic rollout on the app and alignment between engineering and marketing. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d1b51300541356139029602a8cea4bdc` `source_revision_id=srcrev_9389be5adda456c64cb8cba65870e513` `chunk_id=srcchunk_09c964a8f635811f0ac75e3c6937d783` `native_locator=slack:C0AL7EKNHDF:1778082109.181859:1778087146.533499` `source_timestamp=2026-05-06T17:05:46Z`
- The calendar tool will be briefly walked through during tomorrow's standup. `claim:claim_1_9` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d1b51300541356139029602a8cea4bdc` `source_revision_id=srcrev_6491695b6e5ab6f5b4f842d60e2c5733` `chunk_id=srcchunk_0a5a7074388eda2898affb0f43708d9c` `native_locator=slack:C0AL7EKNHDF:1778082109.181859:1778087171.564159` `source_timestamp=2026-05-06T17:06:11Z`
- Marketing requests access to the calendar tool. `claim:claim_1_10` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d1b51300541356139029602a8cea4bdc` `source_revision_id=srcrev_ef61f9b44110b25927e24d0049c608cd` `chunk_id=srcchunk_b9eb49b32addb7fd10d608e9bc6fd87c` `native_locator=slack:C0AL7EKNHDF:1778082109.181859:1778087462.198699` `source_timestamp=2026-05-06T17:11:02Z`
- Access to the calendar tool will be rolled out after setting up access levels on emails due to sensitive rewards/payment knobs. `claim:claim_1_11` `confidence:1.00`
  - citation: `source_document_id=srcdoc_d1b51300541356139029602a8cea4bdc` `source_revision_id=srcrev_389f4ddc96addb62a513d247b1d58e6f` `chunk_id=srcchunk_9f005d95f5f9370a5f85c266a5c095ae` `native_locator=slack:C0AL7EKNHDF:1778082109.181859:1778087612.393989` `source_timestamp=2026-05-06T17:13:32Z`

## Sources

- `source_document_id`: `srcdoc_d1b51300541356139029602a8cea4bdc`
- `source_revision_id`: `srcrev_389f4ddc96addb62a513d247b1d58e6f`
