# ArgoCD Quick Reference & Cheat Sheet

## Installation

```bash
# Install ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Port forward to access UI
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

## CLI Login

```bash
# Login to ArgoCD
argocd login <ARGOCD_SERVER> --username admin --password <PASSWORD>

# Login with port-forward
argocd login localhost:8080 --username admin --password <PASSWORD> --insecure

# Change password
argocd account update-password
```

## Application Management

### Create Application

```bash
# Via CLI
argocd app create <APP_NAME> \
  --repo https://github.com/org/repo.git \
  --path ./manifests \
  --dest-server https://kubernetes.default.svc \
  --dest-namespace default

# Helm chart
argocd app create <APP_NAME> \
  --repo https://charts.example.com \
  --helm-chart <CHART_NAME> \
  --revision <VERSION> \
  --dest-server https://kubernetes.default.svc \
  --dest-namespace default

# Via manifest
kubectl apply -f application.yaml
```

### Application Operations

```bash
# List applications
argocd app list

# Get app details
argocd app get <APP_NAME>

# Get app manifest
argocd app manifests <APP_NAME>

# Get app history
argocd app history <APP_NAME>

# Sync application
argocd app sync <APP_NAME>

# Force sync (ignore sync window)
argocd app sync <APP_NAME> --force

# Sync specific resource
argocd app sync <APP_NAME> --resource <GROUP>/<KIND>/<NAME>

# Diff application
argocd app diff <APP_NAME>

# Rollback to previous revision
argocd app rollback <APP_NAME> <REVISION>

# Refresh application (hard refresh)
argocd app get <APP_NAME> --refresh

# Delete application
argocd app delete <APP_NAME>

# Delete with cascade (delete k8s resources)
argocd app delete <APP_NAME> --cascade

# Set auto-sync
argocd app set <APP_NAME> --sync-policy automated

# Disable auto-sync
argocd app set <APP_NAME> --sync-policy none

# Set sync option
argocd app set <APP_NAME> --sync-option CreateNamespace=true

# Set parameter
argocd app set <APP_NAME> -p key=value

# Watch sync status
argocd app wait <APP_NAME> --sync

# Terminate sync
argocd app terminate-op <APP_NAME>
```

### Application Filters

```bash
# List apps in project
argocd app list --project <PROJECT_NAME>

# List apps in namespace
argocd app list --app-namespace <NAMESPACE>

# List apps by label
argocd app list -l environment=production

# List out-of-sync apps
argocd app list --sync-status OutOfSync
```

## Repository Management

```bash
# List repositories
argocd repo list

# Add HTTPS repository
argocd repo add https://github.com/org/repo.git \
  --username <USER> \
  --password <TOKEN>

# Add SSH repository
argocd repo add git@github.com:org/repo.git \
  --ssh-private-key-path ~/.ssh/id_rsa

# Add Helm repository
argocd repo add https://charts.example.com \
  --type helm \
  --name stable

# Remove repository
argocd repo rm https://github.com/org/repo.git
```

## Cluster Management

```bash
# List clusters
argocd cluster list

# Add cluster from kubeconfig
argocd cluster add <CONTEXT_NAME>

# Add cluster with name
argocd cluster add <CONTEXT_NAME> --name production

# Remove cluster
argocd cluster rm <SERVER_URL>

# Get cluster info
argocd cluster get <SERVER_URL>
```

## Project Management

```bash
# List projects
argocd proj list

# Create project
argocd proj create <PROJECT_NAME>

# Add source repository to project
argocd proj add-source <PROJECT_NAME> https://github.com/org/repo.git

# Add destination to project
argocd proj add-destination <PROJECT_NAME> \
  <SERVER_URL> \
  <NAMESPACE>

# Allow cluster resource
argocd proj allow-cluster-resource <PROJECT_NAME> \
  <GROUP> \
  <KIND>

# Allow namespace resource
argocd proj allow-namespace-resource <PROJECT_NAME> \
  <GROUP> \
  <KIND>

# Delete project
argocd proj delete <PROJECT_NAME>
```

## Account & RBAC

```bash
# List accounts
argocd account list

# Get account details
argocd account get <ACCOUNT_NAME>

# Update password
argocd account update-password \
  --account <ACCOUNT_NAME> \
  --new-password <NEW_PASSWORD>

# Generate token
argocd account generate-token \
  --account <ACCOUNT_NAME>

# Can I operations
argocd account can-i sync applications '*/*'
argocd account can-i create projects
```

## Settings & Configuration

```bash
# Get server version
argocd version

# Get cluster info
argocd cluster list

# Get server settings
argocd settings get

# Update repo credentials
argocd repocreds add https://github.com/org/ \
  --username <USER> \
  --password <TOKEN>

# List repo credentials templates
argocd repocreds list
```

## Notifications

```bash
# Trigger notification
argocd app set <APP_NAME> \
  --annotation notifications.argoproj.io/subscribe.on-sync-succeeded.slack=channel-name
```

## Advanced Operations

```bash
# Patch application
argocd app patch <APP_NAME> --patch '{"spec":{"syncPolicy":{"automated":{"prune":true}}}}'

# Set label
argocd app set <APP_NAME> --label environment=production

# Unset label
argocd app unset <APP_NAME> --label environment

# Resource ignoring
argocd app set <APP_NAME> --ignore-missing-value-files
```

## Kubectl Commands

```bash
# Get all applications
kubectl get applications -n argocd

# Get application details
kubectl describe application <APP_NAME> -n argocd

# Get application YAML
kubectl get application <APP_NAME> -n argocd -o yaml

# Edit application
kubectl edit application <APP_NAME> -n argocd

# Delete application
kubectl delete application <APP_NAME> -n argocd

# Get AppProjects
kubectl get appprojects -n argocd

# Get ApplicationSets
kubectl get applicationsets -n argocd
```

## Common Application Manifest Snippets

### Auto-sync with prune and self-heal

```yaml
syncPolicy:
  automated:
    prune: true
    selfHeal: true
  syncOptions:
  - CreateNamespace=true
```

### Manual sync with retry

```yaml
syncPolicy:
  syncOptions:
  - CreateNamespace=true
  retry:
    limit: 5
    backoff:
      duration: 5s
      factor: 2
      maxDuration: 3m
```

### Ignore differences

```yaml
ignoreDifferences:
- group: apps
  kind: Deployment
  jsonPointers:
  - /spec/replicas
```

### Sync waves

```yaml
metadata:
  annotations:
    argocd.argoproj.io/sync-wave: "1"
```

### Resource hooks

```yaml
metadata:
  annotations:
    argocd.argoproj.io/hook: PreSync
    argocd.argoproj.io/hook-delete-policy: HookSucceeded
```

## Troubleshooting

```bash
# Get detailed app status
argocd app get <APP_NAME> --show-operation --show-params

# View app logs
kubectl logs -n argocd deployment/argocd-server -f

# View application controller logs
kubectl logs -n argocd deployment/argocd-application-controller -f

# View repo server logs
kubectl logs -n argocd deployment/argocd-repo-server -f

# Get sync operation details
kubectl get application <APP_NAME> -n argocd -o jsonpath='{.status.operationState}'

# Force refresh cache
argocd app get <APP_NAME> --hard-refresh

# Debug manifest generation
argocd app manifests <APP_NAME> --source live

# Restart ArgoCD components
kubectl rollout restart deployment -n argocd argocd-server
kubectl rollout restart deployment -n argocd argocd-application-controller
kubectl rollout restart deployment -n argocd argocd-repo-server
```

## Environment Variables

```bash
# Set ArgoCD server
export ARGOCD_SERVER=argocd.example.com

# Set auth token
export ARGOCD_AUTH_TOKEN=<TOKEN>

# Insecure
export ARGOCD_OPTS='--insecure'

# Enable gRPC
export ARGOCD_OPTS='--grpc-web'
```

## Useful Aliases

```bash
# Add to ~/.bashrc or ~/.zshrc
alias argocd-login='argocd login localhost:8080 --insecure'
alias argocd-apps='argocd app list'
alias argocd-sync='argocd app sync'
alias argocd-get='argocd app get'
alias argocd-diff='argocd app diff'
alias argocd-watch='watch -n 2 argocd app list'
```

## Full Application Example

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
  finalizers:
  - resources-finalizer.argocd.argoproj.io
  labels:
    team: platform
  annotations:
    notifications.argoproj.io/subscribe.on-sync-succeeded.slack: team-channel
spec:
  project: default
  source:
    repoURL: https://github.com/org/repo.git
    targetRevision: HEAD
    path: k8s
    helm:
      valueFiles:
      - values.yaml
      parameters:
      - name: image.tag
        value: v1.0.0
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
  ignoreDifferences:
  - group: apps
    kind: Deployment
    jsonPointers:
    - /spec/replicas
  revisionHistoryLimit: 10
```

## Links

- [Official Documentation](https://argo-cd.readthedocs.io/)
- [GitHub Repository](https://github.com/argoproj/argo-cd)
- [Example Apps](https://github.com/argoproj/argocd-example-apps)
- [Slack Community](https://argoproj.github.io/community/join-slack)
