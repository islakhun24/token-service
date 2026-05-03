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

// Service orchestrates symbol collection, normalization, matching, and enrichment.
type Service struct {
	collectors []Collector
	validator  *Validator
	matcher    *Matcher
	builder    *RegistryBuilder
	ranker     *Ranker
}

// NewService creates a new symbol service.
func NewService(collectors []Collector) *Service {
	return &Service{
		collectors: collectors,
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

			items, err := c.FetchSymbols(ctx)
			if err != nil {
				fmt.Printf("Warning: failed to fetch from %s: %v\n", c.Name(), err)
				return nil
			}

			var parsed []ExchangeSymbol

			for _, item := range items {

				canon, ok := Normalize(item)
				if !ok {
					continue
				}

				if !s.validator.Validate(canon) {
					continue
				}

				parsed = append(parsed, ExchangeSymbol{
					Exchange:   c.Name(),
					Raw:        item.Raw,
					Canonical:  canon,
					Categories: item.Categories,
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

	// ranking (by exchange count since market cap is enriched separately)
	ranked := s.ranker.Rank(entries)

	return &FuturesPairsResponse{
		Version: time.Now().UTC().Format(time.RFC3339),
		Total:   len(ranked),
		Data:    ranked,
	}, nil
}

