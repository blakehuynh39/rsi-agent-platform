# Depin K8s Internal Service DNS

Internal DNS follows standard Kubernetes convention: `<service-name>.<namespace>.svc.cluster.local`. Short form `<service-name>.<namespace>` also resolves within cluster.

## Staging (cluster: use1-stage, namespace: story)

| Service | DNS | Port | Replicas |
|---|---|---|---|
| `use1-stage-depin-backend` | `use1-stage-depin-backend.story.svc.cluster.local` | 8080 | 2 |
| `use1-stage-depin-ip-registration-poller` | `use1-stage-depin-ip-registration-poller.story.svc.cluster.local` | 9090 | 2 |
| `use1-stage-depin-ip-registration-submitter` | `use1-stage-depin-ip-registration-submitter.story.svc.cluster.local` | 9090 | 3 |
| `use1-stage-depin-ip-registration-confirmer` | `use1-stage-depin-ip-registration-confirmer.story.svc.cluster.local` | 9090 | 3 |

All `ClusterIP` type. All healthy with matching endpoints (verified 2026-05-12).

## Production (cluster: use1-prod, namespace: story)

**Note**: Production is a separate EKS cluster. `kubectl` on staging will NOT show these services.

| Service | DNS | Port | Replicas |
|---|---|---|---|
| `use1-prod-depin-backend` | `use1-prod-depin-backend.story.svc.cluster.local` | 8080 | 4-8 (HPA) |
| `use1-prod-depin-ip-registration-poller` | `use1-prod-depin-ip-registration-poller.story.svc.cluster.local` | 9090 | 2 |
| `use1-prod-depin-ip-registration-submitter` | `use1-prod-depin-ip-registration-submitter.story.svc.cluster.local` | 9090 | 3 |
| `use1-prod-depin-ip-registration-confirmer` | `use1-prod-depin-ip-registration-confirmer.story.svc.cluster.local` | 9090 | 3 |

Prod backend has AWS ALB ingress serving `depin.storyprotocol.net` and `api.numolabs.ai`. Submitter has IRSA role for Story chain txs.

## Helm Naming Convention

- Backend chart: `fullnameOverride` = exact service name (e.g., `use1-prod-depin-backend`)
- IP Registration chart: `fullnameOverride` sets prefix; each worker gets `{fullnameOverride}-{worker_name}` suffix
- Source: `story-deployments/story/depin-backend/use1-prod.yaml` and `story/depin-ip-registration/use1-prod.yaml`

## Cluster Endpoints

- **use1-stage**: current `kubectl` context (accessible)
- **use1-prod**: `https://6345B193081358905BE7E8B5A00C54DF.gr7.us-east-1.eks.amazonaws.com` (separate, no direct access)

ArgoCD ApplicationSet (`applicationset/templates/applicationset.yaml`) confirms namespace = top-level key under `elements` (e.g., `story`).

## Verification Commands

```bash
# Staging services (accessible)
kubectl get svc,endpoints -n story | grep depin

# Production services (NOT accessible from staging cluster)
# Reference story-deployments/story/depin-backend/use1-prod.yaml instead

# Verify DNS from within cluster (requires pods/exec RBAC — may be denied)
kubectl exec -n story deployment/use1-stage-depin-backend -- nslookup use1-stage-depin-backend.story
```
