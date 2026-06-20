---
title: "Diagnosing 400 Errors on Avatar Upload for Staging Depin Backend"
type: "runbook"
slug: "runbooks/staging-depin-avatar-upload-400"
freshness: "2026-04-08T23:24:05Z"
tags:
  - "aws"
  - "depin"
  - "staging"
  - "troubleshooting"
  - "uploads"
owners:
  - "Aiwei"
  - "Woojin"
source_revision_ids:
  - "srcrev_41c2875f5c1c5e498dc82363ff267eae"
  - "srcrev_56b3d8997729b9a03cdaa22181173e1d"
  - "srcrev_6f84316f28a6e9a7c5754596468c40d1"
  - "srcrev_724a8b11da6441ffa342dd131aeb8366"
  - "srcrev_871b2447da03c6c932f04dd04f1ca1cb"
  - "srcrev_88449465b8bd42815da750e5939e1521"
  - "srcrev_a0e9a7c0da11516913a65bea51643c24"
conflict_state: "none"
---

# Diagnosing 400 Errors on Avatar Upload for Staging Depin Backend

## Summary

A 5MB avatar image upload to staging-depin.storyprotocol.net /v1/me/complete-intro returns HTTP 400 with Starlette error "Error parsing multipart/form-data". Investigation reveals the ingress uses AWS ALB (not nginx), no app-side body size limits, and infra limits are not the cause. Remaining suspects include process-level gunicorn flags, python-multipart version, ALB idle timeout, or client-side transfer. Resolution is pending further inspection of pod startup commands.

## Claims

- The staging-depin ingress is an AWS Application Load Balancer, not nginx, so nginx-specific annotations like proxy-body-size have no effect. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_41d84aca1848e89d7ea308d4e8889817` `source_revision_id=srcrev_871b2447da03c6c932f04dd04f1ca1cb` `chunk_id=srcchunk_d8e446e892b402946a5f90923210d911` `native_locator=slack:C0547N89JUB:1775675483.744039:1775675652.584859` `source_timestamp=2026-04-08T19:14:12Z`
- The depin-backend application (FastAPI) has no body size limits configured in FastAPI, Starlette, gunicorn, or uvicorn config files. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_41d84aca1848e89d7ea308d4e8889817` `source_revision_id=srcrev_41c2875f5c1c5e498dc82363ff267eae` `chunk_id=srcchunk_5626a63dab0abddf1f9460a4c9e835e4` `native_locator=slack:C0547N89JUB:1775675483.744039:1775675798.956049` `source_timestamp=2026-04-08T19:16:38Z`
- Starlette's multipart parser returns "Error parsing multipart/form-data", indicating the request body is truncated or corrupted before reaching the app. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_41d84aca1848e89d7ea308d4e8889817` `source_revision_id=srcrev_41c2875f5c1c5e498dc82363ff267eae` `chunk_id=srcchunk_5626a63dab0abddf1f9460a4c9e835e4` `native_locator=slack:C0547N89JUB:1775675483.744039:1775675798.956049` `source_timestamp=2026-04-08T19:16:38Z`
- The problematic file size is approximately 5MB, ruling out WAF body inspection limits (max 64KB) as the direct cause, but leaving explicit WAF size constraint rules as a possibility (later checked and ruled out). `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_41d84aca1848e89d7ea308d4e8889817` `source_revision_id=srcrev_88449465b8bd42815da750e5939e1521` `chunk_id=srcchunk_d16bc3aa1ea9ffbfa3f81cf57b114fa5` `native_locator=slack:C0547N89JUB:1775675483.744039:1775675897.576699` `source_timestamp=2026-04-08T19:18:17Z`
  - citation: `source_document_id=srcdoc_41d84aca1848e89d7ea308d4e8889817` `source_revision_id=srcrev_a0e9a7c0da11516913a65bea51643c24` `chunk_id=srcchunk_4664f792c09a0bd35aa9043bca885072` `native_locator=slack:C0547N89JUB:1775675483.744039:1775675918.758109` `source_timestamp=2026-04-08T19:18:38Z`
- Subsequent investigation by Woojin confirmed no blocking at the ALB or WAF level, and Cloudflare limits are sufficient, eliminating major infra-level body size limits. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_41d84aca1848e89d7ea308d4e8889817` `source_revision_id=srcrev_56b3d8997729b9a03cdaa22181173e1d` `chunk_id=srcchunk_e7849c6748dbdf50e9dfd9cc1eda0457` `native_locator=slack:C0547N89JUB:1775675483.744039:1775690268.767569` `source_timestamp=2026-04-08T23:17:48Z`
- The remaining suspects are process-level startup flags (e.g., gunicorn --limit-request-body not in config files but possibly in pod command args), outdated python-multipart library version, ALB idle timeout causing mid-transfer drops, or client-side transfer issues. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_41d84aca1848e89d7ea308d4e8889817` `source_revision_id=srcrev_56b3d8997729b9a03cdaa22181173e1d` `chunk_id=srcchunk_e7849c6748dbdf50e9dfd9cc1eda0457` `native_locator=slack:C0547N89JUB:1775675483.744039:1775690268.767569` `source_timestamp=2026-04-08T23:17:48Z`
- The helm chart for depin-backend is expected in piplabs/helm-charts or the app repository, but access to kubectl and GitHub was temporarily restricted for some team members during the investigation. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_41d84aca1848e89d7ea308d4e8889817` `source_revision_id=srcrev_6f84316f28a6e9a7c5754596468c40d1` `chunk_id=srcchunk_bce5f7525598760b25a0299290ced076` `native_locator=slack:C0547N89JUB:1775675483.744039:1775677447.645359` `source_timestamp=2026-04-08T19:44:07Z`
  - citation: `source_document_id=srcdoc_41d84aca1848e89d7ea308d4e8889817` `source_revision_id=srcrev_724a8b11da6441ffa342dd131aeb8366` `chunk_id=srcchunk_cb5e2c6e5356d29cd0a5a33e5baea02c` `native_locator=slack:C0547N89JUB:1775675483.744039:1775690645.671129` `source_timestamp=2026-04-08T23:24:05Z`

## Open Questions

- What is the root cause of the 400 error? Is it a gunicorn startup flag, python-multipart version, ALB idle timeout, or client-side issue?

## Sources

- `source_document_id`: `srcdoc_41d84aca1848e89d7ea308d4e8889817`
- `source_revision_id`: `srcrev_724a8b11da6441ffa342dd131aeb8366`
