# Istio Observability and Monitoring

This file contains examples for implementing observability in Istio.

## Key Observability Features

Istio provides three main pillars of observability:
1. **Metrics**: Request rates, latencies, error rates
2. **Logs**: Request and connection logs
3. **Traces**: Distributed tracing across services

## Metrics

Istio automatically collects metrics about all traffic in the mesh.

### Access Metrics via Prometheus

```bash
# Port-forward to Prometheus
kubectl port-forward svc/prometheus -n istio-system 9090:9090
# Access: http://localhost:9090
```

### Useful Prometheus Queries

```promql
# Request rate
rate(istio_requests_total[5m])

# Request latency (p95)
histogram_quantile(0.95, rate(istio_request_duration_milliseconds_bucket[5m]))

# Error rate
rate(istio_requests_total{response_code=~"5.."}[5m])

# Request by source/destination
sum(rate(istio_requests_total[5m])) by (source_workload, destination_workload)
```

## Distributed Tracing

Enable distributed tracing to track requests across services.

### View Traces via Jaeger

```bash
# Port-forward to Jaeger
kubectl port-forward svc/jaeger -n istio-system 16686:16686
# Access: http://localhost:16686
```

### Configure Tracing Sampling

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio
  namespace: istio-system
data:
  mesh: |
    # Global configuration
    enableTracing: true
    # Sampling rate (0-100)
    traceSamplingRate: 100
```

### Example: Custom Headers for Tracing

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: reviews
spec:
  hosts:
  - reviews
  http:
  - match:
    - headers:
        x-custom-trace:
          exact: "true"
    route:
    - destination:
        host: reviews
  - route:
    - destination:
        host: reviews
```

## Service Graph Visualization

Use Kiali to visualize the service mesh and traffic flows.

### Access Kiali Dashboard

```bash
# Port-forward to Kiali
kubectl port-forward svc/kiali -n istio-system 20000:20000
# Access: http://localhost:20000
# Default credentials: admin/admin
```

### Kiali Features

- **Service Graph**: Visualize service dependencies
- **Traffic Metrics**: See request rates, error rates, latencies
- **Configuration**: Validate Istio configuration
- **Logs**: View pod logs
- **Traces**: Access distributed traces

## Access Logging

Configure access logs to capture detailed request/response information.

### Enable Access Logging

```bash
# Edit ConfigMap
kubectl edit configmap istio -n istio-system
```

In the `mesh` section, add:

```yaml
accessLogFile: /dev/stdout
accessLogFormat: |
  [%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%"
  %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT%
  "%DURATION%" "%RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)%"
  "%REQ(X-FORWARDED-FOR)%" "%REQ(USER-AGENT)%"
  "%REQ(X-REQUEST-ID)%" "%REQ(:AUTHORITY)%" "%UPSTREAM_HOST%"
```

### View Access Logs

```bash
# View logs of sidecar proxy
kubectl logs <pod-name> -n default -c istio-proxy

# Stream logs
kubectl logs -f <pod-name> -n default -c istio-proxy
```

### Example Access Log

```
[2024-01-15T10:30:45.123Z] "GET /api/v1/users HTTP/1.1" 200 - 0 1234 "45" "23" "-" "Mozilla/5.0" "789abc" "api.example.com" "10.0.0.5:8080"
```

## Telemetry Configuration

Advanced telemetry configuration for specific workloads.

```yaml
apiVersion: telemetry.istio.io/v1alpha1
kind: Telemetry
metadata:
  name: custom-metrics
  namespace: default
spec:
  metrics:
  - providers:
    - name: prometheus
    dimensions:
    - request_path
    - source_namespaced_name
    - destination_namespaced_name
  logs:
  - providers:
    - name: stackdriver
    overrides:
    - match:
        mode: SERVER
      disabled: false
```

## Grafana Integration

Grafana dashboards display metrics collected by Prometheus.

### Access Grafana

```bash
# Port-forward to Grafana
kubectl port-forward svc/grafana -n istio-system 3000:3000
# Access: http://localhost:3000
# Default credentials: admin/admin
```

### Built-in Dashboards

- Istio Mesh Dashboard: Overall mesh health and metrics
- Istio Service Dashboard: Per-service metrics
- Istio Workload Dashboard: Per-workload metrics

## Monitoring Best Practices

### 1. Set Up Alerts

Create Prometheus alerts for critical thresholds:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: istio-alerts
  namespace: istio-system
spec:
  groups:
  - name: istio
    interval: 30s
    rules:
    - alert: HighErrorRate
      expr: rate(istio_requests_total{response_code=~"5.."}[5m]) > 0.05
      for: 5m
      annotations:
        summary: "High error rate detected"
    
    - alert: HighLatency
      expr: histogram_quantile(0.95, rate(istio_request_duration_milliseconds_bucket[5m])) > 1000
      for: 5m
      annotations:
        summary: "High latency detected"
```

### 2. Custom Metrics

Export custom application metrics through Prometheus endpoints for correlation with Istio metrics.

### 3. Correlation

Correlate metrics, logs, and traces to investigate issues:

1. Notice high error rate in Grafana
2. Find affected service in Kiali
3. View logs in kubectl
4. Trace request in Jaeger

### 4. Health Checks

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-pod
spec:
  containers:
  - name: app
    livenessProbe:
      httpGet:
        path: /healthz
        port: 8080
    readinessProbe:
      httpGet:
        path: /ready
        port: 8080
```

## Common Observability Patterns

### Pattern 1: Debug a Slow Service

1. Check latency in Grafana
2. Identify service in Kiali
3. View upstream dependencies
4. Check error rates
5. Review distributed trace
6. Check proxy logs

### Pattern 2: Investigate Error Spike

1. Alert triggers high error rate
2. Use Kiali to find affected service
3. View request traces in Jaeger
4. Check application logs
5. Verify authorization policies
6. Rollback or fix deployment

## Cleanup

```bash
# Delete monitoring tools (if using demo profile)
kubectl delete ns prometheus grafana jaeger kiali
```
