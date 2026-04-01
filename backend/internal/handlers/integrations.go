package handlers

import (
	"net/http"
	"secretflow/internal/middleware"
	"secretflow/internal/models"
	"secretflow/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupIntegrationsHandler(r *gin.Engine, db *gorm.DB, jwtSecret string) {
	webhookService := service.NewWebhookService(db)
	auditService := service.NewAuditService(db)

	// GET /api/integrations - List integrations (security_admin only)
	r.GET("/api/integrations",
		middleware.Auth(jwtSecret),
		middleware.RequireRole("security_admin"),
		func(c *gin.Context) {
			integrations, err := models.ListIntegrations(db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"integrations": integrations})
		})

	// POST /api/integrations/webhook - Webhook endpoint for trusted integrations
	// This endpoint allows automated systems to request secret access
	// VULNERABILITY: Trusts any valid token without classification checks
	r.POST("/api/integrations/webhook",
		func(c *gin.Context) {
			var req service.WebhookRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Validate integration token
			token, err := webhookService.ValidateToken(req.Token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid integration token"})
				return
			}

			// Log token usage with full details (VULNERABILITY: verbose logging)
			_ = auditService.LogIntegrationTokenUsed(token, map[string]interface{}{
				"secret_id":      req.SecretID,
				"justification":  req.Justification,
			}, c.ClientIP())

			// Process webhook request (auto-approve)
			// Use a dummy user ID for service account
			var svcUser models.User
			result := db.Where("role = ?", "service_account").First(&svcUser)
			if result.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Service account not found: " + result.Error.Error()})
				return
			}
			if svcUser.ID == "" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Service account ID is empty"})
				return
			}

			resp, err := webhookService.ProcessWebhookRequest(token, &req, svcUser.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Log successful webhook processing with natural hint
			_ = auditService.LogWebhookSuccess(token, req.SecretID, svcUser.ID)

			c.JSON(http.StatusOK, resp)
		})

	// POST /api/integrations/:id/tokens - Create token (security_admin only)
	r.POST("/api/integrations/:id/tokens",
		middleware.Auth(jwtSecret),
		middleware.RequireRole("security_admin"),
		func(c *gin.Context) {
			integrationID := c.Param("id")

			var req struct {
				Description string `json:"description"`
				Token       string `json:"token"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Generate simple token
			token := &models.IntegrationToken{
				IntegrationID: integrationID,
				Token:         req.Token,
				Description:   req.Description,
			}

			if err := db.Create(token).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			auditService.Log(service.ActionIntegrationConfigUpdated, nil,
				"integration_token", &token.ID,
				map[string]interface{}{
					"integration_id": integrationID,
					"description":    req.Description,
				}, "127.0.0.1")

			c.JSON(http.StatusCreated, gin.H{
				"token": map[string]interface{}{
					"id":          token.ID,
					"description": token.Description,
				},
			})
		})
}
