package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vikhyat-sharma/astrology-ai/internal/services"
)

// AuthService interface for dependency injection
type AuthServiceInterface interface {
	ValidateToken(token string) (uuid.UUID, error)
}

// AuthRequired middleware checks for valid JWT token
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Get auth service from context (set in main)
		authService, exists := c.Get("authService")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth service not available"})
			c.Abort()
			return
		}

		userID, err := authService.(*services.AuthService).ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user ID in context for handlers to use
		c.Set("userID", userID)
		c.Next()
	}
}

// CORS middleware adds CORS headers
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Logger middleware logs requests
func Logger() gin.HandlerFunc {
	return gin.Logger()
}

// ErrorHandler middleware handles errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}