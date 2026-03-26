# ArgoCD - Declarative GitOps Continuous Deployment

> **📖 [ARGOCD-COMPLETE-GUIDE.md](ARGOCD-COMPLETE-GUIDE.md)** - Complete reference guide with everything you need

ArgoCD is a declarative, GitOps continuous delivery tool for Kubernetes. It automates the deployment of applications to Kubernetes clusters by using Git repositories as the single source of truth.

## Quick Start

```bash
# 1. Install ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# 2. Access UI
kubectl port-forward svc/argocd-server -n argocd 8080:443

# 3. Get initial password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# 4. Login (username: admin, password: from step 3)
argocd login localhost:8080 --insecure

# 5. Deploy your first app
kubectl apply -f basic-app.yaml
```

## Documentation

### 📖 Complete Guide
**[ARGOCD-COMPLETE-GUIDE.md](ARGOCD-COMPLETE-GUIDE.md)** - Comprehensive guide covering:
- Basic concepts and features
- Installation and setup
- Command reference (CLI cheat sheet)
- Production deployment (HA, security, monitoring)
- CI/CD integration (GitHub Actions, GitLab CI, Jenkins, Azure DevOps)
- Secrets management (Sealed Secrets, Vault, External Secrets)
- Image updater configuration
- Notifications setup (Slack, Teams, Email)
- Troubleshooting and debugging
- Performance optimization
- Advanced patterns (Argo Rollouts, blue-green, canary)
- Multi-tenancy configuration

### 🚀 Day 1 Baseline (Implemented)
- **[platform/day1/README.md](platform/day1/README.md)** - Runbook for project boundaries and RBAC bootstrap
- **[platform/day1/projects/appprojects.yaml](platform/day1/projects/appprojects.yaml)** - `platform-core`, `team-dev`, `team-prod` AppProjects
- **[platform/day1/rbac/argocd-rbac-cm.yaml](platform/day1/rbac/argocd-rbac-cm.yaml)** - Group-to-role mappings
- **[platform/day1/apps/root-app.yaml](platform/day1/apps/root-app.yaml)** - Parent app-of-apps bootstrap

Apply once in order:

```bash
kubectl apply -f argo/platform/day1/projects/appprojects.yaml
kubectl apply -f argo/platform/day1/rbac/argocd-rbac-cm.yaml
kubectl apply -f argo/platform/day1/apps/root-app.yaml
```

### Day 2 Baseline (Started)
- **[platform/day2/README.md](platform/day2/README.md)** - Runbook for environment fan-out
- **[platform/day2/apps/env-applicationset.yaml](platform/day2/apps/env-applicationset.yaml)** - One template generates `demo-dev`, `demo-stage`, `demo-prod`
- **[platform/day1/apps/children/day2-bootstrap.yaml](platform/day1/apps/children/day2-bootstrap.yaml)** - Day 2 bootstrap child app

Behavior:
- `dev`: auto-sync enabled (`prune` + `selfHeal`)
- `stage`: manual sync
- `prod`: manual sync

Direct apply (optional):

```bash
kubectl apply -f argo/platform/day2/apps/env-applicationset.yaml
```

### 🛠️ Setup Scripts
- **[setup.sh](setup.sh)** - Bash installation script for Linux/macOS
- **[setup.ps1](setup.ps1)** - PowerShell installation script for Windows

### 📁 Application Templates
| Template | Use Case |
|----------|----------|
| [basic-app.yaml](basic-app.yaml) | Simple Git-based deployment |
| [helm-app.yaml](helm-app.yaml) | Helm chart with custom values |
| [kustomize-app.yaml](kustomize-app.yaml) | Kustomize overlays |
| [app-of-apps.yaml](app-of-apps.yaml) | Managing multiple apps |
| [app-project.yaml](app-project.yaml) | Project with RBAC |
| [applicationset.yaml](applicationset.yaml) | GitOps at scale |
| [sync-waves-example.yaml](sync-waves-example.yaml) | Ordered deployment with hooks |

## What is ArgoCD?

ArgoCD provides:
- **Declarative setup**: Define applications in Git
- **Automated sync**: Continuous deployment based on Git changes
- **Multi-cluster support**: Manage multiple Kubernetes clusters from one place
- **UI Dashboard**: Visual application status monitoring
- **RBAC & SSO**: Fine-grained access control and enterprise authentication
- **Rollback**: Easy rollback to previous application states
- **Health monitoring**: Continuous health checks for deployed applications

## Core Concepts

- **Application**: A group of Kubernetes resources defined by a manifest in Git
- **Project**: Logical grouping of applications with access restrictions
- **Repository**: Git repository containing application manifests
- **Sync**: Process of applying Git state to the Kubernetes cluster
- **Sync Status**: Indicates if live state matches desired state (Synced/OutOfSync)
- **Health Status**: Indicates resource health (Healthy/Progressing/Degraded/Missing)

## Key Features

✅ **GitOps Best Practices** - Git as single source of truth  
✅ **Automated Sync** - Auto-deploy on Git changes  
✅ **Multi-Source Support** - YAML, Helm, Kustomize, Jsonnet  
✅ **Multi-Cluster** - Manage multiple K8s clusters  
✅ **RBAC** - Fine-grained permission control  
✅ **SSO Integration** - GitHub, GitLab, LDAP, OIDC  
✅ **Rollback** - Easy rollback to any previous version  
✅ **Health Monitoring** - Continuous health checks  
✅ **Sync Waves** - Control deployment order  
✅ **Resource Hooks** - PreSync, Sync, PostSync, SyncFail hooks  
✅ **App of Apps** - Manage multiple applications as one  
✅ **ApplicationSets** - Generate apps dynamically  

## Prerequisites

- Kubernetes cluster 1.16+
- kubectl configured
- Git repository with K8s manifests
- Optional: Helm 3+, Kustomize, ArgoCD CLI

## Common Commands

```bash
# Application operations
argocd app list                    # List all applications
argocd app get <APP>               # Get application details
argocd app create <APP>            # Create application
argocd app sync <APP>              # Sync application
argocd app delete <APP>            # Delete application
argocd app rollback <APP> <REV>    # Rollback to revision

# Repository management
argocd repo list                   # List repositories
argocd repo add <REPO_URL>         # Add repository

# Cluster management
argocd cluster list                # List clusters
argocd cluster add <CONTEXT>       # Add cluster

# Status and monitoring
argocd app get <APP> --refresh     # Refresh and show status
argocd app wait <APP> --health     # Wait for healthy state
argocd app diff <APP>              # Show differences
```

## Example Application

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
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

## Learning Path

1. **Beginner** (1-2 hours)
   - Read [Complete Guide - Basic Concepts](ARGOCD-COMPLETE-GUIDE.md#key-concepts)
   - Run [setup.sh](setup.sh) or [setup.ps1](setup.ps1)
   - Deploy [basic-app.yaml](basic-app.yaml)

2. **Intermediate** (3-5 days)
   - Explore [Helm](helm-app.yaml) and [Kustomize](kustomize-app.yaml) templates
   - Set up [App of Apps pattern](app-of-apps.yaml)
   - Configure [RBAC with AppProjects](app-project.yaml)
   - Review [CI/CD Integration section](ARGOCD-COMPLETE-GUIDE.md#cicd-integration)

3. **Advanced** (1-2 weeks)
   - Study [Production Guide section](ARGOCD-COMPLETE-GUIDE.md#production-guide---ha-security-monitoring)
   - Implement HA and monitoring
   - Configure secrets management
   - Set up [ApplicationSets](applicationset.yaml)
   - Explore [progressive delivery](ARGOCD-COMPLETE-GUIDE.md#progressive-delivery-with-argo-rollouts)

## Troubleshooting

Quick fixes for common issues:

```bash
# Application won't sync
argocd app diff <APP>              # Check differences
argocd app sync <APP> --force      # Force sync

# Check component health
kubectl get pods -n argocd         # Verify all pods running
kubectl logs -n argocd deployment/argocd-application-controller -f

# Reset admin password
kubectl -n argocd patch secret argocd-secret -p '{"stringData": {"admin.password": "newpassword"}}'
```

For detailed troubleshooting, see the [Troubleshooting section](ARGOCD-COMPLETE-GUIDE.md#troubleshooting--advanced-patterns) in the complete guide.

## Resources

- 📖 [Complete ArgoCD Guide](ARGOCD-COMPLETE-GUIDE.md)
- 🌐 [Official Documentation](https://argo-cd.readthedocs.io/)
- 💻 [GitHub Repository](https://github.com/argoproj/argo-cd)
- 💬 [Slack Community](https://argoproj.github.io/community/join-slack)
- 📹 [YouTube Channel](https://www.youtube.com/c/ArgoprojCommunity)

---

**Need Help?** Check the [Complete Guide](ARGOCD-COMPLETE-GUIDE.md) for detailed documentation on all topics.
