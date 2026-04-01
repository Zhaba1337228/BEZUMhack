package service

import (
	"errors"
	"time"

	"secretflow/internal/models"

	"gorm.io/gorm"
)

var (
	ErrInvalidToken       = errors.New("invalid integration token")
	ErrIntegrationDisabled = errors.New("integration is disabled")
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
// PATH 1 & 2 VULNERABILITY: Does not check allowed_secrets or allowed_environments
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

	// Update last used timestamp
	_ = models.UpdateTokenLastUsed(s.db, intToken.ID)

	return intToken, nil
}

// ProcessWebhookRequest processes a webhook request with valid token
// This is the core vulnerability: trusted automation bypasses classification-based approval
func (s *WebhookService) ProcessWebhookRequest(token *models.IntegrationToken, req *WebhookRequest, userID string) (*WebhookResponse, error) {
	now := time.Now()
	accessReq := &models.AccessRequest{
		SecretID:      req.SecretID,
		UserID:        userID,
		Justification: req.Justification,
		Status:        "approved",
		AutoApproved:  true,
		Source:        "webhook",
		CreatedAt:     now,
		DecidedAt:     &now,
	}

	if err := s.db.Create(accessReq).Error; err != nil {
		return nil, err
	}

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
