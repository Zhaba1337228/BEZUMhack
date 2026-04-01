package handlers

import (
	"net/http"
	"secretflow/internal/middleware"
	"secretflow/internal/models"
	"secretflow/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SetupSecretsHandler(r *gin.Engine, db *gorm.DB, jwtSecret string) {
	auditService := service.NewAuditService(db)

	// GET /api/secrets - List secrets (filtered by role)
	r.GET("/api/secrets",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			userID, _ := c.Get("userID")

			classification := c.Query("classification")
			environment := c.Query("environment")
			ownerTeam := c.Query("owner_team")

			secrets, err := models.ListSecrets(db, classification, environment, ownerTeam)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Build response with access status
			type SecretResponse struct {
				ID             string `json:"id"`
				Name           string `json:"name"`
				Description    string `json:"description"`
				Classification string `json:"classification"`
				Environment    string `json:"environment"`
				OwnerTeam      string `json:"owner_team"`
				HasAccess      bool   `json:"has_access"`
				PendingRequest bool   `json:"pending_request"`
			}

			var response []SecretResponse
			for _, secret := range secrets {
				// Check if user has active grant
				grant, err := models.GetActiveGrant(db, userID.(string), secret.ID)
				hasAccess := err == nil && grant != nil && !grant.Revoked

				// Check if user has pending request
				var reqCount int64
				db.Model(&models.AccessRequest{}).
					Where("user_id = ? AND secret_id = ? AND status = ?",
						userID.(string), secret.ID, "pending").Count(&reqCount)

				response = append(response, SecretResponse{
					ID:             secret.ID,
					Name:           secret.Name,
					Description:    secret.Description,
					Classification: secret.Classification,
					Environment:    secret.Environment,
					OwnerTeam:      secret.OwnerTeam,
					HasAccess:      hasAccess,
					PendingRequest: reqCount > 0,
				})
			}

			userIDStr := userID.(string)
			auditService.Log(service.ActionSecretView, &userIDStr, "secrets", nil,
				map[string]interface{}{"count": len(secrets)}, c.ClientIP())

			c.JSON(http.StatusOK, gin.H{"secrets": response})
		})

	// GET /api/secrets/:id - Get secret metadata
	r.GET("/api/secrets/:id",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			secretID := c.Param("id")
			userID, _ := c.Get("userID")

			secret, err := models.GetSecretByUUID(db, secretID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
				return
			}

			// Check if user has active grant
			grant, err := models.GetActiveGrant(db, userID.(string), secretID)
			hasAccess := err == nil && grant != nil && !grant.Revoked

			c.JSON(http.StatusOK, gin.H{
				"secret": map[string]interface{}{
					"id":             secret.ID,
					"name":           secret.Name,
					"description":    secret.Description,
					"classification": secret.Classification,
					"environment":    secret.Environment,
					"owner_team":     secret.OwnerTeam,
					"has_access":     hasAccess,
				},
			})
		})

	// GET /api/secrets/:id/value - Get secret value (requires grant)
	r.GET("/api/secrets/:id/value",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			userID, _ := c.Get("userID")
			secretID := c.Param("id")

			// Check for active grant
			grant, err := models.GetActiveGrant(db, userID.(string), secretID)
			if err != nil || grant == nil || grant.Revoked {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "No active access grant. Please request access.",
				})
				return
			}

			// Get secret value
			secret, err := models.GetSecretByUUID(db, secretID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"secret": map[string]interface{}{
					"id":    secret.ID,
					"name":  secret.Name,
					"value": secret.Value,
				},
			})

			userIDStr := userID.(string)
			_ = auditService.Log(service.ActionSecretValueRevealed, &userIDStr, "secret", &secretID,
				map[string]interface{}{
					"secret_id":       secretID,
					"secret_name":     secret.Name,
					"classification":  secret.Classification,
					"justification":   "Direct secret value reveal endpoint usage",
				}, c.ClientIP())
		})

	// POST /api/secrets/:id/request - Request access
	r.POST("/api/secrets/:id/request",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			userID, _ := c.Get("userID")
			userRole, _ := c.Get("role")
			secretID := c.Param("id")

			var req struct {
				Justification string `json:"justification" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Verify secret exists
			secret, err := models.GetSecretByUUID(db, secretID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
				return
			}

			// Security admins get auto-approved for any secret
			isAutoApproved := userRole == "security_admin"

			// LOW classification is also auto-approved
			if !isAutoApproved {
				requiredApprover := service.GetRequiredApproverRole(secret.Classification)
				isAutoApproved = requiredApprover == ""
			}

			status := "approved"
			if !isAutoApproved {
				status = "pending"
			}

			// Create access request
			accessReq := &models.AccessRequest{
				ID:            uuid.New().String(),
				SecretID:      secretID,
				UserID:        userID.(string),
				Justification: req.Justification,
				Status:        status,
				AutoApproved:  isAutoApproved,
				Source:        "ui",
			}

			if err := models.CreateAccessRequest(db, accessReq); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// If auto-approved, create the access grant immediately
			if isAutoApproved {
				approvalService := service.NewApprovalService(db)
				grant, err := approvalService.CreateAutoApprovedGrant(secretID, userID.(string), accessReq.ID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access grant: " + err.Error()})
					return
				}
				_ = grant // unused for now
			}

			BroadcastNotification(userID.(string), NotificationEvent{
				ID:        accessReq.ID,
				Type:      "request",
				Title:     "Access Request Created",
				Message:   "Your request is " + status + " for " + secret.Name,
				Timestamp: time.Now(),
				UserID:    userID.(string),
			})

			userIDStr := userID.(string)
			auditService.Log(service.ActionSecretAccessRequest, &userIDStr,
				"access_request", &accessReq.ID,
				map[string]interface{}{
					"secret_id":         secretID,
					"classification":    secret.Classification,
					"requires_approval": !isAutoApproved,
					"auto_approved":     isAutoApproved,
				}, c.ClientIP())

			c.JSON(http.StatusCreated, gin.H{
				"request": map[string]interface{}{
					"id":            accessReq.ID,
					"secret_id":     secretID,
					"status":        status,
					"auto_approved": isAutoApproved,
				},
			})
		})
}
