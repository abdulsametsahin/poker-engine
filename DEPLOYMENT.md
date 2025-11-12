# Docker Deployment Guide

This guide will help you deploy the Poker Platform using Docker on Digital Ocean or any other server.

## Prerequisites

- Docker (20.10+)
- Docker Compose (1.29+)
- A server with at least 2GB RAM
- Domain name (optional, for production)

## Quick Start

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd poker-engine
```

### 2. Configure Environment Variables

Copy the example environment file and update it with your values:

```bash
cp .env.example .env
```

Edit `.env` and update the following:

```bash
# IMPORTANT: Change these values for production!
DB_ROOT_PASSWORD=your_secure_root_password
DB_PASSWORD=your_secure_db_password
JWT_SECRET=your_very_long_random_secret_key

# For local testing, keep these as is
REACT_APP_API_URL=http://localhost:8080/api
REACT_APP_WS_URL=ws://localhost:8080/ws
```

### 3. Start the Application

```bash
docker compose up -d --build
```

This single command will:
- Build the Go backend
- Build the React frontend
- Start MySQL database
- Initialize database schema
- Start all services

### 4. Verify Deployment

Check if all services are running:

```bash
docker compose ps
```

You should see multiple services running:
- `poker-db` (MySQL)
- `poker-backend` (Go API)
- `poker-frontend` (React + Nginx)
- `poker-engine-dbgate` (Database Management)
- `poker-loki` (Log Aggregation)
- `poker-promtail` (Log Collector)
- `poker-grafana` (Monitoring)

### 5. Access the Application

- **Frontend**: http://localhost
- **Backend API**: http://localhost:8080/api
- **WebSocket**: ws://localhost:8080/ws
- **DbGate** (Database Management): http://localhost:10080
- **Grafana** (Monitoring): http://localhost:3000

## Production Deployment on Digital Ocean

### 1. Create a Droplet

1. Log in to Digital Ocean
2. Create a new Droplet
3. Choose Ubuntu 22.04 LTS
4. Select at least 2GB RAM plan
5. Add your SSH key

### 2. Install Docker on Droplet

SSH into your droplet and run:

```bash
# Update packages
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Add your user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

### 3. Deploy the Application

```bash
# Clone your repository
git clone <your-repo-url>
cd poker-engine

# Configure environment
cp .env.example .env
nano .env  # Edit with your production values

# For production with a domain, update:
# REACT_APP_API_URL=https://yourdomain.com/api
# REACT_APP_WS_URL=wss://yourdomain.com/ws

# Start the application
docker compose up -d --build
```

### 4. Configure Firewall

```bash
# Allow HTTP, HTTPS, and SSH
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 8080/tcp  # Backend API (optional, can be removed if using reverse proxy)
sudo ufw enable
```

### 5. Point Your Domain

In your domain registrar's DNS settings:
- Add an A record pointing to your droplet's IP address
- Example: `poker.yourdomain.com` → `YOUR_DROPLET_IP`

### 6. Set Up SSL (Optional but Recommended)

For HTTPS, you can use Nginx as a reverse proxy with Let's Encrypt:

```bash
# Install Nginx and Certbot
sudo apt install nginx certbot python3-certbot-nginx -y

# Get SSL certificate
sudo certbot --nginx -d yourdomain.com

# Configure Nginx (example config provided below)
```

## Docker Commands

### View Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f backend
docker compose logs -f frontend
docker compose logs -f db
docker compose logs -f dbgate
```

### Restart Services

```bash
# All services
docker compose restart

# Specific service
docker compose restart backend
```

### Stop Services

```bash
docker compose down
```

### Stop and Remove All Data

```bash
docker compose down -v  # WARNING: This removes database data!
```

### Rebuild After Code Changes

```bash
docker compose up -d --build
```

### Access Database

```bash
docker compose exec db mysql -u poker_user -p poker_platform
```

## Troubleshooting

### Services Won't Start

Check logs for errors:
```bash
docker compose logs
```

### Database Connection Issues

1. Ensure database is healthy:
```bash
docker compose ps
```

2. Check database logs:
```bash
docker compose logs db
```

3. Verify environment variables in `.env`

### Frontend Can't Connect to Backend

1. Check backend is running:
```bash
curl http://localhost:8080/api/tables
```

2. Verify `REACT_APP_API_URL` in `.env` matches your backend URL

3. If using a domain, ensure CORS is properly configured in the backend

### Port Already in Use

If ports 80, 8080, or 3306 are already in use, edit `docker-compose.yml` to use different ports:

```yaml
services:
  frontend:
    ports:
      - "8081:80"  # Access frontend on port 8081
```

## Architecture

```
┌─────────────────┐
│   Frontend      │
│  (React + Nginx)│
│   Port 80       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Backend       │
│  (Go + WS)      │
│   Port 8080     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Database      │
│   (MySQL 8.0)   │
│   Port 3306     │
└─────────────────┘
```

## Backup and Restore

### Backup Database

```bash
docker compose exec db mysqldump -u root -p poker_platform > backup.sql
```

### Restore Database

```bash
docker compose exec -T db mysql -u root -p poker_platform < backup.sql
```

## Performance Optimization

### For Production

1. **Use environment-specific builds**: Rebuild frontend with production API URLs
2. **Enable database backups**: Set up automated backups
3. **Monitor resources**: Use `docker stats` to monitor container resource usage
4. **Scale if needed**: Consider using Docker Swarm or Kubernetes for scaling

### Resource Limits

You can add resource limits in `docker-compose.yml`:

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
```

## Security Checklist

- [ ] Changed default passwords in `.env`
- [ ] Set strong JWT_SECRET
- [ ] Configured firewall (UFW)
- [ ] Set up SSL/TLS (HTTPS)
- [ ] Regular database backups
- [ ] Keep Docker and images updated
- [ ] Use non-root users in containers (already configured)
- [ ] Limit exposed ports

## Support

For issues or questions, please open an issue in the repository.
