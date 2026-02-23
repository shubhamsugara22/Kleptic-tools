# ArgoCD Setup Script for Windows PowerShell
# Install and configure ArgoCD on Kubernetes cluster

param(
    [string]$Namespace = "argocd",
    [string]$Version = "stable"
)

# Colors for output
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Info { Write-Host $args -ForegroundColor Yellow }
function Write-Error { Write-Host $args -ForegroundColor Red }

Write-Success "=== ArgoCD Installation Script for Windows ===`n"

# Check prerequisites
Write-Info "Checking prerequisites..."

# Check if kubectl is installed
try {
    kubectl version --client --short 2>&1 | Out-Null
    Write-Success "✓ kubectl is installed"
} catch {
    Write-Error "kubectl is not installed. Please install kubectl first."
    exit 1
}

# Check cluster connection
try {
    kubectl cluster-info 2>&1 | Out-Null
    Write-Success "✓ Cluster is accessible`n"
} catch {
    Write-Error "Cannot connect to Kubernetes cluster. Check your kubeconfig."
    exit 1
}

# Create namespace
Write-Info "Creating ArgoCD namespace..."
kubectl create namespace $Namespace --dry-run=client -o yaml | kubectl apply -f -
Write-Success "✓ Namespace created`n"

# Install ArgoCD
Write-Info "Installing ArgoCD..."
$installUrl = "https://raw.githubusercontent.com/argoproj/argo-cd/$Version/manifests/install.yaml"
kubectl apply -n $Namespace -f $installUrl

# Wait for ArgoCD to be ready
Write-Info "Waiting for ArgoCD pods to be ready..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n $Namespace --timeout=300s
Write-Success "✓ ArgoCD installed successfully`n"

# Get initial admin password
Write-Info "Retrieving initial admin password..."
$passwordBase64 = kubectl -n $Namespace get secret argocd-initial-admin-secret -o jsonpath="{.data.password}"
$password = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($passwordBase64))
Write-Success "✓ Admin password retrieved`n"

# Access options
Write-Info "How would you like to access ArgoCD?"
Write-Host "1) Port Forward (localhost:8080)"
Write-Host "2) LoadBalancer Service"
Write-Host "3) NodePort Service"
Write-Host "4) Skip (configure manually later)"
$accessChoice = Read-Host "Enter choice [1-4]"

switch ($accessChoice) {
    "1" {
        Write-Info "To access ArgoCD, run the following command in a new terminal:"
        Write-Success "kubectl port-forward svc/argocd-server -n $Namespace 8080:443"
        Write-Success "Then access ArgoCD at: https://localhost:8080"
        
        $startPortForward = Read-Host "Start port forwarding now? (y/n)"
        if ($startPortForward -eq "y") {
            Write-Info "Starting port forward... Press Ctrl+C to stop"
            kubectl port-forward svc/argocd-server -n $Namespace 8080:443
        }
    }
    "2" {
        Write-Info "Patching service to LoadBalancer..."
        kubectl patch svc argocd-server -n $Namespace -p '{\"spec\": {\"type\": \"LoadBalancer\"}}'
        Write-Info "Waiting for external IP..."
        $externalIp = ""
        while ([string]::IsNullOrEmpty($externalIp)) {
            Start-Sleep -Seconds 5
            $externalIp = kubectl get svc argocd-server -n $Namespace -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
        }
        Write-Success "Access ArgoCD at: https://$externalIp"
    }
    "3" {
        Write-Info "Patching service to NodePort..."
        kubectl patch svc argocd-server -n $Namespace -p '{\"spec\": {\"type\": \"NodePort\"}}'
        $nodePort = kubectl get svc argocd-server -n $Namespace -o jsonpath='{.spec.ports[0].nodePort}'
        $nodeIp = kubectl get nodes -o jsonpath='{.items[0].status.addresses[0].address}'
        Write-Success "Access ArgoCD at: https://${nodeIp}:${nodePort}"
    }
    "4" {
        Write-Info "Skipping service configuration"
    }
}

Write-Host ""

# Install ArgoCD CLI
$installCli = Read-Host "Would you like to install ArgoCD CLI? (y/n)"
if ($installCli -eq "y") {
    Write-Info "Installing ArgoCD CLI..."
    
    try {
        # Get latest version
        $latestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/argoproj/argo-cd/releases/latest"
        $version = $latestRelease.tag_name
        
        # Download URL
        $downloadUrl = "https://github.com/argoproj/argo-cd/releases/download/$version/argocd-windows-amd64.exe"
        
        # Download location
        $installPath = "$env:USERPROFILE\AppData\Local\Programs"
        if (-not (Test-Path $installPath)) {
            New-Item -ItemType Directory -Path $installPath -Force | Out-Null
        }
        
        $argocdExe = Join-Path $installPath "argocd.exe"
        
        # Download
        Write-Info "Downloading ArgoCD CLI $version..."
        Invoke-WebRequest -Uri $downloadUrl -OutFile $argocdExe
        
        # Add to PATH if not already there
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($userPath -notlike "*$installPath*") {
            [Environment]::SetEnvironmentVariable("Path", "$userPath;$installPath", "User")
            Write-Info "Added $installPath to PATH. Restart your terminal for changes to take effect."
        }
        
        Write-Success "✓ ArgoCD CLI installed at: $argocdExe"
        
        # Verify installation
        & $argocdExe version --client
    } catch {
        Write-Error "Failed to install ArgoCD CLI: $_"
    }
}

Write-Host ""

# Display credentials
Write-Success "=== ArgoCD Installation Complete ===`n"
Write-Info "Login Credentials:"
Write-Success "Username: admin"
Write-Success "Password: $password"

Write-Host ""
Write-Info "Important: Change the admin password after first login!"
Write-Host "Command: argocd account update-password"
Write-Host ""

# Save credentials to file
$saveCredentials = Read-Host "Save credentials to file? (y/n)"
if ($saveCredentials -eq "y") {
    $credFile = "argocd-credentials.txt"
    @"
ArgoCD Installation Details
===========================
Namespace: $Namespace
Username: admin
Password: $password
Date: $(Get-Date)

Access Commands:
- Port Forward: kubectl port-forward svc/argocd-server -n $Namespace 8080:443
- Get Password: kubectl -n $Namespace get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

Change Password:
argocd account update-password

"@ | Out-File -FilePath $credFile -Encoding UTF8
    Write-Success "✓ Credentials saved to: $credFile"
}

Write-Host ""

# Create sample application
$createSample = Read-Host "Would you like to create a sample guestbook application? (y/n)"
if ($createSample -eq "y") {
    Write-Info "Creating sample application..."
    
    $sampleApp = @"
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: guestbook
  namespace: $Namespace
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps.git
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
"@
    
    $sampleApp | kubectl apply -f -
    
    Write-Success "✓ Sample application created"
    Write-Host "Check the application status in ArgoCD UI or run:"
    Write-Success "kubectl get applications -n $Namespace"
}

Write-Host ""

# Useful commands
Write-Info "=== Useful Commands ==="
Write-Host "1. Get ArgoCD server URL:" -NoNewline; Write-Success " kubectl get svc argocd-server -n $Namespace"
Write-Host "2. List applications:" -NoNewline; Write-Success " kubectl get applications -n $Namespace"
Write-Host "3. Port forward:" -NoNewline; Write-Success " kubectl port-forward svc/argocd-server -n $Namespace 8080:443"
Write-Host "4. View logs:" -NoNewline; Write-Success " kubectl logs -n $Namespace deployment/argocd-server"
Write-Host "5. Delete ArgoCD:" -NoNewline; Write-Success " kubectl delete namespace $Namespace"
Write-Host "6. Login via CLI:" -NoNewline; Write-Success " argocd login localhost:8080 --username admin --password $password --insecure"

Write-Host ""
Write-Success "Setup complete! Happy deploying with ArgoCD! 🚀"
