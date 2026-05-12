package router

import (
	"github.com/gin-gonic/gin"
	"role-management/controller"
)

func RegisterRoleRoutes(r *gin.Engine, roleController *controller.RoleController, systemController *controller.SystemController, authController *controller.AuthController) {
	requireAuth := authController.RequireAuth()
	requireAdmin := authController.RequireAdmin()

	roleRoutes := r.Group("/api/roles", requireAuth)
	{
		roleRoutes.POST("", requireAdmin, roleController.CreateRole)
		roleRoutes.GET("", roleController.ListRoles)
		roleRoutes.GET("/:id", roleController.GetRole)
		roleRoutes.PUT("/:id", requireAdmin, roleController.UpdateRole)
		roleRoutes.DELETE("/:id", requireAdmin, roleController.DeleteRole)
		roleRoutes.POST("/:id/permissions", requireAdmin, roleController.AssignPermissions)
		roleRoutes.GET("/:id/permissions", roleController.GetRolePermissions)
		roleRoutes.GET("/export", requireAdmin, roleController.ExportRoles)
	}

	userRoutes := r.Group("/api/users", requireAuth, requireAdmin)
	{
		userRoutes.POST("", systemController.CreateUser)
		userRoutes.GET("", systemController.ListUsers)
		userRoutes.PUT("/:id", systemController.UpdateUser)
		userRoutes.DELETE("/:id", systemController.DeleteUser)
		userRoutes.POST("/:id/permissions", systemController.AssignUserPermission)
	}

	permissionRoutes := r.Group("/api/permissions", requireAuth)
	{
		permissionRoutes.GET("/tree", systemController.ListPermissions)
		permissionRoutes.POST("/import", requireAdmin, systemController.ImportPermissions)
		permissionRoutes.GET("/audit", requireAdmin, systemController.PermissionAudit)
	}
}
