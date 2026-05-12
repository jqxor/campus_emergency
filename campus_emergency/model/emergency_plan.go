package model

import (
	"time"

	"gorm.io/gorm"
)

type PlanStatus string

const (
	PlanStatusDraft       PlanStatus = "draft"
	PlanStatusActive      PlanStatus = "active"
	PlanStatusInactive    PlanStatus = "inactive"
	PlanStatusExpired     PlanStatus = "expired"
	PlanStatusUnderReview PlanStatus = "under_review"
)

type ScenarioType string

const (
	ScenarioTypeFire          ScenarioType = "fire"
	ScenarioTypeEarthquake    ScenarioType = "earthquake"
	ScenarioTypeTerrorAttack  ScenarioType = "terror_attack"
	ScenarioTypeMedical       ScenarioType = "medical"
	ScenarioTypeOther         ScenarioType = "other"
)

type SafePoint struct {
	Latitude  float64 `json:"latitude" gorm:"type:decimal(10,6)"`
	Longitude float64 `json:"longitude" gorm:"type:decimal(10,6)"`
	Name      string  `json:"name"`
	Capacity  int     `json:"capacity"`
}

type PathPoint struct {
	Latitude  float64 `json:"latitude" gorm:"type:decimal(10,6)"`
	Longitude float64 `json:"longitude" gorm:"type:decimal(10,6)"`
	Order     int     `json:"order"`
}

type Obstacle struct {
	Latitude  float64 `json:"latitude" gorm:"type:decimal(10,6)"`
	Longitude float64 `json:"longitude" gorm:"type:decimal(10,6)"`
	Radius    float64 `json:"radius"`
	Type      string  `json:"type"`
}

type AIOptimization struct {
	OriginalPathLength  float64     `json:"original_path_length"`
	OptimizedPathLength float64     `json:"optimized_path_length"`
	SavingTime          float64     `json:"saving_time"`
	Suggestions         []string    `json:"suggestions"`
	OptimizedPath       []PathPoint `json:"optimized_path"`
	CreatedAt           time.Time   `json:"created_at"`
}

type EmergencyPlan struct {
	gorm.Model
	Name           string          `json:"name" gorm:"size:100;not null"`
	ScenarioType   ScenarioType    `json:"scenario_type" gorm:"not null;index"`
	Description    string          `json:"description" gorm:"type:text"`
	Status         PlanStatus      `json:"status" gorm:"default:draft"`
	Priority       int             `json:"priority" gorm:"default:0"`
	CreatorID      uint            `json:"creator_id" gorm:"not null"`
	CreatorName    string          `json:"creator_name" gorm:"size:50"`
	SafePoints     []SafePoint     `json:"safe_points" gorm:"type:json"`
	EvacuationPath []PathPoint     `json:"evacuation_path" gorm:"type:json"`
	Obstacles      []Obstacle      `json:"obstacles" gorm:"type:json"`
	EstimatedTime  float64         `json:"estimated_time"`
	PeopleCount    int             `json:"people_count"`
	AIOptimization *AIOptimization `json:"ai_optimization" gorm:"type:json"`
	LastOptimizedAt *time.Time      `json:"last_optimized_at"`
	Tags           []string        `json:"tags" gorm:"type:json"`
}

type PlanSearchCondition struct {
	ScenarioType string     `json:"scenario_type"`
	Status       PlanStatus `json:"status"`
	Keyword      string     `json:"keyword"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	Page         int        `json:"page" default:"1"`
	PageSize     int        `json:"page_size" default:"10"`
	SortBy       string     `json:"sort_by" default:"created_at"`
	SortOrder    string     `json:"sort_order" default:"desc"`
}

type PlanImportExport struct {
	Version    string          `json:"version" default:"1.0"`
	Plans      []EmergencyPlan `json:"plans"`
	ExportTime time.Time       `json:"export_time"`
	ExportUser string          `json:"export_user"`
}
