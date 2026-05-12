package service

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"role-management/entity"
	"role-management/repository"
)

type roleService struct {
	roleRepo        repository.RoleRepository
	permRepo        repository.PermissionRepository
	auditLogService AuditLogService
}

func NewRoleService(roleRepo repository.RoleRepository, permRepo repository.PermissionRepository, auditLogService AuditLogService) RoleService {
	return &roleService{roleRepo: roleRepo, permRepo: permRepo, auditLogService: auditLogService}
}

func (s *roleService) CreateRole(ctx context.Context, name, description string) (*entity.Role, error) {
	if strings.TrimSpace(name) == "" { return nil, errors.New("角色名称不能为空") }
	existingRole, err := s.roleRepo.GetByName(ctx, name)
	if err != nil { return nil, err }
	if existingRole != nil { return nil, errors.New("角色名称已存在") }
	role := &entity.Role{Name: name, Description: description, IsActive: true}
	if err := s.roleRepo.Create(ctx, role); err != nil { return nil, err }
	_ = s.auditLogService.LogAction(ctx, "CREATE_ROLE", "创建新角色", map[string]interface{}{"role_id": role.ID, "role_name": role.Name})
	return role, nil
}

func (s *roleService) GetRoleByID(ctx context.Context, id uint64) (*entity.Role, error) {
	if id == 0 { return nil, errors.New("角色ID不能为空") }
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil { return nil, err }
	if role == nil { return nil, errors.New("角色不存在") }
	return role, nil
}

func (s *roleService) UpdateRole(ctx context.Context, id uint64, name, description string) (*entity.Role, error) {
	if id == 0 { return nil, errors.New("角色ID不能为空") }
	if strings.TrimSpace(name) == "" { return nil, errors.New("角色名称不能为空") }
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil { return nil, err }
	if role == nil { return nil, errors.New("角色不存在") }
	if name != role.Name {
		existingRole, err := s.roleRepo.GetByName(ctx, name)
		if err != nil { return nil, err }
		if existingRole != nil && existingRole.ID != id { return nil, errors.New("角色名称已存在") }
	}
	oldName := role.Name
	role.Name = name
	role.Description = description
	if err := s.roleRepo.Update(ctx, role); err != nil { return nil, err }
	_ = s.auditLogService.LogAction(ctx, "UPDATE_ROLE", "更新角色信息", map[string]interface{}{"role_id": role.ID, "old_name": oldName, "new_name": role.Name})
	return role, nil
}

func (s *roleService) DeleteRole(ctx context.Context, id uint64) error {
	if id == 0 { return errors.New("角色ID不能为空") }
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil { return err }
	if role == nil { return errors.New("角色不存在") }
	hasUsers, err := s.roleRepo.CheckUserAssociation(ctx, id)
	if err != nil { return err }
	if hasUsers { return errors.New("无法删除已关联用户的角色，请先解除用户关联") }
	if err := s.roleRepo.Delete(ctx, id); err != nil { return err }
	_ = s.auditLogService.LogAction(ctx, "DELETE_ROLE", "删除角色", map[string]interface{}{"role_id": id, "role_name": role.Name})
	return nil
}

func (s *roleService) ListRoles(ctx context.Context, page, pageSize int) ([]entity.Role, int64, error) {
	if page <= 0 { page = 1 }
	if pageSize <= 0 || pageSize > 100 { pageSize = 20 }
	return s.roleRepo.List(ctx, page, pageSize)
}

func (s *roleService) AssignRolePermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error {
	if roleID == 0 { return errors.New("角色ID不能为空") }
	return s.roleRepo.AssignPermissions(ctx, roleID, permissionIDs)
}

func (s *roleService) GetRolePermissions(ctx context.Context, roleID uint64) ([]entity.Permission, error) {
	if roleID == 0 { return nil, errors.New("角色ID不能为空") }
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil { return nil, err }
	if role == nil { return nil, errors.New("角色不存在") }
	return s.roleRepo.GetRolePermissions(ctx, roleID)
}

func (s *roleService) ExportRoleList(ctx context.Context) (string, error) {
	summaries, err := s.roleRepo.GetPermissionSummary(ctx)
	if err != nil { return "", err }
	filePath := fmt.Sprintf("role_export_%s.csv", time.Now().Format("20060102150405"))
	file, err := os.Create(filePath)
	if err != nil { return "", err }
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.Write([]string{"角色ID", "角色名称", "权限数量", "访问模块"}); err != nil { return "", err }
	for _, summary := range summaries {
		if err := writer.Write([]string{strconv.FormatUint(summary.RoleID, 10), summary.RoleName, strconv.Itoa(summary.PermissionCount), strings.Join(summary.Modules, ",")}); err != nil { return "", err }
	}
	_ = s.auditLogService.LogAction(ctx, "EXPORT_ROLES", "导出角色列表", map[string]interface{}{"file_path": filePath, "count": len(summaries)})
	return filePath, nil
}
