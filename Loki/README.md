# Loki Setup Guide

Loki is a horizontally scalable, highly available, multi-tenant log aggregation system inspired by Prometheus. It's designed to be very cost effective and easy to operate as it does not index the contents of the logs, but rather a set of labels for each log stream.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [Docker Installation](#docker-installation)
  - [Binary Installation](#binary-installation)
  - [Kubernetes Installation](#kubernetes-installation)
- [Basic Configuration](#basic-configuration)
- [Docker Compose Setup](#docker-compose-setup)
- [Grafana Integration](#grafana-integration)
- [Querying Logs](#querying-logs)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- Docker & Docker Compose (for Docker installation)
- Kubernetes cluster (for K8s installation)
- Grafana (optional, for visualization)
- Basic understanding of logging and observability

## Installation

### Docker Installation

**Using Docker Compose (Recommended)**

1. Create a `docker-compose.yml` file:

```yaml
version: '3'

services:
  loki:
    image: grafana/loki:latest
    container_name: loki
    restart: unless-stopped
    ports:
      - "3100:3100"
    volumes:
      - ./loki-config.yml:/etc/loki/local-config.yml:ro
      - loki-data:/loki
    environment:
      - LOKI_CONFIG_FILE=/etc/loki/local-config.yml
    networks:
      - loki-network

volumes:
  loki-data:

networks:
  loki-network:
    driver: bridge
```

2. Run the setup script:

```bash
go run setup.go
```

### Binary Installation

**Linux/macOS**

```bash
# Download the latest release
wget https://github.com/grafana/loki/releases/download/v2.9.3/loki-linux-amd64.zip
unzip loki-linux-amd64.zip
chmod +x loki-linux-amd64

# Run with config
./loki-linux-amd64 -config.file=loki-config.yml
```

**Windows**

```bash
# Download from releases page
# https://github.com/grafana/loki/releases

# Extract and run
.\loki-windows-amd64.exe -config.file=loki-config.yml
```

### Kubernetes Installation

**Using Helm**

```bash
# Add Grafana Helm repository
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Install Loki
helm install loki grafana/loki-stack --namespace logging --create-namespace

# Verify installation
kubectl get pods -n logging
```

## Basic Configuration

### Loki Configuration File (loki-config.yml)

```yaml
auth_enabled: false

ingester:
  chunk_idle_period: 3m
  chunk_retain_period: 1m
  max_chunk_age: 1h
  chunk_encoding: snappy

limits_config:
  enforce_metric_name: false
  reject_old_samples: true
  reject_old_samples_max_age: 168h

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

server:
  http_listen_port: 3100
  log_level: info

storage_config:
  boltdb_shipper:
    active_index_directory: /loki/index
    shared_store: filesystem
  filesystem:
    directory: /loki/chunks

chunk_store_config:
  max_look_back_period: 0s

table_manager:
  retention_deletes_enabled: false
  retention_period: 0s
```

## Docker Compose Setup

1. **Create the project structure:**

```
loki/
├── docker-compose.yml
├── loki-config.yml
└── setup.go
```

2. **Run the automated setup:**

```bash
go run setup.go
```

This will:
- Create the Docker Compose configuration
- Set up the Loki configuration
- Create necessary volumes
- Start the Loki service

3. **Verify Loki is running:**

```bash
curl http://localhost:3100/ready
```

You should see `ready`.

## Grafana Integration

### Add Loki datasource in Grafana

1. Open Grafana (usually http://localhost:3000)
2. Navigate to **Configuration** > **Data Sources**
3. Click **Add data source**
4. Select **Loki**
5. Set HTTP URL to: `http://loki:3100`
6. Click **Save & Test**

### Docker Compose with Grafana

Add to your `docker-compose.yml`:

```yaml
  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana
    depends_on:
      - loki
    networks:
      - loki-network

volumes:
  grafana-data:
```

## Security: Nginx Reverse Proxy & Basic Auth

Loki is secured by an Nginx reverse proxy with basic authentication. By default, Nginx listens on port 8080 and proxies requests to Loki. The default credentials are:

- Username: `admin`
- Password: `admin`

To change the password, update the `nginx.htpasswd` file (use an htpasswd generator).

**Access Loki securely:**

- Use: `http://localhost:8080` (you will be prompted for credentials)
- All direct access to Loki (port 3100) can be firewalled or restricted to internal containers.

### Docker Compose with Nginx Security

The setup includes an Nginx service:

```yaml
  nginx:
    image: nginx:latest
    container_name: loki-nginx
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx.htpasswd:/etc/nginx/.htpasswd:ro
    depends_on:
      - loki
    networks:
      - loki-network
```

- All requests to Loki should go through Nginx for authentication.
- Update `nginx.conf` and `nginx.htpasswd` for custom security needs.

## Querying Logs

### LogQL - Loki Query Language

**Basic Label Query:**
```
{job="prometheus"}
```

**Filter by logs containing text:**
```
{job="prometheus"} |= "error"
```

**Regex filter:**
```
{job="prometheus"} |~ "error.*timeout"
```

**Exclude logs:**
```
{job="prometheus"} != "debug"
```

**Extract fields and aggregate:**
```
sum(rate({job="prometheus"} |= "requests" [5m])) by (status_code)
```

### Examples

**View all logs from a container:**
```
{container="myapp"}
```

**Get error logs:**
```
{job="myapp"} |= "ERROR"
```

**Count logs by level:**
```
sum(count_over_time({job="myapp"} [5m])) by (level)
```

**Recent logs (last 100 lines):**
```
{job="prometheus"} | tail 100
```

## Common Use Cases

### 1. Docker Container Logging

Use Docker's Loki logging driver:

```yaml
services:
  myapp:
    image: myapp:latest
    logging:
      driver: loki
      options:
        loki-url: "http://localhost:3100/loki/api/v1/push"
        loki-batch-size: "400"
```

### 2. Kubernetes Pod Logs

Using Promtail for log collection:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: promtail-config
data:
  promtail-config.yml: |
    clients:
      - url: http://loki:3100/loki/api/v1/push
    
    scrape_configs:
      - job_name: kubernetes-pods
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_name]
            target_label: pod
```

### 3. Multi-tenant Setup

Enable authentication in Loki config:

```yaml
auth_enabled: true

server:
  http_listen_port: 3100
```

## Troubleshooting

### Loki not responding

1. Check if Loki is running:
```bash
docker ps | grep loki
```

2. Check logs:
```bash
docker logs loki
```

3. Test connectivity:
```bash
curl -v http://localhost:3100/loki/api/v1/label
```

### High Memory Usage

- Reduce `chunk_idle_period` in config
- Enable compression (`chunk_encoding: snappy`)
- Lower `max_chunk_age`

### Logs not appearing

1. Verify logging driver is configured correctly
2. Check Promtail scrape configs
3. Verify label names in queries

### Performance Issues

1. Enable caching:
```yaml
cache_config:
  enable_fifocache: true
```

2. Use external object storage instead of filesystem
3. Enable compression
4. Tune retention policies

### Port already in use

Change the port in config and docker-compose:

```yaml
server:
  http_listen_port: 3101  # Change port
```

## Additional Resources

- [Official Loki Documentation](https://grafana.com/docs/loki/latest/)
- [LogQL Query Language](https://grafana.com/docs/loki/latest/logql/)
- [Promtail Documentation](https://grafana.com/docs/loki/latest/clients/promtail/)
- [Grafana Loki GitHub Repository](https://github.com/grafana/loki)
