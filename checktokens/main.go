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

	fmt.Println("=== Integration Tokens ===")
	rows, err := db.Query(`SELECT it.id, it.integration_id, it.token, i.name, i.provider, i.enabled
		FROM integration_tokens it
		JOIN integrations i ON it.integration_id = i.id`)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, integrationID, token, name, provider string
		var enabled bool
		rows.Scan(&id, &integrationID, &token, &name, &provider, &enabled)
		fmt.Printf("  Token: %s | Integration: %s (%s) | ID: %s | Enabled: %v\n", token, name, provider, id, enabled)
	}

	fmt.Println("\n=== Integrations ===")
	rows2, err := db.Query(`SELECT id, name, provider, project_name, enabled FROM integrations`)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var id, name, provider string
		var enabled bool
		var project sql.NullString
		rows2.Scan(&id, &name, &provider, &project, &enabled)
		projectStr := ""
		if project.Valid {
			projectStr = project.String
		}
		fmt.Printf("  ID: %s | Name: %s | Provider: %s | Project: %s | Enabled: %v\n", id, name, provider, projectStr, enabled)
	}
}
