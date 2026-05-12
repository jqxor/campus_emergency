package entity

import "time"

// Location 表示地理位置信息
type Location struct {
	Latitude  float64   `gorm:"not null" json:"latitude"`
	Longitude float64   `gorm:"not null" json:"longitude"`
	Altitude  float64   `json:"altitude"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// UserLocation 用户位置信息
type UserLocation struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"not null;index" json:"user_id"`
	Location  Location  `gorm:"embedded" json:"location"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
