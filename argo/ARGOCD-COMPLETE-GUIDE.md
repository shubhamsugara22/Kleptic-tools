# ArgoCD - Declarative GitOps Continuous Deployment

ArgoCD is a declarative, GitOps continuous delivery tool for Kubernetes. It automates the deployment of applications to Kubernetes clusters by using Git repositories as the single source of truth.

## Quick Links

> ðŸ“‹ **[INDEX.md](INDEX.md)** - Complete documentation index with learning paths, use case navigation, and quick search

### ðŸ“š Documentation Guides
- ðŸ“– **[CHEATSHEET.md](CHEATSHEET.md)** - Quick reference for all ArgoCD commands and operations
- ðŸ—ï¸ **[PRODUCTION-GUIDE.md](PRODUCTION-GUIDE.md)** - High availability, security, monitoring, and disaster recovery
- ðŸ”§ **[INTEGRATION-GUIDE.md](INTEGRATION-GUIDE.md)** - CI/CD pipelines, secrets management, image updater, and notifications
- ðŸ› **[TROUBLESHOOTING-GUIDE.md](TROUBLESHOOTING-GUIDE.md)** - Common issues, debugging, performance tuning, and advanced patterns

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
Git Repository â†’ ArgoCD API â†’ Repository Server â†’ Application Controller â†’ Kubernetes Cluster
     â†‘                                                      â†“
     â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€ Sync Status â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

## Best Practices

### 1. **Repository Structure**
```
example-repo/
â”œâ”€â”€ base/
â”‚   â”œâ”€â”€ kustomization.yaml
â”‚   â””â”€â”€ deployment.yaml
â”œâ”€â”€ overlays/
â”‚   â”œâ”€â”€ dev/
â”‚   â”‚   â””â”€â”€ kustomization.yaml
â”‚   â”œâ”€â”€ staging/
â”‚   â”‚   â””â”€â”€ kustomization.yaml
â”‚   â””â”€â”€ prod/
â”‚       â””â”€â”€ kustomization.yaml
â””â”€â”€ argocd/
    â””â”€â”€ application.yaml
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
# ArgoCD Production Guide - HA, Security, Monitoring

## Table of Contents
- [High Availability Setup](#high-availability-setup)
- [Security Hardening](#security-hardening)
- [Monitoring & Observability](#monitoring--observability)
- [Backup & Disaster Recovery](#backup--disaster-recovery)
- [Performance Tuning](#performance-tuning)
- [Upgrade & Maintenance](#upgrade--maintenance)

## High Availability Setup

### 1. Multi-Replica Configuration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-server
  namespace: argocd
spec:
  replicas: 3  # Multiple replicas for HA
  template:
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/name: argocd-server
              topologyKey: kubernetes.io/hostname
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-application-controller
  namespace: argocd
spec:
  replicas: 3  # Sharded application controller
  template:
    spec:
      env:
      - name: ARGOCD_CONTROLLER_REPLICAS
        value: "3"
      - name: ARGOCD_CONTROLLER_SHARD
        valueFrom:
          fieldRef:
            fieldPath: metadata.name
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-repo-server
  namespace: argocd
spec:
  replicas: 3  # Multiple repo servers
  template:
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/name: argocd-repo-server
              topologyKey: kubernetes.io/hostname
```

### 2. Redis HA

```bash
# Install Redis HA using Helm
helm repo add dandydeveloper https://dandydeveloper.github.io/charts
helm install redis-ha dandydeveloper/redis-ha -n argocd \
  --set replicas=3 \
  --set haproxy.enabled=true

# Update ArgoCD to use Redis HA
kubectl patch deployment argocd-server -n argocd --type='json' \
  -p='[{"op": "add", "path": "/spec/template/spec/containers/0/env/-", 
        "value": {"name": "REDIS_HA_PROXY", "value": "redis-ha-haproxy:6379"}}]'
```

### 3. Load Balancer Configuration

```yaml
apiVersion: v1
kind: Service
metadata:
  name: argocd-server
  namespace: argocd
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: https
    service.beta.kubernetes.io/aws-load-balancer-ssl-ports: "443"
spec:
  type: LoadBalancer
  ports:
  - name: https
    port: 443
    targetPort: 8080
    protocol: TCP
  selector:
    app.kubernetes.io/name: argocd-server
```

### 4. Ingress with TLS

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: argocd-server-ingress
  namespace: argocd
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-passthrough: "true"
    nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
spec:
  ingressClassName: nginx
  rules:
  - host: argocd.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: argocd-server
            port:
              name: https
  tls:
  - hosts:
    - argocd.example.com
    secretName: argocd-server-tls
```

## Security Hardening

### 1. Network Policy

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: argocd-server-network-policy
  namespace: argocd
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: argocd-server
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app.kubernetes.io/name: argocd-repo-server
    ports:
    - protocol: TCP
      port: 8081
  - to:
    - namespaceSelector: {}
      podSelector:
        matchLabels:
          k8s-app: kube-dns
    ports:
    - protocol: UDP
      port: 53
```

### 2. Pod Security Standards

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: argocd
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

### 3. RBAC Least Privilege

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-rbac-cm
  namespace: argocd
data:
  policy.default: role:readonly
  policy.csv: |
    # Developer role - can view and sync apps in their project
    p, role:developer, applications, get, */*, allow
    p, role:developer, applications, sync, */*, allow
    p, role:developer, repositories, get, *, allow
    p, role:developer, clusters, get, *, allow
    
    # DevOps role - full application management
    p, role:devops, applications, *, */*, allow
    p, role:devops, repositories, *, *, allow
    p, role:devops, clusters, get, *, allow
    
    # Admin role - full access
    p, role:admin, *, *, *, allow
    
    # Group mappings
    g, devops-team, role:devops
    g, dev-team, role:developer
    g, platform-team, role:admin
```

### 4. Image Scanning Integration

```yaml
# Using Trivy for image scanning
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: argocd
spec:
  source:
    repoURL: https://github.com/org/repo.git
    targetRevision: HEAD
    path: k8s
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
  # Add image scanning hook
  syncOptions:
  - CreateNamespace=true
---
apiVersion: batch/v1
kind: Job
metadata:
  name: image-scan
  annotations:
    argocd.argoproj.io/hook: PreSync
    argocd.argoproj.io/hook-delete-policy: BeforeHookCreation
spec:
  template:
    spec:
      containers:
      - name: trivy-scan
        image: aquasec/trivy:latest
        command:
        - trivy
        - image
        - --severity
        - HIGH,CRITICAL
        - --exit-code
        - "1"
        - myapp:latest
      restartPolicy: Never
```

### 5. OPA/Gatekeeper Policy Enforcement

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: gatekeeper
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://open-policy-agent.github.io/gatekeeper/charts
    chart: gatekeeper
    targetRevision: 3.13.0
  destination:
    server: https://kubernetes.default.svc
    namespace: gatekeeper-system
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
---
# Example constraint - Require labels
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sRequiredLabels
metadata:
  name: require-labels
spec:
  match:
    kinds:
    - apiGroups: ["apps"]
      kinds: ["Deployment"]
  parameters:
    labels:
    - key: app
    - key: environment
    - key: team
```

## Monitoring & Observability

### 1. Prometheus Metrics

```yaml
apiVersion: v1
kind: ServiceMonitor
metadata:
  name: argocd-metrics
  namespace: argocd
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: argocd-server
  endpoints:
  - port: metrics
---
apiVersion: v1
kind: ServiceMonitor
metadata:
  name: argocd-repo-server-metrics
  namespace: argocd
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: argocd-repo-server
  endpoints:
  - port: metrics
---
apiVersion: v1
kind: ServiceMonitor
metadata:
  name: argocd-applicationset-controller-metrics
  namespace: argocd
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: argocd-applicationset-controller
  endpoints:
  - port: metrics
```

### 2. Key Metrics to Monitor

```promql
# Application sync status
argocd_app_info{sync_status="OutOfSync"}

# Application health status
argocd_app_info{health_status="Degraded"}

# Sync operations
rate(argocd_app_sync_total[5m])

# API request latency
histogram_quantile(0.95, rate(argocd_server_request_duration_seconds_bucket[5m]))

# Repository requests
rate(argocd_repo_request_total[5m])

# Controller queue depth
argocd_app_reconcile_count

# Redis connection status
redis_connected_clients
```

### 3. Grafana Dashboard

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-dashboard
  namespace: monitoring
data:
  argocd-dashboard.json: |
    {
      "dashboard": {
        "title": "ArgoCD",
        "panels": [
          {
            "title": "Applications by Sync Status",
            "targets": [{
              "expr": "count by (sync_status) (argocd_app_info)"
            }]
          },
          {
            "title": "Applications by Health Status",
            "targets": [{
              "expr": "count by (health_status) (argocd_app_info)"
            }]
          },
          {
            "title": "Sync Operations Rate",
            "targets": [{
              "expr": "rate(argocd_app_sync_total[5m])"
            }]
          }
        ]
      }
    }
```

### 4. Log Aggregation

```yaml
# FluentBit configuration for ArgoCD logs
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: logging
data:
  fluent-bit.conf: |
    [INPUT]
        Name              tail
        Path              /var/log/containers/argocd-*_argocd_*.log
        Parser            docker
        Tag               argocd.*
        
    [FILTER]
        Name                kubernetes
        Match               argocd.*
        Kube_URL            https://kubernetes.default.svc:443
        
    [OUTPUT]
        Name                elasticsearch
        Match               argocd.*
        Host                elasticsearch.logging.svc
        Port                9200
        Index               argocd-logs
```

### 5. Alerting Rules

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: argocd-alerts
  namespace: argocd
spec:
  groups:
  - name: argocd
    interval: 30s
    rules:
    - alert: ArgoCDAppOutOfSync
      expr: argocd_app_info{sync_status="OutOfSync"} > 0
      for: 15m
      labels:
        severity: warning
      annotations:
        summary: "ArgoCD application {{ $labels.name }} is out of sync"
        
    - alert: ArgoCDAppUnhealthy
      expr: argocd_app_info{health_status!~"Healthy|Progressing"} > 0
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "ArgoCD application {{ $labels.name }} is unhealthy"
        
    - alert: ArgoCDSyncFailed
      expr: increase(argocd_app_sync_total{phase="Failed"}[5m]) > 0
      labels:
        severity: warning
      annotations:
        summary: "ArgoCD sync failed for {{ $labels.name }}"
```

## Backup & Disaster Recovery

### 1. Backup ArgoCD Configuration

```bash
#!/bin/bash
# backup-argocd.sh

BACKUP_DIR="argocd-backup-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Backup ArgoCD applications
kubectl get applications -n argocd -o yaml > "$BACKUP_DIR/applications.yaml"

# Backup ArgoCD projects
kubectl get appprojects -n argocd -o yaml > "$BACKUP_DIR/appprojects.yaml"

# Backup ArgoCD ApplicationSets
kubectl get applicationsets -n argocd -o yaml > "$BACKUP_DIR/applicationsets.yaml"

# Backup ArgoCD secrets
kubectl get secrets -n argocd -o yaml > "$BACKUP_DIR/secrets.yaml"

# Backup ArgoCD configmaps
kubectl get configmaps -n argocd -o yaml > "$BACKUP_DIR/configmaps.yaml"

# Create tarball
tar -czf "$BACKUP_DIR.tar.gz" "$BACKUP_DIR"
rm -rf "$BACKUP_DIR"

echo "Backup completed: $BACKUP_DIR.tar.gz"
```

### 2. Automated Backup with CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: argocd-backup
  namespace: argocd
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: argocd-backup
          containers:
          - name: backup
            image: bitnami/kubectl:latest
            command:
            - /bin/bash
            - -c
            - |
              BACKUP_DATE=$(date +%Y%m%d-%H%M%S)
              kubectl get applications -n argocd -o yaml > /backup/applications-$BACKUP_DATE.yaml
              kubectl get appprojects -n argocd -o yaml > /backup/appprojects-$BACKUP_DATE.yaml
              kubectl get applicationsets -n argocd -o yaml > /backup/applicationsets-$BACKUP_DATE.yaml
              
              # Upload to S3 (example)
              aws s3 cp /backup s3://my-bucket/argocd-backups/$BACKUP_DATE/ --recursive
            volumeMounts:
            - name: backup
              mountPath: /backup
          volumes:
          - name: backup
            emptyDir: {}
          restartPolicy: OnFailure
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: argocd-backup
  namespace: argocd
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: argocd-backup
  namespace: argocd
rules:
- apiGroups: ["argoproj.io"]
  resources: ["applications", "appprojects", "applicationsets"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["secrets", "configmaps"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: argocd-backup
  namespace: argocd
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: argocd-backup
subjects:
- kind: ServiceAccount
  name: argocd-backup
  namespace: argocd
```

### 3. Restore Procedure

```bash
#!/bin/bash
# restore-argocd.sh

BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup-file.tar.gz>"
    exit 1
fi

# Extract backup
tar -xzf "$BACKUP_FILE"
BACKUP_DIR=$(basename "$BACKUP_FILE" .tar.gz)

# Restore applications
kubectl apply -f "$BACKUP_DIR/applications.yaml"

# Restore projects
kubectl apply -f "$BACKUP_DIR/appprojects.yaml"

# Restore ApplicationSets
kubectl apply -f "$BACKUP_DIR/applicationsets.yaml"

# Restore secrets (be careful with this)
kubectl apply -f "$BACKUP_DIR/secrets.yaml"

# Restore configmaps
kubectl apply -f "$BACKUP_DIR/configmaps.yaml"

echo "Restore completed from $BACKUP_FILE"
```

## Performance Tuning

### 1. Controller Sharding

```yaml
# Scale application-controller with sharding
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: argocd-application-controller
  namespace: argocd
spec:
  replicas: 3  # Number of shards
  template:
    spec:
      containers:
      - name: argocd-application-controller
        env:
        - name: ARGOCD_CONTROLLER_REPLICAS
          value: "3"
        - name: ARGOCD_CONTROLLER_SHARD_NUMBER
          value: "$(POD_NAME)"
        resources:
          requests:
            cpu: 1000m
            memory: 2Gi
          limits:
            cpu: 2000m
            memory: 4Gi
```

### 2. Repository Server Scaling

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-repo-server
  namespace: argocd
spec:
  replicas: 5  # Scale based on repository load
  template:
    spec:
      containers:
      - name: argocd-repo-server
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 1000m
            memory: 2Gi
        env:
        - name: ARGOCD_EXEC_TIMEOUT
          value: "180s"
        - name: ARGOCD_GIT_REQUEST_TIMEOUT
          value: "60s"
```

### 3. Resource Exclusions

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
  namespace: argocd
data:
  resource.exclusions: |
    - apiGroups:
      - "*"
      kinds:
      - ProviderConfigUsage
    - apiGroups:
      - cilium.io
      kinds:
      - CiliumIdentity
  resource.compareoptions: |
    ignoreAggregatedRoles: true
```

### 4. Caching Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
  namespace: argocd
data:
  timeout.reconciliation: "180s"
  timeout.hard.reconciliation: "0"  # Disable hard reconciliation
  application.instanceLabelKey: "argocd.argoproj.io/instance"
```

## Upgrade & Maintenance

### 1. Upgrade ArgoCD

```bash
# Check current version
kubectl get deployment argocd-server -n argocd -o jsonpath='{.spec.template.spec.containers[0].image}'

# Backup before upgrade
./backup-argocd.sh

# Upgrade to specific version
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.9.0/manifests/install.yaml

# Or use latest stable
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Wait for rollout
kubectl rollout status deployment/argocd-server -n argocd
kubectl rollout status deployment/argocd-repo-server -n argocd
kubectl rollout status statefulset/argocd-application-controller -n argocd

# Verify version
argocd version
```

### 2. Maintenance Window

```yaml
# Set maintenance mode - prevent syncs
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
  namespace: argocd
data:
  application.resourceTrackingMethod: "annotation"
  # Add maintenance window
  resource.customizations.health.argoproj.io_Application: |
    hs = {}
    hs.status = "Suspended"
    hs.message = "Maintenance in progress"
    return hs
```

### 3. Database Migration

```bash
# If upgrading from pre-2.0 to 2.x
# Backup existing data
kubectl exec -n argocd argocd-application-controller-0 -- \
  argocd admin export > backup.yaml

# After upgrade, import if needed
kubectl exec -n argocd argocd-application-controller-0 -- \
  argocd admin import < backup.yaml
```

### 4. Health Check

```bash
# Check ArgoCD health
kubectl exec -n argocd argocd-server-xxx -- argocd admin settings resource-overrides health

# Check applications
argocd app list

# Check repositories
argocd repo list

# Test connectivity
argocd cluster list
```

## Additional Resources

- [ArgoCD HA Installation](https://argo-cd.readthedocs.io/en/stable/operator-manual/high_availability/)
- [Security Best Practices](https://argo-cd.readthedocs.io/en/stable/operator-manual/security/)
- [Monitoring Guide](https://argo-cd.readthedocs.io/en/stable/operator-manual/metrics/)
- [Disaster Recovery](https://argo-cd.readthedocs.io/en/stable/operator-manual/disaster_recovery/)
# ArgoCD Integration Guide - CI/CD, Secrets, Image Updates

## Table of Contents
- [CI/CD Integration](#cicd-integration)
- [Secrets Management](#secrets-management)
- [ArgoCD Image Updater](#argocd-image-updater)
- [Notifications Configuration](#notifications-configuration)

## CI/CD Integration

### 1. GitHub Actions Integration

```yaml
# .github/workflows/deploy.yml
name: Build and Deploy

on:
  push:
    branches: [ main ]
    
env:
  ARGOCD_SERVER: argocd.example.com
  APP_NAME: myapp

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Login to Container Registry
      uses: docker/login-action@v2
      with:
        registry: ghcr.io
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Build and push
      uses: docker/build-push-action@v4
      with:
        context: .
        push: true
        tags: |
          ghcr.io/${{ github.repository }}:${{ github.sha }}
          ghcr.io/${{ github.repository }}:latest
        cache-from: type=gha
        cache-to: type=gha,mode=max
  
  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Update image tag in manifests
      run: |
        cd k8s
        sed -i "s|image:.*|image: ghcr.io/${{ github.repository }}:${{ github.sha }}|g" deployment.yaml
        git config user.name "GitHub Actions"
        git config user.email "actions@github.com"
        git add .
        git commit -m "Update image to ${{ github.sha }}"
        git push
    
    - name: Install ArgoCD CLI
      run: |
        curl -sSL -o argocd https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
        chmod +x argocd
        sudo mv argocd /usr/local/bin/
    
    - name: Sync ArgoCD Application
      env:
        ARGOCD_AUTH_TOKEN: ${{ secrets.ARGOCD_TOKEN }}
      run: |
        argocd app sync $APP_NAME --server $ARGOCD_SERVER --auth-token $ARGOCD_AUTH_TOKEN
        argocd app wait $APP_NAME --server $ARGOCD_SERVER --auth-token $ARGOCD_AUTH_TOKEN --health
```

### 2. GitLab CI Integration

```yaml
# .gitlab-ci.yml
stages:
  - build
  - deploy

variables:
  IMAGE_TAG: $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  ARGOCD_SERVER: argocd.example.com

build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
    - docker build -t $IMAGE_TAG .
    - docker push $IMAGE_TAG
  only:
    - main

update-manifest:
  stage: deploy
  image: alpine/git
  before_script:
    - apk add --no-cache sed
  script:
    - git clone https://${CI_USERNAME}:${CI_ACCESS_TOKEN}@gitlab.com/org/k8s-manifests.git
    - cd k8s-manifests
    - sed -i "s|image:.*|image: $IMAGE_TAG|g" deployment.yaml
    - git config user.email "ci@gitlab.com"
    - git config user.name "GitLab CI"
    - git add .
    - git commit -m "Update image to $CI_COMMIT_SHA"
    - git push origin main
  only:
    - main

argocd-sync:
  stage: deploy
  image: argoproj/argocd:latest
  script:
    - argocd login $ARGOCD_SERVER --auth-token $ARGOCD_TOKEN --insecure
    - argocd app sync myapp --force
    - argocd app wait myapp --health
  only:
    - main
```

### 3. Jenkins Pipeline

```groovy
// Jenkinsfile
pipeline {
    agent any
    
    environment {
        DOCKER_REGISTRY = 'registry.example.com'
        ARGOCD_SERVER = 'argocd.example.com'
        APP_NAME = 'myapp'
        IMAGE_TAG = "${env.BUILD_ID}"
    }
    
    stages {
        stage('Build') {
            steps {
                script {
                    docker.build("${DOCKER_REGISTRY}/${APP_NAME}:${IMAGE_TAG}")
                }
            }
        }
        
        stage('Push') {
            steps {
                script {
                    docker.withRegistry("https://${DOCKER_REGISTRY}", 'docker-credentials') {
                        docker.image("${DOCKER_REGISTRY}/${APP_NAME}:${IMAGE_TAG}").push()
                        docker.image("${DOCKER_REGISTRY}/${APP_NAME}:${IMAGE_TAG}").push('latest')
                    }
                }
            }
        }
        
        stage('Update Manifests') {
            steps {
                script {
                    git credentialsId: 'git-credentials', url: 'https://github.com/org/k8s-manifests.git'
                    sh """
                        sed -i "s|image:.*|image: ${DOCKER_REGISTRY}/${APP_NAME}:${IMAGE_TAG}|g" k8s/deployment.yaml
                        git config user.email 'jenkins@example.com'
                        git config user.name 'Jenkins'
                        git add .
                        git commit -m 'Update image to ${IMAGE_TAG}'
                        git push origin main
                    """
                }
            }
        }
        
        stage('Sync ArgoCD') {
            steps {
                withCredentials([string(credentialsId: 'argocd-token', variable: 'ARGOCD_TOKEN')]) {
                    sh """
                        argocd login ${ARGOCD_SERVER} --auth-token ${ARGOCD_TOKEN} --insecure
                        argocd app sync ${APP_NAME}
                        argocd app wait ${APP_NAME} --health
                    """
                }
            }
        }
    }
    
    post {
        success {
            slackSend color: 'good', message: "Deployment successful: ${APP_NAME} - ${IMAGE_TAG}"
        }
        failure {
            slackSend color: 'danger', message: "Deployment failed: ${APP_NAME} - ${IMAGE_TAG}"
        }
    }
}
```

### 4. Azure DevOps Pipeline

```yaml
# azure-pipelines.yml
trigger:
  - main

variables:
  imageTag: '$(Build.BuildId)'
  imageName: 'myapp'
  argocdServer: 'argocd.example.com'
  argocdApp: 'myapp'

stages:
- stage: Build
  jobs:
  - job: BuildAndPush
    pool:
      vmImage: 'ubuntu-latest'
    steps:
    - task: Docker@2
      inputs:
        containerRegistry: 'ACR Connection'
        repository: $(imageName)
        command: 'buildAndPush'
        Dockerfile: '**/Dockerfile'
        tags: |
          $(imageTag)
          latest

- stage: Deploy
  jobs:
  - job: UpdateManifests
    pool:
      vmImage: 'ubuntu-latest'
    steps:
    - checkout: self
      persistCredentials: true
    
    - script: |
        cd k8s
        sed -i "s|image:.*|image: $(containerRegistry)/$(imageName):$(imageTag)|g" deployment.yaml
        git config user.email "azure@devops.com"
        git config user.name "Azure DevOps"
        git add .
        git commit -m "Update image to $(imageTag)"
        git push
      displayName: 'Update image tag'
  
  - job: ArgoSync
    dependsOn: UpdateManifests
    pool:
      vmImage: 'ubuntu-latest'
    steps:
    - script: |
        curl -sSL -o argocd https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
        chmod +x argocd
        sudo mv argocd /usr/local/bin/
      displayName: 'Install ArgoCD CLI'
    
    - script: |
        argocd login $(argocdServer) --auth-token $(argocdToken) --insecure
        argocd app sync $(argocdApp)
        argocd app wait $(argocdApp) --health
      displayName: 'Sync ArgoCD Application'
      env:
        argocdToken: $(ARGOCD_TOKEN)
```

## Secrets Management

### 1. Sealed Secrets

#### Installation

```bash
# Install Sealed Secrets controller
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.0/controller.yaml

# Install kubeseal CLI
wget https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.0/kubeseal-linux-amd64
sudo install -m 755 kubeseal-linux-amd64 /usr/local/bin/kubeseal
```

#### Usage

```bash
# Create a secret
kubectl create secret generic mysecret \
  --from-literal=username=admin \
  --from-literal=password=secretpass \
  --dry-run=client -o yaml > secret.yaml

# Seal the secret
kubeseal --format yaml < secret.yaml > sealed-secret.yaml

# sealed-secret.yaml can be committed to Git
# Apply sealed secret
kubectl apply -f sealed-secret.yaml
```

#### ArgoCD Integration

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: argocd
spec:
  source:
    repoURL: https://github.com/org/repo.git
    targetRevision: HEAD
    path: k8s
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true

# In your repo: k8s/sealed-secret.yaml
---
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: mysecret
  namespace: production
spec:
  encryptedData:
    username: AgBR7P...
    password: AgCQ9K...
```

### 2. External Secrets Operator (ESO)

#### Installation

```bash
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets \
  external-secrets/external-secrets \
  -n external-secrets-system \
  --create-namespace
```

#### AWS Secrets Manager Integration

```yaml
# SecretStore - connects to AWS Secrets Manager
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secretstore
  namespace: production
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets-sa

---
# ExternalSecret - syncs secret from AWS
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: database-credentials
  namespace: production
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secretstore
    kind: SecretStore
  target:
    name: db-secret
    creationPolicy: Owner
  data:
  - secretKey: username
    remoteRef:
      key: prod/database
      property: username
  - secretKey: password
    remoteRef:
      key: prod/database
      property: password
```

#### Azure Key Vault Integration

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: azure-secretstore
  namespace: production
spec:
  provider:
    azurekv:
      authType: ManagedIdentity
      vaultUrl: "https://my-vault.vault.azure.net"
      identityId: "/subscriptions/xxx/resourceGroups/xxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/xxx"

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: app-secrets
  namespace: production
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: azure-secretstore
    kind: SecretStore
  target:
    name: app-secret
  data:
  - secretKey: api-key
    remoteRef:
      key: api-key
```

### 3. HashiCorp Vault Integration

#### Install Vault Agent Injector

```bash
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install vault hashicorp/vault \
  --set "injector.enabled=true" \
  --set "server.enabled=false"
```

#### ArgoCD with Vault

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: production
spec:
  template:
    metadata:
      annotations:
        vault.hashicorp.com/agent-inject: "true"
        vault.hashicorp.com/role: "myapp"
        vault.hashicorp.com/agent-inject-secret-config: "secret/data/myapp/config"
        vault.hashicorp.com/agent-inject-template-config: |
          {{- with secret "secret/data/myapp/config" -}}
          export DATABASE_URL="{{ .Data.data.database_url }}"
          export API_KEY="{{ .Data.data.api_key }}"
          {{- end }}
    spec:
      serviceAccountName: myapp
      containers:
      - name: app
        image: myapp:latest
        command: ["/bin/sh"]
        args:
        - -c
        - source /vault/secrets/config && ./app
```

#### Vault Plugin for ArgoCD

```yaml
# Install argocd-vault-plugin
apiVersion: v1
kind: ConfigMap
metadata:
  name: cmp-plugin
  namespace: argocd
data:
  avp.yaml: |
    apiVersion: argoproj.io/v1alpha1
    kind: ConfigManagementPlugin
    metadata:
      name: argocd-vault-plugin
    spec:
      allowConcurrency: true
      discover:
        find:
          command:
            - sh
            - "-c"
            - "find . -name '*.yaml' | xargs -I {} grep \"<path\\|avp\\.kubernetes\\.io\" {} | grep ."
      generate:
        command:
          - argocd-vault-plugin
          - generate
          - ./
      lockRepo: false

# Usage in manifests
---
apiVersion: v1
kind: Secret
metadata:
  name: myapp-secret
  namespace: production
type: Opaque
stringData:
  username: <path:secret/data/myapp#username>
  password: <path:secret/data/myapp#password>
```

### 4. SOPS (Secrets Operations)

```bash
# Install SOPS
wget https://github.com/mozilla/sops/releases/download/v3.8.1/sops-v3.8.1.linux.amd64
sudo mv sops-v3.8.1.linux.amd64 /usr/local/bin/sops
sudo chmod +x /usr/local/bin/sops

# Encrypt a secret
sops --encrypt --age <AGE_PUBLIC_KEY> secret.yaml > secret.enc.yaml

# Decrypt (for viewing)
sops --decrypt secret.enc.yaml

# ArgoCD with SOPS (using Helm Secrets plugin)
# Install argo-cd helm-secrets plugin
```

## ArgoCD Image Updater

### 1. Installation

```bash
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj-labs/argocd-image-updater/stable/manifests/install.yaml
```

### 2. Basic Configuration

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: argocd
  annotations:
    # Enable image updater
    argocd-image-updater.argoproj.io/image-list: myimage=myregistry/myapp
    
    # Update strategy: latest, semver, digest
    argocd-image-updater.argoproj.io/myimage.update-strategy: semver
    
    # Constraint (semver only)
    argocd-image-updater.argoproj.io/myimage.allow-tags: regexp:^v[0-9]+\.[0-9]+\.[0-9]+$
    
    # Write back method: git or argocd
    argocd-image-updater.argoproj.io/write-back-method: git
    
    # Git branch to update
    argocd-image-updater.argoproj.io/git-branch: main
spec:
  source:
    repoURL: https://github.com/org/repo.git
    targetRevision: HEAD
    path: k8s
    helm:
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
```

### 3. Advanced Image Update Patterns

```yaml
# Multiple images
metadata:
  annotations:
    argocd-image-updater.argoproj.io/image-list: |
      frontend=myregistry/frontend,
      backend=myregistry/backend,
      api=myregistry/api
    argocd-image-updater.argoproj.io/frontend.update-strategy: latest
    argocd-image-updater.argoproj.io/backend.update-strategy: semver:~1.2
    argocd-image-updater.argoproj.io/api.update-strategy: digest

# Private registry authentication
argocd-image-updater.argoproj.io/myimage.pull-secret: pullsecret:argocd/my-registry-secret

# Custom Git commit message
argocd-image-updater.argoproj.io/git-commit-message: |
  build: update {{.Image}} to {{.NewTag}}
  
  Generated by ArgoCD Image Updater
```

## Notifications Configuration

### 1. Install Notifications

```bash
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj-labs/argocd-notifications/stable/manifests/install.yaml
```

### 2. Slack Integration

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: argocd-notifications-secret
  namespace: argocd
stringData:
  slack-token: xoxb-your-slack-token

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
  namespace: argocd
data:
  service.slack: |
    token: $slack-token
  
  template.app-deployed: |
    message: |
      Application {{.app.metadata.name}} is now running new version.
    slack:
      attachments: |
        [{
          "title": "{{ .app.metadata.name}}",
          "title_link":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
          "color": "#18be52",
          "fields": [
          {
            "title": "Sync Status",
            "value": "{{.app.status.sync.status}}",
            "short": true
          },
          {
            "title": "Repository",
            "value": "{{.app.spec.source.repoURL}}",
            "short": true
          }
          ]
        }]
  
  template.app-health-degraded: |
    message: |
      Application {{.app.metadata.name}} has degraded.
    slack:
      attachments: |
        [{
          "title": "{{ .app.metadata.name}}",
          "title_link": "{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
          "color": "#f4c030",
          "fields": [
          {
            "title": "Health Status",
            "value": "{{.app.status.health.status}}",
            "short": true
          }
          ]
        }]
  
  trigger.on-deployed: |
    - description: Application is synced and healthy
      send:
      - app-deployed
      when: app.status.operationState.phase in ['Succeeded'] and app.status.health.status == 'Healthy'
  
  trigger.on-health-degraded: |
    - description: Application has degraded
      send:
      - app-health-degraded
      when: app.status.health.status == 'Degraded'
```

### 3. Subscribe Applications

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: argocd
  annotations:
    notifications.argoproj.io/subscribe.on-deployed.slack: dev-team-channel
    notifications.argoproj.io/subscribe.on-health-degraded.slack: alerts-channel
spec:
  # ... rest of application spec
```

### 4. Microsoft Teams Integration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
  namespace: argocd
data:
  service.teams: |
    recipientUrls:
      devTeam: https://outlook.office.com/webhook/xxx
  
  template.app-deployed: |
    teams:
      themeColor: "#000080"
      summary: "Application {{.app.metadata.name}} deployed"
      sections: |
        [{
          "activityTitle": "Application {{.app.metadata.name}} deployed",
          "facts": [
            {
              "name": "Sync Status",
              "value": "{{.app.status.sync.status}}"
            },
            {
              "name": "Repository",
              "value": "{{.app.spec.source.repoURL}}"
            }
          ]
        }]
```

### 5. Email Notifications

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: argocd-notifications-secret
  namespace: argocd
stringData:
  email-username: smtp-user@example.com
  email-password: smtp-password

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
  namespace: argocd
data:
  service.email.gmail: |
    username: $email-username
    password: $email-password
    host: smtp.gmail.com
    port: 587
    from: $email-username
  
  template.app-sync-failed: |
    email:
      subject: ArgoCD Sync Failed - {{.app.metadata.name}}
    message: |
      Application {{.app.metadata.name}} sync has failed.
      
      Sync Status: {{.app.status.sync.status}}
      Health Status: {{.app.status.health.status}}
      
      View in ArgoCD: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}
```

## Additional Resources

- [ArgoCD Best Practices](https://argo-cd.readthedocs.io/en/stable/user-guide/best_practices/)
- [Image Updater Documentation](https://argocd-image-updater.readthedocs.io/)
- [Notifications Documentation](https://argocd-notifications.readthedocs.io/)
- [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)
- [External Secrets Operator](https://external-secrets.io/)
# ArgoCD Troubleshooting & Advanced Patterns

## Table of Contents
- [Common Issues](#common-issues)
- [Debugging Techniques](#debugging-techniques)
- [Performance Optimization](#performance-optimization)
- [Advanced Patterns](#advanced-patterns)
- [Resource Hooks](#resource-hooks)
- [Multi-Tenancy](#multi-tenancy)

## Common Issues

### 1. Application OutOfSync

**Symptom**: Application shows "OutOfSync" status

**Causes & Solutions**:

```bash
# Check diff between desired and live state
argocd app diff myapp

# Check if it's a known issue (resource tracking, label changes)
kubectl get app myapp -n argocd -o yaml | grep -A 10 status

# Force sync with replace
argocd app sync myapp --force --replace

# If specific resources are causing issues, add to ignore differences
```

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
spec:
  ignoreDifferences:
  - group: apps
    kind: Deployment
    jsonPointers:
    - /spec/replicas  # Ignore HPA-managed replicas
  - group: apps
    kind: StatefulSet
    jsonPointers:
    - /spec/volumeClaimTemplates  # Ignore PVC changes
  - group: ""
    kind: Service
    jqPathExpressions:
    - '.spec.ports[] | select(.port == 8080)'  # Ignore specific port
```

### 2. Application Sync Hanging

**Symptom**: Sync operation stuck in "Progressing"

**Solutions**:

```bash
# Check operation status
argocd app get myapp --show-operation

# Check pod logs
kubectl logs -n argocd deployment/argocd-application-controller -f

# Terminate stuck operation
argocd app terminate-op myapp

# Check for resource hooks that might be blocking
kubectl get pods -n target-namespace -l argocd.argoproj.io/instance=myapp

# Check sync waves - might be waiting for previous wave
kubectl get all -n target-namespace -o jsonpath='{range .items[*]}{.metadata.annotations.argocd\.argoproj\.io/sync-wave}{"\t"}{.kind}/{.metadata.name}{"\n"}{end}' | sort -n
```

### 3. Health Check Failures

**Symptom**: Application showing "Degraded" or "Unknown" health

**Custom Health Check**:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
  namespace: argocd
data:
  # Custom health check for CRD
  resource.customizations.health.mygroup.example.com_MyResource: |
    hs = {}
    if obj.status ~= nil then
      if obj.status.conditions ~= nil then
        for i, condition in ipairs(obj.status.conditions) do
          if condition.type == "Ready" and condition.status == "False" then
            hs.status = "Degraded"
            hs.message = condition.message
            return hs
          end
          if condition.type == "Ready" and condition.status == "True" then
            hs.status = "Healthy"
            hs.message = "Resource is healthy"
            return hs
          end
        end
      end
    end
    hs.status = "Progressing"
    hs.message = "Waiting for resource"
    return hs
```

### 4. Repository Connection Issues

**Symptom**: "Unable to connect to repository"

```bash
# Test repository connection
argocd repo get https://github.com/org/repo.git

# Re-add repository with credentials
argocd repo add https://github.com/org/repo.git \
  --username <username> \
  --password <token>

# For SSH keys
argocd repo add git@github.com:org/repo.git \
  --ssh-private-key-path ~/.ssh/id_rsa

# Check repo-server logs
kubectl logs -n argocd deployment/argocd-repo-server -f

# Test from repo-server pod
kubectl exec -n argocd deployment/argocd-repo-server -- git ls-remote https://github.com/org/repo.git
```

### 5. RBAC Permission Errors

**Symptom**: "permission denied" or "forbidden" errors

```bash
# Check AppProject permissions
kubectl get appproject -n argocd myproject -o yaml

# Check user/group permissions
argocd account can-i sync applications myapp

# Add permissions to AppProject
kubectl patch appproject myproject -n argocd --type merge -p '
{
  "spec": {
    "roles": [{
      "name": "developer",
      "policies": [
        "p, proj:myproject:developer, applications, sync, myproject/*, allow",
        "p, proj:myproject:developer, applications, get, myproject/*, allow"
      ]
    }]
  }
}'
```

## Debugging Techniques

### 1. Enable Debug Logging

```yaml
# Enable debug logs for application controller
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-application-controller
  namespace: argocd
spec:
  template:
    spec:
      containers:
      - name: argocd-application-controller
        command:
        - argocd-application-controller
        - --loglevel=debug
        - --status-processors=20
        - --operation-processors=10
```

### 2. Application Events

```bash
# View application events
kubectl get events -n argocd --field-selector involvedObject.name=myapp

# Describe application resource
kubectl describe application myapp -n argocd

# Get detailed status
argocd app get myapp --output yaml | less
```

### 3. Resource Tracking

```bash
# Check tracked resources
kubectl get configmap argocd-cm -n argocd -o yaml | grep -A 5 "application.resourceTrackingMethod"

# View resources tracked by application
argocd app resources myapp

# Check for orphaned resources
kubectl get all -n target-namespace -l app.kubernetes.io/instance=myapp
```

### 4. Manifest Generation Debugging

```bash
# Generate manifests without applying
argocd app manifests myapp

# View rendered Helm templates
argocd app manifests myapp --helm-set image.tag=v2.0.0

# Test Kustomize build
kubectl kustomize path/to/overlay

# Validate manifests
argocd app manifests myapp | kubectl apply --dry-run=client -f -
```

## Performance Optimization

### 1. Application Controller Tuning

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-application-controller
  namespace: argocd
spec:
  template:
    spec:
      containers:
      - name: argocd-application-controller
        env:
        # Increase workers
        - name: ARGOCD_CONTROLLER_STATUS_PROCESSORS
          value: "50"
        - name: ARGOCD_CONTROLLER_OPERATION_PROCESSORS
          value: "25"
        
        # Adjust reconciliation
        - name: ARGOCD_RECONCILIATION_TIMEOUT
          value: "180s"
        - name: ARGOCD_REPO_SERVER_TIMEOUT_SECONDS
          value: "120"
        
        # Resource management
        resources:
          requests:
            cpu: "2"
            memory: "4Gi"
          limits:
            cpu: "4"
            memory: "8Gi"
```

### 2. Repository Server Optimization

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-repo-server
  namespace: argocd
spec:
  replicas: 3  # Scale horizontally
  template:
    spec:
      containers:
      - name: argocd-repo-server
        env:
        # Enable parallelism
        - name: ARGOCD_EXEC_TIMEOUT
          value: "300s"
        - name: ARGOCD_GIT_ATTEMPTS_COUNT
          value: "3"
        
        # Configure cache
        - name: ARGOCD_REPO_CACHE_EXPIRATION
          value: "24h"
        
        resources:
          requests:
            cpu: "1"
            memory: "2Gi"
          limits:
            cpu: "2"
            memory: "4Gi"
        
        volumeMounts:
        - name: repo-cache
          mountPath: /tmp
      
      volumes:
      - name: repo-cache
        emptyDir:
          sizeLimit: 10Gi
```

### 3. Exclude Resources from Reconciliation

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
  namespace: argocd
data:
  # Exclude resources that change frequently
  resource.exclusions: |
    - apiGroups:
      - "events.k8s.io"
      kinds:
      - Event
      clusters:
      - "*"
    - apiGroups:
      - ""
      kinds:
      - Event
      clusters:
      - "*"
    - apiGroups:
      - "metrics.k8s.io"
      kinds:
      - "*"
      clusters:
      - "*"
```

### 4. Optimize Large Applications

```yaml
# Use App of Apps pattern instead of monolithic apps
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: platform
spec:
  project: default
  source:
    repoURL: https://github.com/org/platform.git
    targetRevision: HEAD
    path: apps
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd
  syncPolicy:
    automated:
      prune: true

# Structure: apps/ directory contains multiple application manifests
# apps/
#   monitoring.yaml
#   logging.yaml
#   ingress.yaml
#   databases.yaml
```

## Advanced Patterns

### 1. Progressive Delivery with Argo Rollouts

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: myapp
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
      - setWeight: 80
      - pause: {duration: 1h}
      
      # Analysis template for automated rollback
      analysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: myapp
  
  revisionHistoryLimit: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
        image: myapp:latest
        ports:
        - containerPort: 8080

---
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: success-rate
spec:
  args:
  - name: service-name
  metrics:
  - name: success-rate
    interval: 5m
    successCondition: result >= 0.95
    failureLimit: 3
    provider:
      prometheus:
        address: http://prometheus.monitoring:9090
        query: |
          sum(rate(
            http_requests_total{service="{{args.service-name}}",status!~"5.."}[5m]
          )) /
          sum(rate(
            http_requests_total{service="{{args.service-name}}"}[5m]
          ))
```

### 2. Blue-Green Deployments

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: myapp
spec:
  replicas: 5
  strategy:
    blueGreen:
      activeService: myapp-active
      previewService: myapp-preview
      autoPromotionEnabled: false
      scaleDownDelaySeconds: 300
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
        image: myapp:latest

---
apiVersion: v1
kind: Service
metadata:
  name: myapp-active
spec:
  selector:
    app: myapp
  ports:
  - port: 80
    targetPort: 8080

---
apiVersion: v1
kind: Service
metadata:
  name: myapp-preview
spec:
  selector:
    app: myapp
  ports:
  - port: 80
    targetPort: 8080
```

### 3. Multi-Source Applications

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: argocd
spec:
  project: default
  sources:
  # Source 1: Helm chart from OCI registry
  - repoURL: registry-1.docker.io/bitnamicharts
    chart: postgresql
    targetRevision: 12.1.9
    helm:
      releaseName: postgres
      valueFiles:
      - $values/postgres/values.yaml
  
  # Source 2: Custom values from Git
  - repoURL: https://github.com/org/configs.git
    targetRevision: HEAD
    ref: values
  
  # Source 3: Application manifests
  - repoURL: https://github.com/org/app.git
    targetRevision: HEAD
    path: k8s
  
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

### 4. Dynamic Environment Generation with ApplicationSets

```yaml
# Git Files Generator - create app per directory
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: microservices
  namespace: argocd
spec:
  generators:
  - git:
      repoURL: https://github.com/org/microservices.git
      revision: HEAD
      files:
      - path: "services/*/config.json"
  template:
    metadata:
      name: '{{path.basename}}'
    spec:
      project: default
      source:
        repoURL: https://github.com/org/microservices.git
        targetRevision: HEAD
        path: 'services/{{path.basename}}'
      destination:
        server: https://kubernetes.default.svc
        namespace: '{{namespace}}'
      syncPolicy:
        automated:
          prune: true
          selfHeal: true

# config.json in each service directory:
# {
#   "namespace": "production",
#   "replicas": 3
# }
```

## Resource Hooks

### 1. PreSync Hook - Database Migration

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migration
  annotations:
    argocd.argoproj.io/hook: PreSync
    argocd.argoproj.io/hook-delete-policy: BeforeHookCreation
    argocd.argoproj.io/sync-wave: "1"
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: myapp:latest
        command: ["./migrate"]
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
      restartPolicy: Never
  backoffLimit: 2
```

### 2. Sync Hook - Schema Update

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: schema-update
  annotations:
    argocd.argoproj.io/hook: Sync
    argocd.argoproj.io/hook-delete-policy: HookSucceeded
    argocd.argoproj.io/sync-wave: "2"
spec:
  template:
    spec:
      containers:
      - name: update
        image: migrate/migrate
        args:
        - "-path=/migrations"
        - "-database=$(DATABASE_URL)"
        - "up"
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
      restartPolicy: OnFailure
```

### 3. PostSync Hook - Smoke Tests

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: smoke-tests
  annotations:
    argocd.argoproj.io/hook: PostSync
    argocd.argoproj.io/hook-delete-policy: HookSucceeded
    argocd.argoproj.io/sync-wave: "5"
spec:
  template:
    spec:
      containers:
      - name: test
        image: curlimages/curl
        command: 
        - /bin/sh
        - -c
        - |
          set -e
          echo "Running smoke tests..."
          curl -f http://myapp.production.svc.cluster.local/health || exit 1
          curl -f http://myapp.production.svc.cluster.local/ready || exit 1
          echo "All tests passed!"
      restartPolicy: Never
  backoffLimit: 3
```

### 4. Sync Failure Hook - Rollback Notification

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: sync-failure-alert
  annotations:
    argocd.argoproj.io/hook: SyncFail
    argocd.argoproj.io/hook-delete-policy: HookSucceeded
spec:
  template:
    spec:
      containers:
      - name: notify
        image: curlimages/curl
        command:
        - /bin/sh
        - -c
        - |
          curl -X POST https://hooks.slack.com/services/XXX \
            -H 'Content-Type: application/json' \
            -d '{"text":"ArgoCD Sync Failed for myapp!"}'
      restartPolicy: Never
```

## Multi-Tenancy

### 1. Namespace-based Isolation

```yaml
# Create AppProject per team
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: team-a
  namespace: argocd
spec:
  description: Team A Project
  
  # Restrict source repositories
  sourceRepos:
  - 'https://github.com/org/team-a-*'
  
  # Restrict destination namespaces
  destinations:
  - namespace: 'team-a-*'
    server: https://kubernetes.default.svc
  
  # Restrict resource types
  clusterResourceWhitelist:
  - group: ''
    kind: Namespace
  namespaceResourceWhitelist:
  - group: 'apps'
    kind: Deployment
  - group: 'apps'
    kind: StatefulSet
  - group: ''
    kind: Service
  - group: ''
    kind: ConfigMap
  - group: ''
    kind: Secret
  
  # Deny deploying certain resources
  namespaceResourceBlacklist:
  - group: ''
    kind: ResourceQuota
  - group: ''
    kind: LimitRange
  
  # RBAC roles
  roles:
  - name: developer
    description: Developers can sync applications
    policies:
    - p, proj:team-a:developer, applications, get, team-a/*, allow
    - p, proj:team-a:developer, applications, sync, team-a/*, allow
    groups:
    - team-a-developers
  
  - name: admin
    description: Admins have full access
    policies:
    - p, proj:team-a:admin, applications, *, team-a/*, allow
    groups:
    - team-a-admins
```

### 2. Cluster-based Isolation

```yaml
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: staging
  namespace: argocd
spec:
  description: Staging Environment
  
  sourceRepos:
  - '*'
  
  # Only deploy to staging cluster
  destinations:
  - namespace: '*'
    server: https://staging-cluster.example.com
  
  # Sync windows
  syncWindows:
  - kind: allow
    schedule: '0 9-17 * * 1-5'  # Weekdays 9am-5pm
    duration: 8h
    applications:
    - '*'
    manualSync: true
  
  - kind: deny
    schedule: '0 0-8,18-23 * * *'  # Outside business hours
    duration: 10h
    applications:
    - '*'
    manualSync: false
```

### 3. Resource Quotas per Project

```yaml
# Namespace for team
apiVersion: v1
kind: Namespace
metadata:
  name: team-a-prod

---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-a-quota
  namespace: team-a-prod
spec:
  hard:
    requests.cpu: "100"
    requests.memory: 200Gi
    persistentvolumeclaims: "50"
    pods: "100"

---
apiVersion: v1
kind: LimitRange
metadata:
  name: team-a-limits
  namespace: team-a-prod
spec:
  limits:
  - max:
      cpu: "4"
      memory: 8Gi
    min:
      cpu: 100m
      memory: 128Mi
    default:
      cpu: "1"
      memory: 1Gi
    defaultRequest:
      cpu: 500m
      memory: 512Mi
    type: Container
```

## Useful Scripts

### 1. Bulk Application Sync

```bash
#!/bin/bash
# sync-all-apps.sh - Sync all applications in a project

PROJECT="myproject"

# Get all apps in project
APPS=$(argocd app list --project $PROJECT -o name)

echo "Syncing applications in project: $PROJECT"

for APP in $APPS; do
  echo "Syncing: $APP"
  argocd app sync $APP --async
done

echo "Waiting for syncs to complete..."
for APP in $APPS; do
  argocd app wait $APP --health
done

echo "All applications synced and healthy!"
```

### 2. Application Health Check

```bash
#!/bin/bash
# check-app-health.sh - Check health of all applications

argocd app list -o json | jq -r '.[] | 
  select(.status.health.status != "Healthy") | 
  "\(.metadata.name): \(.status.health.status) - \(.status.health.message)"'
```

### 3. Find OutOfSync Resources

```bash
#!/bin/bash
# find-outofsync.sh - Find all OutOfSync resources

APP=$1
argocd app diff $APP --local-repo-root . | grep -A 5 "====="
```

## Best Practices Summary

1. **Use App of Apps Pattern**: Break large applications into smaller, manageable apps
2. **Implement Sync Waves**: Control deployment order with `argocd.argoproj.io/sync-wave`
3. **Use Resource Hooks**: Run jobs before/after sync for migrations, tests
4. **Enable Auto-Sync Carefully**: Start with manual sync, enable auto-sync after validation
5. **Configure Ignore Differences**: Avoid false OutOfSync status for expected changes
6. **Use AppProjects**: Implement RBAC and resource restrictions per team/environment
7. **Monitor ArgoCD**: Set up Prometheus metrics and alerting
8. **Regular Backups**: Backup ArgoCD configuration and cluster secrets
9. **Use Notifications**: Configure alerts for sync failures and degraded health
10. **Document Custom Resources**: Add health checks and action configurations for CRDs

## Additional Resources

- [ArgoCD Troubleshooting Guide](https://argo-cd.readthedocs.io/en/stable/operator-manual/troubleshooting/)
- [Performance Tuning](https://argo-cd.readthedocs.io/en/stable/operator-manual/high_availability/)
- [Resource Hooks](https://argo-cd.readthedocs.io/en/stable/user-guide/resource_hooks/)
- [Argo Rollouts](https://argo-rollouts.readthedocs.io/)
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
