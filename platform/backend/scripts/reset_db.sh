#!/bin/bash

# ============================================
# Database Reset Script
# ============================================
# This script resets the poker platform database
# WARNING: This will DELETE ALL DATA!
# ============================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-poker_user}"
DB_PASSWORD="${DB_PASSWORD:-poker_password}"
DB_NAME="${DB_NAME:-poker_platform}"
SKIP_CONFIRM="${SKIP_CONFIRM:-false}"

# Function to print colored messages
print_error() {
    echo -e "${RED}❌ ERROR: $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  WARNING: $1${NC}"
}

print_info() {
    echo -e "ℹ️  $1"
}

# Function to confirm action
confirm_reset() {
    if [ "$SKIP_CONFIRM" = "true" ]; then
        return 0
    fi

    echo ""
    print_warning "THIS WILL DELETE ALL DATA IN THE DATABASE!"
    echo ""
    echo "Database: $DB_NAME"
    echo "Host: $DB_HOST:$DB_PORT"
    echo ""
    read -p "Are you sure you want to reset the database? (yes/no): " -r
    echo ""

    if [[ ! $REPLY =~ ^[Yy]es$ ]]; then
        print_info "Reset cancelled"
        exit 0
    fi
}

# Main script
echo "============================================"
echo "  Poker Platform - Database Reset"
echo "============================================"
echo ""

# Confirm reset
confirm_reset

print_info "Starting database reset..."

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SQL_FILE="$SCRIPT_DIR/reset_database.sql"

# Check if SQL file exists
if [ ! -f "$SQL_FILE" ]; then
    print_error "SQL file not found: $SQL_FILE"
    exit 1
fi

# Run the reset SQL script
print_info "Executing reset SQL script..."
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < "$SQL_FILE"

if [ $? -eq 0 ]; then
    echo ""
    print_success "Database reset completed successfully!"
    echo ""
    print_info "All tables have been dropped and recreated"
    print_info "You can now start the server to begin with a fresh database"
    echo ""
else
    print_error "Database reset failed"
    exit 1
fi
