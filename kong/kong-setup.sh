#!/bin/bash

set -e

echo ""
echo "====================================================="
echo "   🚀 Kong Gateway Setup Script"
echo "====================================================="
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed"
    echo "   Please install Docker from https://www.docker.com/"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not installed"
    echo "   Please install Docker Compose"
    exit 1
fi

# Copy env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📋 Creating .env from .env.example..."
    cp .env.example .env
    echo "✅ .env created"
    echo "⚠️  Please update .env with your values"
    echo ""
else
    echo "✅ .env already exists"
    echo ""
fi

# Start Kong
echo "🐳 Starting Kong Gateway with Docker Compose..."
echo ""

# Try docker compose first (newer syntax), fall back to docker-compose
if docker compose version &> /dev/null; then
    docker compose up -d
else
    docker-compose up -d
fi

echo ""
echo "⏳ Waiting for Kong to be ready (15 seconds)..."
sleep 15

# Health check
echo ""
echo "🏥 Checking Kong health..."
if curl -s http://localhost:8001/status | grep -q "reachable"; then
    echo "✅ Kong is healthy and ready!"
else
    echo "⚠️  Kong might not be fully ready yet"
    echo "   Check logs: docker-compose logs kong-gateway"
fi

echo ""
echo "====================================================="
echo "   ✅ Kong Gateway Setup Complete!"
echo "====================================================="
echo ""
echo "📋 Access Points:"
echo "   Proxy:      http://localhost:8000"
echo "   Admin API:  http://localhost:8001"
echo ""
echo "🧪 Test Kong:"
echo "   curl http://localhost:8001/status"
echo ""
echo "🚀 Run the demo:"
echo "   go run main.go"
echo ""
echo "📊 View logs:"
echo "   docker-compose logs -f kong-gateway"
echo ""
echo "🛑 Stop Kong:"
echo "   docker-compose down"
echo ""
