#!/bin/bash

# ============================================
# Backend Container Entrypoint Script
# ============================================
# This script runs database migrations before
# starting the poker server application
# ============================================

set -e

echo "============================================"
echo "  Poker Platform - Backend Startup"
echo "============================================"
echo ""

# Run database migrations
echo "Running database migrations..."
/root/scripts/migrate.sh up

echo ""
echo "Starting poker server..."

# Start the poker server
exec ./poker-server
