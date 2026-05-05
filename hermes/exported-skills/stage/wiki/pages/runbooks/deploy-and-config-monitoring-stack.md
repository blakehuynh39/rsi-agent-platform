---
title: "Deploy and Config Monitoring Stack"
type: "runbook"
slug: "runbooks/deploy-and-config-monitoring-stack"
freshness: "2026-05-05T06:33:05Z"
tags:
  - "eks"
  - "grafana"
  - "kubernetes"
  - "monitoring"
  - "prometheus"
owners: []
source_revision_ids:
  - "srcrev_2329bc26ad9dab6d07a766e94f30c97b"
conflict_state: "none"
---

# Deploy and Config Monitoring Stack

## Summary

Runbook for deploying the Prometheus/Grafana monitoring stack on EKS using Helm, including namespace creation, secret setup, and Helm commands.

## Claims

- The monitoring stack is deployed in the 'monitoring' Kubernetes namespace. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196) `source_document_id=srcdoc_3704a5e562f18fd813262e6d39cc38c0` `source_revision_id=srcrev_2329bc26ad9dab6d07a766e94f30c97b` `chunk_id=srcchunk_0fcb1d718f182f5322dcb9fe9f1937bf` `native_locator=https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196` `source_timestamp=2026-05-05T06:33:05Z`
- A Kubernetes secret named 'kubepromsecret' is created in the monitoring namespace with a username and password for Grafana Cloud metrics writing. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196) `source_document_id=srcdoc_3704a5e562f18fd813262e6d39cc38c0` `source_revision_id=srcrev_2329bc26ad9dab6d07a766e94f30c97b` `chunk_id=srcchunk_0fcb1d718f182f5322dcb9fe9f1937bf` `native_locator=https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196` `source_timestamp=2026-05-05T06:33:05Z`
- The Prometheus stack is installed via Helm using the chart prometheus-community/kube-prometheus-stack with release name 'prom-prod-us-east-1' and values file 'values-story-prod-us-east-1-eks.yaml'. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196) `source_document_id=srcdoc_3704a5e562f18fd813262e6d39cc38c0` `source_revision_id=srcrev_2329bc26ad9dab6d07a766e94f30c97b` `chunk_id=srcchunk_0fcb1d718f182f5322dcb9fe9f1937bf` `native_locator=https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196` `source_timestamp=2026-05-05T06:33:05Z`
- An alternative deployment using the prometheus-community/prometheus chart is also documented, with the same release name and values file. `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196) `source_document_id=srcdoc_3704a5e562f18fd813262e6d39cc38c0` `source_revision_id=srcrev_2329bc26ad9dab6d07a766e94f30c97b` `chunk_id=srcchunk_0fcb1d718f182f5322dcb9fe9f1937bf` `native_locator=https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196` `source_timestamp=2026-05-05T06:33:05Z`
- The Helm release can be upgraded using 'helm upgrade prom-prod-us-east-1 prometheus-community/kube-prometheus-stack -n monitoring -f values-story-prod-us-east-1-eks.yaml'. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196) `source_document_id=srcdoc_3704a5e562f18fd813262e6d39cc38c0` `source_revision_id=srcrev_2329bc26ad9dab6d07a766e94f30c97b` `chunk_id=srcchunk_0fcb1d718f182f5322dcb9fe9f1937bf` `native_locator=https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196` `source_timestamp=2026-05-05T06:33:05Z`
- Prometheus UI can be accessed locally via port-forward on port 9090 using 'kubectl --namespace monitoring port-forward svc/prom-prod-us-east-1-kube-p-prometheus 9090'. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196) `source_document_id=srcdoc_3704a5e562f18fd813262e6d39cc38c0` `source_revision_id=srcrev_2329bc26ad9dab6d07a766e94f30c97b` `chunk_id=srcchunk_0fcb1d718f182f5322dcb9fe9f1937bf` `native_locator=https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196` `source_timestamp=2026-05-05T06:33:05Z`
- The Helm release can be uninstalled with 'helm uninstall prom-prod-us-east-1 -n monitoring'. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196) `source_document_id=srcdoc_3704a5e562f18fd813262e6d39cc38c0` `source_revision_id=srcrev_2329bc26ad9dab6d07a766e94f30c97b` `chunk_id=srcchunk_0fcb1d718f182f5322dcb9fe9f1937bf` `native_locator=https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196` `source_timestamp=2026-05-05T06:33:05Z`
- The deployment references Grafana Cloud documentation for AWS EKS configuration. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196) `source_document_id=srcdoc_3704a5e562f18fd813262e6d39cc38c0` `source_revision_id=srcrev_2329bc26ad9dab6d07a766e94f30c97b` `chunk_id=srcchunk_0fcb1d718f182f5322dcb9fe9f1937bf` `native_locator=https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196` `source_timestamp=2026-05-05T06:33:05Z`

## Sources

- `source_document_id`: `srcdoc_3704a5e562f18fd813262e6d39cc38c0`
- `source_revision_id`: `srcrev_2329bc26ad9dab6d07a766e94f30c97b`
- `source_url`: [Notion source](https://www.notion.so/KB-Deploy-and-Config-Monitoring-Stack-17b48b3f13104b1eb491b12c33828196)
