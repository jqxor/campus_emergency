package repository

import (
	"context"
	"path_optimization/entity"

	"gorm.io/gorm"
)

// LocationRepository 位置信息仓库
type LocationRepository interface {
	SaveUserLocation(ctx context.Context, userLocation *entity.UserLocation) error
	GetLatestUserLocation(ctx context.Context, userID uint64) (*entity.UserLocation, error)
	CreateLocation(ctx context.Context, location *entity.Location) error
}

type locationRepository struct {
	db *gorm.DB
}

func NewLocationRepository(db *gorm.DB) LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) SaveUserLocation(ctx context.Context, userLocation *entity.UserLocation) error {
	return r.db.WithContext(ctx).Create(userLocation).Error
}

func (r *locationRepository) GetLatestUserLocation(ctx context.Context, userID uint64) (*entity.UserLocation, error) {
	var userLocation entity.UserLocation
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").First(&userLocation)
	if result.Error != nil {
		return nil, result.Error
	}
	return &userLocation, nil
}

func (r *locationRepository) CreateLocation(ctx context.Context, location *entity.Location) error {
	return r.db.WithContext(ctx).Create(location).Error
}
