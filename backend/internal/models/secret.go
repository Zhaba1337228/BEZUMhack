package models

import (
	"time"

	"gorm.io/gorm"
)

type Secret struct {
	ID            string    `gorm:"type:uuid;primary_key" json:"id"`
	Name          string    `gorm:"unique;not null" json:"name"`
	Description   string    `json:"description"`
	Classification string   `gorm:"type:secret_classification;not null" json:"classification"`
	Environment   string    `gorm:"type:secret_environment;not null" json:"environment"`
	OwnerTeam     string    `gorm:"not null" json:"owner_team"`
	Value         string    `gorm:"not null" json:"-"` // Never expose in JSON
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Secret) TableName() string {
	return "secrets"
}

func GetSecretByUUID(db *gorm.DB, uuid string) (*Secret, error) {
	var secret Secret
	result := db.Where("id = ?", uuid).First(&secret)
	return &secret, result.Error
}

func ListSecrets(db *gorm.DB, classification, environment, ownerTeam string) ([]Secret, error) {
	var secrets []Secret
	query := db.Model(&Secret{})

	if classification != "" {
		query = query.Where("classification = ?", classification)
	}
	if environment != "" {
		query = query.Where("environment = ?", environment)
	}
	if ownerTeam != "" {
		query = query.Where("owner_team = ?", ownerTeam)
	}

	result := query.Find(&secrets)
	return secrets, result.Error
}
