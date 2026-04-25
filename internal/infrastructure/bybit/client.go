package bybit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type InstrumentsResponse struct {
	Result struct {
		List []Instrument `json:"list"`
	} `json:"result"`
}

type Instrument struct {
	Symbol       string `json:"symbol"`
	BaseAsset    string `json:"baseCoin"`
	QuoteAsset   string `json:"quoteCoin"`
	Status       string `json:"status"`
	ContractType string `json:"contractType"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: "https://api.bybit.com",
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) GetLinearPerpetuals() ([]Instrument, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/v5/market/instruments-info?category=linear&status=Trading")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bybit status: %d", resp.StatusCode)
	}

	var data InstrumentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var result []Instrument
	for _, inst := range data.Result.List {
		if inst.ContractType == "LinearPerpetual" && inst.QuoteAsset == "USDT" {
			result = append(result, inst)
		}
	}
	return result, nil
}
