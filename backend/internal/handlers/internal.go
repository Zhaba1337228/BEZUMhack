package handlers

import (
	"net/http"
	"secretflow/internal/middleware"
	"secretflow/internal/models"
	"secretflow/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupInternalHandler(r *gin.Engine, db *gorm.DB, jwtSecret string) {
	auditService := service.NewAuditService(db)

	// TEST endpoint - create access request via raw SQL
	r.POST("/api/internal/test/grant", func(c *gin.Context) {
		var req struct {
			SecretID string `json:"secret_id"`
			UserID   string `json:"user_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create access request via raw SQL
		var reqID string
		err := db.Raw(`INSERT INTO access_requests (id, secret_id, user_id, justification, status, auto_approved, source, created_at, decided_at)
			VALUES (gen_random_uuid(), ?, ?, 'Test', 'approved', true, 'test', NOW(), NOW()) RETURNING id`,
			req.SecretID, req.UserID).Scan(&reqID).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "SQL Error: " + err.Error()})
			return
		}

		// Create grant
		var grantID string
		err = db.Raw(`INSERT INTO access_grants (id, request_id, secret_id, user_id, granted_at, expires_at, revoked)
			VALUES (gen_random_uuid(), ?, ?, ?, NOW(), NOW() + INTERVAL '24 hours', false) RETURNING id`,
			reqID, req.SecretID, req.UserID).Scan(&grantID).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Grant SQL Error: " + err.Error()})
			return
		}

		// Get secret value
		var secretValue string
		err = db.Raw("SELECT value FROM secrets WHERE id = ?", req.SecretID).Scan(&secretValue).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Secret SQL Error: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"request_id":    reqID,
			"grant_id":      grantID,
			"secret_value":  secretValue,
		})
	})

	// GET /api/internal/integrations/status - Integration health/status endpoint
	// PATH 1 VULNERABILITY: Leaks auth_token for "debugging connectivity"
	r.GET("/api/internal/integrations/status",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			integrations, err := models.ListIntegrations(db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			type IntegrationStatus struct {
				ID          string          `json:"id"`
				Name        string          `json:"name"`
				Provider    string          `json:"provider"`
				Enabled     bool            `json:"enabled"`
				ProjectName string          `json:"project_name"`
				Status      string          `json:"status"`
				LastSync    *string         `json:"last_sync,omitempty"`
				AuthToken   *string         `json:"auth_token,omitempty"` // PATH 1: Leaks token
				WebhookURL  *string         `json:"webhook_url,omitempty"`
				Config      *models.JSONMap `json:"config,omitempty"`
			}

			var statuses []IntegrationStatus
			for _, integration := range integrations {
				var tokens []models.IntegrationToken
				db.Where("integration_id = ?", integration.ID).Find(&tokens)

				var authToken *string
				var lastSync *string
				if len(tokens) > 0 {
					authToken = &tokens[0].Token // PATH 1: Token leakage
					if tokens[0].LastUsedAt != nil {
						s := tokens[0].LastUsedAt.Format("2006-01-02T15:04:05Z")
						lastSync = &s
					}
				}

				status := "connected"
				if !integration.Enabled {
					status = "disabled"
				}

				var webhookURL *string
				if integration.Config != nil {
					if rawURL, ok := (*integration.Config)["webhook_url"].(string); ok && rawURL != "" {
						webhookURL = &rawURL
					}
				}

				statuses = append(statuses, IntegrationStatus{
					ID:          integration.ID,
					Name:        integration.Name,
					Provider:    integration.Provider,
					Enabled:     integration.Enabled,
					ProjectName: integration.ProjectName,
					Status:      status,
					LastSync:    lastSync,
					AuthToken:   authToken,
					WebhookURL:  webhookURL,
					Config:      integration.Config,
				})
			}

			userID, _ := c.Get("userID")
			userIDStr := userID.(string)
			auditService.Log(service.ActionInternalAPICall, &userIDStr,
				"integration_status", nil,
				map[string]interface{}{
					"integrations_count": len(statuses),
					"purpose":            "operational_health_check",
				}, c.ClientIP())

			c.JSON(http.StatusOK, gin.H{"integrations": statuses})
		})

	// GET /api/internal/integrations/test - Test integration connectivity
	// PATH 1 VULNERABILITY: Leaks full token for "diagnostic purposes"
	r.GET("/api/internal/integrations/test/:id",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			integrationID := c.Param("id")

			var integration models.Integration
			if err := db.Where("id = ?", integrationID).First(&integration).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
				return
			}

			var token models.IntegrationToken
			if err := db.Where("integration_id = ?", integrationID).First(&token).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "No token configured"})
				return
			}

			testResult := map[string]interface{}{
				"integration_id":   integration.ID,
				"integration_name": integration.Name,
				"provider":         integration.Provider,
				"project":          integration.ProjectName,
				"token_valid":      true,
				"token_preview":    token.Token[:8] + "..." + token.Token[len(token.Token)-4:],
				"token_full":       token.Token, // PATH 1: Full token leak
				"connection_test":  "success",
				"last_used":        token.LastUsedAt,
			}

			c.JSON(http.StatusOK, gin.H{"test_result": testResult})
		})

	// POST /api/internal/secrets/grant - DISABLED for CTF (was Path 3)
	r.POST("/api/internal/secrets/grant",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint disabled"})
		})

	// POST /api/internal/apply - DISABLED for CTF (was alternative Path 3)
	r.POST("/api/internal/apply",
		func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint disabled"})
		})
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
