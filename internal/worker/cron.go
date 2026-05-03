package worker

import (
	"context"
	"log"

	"github.com/robfig/cron/v3"

	"futures-symbol-module/internal/repository"
	"futures-symbol-module/internal/symbol"
)

// CronWorker manages scheduled exchange pair sync tasks.
// It only fetches from exchanges (Binance, OKX, Bybit, etc.) — no CMC.
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

// RunJob fetches pairs from all exchanges and stores them to DB.
// CMC enrichment (market cap, categories) is handled separately by CMCWorker.
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
			Base:  entry.Base,
			Quote: entry.Quote,
			Type:  entry.Type,
		}

		// Preserve CMC data if already enriched (don't overwrite with empty)
		if entry.CMCID != nil {
			p.CMCID = entry.CMCID
		}
		if entry.CMCName != nil {
			p.CMCName = entry.CMCName
		}
		if entry.CMCSlug != nil {
			p.CMCSlug = entry.CMCSlug
		}
		if entry.CMCRank != nil {
			p.CMCRank = entry.CMCRank
		}
		if entry.Market.MarketCap > 0 {
			p.MarketCap = entry.Market.MarketCap
		}
		if entry.CirculatingSupply > 0 {
			p.CirculatingSupply = entry.CirculatingSupply
		}
		if entry.TotalSupply > 0 {
			p.TotalSupply = entry.TotalSupply
		}
		if entry.MaxSupply > 0 {
			p.MaxSupply = entry.MaxSupply
		}
		if entry.Market.MarketCap > 0 {
			p.MarketCap = entry.Market.MarketCap
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

	// Sync categories from exchange data (e.g. Binance underlyingSubType)
	catCount := 0
	for _, entry := range resp.Data {
		if len(entry.Categories) > 0 {
			if err := w.repo.ReplaceCategoriesForBase(ctx, entry.Base, entry.Categories); err != nil {
				log.Printf("Failed to sync categories for %s: %v", entry.Base, err)
				continue
			}
			catCount++
		}
	}
	if catCount > 0 {
		log.Printf("Synced categories for %d pairs from exchange data", catCount)
	}
}

func ptr(s string) *string {
	return &s
}
