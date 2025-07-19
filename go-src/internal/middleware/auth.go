package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MockAuth provides a mock authentication middleware for development
// In production, this should be replaced with proper JWT or session-based auth
func MockAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// For development, create a mock user
		// In production, extract from JWT token or session
		mockUserID := uint(1)
		mockUserEmail := "admin@coolify.local"
		mockUserName := "Admin User"

		// Set user information in context for handlers to use
		c.Set("userID", mockUserID)
		c.Set("userEmail", mockUserEmail)
		c.Set("userName", mockUserName)
		c.Set("authenticated", true)

		c.Next()
	})
}

// AuthRequired middleware that checks if user is authenticated
func AuthRequired() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authenticated, exists := c.Get("authenticated")
		if !exists || !authenticated.(bool) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// AdminRequired middleware that checks if user has admin role
func AdminRequired() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check if user is authenticated first
		authenticated, exists := c.Get("authenticated")
		if !exists || !authenticated.(bool) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// In a real implementation, check user role from database
		// For now, allow all authenticated users
		c.Next()
	})
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	if id, ok := userID.(uint); ok {
		return id
	}
	return 0
}

// GetUserEmail extracts user email from context
func GetUserEmail(c *gin.Context) string {
	userEmail, exists := c.Get("userEmail")
	if !exists {
		return ""
	}
	if email, ok := userEmail.(string); ok {
		return email
	}
	return ""
}
