package symbol

// CanonicalSymbol represents a normalized futures symbol.
type CanonicalSymbol struct {
	Base  string `json:"base"`
	Quote string `json:"quote"`
	Type  string `json:"type"`
}

// ExchangeSymbol maps an exchange-specific symbol string to its canonical form.
type ExchangeSymbol struct {
	Exchange   string
	Raw        string
	Canonical  CanonicalSymbol
	Categories []string
}

// RegistryEntry represents a matched symbol across all exchanges.
type RegistryEntry struct {
	Base              string            `json:"base"`
	Quote             string            `json:"quote"`
	Type              string            `json:"type"`
	Rank              int               `json:"rank"`
	Exchanges         map[string]string `json:"exchanges"`
	Available         map[string]bool   `json:"available"`
	CMCID             *int              `json:"cmc_id,omitempty"`
	CMCName           *string           `json:"cmc_name,omitempty"`
	CMCSlug           *string           `json:"cmc_slug,omitempty"`
	CMCRank           *int              `json:"cmc_rank,omitempty"`
	Market            MarketData        `json:"market"`
	CirculatingSupply float64           `json:"circulating_supply,omitempty"`
	TotalSupply       float64           `json:"total_supply,omitempty"`
	MaxSupply         float64           `json:"max_supply,omitempty"`
	Categories        []string          `json:"categories"`
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
