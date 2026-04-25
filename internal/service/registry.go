package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"token-service/internal/domain"
	"token-service/internal/infrastructure/binance"
	"token-service/internal/infrastructure/staticmapping"
	"token-service/pkg/normalizer"

	"go.uber.org/zap"
)

type RegistryRepository interface {
	GetAll(ctx context.Context) ([]domain.CanonicalToken, error)
	GetByCanonical(ctx context.Context, symbol string) (*domain.CanonicalToken, error)
	Create(ctx context.Context, token *domain.CanonicalToken) error
	Update(ctx context.Context, token *domain.CanonicalToken) error
	Delete(ctx context.Context, canonical string) error
	GetExchangeMaps(ctx context.Context, canonicalID int) ([]domain.ExchangeMapping, error)
	SaveExchangeMap(ctx context.Context, m *domain.ExchangeMapping) error
	DeleteExchangeMaps(ctx context.Context, canonicalID int) error
	GetContracts(ctx context.Context, canonicalID int) (map[string]string, error)
	SaveContract(ctx context.Context, canonicalID int, chain, address string) error
	DeleteContracts(ctx context.Context, canonicalID int) error
	GetStats(ctx context.Context) (*domain.TokenStats, error)
	ToggleActive(ctx context.Context, canonical string, active bool) error
}

type RegistryService struct {
	repo          RegistryRepository
	normalizer    *normalizer.SymbolNormalizer
	staticLoader  *staticmapping.Loader
	binanceClient *binance.Client
	logger        *zap.Logger
	mu            sync.RWMutex
}

func NewRegistryService(
	repo RegistryRepository,
	norm *normalizer.SymbolNormalizer,
	loader *staticmapping.Loader,
	binance *binance.Client,
	logger *zap.Logger,
) *RegistryService {
	return &RegistryService{
		repo:          repo,
		normalizer:    norm,
		staticLoader:  loader,
		binanceClient: binance,
		logger:        logger,
	}
}

// Initialize load data dari static mapping ke DB kalau DB kosong
func (s *RegistryService) Initialize(ctx context.Context) error {
	existing, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	if len(existing) > 0 {
		s.logger.Info("Registry already initialized", zap.Int("count", len(existing)))
		return nil
	}

	if s.staticLoader == nil {
		s.logger.Warn("No static loader, skipping initialization")
		return nil
	}

	tokens := s.staticLoader.GetAll()
	s.logger.Info("Initializing from static mapping", zap.Int("count", len(tokens)))

	for _, t := range tokens {
		canonical := &domain.CanonicalToken{
			CanonicalSymbol: strings.ToUpper(t.Symbol),
			CoinGeckoID:     t.CoinGeckoID,
			Name:            t.Name,
			Category:        "unknown",
			IsActive:        true,
			ArkhamReady:     true,
		}

		if err := s.repo.Create(ctx, canonical); err != nil {
			s.logger.Warn("Failed to create token", zap.String("symbol", t.Symbol), zap.Error(err))
			continue
		}

		// Save contracts
		for chain, addr := range t.Contracts {
			if err := s.repo.SaveContract(ctx, canonical.ID, chain, addr); err != nil {
				s.logger.Warn("Failed to save contract", zap.Error(err))
			}
		}
	}

	return nil
}

func (s *RegistryService) ListTokens(ctx context.Context, search string) ([]domain.CanonicalToken, error) {
	tokens, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var result []domain.CanonicalToken
	search = strings.ToLower(search)

	for _, t := range tokens {
		if search != "" &&
			!strings.Contains(strings.ToLower(t.CanonicalSymbol), search) &&
			!strings.Contains(strings.ToLower(t.Name), search) &&
			!strings.Contains(strings.ToLower(t.CoinGeckoID), search) {
			continue
		}

		// Load exchange maps
		maps, _ := s.repo.GetExchangeMaps(ctx, t.ID)
		t.ExchangeMaps = maps

		// Load contracts
		contracts, _ := s.repo.GetContracts(ctx, t.ID)
		t.Contracts = contracts

		result = append(result, t)
	}

	return result, nil
}

func (s *RegistryService) GetToken(ctx context.Context, canonical string) (*domain.CanonicalToken, error) {
	token, err := s.repo.GetByCanonical(ctx, strings.ToUpper(canonical))
	if err != nil {
		return nil, err
	}

	maps, _ := s.repo.GetExchangeMaps(ctx, token.ID)
	token.ExchangeMaps = maps

	contracts, _ := s.repo.GetContracts(ctx, token.ID)
	token.Contracts = contracts

	return token, nil
}

func (s *RegistryService) CreateToken(ctx context.Context, req domain.CreateTokenRequest) (*domain.CanonicalToken, error) {
	canonical := strings.ToUpper(req.Canonical)

	// Check existing
	existing, _ := s.repo.GetByCanonical(ctx, canonical)
	if existing != nil {
		return nil, fmt.Errorf("token %s already exists", canonical)
	}

	token := &domain.CanonicalToken{
		CanonicalSymbol: canonical,
		Name:            req.Name,
		CoinGeckoID:     strings.ToLower(req.CoinGeckoID),
		Category:        req.Category,
		IsActive:        true,
		ArkhamReady:     false,
	}

	if err := s.repo.Create(ctx, token); err != nil {
		return nil, err
	}

	// Save exchange mappings
	mappings := s.buildExchangeMappings(token.ID, req)
	for _, m := range mappings {
		if err := s.repo.SaveExchangeMap(ctx, &m); err != nil {
			s.logger.Warn("Failed to save exchange map", zap.Error(err))
		}
	}

	// Save contracts
	for chain, addr := range req.Contracts {
		if err := s.repo.SaveContract(ctx, token.ID, chain, addr); err != nil {
			s.logger.Warn("Failed to save contract", zap.Error(err))
		}
	}

	return s.GetToken(ctx, canonical)
}

func (s *RegistryService) UpdateToken(ctx context.Context, canonical string, req domain.CreateTokenRequest) (*domain.CanonicalToken, error) {
	canonical = strings.ToUpper(canonical)

	token, err := s.repo.GetByCanonical(ctx, canonical)
	if err != nil {
		return nil, fmt.Errorf("token %s not found", canonical)
	}

	token.Name = req.Name
	token.CoinGeckoID = strings.ToLower(req.CoinGeckoID)
	token.Category = req.Category

	if err := s.repo.Update(ctx, token); err != nil {
		return nil, err
	}

	// Replace exchange mappings
	if err := s.repo.DeleteExchangeMaps(ctx, token.ID); err != nil {
		s.logger.Warn("Failed to delete old maps", zap.Error(err))
	}
	mappings := s.buildExchangeMappings(token.ID, req)
	for _, m := range mappings {
		if err := s.repo.SaveExchangeMap(ctx, &m); err != nil {
			s.logger.Warn("Failed to save exchange map", zap.Error(err))
		}
	}

	// Replace contracts
	if err := s.repo.DeleteContracts(ctx, token.ID); err != nil {
		s.logger.Warn("Failed to delete old contracts", zap.Error(err))
	}
	for chain, addr := range req.Contracts {
		if err := s.repo.SaveContract(ctx, token.ID, chain, addr); err != nil {
			s.logger.Warn("Failed to save contract", zap.Error(err))
		}
	}

	return s.GetToken(ctx, canonical)
}

func (s *RegistryService) DeleteToken(ctx context.Context, canonical string) error {
	return s.repo.Delete(ctx, strings.ToUpper(canonical))
}

func (s *RegistryService) ToggleActive(ctx context.Context, canonical string) (*domain.CanonicalToken, error) {
	token, err := s.repo.GetByCanonical(ctx, strings.ToUpper(canonical))
	if err != nil {
		return nil, err
	}

	if err := s.repo.ToggleActive(ctx, canonical, !token.IsActive); err != nil {
		return nil, err
	}

	return s.GetToken(ctx, canonical)
}

func (s *RegistryService) VerifyArkham(ctx context.Context, canonical string) (*domain.CanonicalToken, error) {
	token, err := s.repo.GetByCanonical(ctx, strings.ToUpper(canonical))
	if err != nil {
		return nil, err
	}

	// TODO: Implement actual Arkham API verification
	// For now, simulate
	token.ArkhamReady = true
	if err := s.repo.Update(ctx, token); err != nil {
		return nil, err
	}

	return s.GetToken(ctx, canonical)
}

func (s *RegistryService) GetExchangeSymbols(ctx context.Context, canonical, exchange string) ([]domain.ExchangeSymbol, error) {
	token, err := s.repo.GetByCanonical(ctx, strings.ToUpper(canonical))
	if err != nil {
		return nil, err
	}

	maps, err := s.repo.GetExchangeMaps(ctx, token.ID)
	if err != nil {
		return nil, err
	}

	var result []domain.ExchangeSymbol
	for _, m := range maps {
		if exchange != "" && !strings.EqualFold(m.Exchange, exchange) {
			continue
		}
		result = append(result, domain.ExchangeSymbol{
			Exchange: m.Exchange,
			Symbol:   m.ExchangeSymbol,
			Base:     m.NormalizedBase,
			Quote:    m.QuoteAsset,
		})
	}

	return result, nil
}

func (s *RegistryService) Normalize(ctx context.Context, exchange, symbol string) (*domain.NormalizeResponse, error) {
	canonical := s.normalizer.ExtractBase(exchange, symbol)

	return &domain.NormalizeResponse{
		Canonical: canonical,
		BaseAsset: canonical,
		Exchange:  exchange,
	}, nil
}

func (s *RegistryService) GetStats(ctx context.Context) (*domain.TokenStats, error) {
	return s.repo.GetStats(ctx)
}

func (s *RegistryService) buildExchangeMappings(canonicalID int, req domain.CreateTokenRequest) []domain.ExchangeMapping {
	var result []domain.ExchangeMapping
	canonical := strings.ToUpper(req.Canonical)

	if req.Binance != "" {
		result = append(result, domain.ExchangeMapping{
			CanonicalID:    canonicalID,
			Exchange:       "binance",
			MarketType:     "perpetual",
			ExchangeSymbol: req.Binance,
			BaseAsset:      s.normalizer.ExtractBase("binance", req.Binance),
			QuoteAsset:     "USDT",
			NormalizedBase: canonical,
			IsActive:       true,
		})
	}

	if req.KuCoin != "" {
		result = append(result, domain.ExchangeMapping{
			CanonicalID:    canonicalID,
			Exchange:       "kucoin",
			MarketType:     "perpetual",
			ExchangeSymbol: req.KuCoin,
			BaseAsset:      s.normalizer.ExtractBase("kucoin", req.KuCoin),
			QuoteAsset:     "USDT",
			NormalizedBase: canonical,
			IsActive:       true,
		})
	}

	if req.Bybit != "" {
		result = append(result, domain.ExchangeMapping{
			CanonicalID:    canonicalID,
			Exchange:       "bybit",
			MarketType:     "linear",
			ExchangeSymbol: req.Bybit,
			BaseAsset:      s.normalizer.ExtractBase("bybit", req.Bybit),
			QuoteAsset:     "USDT",
			NormalizedBase: canonical,
			IsActive:       true,
		})
	}

	if req.OKX != "" {
		result = append(result, domain.ExchangeMapping{
			CanonicalID:    canonicalID,
			Exchange:       "okx",
			MarketType:     "perpetual",
			ExchangeSymbol: req.OKX,
			BaseAsset:      s.normalizer.ExtractBase("okx", req.OKX),
			QuoteAsset:     "USDT",
			NormalizedBase: canonical,
			IsActive:       true,
		})
	}

	return result
}
