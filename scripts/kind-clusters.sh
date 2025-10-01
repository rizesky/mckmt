#!/bin/bash

# Kind cluster management script for MCKMT
# This script manages multiple Kind clusters for MCKMT demo purposes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
MEMORY=${MEMORY:-2048}
CPUS=${CPUS:-2}

# Get cluster count from first argument if it's a number, otherwise use default
if [[ "$1" =~ ^[0-9]+$ ]]; then
    CLUSTER_COUNT=$1
else
    CLUSTER_COUNT=3
fi

# Cluster names
CLUSTER_NAMES=()
for i in $(seq 1 $CLUSTER_COUNT); do
    CLUSTER_NAMES+=("mckmt-cluster-$i")
done

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if kind is installed
check_kind() {
    if ! command -v kind &> /dev/null; then
        print_error "Kind is not installed. Please install Kind first."
        print_status "Installation: https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
        exit 1
    fi
    print_success "Kind is installed: $(kind version)"
}

# Create Kind clusters
create_clusters() {
    print_status "Creating $CLUSTER_COUNT Kind clusters..."
    
    for i in $(seq 1 $CLUSTER_COUNT); do
        CLUSTER_NAME="mckmt-cluster-$i"
        
        # Check if cluster already exists
        if kind get clusters | grep -q "^$CLUSTER_NAME$"; then
            print_warning "Cluster $CLUSTER_NAME already exists, skipping..."
            continue
        fi
        
        print_status "Creating cluster: $CLUSTER_NAME"
        
        # Create cluster with port mappings for each cluster
        cat <<EOF | kind create cluster --name $CLUSTER_NAME --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: $CLUSTER_NAME
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: $((80 + i - 1))
    protocol: TCP
  - containerPort: 443
    hostPort: $((443 + i - 1))
    protocol: TCP
  - containerPort: 30000
    hostPort: $((30000 + i - 1))
    protocol: TCP
  - containerPort: 30001
    hostPort: $((30001 + i - 1))
    protocol: TCP
  - containerPort: 30002
    hostPort: $((30002 + i - 1))
    protocol: TCP
EOF
        
        print_success "Created cluster: $CLUSTER_NAME"
    done
}

# Load Docker images into Kind clusters
load_images() {
    print_status "Loading Docker images into Kind clusters..."
    
    # Build agent image if it doesn't exist
    if ! docker image inspect mckmt-agent:latest &> /dev/null; then
        print_status "Building MCKMT agent image..."
        docker build -f deployments/docker/Dockerfile.agent -t mckmt-agent:latest .
    fi
    
    # Load images into each cluster
    for CLUSTER_NAME in "${CLUSTER_NAMES[@]}"; do
        if kind get clusters | grep -q "^$CLUSTER_NAME$"; then
            print_status "Loading images into $CLUSTER_NAME..."
            kind load docker-image mckmt-agent:latest --name $CLUSTER_NAME
        fi
    done
}

# Deploy agents to clusters
deploy_agents() {
    print_status "Deploying MCKMT agents to clusters..."
    
    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed. Please install kubectl first."
        print_status "Installation: https://kubernetes.io/docs/tasks/tools/"
        return 1
    fi
    
    for i in $(seq 1 $CLUSTER_COUNT); do
        CLUSTER_NAME="mckmt-cluster-$i"
        
        if ! kind get clusters | grep -q "^$CLUSTER_NAME$"; then
            print_warning "Cluster $CLUSTER_NAME not found, skipping agent deployment..."
            continue
        fi
        
        print_status "Deploying agent to $CLUSTER_NAME..."
        
        # Set kubectl context
        kubectl config use-context kind-$CLUSTER_NAME
        
        # Create a temporary kustomization for this cluster
        TEMP_DIR=$(mktemp -d)
        cp -r deployments/k8s/overlays/demo/* "$TEMP_DIR/"
        
        # Update the cluster name in the config
        sed -i "s/mckmt-demo-cluster/$CLUSTER_NAME/g" "$TEMP_DIR/agent-demo-patch.yaml"
        sed -i "s/mckmt-demo/$CLUSTER_NAME/g" "$TEMP_DIR/kustomization.yaml"
        
        # Deploy using kubectl kustomize
        kubectl kustomize "$TEMP_DIR" | kubectl apply -f -
        
        # Clean up temp directory
        rm -rf "$TEMP_DIR"
        
        print_success "Deployed agent to $CLUSTER_NAME"
    done
}

# Stop clusters
stop_clusters() {
    print_status "Stopping Kind clusters..."
    
    for CLUSTER_NAME in "${CLUSTER_NAMES[@]}"; do
        if kind get clusters | grep -q "^$CLUSTER_NAME$"; then
            print_status "Stopping cluster: $CLUSTER_NAME"
            kind delete cluster --name $CLUSTER_NAME
            print_success "Stopped cluster: $CLUSTER_NAME"
        else
            print_warning "Cluster $CLUSTER_NAME not found"
        fi
    done
}

# List clusters
list_clusters() {
    print_status "Listing Kind clusters..."
    kind get clusters
}

# Show cluster status
show_status() {
    print_status "Kind cluster status:"
    for CLUSTER_NAME in "${CLUSTER_NAMES[@]}"; do
        if kind get clusters | grep -q "^$CLUSTER_NAME$"; then
            print_success "✅ $CLUSTER_NAME is running"
            kubectl config use-context kind-$CLUSTER_NAME
            kubectl get pods -n mckmt-agent 2>/dev/null || print_warning "No agent pods found in $CLUSTER_NAME"
        else
            print_error "❌ $CLUSTER_NAME is not running"
        fi
    done
}

# Main function
main() {
    case "${1:-help}" in
        "create")
            check_kind
            create_clusters
            load_images
            deploy_agents
            print_success "Created $CLUSTER_COUNT Kind clusters with MCKMT agents"
            print_status "Clusters: ${CLUSTER_NAMES[*]}"
            print_status "To check status: $0 status"
            ;;
        "stop")
            stop_clusters
            print_success "Stopped all Kind clusters"
            ;;
        "status")
            show_status
            ;;
        "list")
            list_clusters
            ;;
        "help"|*)
            echo "Usage: $0 {create|stop|status|list} [count]"
            echo ""
            echo "Commands:"
            echo "  create [count]  Create Kind clusters with MCKMT agents (default: 3)"
            echo "  stop           Stop all Kind clusters"
            echo "  status         Show cluster and agent status"
            echo "  list           List all Kind clusters"
            echo "  help           Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  MEMORY         Memory per cluster (default: 2048)"
            echo "  CPUS           CPUs per cluster (default: 2)"
            echo ""
            echo "Examples:"
            echo "  $0 create 5    # Create 5 clusters"
            echo "  $0 status      # Check status"
            echo "  $0 stop        # Stop all clusters"
            ;;
    esac
}

main "$@"
