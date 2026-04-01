# Trivy Misconfiguration Scanning Examples

This guide provides ready-to-run Trivy commands focused on misconfiguration scanning.

## What It Scans

Trivy config scanning checks infrastructure and deployment definitions for insecure settings, including:
- Kubernetes manifests
- Terraform
- Dockerfile
- Helm-rendered YAML

## Quick Commands

```bash
# Scan all supported config files in current directory
trivy config .

# Show only high-impact findings
trivy config --severity HIGH,CRITICAL .

# Fail with non-zero exit code when issues are found (CI/CD)
trivy config --severity HIGH,CRITICAL --exit-code 1 .
```

## Targeted Scanning

```bash
# Kubernetes only
trivy config --misconfig-scanners kubernetes .

# Terraform only
trivy config --misconfig-scanners terraform .

# Dockerfile only
trivy config --misconfig-scanners dockerfile .

# Single file
trivy config ./k8s/deployment.yaml
```

## Output for Automation

```bash
# JSON output for pipelines and tooling
trivy config --format json --output trivy-misconfig.json .

# SARIF output for security dashboards
trivy config --format sarif --output trivy-misconfig.sarif .
```

## Noise Reduction

```bash
# Skip common non-relevant directories
trivy config --skip-dirs .git --skip-dirs node_modules --skip-dirs vendor .
```

Use a .trivyignore file to suppress accepted findings:

```text
AVD-KSV-0012
AVD-DS-0002
```

Then run:

```bash
trivy config .
```

## CI/CD Example (GitHub Actions)

```yaml
- name: Trivy misconfiguration scan
  uses: aquasecurity/trivy-action@master
  with:
    scan-type: 'config'
    scan-ref: '.'
    severity: 'HIGH,CRITICAL'
    format: 'sarif'
    output: 'trivy-config.sarif'

- name: Upload SARIF
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: 'trivy-config.sarif'
```

## Common Findings to Expect

- Privileged containers
- Missing runAsNonRoot
- Missing readOnlyRootFilesystem
- Missing CPU and memory limits
- Open network exposure
- Missing encryption settings in IaC

## Repo-Focused Examples

From the repository root:

```bash
# Scan Argo manifests
trivy config ./argo

# Scan Helm chart templates and values
trivy config ./helm

# Scan everything with strict gate
trivy config --severity HIGH,CRITICAL --exit-code 1 .
```
