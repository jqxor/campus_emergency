package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"role-management/entity"
	"role-management/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthUser struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleID   uint64 `json:"role_id"`
	RoleName string `json:"role_name"`
}

type AuthService interface {
	Login(ctx context.Context, username, password string) (token string, user AuthUser, err error)
	Register(ctx context.Context, username, password, email, roleName string) (token string, user AuthUser, err error)
	Me(ctx context.Context, token string) (AuthUser, error)
	Logout(ctx context.Context, token string) error
	ParseBearerToken(header string) string
	IsAdmin(user AuthUser) bool
}

type authService struct {
	userRepo      repository.UserRepository
	roleRepo      repository.RoleRepository
	sessionStore  SessionStore
	sessionTTL    time.Duration
	singleSession bool
	defaultRole   string
	adminRoleName string
}

func NewAuthService(userRepo repository.UserRepository, roleRepo repository.RoleRepository, sessionStore SessionStore) AuthService {
	return &authService{
		userRepo:      userRepo,
		roleRepo:      roleRepo,
		sessionStore:  sessionStore,
		sessionTTL:    24 * time.Hour,
		singleSession: parseBoolEnv("SINGLE_SESSION_PER_USER", false),
		defaultRole:   "student",
		adminRoleName: "admin",
	}
}

func (s *authService) ParseBearerToken(header string) string {
	h := strings.TrimSpace(header)
	if h == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return h
}

func (s *authService) IsAdmin(user AuthUser) bool {
	return strings.EqualFold(user.RoleName, s.adminRoleName)
}

func (s *authService) Login(ctx context.Context, username, password string) (string, AuthUser, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return "", AuthUser{}, errors.New("用户名或密码不能为空")
	}

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", AuthUser{}, err
	}
	if user == nil {
		return "", AuthUser{}, errors.New("用户名或密码错误")
	}
	if !user.IsActive {
		return "", AuthUser{}, errors.New("账号已被禁用")
	}

	matched, err := verifyPassword(user.Password, password)
	if err != nil {
		return "", AuthUser{}, errors.New("密码校验失败")
	}
	if !matched {
		return "", AuthUser{}, errors.New("用户名或密码错误")
	}
	// Best-effort migration: transparently upgrade legacy plaintext passwords to bcrypt.
	if !isBcryptHash(user.Password) {
		hashed, hashErr := hashPassword(password)
		if hashErr == nil {
			user.Password = hashed
			_ = s.userRepo.Update(ctx, user)
		}
	}

	roleID, _ := s.userRepo.GetUserRoleID(ctx, user.ID)
	roleName := ""
	if roleID != 0 {
		role, _ := s.roleRepo.GetByID(ctx, roleID)
		if role != nil {
			roleName = role.Name
		}
	}

	if s.singleSession {
		s.sessionStore.DeleteByUser(user.ID)
	}
	token, _ := s.sessionStore.Create(user.ID, s.sessionTTL)
	return token, AuthUser{ID: user.ID, Username: user.Username, Email: user.Email, RoleID: roleID, RoleName: roleName}, nil
}

func (s *authService) Register(ctx context.Context, username, password, email, roleName string) (string, AuthUser, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	email = strings.TrimSpace(email)
	roleName = strings.TrimSpace(roleName)
	if roleName == "" {
		roleName = s.defaultRole
	}
	if username == "" || password == "" {
		return "", AuthUser{}, errors.New("用户名或密码不能为空")
	}
	if err := validatePassword(password); err != nil {
		return "", AuthUser{}, err
	}
	if email == "" {
		return "", AuthUser{}, errors.New("邮箱不能为空")
	}

	exists, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", AuthUser{}, err
	}
	if exists != nil {
		return "", AuthUser{}, errors.New("用户名已存在")
	}

	role, err := s.roleRepo.GetByName(ctx, roleName)
	if err != nil {
		return "", AuthUser{}, err
	}
	if role == nil {
		return "", AuthUser{}, errors.New("角色不存在")
	}

	hashed, err := hashPassword(password)
	if err != nil {
		return "", AuthUser{}, errors.New("密码处理失败")
	}

	created := &entity.User{Username: username, Password: hashed, Email: email, IsActive: true}
	if err := s.userRepo.Create(ctx, created); err != nil {
		return "", AuthUser{}, err
	}
	if err := s.userRepo.SetUserRole(ctx, created.ID, role.ID); err != nil {
		return "", AuthUser{}, err
	}

	if s.singleSession {
		s.sessionStore.DeleteByUser(created.ID)
	}
	token, _ := s.sessionStore.Create(created.ID, s.sessionTTL)
	return token, AuthUser{ID: created.ID, Username: created.Username, Email: created.Email, RoleID: role.ID, RoleName: role.Name}, nil
}

func (s *authService) Me(ctx context.Context, token string) (AuthUser, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return AuthUser{}, errors.New("未登录")
	}
	sess, ok := s.sessionStore.Get(token)
	if !ok {
		return AuthUser{}, errors.New("登录已过期")
	}
	user, err := s.userRepo.GetByID(ctx, sess.UserID)
	if err != nil {
		return AuthUser{}, err
	}
	if user == nil {
		return AuthUser{}, errors.New("用户不存在")
	}
	roleID, _ := s.userRepo.GetUserRoleID(ctx, user.ID)
	roleName := ""
	if roleID != 0 {
		role, _ := s.roleRepo.GetByID(ctx, roleID)
		if role != nil {
			roleName = role.Name
		}
	}
	return AuthUser{ID: user.ID, Username: user.Username, Email: user.Email, RoleID: roleID, RoleName: roleName}, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	_ = ctx
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("未登录")
	}
	if _, ok := s.sessionStore.Get(token); !ok {
		return errors.New("登录已过期")
	}
	s.sessionStore.Delete(token)
	return nil
}

func isBcryptHash(value string) bool {
	return strings.HasPrefix(value, "$2a$") || strings.HasPrefix(value, "$2b$") || strings.HasPrefix(value, "$2y$")
}

func verifyPassword(stored, provided string) (bool, error) {
	if isBcryptHash(stored) {
		err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(provided))
		if err == nil {
			return true, nil
		}
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return stored == provided, nil
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("密码至少需要 8 位")
	}
	hasLetter := false
	hasDigit := false
	for _, r := range password {
		if r >= '0' && r <= '9' {
			hasDigit = true
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("密码需同时包含字母和数字")
	}
	return nil
}

func parseBoolEnv(key string, def bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return def
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}
