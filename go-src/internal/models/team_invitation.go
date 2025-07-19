package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

// TeamInvitation represents an invitation to join a team
type TeamInvitation struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	TeamID     uint           `json:"team_id" gorm:"not null;index"`
	UUID       string         `json:"uuid" gorm:"uniqueIndex;not null"`
	Email      string         `json:"email" gorm:"not null"`
	Role       string         `json:"role" gorm:"default:'member';not null"`
	Link       *string        `json:"link"`
	Via        string         `json:"via" gorm:"default:'email'"` // email, link, etc.
	ExpiresAt  time.Time      `json:"expires_at" gorm:"not null"`
	AcceptedAt *time.Time     `json:"accepted_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Team Team `json:"team" gorm:"foreignKey:TeamID"`
}

// TableName specifies the table name for TeamInvitation
func (TeamInvitation) TableName() string {
	return "team_invitations"
}

// BeforeCreate is a GORM hook that runs before creating an invitation
func (ti *TeamInvitation) BeforeCreate(tx *gorm.DB) error {
	if ti.UUID == "" {
		ti.UUID = generateInvitationUUID()
	}
	if ti.Role == "" {
		ti.Role = "member"
	}
	if ti.Via == "" {
		ti.Via = "email"
	}
	if ti.ExpiresAt.IsZero() {
		// Default expiration: 7 days from now
		ti.ExpiresAt = time.Now().AddDate(0, 0, 7)
	}
	return nil
}

// IsValid checks if the invitation is still valid (not expired and not accepted)
func (ti *TeamInvitation) IsValid() bool {
	return time.Now().Before(ti.ExpiresAt) && ti.AcceptedAt == nil
}

// IsExpired checks if the invitation has expired
func (ti *TeamInvitation) IsExpired() bool {
	return time.Now().After(ti.ExpiresAt)
}

// IsAccepted checks if the invitation has been accepted
func (ti *TeamInvitation) IsAccepted() bool {
	return ti.AcceptedAt != nil
}

// Accept marks the invitation as accepted
func (ti *TeamInvitation) Accept() {
	now := time.Now()
	ti.AcceptedAt = &now
}

// generateInvitationUUID generates a unique UUID for invitations
func generateInvitationUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
