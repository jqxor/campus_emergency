package service

import (
	"context"
	"errors"
	"strings"

	"role-management/entity"
	"role-management/repository"
)

type UserService interface {
	CreateUser(ctx context.Context, username, password, email string, roleID uint64) (*entity.User, error)
	UpdateUser(ctx context.Context, id uint64, password, email string, roleID uint64, isActive bool) (*entity.User, error)
	DeleteUser(ctx context.Context, id uint64) error
	ListUsers(ctx context.Context, page, pageSize int) ([]map[string]interface{}, int64, error)
	AssignUserPermission(ctx context.Context, userID uint64, permission string) error
}

type userService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
	auditLogService AuditLogService
}

func NewUserService(userRepo repository.UserRepository, roleRepo repository.RoleRepository, auditLogService AuditLogService) UserService {
	return &userService{userRepo: userRepo, roleRepo: roleRepo, auditLogService: auditLogService}
}

func (s *userService) CreateUser(ctx context.Context, username, password, email string, roleID uint64) (*entity.User, error) {
	if strings.TrimSpace(username) == "" { return nil, errors.New("用户名不能为空") }
	if strings.TrimSpace(password) == "" { return nil, errors.New("密码不能为空") }
	if strings.TrimSpace(email) == "" { return nil, errors.New("邮箱不能为空") }
	exists, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil { return nil, err }
	if exists != nil { return nil, errors.New("用户名已存在") }
	if roleID != 0 {
		role, err := s.roleRepo.GetByID(ctx, roleID)
		if err != nil { return nil, err }
		if role == nil { return nil, errors.New("角色不存在") }
	}
	user := &entity.User{Username: username, Password: password, Email: email, IsActive: true}
	if err := s.userRepo.Create(ctx, user); err != nil { return nil, err }
	if err := s.userRepo.SetUserRole(ctx, user.ID, roleID); err != nil { return nil, err }
	_ = s.auditLogService.LogAction(ctx, "CREATE_USER", "创建用户", map[string]interface{}{"user_id": user.ID, "username": user.Username})
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, id uint64, password, email string, roleID uint64, isActive bool) (*entity.User, error) {
	if id == 0 { return nil, errors.New("用户ID不能为空") }
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil { return nil, err }
	if user == nil { return nil, errors.New("用户不存在") }
	if strings.TrimSpace(password) != "" { user.Password = password }
	if strings.TrimSpace(email) != "" { user.Email = email }
	user.IsActive = isActive
	if err := s.userRepo.Update(ctx, user); err != nil { return nil, err }
	if roleID != 0 {
		if _, err := s.roleRepo.GetByID(ctx, roleID); err != nil { return nil, err }
		if err := s.userRepo.SetUserRole(ctx, user.ID, roleID); err != nil { return nil, err }
	}
	_ = s.auditLogService.LogAction(ctx, "UPDATE_USER", "更新用户", map[string]interface{}{"user_id": user.ID})
	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id uint64) error {
	if id == 0 { return errors.New("用户ID不能为空") }
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil { return err }
	if user == nil { return errors.New("用户不存在") }
	if err := s.userRepo.Delete(ctx, id); err != nil { return err }
	_ = s.auditLogService.LogAction(ctx, "DELETE_USER", "删除用户", map[string]interface{}{"user_id": id})
	return nil
}

func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]map[string]interface{}, int64, error) {
	if page <= 0 { page = 1 }
	if pageSize <= 0 || pageSize > 100 { pageSize = 20 }
	users, total, err := s.userRepo.List(ctx, page, pageSize)
	if err != nil { return nil, 0, err }
	items := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		roleID, _ := s.userRepo.GetUserRoleID(ctx, u.ID)
		perm, _ := s.userRepo.GetUserPermission(ctx, u.ID)
		items = append(items, map[string]interface{}{
			"id": u.ID,
			"username": u.Username,
			"email": u.Email,
			"is_active": u.IsActive,
			"role_id": roleID,
			"permission": perm,
			"created_at": u.CreatedAt,
			"updated_at": u.UpdatedAt,
		})
	}
	return items, total, nil
}

func (s *userService) AssignUserPermission(ctx context.Context, userID uint64, permission string) error {
	if userID == 0 { return errors.New("用户ID不能为空") }
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil { return err }
	if user == nil { return errors.New("用户不存在") }
	if strings.TrimSpace(permission) == "" { return errors.New("权限内容不能为空") }
	if err := s.userRepo.SetUserPermission(ctx, userID, permission); err != nil { return err }
	_ = s.auditLogService.LogAction(ctx, "ASSIGN_USER_PERMISSION", "分配用户权限", map[string]interface{}{"user_id": userID})
	return nil
}
