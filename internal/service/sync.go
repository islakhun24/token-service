package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"token-service/internal/domain"
	"token-service/internal/infrastructure/binance"
	"token-service/pkg/normalizer"
	"token-service/pkg/symbolgen"

	"go.uber.org/zap"
)

type SyncRepository interface {
	GetByCanonical(ctx context.Context, symbol string) (*domain.CanonicalToken, error)
	Create(ctx context.Context, token *domain.CanonicalToken) error
	Update(ctx context.Context, token *domain.CanonicalToken) error
	SaveExchangeMap(ctx context.Context, m *domain.ExchangeMapping) error
}

type SyncService struct {
	repo       SyncRepository
	normalizer *normalizer.SymbolNormalizer
	gen        *symbolgen.Generator
	binance    *binance.Client
	logger     *zap.Logger
}

func NewSyncService(
	repo SyncRepository,
	norm *normalizer.SymbolNormalizer,
	binanceClient *binance.Client,
	logger *zap.Logger,
) *SyncService {
	return &SyncService{
		repo:       repo,
		normalizer: norm,
		gen:        symbolgen.NewGenerator(),
		binance:    binanceClient,
		logger:     logger,
	}
}

// ==================== GENERATE (Preview Only) ====================
func (s *SyncService) Generate(ctx context.Context, canonical string) (*domain.GenerateResponse, error) {
	canonical = strings.ToUpper(canonical)
	syms := s.gen.GenerateAll(canonical)

	return &domain.GenerateResponse{
		Canonical: syms.Canonical,
		Binance:   syms.Binance,
		KuCoin:    syms.KuCoin,
		Bybit:     syms.Bybit,
		OKX:       syms.OKX,
		Arkham:    syms.Arkham,
		CoinGecko: syms.CoinGecko,
	}, nil
}

func (s *SyncService) BulkGenerate(ctx context.Context, binanceSymbols []string) (*domain.BulkGenerateResponse, error) {
	var items []domain.GenerateResponse
	seen := make(map[string]bool)

	for _, sym := range binanceSymbols {
		canonical := s.normalizer.ExtractBase("binance", sym)
		if canonical == "" || seen[canonical] {
			continue
		}
		seen[canonical] = true

		resp, _ := s.Generate(ctx, canonical)
		if resp != nil {
			items = append(items, *resp)
		}
	}

	return &domain.BulkGenerateResponse{
		Items: items,
		Total: len(items),
	}, nil
}

// ==================== SYNC BINANCE (Auto-generate all platforms) ====================
func (s *SyncService) SyncBinance(ctx context.Context) (*domain.SyncResult, error) {
	start := time.Now()
	result := &domain.SyncResult{Exchange: "binance"}

	info, err := s.binance.GetFuturesExchangeInfo()
	if err != nil {
		return nil, fmt.Errorf("fetch binance: %w", err)
	}

	for _, sym := range info.Symbols {
		if sym.Status != "TRADING" || sym.ContractType != "PERPETUAL" {
			continue
		}
		if !strings.HasSuffix(sym.Symbol, "USDT") {
			continue
		}

		canonical := s.normalizer.ExtractBase("binance", sym.Symbol)
		if canonical == "" {
			continue
		}

		result.Total++

		// Generate symbols for ALL platforms
		platformSyms := s.gen.GenerateAll(canonical)

		// Upsert canonical token
		token, err := s.upsertToken(ctx, canonical, platformSyms.Arkham)
		if err != nil {
			result.Failed++
			s.logger.Warn("Failed to upsert token", zap.String("symbol", canonical), zap.Error(err))
			continue
		}

		// Upsert exchange mappings for ALL platforms
		mappings := []struct {
			exchange   string
			marketType string
			symbol     string
			baseAsset  string
			quoteAsset string
		}{
			{"binance", "perpetual", platformSyms.Binance, sym.BaseAsset, sym.QuoteAsset},
			{"kucoin", "perpetual", platformSyms.KuCoin, s.extractBase(platformSyms.KuCoin), "USDT"},
			{"bybit", "linear", platformSyms.Bybit, canonical, "USDT"},
			{"okx", "perpetual", platformSyms.OKX, canonical, "USDT"},
		}

		for _, m := range mappings {
			if err := s.upsertExchangeMap(ctx, token.ID, m.exchange, m.marketType, m.symbol, m.baseAsset, m.quoteAsset, canonical); err != nil {
				s.logger.Warn("Failed to save exchange map", zap.String("exchange", m.exchange), zap.Error(err))
			}
		}

		if token.CreatedAt.Equal(token.UpdatedAt) || token.CreatedAt.After(token.UpdatedAt.Add(-1*time.Second)) {
			result.Inserted++
		} else {
			result.Updated++
		}
	}

	result.DurationMs = time.Since(start).Milliseconds()
	s.logger.Info("Binance sync completed",
		zap.Int("total", result.Total),
		zap.Int("inserted", result.Inserted),
		zap.Int("updated", result.Updated),
		zap.Int("failed", result.Failed),
	)

	return result, nil
}

// ==================== SYNC ALL (Same as SyncBinance but with all platforms generated) ====================
func (s *SyncService) SyncAll(ctx context.Context) (*domain.SyncAllResponse, error) {
	// Single sync from Binance generates all platform mappings
	result, err := s.SyncBinance(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.SyncAllResponse{
		Results: []domain.SyncResult{*result},
	}, nil
}

// ==================== HELPERS ====================
func (s *SyncService) upsertToken(ctx context.Context, canonical, coingeckoID string) (*domain.CanonicalToken, error) {
	canonical = strings.ToUpper(canonical)

	existing, err := s.repo.GetByCanonical(ctx, canonical)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		existing.CoinGeckoID = coingeckoID
		existing.UpdatedAt = time.Now()
		if err := s.repo.Update(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	token := &domain.CanonicalToken{
		CanonicalSymbol: canonical,
		Name:            canonical,
		CoinGeckoID:     coingeckoID,
		Category:        "unknown",
		IsActive:        true,
		ArkhamReady:     false,
	}

	if err := s.repo.Create(ctx, token); err != nil {
		return nil, err
	}
	return token, nil
}

func (s *SyncService) upsertExchangeMap(
	ctx context.Context,
	canonicalID int,
	exchange, marketType, symbol, baseAsset, quoteAsset, normalized string,
) error {
	m := &domain.ExchangeMapping{
		CanonicalID:    canonicalID,
		Exchange:       exchange,
		MarketType:     marketType,
		ExchangeSymbol: symbol,
		BaseAsset:      baseAsset,
		QuoteAsset:     quoteAsset,
		NormalizedBase: normalized,
		IsActive:       true,
	}
	return s.repo.SaveExchangeMap(ctx, m)
}

func (s *SyncService) extractBase(symbol string) string {
	symbol = strings.ToUpper(symbol)
	symbol = strings.TrimSuffix(symbol, "USDTM")
	symbol = strings.TrimSuffix(symbol, "USDT")
	return symbol
}
