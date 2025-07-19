package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"coolify-go/internal/auth"
	"coolify-go/internal/models"
	"coolify-go/internal/utils"

	"gorm.io/gorm"
)

// UserService handles user management
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// GetByID returns a user by ID
func (s *UserService) GetByID(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	err := s.db.Preload("TeamMemberships.Team").First(&user, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetByEmail returns a user by email
func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.db.Where("email = ?", strings.ToLower(email)).Preload("TeamMemberships.Team").First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// ListUsers returns a paginated list of users
func (s *UserService) ListUsers(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// Get total count
	if err := s.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users with pagination
	err := s.db.Offset(offset).Limit(limit).Preload("TeamMemberships.Team").Find(&users).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// CreateUserRequest represents user creation request
type CreateUserRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=50"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	Role            string `json:"role" validate:"oneof=admin user"`
	MarketingEmails bool   `json:"marketing_emails"`
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*models.User, error) {
	// Check if email already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", strings.ToLower(req.Email)).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Validate password strength
	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set default role
	role := req.Role
	if role == "" {
		role = "user"
	}

	// Create user
	user := &models.User{
		Name:            req.Name,
		Email:           strings.ToLower(req.Email),
		Password:        hashedPassword,
		Role:            role,
		IsActive:        true,
		MarketingEmails: req.MarketingEmails,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	Name            *string `json:"name" validate:"omitempty,min=2,max=50"`
	Email           *string `json:"email" validate:"omitempty,email"`
	Role            *string `json:"role" validate:"omitempty,oneof=admin user"`
	IsActive        *bool   `json:"is_active"`
	MarketingEmails *bool   `json:"marketing_emails"`
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, userID uint, req *UpdateUserRequest) (*models.User, error) {
	// Get existing user
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Email != nil {
		email := strings.ToLower(*req.Email)
		// Check if email is already in use by another user
		var existingUser models.User
		if err := s.db.Where("email = ? AND id != ?", email, userID).First(&existingUser).Error; err == nil {
			return nil, fmt.Errorf("email %s is already in use", email)
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check existing email: %w", err)
		}
		user.Email = email
	}

	if req.Role != nil {
		user.Role = *req.Role
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if req.MarketingEmails != nil {
		user.MarketingEmails = *req.MarketingEmails
	}

	// Save user
	if err := s.db.Save(user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=50"`
	MarketingEmails bool   `json:"marketing_emails"`
}

// UpdateProfile updates user profile (self-service)
func (s *UserService) UpdateProfile(ctx context.Context, userID uint, req *UpdateProfileRequest) (*models.User, error) {
	// Get existing user
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update profile fields
	user.Name = req.Name
	user.MarketingEmails = req.MarketingEmails

	// Save user
	if err := s.db.Save(user).Error; err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return user, nil
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID uint) error {
	// Check if user exists
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Remove user from all teams (except personal team)
	if err := tx.Where("user_id = ?", userID).Delete(&models.TeamMember{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove team memberships: %w", err)
	}

	// Delete user's personal team
	if err := tx.Where("personal_team = true AND id IN (SELECT team_id FROM team_members WHERE user_id = ?)", userID).Delete(&models.Team{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete personal team: %w", err)
	}

	// Soft delete user
	if err := tx.Delete(user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8"`
}

// ChangePassword changes user password (authenticated)
func (s *UserService) ChangePassword(ctx context.Context, userID uint, req *ChangePasswordRequest) error {
	// Validate password confirmation
	if req.NewPassword != req.ConfirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	// Validate password strength
	if err := auth.ValidatePasswordStrength(req.NewPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Get user
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if !auth.CheckPassword(req.CurrentPassword, user.Password) {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.Password = hashedPassword
	if err := s.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// PasswordResetRequest represents password reset request
type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordReset initiates a password reset process
func (s *UserService) PasswordReset(ctx context.Context, req *PasswordResetRequest) error {
	// Check if user exists
	user, err := s.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if user exists or not for security
		return nil
	}

	// Generate reset token
	resetToken := utils.GenerateRandomString(32)

	// Store reset token (in a real implementation, you'd store this in a separate table or cache)
	// For now, we'll just log it - you should implement proper token storage
	fmt.Printf("Password reset token for %s: %s\n", user.Email, resetToken)

	// TODO: Send email with reset link containing the token
	// You should integrate with an email service here

	return nil
}

// ResetPasswordRequest represents password reset completion request
type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8"`
}

// ResetPassword completes password reset with token
func (s *UserService) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	// Validate password confirmation
	if req.NewPassword != req.ConfirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	// Validate password strength
	if err := auth.ValidatePasswordStrength(req.NewPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// TODO: Validate reset token and get associated user
	// In a real implementation, you'd look up the token in your storage
	// For now, return an error as this is not fully implemented
	return fmt.Errorf("password reset token validation not implemented")

	// This is what the implementation would look like:
	/*
		// Get user from token
		user, err := s.getUserFromResetToken(req.Token)
		if err != nil {
			return fmt.Errorf("invalid or expired reset token")
		}

		// Hash new password
		hashedPassword, err := auth.HashPassword(req.NewPassword)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// Update password
		user.Password = hashedPassword
		if err := s.db.Save(user).Error; err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		// Invalidate reset token
		if err := s.invalidateResetToken(req.Token); err != nil {
			return fmt.Errorf("failed to invalidate reset token: %w", err)
		}

		return nil
	*/
}

// SetCurrentTeam sets the user's current active team
func (s *UserService) SetCurrentTeam(ctx context.Context, userID, teamID uint) error {
	// Verify user is a member of the team
	var membership models.TeamMember
	if err := s.db.Where("user_id = ? AND team_id = ?", userID, teamID).First(&membership).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user is not a member of this team")
		}
		return fmt.Errorf("failed to check team membership: %w", err)
	}

	// Update user's current team
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("current_team_id", teamID).Error; err != nil {
		return fmt.Errorf("failed to update current team: %w", err)
	}

	return nil
}

// GetUserTeams returns all teams a user is a member of
func (s *UserService) GetUserTeams(ctx context.Context, userID uint) ([]*models.Team, error) {
	var teams []*models.Team

	err := s.db.
		Joins("JOIN team_members ON teams.id = team_members.team_id").
		Where("team_members.user_id = ? AND teams.is_active = true", userID).
		Preload("Members.User").
		Find(&teams).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user teams: %w", err)
	}

	return teams, nil
}

// GetUserStats returns user statistics
type UserStats struct {
	TotalApplications int64      `json:"total_applications"`
	TotalServers      int64      `json:"total_servers"`
	TotalTeams        int64      `json:"total_teams"`
	LastLogin         *time.Time `json:"last_login"`
}

// GetUserStats returns statistics for a user
func (s *UserService) GetUserStats(ctx context.Context, userID uint) (*UserStats, error) {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	stats := &UserStats{
		LastLogin: user.LastLogin,
	}

	// Count applications
	if err := s.db.Model(&models.Application{}).Where("user_id = ?", userID).Count(&stats.TotalApplications).Error; err != nil {
		return nil, fmt.Errorf("failed to count applications: %w", err)
	}

	// Count servers
	if err := s.db.Model(&models.Server{}).Where("user_id = ?", userID).Count(&stats.TotalServers).Error; err != nil {
		return nil, fmt.Errorf("failed to count servers: %w", err)
	}

	// Count teams
	if err := s.db.Model(&models.TeamMember{}).Where("user_id = ?", userID).Count(&stats.TotalTeams).Error; err != nil {
		return nil, fmt.Errorf("failed to count teams: %w", err)
	}

	return stats, nil
}
