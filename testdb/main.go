package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=Rrobocopid12 dbname=secretflow sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	fmt.Println("Users count:", count)

	var svcCount int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'service_account'").Scan(&svcCount)
	fmt.Println("Service accounts:", svcCount)

	var svcID string
	err = db.QueryRow("SELECT id FROM users WHERE role = 'service_account' LIMIT 1").Scan(&svcID)
	if err != nil {
		log.Fatalf("Failed to get service account: %v", err)
	}
	fmt.Println("Service account ID:", svcID)
}
