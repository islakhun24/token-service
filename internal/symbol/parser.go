package symbol

import (
	"regexp"
	"strings"
)

var (
	okxPattern      = regexp.MustCompile(`^([A-Z0-9]+)-USDT-SWAP$`)
	bitgetPattern   = regexp.MustCompile(`^([A-Z0-9]+)USDT_(UMCBL|CMCBL)$`)
	mexcPattern     = regexp.MustCompile(`^([A-Z0-9]+)_USDT$`)
	kucoinPattern   = regexp.MustCompile(`^([A-Z0-9]+)-USDTM$`)
	standardPattern = regexp.MustCompile(`^([A-Z0-9]+)USDT$`)
)

// 🔥 NEW: Field-based override
type SymbolInput struct {
	Exchange string

	// raw fallback
	Raw string

	// optional (from API)
	Base  string
	Quote string

	Status string
}

var baseAlias = map[string]string{
	"XBT": "BTC",
}

func normalizeBase(base string) string {
	base = strings.ToUpper(base)

	if v, ok := baseAlias[base]; ok {
		return v
	}

	return base
}

// 🔥 NEW ENTRY POINT (USE THIS)
func Normalize(input SymbolInput) (CanonicalSymbol, bool) {

	// 1. FILTER STATUS
	if input.Status != "" {
		switch input.Exchange {
		case "kucoin":
			if input.Status != "Open" {
				return CanonicalSymbol{}, false
			}
		case "binance":
			if input.Status != "TRADING" {
				return CanonicalSymbol{}, false
			}
		case "bybit":
			if input.Status != "Trading" {
				return CanonicalSymbol{}, false
			}
		case "bitget":
			if !strings.EqualFold(input.Status, "online") &&
				!strings.EqualFold(input.Status, "normal") {
				return CanonicalSymbol{}, false
			}
		}
	}

	// 2. FIELD-BASED (PRIORITY)
	if input.Base != "" && input.Quote == "USDT" {
		return CanonicalSymbol{
			Base:  normalizeBase(strings.ToUpper(input.Base)),
			Quote: "USDT",
			Type:  "PERP",
		}, true
	}

	// 3. FALLBACK → REGEX
	return ParseSymbol(input.Raw)
}

// 🔥 KEEP YOUR EXISTING FUNCTION (slightly improved)
func ParseSymbol(raw string) (CanonicalSymbol, bool) {
	raw = strings.TrimSpace(strings.ToUpper(raw))
	if raw == "" {
		return CanonicalSymbol{}, false
	}

	// OKX
	if matches := okxPattern.FindStringSubmatch(raw); matches != nil {
		return CanonicalSymbol{
			Base:  matches[1],
			Quote: "USDT",
			Type:  "PERP",
		}, true
	}

	// Bitget
	if matches := bitgetPattern.FindStringSubmatch(raw); matches != nil {
		return CanonicalSymbol{
			Base:  matches[1],
			Quote: "USDT",
			Type:  "PERP",
		}, true
	}

	// MEXC
	if matches := mexcPattern.FindStringSubmatch(raw); matches != nil {
		return CanonicalSymbol{
			Base:  matches[1],
			Quote: "USDT",
			Type:  "PERP",
		}, true
	}

	// KuCoin
	if matches := kucoinPattern.FindStringSubmatch(raw); matches != nil {
		return CanonicalSymbol{
			Base:  matches[1],
			Quote: "USDT",
			Type:  "PERP",
		}, true
	}

	// Binance / Bybit
	if matches := standardPattern.FindStringSubmatch(raw); matches != nil {
		return CanonicalSymbol{
			Base:  matches[1],
			Quote: "USDT",
			Type:  "PERP",
		}, true
	}

	return CanonicalSymbol{}, false
}
