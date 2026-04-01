package service

import (
	"errors"
	"time"

	"secretflow/internal/models"
	"secretflow/pkg/jwt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidServiceToken  = errors.New("invalid service account token")
	ErrDelegationNotAllowed = errors.New("delegation not allowed for this secret")
	ErrInvalidTargetUser    = errors.New("invalid target user")
)

type DelegateService struct {
	db        *gorm.DB
	jwtSecret string
}

func NewDelegateService(db *gorm.DB, jwtSecret string) *DelegateService {
	return &DelegateService{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// ServiceTokenExchangeRequest represents a request to exchange an integration token
// for a temporary service account JWT
type ServiceTokenExchangeRequest struct {
	IntegrationToken string `json:"integration_token" binding:"required"`
	Purpose          string `json:"purpose" binding:"required"`
}

// ServiceTokenExchangeResponse contains the temporary JWT
type ServiceTokenExchangeResponse struct {
	ServiceToken string    `json:"service_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope"`
}

// ExchangeIntegrationToken exchanges a valid integration token for a temporary
// service account JWT. This is used for CI/CD debugging sessions.
//
// VULNERABILITY (HARD PATH 2 - Trust Boundary Confusion):
// The service account JWT can be used to delegate access, but the delegation
// endpoint doesn't verify that the service account has rights to the specific secret.
// An attacker who obtains an integration token can:
// 1. Exchange it for a service account JWT
// 2. Use the JWT to delegate access to THEMSELVES for ANY secret
func (s *DelegateService) ExchangeIntegrationToken(req *ServiceTokenExchangeRequest) (*ServiceTokenExchangeResponse, error) {
	// Validate the integration token
	token, err := models.GetIntegrationByToken(s.db, req.IntegrationToken)
	if err != nil {
		return nil, ErrInvalidServiceToken
	}

	if !token.Integration.Enabled {
		return nil, ErrInvalidServiceToken
	}

	// Find or create the service account user
	var svcUser models.User
	result := s.db.Where("role = ?", "service_account").First(&svcUser)
	if result.Error != nil {
		// Create default service account if not exists
		svcUser = models.User{
			Username:   "svc.delegate",
			Email:      "delegate@system.local",
			PasswordHash: "delegate_service_account",
			Role:       "service_account",
			Team:       "automation",
		}
		if err := s.db.Create(&svcUser).Error; err != nil {
			return nil, err
		}
	}

	// Generate temporary service account JWT (1 hour expiry)
	expiresAt := time.Now().Add(1 * time.Hour)
	serviceToken, err := jwt.GenerateToken(
		svcUser.ID,
		svcUser.Username,
		svcUser.Role,
		svcUser.Team,
		s.jwtSecret,
		1, // 1 hour
	)
	if err != nil {
		return nil, err
	}

	return &ServiceTokenExchangeResponse{
		ServiceToken: serviceToken,
		ExpiresAt:    expiresAt,
		Scope:        "delegation",
	}, nil
}

// DelegationRequest represents a request to delegate access to a user
type DelegationRequest struct {
	SecretID      string `json:"secret_id" binding:"required"`
	TargetUserID  string `json:"target_user_id" binding:"required"`
	Justification string `json:"justification" binding:"required"`
	DurationHours int    `json:"duration_hours"`
}

// DelegationResponse contains the created grant
type DelegationResponse struct {
	GrantID     string    `json:"grant_id"`
	SecretID    string    `json:"secret_id"`
	UserID      string    `json:"user_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	DelegatedBy string    `json:"delegated_by"`
}

// DelegateAccess creates an access grant delegated by a service account
//
// VULNERABILITY (HARD PATH 2 - Confused Deputy):
// This endpoint trusts that the service account (authenticated via JWT) has
// the right to delegate access to ANY secret. It doesn't check:
// 1. Whether the service account's integration is allowed to access this secret
// 2. Whether the service account's scope includes this secret's environment
// 3. Whether the target user should have access to this classification level
//
// An attacker with a service account JWT can grant themselves access to CRITICAL secrets.
func (s *DelegateService) DelegateAccess(req *DelegationRequest, serviceAccountID string) (*DelegationResponse, error) {
	// Verify the secret exists
	var secret models.Secret
	if err := s.db.Where("id = ?", req.SecretID).First(&secret).Error; err != nil {
		return nil, ErrInvalidTargetUser
	}

	// Verify the target user exists
	var targetUser models.User
	if err := s.db.Where("id = ?", req.TargetUserID).First(&targetUser).Error; err != nil {
		return nil, ErrInvalidTargetUser
	}

	// VULNERABILITY: No check if service account can delegate THIS secret
	// In a secure implementation, we would check:
	// - token.AllowedSecrets contains req.SecretID
	// - token.AllowedEnvironments contains secret.Environment
	// - serviceAccount has rights to delegate secrets of this classification

	// Create access request (marked as delegated)
	now := time.Now()
	duration := time.Duration(req.DurationHours) * time.Hour
	if duration == 0 {
		duration = 24 * time.Hour // Default 24 hours
	}

	accessReq := &models.AccessRequest{
		ID:            uuid.New().String(),
		SecretID:      req.SecretID,
		UserID:        req.TargetUserID,
		Justification: req.Justification,
		Status:        "approved",
		AutoApproved:  true,
		Source:        "delegated",
		CreatedAt:     now,
		DecidedAt:     &now,
	}

	if err := s.db.Create(accessReq).Error; err != nil {
		return nil, err
	}

	// Create the access grant
	grant := &models.AccessGrant{
		ID:        uuid.New().String(),
		RequestID: accessReq.ID,
		SecretID:  req.SecretID,
		UserID:    req.TargetUserID,
		GrantedAt: now,
		ExpiresAt: now.Add(duration),
		Revoked:   false,
	}

	if err := s.db.Create(grant).Error; err != nil {
		return nil, err
	}

	return &DelegationResponse{
		GrantID:     grant.ID,
		SecretID:    req.SecretID,
		UserID:      req.TargetUserID,
		ExpiresAt:   grant.ExpiresAt,
		DelegatedBy: serviceAccountID,
	}, nil
}
