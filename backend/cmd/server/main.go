package main

import (
	"fmt"
	"log"
	"os"
	"secretflow/internal/config"
	"secretflow/internal/database"
	"secretflow/internal/handlers"
)

func init() {
	// Enable verbose logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	// Load configuration
	cfg := config.Load()

	log.Println("Starting SecretFlow backend...")
	log.Printf("Database: %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

	// Initialize database
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup router
	r := handlers.SetupRouter(db, handlers.RouterConfig{
		JWTSecret:      cfg.JWTSecret,
		JWTExpiry:      cfg.JWTExpiry,
		FrontendURL:    cfg.FrontendURL,
		AllowedOrigins: cfg.AllowedOrigins,
	})

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.BackendHost, cfg.BackendPort)
	log.Printf("Server starting on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
