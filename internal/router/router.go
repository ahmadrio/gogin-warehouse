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
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewProductCategoryRepository(db)
	locationRepo := repository.NewWarehouseLocationRepository(db)
	warehouseRepo := repository.NewWarehouseRepository(db)
	lookupRepo := repository.NewLookupRepository(db)
	sequenceRepo := repository.NewSequenceRepository(db)
	stockRepo := repository.NewStockRepository(db)
	adjustmentRepo := repository.NewStockAdjustmentRepository(db)
	outboundRepo := repository.NewOutboundRepository(db)
	transferRepo := repository.NewStockTransferRepository(db)
	inboundRepo := repository.NewInboundRepository(db)

	authService := service.NewAuthService(userRepo, tokens)
	productService := service.NewProductService(db, productRepo, categoryRepo, auditRepo)
	categoryService := service.NewProductCategoryService(db, categoryRepo, auditRepo)
	locationService := service.NewWarehouseLocationService(db, locationRepo, warehouseRepo, auditRepo)
	warehouseService := service.NewWarehouseService(db, warehouseRepo, auditRepo)
	stockService := service.NewStockService(stockRepo)
	adjustmentService := service.NewStockAdjustmentService(db, adjustmentRepo, warehouseRepo, lookupRepo, sequenceRepo, auditRepo, stockService)
	outboundService := service.NewOutboundService(db, outboundRepo, warehouseRepo, lookupRepo, sequenceRepo, auditRepo, stockService)
	transferService := service.NewStockTransferService(db, transferRepo, warehouseRepo, lookupRepo, sequenceRepo, auditRepo, stockService)
	inboundService := service.NewInboundService(db, inboundRepo, warehouseRepo, lookupRepo, sequenceRepo, auditRepo, stockService)

	authHandler := handler.NewAuthHandler(authService)
	productHandler := handler.NewProductHandler(productService)
	categoryHandler := handler.NewProductCategoryHandler(categoryService)
	locationHandler := handler.NewWarehouseLocationHandler(locationService)
	warehouseHandler := handler.NewWarehouseHandler(warehouseService)
	stockHandler := handler.NewStockHandler(stockService)
	adjustmentHandler := handler.NewStockAdjustmentHandler(adjustmentService)
	outboundHandler := handler.NewOutboundHandler(outboundService)
	transferHandler := handler.NewStockTransferHandler(transferService)
	inboundHandler := handler.NewInboundHandler(inboundService)

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

	products := protected.Group("/products")
	{
		products.GET("", can(auth.ResProducts, auth.OpRead), productHandler.List)
		products.GET("/:id", can(auth.ResProducts, auth.OpRead), productHandler.Get)
		products.POST("", can(auth.ResProducts, auth.OpCreate), productHandler.Create)
		products.PUT("/:id", can(auth.ResProducts, auth.OpUpdate), productHandler.Update)
		products.DELETE("/:id", can(auth.ResProducts, auth.OpDelete), productHandler.Delete)
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
		warehouses.POST("", can(auth.ResWarehouses, auth.OpCreate), warehouseHandler.Create)

		// lokasi (rak) di dalam gudang
		warehouses.GET("/:id/locations", can(auth.ResWarehouseLocations, auth.OpRead), warehouseHandler.ListLocations)
		warehouses.GET("/:id/locations/:location_id", can(auth.ResWarehouseLocations, auth.OpRead), locationHandler.Get)
		warehouses.POST("/:id/locations", can(auth.ResWarehouseLocations, auth.OpCreate), locationHandler.Create)
		warehouses.PUT("/:id/locations/:location_id", can(auth.ResWarehouseLocations, auth.OpUpdate), locationHandler.Update)
		warehouses.DELETE("/:id/locations/:location_id", can(auth.ResWarehouseLocations, auth.OpDelete), locationHandler.Delete)
	}

	protected.GET("/stocks", can(auth.ResStocks, auth.OpRead), stockHandler.List)

	adjustments := protected.Group("/stock-adjustments")
	{
		adjustments.GET("", can(auth.ResStockAdjustments, auth.OpRead), adjustmentHandler.List)
		adjustments.GET("/:id", can(auth.ResStockAdjustments, auth.OpRead), adjustmentHandler.Get)
		adjustments.POST("", can(auth.ResStockAdjustments, auth.OpCreate), adjustmentHandler.Create)
		adjustments.PUT("/:id", can(auth.ResStockAdjustments, auth.OpUpdate), adjustmentHandler.Update)
		adjustments.POST("/:id/approve", can(auth.ResStockAdjustments, auth.OpApprove), adjustmentHandler.Approve)
		adjustments.POST("/:id/reject", can(auth.ResStockAdjustments, auth.OpReject), adjustmentHandler.Reject)
	}

	outbounds := protected.Group("/outbounds")
	{
		outbounds.GET("", can(auth.ResOutbounds, auth.OpRead), outboundHandler.List)
		outbounds.GET("/:id", can(auth.ResOutbounds, auth.OpRead), outboundHandler.Get)
		outbounds.POST("", can(auth.ResOutbounds, auth.OpCreate), outboundHandler.Create)
		outbounds.PUT("/:id", can(auth.ResOutbounds, auth.OpUpdate), outboundHandler.Update)
		outbounds.POST("/:id/approve", can(auth.ResOutbounds, auth.OpApprove), outboundHandler.Approve)
		outbounds.POST("/:id/reject", can(auth.ResOutbounds, auth.OpReject), outboundHandler.Reject)
	}

	transfers := protected.Group("/stock-transfers")
	{
		transfers.GET("", can(auth.ResStockTransfers, auth.OpRead), transferHandler.List)
		transfers.GET("/:id", can(auth.ResStockTransfers, auth.OpRead), transferHandler.Get)
		transfers.POST("", can(auth.ResStockTransfers, auth.OpCreate), transferHandler.Create)
		transfers.PUT("/:id", can(auth.ResStockTransfers, auth.OpUpdate), transferHandler.Update)
		transfers.POST("/:id/approve", can(auth.ResStockTransfers, auth.OpApprove), transferHandler.Approve)
		transfers.POST("/:id/reject", can(auth.ResStockTransfers, auth.OpReject), transferHandler.Reject)
	}

	inbounds := protected.Group("/inbounds")
	{
		inbounds.GET("", can(auth.ResInbounds, auth.OpRead), inboundHandler.List)
		inbounds.GET("/:id", can(auth.ResInbounds, auth.OpRead), inboundHandler.Get)
		inbounds.POST("", can(auth.ResInbounds, auth.OpCreate), inboundHandler.Create)
		inbounds.PUT("/:id", can(auth.ResInbounds, auth.OpUpdate), inboundHandler.Update)
		inbounds.POST("/:id/approve", can(auth.ResInbounds, auth.OpApprove), inboundHandler.Approve)
		inbounds.POST("/:id/reject", can(auth.ResInbounds, auth.OpReject), inboundHandler.Reject)
		// Putaway dikerjakan staff gudang -> pakai permission update.
		inbounds.POST("/:id/putaway", can(auth.ResInbounds, auth.OpUpdate), inboundHandler.Putaway)
	}

	return r
}
