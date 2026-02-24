#!/bin/bash

# Trivy Installation and Setup Script
# This script installs and configures Trivy security scanner

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status messages
print_status() {
    echo -e "${GREEN}[*]${NC} $1"
}

print_error() {
    echo -e "${RED}[!]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

# Detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        ARCH=$(uname -m)
        if [[ $ARCH == "x86_64" ]]; then
            ARCH="amd64"
        elif [[ $ARCH == "aarch64" ]]; then
            ARCH="arm64"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
        ARCH=$(uname -m)
        if [[ $ARCH == "arm64" ]]; then
            ARCH="arm64"
        else
            ARCH="amd64"
        fi
    else
        print_error "Unsupported OS: $OSTYPE"
        exit 1
    fi
}

# Install Trivy on Linux (apt)
install_apt() {
    print_status "Installing Trivy using apt..."
    
    sudo apt-get update
    sudo apt-get install -y wget apt-transport-https gnupg lsb-release
    
    # Add GPG key
    wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo apt-key add -
    
    # Add repository
    echo "deb https://aquasecurity.github.io/trivy-repo/deb $(lsb_release -sc) main" | sudo tee -a /etc/apt/sources.list.d/trivy.list
    
    # Install
    sudo apt-get update
    sudo apt-get install -y trivy
    
    print_status "Trivy installed successfully via apt"
}

# Install Trivy on macOS using Homebrew
install_homebrew() {
    print_status "Installing Trivy using Homebrew..."
    
    if ! command -v brew &> /dev/null; then
        print_error "Homebrew is not installed. Please install Homebrew first."
        exit 1
    fi
    
    brew install trivy
    print_status "Trivy installed successfully via Homebrew"
}

# Install Trivy using curl
install_curl() {
    print_status "Installing Trivy using curl..."
    
    curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
    
    print_status "Trivy installed successfully via curl"
}

# Download and cache vulnerability database
setup_database() {
    print_status "Setting up Trivy database..."
    
    # Create cache directory
    mkdir -p ~/.cache/trivy
    
    # Download vulnerability database
    trivy image --download-db-only
    
    print_status "Database setup complete"
}

# Create sample configuration file
create_config() {
    print_status "Creating sample configuration file..."
    
    cat > trivy.yaml << 'EOF'
# Trivy Configuration File

# Severity levels to report
severity:
  - CRITICAL
  - HIGH
  - MEDIUM

# Skip updating the database
skip-update: false

# Use offline mode
offline-db: false

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
cache:
  dir: /home/user/.cache/trivy

# Exit code to return when vulnerabilities/misconfigurations are found
exit-code: 1

# Ignore specific vulnerabilities
ignorefile: ".trivyignore"

# Custom policy directory (for configuration checks)
skip-dirs:
  - node_modules
  - vendor
  - venv

# Timeout for VM image scanning
timeout: 10m
EOF
    
    print_status "Configuration file created: trivy.yaml"
}

# Create sample .trivyignore file
create_ignore_file() {
    print_status "Creating sample .trivyignore file..."
    
    cat > .trivyignore << 'EOF'
# Example: Ignore specific vulnerability IDs
# CVE-2022-12345
# AVD-AWS-0001

# Example: Ignore by resource type
# /path/to/dockerfile

# Format:
# CVE-XXXX-XXXXX (for vulnerabilities)
# AVD-XXXX-XXXX (for misconfigurations)
EOF
    
    print_status "Ignore file created: .trivyignore"
}

# Test Trivy installation
test_installation() {
    print_status "Testing Trivy installation..."
    
    if ! command -v trivy &> /dev/null; then
        print_error "Trivy installation failed or not in PATH"
        return 1
    fi
    
    TRIVY_VERSION=$(trivy --version)
    print_status "Trivy version: $TRIVY_VERSION"
    
    return 0
}

# Main installation flow
main() {
    print_status "Starting Trivy installation and setup..."
    
    detect_os
    print_status "Detected OS: $OS ($ARCH)"
    
    # Install based on OS
    if [[ $OS == "linux" ]]; then
        if command -v apt-get &> /dev/null; then
            install_apt
        elif command -v yum &> /dev/null; then
            print_warning "YUM-based systems not fully supported in this script"
            print_status "Please follow manual installation: https://aquasecurity.github.io/trivy/latest/getting-started/installation/"
            exit 1
        else
            install_curl
        fi
    elif [[ $OS == "macos" ]]; then
        install_homebrew
    fi
    
    # Test installation
    if test_installation; then
        # Setup database and config
        setup_database
        create_config
        create_ignore_file
        
        print_status "Trivy setup complete!"
        print_status ""
        print_status "Quick start examples:"
        echo -e "${YELLOW}  Scan container image:${NC}"
        echo "    trivy image nginx:latest"
        echo ""
        echo -e "${YELLOW}  Scan filesystem:${NC}"
        echo "    trivy fs ."
        echo ""
        echo -e "${YELLOW}  Scan Git repository:${NC}"
        echo "    trivy repo https://github.com/user/repo.git"
        echo ""
        echo -e "${YELLOW}  Generate SBOM:${NC}"
        echo "    trivy image --format cyclonedx nginx:latest > sbom.xml"
        echo ""
        print_status "For more information, visit: https://aquasecurity.github.io/trivy/"
    else
        print_error "Trivy installation or testing failed"
        exit 1
    fi
}

main "$@"
