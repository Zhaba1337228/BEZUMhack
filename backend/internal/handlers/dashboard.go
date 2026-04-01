package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"secretflow/internal/middleware"
	"secretflow/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// In-memory notification channels for SSE (simple implementation)
// In production, this would use Redis pub/sub or similar
var notificationChannels = make(map[string]map[string]chan NotificationEvent)

// NotificationEvent represents a real-time notification
type NotificationEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id,omitempty"`
}

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
				Classification string `gorm:"column:classification" json:"classification"`
				Count          int64  `gorm:"column:count" json:"count"`
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

	// GET /api/events/stream - Server-Sent Events stream for real-time notifications
	r.GET("/api/events/stream",
		middleware.AuthHeaderOrQuery(jwtSecret, "token"),
		func(c *gin.Context) {
			userID, _ := c.Get("userID")
			userIDStr := userID.(string)

			// Set SSE headers
			c.Writer.Header().Set("Content-Type", "text/event-stream")
			c.Writer.Header().Set("Cache-Control", "no-cache")
			c.Writer.Header().Set("Connection", "keep-alive")
			c.Writer.Header().Set("X-Accel-Buffering", "no")
			c.Writer.Flush()

			// Register channel for this user
			if notificationChannels[userIDStr] == nil {
				notificationChannels[userIDStr] = make(map[string]chan NotificationEvent)
			}
			clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())
			eventChan := make(chan NotificationEvent, 10)
			notificationChannels[userIDStr][clientID] = eventChan

			// Cleanup on disconnect
			defer func() {
				close(eventChan)
				delete(notificationChannels[userIDStr], clientID)
				if len(notificationChannels[userIDStr]) == 0 {
					delete(notificationChannels, userIDStr)
				}
			}()

			// Send initial connection event
			fmt.Fprintf(c.Writer, "event: connected\ndata: {\"message\": \"Connected to notification stream\"}\n\n")
			c.Writer.Flush()

			// Keep connection alive
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-c.Request.Context().Done():
					return
				case <-ticker.C:
					fmt.Fprintf(c.Writer, ": ping\n\n")
					c.Writer.Flush()
				case event := <-eventChan:
					payload, _ := json.Marshal(event)
					fmt.Fprintf(c.Writer, "event: notification\ndata: %s\n\n", payload)
					c.Writer.Flush()
				}
			}
		})
}

// BroadcastNotification sends a notification to a specific user
func BroadcastNotification(userID string, event NotificationEvent) {
	if channels, ok := notificationChannels[userID]; ok {
		for _, ch := range channels {
			select {
			case ch <- event:
			default:
				// Channel full, skip
			}
		}
	}
}
