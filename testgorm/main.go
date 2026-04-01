package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type AccessRequest struct {
	ID            string         `gorm:"type:uuid;primary_key"`
	SecretID      string         `gorm:"type:uuid;not null"`
	UserID        string         `gorm:"type:uuid;not null"`
	Justification string         `gorm:"not null"`
	Status        string         `gorm:"type:request_status;not null"`
	AutoApproved  bool           `gorm:"not null"`
	Source        string         `gorm:"not null"`
	SourceContext *string        `gorm:"type:jsonb"`
	CreatedAt     time.Time      `json:"created_at"`
	DecidedAt     *time.Time     `json:"decided_at,omitempty"`
	DecidedBy     *string        `gorm:"type:uuid"`
}

func (AccessRequest) TableName() string {
	return "access_requests"
}

func main() {
	dsn := "host=localhost port=5432 user=postgres password=Rrobocopid12 dbname=secretflow sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Get valid IDs
	var userID, secretID string
	db.Raw("SELECT id FROM users WHERE username = 'dev.alice'").Scan(&userID)
	db.Raw("SELECT id FROM secrets WHERE name = 'PROD_DB_MASTER_PASSWORD'").Scan(&secretID)

	fmt.Printf("User ID: %s, Secret ID: %s\n", userID, secretID)

	// Create access request with GORM
	reqID := uuid.New().String()
	now := time.Now()
	accessReq := AccessRequest{
		ID:            reqID,
		SecretID:      secretID,
		UserID:        userID,
		Justification: "Test via GORM",
		Status:        "approved",
		AutoApproved:  true,
		Source:        "webhook",
		SourceContext: nil,
		CreatedAt:     now,
		DecidedAt:     &now,
	}

	result := db.Create(&accessReq)
	if result.Error != nil {
		log.Fatalf("Failed to create access request: %v", result.Error)
	}

	fmt.Printf("Access request created: %s\n", accessReq.ID)

	// Create grant
	grantID := uuid.New().String()
	type AccessGrant struct {
		ID        string    `gorm:"type:uuid;primary_key"`
		RequestID string    `gorm:"type:uuid;not null"`
		SecretID  string    `gorm:"type:uuid;not null"`
		UserID    string    `gorm:"type:uuid;not null"`
		GrantedAt time.Time `gorm:"not null"`
		ExpiresAt time.Time `gorm:"not null"`
		Revoked   bool      `gorm:"not null"`
	}

	grant := AccessGrant{
		ID:        grantID,
		RequestID: reqID,
		SecretID:  secretID,
		UserID:    userID,
		GrantedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
		Revoked:   false,
	}

	result2 := db.Create(&grant)
	if result2.Error != nil {
		log.Fatalf("Failed to create grant: %v", result2.Error)
	}

	fmt.Printf("Grant created: %s\n", grantID)

	// Get secret value
	var secretValue string
	db.Raw("SELECT value FROM secrets WHERE id = ?", secretID).Scan(&secretValue)
	fmt.Println("\n========================================")
	fmt.Println("SUCCESS! CRITICAL secret value:", secretValue)
	fmt.Println("========================================")

	// Cleanup
	db.Exec("DELETE FROM access_grants WHERE id = ?", grantID)
	db.Exec("DELETE FROM access_requests WHERE id = ?", reqID)
}
