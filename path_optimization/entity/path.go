package entity

import "time"

// Path 表示导航路径
type Path struct {
	ID            uint64         `gorm:"primaryKey" json:"id"`
	UserID        uint64         `gorm:"not null;index" json:"user_id"`
	StartPoint    Location       `gorm:"embedded;embeddedPrefix:start_" json:"start_point"`
	EndPoint      Location       `gorm:"embedded;embeddedPrefix:end_" json:"end_point"`
	Mode          NavigationMode `gorm:"not null" json:"mode"`
	Distance      float64        `json:"distance"`
	EstimatedTime int64          `json:"estimated_time"`
	Points        []Location     `gorm:"-" json:"points"`
	PathPoints    []PathPoint    `gorm:"foreignKey:PathID" json:"path_points"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// PathPoint 路径点
type PathPoint struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	PathID    uint64    `gorm:"not null;index" json:"path_id"`
	Location  Location  `gorm:"embedded" json:"location"`
	Order     int       `gorm:"not null" json:"order"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// NavigationMode 导航模式
type NavigationMode string

const (
	NavigationModeWalking  NavigationMode = "walking"
	NavigationModeCycling  NavigationMode = "cycling"
	NavigationModeDisabled NavigationMode = "disabled"
)
