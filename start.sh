#!/bin/bash

echo "🎰 Starting Poker Platform..."
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Error: Docker is not installed"
    echo "Please install Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker compose &> /dev/null; then
    echo "❌ Error: Docker Compose is not installed"
    echo "Please install Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Creating from .env.example..."
    cp .env.example .env
    echo "✅ Created .env file"
    echo "⚠️  IMPORTANT: Please edit .env and update the passwords before deploying to production!"
    echo ""
fi

# Build and start services
echo "🚀 Building and starting services..."
docker compose up -d --build

# Wait for services to be ready
echo ""
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check service status
echo ""
echo "📊 Service Status:"
docker compose ps

echo ""
echo "✅ Poker Platform is starting!"
echo ""
echo "📍 Access Points:"
echo "   Frontend:  http://localhost"
echo "   Backend:   http://localhost:8080/api"
echo "   WebSocket: ws://localhost:8080/ws"
echo ""
echo "📝 Useful Commands:"
echo "   View logs:        docker compose logs -f"
echo "   Stop services:    docker compose down"
echo "   Restart:          docker compose restart"
echo ""
echo "📖 For more information, see DEPLOYMENT.md"
