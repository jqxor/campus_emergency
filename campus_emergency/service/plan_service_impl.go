package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"campus-emergency/model"
	"campus-emergency/repository"
)

type planServiceImpl struct {
	planRepo     repository.PlanRepository
	aiOptService AIOptimizationService
}

func NewPlanService(planRepo repository.PlanRepository, aiOptService AIOptimizationService) PlanService {
	return &planServiceImpl{planRepo: planRepo, aiOptService: aiOptService}
}

func (s *planServiceImpl) CreatePlan(plan *model.EmergencyPlan) error {
	if s.planRepo.CheckNameExists(plan.Name, 0) { return errors.New("预案名称已存在") }
	if plan.Priority < 0 || plan.Priority > 10 { return errors.New("优先级必须在0-10之间") }
	if plan.Status == "" { plan.Status = model.PlanStatusDraft }
	if len(plan.EvacuationPath) > 1 { plan.EstimatedTime = s.calculatePathTime(plan.EvacuationPath) }
	return s.planRepo.Create(plan)
}

func (s *planServiceImpl) UpdatePlan(plan *model.EmergencyPlan) error {
	existingPlan, err := s.planRepo.GetByID(plan.ID)
	if err != nil { return fmt.Errorf("获取预案失败: %v", err) }
	if existingPlan.Name != plan.Name && s.planRepo.CheckNameExists(plan.Name, plan.ID) { return errors.New("预案名称已存在") }
	if plan.Priority < 0 || plan.Priority > 10 { return errors.New("优先级必须在0-10之间") }
	if !s.pathsEqual(existingPlan.EvacuationPath, plan.EvacuationPath) { plan.EstimatedTime = s.calculatePathTime(plan.EvacuationPath) }
	return s.planRepo.Update(plan)
}

func (s *planServiceImpl) DeletePlan(id uint) error {
	if _, err := s.planRepo.GetByID(id); err != nil { return fmt.Errorf("预案不存在: %v", err) }
	return s.planRepo.Delete(id)
}
func (s *planServiceImpl) GetPlanByID(id uint) (*model.EmergencyPlan, error) { return s.planRepo.GetByID(id) }
func (s *planServiceImpl) GetPlansByScenarioType(scenarioType model.ScenarioType) ([]*model.EmergencyPlan, error) { return s.planRepo.GetByScenarioType(scenarioType) }
func (s *planServiceImpl) SearchPlans(condition *model.PlanSearchCondition) ([]*model.EmergencyPlan, int64, error) { return s.planRepo.Search(condition) }
func (s *planServiceImpl) GetScenarioTypeStats() (map[model.ScenarioType]int64, error) { return s.planRepo.CountByScenarioType() }

func (s *planServiceImpl) UpdatePlanStatus(id uint, status model.PlanStatus) error {
	plan, err := s.planRepo.GetByID(id)
	if err != nil { return fmt.Errorf("预案不存在: %v", err) }
	if !s.isValidStatusTransition(plan.Status, status) { return errors.New("不允许的状态转换") }
	plan.Status = status
	return s.planRepo.Update(plan)
}

func (s *planServiceImpl) OptimizePlanPath(id uint) (*model.EmergencyPlan, error) {
	plan, err := s.planRepo.GetByID(id)
	if err != nil { return nil, fmt.Errorf("预案不存在: %v", err) }
	if len(plan.EvacuationPath) < 2 { return nil, errors.New("疏散路径点不足，无法进行优化") }
	optimization, err := s.aiOptService.OptimizePath(plan)
	if err != nil { return nil, fmt.Errorf("路径优化失败: %v", err) }
	plan.AIOptimization = optimization
	now := time.Now()
	plan.LastOptimizedAt = &now
	if len(optimization.OptimizedPath) > 0 {
		plan.EvacuationPath = optimization.OptimizedPath
		plan.EstimatedTime = s.calculatePathTime(optimization.OptimizedPath)
	}
	if err := s.planRepo.Update(plan); err != nil { return nil, fmt.Errorf("更新预案失败: %v", err) }
	return plan, nil
}

func (s *planServiceImpl) ImportPlans(jsonData []byte, userID uint, userName string) (int, error) {
	var importData model.PlanImportExport
	if err := json.Unmarshal(jsonData, &importData); err != nil { return 0, fmt.Errorf("解析JSON数据失败: %v", err) }
	importedCount := 0
	for _, importedPlan := range importData.Plans {
		if s.planRepo.CheckNameExists(importedPlan.Name, 0) { continue }
		newPlan := &model.EmergencyPlan{
			Name: importedPlan.Name, ScenarioType: importedPlan.ScenarioType, Description: importedPlan.Description,
			Status: model.PlanStatusDraft, Priority: importedPlan.Priority, CreatorID: userID, CreatorName: userName,
			SafePoints: importedPlan.SafePoints, EvacuationPath: importedPlan.EvacuationPath, Obstacles: importedPlan.Obstacles,
			EstimatedTime: importedPlan.EstimatedTime, PeopleCount: importedPlan.PeopleCount, Tags: importedPlan.Tags,
		}
		if err := s.planRepo.Create(newPlan); err == nil { importedCount++ }
	}
	return importedCount, nil
}

func (s *planServiceImpl) ExportPlans(condition *model.PlanSearchCondition) (*model.PlanImportExport, error) {
	plans, _, err := s.planRepo.Search(condition)
	if err != nil { return nil, fmt.Errorf("搜索预案失败: %v", err) }
	values := make([]model.EmergencyPlan, len(plans))
	for i, plan := range plans { values[i] = *plan }
	return &model.PlanImportExport{Version: "1.0", Plans: values, ExportTime: time.Now()}, nil
}

func (s *planServiceImpl) AddObstacle(planID uint, obstacle *model.Obstacle) (*model.EmergencyPlan, error) {
	plan, err := s.planRepo.GetByID(planID)
	if err != nil { return nil, fmt.Errorf("预案不存在: %v", err) }
	plan.Obstacles = append(plan.Obstacles, *obstacle)
	plan.EstimatedTime = s.calculatePathTime(plan.EvacuationPath)
	if err := s.planRepo.Update(plan); err != nil { return nil, fmt.Errorf("更新障碍物失败: %v", err) }
	return plan, nil
}

func (s *planServiceImpl) RemoveObstacle(planID uint, obstacleIndex int) (*model.EmergencyPlan, error) {
	plan, err := s.planRepo.GetByID(planID)
	if err != nil { return nil, fmt.Errorf("预案不存在: %v", err) }
	if obstacleIndex < 0 || obstacleIndex >= len(plan.Obstacles) { return nil, errors.New("障碍物索引无效") }
	plan.Obstacles = append(plan.Obstacles[:obstacleIndex], plan.Obstacles[obstacleIndex+1:]...)
	plan.EstimatedTime = s.calculatePathTime(plan.EvacuationPath)
	if err := s.planRepo.Update(plan); err != nil { return nil, fmt.Errorf("更新障碍物失败: %v", err) }
	return plan, nil
}

func (s *planServiceImpl) UpdatePath(planID uint, path []model.PathPoint) (*model.EmergencyPlan, error) {
	plan, err := s.planRepo.GetByID(planID)
	if err != nil { return nil, fmt.Errorf("预案不存在: %v", err) }
	plan.EvacuationPath = path
	plan.EstimatedTime = s.calculatePathTime(path)
	if err := s.planRepo.Update(plan); err != nil { return nil, fmt.Errorf("更新路径失败: %v", err) }
	return plan, nil
}

func (s *planServiceImpl) calculatePathTime(path []model.PathPoint) float64 {
	if len(path) < 2 { return 0 }
	totalDistance := 0.0
	averageSpeed := 1.2
	sortedPath := make([]model.PathPoint, len(path))
	copy(sortedPath, path)
	sort.Slice(sortedPath, func(i, j int) bool { return sortedPath[i].Order < sortedPath[j].Order })
	for i := 0; i < len(sortedPath)-1; i++ { totalDistance += s.calculateDistance(sortedPath[i].Latitude, sortedPath[i].Longitude, sortedPath[i+1].Latitude, sortedPath[i+1].Longitude) }
	return totalDistance / averageSpeed
}

func (s *planServiceImpl) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func (s *planServiceImpl) pathsEqual(path1, path2 []model.PathPoint) bool {
	if len(path1) != len(path2) { return false }
	for i, p1 := range path1 {
		p2 := path2[i]
		if p1.Order != p2.Order || p1.Latitude != p2.Latitude || p1.Longitude != p2.Longitude { return false }
	}
	return true
}

func (s *planServiceImpl) isValidStatusTransition(current, target model.PlanStatus) bool {
	validTransitions := map[model.PlanStatus][]model.PlanStatus{
		model.PlanStatusDraft:       {model.PlanStatusActive, model.PlanStatusInactive, model.PlanStatusUnderReview},
		model.PlanStatusUnderReview: {model.PlanStatusDraft, model.PlanStatusActive, model.PlanStatusInactive},
		model.PlanStatusActive:      {model.PlanStatusInactive, model.PlanStatusExpired},
		model.PlanStatusInactive:    {model.PlanStatusDraft, model.PlanStatusActive},
		model.PlanStatusExpired:     {model.PlanStatusDraft, model.PlanStatusActive},
	}
	for _, allowed := range validTransitions[current] { if allowed == target { return true } }
	return false
}
