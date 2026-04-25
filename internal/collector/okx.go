package collector

import (
	"context"
	"encoding/json"
	"net/http"

	"futures-symbol-module/internal/symbol"
	"futures-symbol-module/pkg/httpclient"
)

type OKXCollector struct {
	client *httpclient.Client
}

func NewOKXCollector() *OKXCollector {
	return &OKXCollector{client: httpclient.New()}
}

func (c *OKXCollector) Name() string { return "okx" }

type okxResp struct {
	Data []struct {
		InstID string `json:"instId"`
		State  string `json:"state"`
	} `json:"data"`
}

func (c *OKXCollector) FetchSymbols(ctx context.Context) ([]symbol.SymbolInput, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET",
		"https://www.okx.com/api/v5/public/instruments?instType=SWAP",
		nil,
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data okxResp
	json.NewDecoder(resp.Body).Decode(&data)

	var result []symbol.SymbolInput
	for _, s := range data.Data {
		if s.State != "live" {
			continue
		}

		result = append(result, symbol.SymbolInput{
			Exchange: "okx",
			Raw:      s.InstID,
		})
	}
	return result, nil
}