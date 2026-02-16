# Envoy Proxy

Envoy is a Layer 7 proxy and service mesh communication bus designed for large, modern service-oriented architectures. It was originally built at Lyft and is now a CNCF project.

## Overview

Envoy is an open-source edge and service proxy, designed for large modern service-oriented architectures. It can be deployed in front of services or as a proxy between services for intra-service communication.

### Key Features

- **Layer 7 Proxy**: Advanced HTTP/2 and HTTP/3 support with sophisticated routing
- **Service Mesh Integration**: Core component of Istio and other service mesh platforms
- **Load Balancing**: Automatic load balancing with circuit breaking and outlier detection
- **Observability**: Built-in metrics, distributed tracing, and detailed logging
- **Traffic Management**: Retries, timeouts, rate limiting, and traffic shaping
- **Security**: mTLS support, authorization policies, and authentication
- **Dynamic Configuration**: Configuration can be updated without restarts via xDS APIs

## Installation

### Using the Setup Script

Run the setup script to install and configure Envoy:

```bash
./setup.sh
```

Or use Go:

```bash
go run setup.go
```

## Quick Start

### Basic Configuration

Envoy is configured via a configuration file (typically YAML or JSON). Here's a simple example:

```yaml
static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 10000
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          http_filters:
          - name: envoy.filters.http.router
          route_config:
            name: local_route
            virtual_hosts:
            - name: backend
              domains: ["*"]
              routes:
              - match:
                  prefix: "/"
                route:
                  cluster: backend_service
  clusters:
  - name: backend_service
    type: STATIC
    load_assignment:
      cluster_name: backend_service
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 8080
admin:
  access_log_path: /tmp/admin_access.log
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 9901
```

### Running Envoy

```bash
envoy -c envoy.yaml
```

## Use Cases

- **API Gateway**: Route incoming requests to backend services
- **Service Proxy**: Transparent proxy for inter-service communication
- **Load Balancer**: Distribute traffic across multiple instances
- **Ingress Controller**: Kubernetes ingress alternative to NGINX/Traefik
- **Service Mesh Data Plane**: Part of Istio or other service mesh solutions

## Configuration

Envoy supports multiple configuration methods:

1. **Static File**: Configuration file (YAML/JSON)
2. **Dynamic (xDS)**: Remote configuration via APIs
3. **Hot Restart**: Restart without dropping connections

## Monitoring & Management

- **Admin Interface**: Web UI and API at `localhost:9901`
- **Metrics**: Prometheus-compatible metrics exposure
- **Logging**: Detailed access logs and debug logs
- **Tracing**: Integration with Jaeger, Zipkin, and other tracers

## Documentation

- [Official Envoy Documentation](https://www.envoyproxy.io/)
- [Configuration Reference](https://www.envoyproxy.io/docs/envoy/latest/configuration/configuration)
- [xDS Protocol](https://www.envoyproxy.io/docs/envoy/latest/api/api)

## Docker

Pull the official Envoy Docker image:

```bash
docker pull envoyproxy/envoy:v1.27-latest
```

Run a container:

```bash
docker run -v $(pwd)/envoy.yaml:/etc/envoy/envoy.yaml -p 10000:10000 -p 9901:9901 envoyproxy/envoy:v1.27-latest
```

## Comparison with Kong and Traefik

| Feature | Envoy | Kong | Traefik |
|---------|-------|------|---------|
| Primary Use | Service Mesh Proxy | API Gateway | Ingress Controller |
| Configuration | YAML/JSON | Admin API | File/Labels |
| Kubernetes | Yes (via Istio) | Yes | Native |
| Learning Curve | Steep | Moderate | Easy |
| Performance | Very High | High | Good |

## References

- [Envoy Official Site](https://www.envoyproxy.io/)
- [Istio (uses Envoy)](https://istio.io/)
- [Service Mesh Comparison](https://www.cncf.io/blog/2022/02/10/cncf-service-mesh-landscape/)
