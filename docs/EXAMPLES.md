# MCKMT Examples

This document provides comprehensive examples for using MCKMT with various Kubernetes environments, including Kind, minikube, and cloud providers.

## Table of Contents

- [Quick Start with Kind](#quick-start-with-kind)
- [Using with Minikube](#using-with-minikube)
- [Multi-Cluster Setup](#multi-cluster-setup)
- [Agent Deployment Examples](#agent-deployment-examples)
- [API Usage Examples](#api-usage-examples)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) installed (recommended)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) installed
- Docker installed
- Go 1.21+ installed

## Quick Start with Kind

### Step 1: Create Kind Cluster

```bash
# Use the provided Kind configuration
kind create cluster --config=configs/kind-single.yaml --name=mckmt-demo

# Verify cluster
kubectl cluster-info --context kind-mckmt-demo
```

### Step 2: Start MCKMT Hub

**Option A: Development Mode (Local)**
```bash
# Clone the repository
git clone https://github.com/rizesky/mckmt.git
cd mckmt

# Install dependencies
go mod download

# Start dependencies (PostgreSQL, Redis, etc.)
make deps

# Start the hub API server locally
go run cmd/hub/main.go
```

**Option B: Demo Mode (Docker)**
```bash
# Start complete demo environment
make demo-password
# or for OIDC demo
make demo-oidc
```

The hub will be available at:
- **HTTP API**: http://localhost:8080
- **gRPC API**: localhost:8081
- **API Documentation**: http://localhost:8080/swagger/index.html

### Step 3: Deploy the Agent

**Note**: Cluster registration is now handled automatically by the agent via gRPC. No manual registration needed!

### How Automatic Registration Works

1. **Agent starts** and connects to the hub via gRPC
2. **Agent registers** itself with cluster information (name, labels, etc.)
3. **Hub assigns** a unique cluster ID
4. **Agent begins** heartbeat and operation processing

Deploy the agent using Kustomize:

```bash
# Build agent image
docker build -t mckmt-agent:latest .

# Load image into kind
kind load docker-image mckmt-agent:latest --name mckmt-demo

# Deploy agent using Kustomize
kubectl kustomize deployments/k8s/overlays/demo | kubectl apply -f -
```

### Step 4: Verify Setup

```bash
# Check agent is running
kubectl get pods -l app=mckmt-agent

# Check agent logs
kubectl logs -l app=mckmt-agent

# Check cluster status via API
curl http://localhost:8080/api/v1/clusters/mckmt-demo/status
```

## Using with Minikube

### Step 1: Start Minikube

```bash
# Start minikube with sufficient resources
minikube start --memory=4096 --cpus=2

# Verify cluster is running
kubectl cluster-info
```

### Step 2: Start MCKMT Hub

**Option A: Development Mode (Local)**
```bash
# Clone the repository
git clone https://github.com/rizesky/mckmt.git
cd mckmt

# Install dependencies
go mod download

# Start dependencies (PostgreSQL, Redis, etc.)
make deps

# Start the hub API server locally
go run cmd/hub/main.go
```

**Option B: Demo Mode (Docker)**
```bash
# Start complete demo environment
make demo-password
```

### Step 3: Deploy Agent to Minikube

```bash
# Build agent image
docker build -t mckmt-agent:latest .

# Load image into minikube
minikube image load mckmt-agent:latest

# Deploy agent
kubectl apply -f deployments/k8s/deployment-agent.yaml
```

### Step 4: Verify Setup

```bash
# Check agent is running
kubectl get pods -l app=mckmt-agent

# Check agent logs
kubectl logs -l app=mckmt-agent

# Check cluster status via API
curl http://localhost:8080/api/v1/clusters/minikube-dev/status
```

## Multi-Cluster Setup

### Scenario: Managing Multiple Kind Clusters

```bash
# Create multiple Kind clusters
kind create cluster --name cluster1 --config=configs/kind-single.yaml
kind create cluster --name cluster2 --config=configs/kind-single.yaml
kind create cluster --name cluster3 --config=configs/kind-single.yaml

# Deploy agents to each cluster
kubectl config use-context kind-cluster1
kubectl kustomize deployments/k8s/overlays/demo | kubectl apply -f -

kubectl config use-context kind-cluster2
kubectl kustomize deployments/k8s/overlays/demo | kubectl apply -f -

kubectl config use-context kind-cluster3
kubectl kustomize deployments/k8s/overlays/demo | kubectl apply -f -
```

### Scenario: Managing Multiple Minikube Clusters

```bash
# Start multiple minikube profiles
minikube start --profile=cluster1 --memory=2048 --cpus=2
minikube start --profile=cluster2 --memory=2048 --cpus=2

# Deploy agents to both clusters
# Note: Cluster registration is now handled automatically by the agent via gRPC
kubectl config use-context minikube-cluster1
kubectl apply -f deployments/k8s/deployment-agent.yaml

kubectl config use-context minikube-cluster2
kubectl apply -f deployments/k8s/deployment-agent.yaml
```

## Agent Deployment Examples

### Using Kustomize for Environment-Specific Deployments

#### Demo Environment

```bash
# Deploy to demo environment
kubectl kustomize deployments/k8s/overlays/demo | kubectl apply -f -
```

#### Production Environment

```bash
# Deploy to production environment
kubectl kustomize deployments/k8s/overlays/production | kubectl apply -f -
```

### Custom Configuration

You can customize the agent deployment by modifying the Kustomize overlays:

```bash
# Edit demo configuration
vim deployments/k8s/overlays/demo/agent-demo-patch.yaml

# Apply changes
kubectl kustomize deployments/k8s/overlays/demo | kubectl apply -f -
```

## API Usage Examples

### Cluster Management

#### List All Clusters

```bash
curl -X GET http://localhost:8080/api/v1/clusters
```

#### Get Cluster Details

```bash
curl -X GET http://localhost:8080/api/v1/clusters/{cluster-id}
```

#### Get Cluster Status

```bash
curl -X GET http://localhost:8080/api/v1/clusters/{cluster-id}/status
```

### Operation Management

#### List Operations

```bash
curl -X GET http://localhost:8080/api/v1/operations
```

#### Get Operation Details

```bash
curl -X GET http://localhost:8080/api/v1/operations/{operation-id}
```

#### Cancel Operation

```bash
curl -X POST http://localhost:8080/api/v1/operations/{operation-id}/cancel
```

### Authentication

#### Register User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "password123"
  }'
```

#### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password123"
  }'
```

## Troubleshooting

### Common Issues

#### Agent Not Connecting to Hub

1. **Check network connectivity**:
   ```bash
   # From agent pod
   kubectl exec -it <agent-pod> -- curl http://hub-service:8080/health
   ```

2. **Verify gRPC connection**:
   ```bash
   # Check agent logs
   kubectl logs -l app=mckmt-agent
   ```

3. **Check service discovery**:
   ```bash
   # Verify DNS resolution
   kubectl exec -it <agent-pod> -- nslookup hub-service
   ```

#### Cluster Not Appearing in Hub

1. **Check agent registration**:
   ```bash
   # Check agent logs for registration messages
   kubectl logs -l app=mckmt-agent | grep -i register
   ```

2. **Verify cluster ID assignment**:
   ```bash
   # Check agent environment variables
   kubectl exec -it <agent-pod> -- env | grep MCKMT
   ```

#### Kind Cluster Issues

```bash
# Reset Kind cluster if needed
kind delete cluster --name mckmt-demo
kind create cluster --config=configs/kind-single.yaml --name=mckmt-demo

# Check Kind cluster status
kind get clusters
```

#### Minikube Cluster Issues

```bash
# Reset minikube if needed
minikube delete
minikube start --memory=4096 --cpus=2

# Check minikube status
minikube status

# Access minikube dashboard
minikube dashboard
```

### Debug Commands

#### Check All Resources

```bash
# List all MCKMT resources
kubectl get all -l app.kubernetes.io/part-of=mckmt

# Check namespaces
kubectl get namespaces | grep mckmt
```

#### View Logs

```bash
# Agent logs
kubectl logs -l app=mckmt-agent

# Hub logs (if running in container)
docker logs mckmt-hub
```

#### Test Connectivity

```bash
# Test HTTP API
curl http://localhost:8080/health

# Test gRPC (if grpcurl is installed)
grpcurl -plaintext localhost:8081 list
```

### Performance Tuning

#### Resource Limits

Adjust resource limits in the Kustomize overlays:

```yaml
# deployments/k8s/overlays/production/agent-production-patch.yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "200m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

#### Scaling

```bash
# Scale agent deployment
kubectl scale deployment mckmt-agent --replicas=3

# Check pod distribution
kubectl get pods -o wide
```

## Advanced Examples

### Custom Agent Configuration

Create a custom overlay for your specific needs:

```bash
# Create custom overlay
mkdir -p deployments/k8s/overlays/custom

# Copy from demo overlay
cp -r deployments/k8s/overlays/demo/* deployments/k8s/overlays/custom/

# Modify configuration
vim deployments/k8s/overlays/custom/agent-demo-patch.yaml

# Deploy custom configuration
kubectl kustomize deployments/k8s/overlays/custom | kubectl apply -f -
```

### Integration with CI/CD

```yaml
# .github/workflows/deploy-agent.yml
name: Deploy MCKMT Agent
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy to Kind
        run: |
          kind create cluster --config=configs/kind-single.yaml
          kubectl kustomize deployments/k8s/overlays/demo | kubectl apply -f -
```

This completes the updated examples documentation with Kind prioritized over minikube! 🚀