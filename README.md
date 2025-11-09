# Poker Engine

Production-ready Texas Hold'em poker engine with complete platform implementation.

## Project Structure

```
poker-engine/
├── engine/          # Stateless poker game engine (Go)
├── models/          # Data models
├── platform/        # Full-stack platform (Backend + Frontend)
│   ├── backend/     # Go REST API + WebSocket server
│   └── frontend/    # React + TypeScript + Material-UI
```

## Engine Features

- Stateless game engine
- Multi-player support (2-8 players)
- Side pot calculations for all-in scenarios
- Proper dealer/blind rotation
- Heads-up support
- Tournament and cash game modes
- Action timeouts
- Event system

## Platform

Complete poker platform with:
- User authentication (JWT)
- Table management
- Matchmaking system
- Real-time game synchronization via WebSocket
- MySQL persistence
- React frontend with live poker table view

## Running the Platform

See [platform/README.md](platform/README.md) for detailed setup instructions.

### Quick Start

```bash
# 1. Setup database
mysql -u root -p < platform/backend/scripts/schema.sql

# 2. Start backend
cd platform/backend
cp .env.example .env
# Edit .env with your credentials
go run cmd/server/main.go

# 3. Start frontend (in new terminal)
cd platform/frontend
npm install
npm start
```

Access the platform at `http://localhost:3000`

## Architecture

The platform uses a three-tier architecture:

```
Frontend (React) <--WebSocket--> Backend (Go) <--In-Memory--> Engine (Go)
                                      |
                                      v
                                  MySQL (Persistence)
```

- Frontend communicates only with backend
- Backend acts as bridge between frontend and stateless engine
- Backend manages game lifecycle (start games, advance rounds)
- Real-time sync via WebSocket
