# Istio Traffic Handler - Go Usage Guide

A comprehensive Go application for programmatic Istio traffic management.

## Quick Start

```bash
# Download dependencies
go mod download

# Run the traffic handler
go run traffic-handler.go

# With custom namespace
NAMESPACE=production go run traffic-handler.go

# Build and run
go build -o traffic-handler traffic-handler.go
./traffic-handler
```

## Features

✓ Canary Deployments with traffic splitting  
✓ Gradual rollout automation  
✓ Circuit breaker configuration  
✓ Retry policy management  
✓ Header-based routing  
✓ Dynamic traffic weight updates  

## Basic Examples

### 1. Create Canary Deployment

```go
handler, _ := NewIstioTrafficHandler("default")
handler.CreateDestinationRule("myapp")
handler.CreateCanaryDeployment("myapp", 90, 10) // 90% stable, 10% canary
```

### 2. Update Traffic Weights

```go
handler.UpdateTrafficWeights("myapp", 70, 30) // Shift to 30% canary
```

### 3. Gradual Rollout

```go
stages := []int32{10, 25, 50, 75, 100}
interval := 30 * time.Second
handler.GradualCanaryRollout("myapp", stages, interval)
```

### 4. Apply Circuit Breaker

```go
handler.ApplyCircuitBreaker("myapp", 100, 50)
```

### 5. Create Retry Policy

```go
handler.CreateRetryPolicy("myapp-api", 3, "10s")
```

## Using Makefile

```bash
make help          # Show all commands
make deps          # Install dependencies
make build         # Build binary
make run           # Run the handler
make verify-istio  # Check Istio installation
make list-vs       # List VirtualServices
```

## Environment Variables

- `NAMESPACE`: Target namespace (default: "default")
- `KUBECONFIG`: Path to kubeconfig (default: "~/.kube/config")

## Resources

See these files for more details:
- **README.md**: Istio overview and setup
- **traffic-management.md**: YAML configuration examples
- **CHEATSHEET.md**: Quick command reference
- **traffic-handler.go**: Full source code
