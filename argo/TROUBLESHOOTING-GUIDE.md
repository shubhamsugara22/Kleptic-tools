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
