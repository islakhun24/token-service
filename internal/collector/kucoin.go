package collector

import (
	"context"
	"encoding/json"
	"net/http"

	"futures-symbol-module/internal/symbol"
	"futures-symbol-module/pkg/httpclient"
)

type KucoinCollector struct {
	client *httpclient.Client
}

func NewKucoinCollector() *KucoinCollector {
	return &KucoinCollector{client: httpclient.New()}
}

func (c *KucoinCollector) Name() string { return "kucoin" }

type kucoinResp struct {
	Data []struct {
		Symbol        string `json:"symbol"`
		BaseCurrency  string `json:"baseCurrency"`
		QuoteCurrency string `json:"quoteCurrency"`
		Status        string `json:"status"`
	} `json:"data"`
}

func (c *KucoinCollector) FetchSymbols(ctx context.Context) ([]symbol.SymbolInput, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET",
		"https://api-futures.kucoin.com/api/v1/contracts/active",
		nil,
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data kucoinResp
	json.NewDecoder(resp.Body).Decode(&data)

	var result []symbol.SymbolInput
	for _, s := range data.Data {
		if s.Status != "Open" || s.QuoteCurrency != "USDT" {
			continue
		}

		result = append(result, symbol.SymbolInput{
			Exchange: "kucoin",
			Raw:      s.Symbol,
			Base:     s.BaseCurrency,
			Quote:    s.QuoteCurrency,
			Status:   s.Status,
		})
	}
	return result, nil
}