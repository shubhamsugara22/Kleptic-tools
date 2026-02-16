#!/bin/bash
# Envoy Setup Script
# This script sets up Envoy proxy on the system

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Envoy Proxy Setup ===${NC}\n"

# Check if running on Linux, macOS, or Windows (Git Bash)
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macos"
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
    OS="windows"
else
    OS="unknown"
fi

echo -e "${YELLOW}Detected OS: $OS${NC}\n"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check for Docker
if command_exists docker; then
    echo -e "${GREEN}✓ Docker is installed${NC}"
    DOCKER_AVAILABLE=true
else
    echo -e "${YELLOW}✗ Docker is not installed${NC}"
    DOCKER_AVAILABLE=false
fi

echo ""

# Install Envoy based on OS
case $OS in
    linux)
        echo -e "${BLUE}Installing Envoy on Linux...${NC}\n"
        
        # For Ubuntu/Debian
        if command_exists apt-get; then
            echo "Installing Envoy via apt..."
            sudo apt-get update
            sudo apt-get install -y envoy
            echo -e "${GREEN}✓ Envoy installed via apt${NC}"
        # For CentOS/RHEL
        elif command_exists yum; then
            echo "Installing Envoy via yum..."
            sudo yum install -y envoy
            echo -e "${GREEN}✓ Envoy installed via yum${NC}"
        else
            echo -e "${YELLOW}Could not find package manager. Please install Envoy manually.${NC}"
            echo "Visit: https://www.envoyproxy.io/docs/envoy/latest/install/install"
        fi
        ;;
    
    macos)
        echo -e "${BLUE}Installing Envoy on macOS...${NC}\n"
        
        if command_exists brew; then
            echo "Installing Envoy via Homebrew..."
            brew tap envoyproxy/envoy
            brew install envoy
            echo -e "${GREEN}✓ Envoy installed via Homebrew${NC}"
        else
            echo -e "${YELLOW}Homebrew not found. Please install Homebrew first.${NC}"
            echo "Visit: https://brew.sh"
        fi
        ;;
    
    windows)
        echo -e "${BLUE}Installing Envoy on Windows...${NC}\n"
        
        if command_exists choco; then
            echo "Installing Envoy via Chocolatey..."
            choco install envoy -y
            echo -e "${GREEN}✓ Envoy installed via Chocolatey${NC}"
        elif [ "$DOCKER_AVAILABLE" = true ]; then
            echo "Docker available! You can run Envoy in a container:"
            echo "docker run -v \$(pwd)/envoy.yaml:/etc/envoy/envoy.yaml -p 10000:10000 -p 9901:9901 envoyproxy/envoy:v1.27-latest"
        else
            echo -e "${YELLOW}Could not auto-install Envoy on Windows.${NC}"
            echo "Please use one of these methods:"
            echo "1. Install Chocolatey and run: choco install envoy"
            echo "2. Use Docker: docker pull envoyproxy/envoy"
            echo "3. Download from: https://www.envoyproxy.io/docs/envoy/latest/install/install"
        fi
        ;;
    
    *)
        echo -e "${YELLOW}Unknown OS: $OS${NC}"
        echo "Please visit: https://www.envoyproxy.io/docs/envoy/latest/install/install"
        ;;
esac

echo ""

# Check if Envoy is installed
if command_exists envoy; then
    echo -e "${GREEN}✓ Envoy is installed${NC}"
    ENVOY_VERSION=$(envoy --version 2>/dev/null || echo "Version unknown")
    echo -e "  $ENVOY_VERSION\n"
else
    echo -e "${YELLOW}⚠ Envoy command not found in PATH${NC}"
    echo "  Ensure Envoy is properly installed and added to your PATH\n"
fi

# Create sample configuration if it doesn't exist
if [ ! -f "envoy.yaml" ]; then
    echo -e "${BLUE}Creating sample envoy.yaml configuration...${NC}\n"
    
    cat > envoy.yaml <<'EOF'
static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address:
        address: 0.0.0.0
        port_value: 10000
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          http_filters:
          - name: envoy.filters.http.router
          route_config:
            name: local_route
            virtual_hosts:
            - name: backend
              domains: ["*"]
              routes:
              - match:
                  prefix: "/"
                route:
                  cluster: backend_service
  clusters:
  - name: backend_service
    type: STATIC
    load_assignment:
      cluster_name: backend_service
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 8080
admin:
  access_log_path: /tmp/admin_access.log
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 9901
EOF
    
    echo -e "${GREEN}✓ Sample configuration created: envoy.yaml${NC}\n"
else
    echo -e "${GREEN}✓ envoy.yaml already exists${NC}\n"
fi

# Setup complete
echo -e "${GREEN}=== Setup Complete ===${NC}\n"
echo "Next steps:"
echo "1. Review and modify envoy.yaml as needed"
echo "2. Start Envoy:"
echo "   envoy -c envoy.yaml"
echo "3. Access admin interface:"
echo "   http://localhost:9901"
echo ""
echo "For more information, visit: https://www.envoyproxy.io/"
