package service

import (
	"context"
	"path_optimization/entity"
	"path_optimization/repository"
	"time"
)

// ReportService 报告服务
type ReportService interface {
	ExportNavigationHistoryPDF(ctx context.Context, userID uint64, startTime, endTime time.Time) ([]byte, error)
	GetNavigationSummary(ctx context.Context, userID uint64, startTime, endTime time.Time) (*NavigationSummary, error)
}

// NavigationSummary 导航摘要
type NavigationSummary struct {
	TotalTrips           int       `json:"total_trips"`
	TotalDistance        float64   `json:"total_distance"`
	TotalTime            int64     `json:"total_time"`
	AverageSpeed         float64   `json:"average_speed"`
	MostUsedMode         string    `json:"most_used_mode"`
	ObstaclesEncountered int       `json:"obstacles_encountered"`
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
}

type reportService struct {
	navigationRepo repository.NavigationRecordRepository
	pathRepo       repository.PathRepository
}

func NewReportService(
	navigationRepo repository.NavigationRecordRepository,
	pathRepo repository.PathRepository,
) ReportService {
	return &reportService{
		navigationRepo: navigationRepo,
		pathRepo:       pathRepo,
	}
}

func (s *reportService) ExportNavigationHistoryPDF(ctx context.Context, userID uint64, startTime, endTime time.Time) ([]byte, error) {
	records, err := s.navigationRepo.GetUserRecordsByTimeRange(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	summary, err := s.GetNavigationSummary(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	_ = summary

	detailedRecords := make([]struct {
		Record *entity.NavigationRecord
		Path   *entity.Path
	}, len(records))
	for i, record := range records {
		path, err := s.pathRepo.GetPathByID(ctx, record.PathID)
		if err != nil {
			return nil, err
		}
		detailedRecords[i] = struct {
			Record *entity.NavigationRecord
			Path   *entity.Path
		}{
			Record: record,
			Path:   path,
		}
	}
	_ = detailedRecords

	return []byte("PDF_CONTENT"), nil
}

func (s *reportService) GetNavigationSummary(ctx context.Context, userID uint64, startTime, endTime time.Time) (*NavigationSummary, error) {
	records, err := s.navigationRepo.GetUserRecordsByTimeRange(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return &NavigationSummary{StartDate: startTime, EndDate: endTime}, nil
	}

	totalTrips := len(records)
	totalDistance := 0.0
	totalTime := int64(0)
	obstaclesEncountered := 0
	modeCount := make(map[string]int)

	for _, record := range records {
		totalDistance += record.ActualDistance
		totalTime += record.ActualTime
		obstaclesEncountered += record.ObstaclesEncountered
		modeCount[string(record.Mode)]++
	}

	mostUsedMode := ""
	maxCount := 0
	for mode, count := range modeCount {
		if count > maxCount {
			maxCount = count
			mostUsedMode = mode
		}
	}

	averageSpeed := 0.0
	if totalTime > 0 {
		averageSpeed = totalDistance / float64(totalTime)
	}

	return &NavigationSummary{
		TotalTrips:           totalTrips,
		TotalDistance:        totalDistance,
		TotalTime:            totalTime,
		AverageSpeed:         averageSpeed,
		MostUsedMode:         mostUsedMode,
		ObstaclesEncountered: obstaclesEncountered,
		StartDate:            startTime,
		EndDate:              endTime,
	}, nil
}
