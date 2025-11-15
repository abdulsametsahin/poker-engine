#!/bin/bash

# Optimized Docker build script with BuildKit
# This script enables BuildKit for faster builds with cache mounts

set -e

echo "🚀 Starting optimized Docker build..."

# Enable BuildKit
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Build with progress output
echo "📦 Building services with BuildKit cache..."
docker compose build --parallel

echo "✅ Build complete!"
echo ""
echo "To start the services, run:"
echo "  docker compose up -d"
