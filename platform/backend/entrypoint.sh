#!/bin/bash

# ============================================
# Backend Container Entrypoint Script
# ============================================
# Migrations run automatically on startup
# ============================================

set -e

echo "============================================"
echo "  Poker Platform - Backend Startup"
echo "============================================"
echo ""

echo "Starting poker server..."
echo "(Migrations will run automatically)"
echo ""

# Start the poker server
exec ./poker-server
