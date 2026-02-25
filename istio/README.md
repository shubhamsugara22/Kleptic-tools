# Istio Service Mesh

Istio is an open-source service mesh that provides a uniform way to manage, secure, and observe microservices. It decouples microservices from network and security concerns by handling service-to-service communication.

## What is Istio?

Istio leverages Envoy proxies as sidecars deployed alongside your services to manage all inbound and outbound network communication. The control plane manages these proxies and provides:

- **Traffic Management**: Advanced routing, load balancing, failover, and retries
- **Security**: Mutual TLS (mTLS), authorization policies, and certificate management
- **Observability**: Distributed tracing, metrics, and logs through integration with Prometheus, Grafana, and Jaeger

## Prerequisites

- Kubernetes cluster (1.22+)
- `kubectl` configured to access your cluster
- Minimum 2GB per pod for data plane
- For Windows: WSL2 with kubectl installed

## Quick Start

### Linux/macOS

```bash
cd istio
./setup.sh
```

### Windows

```powershell
cd istio
.\setup.ps1
```

## Installation Steps

1. **Install Istio Control Plane**
   - Downloads and installs the latest Istio release
   - Creates the `istio-system` namespace
   - Deploys Istiod (control plane)

2. **Deploy Sample Applications** (Optional)
   - Install the Bookinfo sample application to explore Istio features

3. **Verify Installation**
   - Check control plane components
   - Verify sidecar injection is enabled

## Core Components

### Control Plane (Istiod)
- **Pilot**: Manages service discovery and traffic routing
- **Citadel**: Manages certificates and security policies
- **Galley**: Validates and distributes configuration

### Data Plane
- **Envoy Proxies**: Deployed as sidecars alongside application containers
- Handles all network traffic between services

## Key Features

### Traffic Management
- Virtual Services and Destination Rules for fine-grained traffic control
- Gateway for managing inbound/outbound traffic
- Service Entries for external services

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: example
spec:
  hosts:
  - example
  http:
  - match:
    - uri:
        prefix: "/api"
    route:
    - destination:
        host: example
        port:
          number: 8080
```

### Security
- Automatic mTLS between services
- Authorization policies for fine-grained access control

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: deny-all
spec:
  rules: []
```

### Observability
- Automatic request metrics (latency, error rates)
- Distributed tracing with Jaeger
- Service graph visualization with Kiali

## Sidecar Injection

Enable automatic sidecar injection for a namespace:

```bash
kubectl label namespace <namespace> istio-injection=enabled
```

Verify injection:

```bash
kubectl get namespace <namespace> --show-labels
```

## Monitoring & Visualization

### Access Kiali Dashboard (Service Mesh UI)
```bash
kubectl port-forward svc/kiali -n istio-system 20000:20000
# Open http://localhost:20000
```

### Access Grafana (Metrics)
```bash
kubectl port-forward svc/grafana -n istio-system 3000:3000
# Open http://localhost:3000
```

### Access Jaeger (Distributed Tracing)
```bash
kubectl port-forward svc/jaeger -n istio-system 16686:16686
# Open http://localhost:16686
```

## Common Use Cases

### Canary Deployments
Route a percentage of traffic to a new version while monitoring metrics.

### Rate Limiting
Apply rate limiting rules based on traffic patterns.

### Circuit Breaker
Automatically fail requests to unhealthy upstream services.

## Uninstall

```bash
istioctl uninstall --purge
# or
kubectl delete namespace istio-system
```

## Resources

- [Official Istio Documentation](https://istio.io/latest/docs/)
- [Istio Architecture](https://istio.io/latest/docs/ops/deployment/architecture/)
- [Task Examples](https://istio.io/latest/docs/tasks/)
- [Troubleshooting Guide](https://istio.io/latest/docs/ops/troubleshooting/)

## Additional Configuration Files

See the configuration examples in this directory for:
- Virtual Services and Destination Rules
- Gateway configurations
- Authorization policies
- Service Entries for external services
