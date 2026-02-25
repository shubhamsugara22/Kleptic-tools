# Istio Cheatsheet

Quick reference for common Istio operations and commands.

## Installation & Uninstallation

```bash
# Install Istio with demo profile
istioctl install --set profile=demo -y

# Install with production profile (minimal)
istioctl install --set profile=production -y

# Install with specific version
istioctl install --set profile=demo --set hub=docker.io/istio --set tag=1.20.0 -y

# Uninstall Istio
istioctl uninstall --purge

# List installed Istio version
istioctl version
```

## Namespace Management

```bash
# Enable sidecar injection for namespace
kubectl label namespace <namespace> istio-injection=enabled

# Disable sidecar injection
kubectl label namespace <namespace> istio-injection=disabled

# Check if injection is enabled
kubectl get namespace <namespace> --show-labels

# Verify sidecars are injected
kubectl get pods -n <namespace> -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[*].name}{"\n"}{end}'
```

## Configuration Management

```bash
# Validate Istio configuration
istioctl analyze

# Validate specific namespace
istioctl analyze -n <namespace>

# List all VirtualServices
kubectl get virtualservices -A
kubectl get vs -A

# List all DestinationRules
kubectl get destinationrules -A
kubectl get dr -A

# List all Gateways
kubectl get gateways -A
kubectl get gw -A

# List all AuthorizationPolicies
kubectl get authorizationpolicies -A
kubectl get ap -A

# List all PeerAuthentications
kubectl get peerauthentication -A
kubectl get pa -A

# Describe a resource
kubectl describe vs <name> -n <namespace>
kubectl describe dr <name> -n <namespace>
kubectl describe ap <name> -n <namespace>
```

## Traffic Management

```bash
# Apply traffic configuration
kubectl apply -f virtualservice.yaml

# Edit VirtualService
kubectl edit vs <name> -n <namespace>

# Delete VirtualService
kubectl delete vs <name> -n <namespace>

# View route configuration
kubectl get vs <name> -n <namespace> -o yaml

# Test route
kubectl exec <pod> -c <container> -- curl http://<service>:<port>/path
```

## Security & mTLS

```bash
# Enable mTLS for namespace
kubectl apply -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
spec:
  mtls:
    mode: STRICT
EOF

# Check mTLS status
kubectl get peerauthentication -A

# Enable PERMISSIVE mode (for migration)
kubectl apply -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
spec:
  mtls:
    mode: PERMISSIVE
EOF

# Create authorization policy
kubectl apply -f authorization-policy.yaml

# View authorization policies
kubectl get ap -A

# Test authorization
kubectl exec <pod> -c <container> -- curl -I http://<service>:<port>
```

## Debugging & Troubleshooting

```bash
# Check control plane components
kubectl get pods -n istio-system

# Check logs of Istiod
kubectl logs -n istio-system -l app=istiod -f

# Check logs of ingress gateway
kubectl logs -n istio-system -l app=istio-ingressgateway -f

# View sidecar proxy configuration
istioctl proxy-config routes <pod> -n <namespace>

# View virtual services in service
istioctl proxy-config virtualservice <pod> -n <namespace>

# Dump complete proxy configuration
istioctl proxy-config all <pod> -n <namespace>

# Check listeners
istioctl proxy-config listener <pod> -n <namespace>

# Check clusters (upstreams)
istioctl proxy-config cluster <pod> -n <namespace>

# Check endpoints
istioctl proxy-config endpoints <pod> -n <namespace>

# View sidecar proxy logs
kubectl logs <pod> -c istio-proxy -n <namespace>

# Follow sidecar logs
kubectl logs -f <pod> -c istio-proxy -n <namespace>

# Check proxy sync status
istioctl proxy-status

# Check if pod is injected
kubectl get pod <pod-name> -n <namespace> -o jsonpath='{.spec.containers[*].name}'
```

## Port Forwarding to Access Dashboards

```bash
# Kiali (Service Mesh UI)
kubectl port-forward svc/kiali -n istio-system 20000:20000
# Access: http://localhost:20000

# Grafana (Metrics)
kubectl port-forward svc/grafana -n istio-system 3000:3000
# Access: http://localhost:3000

# Prometheus (Metrics Database)
kubectl port-forward svc/prometheus -n istio-system 9090:9090
# Access: http://localhost:9090

# Jaeger (Tracing)
kubectl port-forward svc/jaeger -n istio-system 16686:16686
# Access: http://localhost:16686

# Grafana Loki (Logs)
kubectl port-forward svc/loki -n istio-system 3100:3100
# Access: http://localhost:3100
```

## Sample Applications

```bash
# Deploy Bookinfo sample
kubectl apply -f $ISTIO_HOME/samples/bookinfo/platform/kube/bookinfo.yaml

# Create ingress gateway for Bookinfo
kubectl apply -f $ISTIO_HOME/samples/bookinfo/networking/bookinfo-gateway.yaml

# Get Ingress IP
kubectl get svc istio-ingressgateway -n istio-system

# Access Bookinfo
curl http://<INGRESS_IP>/productpage

# Cleanup Bookinfo
kubectl delete -f $ISTIO_HOME/samples/bookinfo/platform/kube/bookinfo.yaml
```

## Profile Comparison

| Feature | Default | Demo | Production | Ambient |
|---------|---------|------|-----------|---------|
| Core Components | ✓ | ✓ | ✓ | ✓ |
| Prometheus | ✗ | ✓ | ✗ | ✗ |
| Grafana | ✗ | ✓ | ✗ | ✗ |
| Kiali | ✗ | ✓ | ✗ | ✗ |
| Jaeger | ✗ | ✓ | ✗ | ✗ |
| Tracing | ✗ | ✓ | ✗ | ✗ |
| Use Case | Production | Learning/Demo | Production Minimal | eBPF-based |

## Common Issues & Solutions

### Issue: Pods not injected with sidecars
```bash
# Check namespace label
kubectl get namespace -L istio-injection

# Check injection webhook
kubectl get mutatingwebhookconfigurations | grep istio

# Restart pods to trigger injection
kubectl rollout restart deployment/<name> -n <namespace>
```

### Issue: Communication failing between services
```bash
# Check authorization policies
kubectl get ap -A

# Switch to PERMISSIVE mode temporarily
kubectl apply -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: <namespace>
spec:
  mtls:
    mode: PERMISSIVE
EOF

# Analyze mesh
istioctl analyze -n <namespace>
```

### Issue: High latency or timeouts
```bash
# Check circuit breaker settings
kubectl get dr -n <namespace> -o yaml

# Check timeout settings in VirtualService
kubectl get vs -n <namespace> -o yaml

# View latency metrics in Grafana
```

### Issue: Gateway not routing traffic
```bash
# Check gateway status
kubectl get gateway -n <namespace>

# Check VirtualService
kubectl get vs -n <namespace>

# Test URL
curl -I http://<INGRESS_IP>/path

# Check ingress gateway logs
kubectl logs -n istio-system -l app=istio-ingressgateway
```

## Performance Tuning

```bash
# Increase resource limits for Istiod
kubectl patch deployment istiod -n istio-system --patch '
spec:
  template:
    spec:
      containers:
      - name: discovery
        resources:
          limits:
            cpu: 2
            memory: 2Gi
          requests:
            cpu: 1
            memory: 1Gi
'

# Update connection pool size
kubectl edit dr <name> -n <namespace>
# Increase maxConnections and http1MaxPendingRequests
```

## Custom Resource Definitions (CRDs)

```bash
# List all Istio CRDs
kubectl get crd -o name | grep istio

# Get CRD details
kubectl explain virtualservice.spec

# Get CRD properties
kubectl explain virtualservice.spec.http
```

## YAML Examples Quick Links

- Virtual Services & Destination Rules: See `traffic-management.md`
- Security Policies: See `security-policies.md`
- Observability: See `observability.md`

## Useful Resources

- [Official Istio Docs](https://istio.io/latest/docs/)
- [Istio Blog](https://istio.io/latest/blog/)
- [Community Discussion](https://discuss.istio.io/)
- [GitHub Issues](https://github.com/istio/istio/issues)
