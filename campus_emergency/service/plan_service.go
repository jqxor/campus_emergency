package service

import "campus-emergency/model"

type AIOptimizationService interface {
	OptimizePath(plan *model.EmergencyPlan) (*model.AIOptimization, error)
}

type PlanService interface {
	CreatePlan(plan *model.EmergencyPlan) error
	UpdatePlan(plan *model.EmergencyPlan) error
	DeletePlan(id uint) error
	GetPlanByID(id uint) (*model.EmergencyPlan, error)
	GetPlansByScenarioType(scenarioType model.ScenarioType) ([]*model.EmergencyPlan, error)
	SearchPlans(condition *model.PlanSearchCondition) ([]*model.EmergencyPlan, int64, error)
	GetScenarioTypeStats() (map[model.ScenarioType]int64, error)
	UpdatePlanStatus(id uint, status model.PlanStatus) error
	OptimizePlanPath(id uint) (*model.EmergencyPlan, error)
	ImportPlans(jsonData []byte, userID uint, userName string) (int, error)
	ExportPlans(condition *model.PlanSearchCondition) (*model.PlanImportExport, error)
	AddObstacle(planID uint, obstacle *model.Obstacle) (*model.EmergencyPlan, error)
	RemoveObstacle(planID uint, obstacleIndex int) (*model.EmergencyPlan, error)
	UpdatePath(planID uint, path []model.PathPoint) (*model.EmergencyPlan, error)
}
