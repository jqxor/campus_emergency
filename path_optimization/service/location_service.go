package service

import (
	"context"
	"path_optimization/entity"
	"path_optimization/repository"
)

// LocationService 用户位置服务
type LocationService interface {
	UpdateUserLocation(ctx context.Context, userID uint64, location *entity.Location) error
	GetUserCurrentLocation(ctx context.Context, userID uint64) (*entity.Location, error)
}

type locationService struct {
	locationRepo repository.LocationRepository
}

func NewLocationService(locationRepo repository.LocationRepository) LocationService {
	return &locationService{locationRepo: locationRepo}
}

func (s *locationService) UpdateUserLocation(ctx context.Context, userID uint64, location *entity.Location) error {
	userLocation := &entity.UserLocation{
		UserID:   userID,
		Location: *location,
	}
	return s.locationRepo.SaveUserLocation(ctx, userLocation)
}

func (s *locationService) GetUserCurrentLocation(ctx context.Context, userID uint64) (*entity.Location, error) {
	userLocation, err := s.locationRepo.GetLatestUserLocation(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &userLocation.Location, nil
}
