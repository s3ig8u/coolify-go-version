package handlers

import (
	"net/http"
	"strings"

	"coolify-go/internal/auth"
	"coolify-go/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthHandler provides authentication endpoints
type AuthHandler struct {
	authService  *services.AuthService
	oauthService *auth.OAuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService, oauthService *auth.OAuthService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		oauthService: oauthService,
	}
}

// RegisterRoutes registers auth endpoints
func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	// Standard auth endpoints
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	r.POST("/logout", h.Logout)
	r.POST("/refresh", h.Refresh)

	// OAuth endpoints
	oauth := r.Group("/oauth")
	{
		oauth.GET("/:provider", h.OAuthRedirect)
		oauth.GET("/:provider/callback", h.OAuthCallback)
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	user, tokens, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already exists") ||
			strings.Contains(err.Error(), "passwords do not match") ||
			strings.Contains(err.Error(), "password validation failed") {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"error":   "Registration failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
		"tokens": tokens,
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	user, tokens, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid credentials") ||
			strings.Contains(err.Error(), "user account is disabled") {
			status = http.StatusUnauthorized
		}

		c.JSON(status, gin.H{
			"error":   "Login failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":              user.ID,
			"name":            user.Name,
			"email":           user.Email,
			"role":            user.Role,
			"current_team_id": user.CurrentTeamID,
			"last_login":      user.LastLogin,
		},
		"tokens": tokens,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get user ID from JWT claims
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No authentication found"})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Logout failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// Refresh handles JWT refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	tokens, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Token refresh failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"tokens":  tokens,
	})
}

// OAuthRedirect redirects to OAuth provider
func (h *AuthHandler) OAuthRedirect(c *gin.Context) {
	provider := auth.OAuthProviderType(c.Param("provider"))

	// Generate state for CSRF protection
	state := generateRandomState()

	// Store state in session (you might want to use a more robust session store)
	c.SetCookie("oauth_state", state, 300, "/", "", false, true) // 5 minutes, httpOnly

	authURL, err := h.oauthService.GetAuthURL(provider, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "OAuth provider not available",
			"details": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

// OAuthCallback handles OAuth callback
func (h *AuthHandler) OAuthCallback(c *gin.Context) {
	provider := auth.OAuthProviderType(c.Param("provider"))
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization code not provided",
		})
		return
	}

	// Verify state to prevent CSRF attacks
	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState != state {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid state parameter",
		})
		return
	}

	// Clear the state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Exchange code for user info
	oauthUser, err := h.oauthService.ExchangeCode(c.Request.Context(), provider, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "OAuth authentication failed",
			"details": err.Error(),
		})
		return
	}

	// Convert to AuthService format
	userInfo := &services.OAuthUserInfo{
		Provider: oauthUser.Provider,
		ID:       oauthUser.ID,
		Name:     oauthUser.Name,
		Email:    oauthUser.Email,
		Avatar:   oauthUser.Avatar,
	}

	// Login or register user
	user, tokens, err := h.authService.OAuthLogin(c.Request.Context(), userInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "OAuth login failed",
			"details": err.Error(),
		})
		return
	}

	// For web requests, redirect to dashboard
	if c.GetHeader("Accept") == "text/html" || c.Query("web") == "true" {
		// Set tokens in secure cookies for web sessions
		c.SetCookie("access_token", tokens.AccessToken, int(tokens.ExpiresIn), "/", "", true, true)
		c.SetCookie("refresh_token", tokens.RefreshToken, 7*24*3600, "/", "", true, true) // 7 days
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	// For API requests, return JSON
	c.JSON(http.StatusOK, gin.H{
		"message": "OAuth login successful",
		"user": gin.H{
			"id":              user.ID,
			"name":            user.Name,
			"email":           user.Email,
			"role":            user.Role,
			"current_team_id": user.CurrentTeamID,
		},
		"tokens": tokens,
	})
}

// Helper functions

// getUserIDFromContext extracts user ID from JWT claims in context
func getUserIDFromContext(c *gin.Context) string {
	claims, exists := c.Get("claims")
	if !exists {
		return ""
	}

	jwtClaims, ok := claims.(*auth.Claims)
	if !ok {
		return ""
	}

	return jwtClaims.UserID.String()
}

// generateRandomState generates a random state string for OAuth CSRF protection
func generateRandomState() string {
	// This is a simple implementation - in production, use a more robust method
	return "state_" + strings.Replace(strings.Replace(http.StatusText(http.StatusOK), " ", "_", -1), "-", "_", -1) + "_12345"
}
