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

	fmt.Println("\n=== Tables in secretflow database ===\n")

	rows, err := db.Query(`
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name`)
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)
		fmt.Printf("  - %s\n", tableName)
	}

	fmt.Println("\n=== Users ===\n")
	rows2, _ := db.Query("SELECT username, role, team FROM users")
	defer rows2.Close()
	for rows2.Next() {
		var username, role string
		var team sql.NullString
		rows2.Scan(&username, &role, &team)
		teamStr := ""
		if team.Valid {
			teamStr = team.String
		}
		fmt.Printf("  %s | %s | %s\n", username, role, teamStr)
	}

	fmt.Println("\n=== Secrets ===\n")
	rows3, _ := db.Query("SELECT name, classification, environment FROM secrets ORDER BY classification, name")
	defer rows3.Close()
	for rows3.Next() {
		var name, classification, environment string
		rows3.Scan(&name, &classification, &environment)
		fmt.Printf("  %s | %s | %s\n", name, classification, environment)
	}

	fmt.Println("\n=== Integrations ===\n")
	rows4, _ := db.Query("SELECT name, provider, project_name FROM integrations")
	defer rows4.Close()
	for rows4.Next() {
		var name, provider, project sql.NullString
		rows4.Scan(&name, &provider, &project)
		projectStr := ""
		if project.Valid {
			projectStr = project.String
		}
		fmt.Printf("  %s | %s | %s\n", name.String, provider.String, projectStr)
	}

	fmt.Println("\n========================================")
}
