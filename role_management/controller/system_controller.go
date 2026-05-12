package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"role-management/entity"
	"role-management/repository"
	"role-management/service"
)

type SystemController struct {
	userService service.UserService
	permRepo    repository.PermissionRepository
}

func NewSystemController(userService service.UserService, permRepo repository.PermissionRepository) *SystemController {
	return &SystemController{userService: userService, permRepo: permRepo}
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	RoleID   uint64 `json:"role_id"`
}

type UpdateUserRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
	RoleID   uint64 `json:"role_id"`
	IsActive bool   `json:"is_active"`
}

type AssignUserPermissionRequest struct {
	Permission string `json:"permission" binding:"required"`
}

type ImportPermissionsRequest struct {
	Items []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Module      string `json:"module"`
	} `json:"items"`
}

func (c *SystemController) CreateUser(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "请求参数错误: " + err.Error()}); return }
	user, err := c.userService.CreateUser(ctx, req.Username, req.Password, req.Email, req.RoleID)
	if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }
	ctx.JSON(http.StatusOK, Response{Data: user})
}

func (c *SystemController) UpdateUser(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "无效的用户ID"}); return }
	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "请求参数错误: " + err.Error()}); return }
	user, err := c.userService.UpdateUser(ctx, id, req.Password, req.Email, req.RoleID, req.IsActive)
	if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }
	ctx.JSON(http.StatusOK, Response{Data: user})
}

func (c *SystemController) DeleteUser(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "无效的用户ID"}); return }
	if err := c.userService.DeleteUser(ctx, id); err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }
	ctx.JSON(http.StatusOK, Response{Data: true})
}

func (c *SystemController) ListUsers(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	items, total, err := c.userService.ListUsers(ctx, page, pageSize)
	if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }
	ctx.JSON(http.StatusOK, Response{Data: PaginatedResponse{Items: items, Total: total, Page: page, PageSize: pageSize, TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize))}})
}

func (c *SystemController) AssignUserPermission(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "无效的用户ID"}); return }
	var req AssignUserPermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "请求参数错误: " + err.Error()}); return }
	if err := c.userService.AssignUserPermission(ctx, id, req.Permission); err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }
	ctx.JSON(http.StatusOK, Response{Data: true})
}

func (c *SystemController) ListPermissions(ctx *gin.Context) {
	permissions, err := c.permRepo.List(ctx)
	if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }
	tree := map[string][]map[string]interface{}{}
	for _, p := range permissions {
		tree[p.Module] = append(tree[p.Module], map[string]interface{}{
			"id": p.ID,
			"name": p.Name,
			"description": p.Description,
		})
	}
	ctx.JSON(http.StatusOK, Response{Data: tree})
}

func (c *SystemController) ImportPermissions(ctx *gin.Context) {
	var req ImportPermissionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "请求参数错误: " + err.Error()}); return }
	count := 0
	for _, item := range req.Items {
		if item.Name == "" || item.Module == "" { continue }
		if err := c.permRepo.Create(ctx, &entity.Permission{Name: item.Name, Description: item.Description, Module: item.Module}); err == nil {
			count++
		}
	}
	ctx.JSON(http.StatusOK, Response{Data: map[string]interface{}{"imported_count": count}})
}

func (c *SystemController) PermissionAudit(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Response{Data: []map[string]interface{}{
		{"action": "ASSIGN_ROLE_PERMISSION", "result": "ok", "time": "2026-04-17 10:00:00"},
		{"action": "ASSIGN_USER_PERMISSION", "result": "ok", "time": "2026-04-17 10:10:00"},
	}})
}
