#!/bin/bash

# Test Kustomize configurations for MCKMT

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if kubectl is installed
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed. Please install kubectl first."
        print_status "Installation: https://kubernetes.io/docs/tasks/tools/"
        exit 1
    fi
    print_success "kubectl is installed: $(kubectl version --client --short 2>/dev/null || echo 'kubectl available')"
}

# Test demo overlay
test_demo() {
    print_status "Testing demo overlay..."
    kubectl kustomize deployments/k8s/overlays/demo
    print_success "Demo overlay is valid"
}

# Test production overlay
test_production() {
    print_status "Testing production overlay..."
    kubectl kustomize deployments/k8s/overlays/production
    print_success "Production overlay is valid"
}

# Test base
test_base() {
    print_status "Testing base configuration..."
    kubectl kustomize deployments/k8s/base
    print_success "Base configuration is valid"
}

# Main function
main() {
    case "${1:-all}" in
        "demo")
            check_kubectl
            test_demo
            ;;
        "production")
            check_kubectl
            test_production
            ;;
        "base")
            check_kubectl
            test_base
            ;;
        "all")
            check_kubectl
            test_base
            test_demo
            test_production
            print_success "All Kustomize configurations are valid!"
            ;;
        "help"|*)
            echo "Usage: $0 {demo|production|base|all}"
            echo ""
            echo "Commands:"
            echo "  demo        Test demo overlay"
            echo "  production  Test production overlay"
            echo "  base        Test base configuration"
            echo "  all         Test all configurations (default)"
            echo "  help        Show this help message"
            ;;
    esac
}

main "$@"
