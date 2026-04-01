package config

import (
	"os"
	"strings"
)

type Config struct {
	DBHost       string
	DBPort       string
	DBName       string
	DBUser       string
	DBPassword   string
	JWTSecret    string
	JWTExpiry    int
	BackendHost  string
	BackendPort  string
	FrontendURL  string
	AllowedOrigins []string
}

func Load() *Config {
	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "secretflow"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "Rrobocopid12"),
		JWTSecret:  getEnv("JWT_SECRET", "change_me_in_production"),
		JWTExpiry:  24,
		BackendHost: getEnv("BACKEND_HOST", "0.0.0.0"),
		BackendPort: getEnv("BACKEND_PORT", "8080"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
	}

	// Parse allowed origins
	origins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	cfg.AllowedOrigins = strings.Split(origins, ",")

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
