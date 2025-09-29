#!/bin/bash

# MCKMT Minikube Demo Script
# This script sets up a complete MCKMT demo with minikube

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLUSTER_NAME="mckmt-demo"
HUB_PORT=8080
GRPC_PORT=8081

echo -e "${BLUE}🚀 Starting MCKMT Minikube Demo${NC}"

# Check prerequisites
echo -e "${YELLOW}📋 Checking prerequisites...${NC}"

if ! command -v minikube &> /dev/null; then
    echo -e "${RED}❌ minikube is not installed. Please install minikube first.${NC}"
    exit 1
fi

if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl is not installed. Please install kubectl first.${NC}"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ docker is not installed. Please install docker first.${NC}"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ go is not installed. Please install Go 1.21+ first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All prerequisites found${NC}"

# Start minikube
echo -e "${YELLOW}🔧 Starting minikube cluster...${NC}"

# Check if minikube is already running
if minikube status --profile=$CLUSTER_NAME &> /dev/null; then
    echo -e "${YELLOW}⚠️  Minikube cluster '$CLUSTER_NAME' already exists. Deleting...${NC}"
    minikube delete --profile=$CLUSTER_NAME
fi

# Start minikube with sufficient resources
minikube start --profile=$CLUSTER_NAME --memory=4096 --cpus=2 --driver=docker

# Set kubectl context
kubectl config use-context $CLUSTER_NAME

echo -e "${GREEN}✅ Minikube cluster started${NC}"

# Build MCKMT binaries
echo -e "${YELLOW}🔨 Building MCKMT binaries...${NC}"

# Build hub and agent
go build -o bin/hub cmd/hub/main.go
go build -o bin/agent cmd/agent/main.go

echo -e "${GREEN}✅ Binaries built${NC}"

# Start dependencies
echo -e "${YELLOW}🐘 Starting dependencies (PostgreSQL, Redis)...${NC}"

# Start dependencies using docker-compose
docker-compose up -d postgres redis prometheus grafana

# Wait for dependencies to be ready
echo -e "${YELLOW}⏳ Waiting for dependencies to be ready...${NC}"
sleep 10

# Run database migrations
echo -e "${YELLOW}🗄️  Running database migrations...${NC}"
go run cmd/migrate/main.go migrate

echo -e "${GREEN}✅ Dependencies ready${NC}"

# Start hub in background
echo -e "${YELLOW}🏢 Starting MCKMT Hub...${NC}"

# Start hub in background
./bin/hub > hub.log 2>&1 &
HUB_PID=$!

# Wait for hub to start
echo -e "${YELLOW}⏳ Waiting for hub to start...${NC}"
sleep 5

# Check if hub is running
if ! curl -s http://localhost:$HUB_PORT/health > /dev/null; then
    echo -e "${RED}❌ Hub failed to start. Check hub.log for details.${NC}"
    kill $HUB_PID 2>/dev/null || true
    exit 1
fi

echo -e "${GREEN}✅ Hub started (PID: $HUB_PID)${NC}"

# Register cluster
echo -e "${YELLOW}📝 Registering cluster with MCKMT...${NC}"

CLUSTER_RESPONSE=$(curl -s -X POST http://localhost:$HUB_PORT/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "'$CLUSTER_NAME'",
    "mode": "agent",
    "labels": {
      "env": "demo",
      "provider": "minikube"
    }
  }')

if echo "$CLUSTER_RESPONSE" | grep -q "error"; then
    echo -e "${RED}❌ Failed to register cluster: $CLUSTER_RESPONSE${NC}"
    kill $HUB_PID 2>/dev/null || true
    exit 1
fi

echo -e "${GREEN}✅ Cluster registered${NC}"

# Create agent deployment
echo -e "${YELLOW}🤖 Creating agent deployment...${NC}"

cat > agent-deployment.yaml << EOF
apiVersion: v1
kind: Namespace
metadata:
  name: mckmt-system
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
        image: busybox
        command: ["/bin/sh"]
        args: ["-c", "while true; do sleep 30; done"]
        env:
        - name: MCKMT_HUB_URL
          value: "host.docker.internal:$GRPC_PORT"
        - name: MCKMT_CLUSTER_ID
          value: "$CLUSTER_NAME"
        - name: MCKMT_AGENT_ID
          value: "agent-001"
        volumeMounts:
        - name: agent-binary
          mountPath: /usr/local/bin/agent
        - name: kubeconfig
          mountPath: /root/.kube
          readOnly: true
      volumes:
      - name: agent-binary
        hostPath:
          path: $(pwd)/bin/agent
          type: File
      - name: kubeconfig
        hostPath:
          path: ~/.kube/config
          type: File
EOF

# Deploy agent
kubectl apply -f agent-deployment.yaml

echo -e "${GREEN}✅ Agent deployment created${NC}"

# Wait for agent to be ready
echo -e "${YELLOW}⏳ Waiting for agent to be ready...${NC}"
kubectl wait --for=condition=available --timeout=60s deployment/mckmt-agent -n mckmt-system

echo -e "${GREEN}✅ Agent is ready${NC}"

# Display status
echo -e "${BLUE}🎉 MCKMT Demo Setup Complete!${NC}"
echo ""
echo -e "${GREEN}📊 Access Points:${NC}"
echo -e "  • Hub API: http://localhost:$HUB_PORT"
echo -e "  • API Docs: http://localhost:$HUB_PORT/swagger/index.html"
echo -e "  • gRPC API: localhost:$GRPC_PORT"
echo -e "  • Minikube Dashboard: minikube dashboard --profile=$CLUSTER_NAME"
echo ""
echo -e "${GREEN}🔧 Useful Commands:${NC}"
echo -e "  • Check clusters: curl http://localhost:$HUB_PORT/api/v1/clusters"
echo -e "  • Check agent logs: kubectl logs -n mckmt-system deployment/mckmt-agent"
echo -e "  • Stop demo: ./scripts/stop-demo.sh"
echo ""
echo -e "${YELLOW}📝 Next Steps:${NC}"
echo -e "  1. Open http://localhost:$HUB_PORT/swagger/index.html in your browser"
echo -e "  2. Try the API examples from docs/EXAMPLES.md"
echo -e "  3. Deploy some applications to your minikube cluster"
echo ""

# Create stop script
cat > scripts/stop-demo.sh << 'EOF'
#!/bin/bash
echo "🛑 Stopping MCKMT Demo..."

# Kill hub process
if [ ! -z "$HUB_PID" ]; then
    kill $HUB_PID 2>/dev/null || true
fi

# Stop minikube
minikube stop --profile=mckmt-demo

# Stop dependencies
docker-compose down

echo "✅ Demo stopped"
EOF

chmod +x scripts/stop-demo.sh

echo -e "${GREEN}✅ Stop script created: scripts/stop-demo.sh${NC}"
echo ""
echo -e "${BLUE}🎯 Demo is ready! The hub is running in the background.${NC}"
echo -e "${YELLOW}💡 To stop the demo, run: ./scripts/stop-demo.sh${NC}"
