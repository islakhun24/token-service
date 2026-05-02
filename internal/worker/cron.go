package worker

import (
	"context"
	"log"

	"github.com/robfig/cron/v3"

	"futures-symbol-module/internal/repository"
	"futures-symbol-module/internal/symbol"
)

// CronWorker manages scheduled tasks.
type CronWorker struct {
	scheduler *cron.Cron
	svc       *symbol.Service
	repo      *repository.PairRepository
	schedule  string
}

// NewCronWorker creates a new CronWorker.
func NewCronWorker(svc *symbol.Service, repo *repository.PairRepository, schedule string) *CronWorker {
	return &CronWorker{
		scheduler: cron.New(),
		svc:       svc,
		repo:      repo,
		schedule:  schedule,
	}
}

// Start begins the cron scheduler and adds the sync job.
func (w *CronWorker) Start() {
	_, err := w.scheduler.AddFunc(w.schedule, func() {
		w.RunJob()
	})
	if err != nil {
		log.Fatalf("Failed to schedule cron job with schedule '%s': %v", w.schedule, err)
	}

	w.scheduler.Start()
	log.Printf("Cron scheduler started: pairs will be synced according to schedule '%s'\n", w.schedule)

	// Run the job immediately on startup asynchronously
	go w.RunJob()
}

// Stop gracefully stops the cron scheduler.
func (w *CronWorker) Stop() {
	ctx := w.scheduler.Stop()
	<-ctx.Done()
	log.Println("Cron scheduler stopped")
}

// RunJob executes the process of fetching pairs and storing them to DB.
func (w *CronWorker) RunJob() {
	log.Println("Starting scheduled pairs sync job...")

	ctx := context.Background()
	resp, err := w.svc.GetFuturesPairs(ctx)
	if err != nil {
		log.Printf("Cron job error getting pairs: %v", err)
		return
	}

	var pairs []repository.Pair
	for _, entry := range resp.Data {
		p := repository.Pair{
			Base:      entry.Base,
			Quote:     entry.Quote,
			Type:      entry.Type,
			MarketCap: entry.Market.MarketCap,
		}

		if entry.CoinGeckoID != "" {
			p.CoinGeckoID = ptr(entry.CoinGeckoID)
		}

		// Check each exchange from the map
		if val, ok := entry.Exchanges["binance"]; ok && val != "" {
			p.Binance = ptr(val)
		}
		if val, ok := entry.Exchanges["bybit"]; ok && val != "" {
			p.Bybit = ptr(val)
		}
		if val, ok := entry.Exchanges["okx"]; ok && val != "" {
			p.OKX = ptr(val)
		}
		if val, ok := entry.Exchanges["mexc"]; ok && val != "" {
			p.Mexc = ptr(val)
		}
		if val, ok := entry.Exchanges["bitget"]; ok && val != "" {
			p.Bitget = ptr(val)
		}
		if val, ok := entry.Exchanges["kucoin"]; ok && val != "" {
			p.Kucoin = ptr(val)
		}

		pairs = append(pairs, p)
	}

	if err := w.repo.BulkUpsert(ctx, pairs); err != nil {
		log.Printf("Cron job error upserting pairs: %v", err)
		return
	}

	log.Printf("Cron job successfully synced %d pairs to database", len(pairs))
}

func ptr(s string) *string {
	return &s
}
