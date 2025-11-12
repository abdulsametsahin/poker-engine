#!/bin/bash

# ============================================
# Database Migration Script
# ============================================
# Manages up/down migrations for the poker platform
# ============================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-poker_user}"
DB_PASSWORD="${DB_PASSWORD:-poker_password}"
DB_NAME="${DB_NAME:-poker_platform}"

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
MIGRATIONS_DIR="$SCRIPT_DIR/../migrations"

# Functions for colored output
print_error() {
    echo -e "${RED}❌ ERROR: $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Wait for MySQL server to be ready
wait_for_mysql() {
    local max_attempts=60
    local attempt=1

    print_info "Waiting for MySQL server to be ready..."
    echo "Host: $DB_HOST:$DB_PORT"

    while [ $attempt -le $max_attempts ]; do
        if mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1" &>/dev/null; then
            print_success "MySQL server is ready"
            return 0
        fi

        if [ $attempt -eq 1 ] || [ $((attempt % 5)) -eq 0 ]; then
            print_info "Attempt $attempt/$max_attempts - waiting..."
        fi

        sleep 2
        attempt=$((attempt + 1))
    done

    print_error "MySQL server not ready after $max_attempts attempts"
    exit 1
}

# Ensure database exists
ensure_database() {
    print_info "Ensuring database '$DB_NAME' exists..."

    if mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" \
        -e "CREATE DATABASE IF NOT EXISTS \`$DB_NAME\`;" &>/dev/null; then
        print_success "Database '$DB_NAME' is ready"
    else
        print_error "Failed to create database '$DB_NAME'"
        exit 1
    fi
}

# Ensure schema_migrations table exists
ensure_migrations_table() {
    print_info "Ensuring schema_migrations table exists..."

    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" <<-EOF &>/dev/null
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INT NOT NULL PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
EOF

    if [ $? -eq 0 ]; then
        print_success "Schema migrations table is ready"
    else
        print_error "Failed to create schema_migrations table"
        exit 1
    fi
}

# Initialize database (wait, create db, create migrations table)
init_database() {
    wait_for_mysql
    ensure_database
    ensure_migrations_table
}

# Get current migration version
get_current_version() {
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" \
        -sN -e "SELECT IFNULL(MAX(version), 0) FROM schema_migrations" 2>/dev/null || echo "0"
}

# Get available migrations
get_available_migrations() {
    local direction=$1
    local migrations=()

    for file in "$MIGRATIONS_DIR"/*_${direction}.sql; do
        if [ -f "$file" ]; then
            # Extract version number from filename (e.g., 001_name_up.sql -> 001)
            local version=$(basename "$file" | grep -oP '^\d+')
            migrations+=("$version")
        fi
    done

    # Sort and remove duplicates
    printf '%s\n' "${migrations[@]}" | sort -u
}

# Run a migration
run_migration() {
    local version=$1
    local direction=$2
    local file_pattern="${MIGRATIONS_DIR}/${version}_*_${direction}.sql"
    local migration_file=$(ls $file_pattern 2>/dev/null | head -n1)

    if [ ! -f "$migration_file" ]; then
        print_error "Migration file not found: $file_pattern"
        return 1
    fi

    local migration_name=$(basename "$migration_file" | sed "s/_${direction}\.sql//")

    print_info "Running migration: $migration_name ($direction)"

    if mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < "$migration_file"; then
        print_success "Migration $migration_name completed"
        return 0
    else
        print_error "Migration $migration_name failed"
        return 1
    fi
}

# Migrate up to a specific version
migrate_up() {
    local target_version=${1:-999}
    local current_version=$(get_current_version)

    print_info "Current version: $current_version"
    print_info "Target version: $target_version"

    # Get all available migrations
    local available_migrations=($(get_available_migrations "up"))

    if [ ${#available_migrations[@]} -eq 0 ]; then
        print_warning "No migrations found"
        return 0
    fi

    local migrations_applied=0

    for version in "${available_migrations[@]}"; do
        if [ "$version" -gt "$current_version" ] && [ "$version" -le "$target_version" ]; then
            if run_migration "$version" "up"; then
                ((migrations_applied++))
            else
                print_error "Migration failed. Stopping."
                return 1
            fi
        fi
    done

    if [ $migrations_applied -eq 0 ]; then
        print_success "Database is up to date (version: $current_version)"
    else
        print_success "Applied $migrations_applied migration(s)"
        print_success "Current version: $(get_current_version)"
    fi
}

# Migrate down by N steps
migrate_down() {
    local steps=${1:-1}
    local current_version=$(get_current_version)

    if [ "$current_version" -eq 0 ]; then
        print_warning "No migrations to rollback"
        return 0
    fi

    print_info "Current version: $current_version"
    print_info "Rolling back $steps step(s)"

    # Confirm rollback
    echo ""
    print_warning "This will ROLLBACK database changes!"
    read -p "Are you sure? (yes/no): " -r
    echo ""

    if [[ ! $REPLY =~ ^[Yy]es$ ]]; then
        print_info "Rollback cancelled"
        return 0
    fi

    # Get migrations to rollback (in reverse order)
    local available_migrations=($(get_available_migrations "down" | sort -rn))
    local migrations_rolled_back=0

    for version in "${available_migrations[@]}"; do
        if [ "$version" -le "$current_version" ] && [ $migrations_rolled_back -lt $steps ]; then
            if run_migration "$version" "down"; then
                ((migrations_rolled_back++))
            else
                print_error "Rollback failed. Stopping."
                return 1
            fi
        fi
    done

    print_success "Rolled back $migrations_rolled_back migration(s)"
    print_success "Current version: $(get_current_version)"
}

# Show migration status
show_status() {
    local current_version=$(get_current_version)
    local available_migrations=($(get_available_migrations "up"))

    echo "============================================"
    echo "  Database Migration Status"
    echo "============================================"
    echo ""
    echo "Database: $DB_NAME"
    echo "Host: $DB_HOST:$DB_PORT"
    echo "Current Version: $current_version"
    echo ""
    echo "Available Migrations:"
    echo ""

    if [ ${#available_migrations[@]} -eq 0 ]; then
        print_warning "No migrations found"
        return
    fi

    for version in "${available_migrations[@]}"; do
        local file_pattern="${MIGRATIONS_DIR}/${version}_*_up.sql"
        local migration_file=$(ls $file_pattern 2>/dev/null | head -n1)
        local migration_name=$(basename "$migration_file" | sed 's/_up\.sql//')

        if [ "$version" -le "$current_version" ]; then
            echo -e "${GREEN}  ✓ $migration_name${NC}"
        else
            echo -e "${YELLOW}  ○ $migration_name (pending)${NC}"
        fi
    done
    echo ""
}

# Create a new migration template
create_migration() {
    local name=$1

    if [ -z "$name" ]; then
        print_error "Migration name is required"
        echo "Usage: $0 create <migration_name>"
        exit 1
    fi

    # Get next version number
    local latest_version=$(ls "$MIGRATIONS_DIR" | grep -oP '^\d+' | sort -n | tail -1)
    local next_version=$(printf "%03d" $((10#$latest_version + 1)))

    # Create migration files
    local up_file="${MIGRATIONS_DIR}/${next_version}_${name}_up.sql"
    local down_file="${MIGRATIONS_DIR}/${next_version}_${name}_down.sql"

    cat > "$up_file" <<EOF
-- Migration: $name
-- Created: $(date +%Y-%m-%d)
-- Description: TODO

-- Your migration SQL here

-- Record migration
INSERT INTO schema_migrations (version) VALUES ($next_version);
EOF

    cat > "$down_file" <<EOF
-- Migration Rollback: $name
-- Description: TODO

-- Your rollback SQL here

-- Remove migration record
DELETE FROM schema_migrations WHERE version = $next_version;
EOF

    print_success "Created migration files:"
    echo "  - $up_file"
    echo "  - $down_file"
}

# Main command dispatcher
case "${1:-}" in
    up)
        init_database
        migrate_up "${2:-999}"
        ;;
    down)
        init_database
        migrate_down "${2:-1}"
        ;;
    status)
        init_database
        show_status
        ;;
    create)
        create_migration "$2"
        ;;
    *)
        echo "============================================"
        echo "  Database Migration Tool"
        echo "============================================"
        echo ""
        echo "Usage: $0 <command> [options]"
        echo ""
        echo "Commands:"
        echo "  up [version]     - Run pending migrations (optionally to specific version)"
        echo "  down [steps]     - Rollback N migrations (default: 1)"
        echo "  status           - Show current migration status"
        echo "  create <name>    - Create new migration files"
        echo ""
        echo "Examples:"
        echo "  $0 up                  # Run all pending migrations"
        echo "  $0 up 3                # Migrate up to version 3"
        echo "  $0 down                # Rollback last migration"
        echo "  $0 down 2              # Rollback last 2 migrations"
        echo "  $0 status              # Show current status"
        echo "  $0 create add_users    # Create new migration"
        echo ""
        echo "Environment Variables:"
        echo "  DB_HOST      (default: localhost)"
        echo "  DB_PORT      (default: 3306)"
        echo "  DB_USER      (default: poker_user)"
        echo "  DB_PASSWORD  (default: poker_password)"
        echo "  DB_NAME      (default: poker_platform)"
        echo ""
        exit 1
        ;;
esac
