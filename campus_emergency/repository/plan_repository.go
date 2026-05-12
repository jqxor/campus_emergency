package repository

import (
	"errors"
	"campus-emergency/model"
	"gorm.io/gorm"
)

type PlanRepository interface {
	Create(plan *model.EmergencyPlan) error
	Update(plan *model.EmergencyPlan) error
	Delete(id uint) error
	GetByID(id uint) (*model.EmergencyPlan, error)
	GetByScenarioType(scenarioType model.ScenarioType) ([]*model.EmergencyPlan, error)
	Search(condition *model.PlanSearchCondition) ([]*model.EmergencyPlan, int64, error)
	CountByScenarioType() (map[model.ScenarioType]int64, error)
	CheckNameExists(name string, excludeID uint) bool
}

type planRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) PlanRepository { return &planRepository{db: db} }
func (r *planRepository) Create(plan *model.EmergencyPlan) error { return r.db.Create(plan).Error }
func (r *planRepository) Update(plan *model.EmergencyPlan) error { return r.db.Save(plan).Error }
func (r *planRepository) Delete(id uint) error { return r.db.Delete(&model.EmergencyPlan{}, id).Error }
func (r *planRepository) GetByID(id uint) (*model.EmergencyPlan, error) {
	var plan model.EmergencyPlan
	result := r.db.First(&plan, id)
	if result.Error != nil { return nil, result.Error }
	return &plan, nil
}
func (r *planRepository) GetByScenarioType(scenarioType model.ScenarioType) ([]*model.EmergencyPlan, error) {
	var plans []*model.EmergencyPlan
	result := r.db.Where("scenario_type = ?", scenarioType).Find(&plans)
	if result.Error != nil { return nil, result.Error }
	return plans, nil
}
func (r *planRepository) Search(condition *model.PlanSearchCondition) ([]*model.EmergencyPlan, int64, error) {
	var plans []*model.EmergencyPlan
	query := r.db.Model(&model.EmergencyPlan{})
	if condition.ScenarioType != "" { query = query.Where("scenario_type = ?", condition.ScenarioType) }
	if condition.Status != "" { query = query.Where("status = ?", condition.Status) }
	if condition.Keyword != "" { query = query.Where("name LIKE ? OR description LIKE ?", "%"+condition.Keyword+"%", "%"+condition.Keyword+"%") }
	if condition.StartTime != nil { query = query.Where("created_at >= ?", condition.StartTime) }
	if condition.EndTime != nil { query = query.Where("created_at <= ?", condition.EndTime) }
	var total int64
	if err := query.Count(&total).Error; err != nil { return nil, 0, err }
	if condition.SortBy == "" { condition.SortBy = "created_at" }
	if condition.SortOrder == "" { condition.SortOrder = "desc" }
	query = query.Order(condition.SortBy + " " + condition.SortOrder)
	page := condition.Page
	pageSize := condition.PageSize
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 100 { pageSize = 10 }
	query = query.Offset((page-1)*pageSize).Limit(pageSize)
	if err := query.Find(&plans).Error; err != nil { return nil, 0, err }
	return plans, total, nil
}
func (r *planRepository) CountByScenarioType() (map[model.ScenarioType]int64, error) {
	var results []struct {
		ScenarioType model.ScenarioType `json:"scenario_type" gorm:"scenario_type"`
		Count        int64              `json:"count" gorm:"count"`
	}
	if err := r.db.Model(&model.EmergencyPlan{}).Select("scenario_type, COUNT(*) as count").Group("scenario_type").Scan(&results).Error; err != nil {
		return nil, err
	}
	countMap := make(map[model.ScenarioType]int64)
	for _, item := range results { countMap[item.ScenarioType] = item.Count }
	return countMap, nil
}
func (r *planRepository) CheckNameExists(name string, excludeID uint) bool {
	var count int64
	query := r.db.Model(&model.EmergencyPlan{}).Where("name = ?", name)
	if excludeID > 0 { query = query.Where("id != ?", excludeID) }
	query.Count(&count)
	return count > 0
}

var _ = errors.New
