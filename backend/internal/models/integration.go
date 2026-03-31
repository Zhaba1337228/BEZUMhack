package models

import (
	"time"

	"gorm.io/gorm"
)

type Integration struct {
	ID          string    `gorm:"type:uuid;primary_key" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Provider    string    `gorm:"type:integration_provider;not null" json:"provider"`
	ProjectName string    `json:"project_name"`
	Enabled     bool      `gorm:"not null" json:"enabled"`
	Config      *JSONMap  `gorm:"type:jsonb" json:"config,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Integration) TableName() string {
	return "integrations"
}

type IntegrationToken struct {
	ID                  string    `gorm:"type:uuid;primary_key" json:"id"`
	IntegrationID       string    `gorm:"type:uuid;not null" json:"integration_id"`
	Token               string    `gorm:"unique;not null" json:"-"` // Never expose in JSON
	Description         string    `json:"description"`
	AllowedSecrets      *[]string `gorm:"type:text[]" json:"allowed_secrets,omitempty"`
	AllowedEnvironments *[]string `gorm:"type:secret_environment[]" json:"allowed_environments,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	LastUsedAt          *time.Time `json:"last_used_at"`

	// Associations
	Integration Integration `gorm:"foreignKey:IntegrationID" json:"integration,omitempty"`
}

func (IntegrationToken) TableName() string {
	return "integration_tokens"
}

func GetIntegrationByToken(db *gorm.DB, token string) (*IntegrationToken, error) {
	var t IntegrationToken
	result := db.Preload("Integration").Where("token = ?", token).First(&t)
	return &t, result.Error
}

func UpdateTokenLastUsed(db *gorm.DB, tokenID string) error {
	now := time.Now()
	return db.Model(&IntegrationToken{}).Where("id = ?", tokenID).Update("last_used_at", &now).Error
}

func ListIntegrations(db *gorm.DB) ([]Integration, error) {
	var integrations []Integration
	result := db.Find(&integrations)
	return integrations, result.Error
}
