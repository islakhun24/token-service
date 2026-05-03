package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"futures-symbol-module/internal/asset"
	"futures-symbol-module/internal/collector"
	"futures-symbol-module/internal/config"
	"futures-symbol-module/internal/handler"
	"futures-symbol-module/internal/repository"
	"futures-symbol-module/internal/symbol"
	"futures-symbol-module/internal/worker"
)

func main() {
	migrateFlag := flag.String("m", "", "Run migration and exit")
	flag.Parse()

	cfg := config.Load()

	// Initialize database (AutoMigrate happens here)
	db, err := repository.InitDB(cfg.DatabaseDSN())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// If ran with -m=all, exit cleanly after DB initialization (migration)
	if *migrateFlag == "all" {
		log.Println("Database migration completed successfully. Exiting.")
		return
	}

	repo := repository.NewPairRepository(db)

	// Initialize collectors (exchange APIs)
	collectors := []symbol.Collector{
		collector.NewBinanceCollector(),
		collector.NewOKXCollector(),
		collector.NewBybitCollector(),
		collector.NewBitgetCollector(),
		collector.NewMexcCollector(),
		collector.NewKucoinCollector(),
	}

	// Initialize symbol service (exchange pair collection only)
	svc := symbol.NewService(collectors)

	// ===== Worker 1: Exchange Pair Sync =====
	cronWorker := worker.NewCronWorker(svc, repo, cfg.CronSchedule)
	cronWorker.Start()
	defer cronWorker.Stop()

	// ===== Worker 2: CoinMarketCap Enrichment =====
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmcClient := asset.NewCoinMarketCapClient()
	cmcWorker := worker.NewCMCWorker(cmcClient, repo)
	cmcWorker.Start(ctx)
	defer cmcWorker.Stop()

	// Initialize handlers
	futuresHandler := handler.NewFuturesHandler(svc)
	dbPairsHandler := handler.NewDBPairsHandler(repo)

	// Setup router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/futures/pairs", futuresHandler.GetPairs)
	r.GET("/api/db/pairs", dbPairsHandler.GetPairs)
	r.GET("/api/categories", dbPairsHandler.GetCategories)

	// Health check
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "token-service",
			"status":  "active",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	log.Printf("🚀 Token Service running on :%s", cfg.Port)
	log.Printf("📡 Endpoint: GET http://localhost:%s/futures/pairs", cfg.Port)
	log.Printf("📡 Endpoint: GET http://localhost:%s/api/db/pairs", cfg.Port)
	log.Printf("📡 Endpoint: GET http://localhost:%s/api/categories", cfg.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
