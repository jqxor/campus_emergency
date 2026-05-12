package repository

import (
	"context"
	"path_optimization/entity"

	"gorm.io/gorm"
)

// ObstacleRepository 障碍物仓库
type ObstacleRepository interface {
	CreateObstacle(ctx context.Context, obstacle *entity.Obstacle) error
	GetActiveObstaclesInArea(ctx context.Context, minLat, maxLat, minLng, maxLng float64) ([]*entity.Obstacle, error)
	CreateWarning(ctx context.Context, warning *entity.ObstacleWarning) error
	UpdateWarningStatus(ctx context.Context, warningID uint64, isConfirmed, isIgnored bool) error
}

type obstacleRepository struct {
	db *gorm.DB
}

func NewObstacleRepository(db *gorm.DB) ObstacleRepository {
	return &obstacleRepository{db: db}
}

func (r *obstacleRepository) CreateObstacle(ctx context.Context, obstacle *entity.Obstacle) error {
	return r.db.WithContext(ctx).Create(obstacle).Error
}

func (r *obstacleRepository) GetActiveObstaclesInArea(ctx context.Context, minLat, maxLat, minLng, maxLng float64) ([]*entity.Obstacle, error) {
	var obstacles []*entity.Obstacle
	result := r.db.WithContext(ctx).Where(
		"is_active = ? AND latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?",
		true, minLat, maxLat, minLng, maxLng,
	).Find(&obstacles)
	if result.Error != nil {
		return nil, result.Error
	}
	return obstacles, nil
}

func (r *obstacleRepository) CreateWarning(ctx context.Context, warning *entity.ObstacleWarning) error {
	return r.db.WithContext(ctx).Create(warning).Error
}

func (r *obstacleRepository) UpdateWarningStatus(ctx context.Context, warningID uint64, isConfirmed, isIgnored bool) error {
	return r.db.WithContext(ctx).Model(&entity.ObstacleWarning{}).Where("id = ?", warningID).Updates(
		map[string]interface{}{
			"is_confirmed": isConfirmed,
			"is_ignored":   isIgnored,
		},
	).Error
}
