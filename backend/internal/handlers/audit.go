package handlers

import (
	"net/http"
	"secretflow/internal/middleware"
	"secretflow/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupAuditHandler(r *gin.Engine, db *gorm.DB, jwtSecret string) {
	// GET /api/audit/logs - List audit logs (security_admin only)
	r.GET("/api/audit/logs",
		middleware.Auth(jwtSecret),
		middleware.RequireRole("security_admin"),
		func(c *gin.Context) {
			userID := c.Query("user_id")
			action := c.Query("action")
			limit := 100

			logs, err := models.ListAuditLogs(db, userID, action, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"logs": logs})
		})

	// GET /api/audit/stats - Get audit statistics (security_admin only)
	r.GET("/api/audit/stats",
		middleware.Auth(jwtSecret),
		middleware.RequireRole("security_admin"),
		func(c *gin.Context) {
			type ActionCount struct {
				Action string `gorm:"column:action"`
				Count  int64  `gorm:"column:count"`
			}

			var actionCounts []ActionCount
			db.Model(&models.AuditLog{}).
				Select("action, COUNT(*) as count").
				Group("action").
				Scan(&actionCounts)

			c.JSON(http.StatusOK, gin.H{
				"by_action": actionCounts,
			})
		})
}
