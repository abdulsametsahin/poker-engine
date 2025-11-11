# Table Recovery & Database Reset Guide

## Table of Contents
- [Table Recovery Feature](#table-recovery-feature)
- [Database Reset](#database-reset)
- [Docker Usage](#docker-usage)
- [Troubleshooting](#troubleshooting)

---

## Table Recovery Feature

### Overview
The poker engine now automatically recovers active tables on server restart. This prevents data loss when the server crashes or is restarted.

### What Gets Recovered?
- ✅ **Cash game tables** (status: waiting or playing)
- ✅ **Tournament tables** (status: starting or in_progress)
- ✅ **Player seats** and chip counts (from last hand completion)
- ✅ **Table configurations** (blinds, buy-ins, max players)

### What Doesn't Get Recovered?
- ❌ **Mid-hand state** (current bets, deck position, hole cards)
- ❌ **In-progress hands** (marked as cancelled on recovery)

### How It Works
1. **Server starts** → Database connection established
2. **Recovery process begins**:
   - Queries all active tables from database
   - Cleans up orphaned/incomplete hand records
   - Recreates engine tables with current player states
   - Adds players back to tables with their chip counts
3. **Game restart** (after 3 second delay):
   - Tables with 2+ active players start a new hand
   - Tournament blinds are restored to current level
   - All callbacks and event handlers are reconnected

### Recovery Logs
When the server starts, you'll see detailed recovery logs:

```
============================================================
🔄 STARTING TABLE RECOVERY PROCESS
============================================================
Found 3 active tables to recover
Recovering table abc123 (status: playing, type: cash)
  ✓ Added player Alice to seat 0 with 950 chips
  ✓ Added player Bob to seat 1 with 1050 chips
✓ Recovered table abc123 with 2 players
✓ Added 2 cash game tables to engine
✓ Added 1 tournament tables to engine
============================================================
📊 RECOVERY STATISTICS:
   Active Tables: 3
   Active Tournaments: 1
   Active Seats: 6
   Incomplete Hands: 2
============================================================
✅ TABLE RECOVERY COMPLETE
============================================================
```

### Limitations & Edge Cases

#### 1. **Mid-Hand Recovery**
If a hand is in progress when the server crashes:
- Current bets are **lost** (not recoverable)
- The hand is marked as cancelled
- Players keep chips from the **last completed hand**
- A new hand starts when recovery completes

**Mitigation**: Chip counts are synced to the database after **every hand completion**, minimizing potential loss.

#### 2. **Tournament Blind Levels**
Tournaments recover with the current blind level and structure. The blind timer restarts based on the stored `next_blind_increase_at` timestamp.

#### 3. **WebSocket Connections**
Players must **reconnect** their WebSocket connections after a server restart. The frontend should handle automatic reconnection.

#### 4. **Matchmaking Queue**
The in-memory matchmaking queue is **not persisted**. Players in the queue when the server restarts must rejoin.

**Recommendation**: The database stores matchmaking entries, so you could extend recovery to restore the queue if needed.

---

## Database Reset

### Overview
Sometimes you need to completely reset the database (e.g., during development, testing, or to clear corrupted data).

### ⚠️ WARNING
**Database reset will DELETE ALL DATA including:**
- User accounts
- Game history
- Tournament data
- All tables and hands
- Matchmaking queues

### Methods to Reset Database

#### Method 1: Using the Shell Script (Recommended)

**Local Development:**
```bash
cd /home/user/poker-engine/platform/backend/scripts
./reset_db.sh
```

**With Custom Credentials:**
```bash
DB_HOST=localhost \
DB_PORT=3306 \
DB_USER=poker_user \
DB_PASSWORD=poker_password \
DB_NAME=poker_platform \
./reset_db.sh
```

**Skip Confirmation (Dangerous!):**
```bash
SKIP_CONFIRM=true ./reset_db.sh
```

#### Method 2: Using Docker Exec (Production)

**Reset database in running Docker container:**
```bash
docker exec -it poker-db mysql \
  -u poker_user \
  -ppoker_password \
  poker_platform \
  < /docker-entrypoint-initdb.d/reset_database.sql
```

**Or using the mounted script:**
```bash
# First, copy the reset script to the container
docker cp platform/backend/scripts/reset_database.sql poker-db:/tmp/reset.sql

# Then execute it
docker exec -it poker-db mysql \
  -u poker_user \
  -ppoker_password \
  poker_platform \
  < /tmp/reset.sql
```

**Alternative: Interactive MySQL shell:**
```bash
# Enter MySQL shell in container
docker exec -it poker-db mysql -u poker_user -ppoker_password poker_platform

# Then run:
SOURCE /tmp/reset.sql;
```

#### Method 3: Direct SQL Execution

**If you have the SQL file:**
```bash
mysql -h localhost -P 3306 -u poker_user -ppoker_password poker_platform < reset_database.sql
```

**Or connect and source:**
```bash
mysql -h localhost -P 3306 -u poker_user -ppoker_password poker_platform
```
```sql
SOURCE /path/to/reset_database.sql;
```

---

## Docker Usage

### Checking Container Status
```bash
# List all containers
docker ps -a

# Check database container logs
docker logs poker-db

# Check backend container logs
docker logs poker-backend
```

### Accessing Database in Docker

**Method 1: Docker exec with mysql command**
```bash
docker exec -it poker-db mysql -u poker_user -ppoker_password poker_platform
```

**Method 2: Using environment variables**
```bash
docker exec -it poker-db mysql \
  -u ${DB_USER:-poker_user} \
  -p${DB_PASSWORD:-poker_password} \
  ${DB_NAME:-poker_platform}
```

### Resetting Database in Production Docker

**Step-by-step:**

1. **Copy the reset script to the container:**
   ```bash
   docker cp platform/backend/scripts/reset_database.sql poker-db:/tmp/reset.sql
   ```

2. **Execute the reset:**
   ```bash
   docker exec -i poker-db mysql \
     -u poker_user \
     -ppoker_password \
     poker_platform \
     < platform/backend/scripts/reset_database.sql
   ```

3. **Restart the backend to trigger recovery (optional):**
   ```bash
   docker restart poker-backend
   ```

4. **Verify the reset:**
   ```bash
   docker exec -it poker-db mysql \
     -u poker_user \
     -ppoker_password \
     -e "SELECT COUNT(*) FROM poker_platform.users;"
   ```

### Complete Docker Reset (Nuclear Option)

If you want to completely reset everything including volumes:

```bash
# Stop all containers
docker-compose down

# Remove volumes (THIS DELETES ALL DATA!)
docker volume rm poker-engine_db_data
docker volume rm poker-engine_loki_data
docker volume rm poker-engine_grafana_data

# Restart everything
docker-compose up -d

# The schema.sql will automatically run and create fresh tables
```

---

## Troubleshooting

### Recovery Issues

**Problem: Tables don't recover on startup**
```bash
# Check backend logs for recovery errors
docker logs poker-backend | grep -A 20 "RECOVERY PROCESS"

# Verify database has active tables
docker exec -it poker-db mysql -u poker_user -ppoker_password \
  -e "SELECT id, status, game_type FROM poker_platform.tables WHERE status IN ('waiting', 'playing');"
```

**Problem: Players don't reconnect after recovery**
- Players need to refresh the frontend or reconnect WebSocket
- Check that table_seats has `left_at = NULL` for active players
```bash
docker exec -it poker-db mysql -u poker_user -ppoker_password \
  -e "SELECT * FROM poker_platform.table_seats WHERE left_at IS NULL;"
```

**Problem: Games don't start after recovery**
- Verify at least 2 players have chips > 0
- Check logs for "Starting game on table X with N players"
- Recovery waits 3 seconds before starting games

### Reset Issues

**Problem: Permission denied when running script**
```bash
chmod +x platform/backend/scripts/reset_db.sh
```

**Problem: MySQL connection refused**
- Verify database container is running: `docker ps | grep poker-db`
- Check connection details in docker-compose.yml
- Try connecting manually first:
  ```bash
  docker exec -it poker-db mysql -u poker_user -ppoker_password -e "SHOW DATABASES;"
  ```

**Problem: Foreign key constraint errors**
- The script disables foreign key checks, but if you still get errors:
  ```sql
  SET FOREIGN_KEY_CHECKS = 0;
  -- Run your DROP/CREATE statements
  SET FOREIGN_KEY_CHECKS = 1;
  ```

### Common Commands

**View all tables:**
```bash
docker exec -it poker-db mysql -u poker_user -ppoker_password \
  -e "SHOW TABLES FROM poker_platform;"
```

**Count records in each table:**
```bash
docker exec -it poker-db mysql -u poker_user -ppoker_password poker_platform << 'EOF'
SELECT 'users' AS table_name, COUNT(*) AS count FROM users
UNION ALL SELECT 'tables', COUNT(*) FROM tables
UNION ALL SELECT 'tournaments', COUNT(*) FROM tournaments
UNION ALL SELECT 'table_seats', COUNT(*) FROM table_seats
UNION ALL SELECT 'hands', COUNT(*) FROM hands;
EOF
```

**Check for active games:**
```bash
docker exec -it poker-db mysql -u poker_user -ppoker_password \
  -e "SELECT t.id, t.name, t.status, COUNT(ts.id) as players
      FROM poker_platform.tables t
      LEFT JOIN poker_platform.table_seats ts ON t.id = ts.table_id AND ts.left_at IS NULL
      GROUP BY t.id;"
```

**View incomplete hands (orphaned data):**
```bash
docker exec -it poker-db mysql -u poker_user -ppoker_password \
  -e "SELECT * FROM poker_platform.hands WHERE completed_at IS NULL;"
```

---

## Best Practices

### Development
1. Use database reset frequently to test fresh installations
2. Create test data scripts to quickly populate the database
3. Test recovery by forcefully killing the backend container

### Production
1. **Never** reset the database in production unless absolutely necessary
2. Take database backups before any destructive operations
3. Monitor recovery logs on every deployment
4. Set up alerts for failed recovery attempts
5. Consider implementing mid-hand state persistence for critical games

### Backup Before Reset
```bash
# Create a backup before resetting
docker exec poker-db mysqldump \
  -u poker_user \
  -ppoker_password \
  poker_platform > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore from backup if needed
docker exec -i poker-db mysql \
  -u poker_user \
  -ppoker_password \
  poker_platform < backup_20250101_120000.sql
```

---

## Future Enhancements

### Potential Improvements
1. **Mid-hand state persistence**: Store deck state and current bets for full recovery
2. **Automatic backups**: Schedule periodic database backups
3. **Recovery metrics**: Track recovery success rate and timing
4. **Graceful shutdown**: Save in-progress hands before server stops
5. **Distributed recovery**: Support multiple server instances with leader election
6. **Recovery notifications**: Notify players when their tables are recovered

---

## Support

For issues or questions:
- Check the logs: `docker logs poker-backend`
- Review this guide's troubleshooting section
- Check the codebase: `platform/backend/internal/recovery/table_recovery.go`
