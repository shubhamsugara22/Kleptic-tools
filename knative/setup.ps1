param(
	[switch]$SkipServing = $false,
	[switch]$SkipEventing = $false,
	[switch]$WithExamples = $false,
	[int]$TimeoutSeconds = 600
)

$ErrorActionPreference = "Stop"

function Write-Info {
	param([string]$Message)
	Write-Host "[*] $Message" -ForegroundColor Yellow
}

function Write-Ok {
	param([string]$Message)
	Write-Host "[+] $Message" -ForegroundColor Green
}

function Write-Err {
	param([string]$Message)
	Write-Host "[!] $Message" -ForegroundColor Red
}

function Wait-NamespacePods {
	param([string]$Namespace)

	Write-Info "Waiting for pods in namespace $Namespace to become Ready"
	kubectl wait --for=condition=Ready pod --all -n $Namespace --timeout "${TimeoutSeconds}s" | Out-Null
}

function Install-Serving {
	Write-Info "Installing Knative Serving CRDs"
	kubectl apply -f https://github.com/knative/serving/releases/latest/download/serving-crds.yaml | Out-Null

	Write-Info "Installing Knative Serving core"
	kubectl apply -f https://github.com/knative/serving/releases/latest/download/serving-core.yaml | Out-Null

	Write-Info "Installing Kourier ingress"
	kubectl apply -f https://github.com/knative/net-kourier/releases/latest/download/kourier.yaml | Out-Null

	Write-Info "Setting Kourier as ingress class"
	kubectl patch configmap/config-network -n knative-serving --type merge --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}' | Out-Null

	Wait-NamespacePods -Namespace "knative-serving"
	Wait-NamespacePods -Namespace "kourier-system"

	Write-Ok "Knative Serving + Kourier installed"
}

function Install-Eventing {
	Write-Info "Installing Knative Eventing CRDs"
	kubectl apply -f https://github.com/knative/eventing/releases/latest/download/eventing-crds.yaml | Out-Null

	Write-Info "Installing Knative Eventing core"
	kubectl apply -f https://github.com/knative/eventing/releases/latest/download/eventing-core.yaml | Out-Null

	Write-Info "Installing in-memory channel"
	kubectl apply -f https://github.com/knative/eventing/releases/latest/download/in-memory-channel.yaml | Out-Null

	Write-Info "Installing MT channel broker"
	kubectl apply -f https://github.com/knative/eventing/releases/latest/download/mt-channel-broker.yaml | Out-Null

	Wait-NamespacePods -Namespace "knative-eventing"

	Write-Ok "Knative Eventing installed"
}

function Deploy-Examples {
	param([string]$ScriptDir)

	Write-Info "Deploying Knative examples"
	kubectl apply -f "$ScriptDir/examples/service-hello.yaml" | Out-Null
	kubectl apply -f "$ScriptDir/examples/service-hello-v2.yaml" | Out-Null

	$previousRevision = kubectl get revision -l serving.knative.dev/service=hello -o jsonpath='{.items[0].metadata.name}'
	(Get-Content "$ScriptDir/examples/service-hello-traffic-split.yaml" -Raw) -replace "PREVIOUS_REVISION_NAME", $previousRevision | kubectl apply -f - | Out-Null

	kubectl apply -f "$ScriptDir/examples/eventing-broker-trigger.yaml" | Out-Null

	Write-Ok "Examples deployed"
}

try {
	kubectl version --client | Out-Null
}
catch {
	Write-Err "kubectl is not installed or not in PATH"
	exit 1
}

try {
	kubectl cluster-info | Out-Null
}
catch {
	Write-Err "Cannot reach Kubernetes cluster. Check kubeconfig/context"
	exit 1
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Write-Info "Starting Knative setup"

if (-not $SkipServing) {
	Install-Serving
}
else {
	Write-Info "Skipping Serving install"
}

if (-not $SkipEventing) {
	Install-Eventing
}
else {
	Write-Info "Skipping Eventing install"
}

if ($WithExamples) {
	Deploy-Examples -ScriptDir $scriptDir
}

Write-Ok "Setup complete"
Write-Info "Run: kubectl get pods -n knative-serving"
Write-Info "Run: kubectl get pods -n knative-eventing"
