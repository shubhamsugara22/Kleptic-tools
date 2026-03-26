# Day 3 Baseline - Secrets + Policy Guardrails

This Day 3 baseline adds:
- External Secrets Operator installation (for secret sync)
- Kyverno installation (for admission policy)
- Initial Kyverno policy pack (labels, resource requests/limits, no privileged containers)
- Dev secrets demo app using `SecretStore` and `ExternalSecret`

## Files

### Bootstrap
- `../day1/apps/children/day3-bootstrap.yaml`

### Day 3 Argo CD apps
- `apps/external-secrets-operator.yaml`
- `apps/kyverno-operator.yaml`
- `apps/kyverno-policy-pack.yaml`
- `apps/dev-secrets-demo.yaml`

### Policies
- `policies/require-labels.yaml`
- `policies/require-requests-limits.yaml`
- `policies/disallow-privileged-containers.yaml`

### Secrets demo
- `secrets/dev/secretstore.yaml`
- `secrets/dev/externalsecret.yaml`

## How It Gets Applied

If Day 1 root app is active, this is auto-bootstrapped by:
- `day3-bootstrap` child application

Optional direct apply:

```bash
kubectl apply -f argo/platform/day3/apps/external-secrets-operator.yaml
kubectl apply -f argo/platform/day3/apps/kyverno-operator.yaml
kubectl apply -f argo/platform/day3/apps/kyverno-policy-pack.yaml
kubectl apply -f argo/platform/day3/apps/dev-secrets-demo.yaml
```

## Verify

```bash
kubectl get applications -n argocd | grep -E 'day3|kyverno|external-secrets|dev-secrets-demo'
kubectl get ns kyverno external-secrets-system
kubectl get cpol
kubectl get externalsecret -n dev-sample
kubectl get secret demo-app-credentials -n dev-sample
```

## Notes

- `dev-secrets-demo` uses Kubernetes provider-based `SecretStore` as a bootstrap example.
- You must create a matching source secret (`demo-app-shared`) in `shared-secrets` namespace and grant read permissions to `external-secrets-reader`.
- Replace this with your real backend (Vault, AWS Secrets Manager, Azure Key Vault, etc.) in production.
