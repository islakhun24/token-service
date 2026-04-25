package collector

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"futures-symbol-module/internal/symbol"
	"futures-symbol-module/pkg/httpclient"
)

type BitgetCollector struct {
	client *httpclient.Client
}

func NewBitgetCollector() *BitgetCollector {
	return &BitgetCollector{
		client: httpclient.New(),
	}
}

func (c *BitgetCollector) Name() string {
	return "bitget"
}

type bitgetResp struct {
	Data []struct {
		Symbol     string `json:"symbol"`
		BaseCoin   string `json:"baseCoin"`
		QuoteCoin  string `json:"quoteCoin"`
		SymbolType string `json:"symbolType"`
		Status     string `json:"status"`
	} `json:"data"`
}

func (c *BitgetCollector) FetchSymbols(ctx context.Context) ([]symbol.SymbolInput, error) {

	req, _ := http.NewRequestWithContext(ctx, "GET",
		"https://api.bitget.com/api/v2/mix/market/contracts?productType=USDT-FUTURES",
		nil,
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data bitgetResp
	json.NewDecoder(resp.Body).Decode(&data)

	var result []symbol.SymbolInput
	for _, s := range data.Data {

		if !strings.EqualFold(s.SymbolType, "perpetual") {
			continue
		}

		if !strings.EqualFold(s.QuoteCoin, "USDT") {
			continue
		}

		result = append(result, symbol.SymbolInput{
			Exchange: "bitget",
			Raw:      s.Symbol,
			Base:     strings.ToUpper(s.BaseCoin),
			Quote:    strings.ToUpper(s.QuoteCoin),
			Status:   s.Status,
		})
	}

	return result, nil
}
