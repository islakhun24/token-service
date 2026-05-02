package collector

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"futures-symbol-module/internal/symbol"
	"futures-symbol-module/pkg/httpclient"
)

type BybitCollector struct {
	client *httpclient.Client
}

func NewBybitCollector() *BybitCollector {
	return &BybitCollector{client: httpclient.New()}
}

func (c *BybitCollector) Name() string { return "bybit" }

type bybitResp struct {
	Result struct {
		List []struct {
			Symbol       string `json:"symbol"`
			BaseCoin     string `json:"baseCoin"`
			QuoteCoin    string `json:"quoteCoin"`
			Status       string `json:"status"`
			ContractType string `json:"contractType"`
		} `json:"list"`
	} `json:"result"`
}

func (c *BybitCollector) FetchSymbols(ctx context.Context) ([]symbol.SymbolInput, error) {
	baseURL := os.Getenv("BYBIT_API_URL")
	if baseURL == "" {
		baseURL = "https://api.bybit.com"
	}

	req, _ := http.NewRequestWithContext(ctx, "GET",
		baseURL+"/v5/market/instruments-info?category=linear",
		nil,
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data bybitResp
	json.NewDecoder(resp.Body).Decode(&data)

	var result []symbol.SymbolInput
	for _, s := range data.Result.List {
		if s.Status != "Trading" || s.ContractType != "LinearPerpetual" || s.QuoteCoin != "USDT" {
			continue
		}

		result = append(result, symbol.SymbolInput{
			Exchange: "bybit",
			Raw:      s.Symbol,
			Base:     s.BaseCoin,
			Quote:    s.QuoteCoin,
			Status:   s.Status,
		})
	}
	return result, nil
}
