# Migration Guide - Currency System Refactor

## Quick Start

### 1. Set Environment Variables

```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=root
export DB_NAME=poker_engine
```

Or create `.env` file in `platform/backend/`:
```bash
cp .env.example .env
# Edit .env with your database credentials
```

### 2. Check Migration Status

```bash
cd platform/backend/scripts
./migrate.sh status
```

### 3. Run Migrations

```bash
# Run all pending migrations
./migrate.sh up
```

This will:
- Create `schema_migrations` tracking table (if needed)
- Create `chip_transactions` audit table
- Set up indexes for performance

### 4. Verify Migration

```bash
# Check that chip_transactions table exists
mysql -u root -proot poker_engine -e "DESCRIBE chip_transactions;"
```

Expected output:
```
+------------------+--------------+------+-----+-------------------+
| Field            | Type         | Null | Key | Default           |
+------------------+--------------+------+-----+-------------------+
| id               | varchar(36)  | NO   | PRI | NULL              |
| user_id          | varchar(36)  | NO   | MUL | NULL              |
| amount           | int          | NO   |     | NULL              |
| balance_before   | int          | NO   |     | NULL              |
| balance_after    | int          | NO   |     | NULL              |
| transaction_type | varchar(50)  | NO   | MUL | NULL              |
| reference_id     | varchar(36)  | YES  | MUL | NULL              |
| description      | text         | YES  |     | NULL              |
| created_at       | timestamp    | YES  | MUL | CURRENT_TIMESTAMP |
+------------------+--------------+------+-----+-------------------+
```

## What Changed

### Database Schema

**New Table:** `chip_transactions`
- Tracks every chip movement (buy-ins, prizes, refunds)
- Stores before/after balances for audit trail
- Indexed for fast queries by user, type, and date

### Backend Code

**New Currency Service:** `platform/backend/internal/currency/`
- Atomic operations with row-level locking
- Balance validation before deductions
- Automatic audit trail creation

**Updated Prize System:**
- Changed from float percentages to integer basis points
- Eliminates floating-point rounding errors
- Exact prize distribution guaranteed

### Migration Files

- `001_initial_schema_up.sql` - Full schema (for new installs)
- `002_add_chip_transactions_up.sql` - Adds audit table (for existing installs)
- Corresponding `_down.sql` files for rollbacks

## Testing the Migration

### 1. Test Transaction Logging

Start the server and register for a tournament:

```bash
cd platform/backend
go run cmd/server/main.go
```

Then check the transactions:

```bash
mysql -u root -proot poker_engine -e "SELECT * FROM chip_transactions ORDER BY created_at DESC LIMIT 5;"
```

You should see:
- Buy-in deduction (negative amount)
- Prize additions (positive amount)
- Before/after balance verification

### 2. Verify Prize Calculations

Create a tournament with 3 players and 1000 chip buy-in:
- Prize pool: 3000 chips
- 1st place: 50% = 1500 chips
- 2nd place: 30% = 900 chips
- 3rd place: 20% = 600 chips
- **Total: 3000 chips** ✅ (exact, no rounding errors)

Check with:
```sql
SELECT
    position,
    prize_amount,
    SUM(prize_amount) OVER () as total_prizes,
    prize_pool
FROM tournament_players tp
JOIN tournaments t ON tp.tournament_id = t.id
WHERE t.id = 'your-tournament-id';
```

### 3. Check Audit Trail

```sql
-- See all transactions for a user
SELECT
    transaction_type,
    amount,
    balance_before,
    balance_after,
    description,
    created_at
FROM chip_transactions
WHERE user_id = 'your-user-id'
ORDER BY created_at DESC;
```

## Rollback Plan

If you need to rollback:

```bash
# Rollback chip_transactions table
./migrate.sh down

# Or rollback multiple migrations
./migrate.sh down 2
```

**⚠️ Warning:** Rolling back will:
- Drop `chip_transactions` table
- **Lose all transaction history**
- Backup first if data is important!

## Common Issues

### Issue: "Table already exists"

If tables were created manually:

```sql
-- Mark migrations as applied
INSERT INTO schema_migrations (version) VALUES (1);
INSERT INTO schema_migrations (version) VALUES (2);
```

### Issue: "User not found" errors

The currency service validates user existence. Ensure:
- User IDs are valid UUIDs
- Users exist in `users` table before operations

### Issue: "Insufficient chips" errors

This is intentional! The new system prevents:
- Negative balances
- Buy-ins without sufficient chips
- Invalid transaction amounts

Check user balance:
```sql
SELECT id, username, chips FROM users WHERE id = 'user-id';
```

## Production Deployment

### Pre-deployment Checklist

- [ ] Test migrations on staging database
- [ ] Backup production database
- [ ] Review transaction history requirements
- [ ] Update environment variables
- [ ] Test rollback procedure

### Deployment Steps

1. **Backup Database**
   ```bash
   mysqldump -u root -p poker_engine > backup_$(date +%Y%m%d_%H%M%S).sql
   ```

2. **Stop Application**
   ```bash
   # Stop your Go server
   ```

3. **Run Migrations**
   ```bash
   cd platform/backend/scripts
   ./migrate.sh up
   ```

4. **Verify Migration**
   ```bash
   ./migrate.sh status
   mysql -u root -p poker_engine -e "SELECT COUNT(*) FROM chip_transactions;"
   ```

5. **Start Application**
   ```bash
   cd platform/backend
   go run cmd/server/main.go
   ```

6. **Monitor Logs**
   - Check for currency service errors
   - Verify transactions are being logged
   - Test buy-in and prize flows

## Future Migrations

To create new migrations:

```bash
./migrate.sh create add_feature_name
```

This generates:
- `XXX_add_feature_name_up.sql`
- `XXX_add_feature_name_down.sql`

Edit these files with your schema changes.

## Support

If you encounter issues:

1. Check migration status: `./migrate.sh status`
2. Review database logs
3. Verify credentials in .env
4. Test MySQL connection: `mysql -u root -proot poker_engine`

## Next Steps

After successful migration:

1. ✅ Test tournament registration
2. ✅ Complete a tournament and verify prize distribution
3. ✅ Check `chip_transactions` table for audit trail
4. ✅ Verify all prizes sum to exact prize pool
5. ✅ Monitor for any currency-related errors

The currency system is now production-ready with:
- ✅ Atomic operations
- ✅ Balance validation
- ✅ Complete audit trail
- ✅ Precise prize calculations
- ✅ Thread-safe operations
