package controller

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"role-management/entity"
	"role-management/service"
)

type RoleController struct { roleService service.RoleService }
func NewRoleController(roleService service.RoleService) *RoleController { return &RoleController{roleService: roleService} }

type CreateRoleRequest struct { Name string `json:"name" binding:"required"`; Description string `json:"description"` }
type UpdateRoleRequest struct { Name string `json:"name" binding:"required"`; Description string `json:"description"` }
type AssignPermissionsRequest struct { PermissionIDs []uint64 `json:"permission_ids"` }
type Response struct { Data interface{} `json:"data,omitempty"`; Error string `json:"error,omitempty"` }
type PaginatedResponse struct { Items interface{} `json:"items"`; Total int64 `json:"total"`; Page int `json:"page"`; PageSize int `json:"page_size"`; TotalPages int `json:"total_pages"` }

func (c *RoleController) CreateRole(ctx *gin.Context) { var req CreateRoleRequest; if err := ctx.ShouldBindJSON(&req); err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "请求参数错误: " + err.Error()}); return }; role, err := c.roleService.CreateRole(ctx, req.Name, req.Description); if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }; ctx.JSON(http.StatusOK, Response{Data: role}) }
func (c *RoleController) GetRole(ctx *gin.Context) { id, err := strconv.ParseUint(ctx.Param("id"), 10, 64); if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "无效的角色ID"}); return }; role, err := c.roleService.GetRoleByID(ctx, id); if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }; ctx.JSON(http.StatusOK, Response{Data: role}) }
func (c *RoleController) UpdateRole(ctx *gin.Context) { id, err := strconv.ParseUint(ctx.Param("id"), 10, 64); if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "无效的角色ID"}); return }; var req UpdateRoleRequest; if err := ctx.ShouldBindJSON(&req); err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "请求参数错误: " + err.Error()}); return }; role, err := c.roleService.UpdateRole(ctx, id, req.Name, req.Description); if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }; ctx.JSON(http.StatusOK, Response{Data: role}) }
func (c *RoleController) DeleteRole(ctx *gin.Context) { id, err := strconv.ParseUint(ctx.Param("id"), 10, 64); if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "无效的角色ID"}); return }; if err := c.roleService.DeleteRole(ctx, id); err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }; ctx.JSON(http.StatusOK, Response{Data: true}) }
func (c *RoleController) ListRoles(ctx *gin.Context) { page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1")); pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20")); roles, total, err := c.roleService.ListRoles(ctx, page, pageSize); if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }; ctx.JSON(http.StatusOK, Response{Data: PaginatedResponse{Items: roles, Total: total, Page: page, PageSize: pageSize, TotalPages: int((total+int64(pageSize)-1)/int64(pageSize))}}) }
func (c *RoleController) AssignPermissions(ctx *gin.Context) { id, err := strconv.ParseUint(ctx.Param("id"), 10, 64); if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "无效的角色ID"}); return }; var req AssignPermissionsRequest; if err := ctx.ShouldBindJSON(&req); err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "请求参数错误: " + err.Error()}); return }; if err := c.roleService.AssignRolePermissions(ctx, id, req.PermissionIDs); err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }; ctx.JSON(http.StatusOK, Response{Data: true}) }
func (c *RoleController) GetRolePermissions(ctx *gin.Context) { id, err := strconv.ParseUint(ctx.Param("id"), 10, 64); if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: "无效的角色ID"}); return }; permissions, err := c.roleService.GetRolePermissions(ctx, id); if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }; ctx.JSON(http.StatusOK, Response{Data: permissions}) }
func (c *RoleController) ExportRoles(ctx *gin.Context) { filePath, err := c.roleService.ExportRoleList(ctx); if err != nil { ctx.JSON(http.StatusBadRequest, Response{Error: err.Error()}); return }; ctx.Header("Content-Type", "text/csv"); ctx.Header("Content-Disposition", "attachment; filename=roles_export.csv"); ctx.File(filePath); go func() { time.Sleep(5 * time.Minute); _ = os.Remove(filePath) }() }

var _ = entity.Role{}
