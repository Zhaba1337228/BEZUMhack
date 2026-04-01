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

func SetupRequestsHandler(r *gin.Engine, db *gorm.DB, jwtSecret string) {
	approvalService := service.NewApprovalService(db)
	auditService := service.NewAuditService(db)

	// GET /api/requests - List requests
	r.GET("/api/requests",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			userID, _ := c.Get("userID")
			userRole, _ := c.Get("role")

			pendingOnly := c.Query("pending") == "true"
			mineOnly := c.Query("mine") == "true"
			status := c.Query("status")

			// Security admin sees all, others see only their requests
			listUserID := ""
			if userRole != "security_admin" {
				listUserID = userID.(string)
			}
			// Explicit override for "My Requests" page
			if mineOnly {
				listUserID = userID.(string)
			}

			requests, err := models.ListRequests(db, listUserID, status, pendingOnly)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"requests": requests})
		})

	// POST /api/requests - Create request
	r.POST("/api/requests",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			userID, _ := c.Get("userID")

			var req struct {
				SecretID      string `json:"secret_id" binding:"required"`
				Justification string `json:"justification" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Verify secret exists
			_, err := models.GetSecretByUUID(db, req.SecretID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
				return
			}

			// Create access request
			accessReq := &models.AccessRequest{
				ID:            uuid.New().String(),
				SecretID:      req.SecretID,
				UserID:        userID.(string),
				Justification: req.Justification,
				Status:        "pending",
				AutoApproved:  false,
				Source:        "api",
			}

			if err := models.CreateAccessRequest(db, accessReq); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, gin.H{"request": accessReq})
		})

	// POST /api/requests/:id/approve - Approve request
	r.POST("/api/requests/:id/approve",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			userID, _ := c.Get("userID")

			reqID := c.Param("id")
			reqObj, _ := models.GetRequestByUUID(db, reqID)

			// Get approver user
			approver, err := models.GetUserByUUID(db, userID.(string))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}

			grant, err := approvalService.Approve(reqID, approver)
			if err != nil {
				if err == service.ErrAlreadyDecided {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Request already decided"})
					return
				}
				if err == service.ErrInsufficientRole {
					c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			userIDStr := userID.(string)
			auditService.Log(service.ActionRequestApproved, &userIDStr,
				"access_request", &reqID,
				map[string]interface{}{"grant_id": grant.ID}, c.ClientIP())

			if reqObj != nil {
				BroadcastNotification(reqObj.UserID, NotificationEvent{
					ID:        reqID,
					Type:      "approval",
					Title:     "Request Approved",
					Message:   "Access approved for " + reqObj.Secret.Name,
					Timestamp: time.Now(),
					UserID:    reqObj.UserID,
				})
			}

			c.JSON(http.StatusOK, gin.H{"grant": grant})
		})

	// POST /api/requests/:id/deny - Deny request
	r.POST("/api/requests/:id/deny",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			userID, _ := c.Get("userID")

			reqID := c.Param("id")
			reqObj, _ := models.GetRequestByUUID(db, reqID)

			denier, err := models.GetUserByUUID(db, userID.(string))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}

			if err := approvalService.Deny(reqID, denier); err != nil {
				if err == service.ErrAlreadyDecided {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Request already decided"})
					return
				}
				if err == service.ErrInsufficientRole {
					c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			userIDStr := userID.(string)
			auditService.Log(service.ActionRequestDenied, &userIDStr,
				"access_request", &reqID, nil, c.ClientIP())

			if reqObj != nil {
				BroadcastNotification(reqObj.UserID, NotificationEvent{
					ID:        reqID,
					Type:      "alert",
					Title:     "Request Denied",
					Message:   "Access denied for " + reqObj.Secret.Name,
					Timestamp: time.Now(),
					UserID:    reqObj.UserID,
				})
			}

			c.JSON(http.StatusOK, gin.H{"status": "denied"})
		})
}
