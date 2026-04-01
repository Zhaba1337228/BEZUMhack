# SecretFlow - Setup Guide

## Quick Start with Docker

### Prerequisites

- Docker Desktop (Windows/Mac) or Docker + Docker Compose (Linux)
- Git

### Step 1: Clone and Configure

```bash
cd secretflow
cp .env.example .env
```

### Step 2: Start all services

```bash
docker-compose up -d
```

### Step 3: Wait for initialization

Wait approximately 30 seconds for the database to initialize and migrations to run.

Check logs:
```bash
docker-compose logs -f backend
```

You should see:
```
Database connection established
Starting SecretFlow server on 0.0.0.0:8080
```

### Step 4: Access the application

- **Frontend:** http://localhost:3000
- **Backend API:** http://localhost:8080
- **Health check:** http://localhost:8080/health

### Step 5: Login

Use these credentials:

| Username | Password | Role |
|----------|----------|------|
| dev.alice | password123 | developer |
| lead.carol | not disclosed (strong random) | team_lead |
| security.dave | not disclosed (strong random) | security_admin |

---

## Manual Setup (Development)

### Backend

```bash
cd backend

# Install dependencies
go mod download

# Run migrations (if not using Docker)
psql -U postgres -d secretflow -f migrations/001_initial_schema.sql

# Start server
go run cmd/server/main.go
```

Environment variables:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=secretflow
export DB_USER=postgres
export DB_PASSWORD=your_password
export JWT_SECRET=your_secret
```

### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start dev server
npm run dev
```

The frontend will be available at http://localhost:3000

---

## Database Schema

The application creates these tables:

1. `users` - User accounts and roles
2. `secrets` - Secret metadata and values
3. `access_requests` - Access request tracking
4. `access_grants` - Active access permissions
5. `integrations` - External integration config
6. `integration_tokens` - Authentication tokens
7. `audit_logs` - Audit trail
8. `debug_config` - Debug/development config

---

## Troubleshooting

### Backend won't start

Check database connection:
```bash
docker-compose logs db
docker-compose logs backend
```

Common issues:
- Database not ready (wait 30 seconds)
- Wrong credentials in .env file

### Frontend shows blank page

Check browser console for errors. Common issues:
- API URL not configured
- Backend not running

### Reset everything

```bash
docker-compose down -v
docker-compose up -d
```

---

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Frontend  │────▶│   Backend   │────▶│  PostgreSQL │
│  (React)    │     │    (Go)     │     │             │
│  Port 3000  │     │   Port 8080 │     │   Port 5432 │
└─────────────┘     └─────────────┘     └─────────────┘
```

---

## Next Steps

1. **Explore the UI** - Browse secrets, create requests
2. **Try the attack scenarios** - See `ATTACK_SCENARIOS.md`
3. **Review the API** - See `API_SPEC.md`
4. **Examine the code** - Understand the vulnerabilities

---

## Security Warning

This application contains **intentional security vulnerabilities** for educational purposes.

**DO NOT:**
- Deploy to production
- Use with real secrets
- Expose to the internet

This is a training tool only.
