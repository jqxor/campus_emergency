package service

import (
	"context"
	"math"
	"path_optimization/entity"
	"path_optimization/repository"
	"sort"
)

// PathCalculationService 路径计算服务
type PathCalculationService interface {
	CalculateOptimalPath(ctx context.Context, start, end *entity.Location, mode entity.NavigationMode, userID uint64) (*entity.Path, error)
	AdjustPathForObstacles(ctx context.Context, originalPath *entity.Path, obstacles []*entity.Obstacle) (*entity.Path, error)
}

type pathCalculationService struct {
	pathRepo     repository.PathRepository
	locationRepo repository.LocationRepository
	obstacleRepo repository.ObstacleRepository
}

func NewPathCalculationService(
	pathRepo repository.PathRepository,
	locationRepo repository.LocationRepository,
	obstacleRepo repository.ObstacleRepository,
) PathCalculationService {
	return &pathCalculationService{
		pathRepo:     pathRepo,
		locationRepo: locationRepo,
		obstacleRepo: obstacleRepo,
	}
}

func (s *pathCalculationService) CalculateOptimalPath(ctx context.Context, start, end *entity.Location, mode entity.NavigationMode, userID uint64) (*entity.Path, error) {
	if err := s.locationRepo.CreateLocation(ctx, start); err != nil {
		return nil, err
	}
	if err := s.locationRepo.CreateLocation(ctx, end); err != nil {
		return nil, err
	}

	var points []entity.Location
	var distance float64
	var estimatedTime int64

	switch mode {
	case entity.NavigationModeWalking:
		points, distance = s.calculateWalkingPath(start, end)
		estimatedTime = int64(distance / 1.3)
	case entity.NavigationModeCycling:
		points, distance = s.calculateCyclingPath(start, end)
		estimatedTime = int64(distance / 5.0)
	case entity.NavigationModeDisabled:
		points, distance = s.calculateDisabledPath(start, end)
		estimatedTime = int64(distance / 0.8)
	default:
		points, distance = s.calculateWalkingPath(start, end)
		estimatedTime = int64(distance / 1.3)
	}

	path := &entity.Path{
		UserID:        userID,
		StartPoint:    *start,
		EndPoint:      *end,
		Mode:          mode,
		Distance:      distance,
		EstimatedTime: estimatedTime,
		Points:        points,
	}

	if err := s.pathRepo.CreatePath(ctx, path); err != nil {
		return nil, err
	}

	pathPoints := make([]*entity.PathPoint, len(points))
	for i, point := range points {
		pathPoints[i] = &entity.PathPoint{
			PathID:   path.ID,
			Location: point,
			Order:    i,
		}
	}
	if err := s.pathRepo.SavePathPoints(ctx, pathPoints); err != nil {
		return nil, err
	}

	path.PathPoints = make([]entity.PathPoint, len(pathPoints))
	for i, p := range pathPoints {
		path.PathPoints[i] = *p
	}
	return path, nil
}

func (s *pathCalculationService) AdjustPathForObstacles(ctx context.Context, originalPath *entity.Path, obstacles []*entity.Obstacle) (*entity.Path, error) {
	if len(obstacles) == 0 {
		return originalPath, nil
	}

	affectedSegments := s.findAffectedPathSegments(originalPath, obstacles)
	if len(affectedSegments) == 0 {
		return originalPath, nil
	}

	newPoints := make([]entity.Location, 0, len(originalPath.Points))
	processedUpTo := 0
	for _, segment := range affectedSegments {
		newPoints = append(newPoints, originalPath.Points[processedUpTo:segment.StartIndex+1]...)
		detourPoints := s.calculateDetour(
			originalPath.Points[segment.StartIndex],
			originalPath.Points[segment.EndIndex],
			segment.Obstacles,
			originalPath.Mode,
		)
		if len(detourPoints) > 1 {
			newPoints = append(newPoints, detourPoints[1:]...)
		}
		processedUpTo = segment.EndIndex
	}

	if processedUpTo < len(originalPath.Points) {
		newPoints = append(newPoints, originalPath.Points[processedUpTo:]...)
	}

	newPath := &entity.Path{
		UserID:     originalPath.UserID,
		StartPoint: originalPath.StartPoint,
		EndPoint:   originalPath.EndPoint,
		Mode:       originalPath.Mode,
		Points:     newPoints,
	}

	newPath.Distance = calculatePathDistance(newPoints)
	switch newPath.Mode {
	case entity.NavigationModeWalking:
		newPath.EstimatedTime = int64(newPath.Distance / 1.3)
	case entity.NavigationModeCycling:
		newPath.EstimatedTime = int64(newPath.Distance / 5.0)
	case entity.NavigationModeDisabled:
		newPath.EstimatedTime = int64(newPath.Distance / 0.8)
	}

	if err := s.pathRepo.CreatePath(ctx, newPath); err != nil {
		return nil, err
	}

	pathPoints := make([]*entity.PathPoint, len(newPoints))
	for i, point := range newPoints {
		pathPoints[i] = &entity.PathPoint{
			PathID:   newPath.ID,
			Location: point,
			Order:    i,
		}
	}
	if err := s.pathRepo.SavePathPoints(ctx, pathPoints); err != nil {
		return nil, err
	}

	newPath.PathPoints = make([]entity.PathPoint, len(pathPoints))
	for i, p := range pathPoints {
		newPath.PathPoints[i] = *p
	}
	return newPath, nil
}

type affectedSegment struct {
	StartIndex int
	EndIndex   int
	Obstacles  []*entity.Obstacle
}

func (s *pathCalculationService) findAffectedPathSegments(path *entity.Path, obstacles []*entity.Obstacle) []affectedSegment {
	var affectedSegments []affectedSegment
	points := path.Points

	for _, obstacle := range obstacles {
		closestPointIndex := -1
		minDistance := math.MaxFloat64
		for i, point := range points {
			distance := calculateDistance(&point, &obstacle.Location)
			if distance < 5.0 && distance < minDistance {
				minDistance = distance
				closestPointIndex = i
			}
		}
		if closestPointIndex != -1 {
			startIndex := max(0, closestPointIndex-2)
			endIndex := min(len(points)-1, closestPointIndex+2)
			found := false
			for i, seg := range affectedSegments {
				if startIndex <= seg.EndIndex && endIndex >= seg.StartIndex {
					affectedSegments[i].StartIndex = min(seg.StartIndex, startIndex)
					affectedSegments[i].EndIndex = max(seg.EndIndex, endIndex)
					affectedSegments[i].Obstacles = append(affectedSegments[i].Obstacles, obstacle)
					found = true
					break
				}
			}
			if !found {
				affectedSegments = append(affectedSegments, affectedSegment{
					StartIndex: startIndex,
					EndIndex:   endIndex,
					Obstacles:  []*entity.Obstacle{obstacle},
				})
			}
		}
	}

	sort.Slice(affectedSegments, func(i, j int) bool {
		return affectedSegments[i].StartIndex < affectedSegments[j].StartIndex
	})
	return affectedSegments
}

func (s *pathCalculationService) calculateDetour(start, end entity.Location, obstacles []*entity.Obstacle, mode entity.NavigationMode) []entity.Location {
	_ = mode
	points := []entity.Location{start}

	dx := end.Longitude - start.Longitude
	dy := end.Latitude - start.Latitude
	distance := math.Sqrt(dx*dx + dy*dy)
	if distance == 0 {
		return []entity.Location{start, end}
	}

	detourDistance := 0.0001
	for _, obstacle := range obstacles {
		detourDistance += 0.00005 * float64(obstacle.Severity)
	}

	midPoint1 := entity.Location{
		Latitude:  start.Latitude + dy*detourDistance/distance,
		Longitude: start.Longitude - dx*detourDistance/distance,
	}
	midPoint2 := entity.Location{
		Latitude:  end.Latitude + dy*detourDistance/distance,
		Longitude: end.Longitude - dx*detourDistance/distance,
	}
	points = append(points, midPoint1, midPoint2, end)
	return points
}

func (s *pathCalculationService) calculateWalkingPath(start, end *entity.Location) ([]entity.Location, float64) {
	return createStraightPathWithPoints(start, end, 20), calculateDistance(start, end)
}

func (s *pathCalculationService) calculateCyclingPath(start, end *entity.Location) ([]entity.Location, float64) {
	return createStraightPathWithPoints(start, end, 10), calculateDistance(start, end)
}

func (s *pathCalculationService) calculateDisabledPath(start, end *entity.Location) ([]entity.Location, float64) {
	path := createStraightPathWithPoints(start, end, 30)
	distance := calculateDistance(start, end) * 1.2
	return path, distance
}

func createStraightPathWithPoints(start, end *entity.Location, numPoints int) []entity.Location {
	points := []entity.Location{*start}
	latStep := (end.Latitude - start.Latitude) / float64(numPoints+1)
	lngStep := (end.Longitude - start.Longitude) / float64(numPoints+1)
	for i := 1; i <= numPoints; i++ {
		points = append(points, entity.Location{
			Latitude:  start.Latitude + latStep*float64(i),
			Longitude: start.Longitude + lngStep*float64(i),
		})
	}
	points = append(points, *end)
	return points
}

func calculateDistance(point1, point2 *entity.Location) float64 {
	const earthRadius = 6371000
	lat1 := degreesToRadians(point1.Latitude)
	lon1 := degreesToRadians(point1.Longitude)
	lat2 := degreesToRadians(point2.Latitude)
	lon2 := degreesToRadians(point2.Longitude)
	dLat := lat2 - lat1
	dLon := lon2 - lon1
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func calculatePathDistance(points []entity.Location) float64 {
	if len(points) < 2 {
		return 0
	}
	totalDistance := 0.0
	for i := 0; i < len(points)-1; i++ {
		totalDistance += calculateDistance(&points[i], &points[i+1])
	}
	return totalDistance
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
