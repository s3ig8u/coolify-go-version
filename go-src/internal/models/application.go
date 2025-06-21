package models

import (
	"time"

	"gorm.io/gorm"
)

// Application represents a deployed application
type Application struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description *string        `json:"description"`
	Slug        string         `json:"slug" gorm:"uniqueIndex;not null"`
	Type        string         `json:"type" gorm:"not null"`            // docker, static, nodejs, etc.
	Status      string         `json:"status" gorm:"default:'stopped'"` // running, stopped, building, error
	Port        int            `json:"port" gorm:"default:3000"`
	Domain      *string        `json:"domain"`
	GitURL      *string        `json:"git_url"`
	GitBranch   string         `json:"git_branch" gorm:"default:'main'"`
	BuildPack   *string        `json:"build_pack"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Foreign Keys
	UserID   uint `json:"user_id" gorm:"not null"`
	TeamID   uint `json:"team_id" gorm:"not null"`
	ServerID uint `json:"server_id" gorm:"not null"`

	// Relationships
	User   User   `json:"user" gorm:"foreignKey:UserID"`
	Team   Team   `json:"team" gorm:"foreignKey:TeamID"`
	Server Server `json:"server" gorm:"foreignKey:ServerID"`
}

// TableName specifies the table name for Application
func (Application) TableName() string {
	return "applications"
}

// BeforeCreate is a GORM hook that runs before creating an application
func (a *Application) BeforeCreate(tx *gorm.DB) error {
	if a.Status == "" {
		a.Status = "stopped"
	}
	if a.GitBranch == "" {
		a.GitBranch = "main"
	}
	if a.Port == 0 {
		a.Port = 3000
	}
	if !a.IsActive {
		a.IsActive = true
	}
	return nil
}
