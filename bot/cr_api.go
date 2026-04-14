package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"shark_bot/pkg/logger"
	"time"
)


type CRAPIClient struct {
	baseURL string
	token   string
	client  *http.Client
}

type CRAPIResponse struct {
	Status string `json:"status"`
	Total  int    `json:"total"`
	Data   []struct {
		DT      string `json:"dt"`
		Num     string `json:"num"`
		CLI     string `json:"cli"`
		Message string `json:"message"`
		Payout  string `json:"payout"`
	} `json:"data"`
	Msg string `json:"msg,omitempty"`
}

func NewCRAPIClient(baseURL, token string) *CRAPIClient {
	return &CRAPIClient{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *CRAPIClient) FetchSMS() ([]SMSResult, error) {
	if c.token == "" {
		return nil, fmt.Errorf("CR API token is missing")
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CR API URL: %w", err)
	}

	q := u.Query()
	q.Set("token", c.token)
	q.Set("records", "200") // Fetch max records to avoid missing any
	u.RawQuery = q.Encode()

	logger.L.Debug("fetching from CR API", "url", u.String())

	resp, err := c.client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status error: %d", resp.StatusCode)
	}

	var apiResp CRAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	if apiResp.Status == "error" {
		return nil, fmt.Errorf("API error: %s", apiResp.Msg)
	}

	var results []SMSResult
	for _, item := range apiResp.Data {
		results = append(results, SMSResult{
			DateTime: item.DT,
			Number:   item.Num,
			Service:  item.CLI,
			Message:  item.Message,
			Account:  "CR-API", // Tag to distinguish from scrapers
		})
	}

	return results, nil
}
