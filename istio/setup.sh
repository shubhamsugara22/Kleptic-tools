#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Istio Service Mesh Setup${NC}"
echo "================================"

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}kubectl is not installed. Please install kubectl first.${NC}"
    exit 1
fi

# Check if cluster is accessible
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}Cannot access Kubernetes cluster. Please check your kubeconfig.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ kubectl found${NC}"
echo -e "${GREEN}✓ Kubernetes cluster accessible${NC}"

# Download and install Istio
echo ""
echo -e "${YELLOW}Downloading Istio...${NC}"

# Get the latest Istio version
ISTIO_VERSION=$(curl -s https://api.github.com/repos/istio/istio/releases/latest | grep tag_name | sed -e 's/.*"\(.*\)".*/\1/')
DOWNLOAD_URL="https://github.com/istio/istio/releases/download/${ISTIO_VERSION}/istio-${ISTIO_VERSION}-linux-amd64.tar.gz"

echo "Latest Istio version: $ISTIO_VERSION"
echo "Download URL: $DOWNLOAD_URL"

# Download Istio
curl -L "$DOWNLOAD_URL" | tar xz

# Add istioctl to PATH
ISTIO_DIR=$(ls -d istio-* 2>/dev/null | head -n 1)
export PATH=$PATH:$PWD/$ISTIO_DIR/bin

echo -e "${GREEN}✓ Istio downloaded${NC}"

# Create istio-system namespace
echo ""
echo -e "${YELLOW}Creating istio-system namespace...${NC}"
kubectl create namespace istio-system --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✓ Namespace created${NC}"

# Install Istio using istioctl
echo ""
echo -e "${YELLOW}Installing Istio control plane...${NC}"
istioctl install --set profile=demo -y
echo -e "${GREEN}✓ Istio control plane installed${NC}"

# Enable sidecar injection for default namespace (optional)
echo ""
echo -e "${YELLOW}Enabling sidecar injection for default namespace...${NC}"
kubectl label namespace default istio-injection=enabled --overwrite
echo -e "${GREEN}✓ Sidecar injection enabled${NC}"

# Wait for Istio components to be ready
echo ""
echo -e "${YELLOW}Waiting for Istio components to be ready...${NC}"
kubectl wait --for=condition=ready pod -l app=istiod -n istio-system --timeout=300s 2>/dev/null || true

# Check installation status
echo ""
echo -e "${GREEN}Verifying Istio installation...${NC}"
kubectl get pods -n istio-system

# Display access instructions
echo ""
echo -e "${GREEN}✓ Istio installation completed!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Access Kiali dashboard (UI):"
echo "   kubectl port-forward svc/kiali -n istio-system 20000:20000"
echo "   Open: http://localhost:20000"
echo ""
echo "2. Access Grafana (Metrics):"
echo "   kubectl port-forward svc/grafana -n istio-system 3000:3000"
echo "   Open: http://localhost:3000"
echo ""
echo "3. Deploy sample application:"
echo "   kubectl apply -f $ISTIO_DIR/samples/bookinfo/platform/kube/bookinfo.yaml"
echo ""
echo "4. Check sidecar injection status:"
echo "   kubectl get namespace default --show-labels"
echo ""
echo "5. View Istio configuration:"
echo "   istioctl analyze"
echo ""
echo -e "${YELLOW}To uninstall Istio:${NC}"
echo "   istioctl uninstall --purge"
