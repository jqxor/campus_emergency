package repository

import (
	"context"
	"errors"

	"role-management/entity"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uint64) error
	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	List(ctx context.Context, page, pageSize int) ([]entity.User, int64, error)
	SetUserRole(ctx context.Context, userID, roleID uint64) error
	GetUserRoleID(ctx context.Context, userID uint64) (uint64, error)
	SetUserPermission(ctx context.Context, userID uint64, permission string) error
	GetUserPermission(ctx context.Context, userID uint64) (string, error)
}

type userRepository struct { db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository { return &userRepository{db: db} }

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&entity.UserRole{}).Error; err != nil { return err }
		if err := tx.Where("user_id = ?", id).Delete(&entity.UserPermission{}).Error; err != nil { return err }
		return tx.Delete(&entity.User{}, id).Error
	})
}

func (r *userRepository) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, page, pageSize int) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&total).Error; err != nil { return nil, 0, err }
	if err := r.db.WithContext(ctx).Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil { return nil, 0, err }
	return users, total, nil
}

func (r *userRepository) SetUserRole(ctx context.Context, userID, roleID uint64) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entity.UserRole{}).Error; err != nil { return err }
	if roleID == 0 { return nil }
	return r.db.WithContext(ctx).Create(&entity.UserRole{UserID: userID, RoleID: roleID}).Error
}

func (r *userRepository) GetUserRoleID(ctx context.Context, userID uint64) (uint64, error) {
	var ur entity.UserRole
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&ur).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return 0, nil }
		return 0, err
	}
	return ur.RoleID, nil
}

func (r *userRepository) SetUserPermission(ctx context.Context, userID uint64, permission string) error {
	up := entity.UserPermission{UserID: userID, Permission: permission}
	return r.db.WithContext(ctx).Save(&up).Error
}

func (r *userRepository) GetUserPermission(ctx context.Context, userID uint64) (string, error) {
	var up entity.UserPermission
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&up).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return "", nil }
		return "", err
	}
	return up.Permission, nil
}
