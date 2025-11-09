# Poker Platform

Full-stack poker platform with React frontend and Go backend that integrates with the poker engine.

## Running Services

### 1. Database Setup

```bash
mysql -u root -p < backend/scripts/schema.sql
```

### 2. Backend

```bash
cd backend
cp .env.example .env
# Edit .env with your MySQL credentials
go mod download
go run cmd/server/main.go
```

Server runs on `http://localhost:8080`

### 3. Frontend

```bash
cd frontend
npm install
npm start
```

Frontend runs on `http://localhost:3000`

## Environment Variables

Edit `backend/.env`:

```
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=poker_platform
JWT_SECRET=your_secret_key
SERVER_PORT=8080
```

## Default Credentials

New users start with 1000 chips. Register at `/login` or use the quick match feature.
