package main

import (
	"context"
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
	cfg := config.Load()

	// Initialize CoinGecko client and worker
	cgClient := asset.NewCoinGeckoClient(cfg)
	cgWorker := worker.NewCoinGeckoWorker(cgClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cgWorker.Start(ctx)
	defer cgWorker.Stop()

	// Give CoinGecko a moment to populate cache
	time.Sleep(2 * time.Second)

	// Initialize collectors
	collectors := []symbol.Collector{
		collector.NewBinanceCollector(),
		collector.NewOKXCollector(),
		collector.NewBybitCollector(),
		collector.NewBitgetCollector(),
		collector.NewMexcCollector(),
		collector.NewKucoinCollector(),
	}

	// Initialize symbol service
	svc := symbol.NewService(collectors, cgClient)

	// Initialize database
	db, err := repository.InitDB(cfg.DatabaseDSN())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	repo := repository.NewPairRepository(db)

	// Initialize cron worker
	cronWorker := worker.NewCronWorker(svc, repo, cfg.CronSchedule)
	cronWorker.Start()
	defer cronWorker.Stop()

	// Initialize handlers
	futuresHandler := handler.NewFuturesHandler(svc)
	dbPairsHandler := handler.NewDBPairsHandler(repo)

	// Setup router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/futures/pairs", futuresHandler.GetPairs)
	r.GET("/api/db/pairs", dbPairsHandler.GetPairs)

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

	log.Printf("🚀 Futures Symbol Module running on :%s", cfg.Port)
	log.Printf("📡 Endpoint: GET http://localhost:%s/futures/pairs", cfg.Port)
	log.Printf("📡 Endpoint: GET http://localhost:%s/api/db/pairs", cfg.Port)

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
