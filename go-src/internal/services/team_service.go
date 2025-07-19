package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"coolify-go/internal/models"

	"gorm.io/gorm"
)

// TeamService handles team-related business logic
type TeamService struct {
	db *gorm.DB
}

// NewTeamService creates a new team service
func NewTeamService(db *gorm.DB) *TeamService {
	return &TeamService{db: db}
}

// CreateTeamRequest represents the request to create a team
type CreateTeamRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=50"`
	Description string `json:"description" validate:"max=255"`
}

// CreateTeam creates a new team with the user as owner
func (s *TeamService) CreateTeam(ctx context.Context, req *CreateTeamRequest, ownerID uint) (*models.Team, error) {
	// Generate slug from name
	slug := s.generateSlug(req.Name)

	// Check if slug is unique
	var existingTeam models.Team
	if err := s.db.Where("slug = ?", slug).First(&existingTeam).Error; err == nil {
		// Slug exists, append number
		slug = s.generateUniqueSlug(slug)
	}

	// Create team
	team := &models.Team{
		Name:         req.Name,
		Description:  &req.Description,
		Slug:         slug,
		PersonalTeam: false,
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
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// Add owner as team member
	member := &models.TeamMember{
		TeamID:      team.ID,
		UserID:      ownerID,
		Role:        models.RoleOwner,
		Permissions: models.JSONB{},
	}

	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to add owner to team: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Load team with relationships
	if err := s.db.Preload("Members.User").First(team, team.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load team: %w", err)
	}

	return team, nil
}

// GetTeamsByUser returns all teams that a user is a member of
func (s *TeamService) GetTeamsByUser(ctx context.Context, userID uint) ([]*models.Team, error) {
	var teams []*models.Team

	err := s.db.
		Joins("JOIN team_members ON teams.id = team_members.team_id").
		Where("team_members.user_id = ? AND teams.is_active = true", userID).
		Preload("Members.User").
		Find(&teams).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get teams for user: %w", err)
	}

	return teams, nil
}

// GetTeam returns a team by ID if the user has access
func (s *TeamService) GetTeam(ctx context.Context, teamID, userID uint) (*models.Team, error) {
	var team models.Team

	err := s.db.
		Joins("JOIN team_members ON teams.id = team_members.team_id").
		Where("teams.id = ? AND team_members.user_id = ? AND teams.is_active = true", teamID, userID).
		Preload("Members.User").
		Preload("Invitations").
		First(&team).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("team not found or access denied")
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return &team, nil
}

// UpdateTeamRequest represents the request to update a team
type UpdateTeamRequest struct {
	Name              string `json:"name" validate:"required,min=3,max=50"`
	Description       string `json:"description" validate:"max=255"`
	CustomServerLimit *int   `json:"custom_server_limit" validate:"omitempty,min=0"`
}

// UpdateTeam updates team information
func (s *TeamService) UpdateTeam(ctx context.Context, teamID uint, req *UpdateTeamRequest, userID uint) (*models.Team, error) {
	// Check if user has permission to update team
	if !s.hasTeamPermission(teamID, userID, models.PermissionTeamManage) {
		return nil, fmt.Errorf("permission denied")
	}

	var team models.Team
	if err := s.db.First(&team, teamID).Error; err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	// Update team fields
	team.Name = req.Name
	team.Description = &req.Description
	team.CustomServerLimit = req.CustomServerLimit

	if err := s.db.Save(&team).Error; err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	// Load updated team with relationships
	if err := s.db.Preload("Members.User").First(&team, team.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load updated team: %w", err)
	}

	return &team, nil
}

// DeleteTeam deletes a team (only owner can delete)
func (s *TeamService) DeleteTeam(ctx context.Context, teamID, userID uint) error {
	// Check if user is owner
	var member models.TeamMember
	err := s.db.Where("team_id = ? AND user_id = ? AND role = ?", teamID, userID, models.RoleOwner).First(&member).Error
	if err != nil {
		return fmt.Errorf("only team owner can delete team")
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete team members
	if err := tx.Where("team_id = ?", teamID).Delete(&models.TeamMember{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete team members: %w", err)
	}

	// Delete team invitations
	if err := tx.Where("team_id = ?", teamID).Delete(&models.TeamInvitation{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete team invitations: %w", err)
	}

	// Delete team
	if err := tx.Delete(&models.Team{}, teamID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete team: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// AddMemberRequest represents the request to add a member to team
type AddMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin member viewer"`
}

// AddMember adds a member to the team
func (s *TeamService) AddMember(ctx context.Context, teamID uint, req *AddMemberRequest, inviterID uint) (*models.TeamMember, error) {
	// Check if user has permission to invite members
	if !s.hasTeamPermission(teamID, inviterID, models.PermissionMemberInvite) {
		return nil, fmt.Errorf("permission denied")
	}

	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found with email: %s", req.Email)
	}

	// Check if user is already a member
	var existingMember models.TeamMember
	if err := s.db.Where("team_id = ? AND user_id = ?", teamID, user.ID).First(&existingMember).Error; err == nil {
		return nil, fmt.Errorf("user is already a team member")
	}

	// Create team member
	member := &models.TeamMember{
		TeamID:      teamID,
		UserID:      user.ID,
		Role:        req.Role,
		Permissions: models.JSONB{},
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// Load member with user data
	if err := s.db.Preload("User").First(member, member.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load member: %w", err)
	}

	return member, nil
}

// UpdateMemberRequest represents the request to update a member's role
type UpdateMemberRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member viewer"`
}

// UpdateMember updates a team member's role
func (s *TeamService) UpdateMember(ctx context.Context, teamID, memberID uint, req *UpdateMemberRequest, updaterID uint) (*models.TeamMember, error) {
	// Check if user has permission to update members
	if !s.hasTeamPermission(teamID, updaterID, models.PermissionMemberUpdate) {
		return nil, fmt.Errorf("permission denied")
	}

	var member models.TeamMember
	if err := s.db.Where("id = ? AND team_id = ?", memberID, teamID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	// Don't allow changing owner role
	if member.Role == models.RoleOwner {
		return nil, fmt.Errorf("cannot change owner role")
	}

	// Update role
	member.Role = req.Role
	if err := s.db.Save(&member).Error; err != nil {
		return nil, fmt.Errorf("failed to update member: %w", err)
	}

	// Load member with user data
	if err := s.db.Preload("User").First(&member, member.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load member: %w", err)
	}

	return &member, nil
}

// RemoveMember removes a member from the team
func (s *TeamService) RemoveMember(ctx context.Context, teamID, memberID uint, removerID uint) error {
	// Check if user has permission to remove members
	if !s.hasTeamPermission(teamID, removerID, models.PermissionMemberRemove) {
		return fmt.Errorf("permission denied")
	}

	var member models.TeamMember
	if err := s.db.Where("id = ? AND team_id = ?", memberID, teamID).First(&member).Error; err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	// Don't allow removing owner
	if member.Role == models.RoleOwner {
		return fmt.Errorf("cannot remove team owner")
	}

	// Delete member
	if err := s.db.Delete(&member).Error; err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

// Helper methods

// generateSlug generates a URL-friendly slug from team name
func (s *TeamService) generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// generateUniqueSlug generates a unique slug by appending a number
func (s *TeamService) generateUniqueSlug(baseSlug string) string {
	for i := 2; i <= 100; i++ {
		slug := fmt.Sprintf("%s-%d", baseSlug, i)
		var team models.Team
		if err := s.db.Where("slug = ?", slug).First(&team).Error; err != nil {
			return slug
		}
	}
	// Fallback to timestamp
	return fmt.Sprintf("%s-%d", baseSlug, time.Now().Unix())
}

// hasTeamPermission checks if user has a specific permission in the team
func (s *TeamService) hasTeamPermission(teamID, userID uint, permission string) bool {
	var member models.TeamMember
	err := s.db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error
	if err != nil {
		return false
	}

	return member.HasPermission(permission)
}
