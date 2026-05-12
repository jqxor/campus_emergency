package entity

import "time"

type Role struct {
	ID          uint64       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string       `gorm:"type:varchar(50);not null;unique" json:"name"`
	Description string       `gorm:"type:text" json:"description"`
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
	IsActive    bool         `gorm:"default:true" json:"is_active"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

type Permission struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Module      string    `gorm:"type:varchar(50);not null" json:"module"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Roles       []Role    `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}

type RolePermission struct {
	RoleID       uint64    `gorm:"primaryKey" json:"role_id"`
	PermissionID uint64    `gorm:"primaryKey" json:"permission_id"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type UserRole struct {
	UserID    uint64    `gorm:"primaryKey" json:"user_id"`
	RoleID    uint64    `gorm:"primaryKey" json:"role_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type RolePermissionSummary struct {
	RoleID          uint64   `json:"role_id"`
	RoleName        string   `json:"role_name"`
	PermissionCount int      `json:"permission_count"`
	Modules         []string `json:"modules"`
}
