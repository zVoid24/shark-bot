package bot

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"shark_bot/pkg/logger"
	"time"
)

type CRAPIClient struct {
	client *http.Client
	url    string
	token  string
	log    *slog.Logger
}

func NewCRAPIClient(apiUrl, token string) *CRAPIClient {
	return &CRAPIClient{
		client: &http.Client{Timeout: 30 * time.Second},
		url:    apiUrl,
		token:  token,
		log:    logger.New("crapi"),
	}
}

type CRAPIResponse struct {
	Status string `json:"status"`
	Total  int    `json:"total"`
	Data   []struct {
		Dt      string `json:"dt"`
		Num     string `json:"num"`
		Cli     string `json:"cli"`
		Message string `json:"message"`
		Payout  string `json:"payout"`
	} `json:"data"`
	Msg string `json:"msg"` // Error message if status is error
}

func (c *CRAPIClient) FetchSMS() ([]SMSResult, error) {
	if c.token == "" {
		return nil, fmt.Errorf("CR API token is missing")
	}

	u, err := url.Parse(c.url)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("token", c.token)
	q.Set("records", "200") // Fetch max records as per documentation
	u.RawQuery = q.Encode()

	c.log.Debug("fetching from CR API", "url", u.String())

	resp, err := c.client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CR API request failed with status: %d", resp.StatusCode)
	}

	var apiResp CRAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if apiResp.Status == "error" {
		return nil, fmt.Errorf("CR API error: %s", apiResp.Msg)
	}

	var results []SMSResult
	for _, item := range apiResp.Data {
		results = append(results, SMSResult{
			DateTime: item.Dt,
			Number:   item.Num,
			Service:  item.Cli,
			Message:  item.Message,
		})
	}

	return results, nil
}
