package service

import (
	"context"
	"math"
	"path_optimization/entity"
	"path_optimization/repository"
	"time"
)

// NavigationService 导航服务
type NavigationService interface {
	StartNavigation(ctx context.Context, userID uint64, pathID uint64) (*entity.NavigationRecord, error)
	UpdateNavigation(ctx context.Context, userID uint64, currentLocation *entity.Location, pathID uint64) (*entity.Path, []*entity.ObstacleWarning, error)
	EndNavigation(ctx context.Context, userID uint64, pathID uint64) error
	CheckPathForObstacles(ctx context.Context, path *entity.Path) ([]*entity.Obstacle, error)
	ConfirmObstacleWarning(ctx context.Context, warningID uint64, userID uint64) error
	IgnoreObstacleWarning(ctx context.Context, warningID uint64, userID uint64) error
}

type navigationService struct {
	pathRepo           repository.PathRepository
	obstacleRepo       repository.ObstacleRepository
	navigationRepo     repository.NavigationRecordRepository
	pathCalculationSvc PathCalculationService
	locationSvc        LocationService
}

func NewNavigationService(
	pathRepo repository.PathRepository,
	obstacleRepo repository.ObstacleRepository,
	navigationRepo repository.NavigationRecordRepository,
	pathCalculationSvc PathCalculationService,
	locationSvc LocationService,
) NavigationService {
	return &navigationService{
		pathRepo:           pathRepo,
		obstacleRepo:       obstacleRepo,
		navigationRepo:     navigationRepo,
		pathCalculationSvc: pathCalculationSvc,
		locationSvc:        locationSvc,
	}
}

func (s *navigationService) StartNavigation(ctx context.Context, userID uint64, pathID uint64) (*entity.NavigationRecord, error) {
	path, err := s.pathRepo.GetPathByID(ctx, pathID)
	if err != nil {
		return nil, err
	}

	record := &entity.NavigationRecord{
		UserID:    userID,
		PathID:    pathID,
		Mode:      path.Mode,
		StartTime: time.Now(),
	}
	if err := s.navigationRepo.CreateRecord(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *navigationService) UpdateNavigation(ctx context.Context, userID uint64, currentLocation *entity.Location, pathID uint64) (*entity.Path, []*entity.ObstacleWarning, error) {
	if err := s.locationSvc.UpdateUserLocation(ctx, userID, currentLocation); err != nil {
		return nil, nil, err
	}

	path, err := s.pathRepo.GetPathByID(ctx, pathID)
	if err != nil {
		return nil, nil, err
	}

	obstacles, err := s.CheckPathForObstacles(ctx, path)
	if err != nil {
		return nil, nil, err
	}

	var warnings []*entity.ObstacleWarning
	newPath := path
	if len(obstacles) > 0 {
		for _, obstacle := range obstacles {
			warning := &entity.ObstacleWarning{
				UserID:     userID,
				PathID:     pathID,
				ObstacleID: obstacle.ID,
			}
			if err := s.obstacleRepo.CreateWarning(ctx, warning); err == nil {
				warnings = append(warnings, warning)
			}
		}

		newPath, err = s.pathCalculationSvc.AdjustPathForObstacles(ctx, path, obstacles)
		if err != nil {
			return nil, warnings, err
		}

		record, err := s.navigationRepo.GetRecordByPathID(ctx, pathID)
		if err == nil && record != nil {
			record.ObstaclesEncountered += len(obstacles)
			_ = s.navigationRepo.CreateRecord(ctx, record)
		}
	}

	return newPath, warnings, nil
}

func (s *navigationService) EndNavigation(ctx context.Context, userID uint64, pathID uint64) error {
	_ = userID
	path, err := s.pathRepo.GetPathByID(ctx, pathID)
	if err != nil {
		return err
	}

	record, err := s.navigationRepo.GetRecordByPathID(ctx, pathID)
	if err != nil {
		return err
	}

	endTime := time.Now()
	return s.navigationRepo.UpdateRecordEndTime(ctx, record.ID, endTime, path.Distance)
}

func (s *navigationService) CheckPathForObstacles(ctx context.Context, path *entity.Path) ([]*entity.Obstacle, error) {
	if len(path.Points) == 0 {
		return []*entity.Obstacle{}, nil
	}

	minLat, maxLat, minLng, maxLng := findPathBoundingBox(path)
	obstacles, err := s.obstacleRepo.GetActiveObstaclesInArea(ctx, minLat, maxLat, minLng, maxLng)
	if err != nil {
		return nil, err
	}

	var pathObstacles []*entity.Obstacle
	for _, obstacle := range obstacles {
		if isObstacleOnPath(path, obstacle) {
			pathObstacles = append(pathObstacles, obstacle)
		}
	}
	return pathObstacles, nil
}

func (s *navigationService) ConfirmObstacleWarning(ctx context.Context, warningID uint64, userID uint64) error {
	_ = userID
	return s.obstacleRepo.UpdateWarningStatus(ctx, warningID, true, false)
}

func (s *navigationService) IgnoreObstacleWarning(ctx context.Context, warningID uint64, userID uint64) error {
	_ = userID
	return s.obstacleRepo.UpdateWarningStatus(ctx, warningID, false, true)
}

func findPathBoundingBox(path *entity.Path) (minLat, maxLat, minLng, maxLng float64) {
	if len(path.Points) == 0 {
		return 0, 0, 0, 0
	}

	minLat, maxLat = path.Points[0].Latitude, path.Points[0].Latitude
	minLng, maxLng = path.Points[0].Longitude, path.Points[0].Longitude
	for _, point := range path.Points {
		if point.Latitude < minLat {
			minLat = point.Latitude
		}
		if point.Latitude > maxLat {
			maxLat = point.Latitude
		}
		if point.Longitude < minLng {
			minLng = point.Longitude
		}
		if point.Longitude > maxLng {
			maxLng = point.Longitude
		}
	}

	latBuffer := (maxLat - minLat) * 0.1
	lngBuffer := (maxLng - minLng) * 0.1
	if maxLat == minLat {
		latBuffer = 0.0005
	}
	if maxLng == minLng {
		lngBuffer = 0.0005
	}
	minLat -= latBuffer
	maxLat += latBuffer
	minLng -= lngBuffer
	maxLng += lngBuffer
	return minLat, maxLat, minLng, maxLng
}

func isObstacleOnPath(path *entity.Path, obstacle *entity.Obstacle) bool {
	for i := 0; i < len(path.Points)-1; i++ {
		p1 := path.Points[i]
		p2 := path.Points[i+1]
		distance := distanceToLineSegment(&p1, &p2, &obstacle.Location)
		if distance < 5.0 {
			return true
		}
	}
	return false
}

func distanceToLineSegment(p1, p2, point *entity.Location) float64 {
	x1, y1 := p1.Longitude, p1.Latitude
	x2, y2 := p2.Longitude, p2.Latitude
	x3, y3 := point.Longitude, point.Latitude

	lenSq := (x2-x1)*(x2-x1) + (y2-y1)*(y2-y1)
	if lenSq == 0 {
		return math.Sqrt((x3-x1)*(x3-x1) + (y3-y1)*(y3-y1))
	}

	t := ((x3-x1)*(x2-x1) + (y3-y1)*(y2-y1)) / lenSq
	t = math.Max(0, math.Min(1, t))

	projX := x1 + t*(x2-x1)
	projY := y1 + t*(y2-y1)
	dx := x3 - projX
	dy := y3 - projY
	return math.Sqrt(dx*dx+dy*dy) * 111000
}
