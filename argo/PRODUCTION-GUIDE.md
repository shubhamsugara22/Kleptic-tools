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
