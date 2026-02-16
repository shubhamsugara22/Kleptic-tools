# Kleptic App Helm Chart

A Helm chart for deploying and managing Kleptic applications on Kubernetes.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+

## Installation

Add the repository and install:

```bash
helm repo add kleptic <repo-url>
helm repo update
helm install kleptic-release kleptic/kleptic-app
```

Or install from local chart:

```bash
helm install kleptic-release ./helm
```

## Configuration

The following table lists the configurable parameters of the chart and their default values.

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `replicaCount` | Number of pod replicas | `1` |
| `image.repository` | Container image repository | `nginx` |
| `image.tag` | Container image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `service.type` | Kubernetes Service type | `ClusterIP` |
| `service.port` | Service port | `80` |
| `service.targetPort` | Container target port | `8080` |
| `resources.requests.cpu` | CPU request | `50m` |
| `resources.requests.memory` | Memory request | `64Mi` |
| `resources.limits.cpu` | CPU limit | `100m` |
| `resources.limits.memory` | Memory limit | `128Mi` |
| `autoscaling.enabled` | Enable horizontal pod autoscaler | `false` |
| `serviceAccount.create` | Enable service account creation | `true` |

## Usage

### Basic Deployment

```bash
helm install my-release ./helm
```

### Custom Values

```bash
helm install my-release ./helm --values custom-values.yaml
```

### Upgrade

```bash
helm upgrade my-release ./helm
```

### Uninstall

```bash
helm uninstall my-release
```

## Customization

### Override Image

```bash
helm install my-release ./helm \
  --set image.repository=myrepo/myimage \
  --set image.tag=v1.0
```

### Enable Autoscaling

```bash
helm install my-release ./helm \
  --set autoscaling.enabled=true \
  --set autoscaling.minReplicas=2 \
  --set autoscaling.maxReplicas=5
```

### Set Environment Variables

Add to `values.yaml` or pass via CLI:

```bash
helm install my-release ./helm \
  --set env[0].name=ENV_VAR \
  --set env[0].value=myvalue
```

## Templates

- **deployment.yaml** - Kubernetes Deployment
- **service.yaml** - Kubernetes Service  
- **serviceaccount.yaml** - Service Account
- **configmap.yaml** - ConfigMap (optional)
- **_helpers.tpl** - Template helpers and functions

## License

See LICENSE file for details.
