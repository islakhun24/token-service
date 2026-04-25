package asset

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"futures-symbol-module/internal/config"
	"futures-symbol-module/internal/symbol"
	"futures-symbol-module/pkg/httpclient"
)

// CoinGeckoClient provides cached CoinGecko data with duplicate-safe resolution.
type CoinGeckoClient struct {
	client     *httpclient.Client
	config     *config.Config
	mu         sync.RWMutex
	coins      []symbol.CoinGeckoCoin
	markets    map[string]symbol.CoinGeckoMarket // key: coingecko_id
	lastList   time.Time
	lastMarkets time.Time
}

// NewCoinGeckoClient creates a new CoinGecko client.
func NewCoinGeckoClient(cfg *config.Config) *CoinGeckoClient {
	return &CoinGeckoClient{
		client:  httpclient.New(),
		config:  cfg,
		markets: make(map[string]symbol.CoinGeckoMarket),
	}
}

// RefreshList fetches and caches /coins/list.
func (c *CoinGeckoClient) RefreshList(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.coingecko.com/api/v3/coins/list", nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("coingecko list status: %d", resp.StatusCode)
	}

	var coins []symbol.CoinGeckoCoin
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return err
	}

	c.mu.Lock()
	c.coins = coins
	c.lastList = time.Now()
	c.mu.Unlock()
	return nil
}

// RefreshMarkets fetches and caches /coins/markets.
func (c *CoinGeckoClient) RefreshMarkets(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=250&page=1&sparkline=false",
		nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("coingecko markets status: %d", resp.StatusCode)
	}

	var markets []symbol.CoinGeckoMarket
	if err := json.NewDecoder(resp.Body).Decode(&markets); err != nil {
		return err
	}

	c.mu.Lock()
	c.markets = make(map[string]symbol.CoinGeckoMarket, len(markets))
	for _, m := range markets {
		c.markets[strings.ToLower(m.ID)] = m
	}
	c.lastMarkets = time.Now()
	c.mu.Unlock()
	return nil
}

// ResolveID resolves a base symbol to a CoinGecko ID using duplicate-safe scoring.
func (c *CoinGeckoClient) ResolveID(base string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.coins) == 0 {
		return ""
	}

	return ResolveCoinGeckoID(base, c.coins)
}

// GetMarketCap returns the market cap for a given CoinGecko ID.
func (c *CoinGeckoClient) GetMarketCap(id string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if m, ok := c.markets[strings.ToLower(id)]; ok {
		return m.MarketCap
	}
	return 0
}

// ResolveCoinGeckoID implements the duplicate-safe scoring resolver.
func ResolveCoinGeckoID(base string, coins []symbol.CoinGeckoCoin) string {
	baseUpper := strings.ToUpper(base)
	baseLower := strings.ToLower(base)

	type scored struct {
		coin  symbol.CoinGeckoCoin
		score float64
	}

	var candidates []scored

	for _, coin := range coins {
		if !strings.EqualFold(coin.Symbol, baseUpper) {
			continue
		}

		score := 0.0

		// +0.5 symbol match (already filtered)
		score += 0.5

		// +0.2 name relevance
		if strings.Contains(strings.ToLower(coin.Name), baseLower) {
			score += 0.2
		}

		// +0.3 known IDs
		knownIDs := map[string]bool{
			"bitcoin": true, "ethereum": true, "solana": true,
			"ripple": true, "binancecoin": true,
		}
		if knownIDs[strings.ToLower(coin.ID)] {
			score += 0.3
		}

		// -0.5 penalty for wrapped/bridged tokens
		lowerID := strings.ToLower(coin.ID)
		if strings.Contains(lowerID, "wrapped") ||
			strings.Contains(lowerID, "wormhole") ||
			strings.Contains(lowerID, "binance-peg") ||
			strings.Contains(lowerID, "bridged") {
			score -= 0.5
		}

		candidates = append(candidates, scored{coin: coin, score: score})
	}

	if len(candidates) == 0 {
		return ""
	}

	// Select highest score, tie-break by shortest ID, then first
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.score > best.score {
			best = c
		} else if c.score == best.score {
			if len(c.coin.ID) < len(best.coin.ID) {
				best = c
			}
		}
	}

	return best.coin.ID
}
