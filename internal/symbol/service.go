package symbol

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// Collector defines the interface for exchange symbol collectors.
type Collector interface {
	Name() string
	FetchSymbols(ctx context.Context) ([]SymbolInput, error)
}

// CoinGeckoResolver resolves CoinGecko IDs and market data.
type CoinGeckoResolver interface {
	ResolveID(base string) string
	GetMarketCap(id string) float64
}

// Service orchestrates symbol collection, normalization, matching, and enrichment.
type Service struct {
	collectors []Collector
	resolver   CoinGeckoResolver
	validator  *Validator
	matcher    *Matcher
	builder    *RegistryBuilder
	ranker     *Ranker
}

// NewService creates a new symbol service.
func NewService(collectors []Collector, resolver CoinGeckoResolver) *Service {
	return &Service{
		collectors: collectors,
		resolver:   resolver,
		validator:  NewValidator(),
		matcher:    NewMatcher(),
		builder:    NewRegistryBuilder(),
		ranker:     NewRanker(),
	}
}

// GetFuturesPairs builds the complete futures pairs response.
func (s *Service) GetFuturesPairs(ctx context.Context) (*FuturesPairsResponse, error) {

	allSymbols := make([]ExchangeSymbol, 0)
	var mu sync.Mutex

	g, ctx := errgroup.WithContext(ctx)

	for _, col := range s.collectors {
		c := col

		g.Go(func() error {

			// 🔥 IMPORTANT: sekarang ambil structured, bukan []string
			items, err := c.FetchSymbols(ctx)
			if err != nil {
				fmt.Printf("Warning: failed to fetch from %s: %v\n", c.Name(), err)
				return nil
			}

			var parsed []ExchangeSymbol

			for _, item := range items {

				// 🔥 GANTI INI (CORE FIX)
				canon, ok := Normalize(item)
				if !ok {
					continue
				}

				if !s.validator.Validate(canon) {
					continue
				}

				parsed = append(parsed, ExchangeSymbol{
					Exchange:  c.Name(),
					Raw:       item.Raw,
					Canonical: canon,
				})
			}

			mu.Lock()
			allSymbols = append(allSymbols, parsed...)
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// matcher
	groups := s.matcher.Match(allSymbols)

	// registry
	entries := s.builder.Build(groups)

	// enrich
	for i := range entries {
		entries[i].CoinGeckoID = s.resolver.ResolveID(entries[i].Base)
		entries[i].Market.MarketCap = s.resolver.GetMarketCap(entries[i].CoinGeckoID)
	}

	// ranking
	ranked := s.ranker.Rank(entries)

	return &FuturesPairsResponse{
		Version: time.Now().UTC().Format(time.RFC3339),
		Total:   len(ranked),
		Data:    ranked,
	}, nil
}
