#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

INSTALL_SERVING=true
INSTALL_EVENTING=true
DEPLOY_EXAMPLES=false
TIMEOUT_SECONDS=600

print_info() {
    echo -e "${YELLOW}[*]${NC} $1"
}

print_ok() {
    echo -e "${GREEN}[+]${NC} $1"
}

print_err() {
    echo -e "${RED}[!]${NC} $1"
}

usage() {
    cat <<EOF
Knative setup script

Usage:
  ./setup.sh [options]

Options:
  --skip-serving          Skip Knative Serving installation
  --skip-eventing         Skip Knative Eventing installation
  --with-examples         Deploy hands-on examples after install
  --timeout <seconds>     Pod readiness timeout (default: 600)
  -h, --help              Show this help
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --skip-serving)
            INSTALL_SERVING=false
            shift
            ;;
        --skip-eventing)
            INSTALL_EVENTING=false
            shift
            ;;
        --with-examples)
            DEPLOY_EXAMPLES=true
            shift
            ;;
        --timeout)
            TIMEOUT_SECONDS="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            print_err "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

if ! command -v kubectl >/dev/null 2>&1; then
    print_err "kubectl is not installed or not in PATH"
    exit 1
fi

if ! kubectl cluster-info >/dev/null 2>&1; then
    print_err "Cannot reach Kubernetes cluster. Check kubeconfig/context"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

wait_namespace_pods() {
    local ns="$1"
    print_info "Waiting for pods in namespace ${ns} to become Ready"
    kubectl wait --for=condition=Ready pod --all -n "$ns" --timeout="${TIMEOUT_SECONDS}s"
}

install_serving() {
    print_info "Installing Knative Serving CRDs"
    kubectl apply -f https://github.com/knative/serving/releases/latest/download/serving-crds.yaml

    print_info "Installing Knative Serving core"
    kubectl apply -f https://github.com/knative/serving/releases/latest/download/serving-core.yaml

    print_info "Installing Kourier ingress"
    kubectl apply -f https://github.com/knative/net-kourier/releases/latest/download/kourier.yaml

    print_info "Setting Kourier as ingress class"
    kubectl patch configmap/config-network -n knative-serving --type merge --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'

    wait_namespace_pods "knative-serving"
    wait_namespace_pods "kourier-system"

    print_ok "Knative Serving + Kourier installed"
}

install_eventing() {
    print_info "Installing Knative Eventing CRDs"
    kubectl apply -f https://github.com/knative/eventing/releases/latest/download/eventing-crds.yaml

    print_info "Installing Knative Eventing core"
    kubectl apply -f https://github.com/knative/eventing/releases/latest/download/eventing-core.yaml

    print_info "Installing in-memory channel"
    kubectl apply -f https://github.com/knative/eventing/releases/latest/download/in-memory-channel.yaml

    print_info "Installing MT channel broker"
    kubectl apply -f https://github.com/knative/eventing/releases/latest/download/mt-channel-broker.yaml

    wait_namespace_pods "knative-eventing"

    print_ok "Knative Eventing installed"
}

deploy_examples() {
    print_info "Deploying Knative examples"
    kubectl apply -f "$SCRIPT_DIR/examples/service-hello.yaml"
    kubectl apply -f "$SCRIPT_DIR/examples/service-hello-v2.yaml"

    PREV_REV=$(kubectl get revision -l serving.knative.dev/service=hello -o jsonpath='{.items[0].metadata.name}')
    sed "s/PREVIOUS_REVISION_NAME/${PREV_REV}/" "$SCRIPT_DIR/examples/service-hello-traffic-split.yaml" | kubectl apply -f -

    kubectl apply -f "$SCRIPT_DIR/examples/eventing-broker-trigger.yaml"
    print_ok "Examples deployed"
}

print_info "Starting Knative setup"

if [[ "$INSTALL_SERVING" == true ]]; then
    install_serving
else
    print_info "Skipping Serving install"
fi

if [[ "$INSTALL_EVENTING" == true ]]; then
    install_eventing
else
    print_info "Skipping Eventing install"
fi

if [[ "$DEPLOY_EXAMPLES" == true ]]; then
    deploy_examples
fi

print_ok "Setup complete"
print_info "Run: kubectl get pods -n knative-serving"
print_info "Run: kubectl get pods -n knative-eventing"
