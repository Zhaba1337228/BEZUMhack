package service

import (
	"errors"
	"time"

	"secretflow/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrAlreadyDecided     = errors.New("request already decided")
	ErrInsufficientRole   = errors.New("insufficient role for this approval")
	ErrInvalidRequest     = errors.New("invalid request")
)

type ApprovalService struct {
	db *gorm.DB
}

func NewApprovalService(db *gorm.DB) *ApprovalService {
	return &ApprovalService{db: db}
}

// GetRequiredApproverRole returns the role required to approve a request based on classification
func GetRequiredApproverRole(classification string) string {
	switch classification {
	case "LOW":
		return "" // No approval needed
	case "MEDIUM":
		return "team_lead" // Any team_lead or security_admin
	case "HIGH":
		return "team_lead" // Team lead for owner team
	case "CRITICAL":
		return "security_admin" // Only security_admin
	default:
		return "team_lead"
	}
}

// CanApprove checks if a user can approve a given request
func (s *ApprovalService) CanApprove(user *models.User, req *models.AccessRequest, secret *models.Secret) bool {
	if user.Role == "security_admin" {
		return true // Security admin can approve anything
	}

	if user.Role != "team_lead" {
		return false // Only team_lead and security_admin can approve
	}

	// For LOW: anyone can approve (no approval needed anyway)
	if secret.Classification == "LOW" {
		return true
	}

	// For MEDIUM: any team_lead can approve
	if secret.Classification == "MEDIUM" {
		return true
	}

	// For HIGH: team_lead must be from same team as secret owner
	if secret.Classification == "HIGH" {
		return user.Team == secret.OwnerTeam
	}

	// CRITICAL: only security_admin (already handled above)
	return false
}

// Approve approves an access request
func (s *ApprovalService) Approve(reqID string, approver *models.User) (*models.AccessGrant, error) {
	// Get request with secret details
	var req models.AccessRequest
	if err := s.db.Preload("Secret").Where("id = ?", reqID).First(&req).Error; err != nil {
		return nil, ErrInvalidRequest
	}

	if req.Status != "pending" {
		return nil, ErrAlreadyDecided
	}

	// Check if auto-approved (trusted flow)
	if req.AutoApproved {
		// Auto-approved requests are already granted
		var grant models.AccessGrant
		if err := s.db.Where("request_id = ?", reqID).First(&grant).Error; err != nil {
			return nil, err
		}
		return &grant, nil
	}

	// Get secret for classification check
	var secret models.Secret
	if err := s.db.Where("id = ?", req.SecretID).First(&secret).Error; err != nil {
		return nil, ErrInvalidRequest
	}

	// Check if approver has permission
	if !s.CanApprove(approver, &req, &secret) {
		return nil, ErrInsufficientRole
	}

	// Update request status
	now := time.Now()
	req.Status = "approved"
	req.DecidedAt = &now
	req.DecidedBy = &approver.ID
	if err := s.db.Save(&req).Error; err != nil {
		return nil, err
	}

	// Create access grant (24 hour expiry)
	grant := &models.AccessGrant{
		ID:        uuid.New().String(),
		RequestID: req.ID,
		SecretID:  req.SecretID,
		UserID:    req.UserID,
		GrantedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
		Revoked:   false,
	}

	if err := s.db.Create(grant).Error; err != nil {
		return nil, err
	}

	return grant, nil
}

// Deny denies an access request
func (s *ApprovalService) Deny(reqID string, denier *models.User) error {
	var req models.AccessRequest
	if err := s.db.Where("id = ?", reqID).First(&req).Error; err != nil {
		return ErrInvalidRequest
	}

	if req.Status != "pending" {
		return ErrAlreadyDecided
	}

	// Check if denier has permission (same as approve)
	var secret models.Secret
	if err := s.db.Where("id = ?", req.SecretID).First(&secret).Error; err != nil {
		return ErrInvalidRequest
	}

	if !s.CanApprove(denier, &req, &secret) {
		return ErrInsufficientRole
	}

	// Update request status
	now := time.Now()
	req.Status = "denied"
	req.DecidedAt = &now
	req.DecidedBy = &denier.ID

	return s.db.Save(&req).Error
}

// CreateAutoApprovedGrant creates a grant without approval (for trusted automation)
func (s *ApprovalService) CreateAutoApprovedGrant(secretID, userID, requestID string) (*models.AccessGrant, error) {
	now := time.Now()

	// Create access grant directly
	grant := &models.AccessGrant{
		ID:        uuid.New().String(),
		RequestID: requestID,
		SecretID:  secretID,
		UserID:    userID,
		GrantedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
		Revoked:   false,
	}

	if err := s.db.Create(grant).Error; err != nil {
		return nil, err
	}

	return grant, nil
}
