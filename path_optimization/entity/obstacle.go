package entity

import "time"

// Obstacle 障碍物信息
type Obstacle struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Location    Location  `gorm:"embedded" json:"location"`
	Type        string    `gorm:"not null" json:"type"`
	Severity    int       `gorm:"not null" json:"severity"`
	Description string    `json:"description"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	StartTime   time.Time `gorm:"not null" json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// ObstacleWarning 障碍物警告
type ObstacleWarning struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	UserID      uint64    `gorm:"not null;index" json:"user_id"`
	PathID      uint64    `gorm:"not null;index" json:"path_id"`
	ObstacleID  uint64    `gorm:"not null;index" json:"obstacle_id"`
	IsConfirmed bool      `gorm:"default:false" json:"is_confirmed"`
	IsIgnored   bool      `gorm:"default:false" json:"is_ignored"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
