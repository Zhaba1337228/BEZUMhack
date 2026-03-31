package models

import (
	"time"

	"gorm.io/gorm"
)

type AuditLog struct {
	ID           string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Timestamp    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"timestamp"`
	UserID       *string   `gorm:"type:uuid" json:"user_id,omitempty"`
	Action       string    `gorm:"not null" json:"action"`
	ResourceType string    `json:"resource_type,omitempty"`
	ResourceID   *string   `gorm:"type:uuid" json:"resource_id,omitempty"`
	Details      *JSONMap  `gorm:"type:jsonb" json:"details,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

func CreateAuditLog(db *gorm.DB, log *AuditLog) error {
	return db.Create(log).Error
}

func ListAuditLogs(db *gorm.DB, userID, action string, limit int) ([]AuditLog, error) {
	var logs []AuditLog
	query := db.Model(&AuditLog{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}

	result := query.Order("timestamp DESC").Limit(limit).Find(&logs)
	return logs, result.Error
}
