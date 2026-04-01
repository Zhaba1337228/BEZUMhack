package middleware

import (
	"net/http"
	"secretflow/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIError represents a standardized API error response
type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	StatusCode int    `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrUnauthorized          = "UNAUTHORIZED"
	ErrForbidden             = "FORBIDDEN"
	ErrInvalidToken          = "INVALID_TOKEN"
	ErrTokenExpired          = "TOKEN_EXPIRED"
	ErrInsufficientRole      = "INSUFFICIENT_ROLE"
	ErrMissingAuthHeader     = "MISSING_AUTH_HEADER"
	ErrInvalidAuthFormat     = "INVALID_AUTH_FORMAT"
)

// Auth middleware with strict JWT validation
// If session is invalid/expired/tampered: immediate 401, no side-effects
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := ExtractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
				"code":  ErrInvalidAuthFormat,
			})
			c.Abort()
			return
		}
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authorization header",
				"code":  ErrMissingAuthHeader,
			})
			c.Abort()
			return
		}

		claims, err := jwt.ValidateToken(token, secret)
		if err != nil {
			// Determine specific error type
			errorCode := ErrInvalidToken
			if strings.Contains(err.Error(), "expired") {
				errorCode = ErrTokenExpired
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  errorCode,
			})
			c.Abort()
			return
		}

		// Set user context - only AFTER successful validation
		SetAuthContext(c, claims.UserID, claims.Username, claims.Role, claims.Team, claims)

		c.Next()
	}
}

// StrictAuth is an alias for Auth with the same behavior
// Use StrictAuth when you want to emphasize that this endpoint
// requires valid authentication and will have NO side-effects
// if auth fails (no inserts, updates, audit events, etc.)
func StrictAuth(secret string) gin.HandlerFunc {
	return Auth(secret)
}

// AuthHeaderOrQuery validates JWT from Authorization header or query param.
// Useful for EventSource/SSE where custom headers are not reliably supported by browsers.
func AuthHeaderOrQuery(secret string, queryParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := ExtractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
				"code":  ErrInvalidAuthFormat,
			})
			c.Abort()
			return
		}
		if token == "" {
			token = c.Query(queryParam)
		}
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authentication token",
				"code":  ErrMissingAuthHeader,
			})
			c.Abort()
			return
		}

		claims, err := jwt.ValidateToken(token, secret)
		if err != nil {
			errorCode := ErrInvalidToken
			if strings.Contains(err.Error(), "expired") {
				errorCode = ErrTokenExpired
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  errorCode,
			})
			c.Abort()
			return
		}

		SetAuthContext(c, claims.UserID, claims.Username, claims.Role, claims.Team, claims)
		c.Next()
	}
}

func SetAuthContext(c *gin.Context, userID, username, role, team string, claims interface{}) {
	c.Set("userID", userID)
	c.Set("username", username)
	c.Set("role", role)
	c.Set("team", team)
	c.Set("claims", claims)
}

func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", nil
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", &APIError{
			Code:       ErrInvalidAuthFormat,
			Message:    "Invalid authorization format. Use: Bearer <token>",
			StatusCode: http.StatusUnauthorized,
		}
	}
	return parts[1], nil
}

// TokenAuth validates integration tokens from request body
// Strict version: validates token BEFORE any processing
func TokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// Try from JSON body first
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err == nil {
			if t, ok := body["token"].(string); ok {
				token = t
			}
		}

		// Try from form data
		if token == "" {
			token = c.PostForm("token")
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing integration token",
				"code":  ErrMissingAuthHeader,
			})
			c.Abort()
			return
		}

		c.Set("integrationToken", token)
		c.Next()
	}
}

// RequireRole middleware checks if user has required role
// Must be used AFTER Auth middleware
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  ErrUnauthorized,
			})
			c.Abort()
			return
		}

		if role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"code":  ErrInsufficientRole,
				"required": requiredRole,
				"current":  role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole checks if user has ANY of the specified roles
func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  ErrUnauthorized,
			})
			c.Abort()
			return
		}

		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
			"code":  ErrInsufficientRole,
			"required": roles,
			"current":  role,
		})
		c.Abort()
	}
}

// ErrorHandler middleware standardizes error responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				if apiErr, ok := err.Err.(*APIError); ok {
					c.JSON(apiErr.StatusCode, gin.H{
						"error":   apiErr.Message,
						"code":    apiErr.Code,
						"details": apiErr.Details,
					})
					return
				}
			}
		}
	}
}
