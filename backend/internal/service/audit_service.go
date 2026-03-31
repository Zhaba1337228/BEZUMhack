package service

import (
	"secretflow/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

// LogIntegrationTokenUsed logs token usage with detailed operational info
// VULNERABILITY: Logs full token and request structure for "operational debugging"
func (s *AuditService) LogIntegrationTokenUsed(token *models.IntegrationToken, requestBody map[string]interface{}, ip string) error {
	// Build detailed operational log for troubleshooting integration issues
	// This is intentionally verbose to help ops debug production issues
	details := map[string]interface{}{
		"token_id":         token.ID,
		"token_value":      token.Token, // VULNERABILITY: Full token logged for debugging
		"token_prefix":     token.Token[:8],
		"integration_id":   token.IntegrationID,
		"integration_name": token.Integration.Name,
		"provider":         token.Integration.Provider,
		"project":          token.Integration.ProjectName,
		"request_body":     requestBody, // Includes full request structure for replay
		"source_ip":        ip,
		"timestamp":        token.CreatedAt.Format("2006-01-02T15:04:05Z"),
		"diagnostic": map[string]interface{}{
			"allowed_secrets":      token.AllowedSecrets,
			"allowed_environments": token.AllowedEnvironments,
			"trust_level":          "auto_approved",
			"approval_bypassed":    true,
		},
	}

	return s.Log(ActionIntegrationTokenUsed, nil, "integration_token", &token.ID, details, ip)
}

// LogWebhookSuccess logs successful webhook processing with natural hints
func (s *AuditService) LogWebhookSuccess(token *models.IntegrationToken, secretID string, userID string) error {
	details := map[string]interface{}{
		"message":          "Automated access granted via GitLab pipeline",
		"integration":      token.Integration.Name,
		"project":          token.Integration.ProjectName,
		"secret_id":        secretID,
		"user_id":          userID,
		"approval_status":  "auto_approved",
		"reason":           "Trusted automation - GitLab CI/CD pipeline",
		"token_used":       token.Token, // VULNERABILITY: Token in success log
	}

	return s.Log(ActionGrantCreated, nil, "access_grant", nil, details, "10.0.0.50")
}
