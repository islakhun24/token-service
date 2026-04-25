package collector

import (
	"context"
	"encoding/json"
	_"fmt"
	"net/http"

	"futures-symbol-module/internal/symbol"
	"futures-symbol-module/pkg/httpclient"
)

type BinanceCollector struct {
	client *httpclient.Client
}

func NewBinanceCollector() *BinanceCollector {
	return &BinanceCollector{client: httpclient.New()}
}

func (c *BinanceCollector) Name() string { return "binance" }

type binanceResp struct {
	Symbols []struct {
		Symbol       string `json:"symbol"`
		BaseAsset    string `json:"baseAsset"`
		QuoteAsset   string `json:"quoteAsset"`
		Status       string `json:"status"`
		ContractType string `json:"contractType"`
	} `json:"symbols"`
}

func (c *BinanceCollector) FetchSymbols(ctx context.Context) ([]symbol.SymbolInput, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://fapi.binance.com/fapi/v1/exchangeInfo", nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data binanceResp
	json.NewDecoder(resp.Body).Decode(&data)

	var result []symbol.SymbolInput
	for _, s := range data.Symbols {
		if s.Status != "TRADING" || s.ContractType != "PERPETUAL" {
			continue
		}
		if s.QuoteAsset != "USDT" {
			continue
		}

		result = append(result, symbol.SymbolInput{
			Exchange: "binance",
			Raw:      s.Symbol,
			Base:     s.BaseAsset,
			Quote:    s.QuoteAsset,
			Status:   s.Status,
		})
	}
	return result, nil
}