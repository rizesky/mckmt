#!/bin/bash

# MCKMT Dashboard Access Script
# This script helps you access the MCKMT monitoring dashboard

echo "🚀 MCKMT Monitoring Dashboard Access"
echo "===================================="
echo ""

# Check if docker-compose is running
if ! docker ps | grep -q "mckmt-grafana"; then
    echo "❌ MCKMT development environment is not running."
    echo "   Please run 'make dev' first to start the services."
    exit 1
fi

echo "✅ MCKMT services are running!"
echo ""

# Get the local IP address
LOCAL_IP=$(hostname -I | awk '{print $1}')

echo "📊 Access URLs:"
echo "   Grafana Dashboard: http://localhost:3000"
echo "   Prometheus:        http://localhost:9090"
echo "   MCKMT Hub API:     http://localhost:8080"
echo "   MCKMT Metrics:     http://localhost:9091/metrics"
echo ""

echo "🔐 Default Credentials:"
echo "   Grafana: admin / admin"
echo "   Prometheus: No authentication required"
echo ""

echo "📈 Dashboard Features:"
echo "   • HTTP Request Rate & Duration"
echo "   • Connected Clusters & Agents"
echo "   • Operations in Progress"
echo "   • Database Connections"
echo "   • Cache Performance"
echo "   • Agent Heartbeats"
echo "   • Operation Success Rate"
echo ""

# Try to open the dashboard in the default browser
if command -v xdg-open > /dev/null; then
    echo "🌐 Opening Grafana dashboard in your browser..."
    xdg-open http://localhost:3000
elif command -v open > /dev/null; then
    echo "🌐 Opening Grafana dashboard in your browser..."
    open http://localhost:3000
else
    echo "💡 Please open http://localhost:3000 in your browser to access the dashboard."
fi
