-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enum types
CREATE TYPE user_role AS ENUM ('developer', 'team_lead', 'security_admin', 'service_account');
CREATE TYPE secret_classification AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL');
CREATE TYPE secret_environment AS ENUM ('dev', 'staging', 'production');
CREATE TYPE request_status AS ENUM ('pending', 'approved', 'denied');
CREATE TYPE integration_provider AS ENUM ('gitlab', 'webhook', 'internal');

-- 1. users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'developer',
    team VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. secrets table
CREATE TABLE secrets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    classification secret_classification NOT NULL,
    environment secret_environment NOT NULL,
    owner_team VARCHAR(50) NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. access_requests table
CREATE TABLE access_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    secret_id UUID NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    justification TEXT NOT NULL,
    status request_status NOT NULL DEFAULT 'pending',
    auto_approved BOOLEAN NOT NULL DEFAULT false,
    source VARCHAR(50) NOT NULL DEFAULT 'ui',
    source_context JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    decided_at TIMESTAMP,
    decided_by UUID REFERENCES users(id)
);

-- 4. access_grants table
CREATE TABLE access_grants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id UUID NOT NULL REFERENCES access_requests(id) ON DELETE CASCADE,
    secret_id UUID NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT false
);

-- 5. integrations table
CREATE TABLE integrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL,
    provider integration_provider NOT NULL,
    project_name VARCHAR(100),
    enabled BOOLEAN NOT NULL DEFAULT true,
    config JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 6. integration_tokens table
CREATE TABLE integration_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    description VARCHAR(100),
    allowed_secrets TEXT[],
    allowed_environments secret_environment[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP
);

-- 7. audit_logs table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id UUID REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    details JSONB,
    ip_address VARCHAR(45)
);

-- 8. debug_config table
CREATE TABLE debug_config (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    sensitive BOOLEAN NOT NULL DEFAULT false,
    internal_only BOOLEAN NOT NULL DEFAULT false
);

-- Indexes
CREATE INDEX idx_access_requests_secret_id ON access_requests(secret_id);
CREATE INDEX idx_access_requests_user_id ON access_requests(user_id);
CREATE INDEX idx_access_requests_status ON access_requests(status);
CREATE INDEX idx_access_grants_user_id ON access_grants(user_id);
CREATE INDEX idx_access_grants_secret_id ON access_grants(secret_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_integration_tokens_token ON integration_tokens(token);
CREATE INDEX idx_secrets_classification ON secrets(classification);
CREATE INDEX idx_secrets_environment ON secrets(environment);

-- Seed Data

-- Users
-- dev.alice: password123 (weak leaked password)
-- Other accounts use strong random passwords generated for local organizer testing.
INSERT INTO users (username, email, password_hash, role, team) VALUES
('dev.alice', 'alice@company.com', '$2a$10$K0BH/NSZpSS25QH57gc/4.isvL12g/MinXpJyThN2NvFvEtTgMmMi', 'developer', 'platform'),
('dev.bob', 'bob@company.com', '$2a$10$ouGoB.H2T44A25irjhJHTeeCrQ2bJavqK2WCjtUvmuZke4zTxyOHu', 'developer', 'backend'),
('lead.carol', 'carol@company.com', '$2a$10$dfhal59orEh37UywQA0g4.vIIRKI8nVBVrzmyz72lEiyGSu2bDPOq', 'team_lead', 'platform'),
('security.dave', 'dave@company.com', '$2a$10$S0vdxlZ9zpA3nIwGVysBdOPiS/rn.YRS.jRgQTXEgnyAdquhSa.Bu', 'security_admin', 'security'),
('svc.gitlab', 'gitlab-ci@system.local', '$2a$10$TJ3NywXqR2obYmJLzMPYjOL3pPJ/TXzOf9jEowFaweKGj/VYRp4um', 'service_account', 'automation');

-- Secrets
INSERT INTO secrets (name, description, classification, environment, owner_team, value) VALUES
('DEV_DB_PASSWORD', 'Development database password', 'LOW', 'dev', 'platform', 'flag{dev_db_8f2c1a7e5b904d3a}'),
('STAGING_API_KEY', 'Staging API gateway key', 'MEDIUM', 'staging', 'backend', 'flag{staging_api_5c91d2e44abf76c8}'),
('PROD_API_KEY', 'Production API gateway key', 'HIGH', 'production', 'backend', 'flag{prod_api_c7e41bf289d56a30}'),
('PROD_DB_MASTER_PASSWORD', 'Production database master password - CRITICAL', 'CRITICAL', 'production', 'platform', 'flag{prod_db_master_3f8a6d1c9e247b50}'),
('AWS_ROOT_ACCESS_KEY', 'AWS root account access key', 'CRITICAL', 'production', 'security', 'flag{aws_root_91b7c4e2d85fa630}'),
('GITLAB_WEBHOOK_SECRET', 'GitLab webhook verification secret', 'MEDIUM', 'production', 'platform', 'flag{gitlab_webhook_2de9a41c7bf58360}');

-- Integrations
INSERT INTO integrations (name, provider, project_name, config) VALUES
('GitLab CI', 'gitlab', 'secretflow-prod', '{"webhook_url": "http://gitlab.internal/hooks/secretflow", "events": ["deploy", "release"]}'),
('Internal Automation', 'internal', 'secretflow-internal', '{"auto_approve_all": true}');

-- Integration Tokens (THE KEY VULNERABILITY - tokens stored in plaintext)
INSERT INTO integration_tokens (integration_id, token, description, allowed_secrets, allowed_environments) VALUES
((SELECT id FROM integrations WHERE name = 'GitLab CI'), 'gf_prod_abc123xyz789', 'GitLab Production Token', NULL, NULL),
((SELECT id FROM integrations WHERE name = 'Internal Automation'), 'internal_svc_key_001', 'Internal Service Token', NULL, NULL);

-- Debug Config (vulnerability surface)
INSERT INTO debug_config (key, value, sensitive, internal_only) VALUES
('integration.gitlab.token', 'gf_prod_abc123xyz789', true, true),
('integration.gitlab.webhook_url', 'http://gitlab.internal/hooks/secretflow', false, false),
('internal.api.trusted_sources', '["webhook", "internal", "service_mesh"]', false, true),
('debug.mode', 'verbose', false, false),
('internal.service.auth_header', 'X-Service-Authorization', true, true);

-- Initial Audit Log
INSERT INTO audit_logs (action, resource_type, details, ip_address) VALUES
('system_init', 'system', '{"message": "SecretFlow initialized"}', '127.0.0.1');

-- Seed audit event for Attack Path 2
INSERT INTO audit_logs (action, resource_type, resource_id, details, ip_address) VALUES
(
    'integration_token_used',
    'integration_token',
    (SELECT id FROM integration_tokens WHERE token = 'gf_prod_abc123xyz789'),
    '{
      "token_value": "gf_prod_abc123xyz789",
      "token_id": "seeded-from-migration",
      "integration_name": "GitLab CI",
      "provider": "gitlab",
      "project": "secretflow-prod",
      "request_body": {
        "justification": "CI/CD deployment seed event"
      },
      "diagnostic": {
        "trust_level": "auto_approved",
        "approval_bypassed": true
      }
    }',
    '127.0.0.1'
);
