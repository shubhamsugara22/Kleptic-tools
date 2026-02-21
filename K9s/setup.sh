#!/usr/bin/env bash
set -euo pipefail

# K9s setup script for local use

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

echo -e "${BLUE}=== K9s Setup ===${NC}\n"

# Detect OS
OS="unknown"
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
  OS="macos"
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
  OS="windows"
fi

echo -e "${YELLOW}Detected OS: $OS${NC}\n"

# Install K9s
case "$OS" in
  macos)
    if command_exists brew; then
      brew install k9s
    else
      echo -e "${YELLOW}Homebrew not found. Install from https://brew.sh${NC}"
      exit 1
    fi
    ;;
  linux)
    if command_exists apt-get; then
      sudo apt-get update
      sudo apt-get install -y k9s
    elif command_exists yum; then
      sudo yum install -y k9s
    else
      echo -e "${YELLOW}No supported package manager found.${NC}"
      echo "Download the binary from: https://github.com/derailed/k9s/releases"
      exit 1
    fi
    ;;
  windows)
    if command_exists choco; then
      choco install k9s -y
    else
      echo -e "${YELLOW}Chocolatey not found. Install from https://chocolatey.org${NC}"
      exit 1
    fi
    ;;
  *)
    echo -e "${YELLOW}Unsupported OS. Install manually from:${NC}"
    echo "https://github.com/derailed/k9s/releases"
    exit 1
    ;;
esac

echo -e "${GREEN}K9s installation complete.${NC}\n"

# Verify kubectl
if ! command_exists kubectl; then
  echo -e "${YELLOW}kubectl not found.${NC}"
  echo "Install kubectl and ensure your kubeconfig is set up."
  exit 1
fi

# Verify k9s
if command_exists k9s; then
  echo -e "${GREEN}K9s is ready:${NC}"
  k9s version || true
else
  echo -e "${YELLOW}K9s is not in PATH after install.${NC}"
  exit 1
fi

echo -e "\n${BLUE}Quick monitoring tips:${NC}"
cat <<'EOF'
- Launch: k9s
- Change namespace: n
- Change context: 0
- Command mode: :
- Common views: :po (pods), :svc (services), :deploy (deployments), :ns (namespaces)
- Filter: /
- Logs: l (while on a pod)
- Describe: d (while on a resource)
EOF

