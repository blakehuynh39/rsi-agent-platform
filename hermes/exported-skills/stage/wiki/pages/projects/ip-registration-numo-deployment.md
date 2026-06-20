---
title: "IP Registration Numo Cluster Deployment"
type: "project"
slug: "projects/ip-registration-numo-deployment"
freshness: "2026-04-16T01:34:04Z"
tags:
  - "deployment"
  - "grafana"
  - "ip-registration"
  - "monitoring"
  - "numo"
  - "story-deployments"
owners:
  - "U07TNT9N4JC"
  - "U083MMT1771"
  - "U0AKJV8710S"
source_revision_ids:
  - "srcrev_23bcf3f647a3b71801e2019c274d8be4"
  - "srcrev_3bb704656c8d166ab838dc799308755a"
  - "srcrev_4a23750a81488bc6885cf3e0b19dbebf"
  - "srcrev_65f60502e56f828146842106d7a590fa"
  - "srcrev_687a104a8c22a0b40915b5757c2a76a1"
  - "srcrev_8e614e7417abf9ec7ba47745a5550cdf"
  - "srcrev_91396f9db4857afa85bca932eca70601"
  - "srcrev_9bd097c80b4fd04b31188e9bf5bba0b3"
  - "srcrev_ce38419b7c5be2c92183786b46fdd1f6"
  - "srcrev_e3466d83bb7ac3e45996700773b9f3f1"
  - "srcrev_eace1f02bc73470917cef696a6582cd0"
  - "srcrev_fb8be2b4b62b1a204443fbc2e6e02c84"
conflict_state: "none"
---

# IP Registration Numo Cluster Deployment

## Summary

Deploying the IP Registration service to the numo Kubernetes cluster and setting up monitoring with Grafana.

## Claims

- Aiwei created PR #142 in story-helm to add a numo cluster definition for IP registration, along with a Grafana dashboard to track gas and wallet balances. `claim:claim_1_1` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bd69e26554aa3a3eba29622ae2d03627` `source_revision_id=srcrev_3bb704656c8d166ab838dc799308755a` `chunk_id=srcchunk_e8fbd03702d5305b964ecbba23af79e4` `native_locator=slack:C0547N89JUB:1776275081.430369:1776275081.430369` `source_timestamp=2026-04-15T17:44:41Z`
- Review of PR #142 identified five issues: raw IP for clusterEndpoint, image tag TODO, Servicemonitor targetPort as int, vault references handling, and serviceAccount configuration. `claim:claim_1_2` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bd69e26554aa3a3eba29622ae2d03627` `source_revision_id=srcrev_e3466d83bb7ac3e45996700773b9f3f1` `chunk_id=srcchunk_9717e9aa0ca0be18d152945be202aa70` `native_locator=slack:C0547N89JUB:1776275081.430369:1776275171.116799` `source_timestamp=2026-04-15T17:46:11Z`
- All five issues in PR #142 were addressed and the PR was approved, but it was noted that story-helm is the legacy repository and the deployment should be moved to story-deployments. `claim:claim_1_3` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bd69e26554aa3a3eba29622ae2d03627` `source_revision_id=srcrev_687a104a8c22a0b40915b5757c2a76a1` `chunk_id=srcchunk_5469da5a8d05e9adc3b67f8913d5e6bb` `native_locator=slack:C0547N89JUB:1776275081.430369:1776275441.610289` `source_timestamp=2026-04-15T17:50:41Z`
  - citation: `source_document_id=srcdoc_bd69e26554aa3a3eba29622ae2d03627` `source_revision_id=srcrev_fb8be2b4b62b1a204443fbc2e6e02c84` `chunk_id=srcchunk_f469047260b00169f6a376d36300eac8` `native_locator=slack:C0547N89JUB:1776275081.430369:1776295481.816509` `source_timestamp=2026-04-15T23:24:41Z`
- The ECR repository for IP registration is in the stage AWS account (783268398689), and image tags available include 9132a28d17dd5c6c1657f5de5d66d4100816590b and dfc1675d504256e92d0a3f13912f3c4197d55915. `claim:claim_1_4` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bd69e26554aa3a3eba29622ae2d03627` `source_revision_id=srcrev_ce38419b7c5be2c92183786b46fdd1f6` `chunk_id=srcchunk_f719d9ffbb349b17c215fc7313c76e90` `native_locator=slack:C0547N89JUB:1776275081.430369:1776295570.119079` `source_timestamp=2026-04-15T23:26:10Z`
- CI/CD for IP registration already automatically bumps the image tag in story-deployments on every push to staging; the current deployed tag is 65699f0ba3e8cee88a22c2a3096afc20126eea50. `claim:claim_1_5` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bd69e26554aa3a3eba29622ae2d03627` `source_revision_id=srcrev_8e614e7417abf9ec7ba47745a5550cdf` `chunk_id=srcchunk_f7c32bc3449546a1c44f51e62404f852` `native_locator=slack:C0547N89JUB:1776275081.430369:1776295596.275669` `source_timestamp=2026-04-15T23:26:36Z`
- Aiwei (U083MMT1771) was granted write access to story-deployments by Woojin (U07TNT9N4JC) and opened PR #105 for the numo deployment. `claim:claim_1_6` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bd69e26554aa3a3eba29622ae2d03627` `source_revision_id=srcrev_4a23750a81488bc6885cf3e0b19dbebf` `chunk_id=srcchunk_53241a2003caa865693d80326d50e27b` `native_locator=slack:C0547N89JUB:1776275081.430369:1776298246.162839` `source_timestamp=2026-04-16T00:10:46Z`
  - citation: `source_document_id=srcdoc_bd69e26554aa3a3eba29622ae2d03627` `source_revision_id=srcrev_eace1f02bc73470917cef696a6582cd0` `chunk_id=srcchunk_f70e498905768961cf76871ab3515d3d` `native_locator=slack:C0547N89JUB:1776275081.430369:1776298528.177499` `source_timestamp=2026-04-16T00:18:08Z`
  - citation: `source_document_id=srcdoc_bd69e26554aa3a3eba29622ae2d03627` `source_revision_id=srcrev_9bd097c80b4fd04b31188e9bf5bba0b3` `chunk_id=srcchunk_36a9968f6749a2f7adc784e94738aeb7` `native_locator=slack:C0547N89JUB:1776275081.430369:1776298875.259159` `source_timestamp=2026-04-16T00:21:15Z`
- PR #105 is blocked on several infrastructure prerequisites: EKS cluster endpoint, IRSA role, Vault paths, and remaining FIXMEs for contract addresses, S3 buckets, and public URLs. `claim:claim_1_7` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bd69e26554aa3a3eba29622ae2d03627` `source_revision_id=srcrev_65f60502e56f828146842106d7a590fa` `chunk_id=srcchunk_b4a3a0764a0a90db87dccb50cf5dda2d` `native_locator=slack:C0547N89JUB:1776275081.430369:1776298954.237639` `source_timestamp=2026-04-16T00:22:34Z`
- Aiwei requested review from U0AKJV8710S and U08V4SFU7LZ for chain RPC and IP registration values to fill remaining placeholders. `claim:claim_1_8` `confidence:1.00`
  - citation: `source_document_id=srcdoc_bd69e26554aa3a3eba29622ae2d03627` `source_revision_id=srcrev_91396f9db4857afa85bca932eca70601` `chunk_id=srcchunk_895770af17ed68b0aa5443db1afe43c9` `native_locator=slack:C0547N89JUB:1776275081.430369:1776298911.403439` `source_timestamp=2026-04-16T00:21:51Z`
  - citation: `source_document_id=srcdoc_bd69e26554aa3a3eba29622ae2d03627` `source_revision_id=srcrev_23bcf3f647a3b71801e2019c274d8be4` `chunk_id=srcchunk_2f0c7a8f30fb0ab7223469eb1becf544` `native_locator=slack:C0547N89JUB:1776275081.430369:1776303244.433209` `source_timestamp=2026-04-16T01:34:04Z`

## Open Questions

- EKS cluster endpoint for numo is missing (FIXME in applicationset/numo.yaml). Needs to be provided before PR #105 can merge.
- FIXMEs in the values file (contract addresses, S3 buckets, public URLs) need actual values. Chain RPC and IP registration related values require input from U08V4SFU7LZ.
- IRSA role for numo must be confirmed and created in Terraform; verify numo is in stage account 783268398689.
- Vault paths numo/depin-backend and numo/depin-ip-registration must be seeded before deployment.

## Sources

- `source_document_id`: `srcdoc_bd69e26554aa3a3eba29622ae2d03627`
- `source_revision_id`: `srcrev_e9ab3cc9fe8ea6a6e5ec9ba235e48dc2`
