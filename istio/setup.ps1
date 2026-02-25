# Istio Service Mesh Setup for Windows

# Exit on error
$ErrorActionPreference = "Stop"

# Colors
function Write-Green { Write-Host $args -ForegroundColor Green }
function Write-Yellow { Write-Host $args -ForegroundColor Yellow }
function Write-Red { Write-Host $args -ForegroundColor Red }

Write-Green "Istio Service Mesh Setup"
Write-Host "================================"

# Check if kubectl is installed
try {
	kubectl cluster-info > $null 2>&1
	Write-Green "✓ kubectl found"
	Write-Green "✓ Kubernetes cluster accessible"
}
catch {
	Write-Red "kubectl is not installed or cannot access cluster."
	Write-Red "Please install kubectl and ensure your kubeconfig is configured."
	exit 1
}

# Create temporary directory for download
$TempDir = (New-Item -ItemType Directory -Path "$env:TEMP\istio-setup" -Force).FullName
Push-Location $TempDir

try {
	# Fetch the latest Istio version
	Write-Yellow "Fetching latest Istio version..."
	$ReleaseInfo = Invoke-WebRequest -Uri "https://api.github.com/repos/istio/istio/releases/latest" -UseBasicParsing | ConvertFrom-Json
	$IstioVersion = $ReleaseInfo.tag_name
    
	Write-Host "Latest Istio version: $IstioVersion"

	# Download Istio
	Write-Yellow "Downloading Istio for Windows..."
	$DownloadUrl = "https://github.com/istio/istio/releases/download/$IstioVersion/istio-$IstioVersion-win.zip"
    
	Write-Host "Download URL: $DownloadUrl"
    
	$ZipFile = "$TempDir\istio-$IstioVersion-win.zip"
    
	# Download file
	$ProgressPreference = 'SilentlyContinue'
	Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipFile -UseBasicParsing
    
	Write-Green "✓ Istio downloaded"

	# Extract
	Write-Yellow "Extracting Istio..."
	Expand-Archive -Path $ZipFile -DestinationPath $TempDir -Force
    
	$IstioDir = Get-ChildItem -Directory -Filter "istio-*" | Select-Object -First 1
	Write-Green "✓ Istio extracted"

	# Add istioctl to PATH
	$IstioPath = $IstioDir.FullName
	$Env:PATH = "$IstioPath\bin;$Env:PATH"

	# Create istio-system namespace
	Write-Yellow "Creating istio-system namespace..."
	kubectl create namespace istio-system --dry-run=client -o yaml | kubectl apply -f -
	Write-Green "✓ Namespace created"

	# Install Istio using istioctl
	Write-Yellow "Installing Istio control plane..."
	& istioctl install --set profile=demo -y
    
	if ($LASTEXITCODE -ne 0) {
		Write-Red "Failed to install Istio"
		exit 1
	}
    
	Write-Green "✓ Istio control plane installed"

	# Enable sidecar injection for default namespace
	Write-Yellow "Enabling sidecar injection for default namespace..."
	kubectl label namespace default istio-injection=enabled --overwrite
	Write-Green "✓ Sidecar injection enabled"

	# Wait for Istio components
	Write-Yellow "Waiting for Istio components to be ready..."
	kubectl wait --for=condition=ready pod -l app=istiod -n istio-system --timeout=300s 2>$null | Out-Null

	# Verify installation
	Write-Green "Verifying Istio installation..."
	kubectl get pods -n istio-system

	Write-Green "`n✓ Istio installation completed!`n"
    
	Write-Yellow "Next steps:"
	Write-Host "1. Access Kiali dashboard (UI):"
	Write-Host "   kubectl port-forward svc/kiali -n istio-system 20000:20000"
	Write-Host "   Open: http://localhost:20000"
	Write-Host ""
	Write-Host "2. Access Grafana (Metrics):"
	Write-Host "   kubectl port-forward svc/grafana -n istio-system 3000:3000"
	Write-Host "   Open: http://localhost:3000"
	Write-Host ""
	Write-Host "3. Deploy sample application:"
	Write-Host "   kubectl apply -f $IstioPath\samples\bookinfo\platform\kube\bookinfo.yaml"
	Write-Host ""
	Write-Host "4. Check sidecar injection status:"
	Write-Host "   kubectl get namespace default --show-labels"
	Write-Host ""
	Write-Host "5. View Istio configuration:"
	Write-Host "   istioctl analyze"
	Write-Host ""
	Write-Yellow "To uninstall Istio:"
	Write-Host "   istioctl uninstall --purge"

	Write-Yellow "`nNote: istioctl path: $IstioPath\bin"
	Write-Yellow "Add to your PATH for permanent access, or use full path to istioctl commands."

}
finally {
	Pop-Location
}
