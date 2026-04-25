package normalizer

import (
	"strings"
	"sync"
)

type SymbolNormalizer struct {
	aliases map[string]string
	mu      sync.RWMutex
}

func NewSymbolNormalizer(initialAliases map[string]string) *SymbolNormalizer {
	aliases := make(map[string]string)
	for k, v := range initialAliases {
		aliases[strings.ToUpper(k)] = strings.ToUpper(v)
	}
	// Default aliases
	defaults := map[string]string{
		"XBT":       "BTC",
		"1000PEPE":  "PEPE",
		"1000SHIB":  "SHIB",
		"1000FLOKI": "FLOKI",
		"1000BONK":  "BONK",
		"1000LUNC":  "LUNC",
		"WETH":      "ETH",
		"WBTC":      "BTC",
		"WBNB":      "BNB",
		"WMATIC":    "MATIC",
		"WSOL":      "SOL",
	}
	for k, v := range defaults {
		if _, ok := aliases[k]; !ok {
			aliases[k] = v
		}
	}
	return &SymbolNormalizer{aliases: aliases}
}

func (n *SymbolNormalizer) AddAlias(exchangeName, canonical string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.aliases[strings.ToUpper(exchangeName)] = strings.ToUpper(canonical)
}

func (n *SymbolNormalizer) NormalizeBaseAsset(base string) string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	base = strings.ToUpper(base)
	if canonical, ok := n.aliases[base]; ok {
		return canonical
	}
	return base
}

func (n *SymbolNormalizer) ExtractBaseFromBinance(symbol string) string {
	symbol = strings.ToUpper(symbol)
	symbol = strings.TrimSuffix(symbol, "USDT")
	symbol = strings.TrimSuffix(symbol, "BUSD")
	symbol = strings.TrimSuffix(symbol, "USD")
	symbol = strings.TrimSuffix(symbol, "USDC")
	if idx := strings.Index(symbol, "_"); idx > 0 {
		symbol = symbol[:idx]
	}
	return n.NormalizeBaseAsset(symbol)
}

func (n *SymbolNormalizer) ExtractBaseFromKuCoin(symbol string) string {
	symbol = strings.ToUpper(symbol)
	symbol = strings.TrimSuffix(symbol, "USDTM")
	symbol = strings.TrimSuffix(symbol, "USDMM")
	parts := strings.Split(symbol, "-")
	if len(parts) >= 2 {
		return n.NormalizeBaseAsset(parts[0])
	}
	return n.NormalizeBaseAsset(symbol)
}

func (n *SymbolNormalizer) ExtractBaseFromOKX(symbol string) string {
	symbol = strings.ToUpper(symbol)
	symbol = strings.TrimSuffix(symbol, "-SWAP")
	symbol = strings.TrimSuffix(symbol, "-PERP")
	parts := strings.Split(symbol, "-")
	if len(parts) >= 2 {
		return n.NormalizeBaseAsset(parts[0])
	}
	return n.NormalizeBaseAsset(symbol)
}

func (n *SymbolNormalizer) ExtractBaseFromBybit(symbol string) string {
	return n.ExtractBaseFromBinance(symbol)
}

func (n *SymbolNormalizer) ExtractBaseFromGateIO(symbol string) string {
	symbol = strings.ToUpper(symbol)
	parts := strings.Split(symbol, "_")
	if len(parts) >= 2 {
		return n.NormalizeBaseAsset(parts[0])
	}
	return n.NormalizeBaseAsset(symbol)
}

func (n *SymbolNormalizer) ExtractBase(exchange, symbol string) string {
	switch strings.ToLower(exchange) {
	case "binance":
		return n.ExtractBaseFromBinance(symbol)
	case "kucoin":
		return n.ExtractBaseFromKuCoin(symbol)
	case "okx":
		return n.ExtractBaseFromOKX(symbol)
	case "bybit":
		return n.ExtractBaseFromBybit(symbol)
	case "gateio":
		return n.ExtractBaseFromGateIO(symbol)
	default:
		return n.NormalizeBaseAsset(symbol)
	}
}
