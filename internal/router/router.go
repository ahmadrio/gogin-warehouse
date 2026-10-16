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
	auditRepo := repository.NewAuditRepository(db)
	categoryRepo := repository.NewProductCategoryRepository(db)
	warehouseRepo := repository.NewWarehouseRepository(db)

	authService := service.NewAuthService(userRepo, tokens)
	categoryService := service.NewProductCategoryService(db, categoryRepo, auditRepo)
	warehouseService := service.NewWarehouseService(db, warehouseRepo, auditRepo)

	authHandler := handler.NewAuthHandler(authService)
	categoryHandler := handler.NewProductCategoryHandler(categoryService)
	warehouseHandler := handler.NewWarehouseHandler(warehouseService)

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

	// can() hanya singkatan agar daftar route mudah dibaca
	can := func(resource, operation string) gin.HandlerFunc {
		return middleware.RequirePermission(auth.Perm(resource, operation))
	}

	categories := protected.Group("/product-categories")
	{
		categories.GET("", can(auth.ResProductCategories, auth.OpRead), categoryHandler.List)
		categories.GET("/:id", can(auth.ResProductCategories, auth.OpRead), categoryHandler.Get)
		categories.POST("", can(auth.ResProductCategories, auth.OpCreate), categoryHandler.Create)
		categories.PUT("/:id", can(auth.ResProductCategories, auth.OpUpdate), categoryHandler.Update)
		categories.DELETE("/:id", can(auth.ResProductCategories, auth.OpDelete), categoryHandler.Delete)
	}

	warehouses := protected.Group("/warehouses")
	{
		warehouses.GET("", can(auth.ResWarehouses, auth.OpRead), warehouseHandler.List)
		warehouses.GET("/:id", can(auth.ResWarehouses, auth.OpRead), warehouseHandler.Get)
		warehouses.GET("/:id/locations", can(auth.ResWarehouseLocations, auth.OpRead), warehouseHandler.ListLocations)
		warehouses.POST("", can(auth.ResWarehouses, auth.OpCreate), warehouseHandler.Create)
	}

	return r
}
