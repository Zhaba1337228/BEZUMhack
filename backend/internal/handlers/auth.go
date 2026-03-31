package handlers

import (
	"net/http"
	"secretflow/internal/middleware"
	"secretflow/internal/models"
	"secretflow/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupAuthHandler(r *gin.Engine, db *gorm.DB, jwtSecret string, jwtExpiry int) {
	authService := service.NewAuthService(db, jwtSecret, jwtExpiry)
	auditService := service.NewAuditService(db)

	// POST /api/auth/login
	r.POST("/api/auth/login", func(c *gin.Context) {
		var req service.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := authService.Login(&req)
		if err != nil {
			auditService.Log(service.ActionLoginFailure, nil, "user", nil,
				map[string]interface{}{"username": req.Username}, c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		auditService.Log(service.ActionLoginSuccess, &resp.User.ID, "user", &resp.User.ID,
			nil, c.ClientIP())

		c.JSON(http.StatusOK, gin.H{
			"token": resp.Token,
			"user":  resp.User,
		})
	})

	// GET /api/auth/me - requires auth
	r.GET("/api/auth/me",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			userID, _ := c.Get("userID")
			user, err := models.GetUserByUUID(db, userID.(string))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"user": user})
		})
}
