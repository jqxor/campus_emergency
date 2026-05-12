package repository

import (
	"context"
	"errors"
	"strings"

	"role-management/entity"
	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	GetByID(ctx context.Context, id uint64) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, page, pageSize int) ([]entity.Role, int64, error)
	CheckUserAssociation(ctx context.Context, roleID uint64) (bool, error)
	GetRolePermissions(ctx context.Context, roleID uint64) ([]entity.Permission, error)
	AssignPermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error
	ClearPermissions(ctx context.Context, roleID uint64) error
	GetPermissionSummary(ctx context.Context) ([]entity.RolePermissionSummary, error)
}

type PermissionRepository interface {
	Create(ctx context.Context, permission *entity.Permission) error
	List(ctx context.Context) ([]entity.Permission, error)
}

type roleRepository struct { db *gorm.DB }
type permissionRepository struct { db *gorm.DB }

func NewRoleRepository(db *gorm.DB) RoleRepository { return &roleRepository{db: db} }
func NewPermissionRepository(db *gorm.DB) PermissionRepository { return &permissionRepository{db: db} }

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error { return r.db.WithContext(ctx).Create(role).Error }
func (r *roleRepository) GetByID(ctx context.Context, id uint64) (*entity.Role, error) {
	var role entity.Role
	if err := r.db.WithContext(ctx).First(&role, id).Error; err != nil { if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }; return nil, err }
	return &role, nil
}
func (r *roleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error; err != nil { if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }; return nil, err }
	return &role, nil
}
func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error { return r.db.WithContext(ctx).Save(role).Error }
func (r *roleRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&entity.RolePermission{}).Error; err != nil { return err }
		return tx.Delete(&entity.Role{}, id).Error
	})
}
func (r *roleRepository) List(ctx context.Context, page, pageSize int) ([]entity.Role, int64, error) {
	var roles []entity.Role; var total int64
	if err := r.db.WithContext(ctx).Model(&entity.Role{}).Count(&total).Error; err != nil { return nil, 0, err }
	if err := r.db.WithContext(ctx).Offset((page-1)*pageSize).Limit(pageSize).Find(&roles).Error; err != nil { return nil, 0, err }
	return roles, total, nil
}
func (r *roleRepository) CheckUserAssociation(ctx context.Context, roleID uint64) (bool, error) { var count int64; err := r.db.WithContext(ctx).Model(&entity.UserRole{}).Where("role_id = ?", roleID).Count(&count).Error; return count > 0, err }
func (r *roleRepository) GetRolePermissions(ctx context.Context, roleID uint64) ([]entity.Permission, error) {
	var permissions []entity.Permission
	err := r.db.WithContext(ctx).Table("permissions").Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").Where("role_permissions.role_id = ?", roleID).Find(&permissions).Error
	return permissions, err
}
func (r *roleRepository) AssignPermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error {
	if err := r.ClearPermissions(ctx, roleID); err != nil { return err }
	if len(permissionIDs) == 0 { return nil }
	items := make([]entity.RolePermission, 0, len(permissionIDs))
	for _, permissionID := range permissionIDs { items = append(items, entity.RolePermission{RoleID: roleID, PermissionID: permissionID}) }
	return r.db.WithContext(ctx).Create(&items).Error
}
func (r *roleRepository) ClearPermissions(ctx context.Context, roleID uint64) error { return r.db.WithContext(ctx).Where("role_id = ?", roleID).Delete(&entity.RolePermission{}).Error }
func (r *roleRepository) GetPermissionSummary(ctx context.Context) ([]entity.RolePermissionSummary, error) {
	var summaries []entity.RolePermissionSummary
	if err := r.db.WithContext(ctx).Raw(`
		SELECT r.id AS role_id, r.name AS role_name, COUNT(DISTINCT rp.permission_id) AS permission_count, GROUP_CONCAT(DISTINCT p.module) AS modules
		FROM roles r
		LEFT JOIN role_permissions rp ON r.id = rp.role_id
		LEFT JOIN permissions p ON rp.permission_id = p.id
		GROUP BY r.id, r.name
		ORDER BY r.id
	`).Scan(&summaries).Error; err != nil { return nil, err }
	for i, summary := range summaries { summaries[i].Modules = []string{}; if summary.Modules != nil && len(summary.Modules) > 0 { summaries[i].Modules = strings.Split(strings.Join(summary.Modules, ","), ",") } }
	return summaries, nil
}

func (r *permissionRepository) Create(ctx context.Context, permission *entity.Permission) error { return r.db.WithContext(ctx).Create(permission).Error }
func (r *permissionRepository) List(ctx context.Context) ([]entity.Permission, error) { var permissions []entity.Permission; if err := r.db.WithContext(ctx).Find(&permissions).Error; err != nil { return nil, err }; return permissions, nil }
