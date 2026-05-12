package service

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"role-management/entity"
	"role-management/repository"
)

func setupAuthTest(t *testing.T) (AuthService, repository.UserRepository, repository.RoleRepository) {
	t.Helper()
	t.Setenv("SINGLE_SESSION_PER_USER", "false")
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}

	err = db.AutoMigrate(
		&entity.Role{},
		&entity.Permission{},
		&entity.RolePermission{},
		&entity.UserRole{},
		&entity.User{},
		&entity.UserPermission{},
		&entity.SessionToken{},
	)
	if err != nil {
		t.Fatalf("migrate test db failed: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	ctx := context.Background()
	_ = roleRepo.Create(ctx, &entity.Role{Name: "student", Description: "student", IsActive: true})
	_ = roleRepo.Create(ctx, &entity.Role{Name: "admin", Description: "admin", IsActive: true})

	auth := NewAuthService(userRepo, roleRepo, NewInMemorySessionStore())
	return auth, userRepo, roleRepo
}

func TestRegisterStoresHashedPasswordAndLoginSuccess(t *testing.T) {
	auth, userRepo, _ := setupAuthTest(t)
	ctx := context.Background()

	token, _, err := auth.Register(ctx, "alice", "abc12345", "alice@example.com", "student")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if token == "" {
		t.Fatalf("expected token from register")
	}

	user, err := userRepo.GetByUsername(ctx, "alice")
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}
	if user == nil {
		t.Fatalf("expected alice to exist")
	}
	if user.Password == "abc12345" {
		t.Fatalf("password should not be stored in plaintext")
	}
	if !isBcryptHash(user.Password) {
		t.Fatalf("expected password to be bcrypt hash")
	}

	loginToken, _, err := auth.Login(ctx, "alice", "abc12345")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginToken == "" {
		t.Fatalf("expected token from login")
	}
}

func TestLoginMigratesLegacyPlaintextPassword(t *testing.T) {
	auth, userRepo, roleRepo := setupAuthTest(t)
	ctx := context.Background()

	studentRole, err := roleRepo.GetByName(ctx, "student")
	if err != nil || studentRole == nil {
		t.Fatalf("prepare role failed: %v", err)
	}
	legacy := &entity.User{Username: "legacy", Password: "legacy123a", Email: "legacy@example.com", IsActive: true}
	if err := userRepo.Create(ctx, legacy); err != nil {
		t.Fatalf("create legacy user failed: %v", err)
	}
	if err := userRepo.SetUserRole(ctx, legacy.ID, studentRole.ID); err != nil {
		t.Fatalf("bind role failed: %v", err)
	}

	_, _, err = auth.Login(ctx, "legacy", "legacy123a")
	if err != nil {
		t.Fatalf("legacy login should succeed: %v", err)
	}

	refreshed, err := userRepo.GetByUsername(ctx, "legacy")
	if err != nil {
		t.Fatalf("reload legacy user failed: %v", err)
	}
	if refreshed == nil {
		t.Fatalf("expected legacy user to exist")
	}
	if refreshed.Password == "legacy123a" {
		t.Fatalf("legacy plaintext password should be migrated")
	}
	if !isBcryptHash(refreshed.Password) {
		t.Fatalf("expected migrated password to be bcrypt hash")
	}
}

func TestRegisterRejectsWeakPassword(t *testing.T) {
	auth, _, _ := setupAuthTest(t)
	ctx := context.Background()

	_, _, err := auth.Register(ctx, "weak", "1234567", "weak@example.com", "student")
	if err == nil {
		t.Fatalf("expected weak password register to fail")
	}
}

func TestLogoutInvalidatesToken(t *testing.T) {
	auth, _, _ := setupAuthTest(t)
	ctx := context.Background()

	token, _, err := auth.Register(ctx, "logout_user", "abc12345", "logout@example.com", "student")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if err := auth.Logout(ctx, token); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	_, err = auth.Me(ctx, token)
	if err == nil {
		t.Fatalf("expected token to be invalid after logout")
	}
}

func TestSingleSessionPerUserEnabled(t *testing.T) {
	t.Setenv("SINGLE_SESSION_PER_USER", "true")
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}

	err = db.AutoMigrate(
		&entity.Role{},
		&entity.Permission{},
		&entity.RolePermission{},
		&entity.UserRole{},
		&entity.User{},
		&entity.UserPermission{},
		&entity.SessionToken{},
	)
	if err != nil {
		t.Fatalf("migrate test db failed: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	ctx := context.Background()
	_ = roleRepo.Create(ctx, &entity.Role{Name: "student", Description: "student", IsActive: true})

	auth := NewAuthService(userRepo, roleRepo, NewInMemorySessionStore())

	_, _, err = auth.Register(ctx, "single_user", "abc12345", "single@example.com", "student")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	firstToken, _, err := auth.Login(ctx, "single_user", "abc12345")
	if err != nil {
		t.Fatalf("first login failed: %v", err)
	}
	secondToken, _, err := auth.Login(ctx, "single_user", "abc12345")
	if err != nil {
		t.Fatalf("second login failed: %v", err)
	}

	if firstToken == secondToken {
		t.Fatalf("expected distinct tokens between two logins")
	}

	if _, err := auth.Me(ctx, firstToken); err == nil {
		t.Fatalf("expected first token to be invalidated in single-session mode")
	}
	if _, err := auth.Me(ctx, secondToken); err != nil {
		t.Fatalf("expected second token to remain valid: %v", err)
	}
}
