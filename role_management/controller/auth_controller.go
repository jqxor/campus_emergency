package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"role-management/service"
)

type AuthController struct {
	auth service.AuthService
}

func NewAuthController(auth service.AuthService) *AuthController {
	return &AuthController{auth: auth}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	RoleName string `json:"role_name"`
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Response{Error: "请求参数错误: " + err.Error()})
		return
	}
	token, user, err := c.auth.Login(ctx, req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, Response{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, Response{Data: gin.H{"token": token, "user": user}})
}

func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Response{Error: "请求参数错误: " + err.Error()})
		return
	}
	token, user, err := c.auth.Register(ctx, req.Username, req.Password, req.Email, req.RoleName)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, Response{Data: gin.H{"token": token, "user": user}})
}

func (c *AuthController) Me(ctx *gin.Context) {
	token := c.auth.ParseBearerToken(ctx.GetHeader("Authorization"))
	user, err := c.auth.Me(ctx, token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, Response{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, Response{Data: user})
}

func (c *AuthController) Logout(ctx *gin.Context) {
	token := c.auth.ParseBearerToken(ctx.GetHeader("Authorization"))
	if err := c.auth.Logout(ctx, token); err != nil {
		ctx.JSON(http.StatusUnauthorized, Response{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, Response{Data: true})
}

func (c *AuthController) RequireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := c.auth.ParseBearerToken(ctx.GetHeader("Authorization"))
		user, err := c.auth.Me(ctx, token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, Response{Error: err.Error()})
			return
		}
		ctx.Set("authUser", user)
		ctx.Next()
	}
}

func (c *AuthController) RequireAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v, ok := ctx.Get("authUser")
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, Response{Error: "未登录"})
			return
		}
		user, _ := v.(service.AuthUser)
		if !c.auth.IsAdmin(user) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, Response{Error: "权限不足"})
			return
		}
		ctx.Next()
	}
}
