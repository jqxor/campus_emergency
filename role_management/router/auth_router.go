package router

import (
	"github.com/gin-gonic/gin"
	"role-management/controller"
)

func RegisterAuthRoutes(r *gin.Engine, authController *controller.AuthController) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authController.Login)
		auth.POST("/register", authController.Register)
		auth.GET("/me", authController.Me)
		auth.POST("/logout", authController.Logout)
	}
}
