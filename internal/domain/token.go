package domain

import "time"

// CanonicalToken = satu token, multi-exchange, multi-chain
type CanonicalToken struct {
	ID              int                `db:"id" json:"id"`
	CanonicalSymbol string             `db:"canonical_symbol" json:"canonicalSymbol"`
	CoinGeckoID     string             `db:"coingecko_id" json:"coingeckoId"`
	Name            string             `db:"name" json:"name"`
	Category        string             `db:"category" json:"category"`
	IsActive        bool               `db:"is_active" json:"isActive"`
	ArkhamReady     bool               `db:"arkham_ready" json:"arkhamReady"`
	CreatedAt       time.Time          `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time          `db:"updated_at" json:"updatedAt"`
	ExchangeMaps    []ExchangeMapping  `json:"exchangeMaps,omitempty"`
	Contracts       map[string]string  `json:"contracts,omitempty"`
}

// ExchangeMapping = cara setiap exchange menyebut token ini
type ExchangeMapping struct {
	ID             int    `db:"id" json:"id"`
	CanonicalID    int    `db:"canonical_id" json:"canonicalId"`
	Exchange       string `db:"exchange" json:"exchange"`
	MarketType     string `db:"market_type" json:"marketType"`
	ExchangeSymbol string `db:"exchange_symbol" json:"exchangeSymbol"`
	BaseAsset      string `db:"base_asset" json:"baseAsset"`
	QuoteAsset     string `db:"quote_asset" json:"quoteAsset"`
	NormalizedBase string `db:"normalized_base" json:"normalizedBase"`
	IsActive       bool   `db:"is_active" json:"isActive"`
}

// TokenContract = contract address per chain
type TokenContract struct {
	ID              int    `db:"id" json:"id"`
	CanonicalID     int    `db:"canonical_id" json:"canonicalId"`
	Chain           string `db:"chain" json:"chain"`
	ContractAddress string `db:"contract_address" json:"contractAddress"`
	Decimals        int    `db:"decimals" json:"decimals"`
}

// ExchangeSymbol = format untuk API/WebSocket call
type ExchangeSymbol struct {
	Exchange string `json:"exchange"`
	Symbol   string `json:"symbol"`
	Base     string `json:"base"`
	Quote    string `json:"quote"`
}

// TokenStats = summary untuk dashboard
type TokenStats struct {
	Total         int `json:"total"`
	Active        int `json:"active"`
	ArkhamReady   int `json:"arkhamReady"`
	MultiExchange int `json:"multiExchange"`
	Aliases       int `json:"aliases"`
}

// CreateTokenRequest = input dari frontend
type CreateTokenRequest struct {
	Canonical   string            `json:"canonical" binding:"required"`
	Name        string            `json:"name"`
	CoinGeckoID string            `json:"coingeckoId" binding:"required"`
	Category    string            `json:"category"`
	Binance     string            `json:"binance"`
	KuCoin      string            `json:"kucoin"`
	Bybit       string            `json:"bybit"`
	OKX         string            `json:"okx"`
	Contracts   map[string]string `json:"contracts"`
}

// NormalizeRequest = input untuk normalize endpoint
type NormalizeRequest struct {
	Exchange string `json:"exchange" binding:"required"`
	Symbol   string `json:"symbol" binding:"required"`
}

// NormalizeResponse = output normalize
type NormalizeResponse struct {
	Canonical string `json:"canonical"`
	BaseAsset string `json:"baseAsset"`
	Exchange  string `json:"exchange"`
}

// SyncResult = hasil sync per exchange
type SyncResult struct {
	Exchange    string `json:"exchange"`
	Total       int    `json:"total"`
	Inserted    int    `json:"inserted"`
	Updated     int    `json:"updated"`
	Failed      int    `json:"failed"`
	DurationMs  int64  `json:"durationMs"`
}

// SyncAllResponse = hasil sync semua exchange
type SyncAllResponse struct {
	Results []SyncResult `json:"results"`
}


// GenerateRequest = input untuk generate symbols dari canonical
type GenerateRequest struct {
	Canonical string `json:"canonical" binding:"required"`
}

// GenerateResponse = generated symbols untuk semua platform
type GenerateResponse struct {
	Canonical  string            `json:"canonical"`
	Binance    string            `json:"binance"`
	KuCoin     string            `json:"kucoin"`
	Bybit      string            `json:"bybit"`
	OKX        string            `json:"okx"`
	Arkham     string            `json:"arkham"`
	CoinGecko  string            `json:"coingecko"`
}

// BulkGenerateRequest = generate dari list binance symbols
type BulkGenerateRequest struct {
	Symbols []string `json:"symbols" binding:"required"`
}

// BulkGenerateResponse = hasil generate banyak token
type BulkGenerateResponse struct {
	Items []GenerateResponse `json:"items"`
	Total int                `json:"total"`
}
