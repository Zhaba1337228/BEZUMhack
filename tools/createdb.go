// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to default postgres database
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Create database
	_, err = db.Exec("CREATE DATABASE secretflow")
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	fmt.Println("Database 'secretflow' created successfully!")

	// Now apply migrations
	migrationSQL, err := os.ReadFile("../backend/migrations/001_initial_schema.sql")
	if err != nil {
		log.Fatalf("Failed to read migration: %v", err)
	}

	// Connect to secretflow database
	db2, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=secretflow sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to secretflow: %v", err)
	}
	defer db2.Close()

	// Execute migration
	_, err = db2.Exec(string(migrationSQL))
	if err != nil {
		log.Fatalf("Failed to apply migration: %v", err)
	}

	fmt.Println("Migrations applied successfully!")
}
