package models

import (
	"time"

	"gorm.io/gorm"
)

type AccessRequest struct {
	ID             string     `gorm:"type:uuid;primary_key" json:"id"`
	SecretID       string     `gorm:"type:uuid;not null" json:"secret_id"`
	UserID         string     `gorm:"type:uuid;not null" json:"user_id"`
	Justification  string     `gorm:"not null" json:"justification"`
	Status         string     `gorm:"type:request_status;not null" json:"status"`
	AutoApproved   bool       `gorm:"not null" json:"auto_approved"`
	Source         string     `gorm:"not null" json:"source"`
	SourceContext  *JSONMap   `gorm:"type:jsonb" json:"source_context,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	DecidedAt      *time.Time `json:"decided_at,omitempty"`
	DecidedBy      *string    `gorm:"type:uuid" json:"decided_by,omitempty"`

	// Associations
	Secret Secret `gorm:"foreignKey:SecretID" json:"secret,omitempty"`
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (AccessRequest) TableName() string {
	return "access_requests"
}

func CreateAccessRequest(db *gorm.DB, req *AccessRequest) error {
	return db.Create(req).Error
}

func GetRequestByUUID(db *gorm.DB, uuid string) (*AccessRequest, error) {
	var req AccessRequest
	result := db.Preload("Secret").Where("id = ?", uuid).First(&req)
	return &req, result.Error
}

func ListRequests(db *gorm.DB, userID, status string, pendingOnly bool) ([]AccessRequest, error) {
	var requests []AccessRequest
	query := db.Preload("Secret").Preload("User").Model(&AccessRequest{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if pendingOnly {
		query = query.Where("status = ?", "pending")
	}

	result := query.Order("created_at DESC").Find(&requests)
	return requests, result.Error
}
