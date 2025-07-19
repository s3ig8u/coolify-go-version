package services

import (
	"context"
	"fmt"
	"time"

	"coolify-go/internal/models"

	"gorm.io/gorm"
)

// InvitationService handles team invitation logic
type InvitationService struct {
	db *gorm.DB
}

// NewInvitationService creates a new invitation service
func NewInvitationService(db *gorm.DB) *InvitationService {
	return &InvitationService{db: db}
}

// CreateInvitationRequest represents the request to create an invitation
type CreateInvitationRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin member viewer"`
}

// CreateInvitation creates a new team invitation
func (s *InvitationService) CreateInvitation(ctx context.Context, teamID uint, req *CreateInvitationRequest, inviterID uint) (*models.TeamInvitation, error) {
	// Check if inviter has permission
	var inviter models.TeamMember
	err := s.db.Where("team_id = ? AND user_id = ?", teamID, inviterID).First(&inviter).Error
	if err != nil {
		return nil, fmt.Errorf("access denied")
	}

	if !inviter.HasPermission(models.PermissionMemberInvite) {
		return nil, fmt.Errorf("permission denied")
	}

	// Check if user is already a member
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		var existingMember models.TeamMember
		if err := s.db.Where("team_id = ? AND user_id = ?", teamID, existingUser.ID).First(&existingMember).Error; err == nil {
			return nil, fmt.Errorf("user is already a team member")
		}
	}

	// Check if there's already a pending invitation
	var existingInvitation models.TeamInvitation
	if err := s.db.Where("team_id = ? AND email = ? AND accepted_at IS NULL AND expires_at > ?", teamID, req.Email, time.Now()).First(&existingInvitation).Error; err == nil {
		return nil, fmt.Errorf("invitation already exists for this email")
	}

	// Create invitation
	invitation := &models.TeamInvitation{
		TeamID:    teamID,
		Email:     req.Email,
		Role:      req.Role,
		Via:       "email",
		ExpiresAt: time.Now().AddDate(0, 0, 7), // 7 days from now
	}

	if err := s.db.Create(invitation).Error; err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Load invitation with team data
	if err := s.db.Preload("Team").First(invitation, invitation.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load invitation: %w", err)
	}

	return invitation, nil
}

// GetInvitation returns an invitation by UUID
func (s *InvitationService) GetInvitation(ctx context.Context, uuid string) (*models.TeamInvitation, error) {
	var invitation models.TeamInvitation

	err := s.db.Where("uuid = ?", uuid).Preload("Team").First(&invitation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invitation not found")
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	return &invitation, nil
}

// AcceptInvitation accepts a team invitation
func (s *InvitationService) AcceptInvitation(ctx context.Context, uuid string, userID uint) (*models.TeamMember, error) {
	// Get invitation
	invitation, err := s.GetInvitation(ctx, uuid)
	if err != nil {
		return nil, err
	}

	// Check if invitation is valid
	if !invitation.IsValid() {
		if invitation.IsExpired() {
			return nil, fmt.Errorf("invitation has expired")
		}
		if invitation.IsAccepted() {
			return nil, fmt.Errorf("invitation has already been accepted")
		}
		return nil, fmt.Errorf("invitation is not valid")
	}

	// Get user
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if invitation email matches user email
	if user.Email != invitation.Email {
		return nil, fmt.Errorf("invitation email does not match user email")
	}

	// Check if user is already a member
	var existingMember models.TeamMember
	if err := s.db.Where("team_id = ? AND user_id = ?", invitation.TeamID, userID).First(&existingMember).Error; err == nil {
		return nil, fmt.Errorf("user is already a team member")
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create team member
	member := &models.TeamMember{
		TeamID:      invitation.TeamID,
		UserID:      userID,
		Role:        invitation.Role,
		Permissions: models.JSONB{},
	}

	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create team member: %w", err)
	}

	// Mark invitation as accepted
	invitation.Accept()
	if err := tx.Save(invitation).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Load member with user and team data
	if err := s.db.Preload("User").Preload("Team").First(member, member.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load member: %w", err)
	}

	return member, nil
}

// RejectInvitation rejects a team invitation
func (s *InvitationService) RejectInvitation(ctx context.Context, uuid string, userID uint) error {
	// Get invitation
	invitation, err := s.GetInvitation(ctx, uuid)
	if err != nil {
		return err
	}

	// Get user
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if invitation email matches user email
	if user.Email != invitation.Email {
		return fmt.Errorf("invitation email does not match user email")
	}

	// Check if invitation is valid
	if !invitation.IsValid() {
		return fmt.Errorf("invitation is not valid")
	}

	// Delete invitation
	if err := s.db.Delete(invitation).Error; err != nil {
		return fmt.Errorf("failed to reject invitation: %w", err)
	}

	return nil
}

// GetTeamInvitations returns all pending invitations for a team
func (s *InvitationService) GetTeamInvitations(ctx context.Context, teamID, userID uint) ([]*models.TeamInvitation, error) {
	// Check if user has access to team
	var member models.TeamMember
	err := s.db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error
	if err != nil {
		return nil, fmt.Errorf("access denied")
	}

	var invitations []*models.TeamInvitation
	err = s.db.Where("team_id = ? AND accepted_at IS NULL", teamID).
		Order("created_at DESC").
		Find(&invitations).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get invitations: %w", err)
	}

	return invitations, nil
}

// GetUserInvitations returns all pending invitations for a user
func (s *InvitationService) GetUserInvitations(ctx context.Context, userEmail string) ([]*models.TeamInvitation, error) {
	var invitations []*models.TeamInvitation

	err := s.db.Where("email = ? AND accepted_at IS NULL AND expires_at > ?", userEmail, time.Now()).
		Preload("Team").
		Order("created_at DESC").
		Find(&invitations).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user invitations: %w", err)
	}

	return invitations, nil
}

// CancelInvitation cancels a team invitation
func (s *InvitationService) CancelInvitation(ctx context.Context, invitationID, teamID, userID uint) error {
	// Check if user has permission
	var member models.TeamMember
	err := s.db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error
	if err != nil {
		return fmt.Errorf("access denied")
	}

	if !member.HasPermission(models.PermissionMemberInvite) {
		return fmt.Errorf("permission denied")
	}

	// Get invitation
	var invitation models.TeamInvitation
	err = s.db.Where("id = ? AND team_id = ?", invitationID, teamID).First(&invitation).Error
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	// Check if invitation is still pending
	if invitation.IsAccepted() {
		return fmt.Errorf("invitation has already been accepted")
	}

	// Delete invitation
	if err := s.db.Delete(&invitation).Error; err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}

	return nil
}

// CleanupExpiredInvitations removes expired invitations
func (s *InvitationService) CleanupExpiredInvitations(ctx context.Context) error {
	result := s.db.Where("expires_at < ? AND accepted_at IS NULL", time.Now()).Delete(&models.TeamInvitation{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired invitations: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		fmt.Printf("Cleaned up %d expired invitations\n", result.RowsAffected)
	}

	return nil
}
