package models_test

import (
	"testing"
	"time"

	"coolify-go/internal/database"
	"coolify-go/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupTestDB creates a temporary PostgreSQL database for testing (with SQLite fallback)
func setupTestDB(t *testing.T) *gorm.DB {
	return database.TestDBWithFallback(t)
}

func TestTeamMemberModel(t *testing.T) {
	db := setupTestDB(t)

	// Create test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
		Role:     "user",
		IsActive: true,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create test team
	team := &models.Team{
		Name:         "Test Team",
		Slug:         "test-team",
		PersonalTeam: false,
		IsActive:     true,
	}
	err = db.Create(team).Error
	require.NoError(t, err)

	// Create team member
	member := &models.TeamMember{
		TeamID:      team.ID,
		UserID:      user.ID,
		Role:        models.RoleAdmin,
		Permissions: models.JSONB{"test": true},
	}
	err = db.Create(member).Error
	require.NoError(t, err)

	// Test permissions
	assert.True(t, member.HasPermission(models.PermissionTeamManage))
	assert.False(t, member.HasPermission(models.PermissionTeamDelete)) // Admins can't delete team
	assert.NotZero(t, member.JoinedAt)

	// Test relationships
	var foundMember models.TeamMember
	err = db.Preload("Team").Preload("User").First(&foundMember, member.ID).Error
	require.NoError(t, err)
	assert.Equal(t, team.Name, foundMember.Team.Name)
	assert.Equal(t, user.Email, foundMember.User.Email)
}

func TestTeamInvitationModel(t *testing.T) {
	db := setupTestDB(t)

	// Create test team
	team := &models.Team{
		Name:     "Test Team",
		Slug:     "test-team",
		IsActive: true,
	}
	err := db.Create(team).Error
	require.NoError(t, err)

	// Create invitation
	invitation := &models.TeamInvitation{
		TeamID:    team.ID,
		Email:     "invite@example.com",
		Role:      models.RoleMember,
		Via:       "email",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	err = db.Create(invitation).Error
	require.NoError(t, err)

	// Test invitation validity
	assert.True(t, invitation.IsValid())
	assert.False(t, invitation.IsExpired())
	assert.False(t, invitation.IsAccepted())
	assert.NotEmpty(t, invitation.UUID)

	// Test accepting invitation
	invitation.Accept()
	assert.True(t, invitation.IsAccepted())
	assert.False(t, invitation.IsValid()) // No longer valid after acceptance
}

func TestUserTeamRelationships(t *testing.T) {
	db := setupTestDB(t)

	// Create test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
		Role:     "user",
		IsActive: true,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create test teams
	team1 := &models.Team{
		Name:     "Team 1",
		Slug:     "team-1",
		IsActive: true,
	}
	err = db.Create(team1).Error
	require.NoError(t, err)

	team2 := &models.Team{
		Name:     "Team 2",
		Slug:     "team-2",
		IsActive: true,
	}
	err = db.Create(team2).Error
	require.NoError(t, err)

	// Add user to teams with different roles
	member1 := &models.TeamMember{
		TeamID: team1.ID,
		UserID: user.ID,
		Role:   models.RoleOwner,
	}
	err = db.Create(member1).Error
	require.NoError(t, err)

	member2 := &models.TeamMember{
		TeamID: team2.ID,
		UserID: user.ID,
		Role:   models.RoleMember,
	}
	err = db.Create(member2).Error
	require.NoError(t, err)

	// Load user with team memberships
	var foundUser models.User
	err = db.Preload("TeamMemberships.Team").First(&foundUser, user.ID).Error
	require.NoError(t, err)

	// Test user team methods
	assert.True(t, foundUser.IsTeamMember(team1.ID))
	assert.True(t, foundUser.IsTeamMember(team2.ID))
	assert.Equal(t, models.RoleOwner, foundUser.GetTeamRole(team1.ID))
	assert.Equal(t, models.RoleMember, foundUser.GetTeamRole(team2.ID))
	assert.True(t, foundUser.HasTeamRole(team1.ID, models.RoleOwner))
	assert.False(t, foundUser.HasTeamRole(team2.ID, models.RoleOwner))
	assert.True(t, foundUser.HasTeamPermission(team1.ID, models.PermissionTeamDelete))
	assert.False(t, foundUser.HasTeamPermission(team2.ID, models.PermissionTeamDelete))
}

func TestTeamModel(t *testing.T) {
	db := setupTestDB(t)

	// Create test team
	team := &models.Team{
		Name:              "Test Team",
		Slug:              "test-team",
		PersonalTeam:      false,
		CustomServerLimit: nil, // Use default
		IsActive:          true,
	}
	err := db.Create(team).Error
	require.NoError(t, err)

	// Test server limit
	assert.Equal(t, 999999999, team.GetServerLimit()) // Default limit
	assert.False(t, team.ServerLimitReached())        // No servers yet

	// Test with custom limit
	customLimit := 5
	team.CustomServerLimit = &customLimit
	assert.Equal(t, 5, team.GetServerLimit())

	// Test team methods
	assert.True(t, team.IsEmpty()) // No applications or servers
}

func TestJSONBType(t *testing.T) {
	db := setupTestDB(t)

	// Create required user and team first
	user := database.CreateTestUser(t, db)
	team := database.CreateTestTeam(t, db)

	// Test JSONB serialization
	permissions := models.JSONB{
		"deploy":     true,
		"manage":     false,
		"admin_only": []string{"delete", "modify"},
		"meta": map[string]interface{}{
			"created_by": "system",
			"version":    1,
		},
	}

	member := &models.TeamMember{
		TeamID:      team.ID,
		UserID:      user.ID,
		Role:        models.RoleMember,
		Permissions: permissions,
	}

	// This will test the Value() method
	err := db.Create(member).Error
	require.NoError(t, err)

	// This will test the Scan() method
	var foundMember models.TeamMember
	err = db.First(&foundMember, member.ID).Error
	require.NoError(t, err)

	// Verify JSONB data integrity
	assert.Equal(t, true, foundMember.Permissions["deploy"])
	assert.Equal(t, false, foundMember.Permissions["manage"])
	assert.IsType(t, []interface{}{}, foundMember.Permissions["admin_only"])
	assert.IsType(t, map[string]interface{}{}, foundMember.Permissions["meta"])
}
