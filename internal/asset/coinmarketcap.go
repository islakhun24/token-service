package asset

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"futures-symbol-module/pkg/httpclient"
)

// CMCCoin represents a cryptocurrency entry from CoinMarketCap listing API.
type CMCCoin struct {
	ID                 int       `json:"id"`
	Name               string    `json:"name"`
	Symbol             string    `json:"symbol"`
	Slug               string    `json:"slug"`
	CMCRank            int       `json:"cmcRank"`
	MarketPairCount    int       `json:"marketPairCount"`
	CirculatingSupply  float64   `json:"circulatingSupply"`
	TotalSupply        float64   `json:"totalSupply"`
	MaxSupply          float64   `json:"maxSupply"`
	Quotes             []CMCQuote `json:"quotes"`
}

// CMCQuote represents a quote in a specific currency.
type CMCQuote struct {
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	MarketCap float64 `json:"marketCap"`
}

// CMCListingResponse represents the CoinMarketCap listing API response.
type CMCListingResponse struct {
	Data struct {
		CryptoCurrencyList []CMCCoin `json:"cryptoCurrencyList"`
		TotalCount         string    `json:"totalCount"`
	} `json:"data"`
	Status struct {
		ErrorCode    string `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	} `json:"status"`
}

// CoinMarketCapClient provides cached CoinMarketCap data.
type CoinMarketCapClient struct {
	client *httpclient.Client
	mu     sync.RWMutex
	coins  map[string]CMCCoin // key: uppercase symbol (e.g. "BTC")
}

// NewCoinMarketCapClient creates a new CoinMarketCap client.
func NewCoinMarketCapClient() *CoinMarketCapClient {
	return &CoinMarketCapClient{
		client: httpclient.New(),
		coins:  make(map[string]CMCCoin),
	}
}

// RefreshData fetches cryptocurrency listing data from CoinMarketCap.
// Fetches in batches until all coins are covered (up to maxCoins).
func (c *CoinMarketCapClient) RefreshData(ctx context.Context, maxCoins int) error {
	allCoins := make(map[string]CMCCoin)
	batchSize := 500
	start := 1

	for start <= maxCoins {
		limit := batchSize
		if start+limit-1 > maxCoins {
			limit = maxCoins - start + 1
		}

		coins, err := c.fetchPage(ctx, start, limit)
		if err != nil {
			log.Printf("[CMC] Error fetching page start=%d limit=%d: %v", start, limit, err)
			break
		}

		if len(coins) == 0 {
			break
		}

		for _, coin := range coins {
			sym := strings.ToUpper(coin.Symbol)
			// Keep the highest ranked (lowest cmcRank) coin for each symbol
			if existing, ok := allCoins[sym]; !ok || (coin.CMCRank > 0 && coin.CMCRank < existing.CMCRank) {
				allCoins[sym] = coin
			}
		}

		log.Printf("[CMC] Fetched %d coins (start=%d, total cached: %d)", len(coins), start, len(allCoins))
		start += batchSize
	}

	c.mu.Lock()
	c.coins = allCoins
	c.mu.Unlock()

	log.Printf("[CMC] Data refreshed: %d unique symbols cached", len(allCoins))
	return nil
}

// fetchPage fetches a single page of cryptocurrency listings.
func (c *CoinMarketCapClient) fetchPage(ctx context.Context, start, limit int) ([]CMCCoin, error) {
	url := fmt.Sprintf(
		"https://api.coinmarketcap.com/data-api/v3/cryptocurrency/listing?start=%d&limit=%d&sortBy=rank&sortType=desc&convert=USD&cryptoType=all&tagType=all&audited=false&aux=ath,atl,high24h,low24h,num_market_pairs,cmc_rank,date_added,max_supply,circulating_supply,total_supply,volume_7d,volume_30d,self_reported_circulating_supply,self_reported_market_cap",
		start, limit,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Required headers for CoinMarketCap unofficial API
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", "https://coinmarketcap.com")
	req.Header.Set("Referer", "https://coinmarketcap.com/")
	req.Header.Set("Platform", "web")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CMC API status: %d", resp.StatusCode)
	}

	var data CMCListingResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.Status.ErrorCode != "0" {
		return nil, fmt.Errorf("CMC API error: %s", data.Status.ErrorMessage)
	}

	return data.Data.CryptoCurrencyList, nil
}

// GetCoin returns CMC data for a given base symbol.
func (c *CoinMarketCapClient) GetCoin(base string) (CMCCoin, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	coin, ok := c.coins[strings.ToUpper(base)]
	return coin, ok
}

// GetMarketCapUSD returns the USD market cap for a coin.
func (c *CoinMarketCapClient) GetMarketCapUSD(coin CMCCoin) float64 {
	for _, q := range coin.Quotes {
		if q.Name == "USD" {
			return q.MarketCap
		}
	}
	return 0
}
