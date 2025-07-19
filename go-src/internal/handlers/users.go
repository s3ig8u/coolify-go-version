package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"coolify-go/internal/services"

	"github.com/gin-gonic/gin"
)

// UserHandler provides user-related endpoints
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// RegisterRoutes registers user endpoints
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	// User profile endpoints
	r.GET("/profile", h.Profile)
	r.PUT("/profile", h.UpdateProfile)

	// Password management
	r.POST("/password/change", h.ChangePassword)
	r.POST("/password/reset", h.PasswordReset)
	r.POST("/password/reset/complete", h.ResetPassword)

	// User teams
	r.GET("/teams", h.GetUserTeams)
	r.POST("/current-team", h.SetCurrentTeam)

	// User stats
	r.GET("/stats", h.GetUserStats)

	// Admin endpoints (requires admin role)
	admin := r.Group("/admin", requireAdminRole())
	{
		admin.GET("/users", h.ListUsers)
		admin.POST("/users", h.CreateUser)
		admin.GET("/users/:id", h.GetUser)
		admin.PUT("/users/:id", h.UpdateUser)
		admin.DELETE("/users/:id", h.DeleteUser)
	}
}

// Profile returns current user profile
func (h *UserHandler) Profile(c *gin.Context) {
	userID := getUserIDUintFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":               user.ID,
			"name":             user.Name,
			"email":            user.Email,
			"avatar":           user.Avatar,
			"role":             user.Role,
			"is_active":        user.IsActive,
			"current_team_id":  user.CurrentTeamID,
			"marketing_emails": user.MarketingEmails,
			"last_login":       user.LastLogin,
			"created_at":       user.CreatedAt,
			"updated_at":       user.UpdatedAt,
		},
	})
}

// UpdateProfile handles profile updates
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := getUserIDUintFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update profile",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":               user.ID,
			"name":             user.Name,
			"email":            user.Email,
			"marketing_emails": user.MarketingEmails,
			"updated_at":       user.UpdatedAt,
		},
	})
}

// ChangePassword handles password changes
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := getUserIDUintFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req services.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "passwords do not match") ||
			strings.Contains(err.Error(), "current password is incorrect") ||
			strings.Contains(err.Error(), "password validation failed") {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"error":   "Failed to change password",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// PasswordReset handles password reset requests
func (h *UserHandler) PasswordReset(c *gin.Context) {
	var req services.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if err := h.userService.PasswordReset(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to initiate password reset",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If an account with that email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset completion
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req services.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if err := h.userService.ResetPassword(c.Request.Context(), &req); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "passwords do not match") ||
			strings.Contains(err.Error(), "password validation failed") ||
			strings.Contains(err.Error(), "invalid or expired") {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"error":   "Failed to reset password",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// GetUserTeams returns teams for current user
func (h *UserHandler) GetUserTeams(c *gin.Context) {
	userID := getUserIDUintFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	teams, err := h.userService.GetUserTeams(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get teams",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"teams": teams,
	})
}

// SetCurrentTeam sets the user's current team
func (h *UserHandler) SetCurrentTeam(c *gin.Context) {
	userID := getUserIDUintFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req struct {
		TeamID uint `json:"team_id" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if err := h.userService.SetCurrentTeam(c.Request.Context(), userID, req.TeamID); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not a member") {
			status = http.StatusForbidden
		}

		c.JSON(status, gin.H{
			"error":   "Failed to set current team",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Current team updated successfully",
	})
}

// GetUserStats returns user statistics
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userID := getUserIDUintFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	stats, err := h.userService.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// Admin endpoints

// ListUsers returns a paginated list of users (admin only)
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	users, total, err := h.userService.ListUsers(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CreateUser creates a new user (admin only)
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already exists") ||
			strings.Contains(err.Error(), "password validation failed") {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// GetUser returns user details (admin only)
func (h *UserHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UpdateUser updates a user (admin only)
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), uint(userID), &req)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already in use") {
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"error":   "Failed to update user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    user,
	})
}

// DeleteUser deletes a user (admin only)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), uint(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// Helper functions

// getUserIDUintFromContext extracts user ID as uint from context
func getUserIDUintFromContext(c *gin.Context) uint {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	if id, ok := userID.(uint); ok {
		return id
	}
	return 0
}

// requireAdminRole middleware that requires admin role
func requireAdminRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		if roleStr, ok := role.(string); !ok || roleStr != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
