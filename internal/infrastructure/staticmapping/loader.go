package staticmapping

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type StaticToken struct {
	Symbol      string            `json:"symbol"`
	CoinGeckoID string            `json:"coingecko_id"`
	Name        string            `json:"name"`
	Contracts   map[string]string `json:"contracts"`
}

type Loader struct {
	tokens map[string]StaticToken
}

func NewLoader(filepath string) (*Loader, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("read static mapping: %w", err)
	}

	var tokens []StaticToken
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("parse static mapping: %w", err)
	}

	m := make(map[string]StaticToken)
	for _, t := range tokens {
		m[strings.ToUpper(t.Symbol)] = t
	}

	return &Loader{tokens: m}, nil
}

func (l *Loader) GetBySymbol(symbol string) (StaticToken, bool) {
	t, ok := l.tokens[strings.ToUpper(symbol)]
	return t, ok
}

func (l *Loader) GetAll() []StaticToken {
	var result []StaticToken
	for _, t := range l.tokens {
		result = append(result, t)
	}
	return result
}

func (l *Loader) Count() int {
	return len(l.tokens)
}
