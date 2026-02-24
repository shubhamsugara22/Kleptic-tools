# Trivy - Container and Infrastructure Security Scanner

Trivy is a comprehensive and versatile security scanner developed by Aqua Security. It scans for vulnerabilities, misconfigurations, secrets, and license issues across multiple artifact types in your application lifecycle.

## Overview

Trivy is designed to be:
- **Fast** - Scanning completes in seconds
- **Lightweight** - Minimal dependencies and resource requirements
- **Accurate** - Low false-positive rates with multiple data sources
- **Comprehensive** - Supports multiple artifact types and formats

## Features

### Vulnerability Scanning
- **Container Images**: Scan Docker/OCI images for known vulnerabilities
- **Filesystems**: Scan directories, repositories, and local file systems
- **Git Repositories**: Detect vulnerabilities in source code dependencies
- **Kubernetes Clusters**: Scan running pods and manifests
- **SBOM (Software Bill of Materials)**: Generate comprehensive dependency lists

### Additional Security Checks
- **Misconfiguration Detection**: Identify security issues in IaC files (Terraform, CloudFormation, Kubernetes manifests)
- **Secret Detection**: Find hardcoded secrets, API keys, and credentials
- **License Compliance**: Monitor open-source license usage

### Supported Artifact Types
- Container Images (Docker, OCI, Podman)
- Kubernetes YAML manifests
- Terraform files
- CloudFormation templates
- Docker Compose files
- Dockerfile
- Git repositories
- Virtual Machine/Filesystem images
- Source code repositories

## Installation

### Using apt (Debian/Ubuntu)
```bash
sudo apt-get install wget apt-transport-https gnupg lsb-release
wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo apt-key add -
echo "deb https://aquasecurity.github.io/trivy-repo/deb $(lsb_release -sc) main" | sudo tee -a /etc/apt/sources.list.d/trivy.list
sudo apt-get update
sudo apt-get install trivy
```

### Using Homebrew (macOS)
```bash
brew install trivy
```

### Using Docker
```bash
docker run aquasec/trivy --version
```

### Using curl
```bash
curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
```

## Quick Start

### Scan a Container Image
```bash
trivy image nginx:latest
trivy image --severity HIGH,CRITICAL python:3.9
```

### Scan a Filesystem
```bash
trivy fs /path/to/repository
trivy fs --severity CRITICAL .
```

### Scan a Git Repository
```bash
trivy repo https://github.com/aquasecurity/trivy.git
```

### Scan Kubernetes Cluster
```bash
trivy k8s cluster
trivy k8s all-namespaces
```

### Generate SBOM
```bash
trivy image --format cyclonedx nginx:latest > sbom.xml
trivy image --format spdx nginx:latest > sbom.spdx.json
```

### Detect Secrets
```bash
trivy fs --scanners secret /path/to/code
```

## Configuration

Create a `trivy.yaml` configuration file:

```yaml
severity:
  - CRITICAL
  - HIGH

skip-update: false
offline-db: false

output: ""
format: "table"

# Scanners to run
scanners:
  - vuln
  - config
  - secret

# Vulnerability sources
vuln-source: "default"

# Cache settings
cache:
  dir: /home/user/.cache/trivy

# Security checks
checks:
  - AVD-AWS-0123
```

Run with config:
```bash
trivy image --config trivy.yaml my-image:latest
```

## Integration Examples

### Docker Scan
```bash
docker scan myimage:latest
```

### CI/CD Pipeline (GitHub Actions)
```yaml
- name: Run Trivy scan
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: 'myimage:latest'
    format: 'sarif'
    output: 'trivy-results.sarif'

- name: Upload Trivy results to GitHub Security tab
  uses: github/codeql-action/upload-sarif@v2
  with:
    sarif_file: 'trivy-results.sarif'
```

### Kubernetes Admission Controller
Use Trivy as an admission webhook to enforce security policies on container image deployments:

```bash
kubectl create configmap trivy-config --from-file=trivy.yaml
kubectl create deployment trivy-server --image=aquasec/trivy:latest -- server --listen 0.0.0.0:8080
```

## Output Formats

Trivy supports multiple output formats:

- **table** (default) - Human-readable table format
- **json** - Structured JSON output for automation
- **sarif** - SARIF format for integration with GitHub Security
- **cyclonedx** - CycloneDX SBOM format
- **spdx** - SPDX SBOM format
- **csv** - Comma-separated values

Example:
```bash
trivy image --format json --output results.json nginx:latest
trivy image --format sarif --output results.sarif nginx:latest
```

## Common Use Cases

### 1. Pre-push Hook
Scan images before pushing to registry:
```bash
#!/bin/bash
IMAGE=$1
trivy image --severity HIGH,CRITICAL "$IMAGE" || exit 1
docker push "$IMAGE"
```

### 2. Registry Scanning
Scan all images in a registry periodically

### 3. Compliance Reports
Generate compliance reports with known vulnerabilities:
```bash
trivy image --format json nginx:latest | jq '.Results[] | {Target, Vulnerabilities}'
```

### 4. Baseline Comparison
Compare vulnerability status over time

## Troubleshooting

### Database Updates
```bash
# Force update of vulnerability database
trivy image --download-db-only

# Skip database check
trivy image --skip-update nginx:latest
```

### Offline Mode
```bash
# Download database for offline use
trivy image --download-db-only --db-repository ghcr.io/aquasecurity/trivy-db:2

# Run offline
trivy image --skip-update --db /path/to/db nginx:latest
```

### Performance Tuning
```bash
# Use specific scanners only
trivy image --scanners vuln nginx:latest

# Set parallel processes
trivy image --severity HIGH nginx:latest
```

## Resources

- **Official Documentation**: https://aquasecurity.github.io/trivy/
- **GitHub Repository**: https://github.com/aquasecurity/trivy
- **Docker Images**: https://hub.docker.com/r/aquasec/trivy
- **Slack Community**: [Aqua Security Slack](https://slack.aquasecurity.com/)

## License

Trivy is open-source software licensed under the Apache 2.0 License.
