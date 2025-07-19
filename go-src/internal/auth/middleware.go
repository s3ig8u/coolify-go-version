package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT and loads user into context
func AuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		header := c.GetHeader("Authorization")
		tokenString, err := ExtractTokenFromHeader(header)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set claims in context (user loading can be done in handlers)
		c.Set("claims", claims)
		c.Next()
	}
}

// Optional: Role-based middleware
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsInterface, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No authentication claims found"})
			c.Abort()
			return
		}

		claims, ok := claimsInterface.(*Claims)
		if !ok || claims.Role != role {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// TODO: Add team membership/role middleware
// func RequireTeamRole(role string) gin.HandlerFunc { ... }
