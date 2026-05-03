package symbol

import (
	_ "fmt"
	"strings"
)

// RegistryBuilder builds the unified registry from matched groups.
type RegistryBuilder struct {
	exchanges []string
}

// NewRegistryBuilder creates a registry builder with known exchange names.
func NewRegistryBuilder() *RegistryBuilder {
	return &RegistryBuilder{
		exchanges: []string{"binance", "okx", "bybit", "bitget", "mexc", "kucoin"},
	}
}

// Build constructs registry entries from matched groups.
func (rb *RegistryBuilder) Build(groups map[string][]ExchangeSymbol) []RegistryEntry {
	entries := make([]RegistryEntry, 0, len(groups))

	for _, group := range groups {
		if len(group) == 0 {
			continue
		}

		canonical := group[0].Canonical
		entry := RegistryEntry{
			Base:      canonical.Base,
			Quote:     canonical.Quote,
			Type:      canonical.Type,
			Exchanges: make(map[string]string),
			Available: make(map[string]bool),
		}

		// Initialize all exchanges as false
		for _, ex := range rb.exchanges {
			entry.Available[ex] = false
		}

		// Mark available exchanges and collect categories
		catSet := make(map[string]bool)
		for _, sym := range group {
			exKey := strings.ToLower(sym.Exchange)
			entry.Exchanges[exKey] = sym.Raw
			entry.Available[exKey] = true

			// Collect categories (from any exchange that provides them)
			for _, cat := range sym.Categories {
				if cat != "" && !catSet[cat] {
					catSet[cat] = true
					entry.Categories = append(entry.Categories, cat)
				}
			}
		}

		entries = append(entries, entry)
	}

	return entries
}
