package models

import (
	"time"

	"gorm.io/gorm"
)

// Team represents a team/organization in the system
type Team struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description *string        `json:"description"`
	Slug        string         `json:"slug" gorm:"uniqueIndex;not null"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Users        []User        `json:"users" gorm:"many2many:user_teams;"`
	Applications []Application `json:"applications" gorm:"foreignKey:TeamID"`
	Servers      []Server      `json:"servers" gorm:"foreignKey:TeamID"`
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
	return nil
}
