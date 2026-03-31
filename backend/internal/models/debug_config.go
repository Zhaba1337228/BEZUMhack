package models

import (
	"gorm.io/gorm"
)

type DebugConfig struct {
	ID          string `gorm:"type:uuid;primary_key" json:"id"`
	Key         string `gorm:"unique;not null" json:"key"`
	Value       string `gorm:"not null" json:"value"`
	Sensitive   bool   `gorm:"not null" json:"sensitive"`
	InternalOnly bool  `gorm:"not null" json:"internal_only"`
}

func (DebugConfig) TableName() string {
	return "debug_config"
}

func GetAllDebugConfig(db *gorm.DB) ([]DebugConfig, error) {
	var configs []DebugConfig
	result := db.Find(&configs)
	return configs, result.Error
}

func GetDebugConfigByKey(db *gorm.DB, key string) (*DebugConfig, error) {
	var config DebugConfig
	result := db.Where("key = ?", key).First(&config)
	return &config, result.Error
}
