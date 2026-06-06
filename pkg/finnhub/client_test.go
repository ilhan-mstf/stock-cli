package finnhub

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchQuote(t *testing.T) {
	tests := []struct {
		name          string
		symbol        string
		apiKey        string
		handler       http.HandlerFunc
		expectedQuote *Quote
		expectedErr   string
	}{
		{
			name:   "success AAPL",
			symbol: "AAPL",
			apiKey: "test_key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("token") != "test_key" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if r.URL.Query().Get("symbol") != "AAPL" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"c":150.25,"d":2.5,"dp":1.69,"h":152.0,"l":149.5,"o":149.8,"pc":147.75,"t":1600000000}`)
			},
			expectedQuote: &Quote{
				Symbol:        "AAPL",
				Current:       150.25,
				Change:        2.5,
				PercentChange: 1.69,
				High:          152.0,
				Low:           149.5,
				Open:          149.8,
				PrevClose:     147.75,
				Timestamp:     1600000000,
			},
		},
		{
			name:   "empty symbol error",
			symbol: "",
			apiKey: "test_key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedErr: "symbol cannot be empty",
		},
		{
			name:   "missing api key error",
			symbol: "AAPL",
			apiKey: "",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedErr: "Finnhub API key is required",
		},
		{
			name:   "unauthorized status 401",
			symbol: "AAPL",
			apiKey: "invalid_key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			expectedErr: "invalid Finnhub API key (401 Unauthorized)",
		},
		{
			name:   "rate limit status 429",
			symbol: "AAPL",
			apiKey: "test_key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
			},
			expectedErr: "Finnhub API rate limit exceeded (429 Too Many Requests)",
		},
		{
			name:   "invalid symbol zero values return",
			symbol: "INVALID",
			apiKey: "test_key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"c":0,"d":null,"dp":null,"h":0,"l":0,"o":0,"pc":0,"t":0}`)
			},
			expectedErr: `symbol "INVALID" not found or invalid`,
		},
		{
			name:   "corrupt JSON response",
			symbol: "AAPL",
			apiKey: "test_key",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"c":150.25, invalid_json:`)
			},
			expectedErr: "failed to decode response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()

			client := NewClient(tc.apiKey, nil)
			client.SetBaseURL(server.URL)

			q, err := client.FetchQuote(context.Background(), tc.symbol)

			if tc.expectedErr != "" {
				if err == nil {
					t.Fatalf("Expected error %q, got nil", tc.expectedErr)
				}
				if !strings.Contains(err.Error(), tc.expectedErr) {
					t.Errorf("Expected error to contain %q, got %q", tc.expectedErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if q.Symbol != tc.expectedQuote.Symbol ||
				q.Current != tc.expectedQuote.Current ||
				q.Change != tc.expectedQuote.Change ||
				q.PercentChange != tc.expectedQuote.PercentChange ||
				q.Timestamp != tc.expectedQuote.Timestamp {
				t.Errorf("Fetched Quote = %+v; want %+v", q, tc.expectedQuote)
			}
		})
	}
}

func TestFetchQuotesConcurrent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		symbol := r.URL.Query().Get("symbol")
		w.Header().Set("Content-Type", "application/json")
		if symbol == "AAPL" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"c":150.0,"d":1.0,"dp":0.67,"h":151.0,"l":149.0,"o":149.5,"pc":149.0,"t":1600000000}`)
		} else if symbol == "MSFT" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"c":250.0,"d":-2.0,"dp":-0.79,"h":252.0,"l":249.0,"o":251.0,"pc":252.0,"t":1600000000}`)
		} else if symbol == "INVALID" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"c":0,"d":null,"dp":null,"h":0,"l":0,"o":0,"pc":0,"t":0}`)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := NewClient("test_key", nil)
	client.SetBaseURL(server.URL)

	symbols := []string{"AAPL", "INVALID", "MSFT"}
	quotes, errs := client.FetchQuotesConcurrent(context.Background(), symbols)

	if len(quotes) != 3 || len(errs) != 3 {
		t.Fatalf("Expected quotes/errors slices of length 3, got quotes:%d errs:%d", len(quotes), len(errs))
	}

	// AAPL success
	if errs[0] != nil {
		t.Errorf("AAPL error = %v; want nil", errs[0])
	}
	if quotes[0].Symbol != "AAPL" || quotes[0].Current != 150.0 {
		t.Errorf("AAPL quote = %+v; check symbol and current", quotes[0])
	}

	// INVALID failure
	if errs[1] == nil {
		t.Errorf("Expected INVALID error, got nil")
	} else if !strings.Contains(errs[1].Error(), `symbol "INVALID" not found or invalid`) {
		t.Errorf("Unexpected INVALID error message: %v", errs[1])
	}
	if quotes[1].Symbol != "" {
		t.Errorf("Expected empty quote at index 1, got %+v", quotes[1])
	}

	// MSFT success
	if errs[2] != nil {
		t.Errorf("MSFT error = %v; want nil", errs[2])
	}
	if quotes[2].Symbol != "MSFT" || quotes[2].Current != 250.0 {
		t.Errorf("MSFT quote = %+v; check symbol and current", quotes[2])
	}
}

func TestFetchQuoteContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"c":150.25,"t":1600000000}`)
	}))
	defer server.Close()

	client := NewClient("test_key", nil)
	client.SetBaseURL(server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel context

	_, err := client.FetchQuote(ctx, "AAPL")
	if err == nil {
		t.Error("Expected error when loading with cancelled context, got nil")
	}
}

func TestFetchBasicFinancials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"symbol":"AAPL","metricType":"all","metric":{"peNormalizedTTM":28.5,"roicTTM":18.4}}`)
	}))
	defer server.Close()

	client := NewClient("test_key", nil)
	client.SetBaseURL(server.URL)

	f, err := client.FetchBasicFinancials(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if f.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", f.Symbol)
	}
	pe, ok := f.Metric["peNormalizedTTM"].(float64)
	if !ok || pe != 28.5 {
		t.Errorf("Expected peNormalizedTTM to be 28.5, got %v", f.Metric["peNormalizedTTM"])
	}
}

func TestFetchCandles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"c":[150.0, 151.0],"h":[152.0, 153.0],"l":[149.0, 149.5],"o":[149.5, 150.5],"s":"ok","t":[1600000000, 1600086400],"v":[1000, 1200]}`)
	}))
	defer server.Close()

	client := NewClient("test_key", nil)
	client.SetBaseURL(server.URL)

	candles, err := client.FetchCandles(context.Background(), "AAPL", "D", 1600000000, 1600086400)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(candles.Close) != 2 || candles.Close[0] != 150.0 || candles.Status != "ok" {
		t.Errorf("Fetched candles mismatch: %+v", candles)
	}
}

func TestFetchInsiderTransactions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"symbol":"AAPL","data":[{"symbol":"AAPL","name":"Cook Tim","share":2000000,"change":-50000,"price":191.2,"transactionDate":"2026-05-12"}]}`)
	}))
	defer server.Close()

	client := NewClient("test_key", nil)
	client.SetBaseURL(server.URL)

	txs, err := client.FetchInsiderTransactions(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(txs) != 1 || txs[0].Name != "Cook Tim" || txs[0].Change != -50000 {
		t.Errorf("Fetched insider transactions mismatch: %+v", txs)
	}
}
