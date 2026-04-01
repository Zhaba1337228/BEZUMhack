# SecretFlow API Specification

## Base URL
```
http://localhost:8080/api
```

## Authentication

Most endpoints require JWT authentication via the `Authorization` header:
```
Authorization: Bearer <token>
```

---

## Endpoints

### Authentication

#### POST /api/auth/login
Login and receive JWT token.

**Request:**
```json
{
  "username": "dev.alice",
  "password": "password123"
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "username": "dev.alice",
    "email": "alice@company.com",
    "role": "developer",
    "team": "platform"
  }
}
```

#### GET /api/auth/me
Get current user info.

**Auth:** Required

**Response (200):**
```json
{
  "user": {
    "id": "uuid",
    "username": "dev.alice",
    ...
  }
}
```

---

### Secrets

#### GET /api/secrets
List available secrets.

**Auth:** Required

**Query Params:**
- `classification` - Filter by classification (LOW, MEDIUM, HIGH, CRITICAL)
- `environment` - Filter by environment (dev, staging, production)
- `owner_team` - Filter by owner team

**Response (200):**
```json
{
  "secrets": [
    {
      "id": "uuid",
      "name": "PROD_DB_MASTER_PASSWORD",
      "description": "Production database master password",
      "classification": "CRITICAL",
      "environment": "production",
      "owner_team": "platform",
      "has_access": false,
      "pending_request": false
    }
  ]
}
```

#### GET /api/secrets/:id
Get secret metadata.

**Auth:** Required

**Response (200):**
```json
{
  "secret": {
    "id": "uuid",
    "name": "PROD_DB_MASTER_PASSWORD",
    "description": "...",
    "classification": "CRITICAL",
    "environment": "production",
    "owner_team": "platform"
  }
}
```

#### GET /api/secrets/:id/value
Get secret value (requires active grant).

**Auth:** Required + Active Grant

**Response (200):**
```json
{
  "secret": {
    "id": "uuid",
    "name": "PROD_DB_MASTER_PASSWORD",
    "value": "flag{prod_db_master_3f8a6d1c9e247b50}"
  }
}
```

**Response (403):**
```json
{
  "error": "No active access grant. Please request access."
}
```

#### POST /api/secrets/:id/request
Request access to a secret.

**Auth:** Required

**Request:**
```json
{
  "justification": "Need access for production deployment"
}
```

**Response (201):**
```json
{
  "request": {
    "id": "uuid",
    "secret_id": "uuid",
    "status": "pending",
    "auto_approved": false,
    "requires_approval_from": "security_admin"
  }
}
```

---

### Access Requests

#### GET /api/requests
List access requests.

**Auth:** Required

**Query Params:**
- `pending=true` - Only pending requests
- `status=pending|approved|denied` - Filter by status

**Response (200):**
```json
{
  "requests": [
    {
      "id": "uuid",
      "secret_id": "uuid",
      "user_id": "uuid",
      "justification": "...",
      "status": "pending",
      "auto_approved": false,
      "source": "ui",
      "created_at": "2024-03-31T10:00:00Z",
      "secret": {...},
      "user": {...}
    }
  ]
}
```

#### POST /api/requests
Create access request.

**Auth:** Required

**Request:**
```json
{
  "secret_id": "uuid",
  "justification": "..."
}
```

#### POST /api/requests/:id/approve
Approve a request.

**Auth:** Required (team_lead or security_admin)

**Response (200):**
```json
{
  "grant": {
    "id": "uuid",
    "request_id": "uuid",
    "secret_id": "uuid",
    "user_id": "uuid",
    "granted_at": "...",
    "expires_at": "..."
  }
}
```

#### POST /api/requests/:id/deny
Deny a request.

**Auth:** Required (team_lead or security_admin)

**Response (200):**
```json
{
  "status": "denied"
}
```

---

### Audit Logs

#### GET /api/audit/logs
List audit logs.

**Auth:** Required (security_admin only)

**Query Params:**
- `user_id` - Filter by user
- `action` - Filter by action type
- `limit` - Max results (default: 100)

**Response (200):**
```json
{
  "logs": [
    {
      "id": "uuid",
      "timestamp": "2024-03-31T10:00:00Z",
      "user_id": "uuid",
      "action": "integration_token_used",
      "resource_type": "integration_token",
      "resource_id": "uuid",
      "details": {"token_prefix": "gf_prod_", ...},
      "ip_address": "10.0.0.1"
    }
  ]
}
```

#### GET /api/audit/stats
Get audit statistics.

**Auth:** Required (security_admin only)

**Response (200):**
```json
{
  "by_action": [
    {"action": "login_success", "count": 42},
    {"action": "integration_token_used", "count": 15}
  ]
}
```

---

### Integrations

#### GET /api/integrations
List integrations.

**Auth:** Required (security_admin only)

**Response (200):**
```json
{
  "integrations": [
    {
      "id": "uuid",
      "name": "GitLab CI",
      "provider": "gitlab",
      "project_name": "secretflow-prod",
      "enabled": true,
      "config": {...}
    }
  ]
}
```

#### POST /api/integrations/webhook
Webhook endpoint for trusted integrations.

**Auth:** Integration token (not user JWT)

**Request:**
```json
{
  "token": "gf_prod_abc123xyz789",
  "secret_id": "uuid",
  "justification": "Automated deployment"
}
```

**Response (200):**
```json
{
  "access_granted": true,
  "secret_value": "flag{prod_db_master_3f8a6d1c9e247b50}",
  "grant_expires_at": "2024-03-31T12:00:00Z",
  "request_id": "uuid"
}
```

---

### Internal API (Vulnerability Surface)

#### GET /api/internal/debug/config
Get debug configuration.

**Auth:** Required (any role)

**Response (200):**
```json
{
  "config": [
    {"key": "integration.gitlab.token", "value": "gf_prod_abc123xyz789", "sensitive": true},
    {"key": "internal.api.trusted_sources", "value": "[\"webhook\", \"internal\", \"service_mesh\"]"}
  ]
}
```

#### GET /api/internal/debug/integration
Get integration debug info.

**Auth:** Required

**Response (200):**
```json
{
  "integrations": [...],
  "tokens": [...]
}
```

#### POST /api/internal/secrets/grant
Create direct access grant.

**Auth:** Required

**Request:**
```json
{
  "secret_id": "uuid",
  "user_id": "uuid",
  "source": "webhook"
}
```

**Response (200):**
```json
{
  "grant": {
    "id": "uuid",
    "auto_approved": true,
    "secret_value": "...",
    "expires_at": "..."
  }
}
```

#### POST /api/internal/apply
Apply auto-approval logic.

**Auth:** None (vulnerability)

**Request:**
```json
{
  "request_id": "uuid",
  "bypass_classification_check": true,
  "source": "internal"
}
```

**Response (200):**
```json
{
  "approved": true,
  "reason": "Internal source - auto-approved",
  "grant_id": "uuid"
}
```

---

### Dashboard

#### GET /api/dashboard/summary
Get dashboard summary.

**Auth:** Required

**Response (200):**
```json
{
  "secrets_by_classification": [
    {"classification": "CRITICAL", "count": 2},
    {"classification": "HIGH", "count": 1}
  ],
  "my_pending_requests": 1,
  "my_active_grants": 3,
  "pending_approvals": 5
}
```

#### GET /api/dashboard/pending
Get pending requests count.

**Auth:** Required (team_lead or security_admin)

**Response (200):**
```json
{
  "pending": 5
}
```

---

## Role Permissions

| Endpoint | developer | team_lead | security_admin |
|----------|-----------|-----------|----------------|
| GET /api/secrets | ✓ | ✓ | ✓ |
| POST /api/secrets/:id/request | ✓ | ✓ | ✓ |
| GET /api/secrets/:id/value | ✓ (with grant) | ✓ (with grant) | ✓ (with grant) |
| GET /api/requests | ✓ (own) | ✓ (own) | ✓ (all) |
| POST /api/requests/:id/approve | ✗ | ✓ | ✓ |
| GET /api/audit/logs | ✗ | ✗ | ✓ |
| GET /api/integrations | ✗ | ✗ | ✓ |
