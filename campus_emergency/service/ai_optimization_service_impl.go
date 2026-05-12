package service

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"campus-emergency/model"
)

type aiOptimizationServiceImpl struct{}

func NewAIOptimizationService() AIOptimizationService { return &aiOptimizationServiceImpl{} }

func (s *aiOptimizationServiceImpl) OptimizePath(plan *model.EmergencyPlan) (*model.AIOptimization, error) {
	if len(plan.EvacuationPath) < 2 { return nil, errors.New("路径点不足，无法进行优化") }
	originalPath := make([]model.PathPoint, len(plan.EvacuationPath))
	copy(originalPath, plan.EvacuationPath)
	sort.Slice(originalPath, func(i, j int) bool { return originalPath[i].Order < originalPath[j].Order })
	originalLength := s.calculatePathLength(originalPath)
	optimizedPath := s.simplifyPath(originalPath, 0.0001)
	if len(optimizedPath) < 2 { optimizedPath = []model.PathPoint{originalPath[0], originalPath[len(originalPath)-1]} }
	for i := range optimizedPath { optimizedPath[i].Order = i }
	optimizedLength := s.calculatePathLength(optimizedPath)
	suggestions := s.generateSuggestions(originalPath, optimizedPath, originalLength, optimizedLength, plan.Obstacles)
	return &model.AIOptimization{OriginalPathLength: originalLength, OptimizedPathLength: optimizedLength, SavingTime: (originalLength - optimizedLength) / 1.2, Suggestions: suggestions, OptimizedPath: optimizedPath, CreatedAt: time.Now()}, nil
}

func (s *aiOptimizationServiceImpl) calculatePathLength(path []model.PathPoint) float64 {
	total := 0.0
	for i := 0; i < len(path)-1; i++ { total += s.calculateDistance(path[i].Latitude, path[i].Longitude, path[i+1].Latitude, path[i+1].Longitude) }
	return total
}
func (s *aiOptimizationServiceImpl) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
func (s *aiOptimizationServiceImpl) simplifyPath(path []model.PathPoint, epsilon float64) []model.PathPoint {
	if len(path) <= 2 { return path }
	maxDist, maxIndex := 0.0, 0
	start, end := path[0], path[len(path)-1]
	for i := 1; i < len(path)-1; i++ {
		dist := s.pointToLineDistance(path[i], start, end)
		if dist > maxDist { maxDist, maxIndex = dist, i }
	}
	if maxDist > epsilon {
		left := s.simplifyPath(path[:maxIndex+1], epsilon)
		right := s.simplifyPath(path[maxIndex:], epsilon)
		return append(left[:len(left)-1], right...)
	}
	return []model.PathPoint{start, end}
}
func (s *aiOptimizationServiceImpl) pointToLineDistance(p, a, b model.PathPoint) float64 {
	ax, ay := a.Longitude, a.Latitude
	bx, by := b.Longitude, b.Latitude
	px, py := p.Longitude, p.Latitude
	abx, aby := bx-ax, by-ay
	apx, apy := px-ax, py-ay
	dot := apx*abx + apy*aby
	lenSq := abx*abx + aby*aby
	var t float64
	if lenSq != 0 { t = math.Max(0, math.Min(1, dot/lenSq)) }
	cx, cy := ax+t*abx, ay+t*aby
	return math.Sqrt((px-cx)*(px-cx) + (py-cy)*(py-cy))
}
func (s *aiOptimizationServiceImpl) generateSuggestions(originalPath, optimizedPath []model.PathPoint, originalLength, optimizedLength float64, obstacles []model.Obstacle) []string {
	var suggestions []string
	if optimizedLength < originalLength*0.9 {
		savingPercent := (1 - optimizedLength/originalLength) * 100
		suggestions = append(suggestions, fmt.Sprintf("路径长度减少了%.1f%%，预计可节省%.1f秒疏散时间", savingPercent, (originalLength-optimizedLength)/1.2))
	}
	if float64(len(optimizedPath)) < float64(len(originalPath))*0.7 {
		savingPercent := (1 - float64(len(optimizedPath))/float64(len(originalPath))) * 100
		suggestions = append(suggestions, fmt.Sprintf("路径点数量减少了%.1f%%，路径更加简洁明了", savingPercent))
	}
	for _, obstacle := range obstacles {
		if s.isPathNearObstacle(optimizedPath, obstacle, 0.001) {
			suggestions = append(suggestions, "路径靠近障碍物，建议至少保持100米以上安全距离")
			break
		}
	}
	if len(suggestions) == 0 { suggestions = append(suggestions, "当前路径已优化，建议定期检查路径上是否有新的障碍物或建筑变化") }
	return suggestions
}
func (s *aiOptimizationServiceImpl) isPathNearObstacle(path []model.PathPoint, obstacle model.Obstacle, threshold float64) bool {
	for _, point := range path {
		distance := math.Sqrt(math.Pow(point.Latitude-obstacle.Latitude, 2) + math.Pow(point.Longitude-obstacle.Longitude, 2))
		if distance < threshold { return true }
	}
	return false
}
