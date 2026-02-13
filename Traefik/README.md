# Traefik Setup Guide

Traefik is a modern HTTP reverse proxy and load balancer that makes deploying microservices easy. It automatically discovers the right configuration for your services through various providers (Docker, Kubernetes, etc.).

## Table of Contents
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [Docker Installation](#docker-installation)
  - [Binary Installation](#binary-installation)
  - [Kubernetes Installation](#kubernetes-installation)
- [Basic Configuration](#basic-configuration)
- [Configuration Examples](#configuration-examples)
- [SSL/TLS with Let's Encrypt](#ssltls-with-lets-encrypt)
- [Monitoring & Dashboard](#monitoring--dashboard)
- [Common Use Cases](#common-use-cases)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- Docker (for Docker installation)
- Kubernetes cluster (for K8s installation)
- Domain name (for SSL/TLS setup)
- Basic understanding of reverse proxies

## Installation

### Docker Installation

**Using Docker Compose (Recommended)**

1. Create a `docker-compose.yml` file:

```yaml
version: '3'

services:
  traefik:
    image: traefik:v3.0
    container_name: traefik
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"  # Dashboard
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik.yml:/traefik.yml:ro
      - ./acme.json:/acme.json
    networks:
      - traefik-network

networks:
  traefik-network:
    external: true
```

2. Create the network:
```bash
docker network create traefik-network
```

3. Create `traefik.yml` configuration file (see [Basic Configuration](#basic-configuration))

4. Create `acme.json` for SSL certificates:
```bash
touch acme.json
chmod 600 acme.json
```

5. Start Traefik:
```bash
docker-compose up -d
```

**Using Docker CLI**

```bash
docker run -d \
  -p 80:80 \
  -p 443:443 \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v $PWD/traefik.yml:/traefik.yml:ro \
  --name traefik \
  traefik:v3.0
```

### Binary Installation

**Linux/macOS**

```bash
# Download the binary
wget https://github.com/traefik/traefik/releases/download/v3.0.0/traefik_v3.0.0_linux_amd64.tar.gz

# Extract
tar -xzf traefik_v3.0.0_linux_amd64.tar.gz

# Move to /usr/local/bin
sudo mv traefik /usr/local/bin/

# Verify installation
traefik version
```

**Windows**

Download from: https://github.com/traefik/traefik/releases

Add to PATH and run:
```powershell
traefik.exe version
```

### Kubernetes Installation

**Using Helm**

```bash
# Add Traefik Helm repository
helm repo add traefik https://traefik.github.io/charts
helm repo update

# Install Traefik
helm install traefik traefik/traefik \
  --namespace traefik \
  --create-namespace
```

**Using kubectl**

```bash
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/master/docs/content/reference/dynamic-configuration/kubernetes-crd-definition-v1.yml
kubectl apply -f https://raw.githubusercontent.com/traefik/traefik/master/docs/content/reference/dynamic-configuration/kubernetes-crd-rbac.yml
```

## Basic Configuration

Create a `traefik.yml` file:

```yaml
# Static Configuration
api:
  dashboard: true
  insecure: true  # Set to false in production

entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
    network: traefik-network

log:
  level: INFO

accessLog: {}
```

## Configuration Examples

### Example 1: Simple Web Application

**docker-compose.yml for your app:**

```yaml
version: '3'

services:
  webapp:
    image: nginx:alpine
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.webapp.rule=Host(`example.com`)"
      - "traefik.http.routers.webapp.entrypoints=web"
      - "traefik.http.services.webapp.loadbalancer.server.port=80"
    networks:
      - traefik-network

networks:
  traefik-network:
    external: true
```

### Example 2: Multiple Services with Path-Based Routing

```yaml
version: '3'

services:
  api:
    image: myapi:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.api.rule=Host(`example.com`) && PathPrefix(`/api`)"
      - "traefik.http.routers.api.entrypoints=web"
      - "traefik.http.services.api.loadbalancer.server.port=8000"
    networks:
      - traefik-network

  frontend:
    image: myfrontend:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.frontend.rule=Host(`example.com`)"
      - "traefik.http.routers.frontend.entrypoints=web"
      - "traefik.http.services.frontend.loadbalancer.server.port=3000"
    networks:
      - traefik-network

networks:
  traefik-network:
    external: true
```

### Example 3: With Middleware (Basic Auth)

```yaml
version: '3'

services:
  admin:
    image: admin-panel:latest
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.admin.rule=Host(`admin.example.com`)"
      - "traefik.http.routers.admin.entrypoints=web"
      - "traefik.http.routers.admin.middlewares=auth"
      - "traefik.http.middlewares.auth.basicauth.users=admin:$$apr1$$H6uskkkW$$IgXLP6ewTrSuBkTrqE8wj/"
      - "traefik.http.services.admin.loadbalancer.server.port=8080"
    networks:
      - traefik-network

networks:
  traefik-network:
    external: true
```

Generate password hash:
```bash
echo $(htpasswd -nb admin password) | sed -e s/\\$/\\$\\$/g
```

## SSL/TLS with Let's Encrypt

Update your `traefik.yml`:

```yaml
api:
  dashboard: true

entryPoints:
  web:
    address: ":80"
    http:
      redirections:
        entryPoint:
          to: websecure
          scheme: https
  websecure:
    address: ":443"
    http:
      tls:
        certResolver: letsencrypt

certificatesResolvers:
  letsencrypt:
    acme:
      email: your-email@example.com
      storage: /acme.json
      httpChallenge:
        entryPoint: web

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
    network: traefik-network

log:
  level: INFO
```

Update service labels:

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.myapp.rule=Host(`example.com`)"
  - "traefik.http.routers.myapp.entrypoints=websecure"
  - "traefik.http.routers.myapp.tls.certresolver=letsencrypt"
  - "traefik.http.services.myapp.loadbalancer.server.port=80"
```

## Monitoring & Dashboard

Access the Traefik dashboard at: `http://localhost:8080`

**Secure the dashboard:**

```yaml
api:
  dashboard: true
  # Remove insecure: true

# Add middleware for authentication
http:
  middlewares:
    dashboard-auth:
      basicAuth:
        users:
          - "admin:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"

# Apply to dashboard router
  routers:
    dashboard:
      rule: "Host(`traefik.example.com`)"
      service: api@internal
      middlewares:
        - dashboard-auth
```

**Prometheus Metrics:**

Add to `traefik.yml`:

```yaml
metrics:
  prometheus:
    addEntryPointsLabels: true
    addServicesLabels: true
```

Metrics available at: `http://localhost:8080/metrics`

## Common Use Cases

### Load Balancing Multiple Instances

```yaml
services:
  web:
    image: nginx:alpine
    deploy:
      replicas: 3
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.web.rule=Host(`example.com`)"
      - "traefik.http.services.web.loadbalancer.server.port=80"
    networks:
      - traefik-network
```

### Rate Limiting

```yaml
labels:
  - "traefik.http.middlewares.ratelimit.ratelimit.average=100"
  - "traefik.http.middlewares.ratelimit.ratelimit.burst=50"
  - "traefik.http.routers.myapp.middlewares=ratelimit"
```

### CORS Headers

```yaml
labels:
  - "traefik.http.middlewares.cors.headers.accesscontrolallowmethods=GET,OPTIONS,PUT,POST,DELETE"
  - "traefik.http.middlewares.cors.headers.accesscontrolalloworigin=*"
  - "traefik.http.middlewares.cors.headers.accesscontrolmaxage=100"
  - "traefik.http.middlewares.cors.headers.addvaryheader=true"
  - "traefik.http.routers.myapp.middlewares=cors"
```

### Circuit Breaker

```yaml
labels:
  - "traefik.http.middlewares.circuitbreaker.circuitbreaker.expression=NetworkErrorRatio() > 0.5"
  - "traefik.http.routers.myapp.middlewares=circuitbreaker"
```

## Troubleshooting

### Check Traefik Logs

**Docker:**
```bash
docker logs traefik
```

**Docker Compose:**
```bash
docker-compose logs -f traefik
```

### Common Issues

**1. Services not appearing in dashboard**
- Check if `traefik.enable=true` label is set
- Verify the service is on the same network as Traefik
- Check Docker socket is mounted correctly

**2. SSL certificate issues**
- Ensure `acme.json` has correct permissions (600)
- Verify email is set in Let's Encrypt configuration
- Check if port 80 is accessible from the internet

**3. 404 Gateway Not Found**
- Verify router rules (Host, PathPrefix)
- Check entrypoints configuration
- Ensure service port matches container port

### Debug Mode

Enable debug logging:

```yaml
log:
  level: DEBUG
```

### Verify Configuration

```bash
traefik --configFile=traefik.yml --validate
```

## Useful Commands

```bash
# View Traefik version
docker exec traefik traefik version

# Reload configuration
docker restart traefik

# Check network
docker network inspect traefik-network

# Test routing
curl -H "Host: example.com" http://localhost

# View certificates
docker exec traefik cat /acme.json
```

## Resources

- [Official Documentation](https://doc.traefik.io/traefik/)
- [GitHub Repository](https://github.com/traefik/traefik)
- [Community Forum](https://community.traefik.io/)
- [Docker Hub](https://hub.docker.com/_/traefik)

## License

Traefik is open-source software licensed under the MIT License.
