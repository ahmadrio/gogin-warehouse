package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"warehouse-management/internal/auth"
	"warehouse-management/internal/config"
	"warehouse-management/internal/handler"
	"warehouse-management/internal/middleware"
	"warehouse-management/internal/repository"
	"warehouse-management/internal/service"
)

func New(cfg *config.Config, db *gorm.DB) *gin.Engine {
	// ---- 1. Susun dependency: repository -> service -> handler ----
	tokens := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.TTL)

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, tokens)
	authHandler := handler.NewAuthHandler(authService)

	// ---- 2. Daftarkan route ----
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	// Route publik
	v1.POST("/auth/login", authHandler.Login)

	// Semua route di grup ini wajib login
	protected := v1.Group("")
	protected.Use(middleware.Authenticate(tokens, authService))

	protected.GET("/auth/me", authHandler.Me)

	return r
}
