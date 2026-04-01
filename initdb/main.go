package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=localhost port=5432 user=postgres password=Rrobocopid12 dbname=postgres sslmode=disable"

	fmt.Println("Connecting to PostgreSQL...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// Check if database exists
	fmt.Println("Checking database 'secretflow'...")
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'secretflow')").Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check database: %v", err)
	}

	if exists {
		fmt.Println("Database 'secretflow' already exists, dropping...")
		// Terminate all connections to secretflow
		_, err = db.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity
			WHERE datname = 'secretflow' AND pid <> pg_backend_pid()`)
		if err != nil {
			log.Printf("Warning: Failed to terminate connections: %v", err)
		}
		_, err = db.Exec("DROP DATABASE secretflow")
		if err != nil {
			log.Fatalf("Failed to drop database: %v", err)
		}
	}

	_, err = db.Exec("CREATE DATABASE secretflow")
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	fmt.Println("Database 'secretflow' created!")

	// Read migration file
	fmt.Println("Reading migration file...")
	migrationSQL, err := os.ReadFile("D:/hackaton/BEZUMhack/backend/migrations/001_initial_schema.sql")
	if err != nil {
		log.Fatalf("Failed to read migration: %v", err)
	}

	// Connect to secretflow database
	connStr2 := "host=localhost port=5432 user=postgres password=Rrobocopid12 dbname=secretflow sslmode=disable"
	db2, err := sql.Open("postgres", connStr2)
	if err != nil {
		log.Fatalf("Failed to connect to secretflow: %v", err)
	}
	defer db2.Close()

	// Execute migration as single script with BEGIN/COMMIT
	fmt.Println("Applying migrations...")
	_, err = db2.Exec("BEGIN")
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	// Split into statements properly - by semicolon followed by newline or EOF
	statements := strings.Split(string(migrationSQL), ";\n")
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		// Add back the semicolon if not last statement
		if i < len(statements)-1 {
			stmt = stmt + ";"
		}
		_, err = db2.Exec(stmt)
		if err != nil {
			log.Printf("Error at statement %d: %v", i+1, err)
			log.Printf("Statement: %s\n", stmt[:min(len(stmt), 100)])
			db2.Exec("ROLLBACK")
			return
		}
	}

	_, err = db2.Exec("COMMIT")
	if err != nil {
		log.Fatalf("Failed to commit: %v", err)
	}

	fmt.Println("\n========================================")
	fmt.Println("Database initialized successfully!")
	fmt.Println("========================================")
	fmt.Println("\nConnection details:")
	fmt.Println("  Host: localhost:5432")
	fmt.Println("  Database: secretflow")
	fmt.Println("  User: postgres")
	fmt.Println("  Password: Rrobocopid12")
	fmt.Println("\nSeed users:")
	fmt.Println("  - dev.alice (developer, password: password123)")
	fmt.Println("  - dev.bob (developer, strong random password)")
	fmt.Println("  - lead.carol (team_lead, strong random password)")
	fmt.Println("  - security.dave (security_admin, strong random password)")
	fmt.Println("  - svc.gitlab (service_account, strong random password)")
	fmt.Println("\nSecrets:")
	fmt.Println("  - PROD_DB_MASTER_PASSWORD (CRITICAL)")
	fmt.Println("  - AWS_ROOT_ACCESS_KEY (CRITICAL)")
	fmt.Println("========================================")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
