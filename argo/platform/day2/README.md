# Day 2 Baseline - Environment Fan-out with ApplicationSet

This Day 2 baseline adds one ApplicationSet that generates Argo CD Applications for:
- `dev` (auto-sync enabled)
- `stage` (manual sync)
- `prod` (manual sync)

## Files

- `apps/env-applicationset.yaml`: Single ApplicationSet template with env-specific values
- `workloads/demo/dev/configmap.yaml`: Dev workload sample
- `workloads/demo/stage/configmap.yaml`: Stage workload sample
- `workloads/demo/prod/configmap.yaml`: Prod workload sample

## How It Gets Bootstrapped

Day 1 root app now includes:
- `../day1/apps/children/day2-bootstrap.yaml`

That child app syncs `argo/platform/day2/apps`, so the ApplicationSet is managed automatically.

## First-Time Apply

If Day 1 is already running and synced, commit/push and Argo CD will pick this up.

If you want to apply directly for first test:

```bash
kubectl apply -f argo/platform/day2/apps/env-applicationset.yaml
```

## Verify

```bash
kubectl get applicationsets -n argocd
argocd app list | grep demo-
argocd app get demo-dev
argocd app get demo-stage
argocd app get demo-prod
```

Expected behavior:
- `demo-dev` auto-syncs (`prune` and `selfHeal` on)
- `demo-stage` requires manual sync
- `demo-prod` requires manual sync

## Customize Next

- Replace demo workload paths with your real app manifests.
- Add more environments or clusters by extending the list generator.
- Move to matrix or cluster generators in Week 4 for multi-cluster fan-out.
