# Observability Stack

This directory contains the configuration for the observability stack using Grafana, Loki, and Promtail.

## Components

- **Grafana**: Visualization and dashboards (http://localhost:3000)
- **Loki**: Log aggregation system
- **Promtail**: Log collector that scrapes logs from Docker containers

## Quick Start

The observability stack is automatically started when you run:

```bash
docker compose up -d
```

## Accessing Grafana

1. Open your browser and navigate to http://localhost:3000
2. Login with default credentials:
   - Username: `admin`
   - Password: `admin` (you'll be prompted to change this on first login)

You can customize these credentials by setting environment variables:
```bash
export GRAFANA_USER=yourusername
export GRAFANA_PASSWORD=yourpassword
```

## Viewing Logs

1. In Grafana, click on "Explore" in the left sidebar
2. Select "Loki" as the data source (it's pre-configured)
3. Use LogQL queries to filter logs:

### Example Queries

**View all logs from backend:**
```logql
{job="poker-backend"}
```

**View all logs from frontend:**
```logql
{job="poker-frontend"}
```

**View all logs from database:**
```logql
{job="poker-db"}
```

**Filter by log level (if using structured logging):**
```logql
{job="poker-backend"} |= "error"
```

**View logs from all poker services:**
```logql
{job=~"poker-.*"}
```

## Log Retention

By default, logs are stored in Docker volumes:
- `loki_data`: Stores Loki's log data
- `grafana_data`: Stores Grafana dashboards and settings

To clear logs and start fresh:
```bash
docker compose down -v
```

## Architecture

```
Docker Containers → Promtail → Loki → Grafana
                    (collect)  (store) (visualize)
```

1. **Promtail** scrapes logs from Docker containers via Docker socket
2. **Loki** aggregates and indexes logs
3. **Grafana** provides a UI to query and visualize logs

## Troubleshooting

### Logs not appearing

1. Check if Promtail is running:
   ```bash
   docker logs poker-promtail
   ```

2. Check if Loki is healthy:
   ```bash
   curl http://localhost:3100/ready
   ```

3. Verify Promtail can reach Loki:
   ```bash
   docker exec poker-promtail wget -O- http://loki:3100/ready
   ```

### Cannot access Grafana

1. Ensure Grafana container is running:
   ```bash
   docker ps | grep grafana
   ```

2. Check Grafana logs:
   ```bash
   docker logs poker-grafana
   ```

## Configuration Files

- `loki-config.yaml`: Loki configuration
- `promtail-config.yaml`: Promtail configuration (defines how to scrape logs)
- `grafana/provisioning/datasources/loki.yaml`: Auto-provisions Loki as a datasource in Grafana
