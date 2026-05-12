package repository

import (
	"context"
	"path_optimization/entity"
	"time"

	"gorm.io/gorm"
)

// NavigationRecordRepository 导航记录仓库
type NavigationRecordRepository interface {
	CreateRecord(ctx context.Context, record *entity.NavigationRecord) error
	UpdateRecordEndTime(ctx context.Context, recordID uint64, endTime time.Time, actualDistance float64) error
	GetUserRecordsByTimeRange(ctx context.Context, userID uint64, startTime, endTime time.Time) ([]*entity.NavigationRecord, error)
	GetRecordByPathID(ctx context.Context, pathID uint64) (*entity.NavigationRecord, error)
}

type navigationRecordRepository struct {
	db *gorm.DB
}

func NewNavigationRecordRepository(db *gorm.DB) NavigationRecordRepository {
	return &navigationRecordRepository{db: db}
}

func (r *navigationRecordRepository) CreateRecord(ctx context.Context, record *entity.NavigationRecord) error {
	if record.ID == 0 {
		return r.db.WithContext(ctx).Create(record).Error
	}
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *navigationRecordRepository) UpdateRecordEndTime(ctx context.Context, recordID uint64, endTime time.Time, actualDistance float64) error {
	record, err := r.getRecordByID(ctx, recordID)
	if err != nil {
		return err
	}
	actualTime := int64(endTime.Sub(record.StartTime).Seconds())
	return r.db.WithContext(ctx).Model(&entity.NavigationRecord{}).Where("id = ?", recordID).Updates(
		map[string]interface{}{
			"end_time":        endTime,
			"actual_distance": actualDistance,
			"actual_time":     actualTime,
		},
	).Error
}

func (r *navigationRecordRepository) GetUserRecordsByTimeRange(ctx context.Context, userID uint64, startTime, endTime time.Time) ([]*entity.NavigationRecord, error) {
	var records []*entity.NavigationRecord
	result := r.db.WithContext(ctx).Where(
		"user_id = ? AND start_time BETWEEN ? AND ?",
		userID, startTime, endTime,
	).Order("start_time DESC").Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	return records, nil
}

func (r *navigationRecordRepository) GetRecordByPathID(ctx context.Context, pathID uint64) (*entity.NavigationRecord, error) {
	var record entity.NavigationRecord
	result := r.db.WithContext(ctx).Where("path_id = ?", pathID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func (r *navigationRecordRepository) getRecordByID(ctx context.Context, recordID uint64) (*entity.NavigationRecord, error) {
	var record entity.NavigationRecord
	result := r.db.WithContext(ctx).Where("id = ?", recordID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}
