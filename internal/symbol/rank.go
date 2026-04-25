package symbol

import (
	"sort"
)

// Ranker sorts and ranks registry entries by market cap.
type Ranker struct{}

// NewRanker creates a new ranker.
func NewRanker() *Ranker {
	return &Ranker{}
}

// Rank sorts entries by market cap descending and assigns rank.
// Entries with market_cap == 0 are placed at the end.
func (r *Ranker) Rank(entries []RegistryEntry) []RegistryEntry {
	// Separate entries with and without market cap
	var withCap, withoutCap []RegistryEntry
	for _, e := range entries {
		if e.Market.MarketCap > 0 {
			withCap = append(withCap, e)
		} else {
			withoutCap = append(withoutCap, e)
		}
	}

	// Sort by market cap descending
	sort.Slice(withCap, func(i, j int) bool {
		return withCap[i].Market.MarketCap > withCap[j].Market.MarketCap
	})

	// Assign rank
	for i := range withCap {
		withCap[i].Rank = i + 1
	}

	// Append zero-cap entries without rank
	for i := range withoutCap {
		withoutCap[i].Rank = 0
	}

	return append(withCap, withoutCap...)
}
