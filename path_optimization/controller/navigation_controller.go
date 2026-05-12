package controller

import (
	"net/http"
	"path_optimization/entity"
	"path_optimization/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// NavigationController 导航控制器
type NavigationController struct {
	navigationSvc      service.NavigationService
	pathCalculationSvc service.PathCalculationService
	locationSvc        service.LocationService
	reportSvc          service.ReportService
}

func NewNavigationController(
	navigationSvc service.NavigationService,
	pathCalculationSvc service.PathCalculationService,
	locationSvc service.LocationService,
	reportSvc service.ReportService,
) *NavigationController {
	return &NavigationController{
		navigationSvc:      navigationSvc,
		pathCalculationSvc: pathCalculationSvc,
		locationSvc:        locationSvc,
		reportSvc:          reportSvc,
	}
}

func (c *NavigationController) CalculatePathHandler(ctx *gin.Context) {
	var req struct {
		StartLat float64               `json:"start_lat" binding:"required"`
		StartLng float64               `json:"start_lng" binding:"required"`
		EndLat   float64               `json:"end_lat" binding:"required"`
		EndLng   float64               `json:"end_lng" binding:"required"`
		Mode     entity.NavigationMode `json:"mode" binding:"required,oneof=walking cycling disabled"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := strconv.ParseUint(ctx.GetHeader("X-User-ID"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未提供用户ID"})
		return
	}

	start := &entity.Location{Latitude: req.StartLat, Longitude: req.StartLng}
	end := &entity.Location{Latitude: req.EndLat, Longitude: req.EndLng}

	path, err := c.pathCalculationSvc.CalculateOptimalPath(ctx.Request.Context(), start, end, req.Mode, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "路径计算失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, path)
}

func (c *NavigationController) StartNavigationHandler(ctx *gin.Context) {
	pathID, err := strconv.ParseUint(ctx.Param("path_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的路径ID"})
		return
	}
	userID, err := strconv.ParseUint(ctx.GetHeader("X-User-ID"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未提供用户ID"})
		return
	}

	record, err := c.navigationSvc.StartNavigation(ctx.Request.Context(), userID, pathID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "导航开始失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, record)
}

func (c *NavigationController) UpdateNavigationHandler(ctx *gin.Context) {
	pathID, err := strconv.ParseUint(ctx.Param("path_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的路径ID"})
		return
	}

	var req struct {
		CurrentLat float64 `json:"current_lat" binding:"required"`
		CurrentLng float64 `json:"current_lng" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := strconv.ParseUint(ctx.GetHeader("X-User-ID"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未提供用户ID"})
		return
	}

	currentLocation := &entity.Location{Latitude: req.CurrentLat, Longitude: req.CurrentLng}
	path, warnings, err := c.navigationSvc.UpdateNavigation(ctx.Request.Context(), userID, currentLocation, pathID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "导航更新失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"path": path, "warnings": warnings})
}

func (c *NavigationController) EndNavigationHandler(ctx *gin.Context) {
	pathID, err := strconv.ParseUint(ctx.Param("path_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的路径ID"})
		return
	}
	userID, err := strconv.ParseUint(ctx.GetHeader("X-User-ID"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未提供用户ID"})
		return
	}

	if err := c.navigationSvc.EndNavigation(ctx.Request.Context(), userID, pathID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "导航结束失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "导航已结束"})
}

func (c *NavigationController) ConfirmObstacleWarningHandler(ctx *gin.Context) {
	warningID, err := strconv.ParseUint(ctx.Param("warning_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的警告ID"})
		return
	}
	userID, err := strconv.ParseUint(ctx.GetHeader("X-User-ID"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未提供用户ID"})
		return
	}

	if err := c.navigationSvc.ConfirmObstacleWarning(ctx.Request.Context(), warningID, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "确认警告失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "已确认障碍物警告，路径将更新"})
}

func (c *NavigationController) IgnoreObstacleWarningHandler(ctx *gin.Context) {
	warningID, err := strconv.ParseUint(ctx.Param("warning_id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的警告ID"})
		return
	}
	userID, err := strconv.ParseUint(ctx.GetHeader("X-User-ID"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未提供用户ID"})
		return
	}

	if err := c.navigationSvc.IgnoreObstacleWarning(ctx.Request.Context(), warningID, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "忽略警告失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "已忽略障碍物警告，将继续使用当前路径"})
}

func (c *NavigationController) ExportNavigationHistoryHandler(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	startTime, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的开始日期格式，应为YYYY-MM-DD"})
		return
	}
	endTime, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的结束日期格式，应为YYYY-MM-DD"})
		return
	}
	endTime = endTime.Add(24*time.Hour - time.Second)

	userID, err := strconv.ParseUint(ctx.GetHeader("X-User-ID"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未提供用户ID"})
		return
	}

	pdfContent, err := c.reportSvc.ExportNavigationHistoryPDF(ctx.Request.Context(), userID, startTime, endTime)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "导出历史记录失败: " + err.Error()})
		return
	}

	ctx.Header("Content-Type", "application/pdf")
	ctx.Header("Content-Disposition", "attachment; filename=\"navigation_history_"+startDateStr+"_to_"+endDateStr+".pdf\"")
	ctx.Data(http.StatusOK, "application/pdf", pdfContent)
}

func (c *NavigationController) GetNavigationSummaryHandler(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	startTime, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的开始日期格式，应为YYYY-MM-DD"})
		return
	}
	endTime, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的结束日期格式，应为YYYY-MM-DD"})
		return
	}
	endTime = endTime.Add(24*time.Hour - time.Second)

	userID, err := strconv.ParseUint(ctx.GetHeader("X-User-ID"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未提供用户ID"})
		return
	}

	summary, err := c.reportSvc.GetNavigationSummary(ctx.Request.Context(), userID, startTime, endTime)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取导航摘要失败: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, summary)
}
