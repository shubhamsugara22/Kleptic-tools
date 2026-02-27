# ArgoCD - Declarative GitOps Continuous Deployment

ArgoCD is a declarative, GitOps continuous delivery tool for Kubernetes. It automates the deployment of applications to Kubernetes clusters by using Git repositories as the single source of truth.

## Quick Links

> 📋 **[INDEX.md](INDEX.md)** - Complete documentation index with learning paths, use case navigation, and quick search

### 📚 Documentation Guides
- 📖 **[CHEATSHEET.md](CHEATSHEET.md)** - Quick reference for all ArgoCD commands and operations
- 🏗️ **[PRODUCTION-GUIDE.md](PRODUCTION-GUIDE.md)** - High availability, security, monitoring, and disaster recovery
- 🔧 **[INTEGRATION-GUIDE.md](INTEGRATION-GUIDE.md)** - CI/CD pipelines, secrets management, image updater, and notifications
- 🐛 **[TROUBLESHOOTING-GUIDE.md](TROUBLESHOOTING-GUIDE.md)** - Common issues, debugging, performance tuning, and advanced patterns

### Setup Scripts
- [setup.sh](setup.sh) - Bash installation script for Linux/macOS
- [setup.ps1](setup.ps1) - PowerShell installation script for Windows

### Template Files
- [basic-app.yaml](basic-app.yaml) - Basic application template
- [helm-app.yaml](helm-app.yaml) - Helm chart application
- [kustomize-app.yaml](kustomize-app.yaml) - Kustomize overlay application
- [app-of-apps.yaml](app-of-apps.yaml) - App of Apps pattern
- [app-project.yaml](app-project.yaml) - AppProject with RBAC
- [applicationset.yaml](applicationset.yaml) - ApplicationSet for GitOps at scale
- [sync-waves-example.yaml](sync-waves-example.yaml) - Sync waves and hooks demonstration

## Table of Contents

- [Overview](#overview)
- [Key Concepts](#key-concepts)
- [ArgoCD Features We Use](#argocd-features-we-use)
- [Application Templates](#application-templates)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Basic Setup](#basic-setup)
- [Common Commands](#common-commands)
- [Architecture](#architecture)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

ArgoCD provides:
- **Declarative setup**: Define applications in Git
- **Automated sync**: Continuous deployment based on Git changes
- **Multi-cluster support**: Manage multiple Kubernetes clusters
- **UI Dashboard**: Visual application status monitoring
- **RBAC**: Fine-grained access control
- **Web hooks & SSO**: Enterprise integrations

## Key Concepts

### Application
A group of Kubernetes resources defined by a manifest in a Git repository.

### Project
A logical grouping of applications with restrictions on which namespaces and clusters can be deployed to.

### Repository
A Git repository containing application manifests.

### Cluster
A Kubernetes cluster where applications are deployed.

### Sync Status
- **Synced**: Application resources match Git manifest
- **Out of Sync**: Application differs from Git source
- **Unknown**: Sync status cannot be determined

### Sync Policy
- **Manual**: Requires manual sync to deploy
- **Automatic**: Auto-syncs when Git changes, with optional pruning and self-heal

## ArgoCD Features We Use

### 1. **Automated Sync**
Automatically deploys changes from Git to Kubernetes clusters without manual intervention.

```yaml
syncPolicy:
  automated:
    prune: true      # Auto-delete resources removed from Git
    selfHeal: true   # Revert manual changes to cluster
```

### 2. **Multi-Cluster Deployment**
Manage multiple Kubernetes clusters from a single ArgoCD instance.

```bash
# Add production cluster
argocd cluster add prod-cluster --name production

# Add staging cluster
argocd cluster add staging-cluster --name staging
```

### 3. **Application Health Monitoring**
Continuous health checks for deployed applications with visual status indicators.

- **Healthy**: All resources running as expected
- **Progressing**: Deployment in progress
- **Degraded**: Issues detected
- **Suspended**: Application paused
- **Missing**: Resources not found

### 4. **GitOps with Multiple Sources**
Support for various manifest formats:
- **Plain YAML/JSON**: Kubernetes manifests
- **Helm Charts**: Package manager for Kubernetes
- **Kustomize**: Template-free customization
- **Jsonnet**: Data templating language
- **Custom Config Management Plugins**

### 5. **App of Apps Pattern**
Manage multiple applications using a parent application.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: root-app
  namespace: argocd
spec:
  source:
    repoURL: https://github.com/org/apps.git
    path: applications/
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd
```

### 6. **Rollback & History**
Track deployment history and rollback to previous versions.

```bash
# View history
argocd app history example-app

# Rollback to specific revision
argocd app rollback example-app 5
```

### 7. **Resource Hooks**
Execute actions before/after sync operations.

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: pre-sync-migration
  annotations:
    argocd.argoproj.io/hook: PreSync
    argocd.argoproj.io/hook-delete-policy: HookSucceeded
spec:
  template:
    spec:
      containers:
      - name: migration
        image: migrations:latest
        command: ["migrate"]
      restartPolicy: Never
```

Available hooks:
- `PreSync`: Before sync operation
- `Sync`: During sync operation
- `PostSync`: After sync operation
- `SyncFail`: On sync failure
- `Skip`: Skip sync

### 8. **Sync Waves**
Control deployment order using annotations.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: database-config
  annotations:
    argocd.argoproj.io/sync-wave: "0"  # Deploy first
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  annotations:
    argocd.argoproj.io/sync-wave: "1"  # Deploy second
```

### 9. **Diff Customization**
Ignore specific fields during diff comparisons.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: example
spec:
  ignoreDifferences:
  - group: apps
    kind: Deployment
    jsonPointers:
    - /spec/replicas  # Ignore replica differences
```

### 10. **SSO Integration**
Integrate with identity providers (GitHub, GitLab, LDAP, OIDC).

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
data:
  dex.config: |
    connectors:
    - type: github
      id: github
      name: GitHub
      config:
        clientID: $GITHUB_CLIENT_ID
        clientSecret: $GITHUB_CLIENT_SECRET
        orgs:
        - name: my-org
```

### 11. **Progressive Delivery**
Integration with Argo Rollouts for canary/blue-green deployments.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: rollout-canary
spec:
  replicas: 5
  strategy:
    canary:
      steps:
      - setWeight: 20
      - pause: {duration: 1h}
      - setWeight: 40
      - pause: {duration: 1h}
      - setWeight: 60
      - pause: {duration: 1h}
```

### 12. **Secrets Management**
Integration with external secret managers.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
spec:
  source:
    helm:
      parameters:
      - name: db.password
        value: $ARGOCD_ENV_DB_PASSWORD  # From external source
```

## Application Templates

### 1. **Basic Application Template**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
  finalizers:
  - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: https://github.com/myorg/myapp.git
    targetRevision: HEAD
    path: k8s
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
```

### 2. **Helm Chart Application**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: nginx-helm
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://charts.bitnami.com/bitnami
    chart: nginx
    targetRevision: 13.2.0
    helm:
      releaseName: my-nginx
      parameters:
      - name: service.type
        value: LoadBalancer
      - name: replicaCount
        value: "3"
      valueFiles:
      - values-production.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: web
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
```

### 3. **Kustomize Application**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: kustomize-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/myorg/kustomize-app.git
    targetRevision: main
    path: overlays/production
    kustomize:
      namePrefix: prod-
      nameSuffix: -v1
      images:
      - myapp:1.2.3
      commonLabels:
        environment: production
        team: platform
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

### 4. **Multi-Source Application**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: multi-source-app
  namespace: argocd
spec:
  project: default
  sources:
  - repoURL: https://github.com/myorg/helm-charts.git
    path: charts/myapp
    targetRevision: main
    helm:
      valueFiles:
      - $values/values-prod.yaml
  - repoURL: https://github.com/myorg/config.git
    targetRevision: main
    ref: values
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

### 5. **App of Apps Template**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: app-of-apps
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/myorg/applications.git
    targetRevision: HEAD
    path: apps
    directory:
      recurse: true
      jsonnet: {}
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

### 6. **Environment-Specific Template**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: api-production
  namespace: argocd
  labels:
    environment: production
    team: backend
spec:
  project: production-apps
  source:
    repoURL: https://github.com/myorg/api.git
    targetRevision: v1.5.0  # Pinned version for production
    path: k8s/overlays/production
  destination:
    server: https://prod-cluster.example.com
    namespace: api-prod
  syncPolicy:
    automated:
      prune: false  # Manual pruning in production
      selfHeal: false  # Manual healing in production
    syncOptions:
    - CreateNamespace=true
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
```

### 7. **ApplicationSet Template (GitOps at Scale)**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: microservices
  namespace: argocd
spec:
  generators:
  - git:
      repoURL: https://github.com/myorg/microservices.git
      revision: HEAD
      directories:
      - path: services/*
  template:
    metadata:
      name: '{{path.basename}}'
    spec:
      project: default
      source:
        repoURL: https://github.com/myorg/microservices.git
        targetRevision: HEAD
        path: '{{path}}'
      destination:
        server: https://kubernetes.default.svc
        namespace: '{{path.basename}}'
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
        syncOptions:
        - CreateNamespace=true
```

### 8. **AppProject Template**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: team-platform
  namespace: argocd
spec:
  description: Platform team applications
  sourceRepos:
  - https://github.com/myorg/*
  - https://charts.helm.sh/stable
  destinations:
  - namespace: 'platform-*'
    server: https://kubernetes.default.svc
  - namespace: 'monitoring'
    server: https://kubernetes.default.svc
  clusterResourceWhitelist:
  - group: ''
    kind: Namespace
  - group: 'rbac.authorization.k8s.io'
    kind: ClusterRole
  namespaceResourceWhitelist:
  - group: '*'
    kind: '*'
  roles:
  - name: platform-admin
    description: Full access to platform apps
    policies:
    - p, proj:team-platform:platform-admin, applications, *, team-platform/*, allow
    groups:
    - platform-team
```

### 9. **Private Repository Application**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: private-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: git@github.com:myorg/private-repo.git
    targetRevision: main
    path: deployments
  destination:
    server: https://kubernetes.default.svc
    namespace: apps
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

**Note**: Configure SSH key or credentials first:
```bash
argocd repo add git@github.com:myorg/private-repo.git \
  --ssh-private-key-path ~/.ssh/argocd_deploy_key
```

### 10. **With Resource Tracking & Notifications**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: critical-service
  namespace: argocd
  annotations:
    notifications.argoproj.io/subscribe.on-sync-succeeded.slack: platform-team
    notifications.argoproj.io/subscribe.on-sync-failed.slack: platform-alerts
spec:
  project: default
  source:
    repoURL: https://github.com/myorg/critical-service.git
    targetRevision: HEAD
    path: manifests
  destination:
    server: https://kubernetes.default.svc
    namespace: critical
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
    - PrunePropagationPolicy=foreground
    - PruneLast=true
  revisionHistoryLimit: 10
```

## Prerequisites

### Required
- Kubernetes cluster 1.16+
- kubectl configured to access your cluster
- Git repository with Kubernetes manifests

### Optional
- Helm 3+ (for Helm chart support)
- Kustomize (for kustomization support)
- ArgoCD CLI

## Installation

### 1. Create Namespace

```bash
kubectl create namespace argocd
```

### 2. Install ArgoCD

```bash
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

### 3. Install ArgoCD CLI (Optional)

**On macOS:**
```bash
brew install argocd
```

**On Linux:**
```bash
curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
sudo install -m 555 argocd-linux-amd64 /usr/local/bin/argocd
rm argocd-linux-amd64
```

**On Windows (PowerShell):**
```powershell
$version=$(curl.exe -s https://api.github.com/repos/argoproj/argo-cd/releases/latest | grep tag_name | cut -d '"' -f 4)
$url = "https://github.com/argoproj/argo-cd/releases/download/$version/argocd-windows-amd64.exe"
$output = "$env:USERPROFILE\AppData\Local\Programs\argocd.exe"
Invoke-WebRequest -Uri $url -OutFile $output
$env:Path += ";$env:USERPROFILE\AppData\Local\Programs"
```

## Basic Setup

### 1. Access the ArgoCD UI

#### Port Forward
```bash
kubectl port-forward svc/argocd-server -n argocd 8080:443
```
Access: `https://localhost:8080`

#### Load Balancer (if configured)
```bash
kubectl patch svc argocd-server -n argocd -p '{"spec": {"type": "LoadBalancer"}}'
```

### 2. Get Initial Admin Password

```bash
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo
```

- **Default username**: `admin`
- **Default password**: (from command above)

### 3. Login via CLI

```bash
argocd login <ARGOCD_SERVER> --username admin --password <PASSWORD>
```

### 4. Change Admin Password (Recommended)

```bash
argocd account update-password --account admin --current-password <OLD_PASSWORD> --new-password <NEW_PASSWORD>
```

### 5. Add Git Repository

```bash
argocd repo add https://github.com/example/example-app.git \
  --username <USERNAME> \
  --password <PASSWORD>
```

For SSH repositories:
```bash
argocd repo add git@github.com:example/example-app.git \
  --ssh-private-key-path ~/.ssh/id_rsa
```

### 6. Add Kubernetes Cluster

```bash
argocd cluster add <KUBE_CONTEXT_NAME>
```

View available clusters after adding:
```bash
argocd cluster list
```

### 7. Create an Application

**Via CLI:**
```bash
argocd app create example-app \
  --repo https://github.com/example/example-app.git \
  --revision HEAD \
  --path ./ \
  --dest-server https://kubernetes.default.svc \
  --dest-namespace default
```

**Via Application Manifest (YAML):**

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: example-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/example/example-app.git
    targetRevision: HEAD
    path: ./
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
```

### 8. Sync Application

```bash
argocd app sync example-app
```

Monitor sync status:
```bash
argocd app get example-app
```

## Common Commands

### Application Management

```bash
# List all applications
argocd app list

# Get application status
argocd app get <APP_NAME>

# Create application
argocd app create <APP_NAME> \
  --repo <REPO_URL> \
  --path <PATH> \
  --dest-server <SERVER> \
  --dest-namespace <NAMESPACE>

# Delete application
argocd app delete <APP_NAME>

# Sync application
argocd app sync <APP_NAME>

# Rollback to previous sync
argocd app rollback <APP_NAME>

# Update application parameters
argocd app set <APP_NAME> -p key=value
```

### Repository Management

```bash
# List repositories
argocd repo list

# Add repository
argocd repo add <REPO_URL> --username <USER> --password <PASS>

# Remove repository
argocd repo remove <REPO_URL>
```

### Cluster Management

```bash
# List clusters
argocd cluster list

# Add cluster
argocd cluster add <CONTEXT_NAME>

# Remove cluster
argocd cluster rm <SERVER>
```

### Account & RBAC

```bash
# List accounts
argocd account list

# Create user
argocd account create <USERNAME>

# Set user password
argocd account update-password --account <USERNAME> --new-password <PASSWORD>

# Create API token
argocd account generate-token --account <USERNAME>
```

## Architecture

### Components

- **API Server**: REST API for UI, CLI, and webhooks
- **Repository Server**: Git repository poller and manifest generator
- **Application Controller**: Monitors applications and performs reconciliation
- **Server**: Handles UI requests and API server caching
- **Dex Server**: OpenID Connect provider for authentication

### Workflow

```
Git Repository → ArgoCD API → Repository Server → Application Controller → Kubernetes Cluster
     ↑                                                      ↓
     └──────────────────── Sync Status ──────────────────┘
```

## Best Practices

### 1. **Repository Structure**
```
example-repo/
├── base/
│   ├── kustomization.yaml
│   └── deployment.yaml
├── overlays/
│   ├── dev/
│   │   └── kustomization.yaml
│   ├── staging/
│   │   └── kustomization.yaml
│   └── prod/
│       └── kustomization.yaml
└── argocd/
    └── application.yaml
```

### 2. **Sync Policies**
```yaml
syncPolicy:
  automated:
    prune: true        # Delete resources not in Git
    selfHeal: true     # Sync when drift detected
  syncOptions:
    - CreateNamespace=true
```

### 3. **RBAC Configuration**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-rbac-cm
  namespace: argocd
data:
  policy.default: role:admin
  policy.csv: |
    p, role:admin, applications, *, */*, allow
    p, role:viewer, applications, get, */*, allow
    g, my-group, role:admin
```

### 4. **Notifications & Webhooks**
- Configure webhooks in Git repositories for push events
- Set up notifications for application sync events
- Use ArgoCD notification service for Slack, PagerDuty, etc.

### 5. **Multi-tenancy**
Use AppProjects to isolate teams:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: team-a
  namespace: argocd
spec:
  sourceRepos:
  - https://github.com/team-a/*
  destinations:
  - namespace: team-a-*
    server: https://kubernetes.default.svc
```

## Troubleshooting

### Check ArgoCD Status

```bash
# Check pod status
kubectl get pods -n argocd

# Check pod logs
kubectl logs -n argocd deployment/argocd-server

# Describe application
kubectl describe app <APP_NAME> -n argocd
```

### Common Issues

#### 1. **Application Out of Sync**
```bash
# Force sync
argocd app sync <APP_NAME> --force

# Check what's different
argocd app diff <APP_NAME>
```

#### 2. **Repository Connection Failed**
```bash
# Verify repository credentials
argocd repo list --repo-url <REPO_URL>

# Test connection
argocd repo create-from git@github.com:example/repo.git
```

#### 3. **Cannot Access UI**
```bash
# Check service
kubectl get svc -n argocd argocd-server

# Port forward
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

### Get Help

- **Documentation**: https://argo-cd.readthedocs.io/
- **GitHub Issues**: https://github.com/argoproj/argo-cd/issues
- **Community Slack**: https://argoproj.github.io/community/join-slack

## Useful Links

- [Official Documentation](https://argo-cd.readthedocs.io/)
- [GitHub Repository](https://github.com/argoproj/argo-cd)
- [Helm Chart](https://github.com/argoproj/argo-helm)
- [ArgoCD Examples](https://github.com/argoproj/argocd-example-apps)
