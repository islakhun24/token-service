package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"token-service/internal/config"
	"token-service/internal/handler"
	"token-service/internal/infrastructure/binance"
	"token-service/internal/infrastructure/postgres"
	"token-service/internal/infrastructure/staticmapping"
	"token-service/internal/service"
	"token-service/pkg/normalizer"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Database
	db, err := postgres.NewDB(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect database", zap.Error(err))
	}
	defer db.Close()

	// Run migrations
	if err := postgres.Migrate(db); err != nil {
		logger.Fatal("Failed to migrate", zap.Error(err))
	}

	// Normalizer
	norm := normalizer.NewSymbolNormalizer(cfg.Normalizer.Aliases)

	// Static mapping loader
	staticLoader, err := staticmapping.NewLoader(cfg.StaticMapping.FilePath)
	if err != nil {
		logger.Warn("Static mapping not loaded, using empty", zap.Error(err))
	}

	// Exchange clients
	binanceClient := binance.NewClient(cfg.Binance.BaseURL)

	// Repository
	repo := postgres.NewRegistryRepository(db)

	// Services
	registryService := service.NewRegistryService(repo, norm, staticLoader, binanceClient, logger)
	syncService := service.NewSyncService(repo, norm, binanceClient, logger)

	// Load initial data
	ctx := context.Background()
	if err := registryService.Initialize(ctx); err != nil {
		logger.Warn("Failed to initialize registry", zap.Error(err))
	}

	// Handlers
	registryHandler := handler.NewRegistryHandler(registryService, logger)
	syncHandler := handler.NewSyncHandler(syncService, logger)

	// HTTP Server
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// Routes
	api := r.Group("/api/v1")
	{
		// Token Registry
		tokens := api.Group("/tokens")
		{
			tokens.GET("", registryHandler.ListTokens)
			tokens.POST("", registryHandler.CreateToken)
			tokens.GET("/:canonical", registryHandler.GetToken)
			tokens.PUT("/:canonical", registryHandler.UpdateToken)
			tokens.DELETE("/:canonical", registryHandler.DeleteToken)
			tokens.POST("/:canonical/verify", registryHandler.VerifyArkham)
			tokens.POST("/:canonical/toggle", registryHandler.ToggleActive)
		}

		api.GET("/exchanges/:exchange/symbols", registryHandler.GetExchangeSymbols)
		api.GET("/normalize", registryHandler.Normalize)
		api.GET("/stats", registryHandler.GetStats)

		// Symbol Generation (Preview)
		api.POST("/generate", syncHandler.GeneratePreview)
		api.POST("/generate/bulk", syncHandler.BulkGeneratePreview)

		// Dynamic Sync (Insert to DB)
		api.POST("/sync/binance", syncHandler.SyncBinance)
		api.POST("/sync/all", syncHandler.SyncAll)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		logger.Info("Token service starting", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
