package models

import (
	"time"

	"gorm.io/gorm"
)

// Team represents a team/organization in the system
type Team struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	Name              string         `json:"name" gorm:"not null"`
	Description       *string        `json:"description"`
	Slug              string         `json:"slug" gorm:"uniqueIndex;not null"`
	PersonalTeam      bool           `json:"personal_team" gorm:"default:false"`
	CustomServerLimit *int           `json:"custom_server_limit"`
	ShowBoarding      bool           `json:"show_boarding" gorm:"default:true"`
	IsActive          bool           `json:"is_active" gorm:"default:true"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Members      []TeamMember     `json:"members" gorm:"foreignKey:TeamID"`
	Invitations  []TeamInvitation `json:"invitations" gorm:"foreignKey:TeamID"`
	Applications []Application    `json:"applications" gorm:"foreignKey:TeamID"`
	Servers      []Server         `json:"servers" gorm:"foreignKey:TeamID"`
}

// TableName specifies the table name for Team
func (Team) TableName() string {
	return "teams"
}

// BeforeCreate is a GORM hook that runs before creating a team
func (t *Team) BeforeCreate(tx *gorm.DB) error {
	if !t.IsActive {
		t.IsActive = true
	}
	if t.ShowBoarding {
		t.ShowBoarding = true
	}
	return nil
}

// GetServerLimit returns the server limit for this team
func (t *Team) GetServerLimit() int {
	if t.CustomServerLimit != nil {
		return *t.CustomServerLimit
	}
	// Default limit for self-hosted
	return 999999999
}

// ServerLimitReached checks if the team has reached its server limit
func (t *Team) ServerLimitReached() bool {
	return len(t.Servers) >= t.GetServerLimit()
}

// IsEmpty checks if the team has no resources
func (t *Team) IsEmpty() bool {
	return len(t.Applications) == 0 && len(t.Servers) == 0
}

// GetOwner returns the team owner (member with owner role)
func (t *Team) GetOwner() *TeamMember {
	for _, member := range t.Members {
		if member.Role == RoleOwner {
			return &member
		}
	}
	return nil
}

// HasMember checks if a user is a member of this team
func (t *Team) HasMember(userID uint) bool {
	for _, member := range t.Members {
		if member.UserID == userID {
			return true
		}
	}
	return false
}

// GetMemberRole returns the role of a user in this team
func (t *Team) GetMemberRole(userID uint) string {
	for _, member := range t.Members {
		if member.UserID == userID {
			return member.Role
		}
	}
	return ""
}
