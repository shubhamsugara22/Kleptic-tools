# ArgoCD - Declarative GitOps Continuous Deployment

ArgoCD is a declarative, GitOps continuous delivery tool for Kubernetes. It automates the deployment of applications to Kubernetes clusters by using Git repositories as the single source of truth.

## Table of Contents

- [Overview](#overview)
- [Key Concepts](#key-concepts)
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
