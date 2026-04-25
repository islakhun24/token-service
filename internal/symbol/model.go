package symbol

// CanonicalSymbol represents a normalized futures symbol.
type CanonicalSymbol struct {
	Base  string `json:"base"`
	Quote string `json:"quote"`
	Type  string `json:"type"`
}

// ExchangeSymbol maps an exchange-specific symbol string to its canonical form.
type ExchangeSymbol struct {
	Exchange  string
	Raw       string
	Canonical CanonicalSymbol
}

// RegistryEntry represents a matched symbol across all exchanges.
type RegistryEntry struct {
	Base        string            `json:"base"`
	Quote       string            `json:"quote"`
	Type        string            `json:"type"`
	Rank        int               `json:"rank"`
	Exchanges   map[string]string `json:"exchanges"`
	Available   map[string]bool   `json:"available"`
	CoinGeckoID string            `json:"coingecko_id"`
	Market      MarketData        `json:"market"`
}

// MarketData holds market enrichment data.
type MarketData struct {
	MarketCap float64 `json:"market_cap"`
}

// FuturesPairsResponse is the strict JSON output format.
type FuturesPairsResponse struct {
	Version string          `json:"version"`
	Total   int             `json:"total"`
	Data    []RegistryEntry `json:"data"`
}

// CoinGeckoCoin represents an entry from /coins/list.
type CoinGeckoCoin struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

// CoinGeckoMarket represents an entry from /coins/markets.
type CoinGeckoMarket struct {
	ID           string  `json:"id"`
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	MarketCap    float64 `json:"market_cap"`
	CurrentPrice float64 `json:"current_price"`
}
