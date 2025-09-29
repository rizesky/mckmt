#!/bin/bash

# Test script to verify caching functionality
echo "🧪 Testing MCKMT Caching Implementation"
echo "========================================"

# Start dependencies
echo "📦 Starting dependencies..."
docker compose up -d postgresql redis

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Build the application
echo "🔨 Building application..."
go build -o bin/hub cmd/hub/main.go

# Start the hub in background
echo "🚀 Starting hub server..."
./bin/hub &
HUB_PID=$!

# Wait for hub to start
sleep 5

# Test cluster resource caching
echo "🔍 Testing cluster resource caching..."

# Create a test cluster first
echo "Creating test cluster..."
CLUSTER_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "name": "test-cluster",
    "mode": "agent",
    "labels": {"env": "test"}
  }')

echo "Cluster creation response: $CLUSTER_RESPONSE"

# Extract cluster ID (this is a simplified approach)
CLUSTER_ID=$(echo $CLUSTER_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -n "$CLUSTER_ID" ]; then
    echo "✅ Test cluster created with ID: $CLUSTER_ID"
    
    # Test resource listing (first call - cache miss)
    echo "📋 Testing resource listing (first call - should be cache miss)..."
    RESOURCE_RESPONSE1=$(curl -s "http://localhost:8080/api/v1/clusters/$CLUSTER_ID/resources?kind=Pod" \
      -H "Authorization: Bearer test-token")
    echo "First call response: $RESOURCE_RESPONSE1"
    
    # Test resource listing (second call - cache hit)
    echo "📋 Testing resource listing (second call - should be cache hit)..."
    RESOURCE_RESPONSE2=$(curl -s "http://localhost:8080/api/v1/clusters/$CLUSTER_ID/resources?kind=Pod" \
      -H "Authorization: Bearer test-token")
    echo "Second call response: $RESOURCE_RESPONSE2"
    
    # Check if cached field is present
    if echo "$RESOURCE_RESPONSE2" | grep -q '"cached":true'; then
        echo "✅ Cache hit detected in second call!"
    else
        echo "❌ Cache hit not detected in second call"
    fi
else
    echo "❌ Failed to create test cluster"
fi

# Test metrics endpoint
echo "📊 Testing metrics endpoint..."
METRICS_RESPONSE=$(curl -s http://localhost:8080/api/v1/metrics)
echo "Metrics response: $METRICS_RESPONSE"

# Cleanup
echo "🧹 Cleaning up..."
kill $HUB_PID 2>/dev/null
docker compose down

echo "✅ Cache testing completed!"
