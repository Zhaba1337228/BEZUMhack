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
	approvalService := service.NewApprovalService(db)
	auditService := service.NewAuditService(db)

	// GET /api/internal/integrations/status - Integration health/status endpoint
	// VULNERABILITY: Leaks sensitive integration metadata in a plausible way
	r.GET("/api/internal/integrations/status",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			// This endpoint is meant for ops to verify integration connectivity
			// It returns detailed status including auth tokens for debugging
			integrations, err := models.ListIntegrations(db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			type IntegrationStatus struct {
				ID             string                 `json:"id"`
				Name           string                 `json:"name"`
				Provider       string                 `json:"provider"`
				Enabled        bool                   `json:"enabled"`
				ProjectName    string                 `json:"project_name"`
				Status         string                 `json:"status"`
				LastSync       *string                `json:"last_sync,omitempty"`
				AuthToken      *string                `json:"auth_token,omitempty"` // VULNERABILITY: Leaks token
				WebhookURL     *string                `json:"webhook_url,omitempty"`
				Config         *models.JSONMap        `json:"config,omitempty"`
			}

			var statuses []IntegrationStatus
			for _, integration := range integrations {
				// Get associated tokens
				var tokens []models.IntegrationToken
				db.Where("integration_id = ?", integration.ID).Find(&tokens)

				var authToken *string
				var lastSync *string
				if len(tokens) > 0 {
					// VULNERABILITY: Exposing full token for "debugging connectivity"
					authToken = &tokens[0].Token
					if tokens[0].LastUsedAt != nil {
						s := tokens[0].LastUsedAt.Format("2006-01-02T15:04:05Z")
						lastSync = &s
					}
				}

				status := "connected"
				if !integration.Enabled {
					status = "disabled"
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
					Config:      integration.Config,
				})
			}

			// Log this access for audit
			userID, _ := c.Get("userID")
			userIDStr := userID.(string)
			auditService.Log(service.ActionInternalAPICall, &userIDStr,
				"integration_status", nil,
				map[string]interface{}{
					"integrations_count": len(statuses),
					"purpose": "operational_health_check",
				}, c.ClientIP())

			c.JSON(http.StatusOK, gin.H{"integrations": statuses})
		})

	// GET /api/internal/integrations/test - Test integration connectivity
	// Returns detailed diagnostic info including token validation results
	r.GET("/api/internal/integrations/test/:id",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			integrationID := c.Param("id")

			var integration models.Integration
			if err := db.Where("id = ?", integrationID).First(&integration).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
				return
			}

			// Get token for testing
			var token models.IntegrationToken
			if err := db.Where("integration_id = ?", integrationID).First(&token).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "No token configured"})
				return
			}

			// Simulate connectivity test
			testResult := map[string]interface{}{
				"integration_id":   integration.ID,
				"integration_name": integration.Name,
				"provider":         integration.Provider,
				"project":          integration.ProjectName,
				"token_valid":      true,
				"token_preview":    token.Token[:8] + "..." + token.Token[len(token.Token)-4:], // VULNERABILITY: Leaks token structure
				"token_full":       token.Token, // VULNERABILITY: Full token for "diagnostic purposes"
				"connection_test":  "success",
				"last_used":        token.LastUsedAt,
				"diagnostic_info": map[string]interface{}{
					"allowed_secrets":        token.AllowedSecrets,
					"allowed_environments":   token.AllowedEnvironments,
					"webhook_endpoint":       "/api/integrations/webhook",
					"expected_token_format":  "gf_prod_* or internal_*",
				},
			}

			c.JSON(http.StatusOK, gin.H{"test_result": testResult})
		})

	// POST /api/internal/secrets/grant - VULNERABILITY: Trusts source field without verification
	r.POST("/api/internal/secrets/grant",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			userID, _ := c.Get("userID")

			var req struct {
				SecretID      string          `json:"secret_id" binding:"required"`
				UserID        string          `json:"user_id" binding:"required"`
				Source        string          `json:"source" binding:"required"`
				SourceContext *models.JSONMap `json:"source_context"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// VULNERABILITY: Only checks source string, not actual origin
			trustedSources := []string{"webhook", "internal", "service_mesh"}
			if !contains(trustedSources, req.Source) {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Invalid source. Must be one of: webhook, internal, service_mesh",
				})
				return
			}

			// VULNERABILITY: source_context is caller-controlled and not validated
			// It should verify the context matches a real integration event
			sourceContext := req.SourceContext
			if sourceContext == nil {
				sourceContext = &models.JSONMap{}
			}
			(*sourceContext)["validated"] = false // Should be set by real validation

			// Create access request
			accessReq := &models.AccessRequest{
				SecretID:      req.SecretID,
				UserID:        req.UserID,
				Justification: "Internal API grant",
				Status:        "approved",
				AutoApproved:  true,
				Source:        req.Source,
				SourceContext: sourceContext,
			}

			if err := db.Create(accessReq).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Create grant immediately
			grant, err := approvalService.CreateAutoApprovedGrant(req.SecretID, req.UserID, accessReq.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			userIDStr := userID.(string)
			auditService.Log(service.ActionInternalAPICall, &userIDStr,
				"access_grant", &grant.ID,
				map[string]interface{}{
					"source":         req.Source,
					"source_context": sourceContext,
					"auto_approved":  true,
				}, c.ClientIP())

			// Get secret value to return
			var secret models.Secret
			if err := db.Where("id = ?", req.SecretID).First(&secret).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Secret not found"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"grant": map[string]interface{}{
					"id":            grant.ID,
					"auto_approved": true,
					"secret_value":  secret.Value,
					"expires_at":    grant.ExpiresAt,
				},
			})
		})

	// POST /api/internal/apply - VULNERABILITY: Missing auth check, bypasses classification
	r.POST("/api/internal/apply",
		func(c *gin.Context) {
			// VULNERABILITY: No authentication check
			// Assumes internal endpoints aren't discoverable
			var req struct {
				RequestID            string          `json:"request_id"`
				BypassClassification bool            `json:"bypass_classification_check"`
				Source               string          `json:"source"`
				SourceContext        *models.JSONMap `json:"source_context"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if req.BypassClassification && req.Source == "internal" {
				// VULNERABILITY: Bypasses classification-based approval
				var accessReq models.AccessRequest
				if err := db.Where("id = ?", req.RequestID).First(&accessReq).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
					return
				}

				accessReq.Status = "approved"
				accessReq.AutoApproved = true
				db.Save(&accessReq)

				// Create grant
				grant, _ := approvalService.CreateAutoApprovedGrant(
					accessReq.SecretID,
					accessReq.UserID,
					accessReq.ID,
				)

				c.JSON(http.StatusOK, gin.H{
					"approved": true,
					"reason":   "Internal source - auto-approved",
					"grant_id": grant.ID,
				})
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
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
