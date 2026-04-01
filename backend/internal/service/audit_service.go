package service

import (
	"secretflow/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

type AuditAction string

const (
	ActionLoginSuccess        AuditAction = "login_success"
	ActionLoginFailure        AuditAction = "login_failure"
	ActionSecretView          AuditAction = "secret_view"
	ActionSecretAccessRequest AuditAction = "secret_access_request"
	ActionRequestApproved     AuditAction = "access_request_approved"
	ActionRequestDenied       AuditAction = "access_request_denied"
	ActionGrantCreated        AuditAction = "access_grant_created"
	ActionGrantRevoked        AuditAction = "access_grant_revoked"
	ActionIntegrationTokenUsed AuditAction = "integration_token_used"
	ActionIntegrationConfigUpdated AuditAction = "integration_config_updated"
	ActionInternalAPICall     AuditAction = "internal_api_call"
)

func (s *AuditService) Log(action AuditAction, userID *string, resourceType string, resourceID *string, details map[string]interface{}, ip string) error {
	log := &models.AuditLog{
		ID:           uuid.New().String(),
		UserID:       userID,
		Action:       string(action),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      (*models.JSONMap)(&details),
		IPAddress:    ip,
	}

	return models.CreateAuditLog(s.db, log)
}

// LogIntegrationTokenUsed logs token usage
// SECURITY FIX: No longer logs full token value - only token ID and masked prefix
func (s *AuditService) LogIntegrationTokenUsed(token *models.IntegrationToken, requestBody map[string]interface{}, ip string) error {
	// Log only non-sensitive operational data
	// Token value is NEVER logged to prevent replay attacks
	details := map[string]interface{}{
		"token_id":         token.ID,
		"token_prefix":     token.Token[:min(8, len(token.Token))] + "...", // Masked
		"integration_id":   token.IntegrationID,
		"integration_name": token.Integration.Name,
		"provider":         token.Integration.Provider,
		"project":          token.Integration.ProjectName,
		"request_type":     "webhook",
		"source_ip":        ip,
		"timestamp":        token.CreatedAt.Format("2006-01-02T15:04:05Z"),
		// SECURITY: token_value, request_body removed to prevent replay
	}

	return s.Log(ActionIntegrationTokenUsed, nil, "integration_token", &token.ID, details, ip)
}

// LogWebhookSuccess logs successful webhook processing
// SECURITY FIX: No longer logs token value
func (s *AuditService) LogWebhookSuccess(token *models.IntegrationToken, secretID string, userID string) error {
	details := map[string]interface{}{
		"message":          "Automated access granted via trusted integration",
		"integration":      token.Integration.Name,
		"project":          token.Integration.ProjectName,
		"secret_id":        secretID,
		"user_id":          userID,
		"approval_status":  "auto_approved",
		"reason":           "Trusted automation - integration token validated",
		"token_id":         token.ID, // Only log ID, not value
	}

	return s.Log(ActionGrantCreated, nil, "access_grant", nil, details, "10.0.0.50")
}
