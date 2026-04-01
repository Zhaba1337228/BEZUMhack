package handlers

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"secretflow/internal/middleware"
	"secretflow/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type auditLogResponse struct {
	ID           string          `json:"id"`
	Timestamp    time.Time       `json:"timestamp"`
	UserID       *string         `json:"user_id,omitempty"`
	Username     string          `json:"username,omitempty"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type,omitempty"`
	ResourceID   *string         `json:"resource_id,omitempty"`
	Details      *models.JSONMap `json:"details,omitempty"`
	IPAddress    string          `json:"ip_address,omitempty"`
	Risky        bool            `json:"risky"`
	SecretID     string          `json:"secret_id,omitempty"`
	Purpose      string          `json:"purpose,omitempty"`
}

type timelineEvent struct {
	Timestamp     time.Time `json:"timestamp"`
	EventType     string    `json:"event_type"`
	ActorID       string    `json:"actor_id,omitempty"`
	ActorUsername string    `json:"actor_username,omitempty"`
	TargetUserID  string    `json:"target_user_id,omitempty"`
	TargetUser    string    `json:"target_user,omitempty"`
	Status        string    `json:"status,omitempty"`
	Justification string    `json:"justification,omitempty"`
	Source        string    `json:"source,omitempty"`
	IPAddress     string    `json:"ip_address,omitempty"`
	Details       string    `json:"details,omitempty"`
}

func SetupAuditHandler(r *gin.Engine, db *gorm.DB, jwtSecret string) {
	// GET /api/audit/logs - List audit logs
	r.GET("/api/audit/logs",
		middleware.StrictAuth(jwtSecret),
		middleware.RequireRole("security_admin"),
		func(c *gin.Context) {
			userID := c.Query("user_id")
			action := c.Query("action")
			secretID := c.Query("secret_id")
			riskyOnly := c.Query("risky") == "true"
			limit := parseLimit(c.Query("limit"), 200)

			logs, err := models.ListAuditLogs(db, userID, action, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			enriched, err := enrichLogs(db, logs)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			filtered := filterLogs(enriched, action, secretID, riskyOnly)
			c.JSON(http.StatusOK, gin.H{"logs": filtered})
		})

	// GET /api/audit/export.csv - CSV export for logs
	r.GET("/api/audit/export.csv",
		middleware.AuthHeaderOrQuery(jwtSecret, "token"),
		middleware.RequireRole("security_admin"),
		func(c *gin.Context) {
			userID := c.Query("user_id")
			action := c.Query("action")
			secretID := c.Query("secret_id")
			riskyOnly := c.Query("risky") == "true"
			limit := parseLimit(c.Query("limit"), 1000)

			logs, err := models.ListAuditLogs(db, userID, action, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			enriched, err := enrichLogs(db, logs)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			filtered := filterLogs(enriched, action, secretID, riskyOnly)

			c.Header("Content-Type", "text/csv; charset=utf-8")
			c.Header("Content-Disposition", "attachment; filename=audit_logs.csv")

			writer := csv.NewWriter(c.Writer)
			defer writer.Flush()

			_ = writer.Write([]string{"timestamp", "action", "resource_type", "resource_id", "user_id", "username", "ip_address", "risky", "secret_id", "purpose", "details_json"})
			for _, log := range filtered {
				detailsJSON := ""
				if log.Details != nil {
					if b, err := json.Marshal(log.Details); err == nil {
						detailsJSON = string(b)
					}
				}
				resourceID := ""
				userIDVal := ""
				if log.ResourceID != nil {
					resourceID = *log.ResourceID
				}
				if log.UserID != nil {
					userIDVal = *log.UserID
				}
				_ = writer.Write([]string{
					log.Timestamp.Format(time.RFC3339),
					log.Action,
					log.ResourceType,
					resourceID,
					userIDVal,
					log.Username,
					log.IPAddress,
					strconv.FormatBool(log.Risky),
					log.SecretID,
					log.Purpose,
					detailsJSON,
				})
			}
		})

	// GET /api/audit/timeline/:secretID - Timeline by secret
	r.GET("/api/audit/timeline/:secretID",
		middleware.StrictAuth(jwtSecret),
		middleware.RequireRole("security_admin"),
		func(c *gin.Context) {
			secretID := c.Param("secretID")
			limit := parseLimit(c.Query("limit"), 300)

			secret, err := models.GetSecretByUUID(db, secretID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
				return
			}

			timeline, err := buildSecretTimeline(db, secretID, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"secret": gin.H{
					"id":             secret.ID,
					"name":           secret.Name,
					"classification": secret.Classification,
					"environment":    secret.Environment,
					"owner_team":     secret.OwnerTeam,
				},
				"timeline": timeline,
			})
		})

	// GET /api/audit/stats - Get audit statistics
	r.GET("/api/audit/stats",
		middleware.StrictAuth(jwtSecret),
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

func parseLimit(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	if v > 5000 {
		return 5000
	}
	return v
}

func enrichLogs(db *gorm.DB, logs []models.AuditLog) ([]auditLogResponse, error) {
	userIDs := make(map[string]struct{})
	for _, log := range logs {
		if log.UserID != nil && *log.UserID != "" {
			userIDs[*log.UserID] = struct{}{}
		}
		if log.Details != nil {
			if uid, ok := (*log.Details)["user_id"].(string); ok && uid != "" {
				userIDs[uid] = struct{}{}
			}
		}
	}

	userMap, err := loadUsernames(db, userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]auditLogResponse, 0, len(logs))
	for _, log := range logs {
		username := ""
		if log.UserID != nil {
			username = userMap[*log.UserID]
		}
		secretID := extractSecretID(log)
		purpose := extractPurpose(log)
		result = append(result, auditLogResponse{
			ID:           log.ID,
			Timestamp:    log.Timestamp,
			UserID:       log.UserID,
			Username:     username,
			Action:       log.Action,
			ResourceType: log.ResourceType,
			ResourceID:   log.ResourceID,
			Details:      log.Details,
			IPAddress:    log.IPAddress,
			Risky:        isRisky(log, secretID),
			SecretID:     secretID,
			Purpose:      purpose,
		})
	}
	return result, nil
}

func filterLogs(logs []auditLogResponse, actionFilter, secretID string, riskyOnly bool) []auditLogResponse {
	filtered := make([]auditLogResponse, 0, len(logs))
	actionFilter = strings.TrimSpace(strings.ToLower(actionFilter))
	for _, log := range logs {
		if actionFilter != "" && !strings.Contains(strings.ToLower(log.Action), actionFilter) {
			continue
		}
		if secretID != "" && log.SecretID != secretID {
			continue
		}
		if riskyOnly && !log.Risky {
			continue
		}
		filtered = append(filtered, log)
	}
	return filtered
}

func buildSecretTimeline(db *gorm.DB, secretID string, limit int) ([]timelineEvent, error) {
	events := make([]timelineEvent, 0, limit)

	var requests []models.AccessRequest
	if err := db.Preload("Secret").Where("secret_id = ?", secretID).Order("created_at DESC").Limit(limit).Find(&requests).Error; err != nil {
		return nil, err
	}

	userIDs := make(map[string]struct{})
	for _, req := range requests {
		userIDs[req.UserID] = struct{}{}
		if req.DecidedBy != nil {
			userIDs[*req.DecidedBy] = struct{}{}
		}
	}

	var grants []models.AccessGrant
	if err := db.Where("secret_id = ?", secretID).Order("granted_at DESC").Limit(limit).Find(&grants).Error; err != nil {
		return nil, err
	}
	for _, g := range grants {
		userIDs[g.UserID] = struct{}{}
	}

	var usageLogs []models.AuditLog
	if err := db.Where("action = ? AND (resource_id = ? OR details->>'secret_id' = ?)", "secret_value_revealed", secretID, secretID).
		Order("timestamp DESC").Limit(limit).Find(&usageLogs).Error; err != nil {
		return nil, err
	}
	for _, l := range usageLogs {
		if l.UserID != nil {
			userIDs[*l.UserID] = struct{}{}
		}
	}

	userMap, err := loadUsernames(db, userIDs)
	if err != nil {
		return nil, err
	}

	for _, req := range requests {
		events = append(events, timelineEvent{
			Timestamp:     req.CreatedAt,
			EventType:     "request_created",
			ActorID:       req.UserID,
			ActorUsername: userMap[req.UserID],
			TargetUserID:  req.UserID,
			TargetUser:    userMap[req.UserID],
			Status:        req.Status,
			Justification: req.Justification,
			Source:        req.Source,
			Details:       "Access request submitted",
		})
		if req.DecidedAt != nil && req.DecidedBy != nil {
			events = append(events, timelineEvent{
				Timestamp:     *req.DecidedAt,
				EventType:     "request_decided",
				ActorID:       *req.DecidedBy,
				ActorUsername: userMap[*req.DecidedBy],
				TargetUserID:  req.UserID,
				TargetUser:    userMap[req.UserID],
				Status:        req.Status,
				Justification: req.Justification,
				Source:        req.Source,
				Details:       "Request was " + req.Status,
			})
		}
	}

	for _, grant := range grants {
		events = append(events, timelineEvent{
			Timestamp:     grant.GrantedAt,
			EventType:     "grant_created",
			TargetUserID:  grant.UserID,
			TargetUser:    userMap[grant.UserID],
			Status:        "approved",
			Details:       "Access grant active until " + grant.ExpiresAt.Format(time.RFC3339),
		})
	}

	for _, log := range usageLogs {
		actorID := ""
		actorName := ""
		if log.UserID != nil {
			actorID = *log.UserID
			actorName = userMap[*log.UserID]
		}
		events = append(events, timelineEvent{
			Timestamp:     log.Timestamp,
			EventType:     "secret_used",
			ActorID:       actorID,
			ActorUsername: actorName,
			IPAddress:     log.IPAddress,
			Status:        "revealed",
			Details:       extractPurpose(log),
		})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})
	if len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

func loadUsernames(db *gorm.DB, ids map[string]struct{}) (map[string]string, error) {
	result := make(map[string]string)
	if len(ids) == 0 {
		return result, nil
	}
	idList := make([]string, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	var users []models.User
	if err := db.Where("id IN ?", idList).Find(&users).Error; err != nil {
		return nil, err
	}
	for _, u := range users {
		result[u.ID] = u.Username
	}
	return result, nil
}

func extractSecretID(log models.AuditLog) string {
	if log.ResourceType == "secret" && log.ResourceID != nil {
		return *log.ResourceID
	}
	if log.Details == nil {
		return ""
	}
	if sid, ok := (*log.Details)["secret_id"].(string); ok {
		return sid
	}
	if reqBody, ok := (*log.Details)["request_body"].(map[string]interface{}); ok {
		if sid, ok := reqBody["secret_id"].(string); ok {
			return sid
		}
	}
	return ""
}

func extractPurpose(log models.AuditLog) string {
	if log.Details == nil {
		return ""
	}
	if justification, ok := (*log.Details)["justification"].(string); ok {
		return justification
	}
	if reason, ok := (*log.Details)["reason"].(string); ok {
		return reason
	}
	if message, ok := (*log.Details)["message"].(string); ok {
		return message
	}
	if reqBody, ok := (*log.Details)["request_body"].(map[string]interface{}); ok {
		if justification, ok := reqBody["justification"].(string); ok {
			return justification
		}
	}
	return ""
}

func isRisky(log models.AuditLog, secretID string) bool {
	action := strings.ToLower(log.Action)
	if strings.Contains(action, "failure") || strings.Contains(action, "denied") {
		return true
	}
	if action == "internal_api_call" || action == "integration_token_used" {
		return true
	}
	if strings.Contains(action, "grant") && log.Details != nil {
		if approvalStatus, ok := (*log.Details)["approval_status"].(string); ok && strings.Contains(strings.ToLower(approvalStatus), "auto") {
			return true
		}
	}
	if log.Details != nil {
		if classification, ok := (*log.Details)["classification"].(string); ok {
			if classification == "CRITICAL" || classification == "HIGH" {
				return true
			}
		}
	}
	if secretID != "" {
		return true
	}
	return false
}
