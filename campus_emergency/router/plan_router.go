package router

import (
	"campus-emergency/controller"
	"github.com/gin-gonic/gin"
)

func RegisterPlanRoutes(router *gin.Engine, planController *controller.PlanController) {
	planRoutes := router.Group("/api/plans")
	{
		planRoutes.POST("", planController.CreatePlan)
		planRoutes.PUT("/:id", planController.UpdatePlan)
		planRoutes.DELETE("/:id", planController.DeletePlan)
		planRoutes.GET("/:id", planController.GetPlan)
		planRoutes.GET("/search", planController.SearchPlans)
		planRoutes.GET("/scenario/:scenario_type", planController.GetPlansByScenarioType)
		planRoutes.PATCH("/:id/status", planController.UpdatePlanStatus)
		planRoutes.POST("/:id/optimize", planController.OptimizePlanPath)
		planRoutes.POST("/import", planController.ImportPlans)
		planRoutes.GET("/export", planController.ExportPlans)
		planRoutes.PUT("/:id/path", planController.UpdatePath)
	}
}
