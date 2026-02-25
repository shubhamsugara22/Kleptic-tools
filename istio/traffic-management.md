# Virtual Services and Destination Rules

This file contains example configurations for traffic management in Istio.

## VirtualService Example

A VirtualService allows you to configure how traffic is routed to services.

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: bookinfo
  namespace: default
spec:
  hosts:
  - bookinfo
  http:
  - match:
    - uri:
        prefix: "/api"
      headers:
        user-agent:
          regex: ".*Chrome.*"
    route:
    - destination:
        host: bookinfo
        subset: v1
      weight: 80
    - destination:
        host: bookinfo
        subset: v2
      weight: 20
    timeout: 10s
    retries:
      attempts: 3
      perTryTimeout: 2s
```

## DestinationRule Example

A DestinationRule defines policies that apply to traffic after it has been routed to a service.

```yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: bookinfo
  namespace: default
spec:
  host: bookinfo
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        http1MaxPendingRequests: 50
        maxRequestsPerConnection: 2
    loadBalancer:
      simple: ROUND_ROBIN
    outlierDetection:
      consecutive5xxErrors: 5
      interval: 30s
      baseEjectionTime: 30s
  subsets:
  - name: v1
    labels:
      version: v1
  - name: v2
    labels:
      version: v2
    trafficPolicy:
      loadBalancer:
        simple: LEAST_CONN
```

## Gateway Example

A Gateway manages inbound and outbound traffic for the mesh.

```yaml
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: bookinfo-gateway
  namespace: default
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "bookinfo.example.com"
  - port:
      number: 443
      name: https
      protocol: HTTPS
    tls:
      mode: SIMPLE
      credentialName: bookinfo-cert
    hosts:
    - "bookinfo.example.com"
---
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: bookinfo-route
  namespace: default
spec:
  hosts:
  - "bookinfo.example.com"
  gateways:
  - bookinfo-gateway
  http:
  - match:
    - uri:
        prefix: "/productpage"
    - uri:
        prefix: "/static"
    - uri:
        prefix: "/login"
    - uri:
        prefix: "/logout"
    - uri:
        prefix: "/api/v1/products"
    route:
    - destination:
        host: productpage
        port:
          number: 9080
```

## Canary Deployment Example

Gradually shift traffic from v1 to v2 for canary deployments.

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: reviews-canary
spec:
  hosts:
  - reviews
  http:
  - match:
    - headers:
        user:
          exact: "tester"
    route:
    - destination:
        host: reviews
        subset: v2
  - route:
    - destination:
        host: reviews
        subset: v1
      weight: 90
    - destination:
        host: reviews
        subset: v2
      weight: 10
---
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: reviews
spec:
  host: reviews
  subsets:
  - name: v1
    labels:
      version: v1
  - name: v2
    labels:
      version: v2
```

## Circuit Breaker Example

Implement circuit breaker pattern with outlier detection.

```yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: productpage-circuit-breaker
spec:
  host: productpage
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 10
      http:
        http1MaxPendingRequests: 10
        maxRequestsPerConnection: 1
    outlierDetection:
      consecutive5xxErrors: 5
      interval: 30s
      baseEjectionTime: 30s
      maxEjectionPercent: 50
      minRequestVolume: 10
```

## Apply Configuration

```bash
# Apply a single configuration
kubectl apply -f virtualservice-example.yaml

# Apply all configurations in a directory
kubectl apply -f ./

# Check resources
kubectl get virtualservices
kubectl get destinationrules
kubectl get gateways

# Describe a resource
kubectl describe vs bookinfo
```
