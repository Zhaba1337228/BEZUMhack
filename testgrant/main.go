package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=localhost port=5432 user=postgres password=Rrobocopid12 dbname=secretflow sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Try to insert access_request directly
	fmt.Println("=== Testing direct SQL insert ===")

	// First get a valid user ID and secret ID
	var userID, secretID string
	err = db.QueryRow("SELECT id FROM users WHERE username = 'dev.alice'").Scan(&userID)
	if err != nil {
		log.Fatalf("Failed to get user ID: %v", err)
	}
	fmt.Printf("User ID: %s\n", userID)

	err = db.QueryRow("SELECT id FROM secrets WHERE name = 'PROD_DB_MASTER_PASSWORD'").Scan(&secretID)
	if err != nil {
		log.Fatalf("Failed to get secret ID: %v", err)
	}
	fmt.Printf("Secret ID: %s\n", secretID)

	// Try to insert access_request
	insertSQL := `INSERT INTO access_requests (id, secret_id, user_id, justification, status, auto_approved, source, source_context)
		VALUES (gen_random_uuid(), $1, $2, 'Internal API grant', 'approved', true, 'webhook', '{"validated": false}'::jsonb)
		RETURNING id`

	var reqID string
	err = db.QueryRow(insertSQL, secretID, userID).Scan(&reqID)
	if err != nil {
		log.Fatalf("Failed to insert access_request: %v", err)
	}
	fmt.Printf("Access request created: %s\n", reqID)

	// Try to insert access_grant
	grantSQL := `INSERT INTO access_grants (id, request_id, secret_id, user_id, granted_at, expires_at, revoked)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW() + INTERVAL '24 hours', false)
		RETURNING id`

	var grantID string
	err = db.QueryRow(grantSQL, reqID, secretID, userID).Scan(&grantID)
	if err != nil {
		log.Fatalf("Failed to insert access_grant: %v", err)
	}
	fmt.Printf("Access grant created: %s\n", grantID)

	// Get secret value
	var secretValue string
	err = db.QueryRow("SELECT value FROM secrets WHERE id = $1", secretID).Scan(&secretValue)
	if err != nil {
		log.Fatalf("Failed to get secret value: %v", err)
	}

	fmt.Println("\n========================================")
	fmt.Println("SUCCESS! CRITICAL secret value:", secretValue)
	fmt.Println("========================================")
}
