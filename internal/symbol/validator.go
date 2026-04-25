package symbol

import (
	"strings"
)

// Validator validates canonical symbols.
type Validator struct{}

// NewValidator creates a new symbol validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks if a canonical symbol meets requirements.
func (v *Validator) Validate(sym CanonicalSymbol) bool {
	// Reject empty base or quote
	if sym.Base == "" || sym.Quote == "" {
		return false
	}

	// Only USDT quote pairs (production futures focus)
	if !strings.EqualFold(sym.Quote, "USDT") {
		return false
	}

	// Must be PERP type
	if !strings.EqualFold(sym.Type, "PERP") {
		return false
	}

	return true
}
