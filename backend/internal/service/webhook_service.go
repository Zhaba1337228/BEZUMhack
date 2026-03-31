package service

import (
	"errors"
	"time"

	"secretflow/internal/models"

	"gorm.io/gorm"
)

var (
	ErrInvalidToken      = errors.New("invalid integration token")
	ErrIntegrationDisabled = errors.New("integration is disabled")
	ErrTokenExpired      = errors.New("token has expired")
)

type WebhookService struct {
	db *gorm.DB
}

func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{db: db}
}

type WebhookRequest struct {
	Token         string `json:"token" binding:"required"`
	SecretID      string `json:"secret_id" binding:"required"`
	Justification string `json:"justification" binding:"required"`
}

type WebhookResponse struct {
	AccessGranted bool       `json:"access_granted"`
	SecretValue   string     `json:"secret_value,omitempty"`
	GrantExpiresAt time.Time `json:"grant_expires_at"`
	RequestID     string     `json:"request_id"`
}

// ValidateToken validates an integration token
// VULNERABILITY: Does not check allowed_secrets or allowed_environments
func (s *WebhookService) ValidateToken(token string) (*models.IntegrationToken, error) {
	intToken, err := models.GetIntegrationByToken(s.db, token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if !intToken.Integration.Enabled {
		return nil, ErrIntegrationDisabled
	}

	// VULNERABILITY: These fields exist but are never checked
	// In a real system, this would restrict which secrets/environments the token can access
	// if intToken.AllowedSecrets != nil {
	//     check if secretID is in allowed list
	// }
	// if intToken.AllowedEnvironments != nil {
	//     check if environment is in allowed list
	// }

	// Update last used timestamp
	_ = models.UpdateTokenLastUsed(s.db, intToken.ID)

	return intToken, nil
}

// ProcessWebhookRequest processes a webhook request with valid token
// This is the core vulnerability: trusted automation bypasses classification-based approval
func (s *WebhookService) ProcessWebhookRequest(token *models.IntegrationToken, req *WebhookRequest, userID string) (*WebhookResponse, error) {
	// Create access request with auto_approved = true
	// This bypasses the normal approval workflow entirely
	now := time.Now()
	accessReq := &models.AccessRequest{
		SecretID:      req.SecretID,
		UserID:        userID,
		Justification: req.Justification,
		Status:        "approved",
		AutoApproved:  true,
		Source:        "webhook",
		SourceContext: &models.JSONMap{
			"integration_id": token.IntegrationID,
			"token_id":       token.ID,
			"integration_name": token.Integration.Name,
			"project":        token.Integration.ProjectName,
			"pipeline":       "gitlab-ci",
			"trigger":        "deployment",
		},
		CreatedAt:   now,
		DecidedAt:   &now,
	}

	if err := s.db.Create(accessReq).Error; err != nil {
		return nil, err
	}

	// Create access grant immediately (bypasses normal approval)
	// This is the critical vulnerability: CRITICAL secrets should still require approval
	grant := &models.AccessGrant{
		RequestID: accessReq.ID,
		SecretID:  req.SecretID,
		UserID:    userID,
		GrantedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
		Revoked:   false,
	}

	if err := s.db.Create(grant).Error; err != nil {
		return nil, err
	}

	// Get secret value
	var secret models.Secret
	if err := s.db.Where("id = ?", req.SecretID).First(&secret).Error; err != nil {
		return nil, err
	}

	return &WebhookResponse{
		AccessGranted:  true,
		SecretValue:    secret.Value,
		GrantExpiresAt: grant.ExpiresAt,
		RequestID:      accessReq.ID,
	}, nil
}
