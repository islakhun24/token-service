package okx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type InstrumentsResponse struct {
	Data []Instrument `json:"data"`
}

type Instrument struct {
	Symbol     string `json:"instId"`
	BaseAsset  string `json:"baseCcy"`
	QuoteAsset string `json:"quoteCcy"`
	Status     string `json:"state"`
	InstType   string `json:"instType"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: "https://www.okx.com",
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) GetSwapPerpetuals() ([]Instrument, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/v5/public/instruments?instType=SWAP")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("okx status: %d", resp.StatusCode)
	}

	var data InstrumentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var result []Instrument
	for _, inst := range data.Data {
		if inst.Status == "live" && inst.QuoteAsset == "USDT" {
			result = append(result, inst)
		}
	}
	return result, nil
}
