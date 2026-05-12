package repository

import (
	"context"
	"path_optimization/entity"

	"gorm.io/gorm"
)

// PathRepository 路径仓库
type PathRepository interface {
	CreatePath(ctx context.Context, path *entity.Path) error
	GetPathByID(ctx context.Context, pathID uint64) (*entity.Path, error)
	UpdatePath(ctx context.Context, path *entity.Path) error
	SavePathPoints(ctx context.Context, pathPoints []*entity.PathPoint) error
}

type pathRepository struct {
	db *gorm.DB
}

func NewPathRepository(db *gorm.DB) PathRepository {
	return &pathRepository{db: db}
}

func (r *pathRepository) CreatePath(ctx context.Context, path *entity.Path) error {
	return r.db.WithContext(ctx).Create(path).Error
}

func (r *pathRepository) GetPathByID(ctx context.Context, pathID uint64) (*entity.Path, error) {
	var path entity.Path
	result := r.db.WithContext(ctx).Where("id = ?", pathID).First(&path)
	if result.Error != nil {
		return nil, result.Error
	}

	var pathPoints []entity.PathPoint
	if err := r.db.WithContext(ctx).Where("path_id = ?", pathID).Order("`order` ASC").Find(&pathPoints).Error; err != nil {
		return nil, err
	}

	path.Points = make([]entity.Location, len(pathPoints))
	for i, point := range pathPoints {
		path.Points[i] = point.Location
	}
	path.PathPoints = pathPoints
	return &path, nil
}

func (r *pathRepository) UpdatePath(ctx context.Context, path *entity.Path) error {
	return r.db.WithContext(ctx).Save(path).Error
}

func (r *pathRepository) SavePathPoints(ctx context.Context, pathPoints []*entity.PathPoint) error {
	if len(pathPoints) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(pathPoints, 100).Error
}
