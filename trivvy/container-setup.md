# Trivy - Basic Container Setup

This quick guide shows the simplest way to scan a container image using Trivy.

## Option 1: Run Trivy as a container

Use this when you do not want to install Trivy locally.

```bash
# Scan a public image
# Windows PowerShell users: replace $PWD with ${PWD} or an absolute path

docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $PWD/.trivy-cache:/root/.cache/ \
  aquasec/trivy:latest image nginx:latest
```

Notes:
- The Docker socket mount lets Trivy inspect local images.
- The cache mount speeds up repeated scans.

## Option 2: Install Trivy and scan locally

```bash
# Verify installation
trivy --version

# Scan a local image
docker pull nginx:latest
trivy image nginx:latest
```

## Common flags

```bash
# Only show HIGH and CRITICAL issues
trivy image --severity HIGH,CRITICAL nginx:latest

# Output JSON for CI pipelines
trivy image --format json --output trivy-results.json nginx:latest

# Do not update the database (useful for air-gapped or fast re-runs)
trivy image --skip-update nginx:latest
```

## Simple CI example (GitHub Actions)

```yaml
- name: Trivy image scan
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: 'nginx:latest'
    severity: 'HIGH,CRITICAL'
    format: 'sarif'
    output: 'trivy-results.sarif'
```

## Troubleshooting

- If you see permission errors, make sure your user can access the Docker socket.
- If the scan is slow on first run, the vulnerability database is downloading.
