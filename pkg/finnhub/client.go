package finnhub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Quote struct {
	Symbol        string  `json:"symbol"`
	Current       float64 `json:"c"`
	Change        float64 `json:"d"`
	PercentChange float64 `json:"dp"`
	High          float64 `json:"h"`
	Low           float64 `json:"l"`
	Open          float64 `json:"o"`
	PrevClose     float64 `json:"pc"`
	Timestamp     int64   `json:"t"`
}

type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Finnhub client.
// If httpClient is nil, a default client with a 5-second timeout is used.
func NewClient(apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	return &Client{
		apiKey:     apiKey,
		httpClient: httpClient,
		baseURL:    "https://finnhub.io/api/v1",
	}
}

// SetBaseURL overrides the default base URL (useful for testing/mocking).
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// FetchQuote retrieves the current quote for a specific ticker symbol.
func (c *Client) FetchQuote(ctx context.Context, symbol string) (*Quote, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	if c.apiKey == "" {
		return nil, fmt.Errorf("Finnhub API key is required")
	}

	u, err := url.Parse(c.baseURL + "/quote")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("symbol", symbol)
	q.Set("token", c.apiKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid Finnhub API key (401 Unauthorized)")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("Finnhub API rate limit exceeded (429 Too Many Requests)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Finnhub API returned HTTP status %d", resp.StatusCode)
	}

	var quote Quote
	if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	quote.Symbol = symbol
	if quote.Timestamp == 0 {
		return nil, fmt.Errorf("symbol %q not found or invalid", symbol)
	}

	return &quote, nil
}

// FetchQuotesConcurrent retrieves quotes for multiple ticker symbols in parallel.
// Returns a slice of quotes and a slice of errors matching the index of the requested symbols.
func (c *Client) FetchQuotesConcurrent(ctx context.Context, symbols []string) ([]Quote, []error) {
	quotes := make([]Quote, len(symbols))
	errs := make([]error, len(symbols))

	var wg sync.WaitGroup
	for i, symbol := range symbols {
		wg.Add(1)
		go func(idx int, sym string) {
			defer wg.Done()
			q, err := c.FetchQuote(ctx, sym)
			if err != nil {
				errs[idx] = err
			} else {
				quotes[idx] = *q
			}
		}(i, symbol)
	}
	wg.Wait()

	return quotes, errs
}

type BasicFinancials struct {
	Symbol     string                 `json:"symbol"`
	MetricType string                 `json:"metricType"`
	Metric     map[string]interface{} `json:"metric"`
}

type Candles struct {
	Close     []float64 `json:"c"`
	High      []float64 `json:"h"`
	Low       []float64 `json:"l"`
	Open      []float64 `json:"o"`
	Status    string    `json:"s"`
	Timestamp []int64   `json:"t"`
	Volume    []int64   `json:"v"`
}

type InsiderTransaction struct {
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name"`
	Share           int64   `json:"share"`
	Change          int64   `json:"change"`
	Price           float64 `json:"price"`
	TransactionDate string  `json:"transactionDate"`
}

type InsiderTransactionsResponse struct {
	Symbol string               `json:"symbol"`
	Data   []InsiderTransaction `json:"data"`
}

// FetchBasicFinancials retrieves key statistics and financial ratios for a symbol.
func (c *Client) FetchBasicFinancials(ctx context.Context, symbol string) (*BasicFinancials, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	if c.apiKey == "" {
		return nil, fmt.Errorf("Finnhub API key is required")
	}

	u, err := url.Parse(c.baseURL + "/stock/metric")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("symbol", symbol)
	q.Set("metric", "all")
	q.Set("token", c.apiKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid Finnhub API key (401 Unauthorized)")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("Finnhub API rate limit exceeded (429 Too Many Requests)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Finnhub API returned HTTP status %d", resp.StatusCode)
	}

	var financials BasicFinancials
	if err := json.NewDecoder(resp.Body).Decode(&financials); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	financials.Symbol = symbol
	if financials.Metric == nil {
		return nil, fmt.Errorf("no fundamental metrics returned for symbol %q", symbol)
	}

	return &financials, nil
}

// FetchCandles retrieves historical price candles for a symbol.
func (c *Client) FetchCandles(ctx context.Context, symbol string, resolution string, from, to int64) (*Candles, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	if c.apiKey == "" {
		return nil, fmt.Errorf("Finnhub API key is required")
	}

	u, err := url.Parse(c.baseURL + "/stock/candle")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("symbol", symbol)
	q.Set("resolution", resolution)
	q.Set("from", fmt.Sprintf("%d", from))
	q.Set("to", fmt.Sprintf("%d", to))
	q.Set("token", c.apiKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid Finnhub API key (401 Unauthorized)")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("Finnhub API rate limit exceeded (429 Too Many Requests)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Finnhub API returned HTTP status %d", resp.StatusCode)
	}

	var candles Candles
	if err := json.NewDecoder(resp.Body).Decode(&candles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if candles.Status == "no_data" || len(candles.Close) == 0 {
		return nil, fmt.Errorf("no historical candle data returned for symbol %q", symbol)
	}

	return &candles, nil
}

// FetchInsiderTransactions retrieves executive and institutional insider transactions.
func (c *Client) FetchInsiderTransactions(ctx context.Context, symbol string) ([]InsiderTransaction, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}
	if c.apiKey == "" {
		return nil, fmt.Errorf("Finnhub API key is required")
	}

	u, err := url.Parse(c.baseURL + "/stock/insider-transactions")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("symbol", symbol)
	q.Set("token", c.apiKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid Finnhub API key (401 Unauthorized)")
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("Finnhub API rate limit exceeded (429 Too Many Requests)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Finnhub API returned HTTP status %d", resp.StatusCode)
	}

	var response InsiderTransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data, nil
}
