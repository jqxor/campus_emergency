package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"role-management/controller"
	"role-management/entity"
	"role-management/repository"
	"role-management/router"
	"role-management/service"
	"golang.org/x/crypto/bcrypt"
)

func getenv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func normalizeAddr(port string) string {
	p := strings.TrimSpace(port)
	if p == "" {
		return ":0"
	}
	if strings.HasPrefix(p, ":") {
		return p
	}
	return ":" + p
}

func parseAllowedOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return []string{"*"}
	}
	return result
}

func isOriginAllowed(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}
	if slices.Contains(allowed, "*") {
		return true
	}
	return slices.Contains(allowed, origin)
}

func isPublicRoleRoute(path string) bool {
	if path == "/healthz" || path == "/readyz" {
		return true
	}
	if strings.HasPrefix(path, "/api/auth/login") || strings.HasPrefix(path, "/api/auth/register") {
		return true
	}
	return false
}

func applyCORS(c *gin.Context, allowed []string) {
	origin := strings.TrimSpace(c.GetHeader("Origin"))
	if origin == "" {
		if slices.Contains(allowed, "*") {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		return
	}
	if !isOriginAllowed(origin, allowed) {
		return
	}
	if slices.Contains(allowed, "*") {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		return
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Vary", "Origin")
}

func seedDefaults(ctx context.Context, roleRepo repository.RoleRepository, userRepo repository.UserRepository) {
	ensureRole := func(name, desc string) *entity.Role {
		role, err := roleRepo.GetByName(ctx, name)
		if err == nil && role != nil {
			return role
		}
		r := &entity.Role{Name: name, Description: desc, IsActive: true}
		_ = roleRepo.Create(ctx, r)
		role, _ = roleRepo.GetByName(ctx, name)
		return role
	}

	adminRole := ensureRole("admin", "系统管理员")
	_ = ensureRole("teacher", "教师")
	_ = ensureRole("student", "学生")

	if adminRole == nil {
		return
	}

	admin, err := userRepo.GetByUsername(ctx, "admin")
	if err != nil {
		return
	}
	adminInitPassword := strings.TrimSpace(os.Getenv("ADMIN_INIT_PASSWORD"))
	if admin == nil {
		if adminInitPassword == "" {
			log.Printf("[warn] admin account not created: ADMIN_INIT_PASSWORD is empty")
			return
		}
		hashed, hashErr := bcrypt.GenerateFromPassword([]byte(adminInitPassword), bcrypt.DefaultCost)
		if hashErr != nil {
			log.Printf("[warn] admin account not created: hash password failed: %v", hashErr)
			return
		}
		admin = &entity.User{Username: "admin", Password: string(hashed), Email: "admin@campus.edu", IsActive: true}
		_ = userRepo.Create(ctx, admin)
		log.Printf("[info] admin account initialized from environment")
	}
	if admin != nil {
		roleID, _ := userRepo.GetUserRoleID(ctx, admin.ID)
		if roleID == 0 {
			_ = userRepo.SetUserRole(ctx, admin.ID, adminRole.ID)
		}
	}
}

func main() {
	addr := normalizeAddr(getenv("PORT", "8082"))
	dbPath := getenv("DB_PATH", "role_management.db")
	defaultOrigins := parseAllowedOrigins(getenv("CORS_ALLOWED_ORIGINS", "*"))
	publicOrigins := parseAllowedOrigins(getenv("CORS_PUBLIC_ALLOWED_ORIGINS", strings.Join(defaultOrigins, ",")))
	protectedOrigins := parseAllowedOrigins(getenv("CORS_PROTECTED_ALLOWED_ORIGINS", strings.Join(defaultOrigins, ",")))

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db failed: %v", err)
	}
	sqlDB, _ := db.DB()
	if err := db.AutoMigrate(&entity.Role{}, &entity.Permission{}, &entity.RolePermission{}, &entity.UserRole{}, &entity.User{}, &entity.UserPermission{}, &entity.SessionToken{}); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}

	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	seedDefaults(context.Background(), roleRepo, userRepo)

	sessionStore := service.NewDBSessionStore(db)
	authSvc := service.NewAuthService(userRepo, roleRepo, sessionStore)
	audit := service.NoopAuditLogService{}
	roleSvc := service.NewRoleService(roleRepo, permRepo, audit)
	userSvc := service.NewUserService(userRepo, roleRepo, audit)
	roleController := controller.NewRoleController(roleSvc)
	systemController := controller.NewSystemController(userSvc, permRepo)
	authController := controller.NewAuthController(authSvc)

	cleanupStop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				deleted := sessionStore.CleanupExpired()
				if deleted > 0 {
					log.Printf("[info] cleaned expired sessions: %d", deleted)
				}
			case <-cleanupStop:
				return
			}
		}
	}()

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "role_management"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		if sqlDB == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "reason": "db"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 800*time.Millisecond)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "reason": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		isPublic := isPublicRoleRoute(path)
		currentAllowed := protectedOrigins
		if isPublic {
			currentAllowed = publicOrigins
		}

		origin := strings.TrimSpace(c.GetHeader("Origin"))
		applyCORS(c, currentAllowed)
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			if origin != "" && !isOriginAllowed(origin, currentAllowed) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	router.RegisterAuthRoutes(r, authController)
	router.RegisterRoleRoutes(r, roleController, systemController, authController)

	srv := &http.Server{Addr: addr, Handler: r}
	go func() {
		log.Printf("server listening on %s (db=%s)", addr, dbPath)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server start failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	close(cleanupStop)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	if sqlDB != nil {
		_ = sqlDB.Close()
	}
}
