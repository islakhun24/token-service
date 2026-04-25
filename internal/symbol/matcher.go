package symbol

import (
	"fmt"
	"strings"
)

// Matcher groups exchange symbols by their canonical identity.
type Matcher struct{}

// NewMatcher creates a new symbol matcher.
func NewMatcher() *Matcher {
	return &Matcher{}
}

// GroupKey generates a unique grouping key for a canonical symbol.
func (m *Matcher) GroupKey(sym CanonicalSymbol) string {
	return fmt.Sprintf("%s_%s_%s", strings.ToUpper(sym.Base), strings.ToUpper(sym.Quote), strings.ToUpper(sym.Type))
}

// Match groups a slice of exchange symbols by canonical identity.
func (m *Matcher) Match(symbols []ExchangeSymbol) map[string][]ExchangeSymbol {
	groups := make(map[string][]ExchangeSymbol)
	for _, sym := range symbols {
		key := m.GroupKey(sym.Canonical)
		groups[key] = append(groups[key], sym)
	}
	return groups
}
