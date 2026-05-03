package worker

import (
	"context"
	"log"
	"time"

	"futures-symbol-module/internal/asset"
	"futures-symbol-module/internal/repository"
)

// CMCWorker handles CoinMarketCap data enrichment.
// It fetches market data and enriches pairs in DB with CMC ID, market cap, supply data.
type CMCWorker struct {
	client   *asset.CoinMarketCapClient
	repo     *repository.PairRepository
	stop     chan struct{}
	schedule time.Duration
	maxCoins int
}

// NewCMCWorker creates a new CoinMarketCap worker.
func NewCMCWorker(client *asset.CoinMarketCapClient, repo *repository.PairRepository) *CMCWorker {
	return &CMCWorker{
		client:   client,
		repo:     repo,
		stop:     make(chan struct{}),
		schedule: 5 * time.Minute,
		maxCoins: 2000,
	}
}

// Start begins the CMC enrichment loop.
func (w *CMCWorker) Start(ctx context.Context) {
	go w.loop(ctx)
}

// Stop signals the worker to stop.
func (w *CMCWorker) Stop() {
	close(w.stop)
}

// loop runs CMC enrichment on a schedule.
func (w *CMCWorker) loop(ctx context.Context) {
	// Run immediately on startup
	w.runEnrichment(ctx)

	ticker := time.NewTicker(w.schedule)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.runEnrichment(ctx)
		case <-w.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}

// runEnrichment fetches CMC data and enriches pairs in DB.
func (w *CMCWorker) runEnrichment(ctx context.Context) {
	log.Println("[CMCWorker] Starting enrichment cycle...")

	// Step 1: Fetch market data from CoinMarketCap
	if err := w.client.RefreshData(ctx, w.maxCoins); err != nil {
		log.Printf("[CMCWorker] Failed to refresh CMC data: %v", err)
		return
	}

	// Step 2: Enrich pairs in DB
	w.enrichPairsInDB(ctx)

	log.Println("[CMCWorker] Enrichment cycle complete")
}

// enrichPairsInDB updates all pairs with CMC data (ID, market cap, supply, etc).
func (w *CMCWorker) enrichPairsInDB(ctx context.Context) {
	pairs, err := w.repo.GetAll(ctx)
	if err != nil {
		log.Printf("[CMCWorker] Failed to fetch pairs: %v", err)
		return
	}

	updated := 0
	for i := range pairs {
		coin, ok := w.client.GetCoin(pairs[i].Base)
		if !ok {
			continue
		}

		mcap := w.client.GetMarketCapUSD(coin)

		// Update CMC fields
		pairs[i].CMCID = &coin.ID
		pairs[i].CMCName = &coin.Name
		pairs[i].CMCSlug = &coin.Slug
		if coin.CMCRank > 0 {
			pairs[i].CMCRank = &coin.CMCRank
		}
		if mcap > 0 {
			pairs[i].MarketCap = mcap
		}
		pairs[i].CirculatingSupply = coin.CirculatingSupply
		pairs[i].TotalSupply = coin.TotalSupply
		pairs[i].MaxSupply = coin.MaxSupply

		updated++
	}

	if updated > 0 {
		if err := w.repo.BulkUpsert(ctx, pairs); err != nil {
			log.Printf("[CMCWorker] Failed to update pairs: %v", err)
			return
		}
	}

	log.Printf("[CMCWorker] Enriched %d/%d pairs with CMC data", updated, len(pairs))
}
