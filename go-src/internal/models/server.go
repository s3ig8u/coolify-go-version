package models

import (
	"time"

	"gorm.io/gorm"
)

// Server represents a deployment server
type Server struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description *string        `json:"description"`
	Host        string         `json:"host" gorm:"not null"`
	Port        int            `json:"port" gorm:"default:22"`
	Username    string         `json:"username" gorm:"not null"`
	SSHKey      string         `json:"ssh_key" gorm:"not null"`
	Status      string         `json:"status" gorm:"default:'offline'"` // online, offline, error
	Type        string         `json:"type" gorm:"default:'vps'"`       // vps, dedicated, cloud
	Provider    *string        `json:"provider"`                        // aws, digitalocean, hetzner, etc.
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Foreign Keys
	UserID uint `json:"user_id" gorm:"not null"`
	TeamID uint `json:"team_id" gorm:"not null"`

	// Relationships
	User         User          `json:"user" gorm:"foreignKey:UserID"`
	Team         Team          `json:"team" gorm:"foreignKey:TeamID"`
	Applications []Application `json:"applications" gorm:"foreignKey:ServerID"`
}

// TableName specifies the table name for Server
func (Server) TableName() string {
	return "servers"
}

// BeforeCreate is a GORM hook that runs before creating a server
func (s *Server) BeforeCreate(tx *gorm.DB) error {
	if s.Status == "" {
		s.Status = "offline"
	}
	if s.Type == "" {
		s.Type = "vps"
	}
	if s.Port == 0 {
		s.Port = 22
	}
	if !s.IsActive {
		s.IsActive = true
	}
	return nil
}
