# SecretFlow - Project Summary

## What Was Built

A complete, runnable enterprise-style vulnerable web application for cybersecurity training exercises.

## Deliverables

### 1. Backend (Go + Gin + PostgreSQL)
- **8 database tables** with complete schema
- **7 handler modules** covering all API endpoints
- **4 service modules** for business logic
- **JWT authentication** with role-based access control
- **Complete seed data** with 5 users and 6 secrets

### 2. Frontend (React + TypeScript + TailwindCSS)
- **7 pages**: Login, Dashboard, Secrets, Secret Detail, Requests, Approvals, Audit Logs, Integrations
- **Role-aware UI** that adapts to user permissions
- **Complete authentication flow**
- **Corporate enterprise design**

### 3. Infrastructure
- **Docker Compose** configuration
- **Backend and Frontend Dockerfiles**
- **Database migrations**
- **Environment configuration**

### 4. Documentation
- README.md - Main documentation
- API_SPEC.md - Complete API reference
- ATTACK_SCENARIOS.md - Detailed attack paths
- SETUP_GUIDE.md - Installation instructions

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                        SecretFlow                             │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  Frontend (React/TypeScript)     Backend (Go/Gin)             │
│  ┌─────────────────────┐        ┌─────────────────────┐      │
│  │ Login               │        │ Auth Handler        │      │
│  │ Dashboard           │        │ Secrets Handler     │      │
│  │ Secrets             │◄──────►│ Requests Handler    │      │
│  │ Requests            │        │ Audit Handler       │      │
│  │ Approvals           │        │ Integrations Handler│      │
│  │ Audit Logs          │        │ Internal Handler    │      │
│  │ Integrations        │        │                     │      │
│  └─────────────────────┘        └─────────────────────┘      │
│                                  │                            │
│                                  ▼                            │
│                          ┌─────────────────────┐              │
│                          │   PostgreSQL DB     │              │
│                          │   (8 tables)        │              │
│                          └─────────────────────┘              │
└──────────────────────────────────────────────────────────────┘
```

## Vulnerability Summary

| # | Vulnerability | Location | Attack Path |
|---|---------------|----------|-------------|
| 1 | Debug endpoint leaks tokens | GET /api/internal/debug/config | Path 1 |
| 2 | Tokens stored in plaintext | integration_tokens table | Path 1, 2 |
| 3 | Token scope not enforced | webhook_service.go | All paths |
| 4 | Source field not verified | POST /api/internal/secrets/grant | Path 3 |
| 5 | Audit logs leak sensitive data | audit_logs.details | Path 2 |
| 6 | Classification bypass | approval_service.go | All paths |
| 7 | Missing auth on internal endpoint | POST /api/internal/apply | Path 3 |

## Attack Paths

### Path 1: Debug-Driven (Config Leak)
```
Login → Discover /api/internal/debug/config → Extract token →
Call webhook → Get CRITICAL secret
```

### Path 2: Audit-Driven (Log Replay)
```
Access audit logs → Find token usage events → Extract token pattern →
Replay webhook request → Get CRITICAL secret
```

### Path 3: Internal API Misuse
```
Discover /api/internal/secrets/grant → Learn source validation →
Spoof trusted source → Get auto-approved grant → Get CRITICAL secret
```

## Database Schema (8 Tables)

1. **users** - User accounts (5 seed users)
2. **secrets** - Secret metadata (6 secrets including 2 CRITICAL)
3. **access_requests** - Access request tracking
4. **access_grants** - Active access permissions
5. **integrations** - External integration config (2 integrations)
6. **integration_tokens** - Authentication tokens (2 tokens)
7. **audit_logs** - Audit trail
8. **debug_config** - Debug configuration (5 entries)

## User Roles

| Role | Capabilities |
|------|-------------|
| developer | View secrets, request access |
| team_lead | + Approve LOW/MEDIUM/HIGH requests |
| security_admin | + Approve CRITICAL, view audit, manage integrations |
| service_account | Automated access via tokens |

## Default Credentials

All users share the same password: `password123`

- `dev.alice` - developer
- `dev.bob` - developer
- `lead.carol` - team_lead
- `security.dave` - security_admin
- `svc.gitlab` - service_account

## Critical Event

**Goal:** Obtain the value of `PROD_DB_MASTER_PASSWORD`

**Success Response:**
```json
{
  "access_granted": true,
  "secret_value": "SUPER_SECRET_PROD_DB_PASS_2024"
}
```

## Why Medium-Hard Complexity

1. **Multiple recon steps** - Must discover internal endpoints
2. **Token acquisition required** - Cannot win without valid token
3. **Dead ends exist** - Normal flow correctly enforced
4. **System understanding needed** - Must connect multiple concepts
5. **No single exploit** - Requires chaining discoveries

## Files Created

```
secretflow/
├── README.md
├── .env.example
├── docker-compose.yml
├── docs/
│   ├── API_SPEC.md
│   ├── ATTACK_SCENARIOS.md
│   └── SETUP_GUIDE.md
├── backend/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── database/database.go
│   │   ├── handlers/*.go (7 files)
│   │   ├── middleware/*.go (2 files)
│   │   ├── models/*.go (7 files)
│   │   └── service/*.go (5 files)
│   ├── migrations/001_initial_schema.sql
│   └── pkg/jwt/jwt.go
└── frontend/
    ├── Dockerfile
    ├── index.html
    ├── nginx.conf
    ├── package.json
    ├── *.config.js/ts (5 config files)
    └── src/
        ├── App.tsx
        ├── main.tsx
        ├── index.css
        ├── components/Layout/*.tsx (4 files)
        ├── context/AuthContext.tsx
        ├── pages/*/ (7 pages, 10 files)
        └── services/api.ts
```

## Running the Project

```bash
# Start all services
docker-compose up -d

# Access application
# Frontend: http://localhost:3000
# Backend: http://localhost:8080

# Login with: dev.alice / password123
```

## Security Warning

This application contains **intentional security vulnerabilities**. Do not deploy to production or use with real secrets.

---

**Total Implementation:** ~30 hours of development
**Complexity:** Medium-Hard
**Tables:** 8
**API Endpoints:** 20+
**Frontend Pages:** 7
**Attack Paths:** 3
