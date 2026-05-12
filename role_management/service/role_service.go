package service

import (
	"context"
	"role-management/entity"
)

type RoleService interface {
	CreateRole(ctx context.Context, name, description string) (*entity.Role, error)
	GetRoleByID(ctx context.Context, id uint64) (*entity.Role, error)
	UpdateRole(ctx context.Context, id uint64, name, description string) (*entity.Role, error)
	DeleteRole(ctx context.Context, id uint64) error
	ListRoles(ctx context.Context, page, pageSize int) ([]entity.Role, int64, error)
	AssignRolePermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error
	GetRolePermissions(ctx context.Context, roleID uint64) ([]entity.Permission, error)
	ExportRoleList(ctx context.Context) (string, error)
}
