# ArgoCD Documentation Index

Complete reference guide for ArgoCD setup, deployment, and management in Kubernetes environments.

## 🚀 Getting Started

### Quick Start
1. Read [README.md](README.md) for basic concepts and setup
2. Use [setup.sh](setup.sh) (Linux/macOS) or [setup.ps1](setup.ps1) (Windows) to install ArgoCD
3. Check [CHEATSHEET.md](CHEATSHEET.md) for quick command reference

### First Application
```bash
# Install ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Access UI
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Get password
argocd admin initial-password -n argocd

# Deploy your first app using a template
kubectl apply -f basic-app.yaml
```

## 📚 Documentation Structure

### Core Documentation

#### [README.md](README.md) - Main Reference Guide
**When to use**: Starting point for understanding ArgoCD
- ArgoCD overview and key concepts
- 12 key features with examples
- 10 application templates (inline)
- Basic installation and setup
- Command reference
- Architecture overview
- Best practices

**Topics covered**:
- Application, AppProject, ApplicationSet concepts
- Automated sync and self-healing
- Multi-cluster management
- Health status monitoring
- Sync hooks and waves
- RBAC and SSO integration
- Helm, Kustomize, Jsonnet support
- Progressive delivery
- Resource hooks
- Application dependencies

#### [CHEATSHEET.md](CHEATSHEET.md) - Quick Reference
**When to use**: Looking for specific commands or quick syntax
- Installation commands
- Login and authentication
- Application CRUD operations
- Repository and cluster management
- kubectl commands for ArgoCD resources
- Manifest snippets
- Quick troubleshooting

**Command categories**:
- Installation (kubectl, Helm, local)
- Login and context
- App management (create, get, list, sync, delete, rollback)
- Repository operations
- Cluster operations
- Debugging commands

#### [PRODUCTION-GUIDE.md](PRODUCTION-GUIDE.md) - Enterprise Deployment
**When to use**: Preparing ArgoCD for production environments
- High availability setup
- Security hardening
- Monitoring and observability
- Backup and disaster recovery
- Performance tuning
- Upgrade and maintenance

**Topics covered**:
- Multi-replica deployments (application controller, repo server, server, Redis)
- Redis HA with Sentinel and HAProxy
- Horizontal Pod Autoscaling
- Network policies and pod security
- RBAC configuration
- Image scanning and admission control
- OPA/Gatekeeper policies
- Prometheus metrics and ServiceMonitors
- Grafana dashboards
- Log aggregation (ELK/Loki)
- Alerting rules
- Backup automation (CronJob)
- Disaster recovery procedures
- Controller sharding
- Resource caching and exclusions
- Rolling upgrades

#### [INTEGRATION-GUIDE.md](INTEGRATION-GUIDE.md) - CI/CD and Integrations
**When to use**: Integrating ArgoCD with CI/CD pipelines and external tools
- CI/CD pipeline integration
- Secrets management solutions
- Image updater configuration
- Notifications setup

**Topics covered**:
- GitHub Actions workflows
- GitLab CI pipelines
- Jenkins pipelines
- Azure DevOps pipelines
- Sealed Secrets setup and usage
- External Secrets Operator (AWS, Azure)
- HashiCorp Vault integration
- SOPS encryption
- ArgoCD Image Updater (semver, digest, latest)
- Slack notifications
- Microsoft Teams notifications
- Email notifications

#### [TROUBLESHOOTING-GUIDE.md](TROUBLESHOOTING-GUIDE.md) - Advanced Topics
**When to use**: Debugging issues or implementing advanced patterns
- Common issues and solutions
- Debugging techniques
- Performance optimization
- Advanced deployment patterns
- Resource hooks
- Multi-tenancy

**Topics covered**:
- OutOfSync troubleshooting
- Sync operation debugging
- Custom health checks
- Repository connection issues
- RBAC permission errors
- Debug logging
- Application events
- Resource tracking
- Manifest generation debugging
- Controller and repo server tuning
- Resource exclusions
- Progressive delivery with Argo Rollouts
- Blue-green deployments
- Multi-source applications
- ApplicationSet generators
- PreSync, Sync, PostSync, SyncFail hooks
- Namespace-based isolation
- Cluster-based isolation
- Resource quotas

## 📁 Template Files

### Application Templates

| Template | Use Case | Complexity |
|----------|----------|------------|
| [basic-app.yaml](basic-app.yaml) | Simple Git-based deployment | Beginner |
| [helm-app.yaml](helm-app.yaml) | Helm chart with custom values | Intermediate |
| [kustomize-app.yaml](kustomize-app.yaml) | Kustomize overlays | Intermediate |
| [app-of-apps.yaml](app-of-apps.yaml) | Managing multiple apps | Advanced |
| [app-project.yaml](app-project.yaml) | Project with RBAC | Advanced |
| [applicationset.yaml](applicationset.yaml) | GitOps at scale | Advanced |
| [sync-waves-example.yaml](sync-waves-example.yaml) | Ordered deployment | Advanced |

### Template Quick Reference

```bash
# Basic application
kubectl apply -f basic-app.yaml

# Helm application with custom values
kubectl apply -f helm-app.yaml

# Kustomize overlay
kubectl apply -f kustomize-app.yaml

# App of Apps pattern
kubectl apply -f app-of-apps.yaml

# Create AppProject with RBAC
kubectl apply -f app-project.yaml

# ApplicationSet for multiple environments
kubectl apply -f applicationset.yaml

# Deployment with hooks and sync waves
kubectl apply -f sync-waves-example.yaml
```

## 🔧 Setup Scripts

### [setup.sh](setup.sh) - Bash Installation Script
**Platforms**: Linux, macOS, WSL
```bash
chmod +x setup.sh
./setup.sh
```

**Features**:
- Interactive prompts for method and version
- kubectl/Helm installation
- Port forwarding setup
- Initial password retrieval
- ArgoCD CLI installation
- Login automation

### [setup.ps1](setup.ps1) - PowerShell Installation Script
**Platforms**: Windows, PowerShell Core
```powershell
.\setup.ps1
```

**Features**:
- Windows-native installation
- Kubectl installation via Chocolatey
- Helm 3 installation
- Port forwarding with job management
- Credential management
- CLI installation with PATH setup

## 🎯 Use Case Navigation

### I want to...

#### Learn ArgoCD Basics
1. Read [README.md - Overview](README.md#overview)
2. Understand [Key Concepts](README.md#key-concepts)
3. Try [Basic Setup](README.md#basic-setup)
4. Deploy using [basic-app.yaml](basic-app.yaml)

#### Install ArgoCD
1. Check [Prerequisites](README.md#prerequisites)
2. Run [setup.sh](setup.sh) or [setup.ps1](setup.ps1)
3. Access UI and CLI per [Installation](README.md#installation)
4. Refer to [CHEATSHEET.md - Installation](CHEATSHEET.md#installation)

#### Deploy My First Application
1. Review [Application Templates](README.md#application-templates)
2. Choose appropriate template ([basic-app.yaml](basic-app.yaml) for simple apps)
3. Customize the template
4. Apply: `kubectl apply -f <template>.yaml`
5. Monitor: `argocd app get myapp`

#### Set Up Production Environment
1. Read entire [PRODUCTION-GUIDE.md](PRODUCTION-GUIDE.md)
2. Implement [High Availability](PRODUCTION-GUIDE.md#high-availability)
3. Configure [Security Hardening](PRODUCTION-GUIDE.md#security-hardening)
4. Set up [Monitoring](PRODUCTION-GUIDE.md#monitoring--observability)
5. Establish [Backup Procedures](PRODUCTION-GUIDE.md#backup--disaster-recovery)

#### Integrate with CI/CD
1. Choose your CI/CD platform in [INTEGRATION-GUIDE.md](INTEGRATION-GUIDE.md)
2. Implement pipeline (GitHub Actions, GitLab CI, Jenkins, Azure DevOps)
3. Configure automated sync
4. Set up [Notifications](INTEGRATION-GUIDE.md#notifications-configuration)

#### Manage Secrets
1. Review [Secrets Management](INTEGRATION-GUIDE.md#secrets-management)
2. Choose solution: Sealed Secrets, External Secrets Operator, Vault, or SOPS
3. Implement per guide
4. Test secret sync

#### Implement Progressive Delivery
1. Read [Advanced Patterns](TROUBLESHOOTING-GUIDE.md#advanced-patterns)
2. Install Argo Rollouts
3. Create Rollout manifest with canary/blue-green strategy
4. Configure AnalysisTemplates for automated decisions
5. Monitor rollout progress

#### Troubleshoot Issues
1. Check [Common Issues](TROUBLESHOOTING-GUIDE.md#common-issues)
2. Use [Debugging Techniques](TROUBLESHOOTING-GUIDE.md#debugging-techniques)
3. Review logs: `kubectl logs -n argocd deployment/argocd-application-controller`
4. Check app status: `argocd app get myapp`

#### Optimize Performance
1. Review [Performance Optimization](TROUBLESHOOTING-GUIDE.md#performance-optimization)
2. Tune application controller
3. Scale repo server
4. Configure resource exclusions
5. Monitor with Prometheus

#### Implement Multi-Tenancy
1. Read [Multi-Tenancy](TROUBLESHOOTING-GUIDE.md#multi-tenancy)
2. Create AppProjects per team
3. Configure RBAC roles
4. Set up resource quotas
5. Implement sync windows

## 📊 Feature Matrix

| Feature | README | CHEATSHEET | PRODUCTION | INTEGRATION | TROUBLESHOOTING |
|---------|--------|------------|------------|-------------|-----------------|
| Basic Concepts | ✅ | ⚠️ | ❌ | ❌ | ❌ |
| Installation | ✅ | ✅ | ⚠️ | ❌ | ❌ |
| Commands | ⚠️ | ✅ | ❌ | ❌ | ❌ |
| Templates | ✅ | ⚠️ | ❌ | ❌ | ❌ |
| High Availability | ❌ | ❌ | ✅ | ❌ | ❌ |
| Security | ⚠️ | ❌ | ✅ | ⚠️ | ❌ |
| Monitoring | ⚠️ | ❌ | ✅ | ❌ | ⚠️ |
| Backup/DR | ❌ | ❌ | ✅ | ❌ | ❌ |
| CI/CD | ❌ | ❌ | ❌ | ✅ | ❌ |
| Secrets | ❌ | ❌ | ⚠️ | ✅ | ❌ |
| Image Updater | ⚠️ | ❌ | ❌ | ✅ | ❌ |
| Notifications | ⚠️ | ❌ | ❌ | ✅ | ❌ |
| Troubleshooting | ⚠️ | ⚠️ | ❌ | ❌ | ✅ |
| Performance | ❌ | ❌ | ✅ | ❌ | ✅ |
| Advanced Patterns | ⚠️ | ❌ | ❌ | ❌ | ✅ |
| Multi-Tenancy | ⚠️ | ❌ | ❌ | ❌ | ✅ |

Legend: ✅ Comprehensive | ⚠️ Basic/Referenced | ❌ Not Covered

## 🔍 Quick Search

### By Technology
- **Kubernetes**: All documents
- **Helm**: README, helm-app.yaml, INTEGRATION-GUIDE
- **Kustomize**: README, kustomize-app.yaml
- **Prometheus**: PRODUCTION-GUIDE, TROUBLESHOOTING-GUIDE
- **Grafana**: PRODUCTION-GUIDE
- **Redis**: PRODUCTION-GUIDE
- **GitHub Actions**: INTEGRATION-GUIDE
- **GitLab CI**: INTEGRATION-GUIDE
- **Jenkins**: INTEGRATION-GUIDE
- **Azure DevOps**: INTEGRATION-GUIDE
- **Vault**: INTEGRATION-GUIDE
- **Sealed Secrets**: INTEGRATION-GUIDE
- **Argo Rollouts**: TROUBLESHOOTING-GUIDE

### By Task
- **Installation**: README, CHEATSHEET, setup scripts
- **Deployment**: README, all template files
- **Monitoring**: PRODUCTION-GUIDE
- **Security**: PRODUCTION-GUIDE
- **Debugging**: TROUBLESHOOTING-GUIDE
- **Automation**: INTEGRATION-GUIDE
- **Scaling**: PRODUCTION-GUIDE, TROUBLESHOOTING-GUIDE

## 📈 Learning Path

### Beginner Path (1-2 days)
1. Read [README.md](README.md) - Understand concepts (30 min)
2. Run installation script (15 min)
3. Deploy [basic-app.yaml](basic-app.yaml) (15 min)
4. Explore UI and [CHEATSHEET.md](CHEATSHEET.md) (1 hour)
5. Try [helm-app.yaml](helm-app.yaml) (30 min)
6. Practice common commands (1 hour)

### Intermediate Path (3-5 days)
1. Complete Beginner Path
2. Deploy [kustomize-app.yaml](kustomize-app.yaml) (1 hour)
3. Implement [app-of-apps.yaml](app-of-apps.yaml) (2 hours)
4. Configure [app-project.yaml](app-project.yaml) with RBAC (2 hours)
5. Set up [sync-waves-example.yaml](sync-waves-example.yaml) (2 hours)
6. Review [INTEGRATION-GUIDE.md - CI/CD](INTEGRATION-GUIDE.md#cicd-integration) (3 hours)
7. Implement secrets management (3 hours)

### Advanced Path (1-2 weeks)
1. Complete Intermediate Path
2. Study [PRODUCTION-GUIDE.md](PRODUCTION-GUIDE.md) entirely (4 hours)
3. Implement HA setup (1 day)
4. Configure monitoring and alerting (1 day)
5. Set up backup/DR (4 hours)
6. Implement [applicationset.yaml](applicationset.yaml) at scale (1 day)
7. Study [TROUBLESHOOTING-GUIDE.md](TROUBLESHOOTING-GUIDE.md) (4 hours)
8. Implement progressive delivery with Rollouts (1 day)
9. Configure multi-tenancy (1 day)

## 🛠️ Common Commands Quick Access

```bash
# Install ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Access UI
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Login CLI
argocd login localhost:8080

# Create application from template
kubectl apply -f basic-app.yaml

# Check application status
argocd app get myapp

# Sync application
argocd app sync myapp

# View application diff
argocd app diff myapp

# Monitor sync status
argocd app wait myapp --health

# Delete application
argocd app delete myapp
```

## 🆘 Emergency Procedures

### Application Won't Sync
1. Check [Troubleshooting - OutOfSync](TROUBLESHOOTING-GUIDE.md#1-application-outofsync)
2. View diff: `argocd app diff myapp`
3. Check logs: `kubectl logs -n argocd deployment/argocd-application-controller`
4. Force sync: `argocd app sync myapp --force`

### ArgoCD Down
1. Check [Production Guide - Backup & DR](PRODUCTION-GUIDE.md#backup--disaster-recovery)
2. Verify pods: `kubectl get pods -n argocd`
3. Check events: `kubectl get events -n argocd`
4. Review logs of failed pods
5. Restore from backup if needed

### Performance Issues
1. Review [Troubleshooting - Performance](TROUBLESHOOTING-GUIDE.md#performance-optimization)
2. Check resource usage: `kubectl top pods -n argocd`
3. Scale components: `kubectl scale deployment argocd-repo-server -n argocd --replicas=3`
4. Tune controller processors
5. Implement resource exclusions

## 📞 Additional Resources

### Official Documentation
- [ArgoCD Official Docs](https://argo-cd.readthedocs.io/)
- [Argo Rollouts](https://argo-rollouts.readthedocs.io/)
- [ArgoCD GitHub](https://github.com/argoproj/argo-cd)

### Community
- [ArgoCD Slack](https://argoproj.github.io/community/join-slack)
- [GitHub Discussions](https://github.com/argoproj/argo-cd/discussions)
- [YouTube Tutorials](https://www.youtube.com/c/ArgoprojCommunity)

### Related Tools
- Helm: `https://helm.sh/`
- Kustomize: `https://kustomize.io/`
- Sealed Secrets: `https://github.com/bitnami-labs/sealed-secrets`
- External Secrets: `https://external-secrets.io/`

---

**Last Updated**: Check file timestamps for latest updates

**Maintainer**: Kleptic Tools Team

**Feedback**: Submit issues or suggestions to improve this documentation
