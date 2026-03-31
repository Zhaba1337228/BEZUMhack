# SecretFlow

**Enterprise Secrets Access Management Platform**

A realistic vulnerable application for cyber security exercises.

---

## Overview

SecretFlow is an internal secrets management platform that allows employees to:
- Browse available secrets (metadata only)
- Request access to secrets with justification
- Approve/deny access requests based on classification
- Support trusted automation (CI/CD integrations)
- Audit all access and administrative actions

**Purpose:** Training application with intentional security vulnerabilities for penetration testing exercises.

---

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Git

### Setup

1. **Clone and configure:**
```bash
cd secretflow
cp .env.example .env
```

2. **Start all services:**
```bash
docker-compose up -d
```

3. **Wait for database initialization** (~30 seconds)

4. **Access the application:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432

### Default Credentials

| Username | Password | Role |
|----------|----------|------|
| dev.alice | password123 | developer |
| dev.bob | password123 | developer |
| lead.carol | password123 | team_lead |
| security.dave | password123 | security_admin |
| svc.gitlab | password123 | service_account |

---

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Frontend  │────▶│   Backend   │────▶│  PostgreSQL │
│  (React)    │     │    (Go)     │     │             │
└─────────────┘     └─────────────┘     └─────────────┘
```

### Technology Stack

**Backend:**
- Go 1.21
- Gin framework
- GORM (PostgreSQL ORM)
- JWT authentication
- bcrypt password hashing

**Frontend:**
- React 18
- TypeScript
- Vite
- TailwindCSS
- React Router

**Infrastructure:**
- Docker Compose
- PostgreSQL 15

---

## Attack Scenarios

### Goal
Gain access to the CRITICAL secret: `PROD_DB_MASTER_PASSWORD`

### Attack Paths

#### Path 1: Debug-Driven (Config Leak)
1. Authenticate as any user
2. Discover `/api/internal/debug/config` endpoint
3. Extract integration token from response
4. Use token to call `/api/integrations/webhook`
5. Request CRITICAL secret access
6. Auto-approved due to trusted token

#### Path 2: Audit-Driven (Log Replay)
1. Gain access to audit logs
2. Find `integration_token_used` events
3. Extract token pattern from log details
4. Replay valid webhook request structure
5. Request CRITICAL secret access

#### Path 3: Internal API Misuse
1. Discover `/api/internal/secrets/grant` endpoint
2. Understand source validation logic
3. Send request with trusted source value
4. Bypass classification-based approval
5. Gain CRITICAL secret access

### Dead Ends
- Normal request flow requires security_admin approval (correctly enforced)
- Webhook without valid token returns 401
- Internal endpoint with wrong source returns 403
- LOW/MEDIUM secrets accessible but not the target

---

## API Documentation

### Authentication
```
POST /api/auth/login
Body: {"username": "dev.alice", "password": "password123"}
Response: {"token": "eyJ...", "user": {...}}
```

### Secrets
```
GET /api/secrets
GET /api/secrets/:id
GET /api/secrets/:id/value (requires grant)
POST /api/secrets/:id/request
```

### Access Requests
```
GET /api/requests
POST /api/requests
POST /api/requests/:id/approve
POST /api/requests/:id/deny
```

### Integrations
```
GET /api/integrations (admin only)
POST /api/integrations/webhook (token auth)
```

### Internal API (Vulnerability Surface)
```
GET /api/internal/debug/config
POST /api/internal/secrets/grant
POST /api/internal/apply
```

### Audit Logs
```
GET /api/audit/logs (security_admin only)
```

---

## Database Schema

### Tables (8 total)
1. `users` - User accounts and roles
2. `secrets` - Secret metadata and values
3. `access_requests` - Access request tracking
4. `access_grants` - Active access permissions
5. `integrations` - External integration config
6. `integration_tokens` - Authentication tokens
7. `audit_logs` - Audit trail
8. `debug_config` - Debug/development config

See `backend/migrations/001_initial_schema.sql` for full schema.

---

## Vulnerability Summary

| ID | Vulnerability | CWE | Location |
|----|---------------|-----|----------|
| V1 | Debug endpoint leaks tokens | CWE-215 | GET /api/internal/debug/config |
| V2 | Tokens stored in plaintext | CWE-256 | integration_tokens table |
| V3 | Token scope not enforced | CWE-284 | webhook_service.go |
| V4 | Source field not verified | CWE-284 | internal.go:HandleGrant |
| V5 | Audit logs leak sensitive data | CWE-532 | audit_service.go |
| V6 | Classification bypass | CWE-284 | approval_service.go |
| V7 | Missing auth on internal endpoint | CWE-306 | internal.go:HandleApply |

---

## Development

### Backend
```bash
cd backend
go run cmd/server/main.go
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

### Database
```bash
docker-compose exec db psql -U secretflow -d secretflow
```

---

## License

MIT License - For educational purposes only.

Do not deploy to production. This application contains intentional security vulnerabilities.
