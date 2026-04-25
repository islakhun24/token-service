package kucoin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type FuturesSymbolsResponse struct {
	Data []FuturesSymbol `json:"data"`
}

type FuturesSymbol struct {
	Symbol     string `json:"symbol"`
	BaseAsset  string `json:"baseCurrency"`
	QuoteAsset string `json:"quoteCurrency"`
	Status     string `json:"status"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: "https://api-futures.kucoin.com",
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) GetFuturesSymbols() ([]FuturesSymbol, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v1/contracts/active")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("kucoin status: %d", resp.StatusCode)
	}

	var data FuturesSymbolsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Data, nil
}
