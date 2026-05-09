---
title: "Our K8S Clusters"
type: "system"
slug: "systems/our-k8s-clusters"
freshness: "2023-11-28T01:59:00Z"
tags:
  - "eks"
  - "infrastructure"
  - "kubernetes"
owners: []
source_revision_ids:
  - "srcrev_264401d142978c01f62c06f79977e091"
conflict_state: "none"
---

# Our K8S Clusters

## Summary

Details of our EKS (K8S) clusters including staging and production environments.

## Claims

- There is a current staging cluster (#243963068353/us-east-2/stag-story-eks-LzD98NGd) used for upcoming Hackathon, with API connected, Prometheus monitoring, GrafanaLab, rules on HTTP errors, application level monitoring, error logs (possibly Fluentd DaemonSet), and API uptime robot. `claim:claim_k8s_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Our-K8S-Clusters-253810bc21c849d6a13265c4e7c02dd1) `source_document_id=srcdoc_c107a922aec6f0290e3b46c3f50be5bd` `source_revision_id=srcrev_264401d142978c01f62c06f79977e091` `chunk_id=srcchunk_45e831d79dc0be2c9ce71436266c2ac4` `native_locator=https://www.notion.so/Our-K8S-Clusters-253810bc21c849d6a13265c4e7c02dd1` `source_timestamp=2023-11-28T01:59:00Z`
- There is a current production cluster (#243963068353/us-east-1/stag-story-eks-LzD98NGd) that is not going to be used. `claim:claim_k8s_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Our-K8S-Clusters-253810bc21c849d6a13265c4e7c02dd1) `source_document_id=srcdoc_c107a922aec6f0290e3b46c3f50be5bd` `source_revision_id=srcrev_264401d142978c01f62c06f79977e091` `chunk_id=srcchunk_45e831d79dc0be2c9ce71436266c2ac4` `native_locator=https://www.notion.so/Our-K8S-Clusters-253810bc21c849d6a13265c4e7c02dd1` `source_timestamp=2023-11-28T01:59:00Z`
- There is a new staging cluster (#478656756051/us-west-2/eks-H2hOTBd3) with HTTP ingress (no HTTPS yet), no ArgoCD (manually deploy at the moment), and permissions and monitoring to be set up. `claim:claim_k8s_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Our-K8S-Clusters-253810bc21c849d6a13265c4e7c02dd1) `source_document_id=srcdoc_c107a922aec6f0290e3b46c3f50be5bd` `source_revision_id=srcrev_264401d142978c01f62c06f79977e091` `chunk_id=srcchunk_45e831d79dc0be2c9ce71436266c2ac4` `native_locator=https://www.notion.so/Our-K8S-Clusters-253810bc21c849d6a13265c4e7c02dd1` `source_timestamp=2023-11-28T01:59:00Z`

## Open Questions

- What is the 'Please see' page? (https://www.notion.so/848ef57add7647878823caf040aa535d)
- What is the pre-requisite page mentioned? (https://www.notion.so/04c9bd60cb144018987991f681148efa)

## Related Pages

- `concepts/devex`

## Sources

- `source_document_id`: `srcdoc_c107a922aec6f0290e3b46c3f50be5bd`
- `source_revision_id`: `srcrev_264401d142978c01f62c06f79977e091`
- `source_url`: [Notion source](https://www.notion.so/Our-K8S-Clusters-253810bc21c849d6a13265c4e7c02dd1)
