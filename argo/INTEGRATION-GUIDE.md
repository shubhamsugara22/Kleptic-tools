# ArgoCD Integration Guide - CI/CD, Secrets, Image Updates

## Table of Contents
- [CI/CD Integration](#cicd-integration)
- [Secrets Management](#secrets-management)
- [ArgoCD Image Updater](#argocd-image-updater)
- [Notifications Configuration](#notifications-configuration)

## CI/CD Integration

### 1. GitHub Actions Integration

```yaml
# .github/workflows/deploy.yml
name: Build and Deploy

on:
  push:
    branches: [ main ]
    
env:
  ARGOCD_SERVER: argocd.example.com
  APP_NAME: myapp

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Login to Container Registry
      uses: docker/login-action@v2
      with:
        registry: ghcr.io
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Build and push
      uses: docker/build-push-action@v4
      with:
        context: .
        push: true
        tags: |
          ghcr.io/${{ github.repository }}:${{ github.sha }}
          ghcr.io/${{ github.repository }}:latest
        cache-from: type=gha
        cache-to: type=gha,mode=max
  
  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Update image tag in manifests
      run: |
        cd k8s
        sed -i "s|image:.*|image: ghcr.io/${{ github.repository }}:${{ github.sha }}|g" deployment.yaml
        git config user.name "GitHub Actions"
        git config user.email "actions@github.com"
        git add .
        git commit -m "Update image to ${{ github.sha }}"
        git push
    
    - name: Install ArgoCD CLI
      run: |
        curl -sSL -o argocd https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
        chmod +x argocd
        sudo mv argocd /usr/local/bin/
    
    - name: Sync ArgoCD Application
      env:
        ARGOCD_AUTH_TOKEN: ${{ secrets.ARGOCD_TOKEN }}
      run: |
        argocd app sync $APP_NAME --server $ARGOCD_SERVER --auth-token $ARGOCD_AUTH_TOKEN
        argocd app wait $APP_NAME --server $ARGOCD_SERVER --auth-token $ARGOCD_AUTH_TOKEN --health
```

### 2. GitLab CI Integration

```yaml
# .gitlab-ci.yml
stages:
  - build
  - deploy

variables:
  IMAGE_TAG: $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  ARGOCD_SERVER: argocd.example.com

build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
    - docker build -t $IMAGE_TAG .
    - docker push $IMAGE_TAG
  only:
    - main

update-manifest:
  stage: deploy
  image: alpine/git
  before_script:
    - apk add --no-cache sed
  script:
    - git clone https://${CI_USERNAME}:${CI_ACCESS_TOKEN}@gitlab.com/org/k8s-manifests.git
    - cd k8s-manifests
    - sed -i "s|image:.*|image: $IMAGE_TAG|g" deployment.yaml
    - git config user.email "ci@gitlab.com"
    - git config user.name "GitLab CI"
    - git add .
    - git commit -m "Update image to $CI_COMMIT_SHA"
    - git push origin main
  only:
    - main

argocd-sync:
  stage: deploy
  image: argoproj/argocd:latest
  script:
    - argocd login $ARGOCD_SERVER --auth-token $ARGOCD_TOKEN --insecure
    - argocd app sync myapp --force
    - argocd app wait myapp --health
  only:
    - main
```

### 3. Jenkins Pipeline

```groovy
// Jenkinsfile
pipeline {
    agent any
    
    environment {
        DOCKER_REGISTRY = 'registry.example.com'
        ARGOCD_SERVER = 'argocd.example.com'
        APP_NAME = 'myapp'
        IMAGE_TAG = "${env.BUILD_ID}"
    }
    
    stages {
        stage('Build') {
            steps {
                script {
                    docker.build("${DOCKER_REGISTRY}/${APP_NAME}:${IMAGE_TAG}")
                }
            }
        }
        
        stage('Push') {
            steps {
                script {
                    docker.withRegistry("https://${DOCKER_REGISTRY}", 'docker-credentials') {
                        docker.image("${DOCKER_REGISTRY}/${APP_NAME}:${IMAGE_TAG}").push()
                        docker.image("${DOCKER_REGISTRY}/${APP_NAME}:${IMAGE_TAG}").push('latest')
                    }
                }
            }
        }
        
        stage('Update Manifests') {
            steps {
                script {
                    git credentialsId: 'git-credentials', url: 'https://github.com/org/k8s-manifests.git'
                    sh """
                        sed -i "s|image:.*|image: ${DOCKER_REGISTRY}/${APP_NAME}:${IMAGE_TAG}|g" k8s/deployment.yaml
                        git config user.email 'jenkins@example.com'
                        git config user.name 'Jenkins'
                        git add .
                        git commit -m 'Update image to ${IMAGE_TAG}'
                        git push origin main
                    """
                }
            }
        }
        
        stage('Sync ArgoCD') {
            steps {
                withCredentials([string(credentialsId: 'argocd-token', variable: 'ARGOCD_TOKEN')]) {
                    sh """
                        argocd login ${ARGOCD_SERVER} --auth-token ${ARGOCD_TOKEN} --insecure
                        argocd app sync ${APP_NAME}
                        argocd app wait ${APP_NAME} --health
                    """
                }
            }
        }
    }
    
    post {
        success {
            slackSend color: 'good', message: "Deployment successful: ${APP_NAME} - ${IMAGE_TAG}"
        }
        failure {
            slackSend color: 'danger', message: "Deployment failed: ${APP_NAME} - ${IMAGE_TAG}"
        }
    }
}
```

### 4. Azure DevOps Pipeline

```yaml
# azure-pipelines.yml
trigger:
  - main

variables:
  imageTag: '$(Build.BuildId)'
  imageName: 'myapp'
  argocdServer: 'argocd.example.com'
  argocdApp: 'myapp'

stages:
- stage: Build
  jobs:
  - job: BuildAndPush
    pool:
      vmImage: 'ubuntu-latest'
    steps:
    - task: Docker@2
      inputs:
        containerRegistry: 'ACR Connection'
        repository: $(imageName)
        command: 'buildAndPush'
        Dockerfile: '**/Dockerfile'
        tags: |
          $(imageTag)
          latest

- stage: Deploy
  jobs:
  - job: UpdateManifests
    pool:
      vmImage: 'ubuntu-latest'
    steps:
    - checkout: self
      persistCredentials: true
    
    - script: |
        cd k8s
        sed -i "s|image:.*|image: $(containerRegistry)/$(imageName):$(imageTag)|g" deployment.yaml
        git config user.email "azure@devops.com"
        git config user.name "Azure DevOps"
        git add .
        git commit -m "Update image to $(imageTag)"
        git push
      displayName: 'Update image tag'
  
  - job: ArgoSync
    dependsOn: UpdateManifests
    pool:
      vmImage: 'ubuntu-latest'
    steps:
    - script: |
        curl -sSL -o argocd https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
        chmod +x argocd
        sudo mv argocd /usr/local/bin/
      displayName: 'Install ArgoCD CLI'
    
    - script: |
        argocd login $(argocdServer) --auth-token $(argocdToken) --insecure
        argocd app sync $(argocdApp)
        argocd app wait $(argocdApp) --health
      displayName: 'Sync ArgoCD Application'
      env:
        argocdToken: $(ARGOCD_TOKEN)
```

## Secrets Management

### 1. Sealed Secrets

#### Installation

```bash
# Install Sealed Secrets controller
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.0/controller.yaml

# Install kubeseal CLI
wget https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.0/kubeseal-linux-amd64
sudo install -m 755 kubeseal-linux-amd64 /usr/local/bin/kubeseal
```

#### Usage

```bash
# Create a secret
kubectl create secret generic mysecret \
  --from-literal=username=admin \
  --from-literal=password=secretpass \
  --dry-run=client -o yaml > secret.yaml

# Seal the secret
kubeseal --format yaml < secret.yaml > sealed-secret.yaml

# sealed-secret.yaml can be committed to Git
# Apply sealed secret
kubectl apply -f sealed-secret.yaml
```

#### ArgoCD Integration

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: argocd
spec:
  source:
    repoURL: https://github.com/org/repo.git
    targetRevision: HEAD
    path: k8s
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true

# In your repo: k8s/sealed-secret.yaml
---
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: mysecret
  namespace: production
spec:
  encryptedData:
    username: AgBR7P...
    password: AgCQ9K...
```

### 2. External Secrets Operator (ESO)

#### Installation

```bash
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets \
  external-secrets/external-secrets \
  -n external-secrets-system \
  --create-namespace
```

#### AWS Secrets Manager Integration

```yaml
# SecretStore - connects to AWS Secrets Manager
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secretstore
  namespace: production
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets-sa

---
# ExternalSecret - syncs secret from AWS
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: database-credentials
  namespace: production
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secretstore
    kind: SecretStore
  target:
    name: db-secret
    creationPolicy: Owner
  data:
  - secretKey: username
    remoteRef:
      key: prod/database
      property: username
  - secretKey: password
    remoteRef:
      key: prod/database
      property: password
```

#### Azure Key Vault Integration

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: azure-secretstore
  namespace: production
spec:
  provider:
    azurekv:
      authType: ManagedIdentity
      vaultUrl: "https://my-vault.vault.azure.net"
      identityId: "/subscriptions/xxx/resourceGroups/xxx/providers/Microsoft.ManagedIdentity/userAssignedIdentities/xxx"

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: app-secrets
  namespace: production
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: azure-secretstore
    kind: SecretStore
  target:
    name: app-secret
  data:
  - secretKey: api-key
    remoteRef:
      key: api-key
```

### 3. HashiCorp Vault Integration

#### Install Vault Agent Injector

```bash
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install vault hashicorp/vault \
  --set "injector.enabled=true" \
  --set "server.enabled=false"
```

#### ArgoCD with Vault

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: production
spec:
  template:
    metadata:
      annotations:
        vault.hashicorp.com/agent-inject: "true"
        vault.hashicorp.com/role: "myapp"
        vault.hashicorp.com/agent-inject-secret-config: "secret/data/myapp/config"
        vault.hashicorp.com/agent-inject-template-config: |
          {{- with secret "secret/data/myapp/config" -}}
          export DATABASE_URL="{{ .Data.data.database_url }}"
          export API_KEY="{{ .Data.data.api_key }}"
          {{- end }}
    spec:
      serviceAccountName: myapp
      containers:
      - name: app
        image: myapp:latest
        command: ["/bin/sh"]
        args:
        - -c
        - source /vault/secrets/config && ./app
```

#### Vault Plugin for ArgoCD

```yaml
# Install argocd-vault-plugin
apiVersion: v1
kind: ConfigMap
metadata:
  name: cmp-plugin
  namespace: argocd
data:
  avp.yaml: |
    apiVersion: argoproj.io/v1alpha1
    kind: ConfigManagementPlugin
    metadata:
      name: argocd-vault-plugin
    spec:
      allowConcurrency: true
      discover:
        find:
          command:
            - sh
            - "-c"
            - "find . -name '*.yaml' | xargs -I {} grep \"<path\\|avp\\.kubernetes\\.io\" {} | grep ."
      generate:
        command:
          - argocd-vault-plugin
          - generate
          - ./
      lockRepo: false

# Usage in manifests
---
apiVersion: v1
kind: Secret
metadata:
  name: myapp-secret
  namespace: production
type: Opaque
stringData:
  username: <path:secret/data/myapp#username>
  password: <path:secret/data/myapp#password>
```

### 4. SOPS (Secrets Operations)

```bash
# Install SOPS
wget https://github.com/mozilla/sops/releases/download/v3.8.1/sops-v3.8.1.linux.amd64
sudo mv sops-v3.8.1.linux.amd64 /usr/local/bin/sops
sudo chmod +x /usr/local/bin/sops

# Encrypt a secret
sops --encrypt --age <AGE_PUBLIC_KEY> secret.yaml > secret.enc.yaml

# Decrypt (for viewing)
sops --decrypt secret.enc.yaml

# ArgoCD with SOPS (using Helm Secrets plugin)
# Install argo-cd helm-secrets plugin
```

## ArgoCD Image Updater

### 1. Installation

```bash
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj-labs/argocd-image-updater/stable/manifests/install.yaml
```

### 2. Basic Configuration

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: argocd
  annotations:
    # Enable image updater
    argocd-image-updater.argoproj.io/image-list: myimage=myregistry/myapp
    
    # Update strategy: latest, semver, digest
    argocd-image-updater.argoproj.io/myimage.update-strategy: semver
    
    # Constraint (semver only)
    argocd-image-updater.argoproj.io/myimage.allow-tags: regexp:^v[0-9]+\.[0-9]+\.[0-9]+$
    
    # Write back method: git or argocd
    argocd-image-updater.argoproj.io/write-back-method: git
    
    # Git branch to update
    argocd-image-updater.argoproj.io/git-branch: main
spec:
  source:
    repoURL: https://github.com/org/repo.git
    targetRevision: HEAD
    path: k8s
    helm:
      parameters:
      - name: image.tag
        value: v1.0.0
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

### 3. Advanced Image Update Patterns

```yaml
# Multiple images
metadata:
  annotations:
    argocd-image-updater.argoproj.io/image-list: |
      frontend=myregistry/frontend,
      backend=myregistry/backend,
      api=myregistry/api
    argocd-image-updater.argoproj.io/frontend.update-strategy: latest
    argocd-image-updater.argoproj.io/backend.update-strategy: semver:~1.2
    argocd-image-updater.argoproj.io/api.update-strategy: digest

# Private registry authentication
argocd-image-updater.argoproj.io/myimage.pull-secret: pullsecret:argocd/my-registry-secret

# Custom Git commit message
argocd-image-updater.argoproj.io/git-commit-message: |
  build: update {{.Image}} to {{.NewTag}}
  
  Generated by ArgoCD Image Updater
```

## Notifications Configuration

### 1. Install Notifications

```bash
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj-labs/argocd-notifications/stable/manifests/install.yaml
```

### 2. Slack Integration

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: argocd-notifications-secret
  namespace: argocd
stringData:
  slack-token: xoxb-your-slack-token

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
  namespace: argocd
data:
  service.slack: |
    token: $slack-token
  
  template.app-deployed: |
    message: |
      Application {{.app.metadata.name}} is now running new version.
    slack:
      attachments: |
        [{
          "title": "{{ .app.metadata.name}}",
          "title_link":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
          "color": "#18be52",
          "fields": [
          {
            "title": "Sync Status",
            "value": "{{.app.status.sync.status}}",
            "short": true
          },
          {
            "title": "Repository",
            "value": "{{.app.spec.source.repoURL}}",
            "short": true
          }
          ]
        }]
  
  template.app-health-degraded: |
    message: |
      Application {{.app.metadata.name}} has degraded.
    slack:
      attachments: |
        [{
          "title": "{{ .app.metadata.name}}",
          "title_link": "{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
          "color": "#f4c030",
          "fields": [
          {
            "title": "Health Status",
            "value": "{{.app.status.health.status}}",
            "short": true
          }
          ]
        }]
  
  trigger.on-deployed: |
    - description: Application is synced and healthy
      send:
      - app-deployed
      when: app.status.operationState.phase in ['Succeeded'] and app.status.health.status == 'Healthy'
  
  trigger.on-health-degraded: |
    - description: Application has degraded
      send:
      - app-health-degraded
      when: app.status.health.status == 'Degraded'
```

### 3. Subscribe Applications

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: argocd
  annotations:
    notifications.argoproj.io/subscribe.on-deployed.slack: dev-team-channel
    notifications.argoproj.io/subscribe.on-health-degraded.slack: alerts-channel
spec:
  # ... rest of application spec
```

### 4. Microsoft Teams Integration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
  namespace: argocd
data:
  service.teams: |
    recipientUrls:
      devTeam: https://outlook.office.com/webhook/xxx
  
  template.app-deployed: |
    teams:
      themeColor: "#000080"
      summary: "Application {{.app.metadata.name}} deployed"
      sections: |
        [{
          "activityTitle": "Application {{.app.metadata.name}} deployed",
          "facts": [
            {
              "name": "Sync Status",
              "value": "{{.app.status.sync.status}}"
            },
            {
              "name": "Repository",
              "value": "{{.app.spec.source.repoURL}}"
            }
          ]
        }]
```

### 5. Email Notifications

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: argocd-notifications-secret
  namespace: argocd
stringData:
  email-username: smtp-user@example.com
  email-password: smtp-password

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
  namespace: argocd
data:
  service.email.gmail: |
    username: $email-username
    password: $email-password
    host: smtp.gmail.com
    port: 587
    from: $email-username
  
  template.app-sync-failed: |
    email:
      subject: ArgoCD Sync Failed - {{.app.metadata.name}}
    message: |
      Application {{.app.metadata.name}} sync has failed.
      
      Sync Status: {{.app.status.sync.status}}
      Health Status: {{.app.status.health.status}}
      
      View in ArgoCD: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}
```

## Additional Resources

- [ArgoCD Best Practices](https://argo-cd.readthedocs.io/en/stable/user-guide/best_practices/)
- [Image Updater Documentation](https://argocd-image-updater.readthedocs.io/)
- [Notifications Documentation](https://argocd-notifications.readthedocs.io/)
- [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)
- [External Secrets Operator](https://external-secrets.io/)
