package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RouterConfig struct {
	JWTSecret      string
	JWTExpiry      int
	FrontendURL    string
	AllowedOrigins []string
}

func SetupRouter(db *gorm.DB, cfg RouterConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// CORS configuration - allow multiple origins
	r.Use(func(c *gin.Context) {
		// Check if origin is allowed
		origin := c.Request.Header.Get("Origin")
		allowedOrigins := cfg.AllowedOrigins
		if len(allowedOrigins) == 0 {
			allowedOrigins = []string{
				cfg.FrontendURL,
				"http://localhost:5173",
				"http://localhost:3000",
				"http://127.0.0.1:5173",
				"http://127.0.0.1:3000",
			}
		}

		allowed := false
		for _, o := range allowedOrigins {
			if origin == o {
				allowed = true
				break
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if cfg.FrontendURL != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.FrontendURL)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Service-Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup all handlers
	SetupAuthHandler(r, db, cfg.JWTSecret, cfg.JWTExpiry)
	SetupSecretsHandler(r, db, cfg.JWTSecret)
	SetupRequestsHandler(r, db, cfg.JWTSecret)
	SetupAuditHandler(r, db, cfg.JWTSecret)
	SetupIntegrationsHandler(r, db, cfg.JWTSecret)
	SetupInternalHandler(r, db, cfg.JWTSecret)
	SetupDashboardHandler(r, db, cfg.JWTSecret)
	SetupDelegateHandler(r, db, cfg.JWTSecret) // NEW: Delegation endpoints (HARD Path 2)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
