# Database Migrations

This directory contains database migrations for the poker platform.

## Migration Naming Convention

Migrations follow this naming pattern:
```
XXX_description_[up|down].sql
```

- `XXX` - Three-digit version number (001, 002, 003, etc.)
- `description` - Snake_case description of the migration
- `up` - SQL for applying the migration
- `down` - SQL for rolling back the migration

## Available Migrations

| Version | Name | Description |
|---------|------|-------------|
| 001 | initial_schema | Creates all core tables (users, tables, tournaments, etc.) |
| 002 | add_chip_transactions | Adds chip_transactions audit table for tracking all chip movements |

## Usage

### Run All Pending Migrations
```bash
cd platform/backend/scripts
./migrate.sh up
```

### Check Migration Status
```bash
./migrate.sh status
```

### Rollback Last Migration
```bash
./migrate.sh down
```

### Rollback Multiple Migrations
```bash
./migrate.sh down 3  # Rollback last 3 migrations
```

### Create New Migration
```bash
./migrate.sh create add_user_preferences
```

This creates two files:
- `XXX_add_user_preferences_up.sql`
- `XXX_add_user_preferences_down.sql`

## Environment Variables

Configure database connection using these environment variables:

```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=poker_user
export DB_PASSWORD=poker_password
export DB_NAME=poker_platform
```

## Migration Guidelines

### Writing Migrations

**DO:**
- ✅ Use `IF NOT EXISTS` for CREATE TABLE statements
- ✅ Add indexes for foreign keys and frequently queried columns
- ✅ Include comments explaining the migration purpose
- ✅ Test both up AND down migrations
- ✅ Keep migrations idempotent when possible
- ✅ Use transactions for data migrations

**DON'T:**
- ❌ Modify existing migration files after they've been applied
- ❌ Delete migration files
- ❌ Skip version numbers
- ❌ Mix schema changes with data changes in the same migration

### Example Migration

**003_add_user_settings_up.sql:**
```sql
-- Migration: Add user settings table
-- Created: 2025-11-12
-- Description: Stores user preferences and settings

CREATE TABLE IF NOT EXISTS user_settings (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    setting_key VARCHAR(50) NOT NULL,
    setting_value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_user_setting (user_id, setting_key),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Record migration
INSERT INTO schema_migrations (version) VALUES (3);
```

**003_add_user_settings_down.sql:**
```sql
-- Migration Rollback: Add user settings table
-- Description: Removes user_settings table

DROP TABLE IF EXISTS user_settings;

-- Remove migration record
DELETE FROM schema_migrations WHERE version = 3;
```

## Schema Migrations Table

The `schema_migrations` table tracks which migrations have been applied:

```sql
CREATE TABLE schema_migrations (
    version INT PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Troubleshooting

### "Table already exists" error
If you get this error, the table was created manually. Either:
1. Drop the table and re-run the migration
2. Insert the version manually: `INSERT INTO schema_migrations (version) VALUES (X);`

### "Cannot connect to MySQL"
Check your environment variables and ensure MySQL is running:
```bash
mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p
```

### Migration failed midway
1. Check the error message
2. Manually fix the database if needed
3. Update `schema_migrations` to reflect actual state
4. Re-run or rollback as needed

## Best Practices

1. **Always create both up and down migrations** - Rollbacks should be possible
2. **Test migrations on a copy of production data** - Catch issues before production
3. **Keep migrations small and focused** - One logical change per migration
4. **Document breaking changes** - Add comments for any backwards-incompatible changes
5. **Backup before major migrations** - Always have a restore point

## CI/CD Integration

Add to your deployment pipeline:

```bash
# Run pending migrations
./scripts/migrate.sh up

# If migrations fail, deployment should stop
if [ $? -ne 0 ]; then
    echo "Migrations failed"
    exit 1
fi
```
