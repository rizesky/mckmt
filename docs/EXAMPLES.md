# MCKMT Examples

This document provides comprehensive examples for using MCKMT with various Kubernetes environments, including minikube, kind, and cloud providers.

## Table of Contents

- [Quick Start with Minikube](#quick-start-with-minikube)
- [Using with Kind](#using-with-kind)
- [Multi-Cluster Setup](#multi-cluster-setup)
- [Agent Deployment Examples](#agent-deployment-examples)
- [API Usage Examples](#api-usage-examples)
- [Troubleshooting](#troubleshooting)

## Quick Start with Minikube

### Prerequisites

- [minikube](https://minikube.sigs.k8s.io/docs/start/) installed
- [kubectl](https://kubernetes.io/docs/tasks/tools/) installed
- Docker installed
- Go 1.21+ installed

### Step 1: Start Minikube

```bash
# Start minikube with sufficient resources
minikube start --memory=4096 --cpus=2

# Verify cluster is running
kubectl cluster-info
```

### Step 2: Start MCKMT Hub

```bash
# Clone the repository
git clone https://github.com/rizesky/mckmt.git
cd mckmt

# Start dependencies (PostgreSQL, Redis, etc.)
make deps

# Start the hub API server
go run cmd/hub/main.go
```

The hub will be available at:
- **HTTP API**: http://localhost:8080
- **gRPC API**: localhost:8081
- **API Documentation**: http://localhost:8080/swagger/index.html

### Step 3: Register the Minikube Cluster

```bash
# Register the minikube cluster with MCKMT
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "minikube-dev",
    "mode": "agent",
    "labels": {
      "env": "development",
      "provider": "minikube"
    }
  }'
```

### Step 4: Deploy the Agent

Create a Kubernetes manifest for the agent:

```yaml
# agent-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mckmt-agent
  namespace: mckmt-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mckmt-agent
  template:
    metadata:
      labels:
        app: mckmt-agent
    spec:
      serviceAccountName: mckmt-agent
      containers:
      - name: agent
        image: mckmt-agent:latest
        env:
        - name: MCKMT_HUB_URL
          value: "host.docker.internal:8081"
        - name: MCKMT_CLUSTER_ID
          value: "minikube-dev"
        - name: MCKMT_AGENT_ID
          value: "agent-001"
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mckmt-agent
  namespace: mckmt-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: mckmt-agent
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets", "statefulsets", "daemonsets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["batch"]
  resources: ["jobs", "cronjobs"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: [""]
  resources: ["pods/exec"]
  verbs: ["create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: mckmt-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: mckmt-agent
subjects:
- kind: ServiceAccount
  name: mckmt-agent
  namespace: mckmt-system
---
apiVersion: v1
kind: Namespace
metadata:
  name: mckmt-system
```

Deploy the agent:

```bash
# Create namespace and deploy agent
kubectl apply -f agent-deployment.yaml

# Check agent status
kubectl get pods -n mckmt-system
kubectl logs -n mckmt-system deployment/mckmt-agent
```

### Step 5: Verify Connection

```bash
# Check cluster status
curl http://localhost:8080/api/v1/clusters

# Check agent heartbeat
curl http://localhost:8080/api/v1/clusters/minikube-dev/status
```

## Using with Kind

### Step 1: Create Kind Cluster

```bash
# Create kind cluster configuration
cat > kind-config.yaml << EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
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
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF

# Create the cluster
kind create cluster --config=kind-config.yaml --name=mckmt-demo

# Verify cluster
kubectl cluster-info --context kind-mckmt-demo
```

### Step 2: Deploy MCKMT Hub

```bash
# Start MCKMT hub
make deps
go run cmd/hub/main.go
```

### Step 3: Register Kind Cluster

```bash
# Register the kind cluster
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "kind-demo",
    "mode": "agent",
    "labels": {
      "env": "demo",
      "provider": "kind"
    }
  }'
```

### Step 4: Deploy Agent to Kind

```bash
# Build agent image
docker build -t mckmt-agent:latest .

# Load image into kind
kind load docker-image mckmt-agent:latest --name mckmt-demo

# Deploy agent (use the same manifest as minikube)
kubectl apply -f agent-deployment.yaml --context kind-mckmt-demo
```

## Multi-Cluster Setup

### Scenario: Managing Multiple Minikube Clusters

```bash
# Start multiple minikube profiles
minikube start --profile=cluster1 --memory=2048 --cpus=2
minikube start --profile=cluster2 --memory=2048 --cpus=2

# Register both clusters
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cluster1-prod",
    "mode": "agent",
    "labels": {
      "env": "production",
      "region": "us-west-1"
    }
  }'

curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cluster2-staging",
    "mode": "agent",
    "labels": {
      "env": "staging",
      "region": "us-west-1"
    }
  }'

# Deploy agents to both clusters
kubectl config use-context minikube-cluster1
kubectl apply -f agent-deployment.yaml

kubectl config use-context minikube-cluster2
kubectl apply -f agent-deployment.yaml
```

## Agent Deployment Examples

### Docker Compose for Local Development

```yaml
# docker-compose.agent.yml
version: '3.8'
services:
  mckmt-agent:
    build: .
    command: ["./agent"]
    environment:
      - MCKMT_HUB_URL=host.docker.internal:8081
      - MCKMT_CLUSTER_ID=local-dev
      - MCKMT_AGENT_ID=agent-local
    volumes:
      - ~/.kube/config:/root/.kube/config:ro
    depends_on:
      - hub
    networks:
      - mckmt-network

  hub:
    build: .
    command: ["./hub"]
    ports:
      - "8080:8080"
      - "8081:8081"
    environment:
      - MCKMT_DATABASE_HOST=postgres
      - MCKMT_REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    networks:
      - mckmt-network

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=mckmt
      - POSTGRES_USER=mckmt
      - POSTGRES_PASSWORD=mckmt
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - mckmt-network

  redis:
    image: redis:7-alpine
    networks:
      - mckmt-network

volumes:
  postgres_data:

networks:
  mckmt-network:
    driver: bridge
```

### Kubernetes Deployment with ConfigMap

```yaml
# agent-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mckmt-agent-config
  namespace: mckmt-system
data:
  config.yaml: |
    agent:
      hub_url: "hub.mckmt.svc.cluster.local:8081"
      cluster_id: "production-cluster"
      agent_id: "agent-001"
      heartbeat_interval: "30s"
    logging:
      level: "info"
      format: "json"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mckmt-agent
  namespace: mckmt-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mckmt-agent
  template:
    metadata:
      labels:
        app: mckmt-agent
    spec:
      serviceAccountName: mckmt-agent
      containers:
      - name: agent
        image: mckmt-agent:latest
        command: ["./agent"]
        volumeMounts:
        - name: config
          mountPath: /etc/mckmt
        - name: kubeconfig
          mountPath: /root/.kube
          readOnly: true
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
      volumes:
      - name: config
        configMap:
          name: mckmt-agent-config
      - name: kubeconfig
        secret:
          secretName: mckmt-agent-kubeconfig
```

## API Usage Examples

### Cluster Management

```bash
# List all clusters
curl http://localhost:8080/api/v1/clusters

# Get specific cluster details
curl http://localhost:8080/api/v1/clusters/{cluster-id}

# Update cluster labels
curl -X PATCH http://localhost:8080/api/v1/clusters/{cluster-id} \
  -H "Content-Type: application/json" \
  -d '{
    "labels": {
      "env": "production",
      "version": "1.2.3"
    }
  }'

# Delete cluster
curl -X DELETE http://localhost:8080/api/v1/clusters/{cluster-id}
```

### Deploying Applications

```bash
# Deploy a simple nginx deployment
curl -X POST http://localhost:8080/api/v1/clusters/{cluster-id}/manifests \
  -H "Content-Type: application/yaml" \
  --data-binary @nginx-deployment.yaml

# Check operation status
curl http://localhost:8080/api/v1/operations/{operation-id}

# Get operation logs
curl http://localhost:8080/api/v1/operations/{operation-id}/logs
```

### Resource Management

```bash
# List cluster resources
curl "http://localhost:8080/api/v1/clusters/{cluster-id}/resources?kind=Deployment"

# Sync cluster state
curl -X POST http://localhost:8080/api/v1/clusters/{cluster-id}/sync

# Get cluster metrics
curl http://localhost:8080/api/v1/clusters/{cluster-id}/metrics
```

### Executing Commands

```bash
# Execute kubectl command
curl -X POST http://localhost:8080/api/v1/clusters/{cluster-id}/kubectl/exec \
  -H "Content-Type: application/json" \
  -d '{
    "command": ["kubectl", "get", "pods", "-A"],
    "timeout": "30s"
  }'
```

## Troubleshooting

### Common Issues

#### Agent Connection Issues

```bash
# Check agent logs
kubectl logs -n mckmt-system deployment/mckmt-agent

# Verify network connectivity
kubectl exec -n mckmt-system deployment/mckmt-agent -- nslookup hub.mckmt.svc.cluster.local

# Check gRPC connectivity
kubectl exec -n mckmt-system deployment/mckmt-agent -- telnet hub.mckmt.svc.cluster.local 8081
```

#### Hub API Issues

```bash
# Check hub logs
docker logs mckmt-hub

# Verify database connection
docker exec mckmt-postgres psql -U mckmt -d mckmt -c "SELECT COUNT(*) FROM clusters;"

# Check Redis connection
docker exec mckmt-redis redis-cli ping
```

#### Minikube Specific Issues

```bash
# Reset minikube if needed
minikube delete
minikube start --memory=4096 --cpus=2

# Check minikube status
minikube status

# Access minikube dashboard
minikube dashboard
```

### Debugging Commands

```bash
# Enable debug logging
export MCKMT_LOGGING_LEVEL=debug
go run cmd/hub/main.go

# Check gRPC server status
grpcurl -plaintext localhost:8081 list

# Test gRPC connection
grpcurl -plaintext localhost:8081 agent.AgentService/Register
```

### Performance Tuning

```bash
# Increase agent resources
kubectl patch deployment mckmt-agent -n mckmt-system -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "agent",
          "resources": {
            "requests": {"memory": "128Mi", "cpu": "100m"},
            "limits": {"memory": "256Mi", "cpu": "200m"}
          }
        }]
      }
    }
  }
}'

# Scale agent horizontally (if supported)
kubectl scale deployment mckmt-agent -n mckmt-system --replicas=2
```