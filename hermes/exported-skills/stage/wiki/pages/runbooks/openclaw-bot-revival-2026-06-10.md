---
title: "OpenClaw Bot Revival Incident (2026-06-10)"
type: "runbook"
slug: "runbooks/openclaw-bot-revival-2026-06-10"
freshness: "2026-06-10T10:25:52Z"
tags:
  - "incident"
  - "openclaw"
  - "security"
  - "slack"
owners: []
source_revision_ids:
  - "srcrev_810a8840b5b19cdec7019493bf44b00f"
  - "srcrev_8de0ad926156fda4df84ad16a8adb75e"
conflict_state: "none"
---

# OpenClaw Bot Revival Incident (2026-06-10)

## Summary

On 2026-06-10, a user revived the OpenClaw Slack bot. Runtime and gateway were operational, but config recovery failures, a CRITICAL security audit flag for group policy/tools, Composio auth failure, and a liveness warning were observed.

## Claims

- User revived the OpenClaw bot on 2026-06-10. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9efa3b1f8b79c3094352fadf14bd3e18` `source_revision_id=srcrev_8de0ad926156fda4df84ad16a8adb75e` `chunk_id=srcchunk_7a8174a472f0d68f23d78c94131bd0b5` `native_locator=slack:C0547N89JUB:1781087072.163719:1781087072.163719` `source_timestamp=2026-06-10T10:24:32Z`
- After revival, runtime was up, gateway reachable, Slack socket connected, and Slack health was OK. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9efa3b1f8b79c3094352fadf14bd3e18` `source_revision_id=srcrev_810a8840b5b19cdec7019493bf44b00f` `chunk_id=srcchunk_62efb2b04e96b567469f320b8cccbbaf` `native_locator=slack:C0547N89JUB:1781087072.163719:1781087152.134929` `source_timestamp=2026-06-10T10:25:52Z`
- The restart experienced repeated config recovery failures with error EROFS: read-only file system when copying /home/openclaw/.openclaw/openclaw.json.last-good to openclaw.json. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9efa3b1f8b79c3094352fadf14bd3e18` `source_revision_id=srcrev_810a8840b5b19cdec7019493bf44b00f` `chunk_id=srcchunk_62efb2b04e96b567469f320b8cccbbaf` `native_locator=slack:C0547N89JUB:1781087072.163719:1781087152.134929` `source_timestamp=2026-06-10T10:25:52Z`
- A security audit flagged an open Slack group policy with elevated/runtime/filesystem tools as CRITICAL, and recommended tightening group policy and tool exposure. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9efa3b1f8b79c3094352fadf14bd3e18` `source_revision_id=srcrev_810a8840b5b19cdec7019493bf44b00f` `chunk_id=srcchunk_62efb2b04e96b567469f320b8cccbbaf` `native_locator=slack:C0547N89JUB:1781087072.163719:1781087152.134929` `source_timestamp=2026-06-10T10:25:52Z`
- There was a Composio auth failure and one liveness warning, but Slack itself was functional. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_9efa3b1f8b79c3094352fadf14bd3e18` `source_revision_id=srcrev_810a8840b5b19cdec7019493bf44b00f` `chunk_id=srcchunk_62efb2b04e96b567469f320b8cccbbaf` `native_locator=slack:C0547N89JUB:1781087072.163719:1781087152.134929` `source_timestamp=2026-06-10T10:25:52Z`

## Open Questions

- How should the group policy and tool exposure be tightened to address the CRITICAL security audit?
- Is the bot fully safe for production given the identified issues?
- What caused the read-only file system error during config recovery?
- What is the impact of the Composio auth failure and how should it be resolved?

## Sources

- `source_document_id`: `srcdoc_9efa3b1f8b79c3094352fadf14bd3e18`
- `source_revision_id`: `srcrev_810a8840b5b19cdec7019493bf44b00f`
