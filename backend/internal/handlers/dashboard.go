package handlers

import (
	"net/http"
	"secretflow/internal/middleware"
	"secretflow/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupDashboardHandler(r *gin.Engine, db *gorm.DB, jwtSecret string) {
	// GET /api/dashboard/summary - Dashboard summary
	r.GET("/api/dashboard/summary",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			userID, _ := c.Get("userID")
			userRole, _ := c.Get("role")

			// Count total secrets
			var totalSecrets int64
			db.Model(&models.Secret{}).Count(&totalSecrets)

			// Count secrets by classification
			type ClassCount struct {
				Classification string `gorm:"column:classification"`
				Count          int64  `gorm:"column:count"`
			}
			var classCounts []ClassCount
			db.Model(&models.Secret{}).
				Select("classification, COUNT(*) as count").
				Group("classification").
				Scan(&classCounts)

			// Count user's pending requests
			var pendingCount int64
			db.Model(&models.AccessRequest{}).
				Where("user_id = ? AND status = ?", userID, "pending").
				Count(&pendingCount)

			// Count grants
			var grantCount int64
			db.Model(&models.AccessGrant{}).
				Where("user_id = ? AND revoked = ? AND expires_at > ?", userID, false, time.Now()).
				Count(&grantCount)

			// For admins: count pending approvals
			var pendingApprovals int64
			if userRole == "team_lead" || userRole == "security_admin" {
				db.Model(&models.AccessRequest{}).
					Where("status = ?", "pending").
					Count(&pendingApprovals)
			}

			c.JSON(http.StatusOK, gin.H{
				"total_secrets":             totalSecrets,
				"secrets_by_classification": classCounts,
				"my_pending_requests":       pendingCount,
				"my_active_grants":          grantCount,
				"pending_approvals":         pendingApprovals,
			})
		})

	// GET /api/dashboard/pending - Pending requests count
	r.GET("/api/dashboard/pending",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			userRole, _ := c.Get("role")

			if userRole != "team_lead" && userRole != "security_admin" {
				c.JSON(http.StatusOK, gin.H{"pending": 0})
				return
			}

			var count int64
			db.Model(&models.AccessRequest{}).
				Where("status = ?", "pending").
				Count(&count)

			c.JSON(http.StatusOK, gin.H{"pending": count})
		})
}
