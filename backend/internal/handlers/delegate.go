package handlers

import (
	"net/http"
	"secretflow/internal/middleware"
	"secretflow/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupDelegateHandler(r *gin.Engine, db *gorm.DB, jwtSecret string) {
	delegateService := service.NewDelegateService(db, jwtSecret)

	// POST /api/service-account/exchange - Exchange integration token for service account JWT
	// This endpoint allows CI/CD systems to get a temporary JWT for debugging
	//
	// HARD PATH 2 VULNERABILITY:
	// This endpoint is the entry point for the trust boundary confusion attack.
	// An attacker who obtains an integration token (via Path 1 or other means)
	// can exchange it for a service account JWT, which can then be used to
	// delegate access to themselves.
	r.POST("/api/service-account/exchange",
		func(c *gin.Context) {
			var req service.ServiceTokenExchangeRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			resp, err := delegateService.ExchangeIntegrationToken(&req)
			if err != nil {
				if err == service.ErrInvalidServiceToken {
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": "Invalid integration token",
						"code":  middleware.ErrInvalidToken,
					})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, resp)
		})

	// POST /api/delegate/access - Delegate access to a user
	// Requires service account JWT (obtained from /api/service-account/exchange)
	//
	// HARD PATH 2 VULNERABILITY (Confused Deputy / Trust Boundary Confusion):
	// The endpoint trusts the service account JWT and creates a grant without
	// verifying that the service account has rights to delegate THIS specific secret.
	//
	// Attack chain:
	// 1. Attacker gets integration token (via Path 1: /api/internal/integrations/status)
	// 2. Attacker exchanges token for service account JWT via /api/service-account/exchange
	// 3. Attacker calls /api/delegate/access with their own user ID as target
	// 4. Attacker gains access to CRITICAL secret
	r.POST("/api/delegate/access",
		middleware.StrictAuth(jwtSecret),
		middleware.RequireRole("service_account"),
		func(c *gin.Context) {
			var req service.DelegationRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			serviceAccountID, _ := c.Get("userID")

			resp, err := delegateService.DelegateAccess(&req, serviceAccountID.(string))
			if err != nil {
				switch err {
				case service.ErrInvalidTargetUser:
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid secret or target user"})
				default:
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				}
				return
			}

			c.JSON(http.StatusOK, resp)
		})

	// GET /api/delegate/info - Informational endpoint about delegation
	// This is a "hint" endpoint that helps CTF players understand the delegation flow
	r.GET("/api/delegate/info",
		middleware.Auth(jwtSecret),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"delegation_info": map[string]interface{}{
					"description": "Service accounts can delegate access to users via the /api/delegate/access endpoint",
					"flow": []string{
						"1. Obtain integration token from /api/internal/integrations/status",
						"2. Exchange integration token for service account JWT at /api/service-account/exchange",
						"3. Use service account JWT to call /api/delegate/access",
					},
					"note": "Service accounts have elevated privileges for CI/CD automation",
				},
			})
		})
}
