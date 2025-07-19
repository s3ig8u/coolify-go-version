package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Email           string         `json:"email" gorm:"uniqueIndex;not null"`
	Password        string         `json:"-" gorm:"not null"` // "-" means don't include in JSON
	Name            string         `json:"name" gorm:"not null"`
	Avatar          *string        `json:"avatar"`
	Role            string         `json:"role" gorm:"default:'user'"`
	IsActive        bool           `json:"is_active" gorm:"default:true"`
	CurrentTeamID   *uint          `json:"current_team_id"`
	LastLogin       *time.Time     `json:"last_login"`
	MarketingEmails bool           `json:"marketing_emails" gorm:"default:false"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	TeamMemberships []TeamMember  `json:"team_memberships" gorm:"foreignKey:UserID"`
	Applications    []Application `json:"applications" gorm:"foreignKey:UserID"`
	Servers         []Server      `json:"servers" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// BeforeCreate is a GORM hook that runs before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = "user"
	}
	if !u.IsActive {
		u.IsActive = true
	}
	return nil
}

// GetTeams returns all teams the user is a member of
func (u *User) GetTeams() []Team {
	var teams []Team
	for _, membership := range u.TeamMemberships {
		teams = append(teams, membership.Team)
	}
	return teams
}

// HasTeamRole checks if the user has a specific role in a team
func (u *User) HasTeamRole(teamID uint, role string) bool {
	for _, membership := range u.TeamMemberships {
		if membership.TeamID == teamID && membership.Role == role {
			return true
		}
	}
	return false
}

// HasTeamPermission checks if the user has a specific permission in a team
func (u *User) HasTeamPermission(teamID uint, permission string) bool {
	for _, membership := range u.TeamMemberships {
		if membership.TeamID == teamID {
			return membership.HasPermission(permission)
		}
	}
	return false
}

// IsTeamMember checks if the user is a member of a specific team
func (u *User) IsTeamMember(teamID uint) bool {
	for _, membership := range u.TeamMemberships {
		if membership.TeamID == teamID {
			return true
		}
	}
	return false
}

// GetTeamRole returns the user's role in a specific team
func (u *User) GetTeamRole(teamID uint) string {
	for _, membership := range u.TeamMemberships {
		if membership.TeamID == teamID {
			return membership.Role
		}
	}
	return ""
}
