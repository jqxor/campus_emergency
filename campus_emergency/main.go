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

	"campus-emergency/controller"
	"campus-emergency/model"
	"campus-emergency/repository"
	"campus-emergency/router"
	"campus-emergency/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
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

func isPublicRoute(path string) bool {
	return path == "/healthz" || path == "/readyz"
}

func main() {
	addr := normalizeAddr(getenv("PORT", "8081"))
	dbPath := getenv("DB_PATH", "campus_emergency.db")
	defaultOrigins := parseAllowedOrigins(getenv("CORS_ALLOWED_ORIGINS", "*"))
	publicOrigins := parseAllowedOrigins(getenv("CORS_PUBLIC_ALLOWED_ORIGINS", strings.Join(defaultOrigins, ",")))
	apiOrigins := parseAllowedOrigins(getenv("CORS_API_ALLOWED_ORIGINS", strings.Join(defaultOrigins, ",")))

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db failed: %v", err)
	}
	sqlDB, _ := db.DB()
	if err := db.AutoMigrate(&model.EmergencyPlan{}); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}

	planRepo := repository.NewPlanRepository(db)
	aiSvc := service.NewAIOptimizationService()
	planSvc := service.NewPlanService(planRepo, aiSvc)
	planController := controller.NewPlanController(planSvc)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "campus_emergency"})
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
		currentAllowed := apiOrigins
		if isPublicRoute(path) {
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
	router.RegisterPlanRoutes(r, planController)

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	if sqlDB != nil {
		_ = sqlDB.Close()
	}
}
