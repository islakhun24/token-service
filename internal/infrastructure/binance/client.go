package binance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type FuturesExchangeInfo struct {
	Symbols []FuturesSymbol `json:"symbols"`
}

type FuturesSymbol struct {
	Symbol       string `json:"symbol"`
	BaseAsset    string `json:"baseAsset"`
	QuoteAsset   string `json:"quoteAsset"`
	Status       string `json:"status"`
	ContractType string `json:"contractType"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) GetFuturesExchangeInfo() (*FuturesExchangeInfo, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/fapi/v1/exchangeInfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("binance status: %d", resp.StatusCode)
	}

	var data FuturesExchangeInfo
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
