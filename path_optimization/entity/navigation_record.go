package entity

import "time"

// NavigationRecord 导航历史记录
type NavigationRecord struct {
	ID                   uint64         `gorm:"primaryKey" json:"id"`
	UserID               uint64         `gorm:"not null;index" json:"user_id"`
	PathID               uint64         `gorm:"not null;index" json:"path_id"`
	Mode                 NavigationMode `gorm:"not null" json:"mode"`
	StartTime            time.Time      `gorm:"not null" json:"start_time"`
	EndTime              time.Time      `json:"end_time"`
	ActualDistance       float64        `json:"actual_distance"`
	ActualTime           int64          `json:"actual_time"`
	ObstaclesEncountered int            `json:"obstacles_encountered"`
	CreatedAt            time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}
