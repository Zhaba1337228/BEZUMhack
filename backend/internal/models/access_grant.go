package models

import (
	"time"

	"gorm.io/gorm"
)

type AccessGrant struct {
	ID         string    `gorm:"type:uuid;primary_key" json:"id"`
	RequestID  string    `gorm:"type:uuid;not null" json:"request_id"`
	SecretID   string    `gorm:"type:uuid;not null" json:"secret_id"`
	UserID     string    `gorm:"type:uuid;not null" json:"user_id"`
	GrantedAt  time.Time `gorm:"not null" json:"granted_at"`
	ExpiresAt  time.Time `gorm:"not null" json:"expires_at"`
	Revoked    bool      `gorm:"not null" json:"revoked"`

	// Associations
	Secret Secret `gorm:"foreignKey:SecretID" json:"secret,omitempty"`
}

func (AccessGrant) TableName() string {
	return "access_grants"
}

func CreateAccessGrant(db *gorm.DB, grant *AccessGrant) error {
	return db.Create(grant).Error
}

func GetActiveGrant(db *gorm.DB, userID, secretID string) (*AccessGrant, error) {
	var grant AccessGrant
	result := db.Where("user_id = ? AND secret_id = ? AND revoked = ? AND expires_at > ?",
		userID, secretID, false, time.Now()).First(&grant)
	return &grant, result.Error
}

func RevokeGrant(db *gorm.DB, grantID string) error {
	return db.Model(&AccessGrant{}).Where("id = ?", grantID).Update("revoked", true).Error
}
