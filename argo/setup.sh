#!/bin/bash
# ArgoCD Setup Script
# Install and configure ArgoCD on Kubernetes cluster

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
ARGOCD_NAMESPACE="argocd"
ARGOCD_VERSION="stable"

echo -e "${GREEN}=== ArgoCD Installation Script ===${NC}\n"

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}kubectl is not installed. Please install kubectl first.${NC}"
    exit 1
fi

# Check cluster connection
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}Cannot connect to Kubernetes cluster. Check your kubeconfig.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ kubectl is installed and cluster is accessible${NC}\n"

# Create namespace
echo -e "${YELLOW}Creating ArgoCD namespace...${NC}"
kubectl create namespace $ARGOCD_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✓ Namespace created${NC}\n"

# Install ArgoCD
echo -e "${YELLOW}Installing ArgoCD...${NC}"
kubectl apply -n $ARGOCD_NAMESPACE -f https://raw.githubusercontent.com/argoproj/argo-cd/$ARGOCD_VERSION/manifests/install.yaml

# Wait for ArgoCD to be ready
echo -e "${YELLOW}Waiting for ArgoCD pods to be ready...${NC}"
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n $ARGOCD_NAMESPACE --timeout=300s
echo -e "${GREEN}✓ ArgoCD installed successfully${NC}\n"

# Get initial admin password
echo -e "${YELLOW}Retrieving initial admin password...${NC}"
ARGOCD_PASSWORD=$(kubectl -n $ARGOCD_NAMESPACE get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)
echo -e "${GREEN}✓ Admin password retrieved${NC}\n"

# Option to expose ArgoCD server
echo -e "${YELLOW}How would you like to access ArgoCD?${NC}"
echo "1) Port Forward (localhost:8080)"
echo "2) LoadBalancer Service"
echo "3) NodePort Service"
echo "4) Skip (configure manually later)"
read -p "Enter choice [1-4]: " access_choice

case $access_choice in
    1)
        echo -e "${YELLOW}Setting up port forwarding...${NC}"
        echo -e "${GREEN}Run the following command to access ArgoCD:${NC}"
        echo -e "kubectl port-forward svc/argocd-server -n $ARGOCD_NAMESPACE 8080:443"
        echo -e "${GREEN}Access ArgoCD at: https://localhost:8080${NC}"
        ;;
    2)
        echo -e "${YELLOW}Patching service to LoadBalancer...${NC}"
        kubectl patch svc argocd-server -n $ARGOCD_NAMESPACE -p '{"spec": {"type": "LoadBalancer"}}'
        echo -e "${YELLOW}Waiting for external IP...${NC}"
        kubectl get svc argocd-server -n $ARGOCD_NAMESPACE -w
        ;;
    3)
        echo -e "${YELLOW}Patching service to NodePort...${NC}"
        kubectl patch svc argocd-server -n $ARGOCD_NAMESPACE -p '{"spec": {"type": "NodePort"}}'
        NODE_PORT=$(kubectl get svc argocd-server -n $ARGOCD_NAMESPACE -o jsonpath='{.spec.ports[0].nodePort}')
        NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[0].address}')
        echo -e "${GREEN}Access ArgoCD at: https://$NODE_IP:$NODE_PORT${NC}"
        ;;
    4)
        echo -e "${YELLOW}Skipping service configuration${NC}"
        ;;
esac

echo ""

# Install ArgoCD CLI (optional)
read -p "Would you like to install ArgoCD CLI? (y/n): " install_cli

if [[ $install_cli == "y" || $install_cli == "Y" ]]; then
    echo -e "${YELLOW}Installing ArgoCD CLI...${NC}"
    
    # Detect OS
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    if [[ "$ARCH" == "x86_64" ]]; then
        ARCH="amd64"
    elif [[ "$ARCH" == "aarch64" ]]; then
        ARCH="arm64"
    fi
    
    VERSION=$(curl -s https://api.github.com/repos/argoproj/argo-cd/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    
    curl -sSL -o argocd-${OS}-${ARCH} https://github.com/argoproj/argo-cd/releases/download/${VERSION}/argocd-${OS}-${ARCH}
    sudo install -m 555 argocd-${OS}-${ARCH} /usr/local/bin/argocd
    rm argocd-${OS}-${ARCH}
    
    echo -e "${GREEN}✓ ArgoCD CLI installed${NC}"
    argocd version --client
fi

echo ""

# Display credentials
echo -e "${GREEN}=== ArgoCD Installation Complete ===${NC}\n"
echo -e "${YELLOW}Login Credentials:${NC}"
echo -e "Username: ${GREEN}admin${NC}"
echo -e "Password: ${GREEN}$ARGOCD_PASSWORD${NC}"

echo -e "\n${YELLOW}Important: Change the admin password after first login!${NC}"
echo -e "Command: argocd account update-password\n"

# Create sample application (optional)
read -p "Would you like to create a sample guestbook application? (y/n): " create_sample

if [[ $create_sample == "y" || $create_sample == "Y" ]]; then
    echo -e "${YELLOW}Creating sample application...${NC}"
    
    cat <<EOF | kubectl apply -f -
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: guestbook
  namespace: $ARGOCD_NAMESPACE
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
EOF
    
    echo -e "${GREEN}✓ Sample application created${NC}"
    echo -e "Check the application status in ArgoCD UI or run:"
    echo -e "kubectl get applications -n $ARGOCD_NAMESPACE"
fi

echo ""

# Useful commands
echo -e "${YELLOW}=== Useful Commands ===${NC}"
echo -e "1. Get ArgoCD server URL: ${GREEN}kubectl get svc argocd-server -n $ARGOCD_NAMESPACE${NC}"
echo -e "2. List applications: ${GREEN}kubectl get applications -n $ARGOCD_NAMESPACE${NC}"
echo -e "3. Port forward: ${GREEN}kubectl port-forward svc/argocd-server -n $ARGOCD_NAMESPACE 8080:443${NC}"
echo -e "4. View logs: ${GREEN}kubectl logs -n $ARGOCD_NAMESPACE deployment/argocd-server${NC}"
echo -e "5. Delete ArgoCD: ${GREEN}kubectl delete namespace $ARGOCD_NAMESPACE${NC}"

echo -e "\n${GREEN}Setup complete! Happy deploying with ArgoCD! 🚀${NC}"
