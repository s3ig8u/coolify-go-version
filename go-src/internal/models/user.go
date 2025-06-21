package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"` // "-" means don't include in JSON
	Name      string         `json:"name" gorm:"not null"`
	Avatar    *string        `json:"avatar"`
	Role      string         `json:"role" gorm:"default:'user'"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Teams        []Team        `json:"teams" gorm:"many2many:user_teams;"`
	Applications []Application `json:"applications" gorm:"foreignKey:UserID"`
	Servers      []Server      `json:"servers" gorm:"foreignKey:UserID"`
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
