package collector

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"futures-symbol-module/internal/symbol"
	"futures-symbol-module/pkg/httpclient"
)

type MexcCollector struct {
	client *httpclient.Client
}

func NewMexcCollector() *MexcCollector {
	return &MexcCollector{client: httpclient.New()}
}

func (c *MexcCollector) Name() string { return "mexc" }

type mexcResp struct {
	Data []struct {
		Symbol string `json:"symbol"`
	} `json:"data"`
}

func (c *MexcCollector) FetchSymbols(ctx context.Context) ([]symbol.SymbolInput, error) {
	baseURL := os.Getenv("MEXC_API_URL")
	if baseURL == "" {
		baseURL = "https://contract.mexc.com"
	}

	req, _ := http.NewRequestWithContext(ctx, "GET",
		baseURL+"/api/v1/contract/detail",
		nil,
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data mexcResp
	json.NewDecoder(resp.Body).Decode(&data)

	var result []symbol.SymbolInput
	for _, s := range data.Data {

		// fallback parsing only
		result = append(result, symbol.SymbolInput{
			Exchange: "mexc",
			Raw:      s.Symbol,
		})
	}
	return result, nil
}