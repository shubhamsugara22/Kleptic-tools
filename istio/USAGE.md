# Istio Traffic Handler - Usage Guide

A Go application for inspecting and managing Istio traffic configurations programmatically.

## Build

```bash
# Build the executable
go build -o traffic-handler traffic-handler.go
```

Windows:
```powershell
go build -o traffic-handler.exe traffic-handler.go
```

## Prerequisites

- Kubernetes cluster running
- Istio installed on the cluster
- `kubectl` configured to access your cluster
- `KUBECONFIG` environment variable set (or kubeconfig at `~/.kube/config`)

## Usage

### Run the Traffic Handler

By default, it operates on the `default` namespace:

```bash
# Linux/macOS
./traffic-handler

# Windows
.\traffic-handler.exe
```

To specify a different namespace:

```bash
# Linux/macOS
export NAMESPACE=production
./traffic-handler

# Windows PowerShell
$env:NAMESPACE="production"
.\traffic-handler.exe
```

### What It Does

The traffic handler connects to your Kubernetes cluster and displays:

1. **VirtualServices** - Traffic routing rules, hosts, and weight distribution
2. **DestinationRules** - Service subsets, versions, and traffic policies
3. **Gateways** - Ingress configuration and exposed services

### Sample Output

```
=== Istio Traffic Handler ===

=== Istio Traffic Management Status ===
Namespace: default

Virtual Services in namespace 'default':
==========================================
- Name: reviews
  Hosts: [reviews]
  Routes:
    -> reviews (subset: v1): 80%
    -> reviews (subset: v2): 20%

- Name: ratings
  Hosts: [ratings]
  Routes:
    -> ratings (subset: v1): 100%

Destination Rules in namespace 'default':
==========================================
- Name: reviews
  Host: reviews
  Subsets:
    - v1 (labels: map[version:v1])
    - v2 (labels: map[version:v2])

- Name: ratings
  Host: ratings
  Subsets:
    - v1 (labels: map[version:v1])

Gateways in namespace 'default':
==========================================
- Name: bookinfo-gateway
  Selector: map[istio:ingressgateway]
  Servers:
    - Port: 80 (HTTP)
      Hosts: [*]

=== Operations completed ===
```

## Available Functions

The `IstioTrafficHandler` struct provides these methods:

### Listing Resources

- `ListVirtualServices()` - List all VirtualServices with routing info
- `ListDestinationRules()` - List all DestinationRules with subsets
- `ListGateways()` - List all Gateways with server configuration
- `ShowStatus()` - Display complete Istio configuration overview

### Resource Details

- `GetVirtualService(name)` - Get detailed information about a specific VirtualService

### Resource Management

- `DeleteVirtualService(name)` - Delete a VirtualService
- `DeleteDestinationRule(name)` - Delete a DestinationRule

## Creating and Modifying Resources

This tool is designed to **inspect** Istio configurations. To create or modify resources, use `kubectl apply` with YAML files:

```bash
# Apply a VirtualService
kubectl apply -f virtualservice.yaml

# Apply a DestinationRule
kubectl apply -f destinationrule.yaml

# Apply both
kubectl apply -f canary-deployment.yaml
```

See [traffic-management.md](traffic-management.md) for complete YAML examples including:
- Canary deployments with traffic splitting
- A/B testing configurations
- Circuit breakers and outlier detection
- Retry policies and timeouts
- Header-based routing
- Fault injection for testing

## Customizing the Tool

Modify the `main()` function in [traffic-handler.go](traffic-handler.go) to add custom logic:

```go
func main() {
    namespace := os.Getenv("NAMESPACE")
    if namespace == "" {
        namespace = "default"
    }

    handler, err := NewIstioTrafficHandler(namespace)
    if err != nil {
        log.Fatalf("Failed to create traffic handler: %v", err)
    }

    // Your custom logic here
    handler.ShowStatus()
    
    // Get specific VirtualService details
    handler.GetVirtualService("my-service")
    
    // Delete a resource
    // handler.DeleteVirtualService("old-service")
}
```

## Using the Makefile

The Makefile provides convenient commands:

```bash
make help          # Show all available commands
make deps          # Download dependencies
make build         # Build the executable
make run           # Run the traffic handler
make test          # Run tests
make lint          # Run linter
make clean         # Clean build artifacts
make verify-istio  # Check Istio installation
make list-vs       # List all VirtualServices
make list-dr       # List all DestinationRules
```

## Troubleshooting

### "failed to create Istio client"

**Cause**: Kubernetes configuration issue

**Solutions**:
- Ensure kubeconfig is properly configured: `kubectl config view`
- Verify cluster is accessible: `kubectl cluster-info`
- Check KUBECONFIG environment variable: `echo $KUBECONFIG`
- Try setting it explicitly: `export KUBECONFIG=~/.kube/config`

### "failed to list virtual services"

**Cause**: Permission or Istio installation issue

**Solutions**:
- Verify Istio is installed: `kubectl get pods -n istio-system`
- Check you have permissions: `kubectl auth can-i list virtualservices`
- Verify the API resource exists: `kubectl api-resources | grep istio`

### "No VirtualServices found"

**Cause**: Resources are in a different namespace or don't exist

**Solutions**:
- Set the correct namespace: `export NAMESPACE=your-namespace`
- Check all namespaces: `kubectl get virtualservices --all-namespaces`
- Verify resources exist: `kubectl get vs,dr,gateway -A`

### Build Errors

If you encounter build errors:

```bash
# Clean and rebuild
rm go.mod go.sum
go mod init istio-traffic-handler
go get istio.io/client-go@v1.20.0
go get k8s.io/client-go@v0.29.0
go mod tidy
go build -o traffic-handler traffic-handler.go
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NAMESPACE` | Kubernetes namespace to target | `default` |
| `KUBECONFIG` | Path to kubeconfig file | `~/.kube/config` |

## Next Steps

1. **Explore Examples**: Check [traffic-management.md](traffic-management.md) for YAML examples
2. **Learn Commands**: Review [CHEATSHEET.md](CHEATSHEET.md) for quick command reference
3. **Setup Istio**: Follow [README.md](README.md) for Istio installation and configuration
4. **Test Features**: Try the sample configurations in the repo

## Related Resources

- [Istio Official Documentation](https://istio.io/latest/docs/)
- [Istio Traffic Management](https://istio.io/latest/docs/concepts/traffic-management/)
- [Istio Security](https://istio.io/latest/docs/concepts/security/)
- [Istio Observability](https://istio.io/latest/docs/concepts/observability/)
