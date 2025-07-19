package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"coolify-go/internal/auth"
	"coolify-go/internal/models"
	"coolify-go/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthService handles authentication business logic
type AuthService struct {
	db          *gorm.DB
	userService *UserService
	jwtManager  *auth.JWTManager
}

// NewAuthService creates a new auth service
func NewAuthService(db *gorm.DB, userService *UserService, jwtSecret string) *AuthService {
	return &AuthService{
		db:          db,
		userService: userService,
		jwtManager:  auth.NewJWTManager(jwtSecret),
	}
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=50"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8"`
	AgreeToTerms    bool   `json:"agree_to_terms" validate:"required"`
	MarketingEmails bool   `json:"marketing_emails"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*models.User, *auth.TokenPair, error) {
	// Validate password confirmation
	if req.Password != req.ConfirmPassword {
		return nil, nil, errors.New("passwords do not match")
	}

	// Validate password strength
	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		return nil, nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Check if email already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", strings.ToLower(req.Email)).First(&existingUser).Error; err == nil {
		return nil, nil, errors.New("user with this email already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Name:            req.Name,
		Email:           strings.ToLower(req.Email),
		Password:        hashedPassword,
		Role:            "user",
		IsActive:        true,
		MarketingEmails: req.MarketingEmails,
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create user
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Check if this is the first user (admin)
	var userCount int64
	tx.Model(&models.User{}).Count(&userCount)
	if userCount == 1 {
		user.Role = "admin"
		if err := tx.Save(user).Error; err != nil {
			tx.Rollback()
			return nil, nil, fmt.Errorf("failed to set admin role: %w", err)
		}
	}

	// Create personal team for user
	team := &models.Team{
		Name:         req.Name + "'s Team",
		Slug:         s.generateTeamSlug(req.Name+"'s Team", tx),
		PersonalTeam: true,
		IsActive:     true,
	}

	if err := tx.Create(team).Error; err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to create personal team: %w", err)
	}

	// Add user as owner of personal team
	member := &models.TeamMember{
		TeamID:      team.ID,
		UserID:      user.ID,
		Role:        models.RoleOwner,
		Permissions: models.JSONB{},
	}

	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to add user to personal team: %w", err)
	}

	// Set current team
	user.CurrentTeamID = &team.ID
	if err := tx.Save(user).Error; err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to set current team: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Generate tokens
	userUUID := uuid.New()
	teamUUID := uuid.New()
	tokens, err := s.jwtManager.GenerateTokenPair(userUUID, user.Email, &teamUUID, user.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return user, tokens, nil
}

// Login authenticates a user and returns a token pair
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*models.User, *auth.TokenPair, error) {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", strings.ToLower(req.Email)).Preload("TeamMemberships.Team").First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, errors.New("invalid credentials")
		}
		return nil, nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, nil, errors.New("user account is disabled")
	}

	// Verify password
	if !auth.CheckPassword(req.Password, user.Password) {
		return nil, nil, errors.New("invalid credentials")
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	if err := s.db.Save(&user).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to update last login: %w", err)
	}

	// Get or create personal team
	if user.CurrentTeamID == nil {
		personalTeam := s.getOrCreatePersonalTeam(&user)
		user.CurrentTeamID = &personalTeam.ID
		if err := s.db.Save(&user).Error; err != nil {
			return nil, nil, fmt.Errorf("failed to set current team: %w", err)
		}
	}

	// Generate tokens
	userUUID := uuid.New()
	var teamUUID *uuid.UUID
	if user.CurrentTeamID != nil {
		tUUID := uuid.New()
		teamUUID = &tUUID
	}

	tokens, err := s.jwtManager.GenerateTokenPair(userUUID, user.Email, teamUUID, user.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &user, tokens, nil
}

// OAuthUserInfo represents OAuth user information
type OAuthUserInfo struct {
	Provider string
	ID       string
	Name     string
	Email    string
	Avatar   string
}

// OAuthLogin handles OAuth login/registration
func (s *AuthService) OAuthLogin(ctx context.Context, userInfo *OAuthUserInfo) (*models.User, *auth.TokenPair, error) {
	// Check if user exists by email
	var user models.User
	err := s.db.Where("email = ?", strings.ToLower(userInfo.Email)).Preload("TeamMemberships.Team").First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// Check if registration is enabled (you might want to add this to settings)
		// For now, allow OAuth registration
		return s.createOAuthUser(ctx, userInfo)
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to find user: %w", err)
	}

	// User exists, update info and login
	return s.loginExistingOAuthUser(ctx, &user, userInfo)
}

// createOAuthUser creates a new user from OAuth information
func (s *AuthService) createOAuthUser(ctx context.Context, userInfo *OAuthUserInfo) (*models.User, *auth.TokenPair, error) {
	// Generate a random password for OAuth users
	randomPassword := utils.GenerateRandomString(32)
	hashedPassword, err := auth.HashPassword(randomPassword)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Name:     userInfo.Name,
		Email:    strings.ToLower(userInfo.Email),
		Password: hashedPassword, // OAuth users still need a password field
		Avatar:   &userInfo.Avatar,
		Role:     "user",
		IsActive: true,
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create user
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Check if this is the first user (admin)
	var userCount int64
	tx.Model(&models.User{}).Count(&userCount)
	if userCount == 1 {
		user.Role = "admin"
		if err := tx.Save(user).Error; err != nil {
			tx.Rollback()
			return nil, nil, fmt.Errorf("failed to set admin role: %w", err)
		}
	}

	// Create personal team
	team := &models.Team{
		Name:         user.Name + "'s Team",
		Slug:         s.generateTeamSlug(user.Name+"'s Team", tx),
		PersonalTeam: true,
		IsActive:     true,
	}

	if err := tx.Create(team).Error; err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to create personal team: %w", err)
	}

	// Add user as owner of personal team
	member := &models.TeamMember{
		TeamID:      team.ID,
		UserID:      user.ID,
		Role:        models.RoleOwner,
		Permissions: models.JSONB{},
	}

	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to add user to personal team: %w", err)
	}

	// Set current team
	user.CurrentTeamID = &team.ID
	if err := tx.Save(user).Error; err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to set current team: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Generate tokens
	userUUID := uuid.New()
	teamUUID := uuid.New()
	tokens, err := s.jwtManager.GenerateTokenPair(userUUID, user.Email, &teamUUID, user.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return user, tokens, nil
}

// loginExistingOAuthUser logs in an existing user via OAuth
func (s *AuthService) loginExistingOAuthUser(ctx context.Context, user *models.User, userInfo *OAuthUserInfo) (*models.User, *auth.TokenPair, error) {
	// Check if user is active
	if !user.IsActive {
		return nil, nil, errors.New("user account is disabled")
	}

	// Update user info from OAuth
	user.Name = userInfo.Name
	if userInfo.Avatar != "" {
		user.Avatar = &userInfo.Avatar
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now

	if err := s.db.Save(user).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Ensure user has a personal team
	if user.CurrentTeamID == nil {
		personalTeam := s.getOrCreatePersonalTeam(user)
		user.CurrentTeamID = &personalTeam.ID
		if err := s.db.Save(user).Error; err != nil {
			return nil, nil, fmt.Errorf("failed to set current team: %w", err)
		}
	}

	// Generate tokens
	userUUID := uuid.New()
	var teamUUID *uuid.UUID
	if user.CurrentTeamID != nil {
		tUUID := uuid.New()
		teamUUID = &tUUID
	}

	tokens, err := s.jwtManager.GenerateTokenPair(userUUID, user.Email, teamUUID, user.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return user, tokens, nil
}

// Logout invalidates a session/token
func (s *AuthService) Logout(ctx context.Context, userID string) error {
	// For JWT tokens, we don't need to do anything server-side
	// The client should discard the token
	// In a more sophisticated implementation, you might maintain a blacklist
	return nil
}

// Refresh issues a new token pair from a refresh token
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	return s.jwtManager.RefreshTokenPair(refreshToken)
}

// ValidateToken validates a JWT token and returns claims
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*auth.Claims, error) {
	return s.jwtManager.ValidateToken(tokenString)
}

// Helper methods

// generateTeamSlug generates a unique slug for team
func (s *AuthService) generateTeamSlug(name string, tx *gorm.DB) string {
	baseSlug := utils.Slugify(name)
	slug := baseSlug

	for i := 2; i <= 100; i++ {
		var team models.Team
		if err := tx.Where("slug = ?", slug).First(&team).Error; err != nil {
			break // Slug is available
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, i)
	}

	// Fallback to timestamp if all numbered variants are taken
	if i := 100; i > 0 {
		var team models.Team
		if err := tx.Where("slug = ?", slug).First(&team).Error; err == nil {
			slug = fmt.Sprintf("%s-%d", baseSlug, time.Now().Unix())
		}
	}

	return slug
}

// getOrCreatePersonalTeam ensures user has a personal team
func (s *AuthService) getOrCreatePersonalTeam(user *models.User) *models.Team {
	// Try to find existing personal team
	var personalTeam models.Team
	err := s.db.Joins("JOIN team_members ON teams.id = team_members.team_id").
		Where("team_members.user_id = ? AND teams.personal_team = true", user.ID).
		First(&personalTeam).Error

	if err == nil {
		return &personalTeam // Found existing personal team
	}

	// Create new personal team
	team := &models.Team{
		Name:         user.Name + "'s Team",
		Slug:         s.generateTeamSlug(user.Name+"'s Team", s.db),
		PersonalTeam: true,
		IsActive:     true,
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create team
	if err := tx.Create(team).Error; err != nil {
		tx.Rollback()
		return nil
	}

	// Add user as owner
	member := &models.TeamMember{
		TeamID:      team.ID,
		UserID:      user.ID,
		Role:        models.RoleOwner,
		Permissions: models.JSONB{},
	}

	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		return nil
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil
	}

	return team
}
