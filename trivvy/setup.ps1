# Trivy Installation and Setup Script for Windows PowerShell

param(
	[switch]$SkipDatabase = $false,
	[switch]$CreateConfig = $true
)

$ErrorActionPreference = "Stop"

# Color functions for output
function Write-Status {
	param([string]$Message)
	Write-Host "[*] $Message" -ForegroundColor Green
}

function Write-Error-Custom {
	param([string]$Message)
	Write-Host "[!] $Message" -ForegroundColor Red
}

function Write-Warning-Custom {
	param([string]$Message)
	Write-Host "[!] $Message" -ForegroundColor Yellow
}

# Check for prerequisites
function Test-Prerequisites {
	Write-Status "Checking prerequisites..."
    
	# Check for Chocolatey or Scoop
	$hasChoco = Get-Command choco -ErrorAction SilentlyContinue
	$hasScoop = Get-Command scoop -ErrorAction SilentlyContinue
    
	if (-not $hasChoco -and -not $hasScoop) {
		Write-Warning-Custom "Neither Chocolatey nor Scoop found"
		Write-Status "Installing Scoop..."
        
		# Check PowerShell execution policy
		$policy = Get-ExecutionPolicy
		if ($policy -ne "Unrestricted" -and $policy -ne "Bypass") {
			Write-Warning-Custom "Current execution policy: $policy"
			Write-Status "Setting execution policy to Bypass for current user..."
			Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope CurrentUser -Force
		}
        
		# Install Scoop
		iex "& {$(irm get.scoop.sh)}"
        
		return $true
	}
    
	return $false
}

# Install Trivy using Scoop
function Install-TriviScoop {
	Write-Status "Installing Trivy using Scoop..."
    
	try {
		scoop bucket add security https://github.com/aquasecurity/scoop-trivy.git
		scoop install trivy
		Write-Status "Trivy installed successfully via Scoop"
		return $true
	}
	catch {
		Write-Error-Custom "Failed to install Trivy via Scoop: $_"
		return $false
	}
}

# Install Trivy using Chocolatey
function Install-TrivyChocolatey {
	Write-Status "Installing Trivy using Chocolatey..."
    
	try {
		choco install trivy -y
		Write-Status "Trivy installed successfully via Chocolatey"
		return $true
	}
	catch {
		Write-Error-Custom "Failed to install Trivy via Chocolatey: $_"
		return $false
	}
}

# Install Trivy manually from GitHub releases
function Install-TrivyManual {
	Write-Status "Downloading Trivy from GitHub releases..."
    
	try {
		$arch = if ([System.Environment]::Is64BitOperatingSystem) { "win-64" } else { "win-32" }
		$version = "latest"  # or specify a specific version like "v0.45.0"
        
		$url = "https://api.github.com/repos/aquasecurity/trivy/releases/$version"
		$release = (Invoke-RestMethod $url)
        
		$assetName = "trivy_*_windows-$arch.zip"
		$asset = $release.assets | Where-Object { $_.name -like $assetName } | Select-Object -First 1
        
		if (-not $asset) {
			Write-Error-Custom "Trivy binary not found for architecture: $arch"
			return $false
		}
        
		$downloadUrl = $asset.browser_download_url
		$outputPath = "$env:TEMP\trivy.zip"
        
		Write-Status "Downloading from: $downloadUrl"
		Invoke-WebRequest -Uri $downloadUrl -OutFile $outputPath
        
		# Create installation directory
		$installDir = "$env:ProgramFiles\trivy"
		if (-not (Test-Path $installDir)) {
			New-Item -ItemType Directory -Path $installDir -Force | Out-Null
		}
        
		# Extract
		Write-Status "Extracting Trivy..."
		Expand-Archive -Path $outputPath -DestinationPath $installDir -Force
        
		# Add to PATH if not already there
		$pathEnvVar = [Environment]::GetEnvironmentVariable("PATH", "User")
		if ($pathEnvVar -notmatch [regex]::Escape($installDir)) {
			Write-Status "Adding Trivy to PATH..."
			$newPath = "$pathEnvVar;$installDir"
			[Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
		}
        
		# Cleanup
		Remove-Item $outputPath -Force
        
		Write-Status "Trivy installed successfully at: $installDir"
		Write-Status "Please restart PowerShell for PATH changes to take effect"
		return $true
	}
	catch {
		Write-Error-Custom "Failed to install Trivy manually: $_"
		return $false
	}
}

# Setup database
function Setup-Database {
	Write-Status "Setting up Trivy database..."
    
	try {
		$cacheDir = "$env:USERPROFILE\.cache\trivy"
		if (-not (Test-Path $cacheDir)) {
			New-Item -ItemType Directory -Path $cacheDir -Force | Out-Null
		}
        
		Write-Status "Downloading vulnerability database (this may take a few minutes)..."
		& trivy image --download-db-only
        
		Write-Status "Database setup complete"
		return $true
	}
	catch {
		Write-Error-Custom "Failed to setup database: $_"
		return $false
	}
}

# Create sample configuration file
function Create-ConfigFile {
	Write-Status "Creating sample configuration file..."
    
	$configContent = @"
# Trivy Configuration File for Windows

# Severity levels to report
severity:
  - CRITICAL
  - HIGH
  - MEDIUM

# Skip updating the database
skip-update: false

# Output format: table, json, sarif, cyclonedx, spdx, csv
format: "table"

# Scanners to use: vuln, config, secret, license
scanners:
  - vuln
  - config
  - secret

# Vulnerability source
vuln-source: "default"

# Cache directory
# Note: On Windows, use forward slashes or double backslashes
# cache:
#   dir: C:/Users/YourUsername/.cache/trivy

# Exit code to return when vulnerabilities/misconfigurations are found
exit-code: 1

# Skip specific directories
skip-dirs:
  - node_modules
  - vendor
  - packages

# Timeout for scanning
timeout: 10m
"@
    
	try {
		$configContent | Out-File -FilePath "trivy.yaml" -Encoding UTF8
		Write-Status "Configuration file created: trivy.yaml"
		return $true
	}
	catch {
		Write-Error-Custom "Failed to create config file: $_"
		return $false
	}
}

# Create .trivyignore file
function Create-IgnoreFile {
	Write-Status "Creating sample .trivyignore file..."
    
	$ignoreContent = @"
# Trivy Ignore File
# 
# Ignore specific vulnerabilities by CVE ID:
# CVE-2022-12345
#
# Ignore specific misconfigurations by AVD ID:
# AVD-AWS-0001
#
# Add one entry per line
"@
    
	try {
		$ignoreContent | Out-File -FilePath ".trivyignore" -Encoding UTF8
		Write-Status "Ignore file created: .trivyignore"
		return $true
	}
	catch {
		Write-Error-Custom "Failed to create ignore file: $_"
		return $false
	}
}

# Test Trivy installation
function Test-Installation {
	Write-Status "Testing Trivy installation..."
    
	try {
		$version = & trivy --version
		Write-Status "Trivy version: $version"
		return $true
	}
	catch {
		Write-Error-Custom "Trivy not found in PATH or execution failed"
		return $false
	}
}

# Main installation flow
function Main {
	Write-Status "Starting Trivy installation and setup for Windows..."
	Write-Status ""
    
	# Check prerequisites
	$installed = Test-Prerequisites
    
	# Install Trivy
	$installSuccess = $false
    
	# Try Scoop first
	if (Get-Command scoop -ErrorAction SilentlyContinue) {
		$installSuccess = Install-TriviScoop
	}
	elseif (Get-Command choco -ErrorAction SilentlyContinue) {
		$installSuccess = Install-TrivyChocolatey
	}
	else {
		$installSuccess = Install-TrivyManual
	}
    
	if (-not $installSuccess) {
		Write-Error-Custom "Trivy installation failed"
		exit 1
	}
    
	# Test installation
	if (-not (Test-Installation)) {
		Write-Error-Custom "Trivy installation test failed"
		exit 1
	}
    
	# Setup database
	if (-not $SkipDatabase) {
		Setup-Database
	}
    
	# Create configuration files
	if ($CreateConfig) {
		Create-ConfigFile
		Create-IgnoreFile
	}
    
	Write-Status ""
	Write-Status "Trivy setup complete!"
	Write-Status ""
	Write-Host "Quick start examples:" -ForegroundColor Yellow
	Write-Host "  Scan container image:"
	Write-Host "    trivy image nginx:latest"
	Write-Host ""
	Write-Host "  Scan filesystem:"
	Write-Host "    trivy fs ."
	Write-Host ""
	Write-Host "  Generate SBOM:"
	Write-Host "    trivy image --format cyclonedx nginx:latest > sbom.xml"
	Write-Host ""
	Write-Status "For more information, visit: https://aquasecurity.github.io/trivy/"
}

# Run main function
Main
