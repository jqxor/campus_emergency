package entity

import "time"

type User struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"type:varchar(64);not null;unique" json:"username"`
	Password  string    `gorm:"type:varchar(128);not null" json:"password,omitempty"`
	Email     string    `gorm:"type:varchar(128);not null;unique" json:"email"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type UserPermission struct {
	UserID      uint64    `gorm:"primaryKey" json:"user_id"`
	Permission  string    `gorm:"type:text;not null" json:"permission"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type SessionToken struct {
	Token     string    `gorm:"primaryKey;type:varchar(128)" json:"token"`
	UserID    uint64    `gorm:"index;not null" json:"user_id"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
