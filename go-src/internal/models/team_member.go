package models

import (
	"encoding/json"
	"time"

	"database/sql/driver"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// TeamMember represents a user's membership in a team with role and permissions
type TeamMember struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	TeamID      uint           `json:"team_id" gorm:"not null;index"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Role        string         `json:"role" gorm:"default:'member';not null"`
	Permissions JSONB          `json:"permissions" gorm:"default:'{}'"`
	JoinedAt    time.Time      `json:"joined_at" gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Team Team `json:"team" gorm:"foreignKey:TeamID"`
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// JSONB is a custom type for PostgreSQL JSONB fields
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return "null", nil
	}
	bytes, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return gorm.ErrInvalidData
	}

	return json.Unmarshal(bytes, j)
}

// GormDataType gorm common data type
func (JSONB) GormDataType() string {
	return "json"
}

// GormDBDataType returns different types based on database
func (JSONB) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "jsonb"
	case "sqlite":
		return "text"
	default:
		return "json"
	}
}

// TableName specifies the table name for TeamMember
func (TeamMember) TableName() string {
	return "team_members"
}

// BeforeCreate is a GORM hook that runs before creating a team member
func (tm *TeamMember) BeforeCreate(tx *gorm.DB) error {
	if tm.Role == "" {
		tm.Role = "member"
	}
	if tm.Permissions == nil {
		tm.Permissions = JSONB{}
	}
	if tm.JoinedAt.IsZero() {
		tm.JoinedAt = time.Now()
	}
	return nil
}

// Role constants
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleViewer = "viewer"
)

// Permission constants
const (
	PermissionTeamManage   = "team:manage"
	PermissionTeamDelete   = "team:delete"
	PermissionMemberInvite = "member:invite"
	PermissionMemberRemove = "member:remove"
	PermissionMemberUpdate = "member:update"
	PermissionServerManage = "server:manage"
	PermissionAppDeploy    = "app:deploy"
	PermissionAppManage    = "app:manage"
)

// HasPermission checks if the team member has a specific permission
func (tm *TeamMember) HasPermission(permission string) bool {
	// Owner has all permissions
	if tm.Role == RoleOwner {
		return true
	}

	// Check role-based permissions
	switch tm.Role {
	case RoleAdmin:
		return permission != PermissionTeamDelete // Admins can't delete team
	case RoleMember:
		return permission == PermissionAppDeploy || permission == PermissionAppManage
	case RoleViewer:
		return false // Viewers have no write permissions
	}

	// Check custom permissions in JSONB
	if tm.Permissions != nil {
		if perm, exists := tm.Permissions[permission]; exists {
			if enabled, ok := perm.(bool); ok {
				return enabled
			}
		}
	}

	return false
}
