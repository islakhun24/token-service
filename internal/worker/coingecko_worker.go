package worker

import (
	"context"
	"log"
	"time"

	"futures-symbol-module/internal/asset"
)

// CoinGeckoWorker periodically refreshes CoinGecko caches.
type CoinGeckoWorker struct {
	client *asset.CoinGeckoClient
	stop   chan struct{}
}

// NewCoinGeckoWorker creates a new worker.
func NewCoinGeckoWorker(client *asset.CoinGeckoClient) *CoinGeckoWorker {
	return &CoinGeckoWorker{
		client: client,
		stop:   make(chan struct{}),
	}
}

// Start begins the background refresh loops.
func (w *CoinGeckoWorker) Start(ctx context.Context) {
	// Initial load
	if err := w.client.RefreshList(ctx); err != nil {
		log.Printf("[CoinGeckoWorker] initial list refresh failed: %v", err)
	}
	if err := w.client.RefreshMarkets(ctx); err != nil {
		log.Printf("[CoinGeckoWorker] initial markets refresh failed: %v", err)
	}

	// List refresher (every 6 hours)
	go w.loop(ctx, 6*time.Hour, w.client.RefreshList, "list")

	// Markets refresher (every 60 seconds)
	go w.loop(ctx, 60*time.Second, w.client.RefreshMarkets, "markets")
}

// Stop signals the worker to stop.
func (w *CoinGeckoWorker) Stop() {
	close(w.stop)
}

func (w *CoinGeckoWorker) loop(ctx context.Context, interval time.Duration, fn func(context.Context) error, name string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := fn(ctx); err != nil {
				log.Printf("[CoinGeckoWorker] %s refresh failed: %v", name, err)
			} else {
				log.Printf("[CoinGeckoWorker] %s refreshed", name)
			}
		case <-w.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}
