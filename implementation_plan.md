# Implementation Plan - Stock Market CLI Dashboard Extensions

This document details the plan to implement macro scans, sector rankings, fundamental metrics, insider summaries, and technical indicator scans in the `stock` CLI.

## Proposed Changes

We will build these extensions in the `/Users/m.ilhan/personal/v2/stock-cli` workspace.

### 1. API Client Extensions (`pkg/finnhub`)

We need to add structures and methods to interface with three new Finnhub endpoints.

#### [MODIFY] [pkg/finnhub/client.go](file:///Users/m.ilhan/personal/v2/stock-cli/pkg/finnhub/client.go)

##### Basic Financials
Add `Metrics` and `BasicFinancials` structures to map the `/stock/metric` response:
```go
type BasicFinancials struct {
    Symbol     string                 `json:"symbol"`
    MetricType string                 `json:"metricType"`
    Metric     map[string]interface{} `json:"metric"`
}
```
*Note: Because Finnhub basic metrics return dynamic fields, using a map of string to interface is highly robust and avoids parsing issues. We will extract values like `peNormalizedTTM`, `peBasicExclExtraTTM`, `psTTM`, `pegTTM`, `roicTTM`, `roeTTM`, `freeCashFlowYieldTTM`, `totalDebt/totalEquityTTM`, `epsGrowthTTMYoy`, and `revenueGrowthTTMYoy` safely using helper functions.*

##### Candles
Add `Candles` structure to map `/stock/candle`:
```go
type Candles struct {
    Close     []float64 `json:"c"`
    High      []float64 `json:"h"`
    Low       []float64 `json:"l"`
    Open      []float64 `json:"o"`
    Status    string    `json:"s"`
    Timestamp []int64   `json:"t"`
    Volume    []int64   `json:"v"`
}
```

##### Insider Transactions
Add `InsiderTransaction` and `InsiderTransactionsResponse` structures to map `/stock/insider-transactions`:
```go
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
```

##### Client Methods
Add:
- `FetchBasicFinancials(ctx context.Context, symbol string) (*BasicFinancials, error)`
- `FetchCandles(ctx context.Context, symbol string, resolution string, from, to int64) (*Candles, error)`
- `FetchInsiderTransactions(ctx context.Context, symbol string) ([]InsiderTransaction, error)`

---

### 2. Technical Math helpers (`pkg/technical`)

#### [NEW] [pkg/technical/indicators.go](file:///Users/m.ilhan/personal/v2/stock-cli/pkg/technical/indicators.go)
Calculates indicators in pure Go without external dependencies:
*   `CalculateSMA(prices []float64, period int) (float64, error)`: Computes the Simple Moving Average of the last `period` closing prices. Returns error if prices length is less than `period`.
*   `CalculateRSI(prices []float64, period int) (float64, error)`: Computes the Relative Strength Index (RSI) using Wilder's smoothing technique. Returns the RSI value of the most recent close.
    *   *Wilder's RSI formula*:
        1. Calculate initial average gain/loss of the first `period` changes.
        2. Smooth successive gains/losses: `avgGain = (prevAvgGain * 13 + currentGain) / 14`.
        3. Compute RS: `avgGain / avgLoss`.
        4. Compute RSI: `100 - (100 / (1 + RS))`.

---

### 3. Cobra CLI Subcommands (`cmd/`)

We will add the new command logic.

#### [NEW] [cmd/macro.go](file:///Users/m.ilhan/personal/v2/stock-cli/cmd/macro.go)
Fetches quotes for `SPY`, `QQQ`, `DIA`, `IWM`, and `VIXY` (volatility proxy) concurrently.
Formats them into a market index table. Includes the 10-year Treasury yield or proxies where possible.

#### [NEW] [cmd/sector.go](file:///Users/m.ilhan/personal/v2/stock-cli/cmd/sector.go)
Queries Select Sector SPDR ETFs (`XLK`, `XLF`, `XLV`, `XLY`, `XLI`, `XLE`, `XLU`, `XLP`, `XLB`, `XLRE`, `XLC`) in parallel.
Sorts the sectors by daily percent change, printing a ranked leaderboard using color codes (Green for positive, Red for negative).

#### [NEW] [cmd/analyze.go](file:///Users/m.ilhan/personal/v2/stock-cli/cmd/analyze.go)
Fetches and formats basic financials.
Validates outputs and aligns them in a nice metrics card table showing:
*   Valuation: PE (TTM), Forward PE, PS, PEG.
*   Quality: ROIC, ROE, FCF Yield.
*   Financial Health: Debt-to-Equity ratio.
*   Growth: Revenue Growth (YoY), EPS Growth (YoY).

#### [NEW] [cmd/technical.go](file:///Users/m.ilhan/personal/v2/stock-cli/cmd/technical.go)
Queries 300 calendar days of daily candles (to cover at least 200 trading days).
Calculates:
*   SMA-50 and SMA-200.
*   14-day RSI (highlights $>70$ as overbought in red, $<30$ as oversold in green).
*   Trading volume breakout factor (today's volume vs 20-day average volume).

#### [NEW] [cmd/insider.go](file:///Users/m.ilhan/personal/v2/stock-cli/cmd/insider.go)
Queries the last 6 months of insider transactions.
Summarizes transactions: aggregates total shares bought and sold by key executives, computing net insider balance.

#### [NEW] [cmd/dashboard.go](file:///Users/m.ilhan/personal/v2/stock-cli/cmd/dashboard.go)
Aggregated dashboard command that pulls the indices (Macro), sector performance rank (Sector), and current watchlist quotes, presenting a single clean 10-minute snapshot.

---

## Verification Plan

### Automated Tests
*   `pkg/technical` Unit Tests: Tests SMA and RSI calculations against known mathematical inputs.
*   `pkg/finnhub` Client Mock Tests: Verifies `/stock/metric`, `/stock/candle`, and `/stock/insider-transactions` endpoints parser.
*   Complete tests: `go test -v ./...`

### Manual Verification
1.  **Macro & Sectors**:
    - `go run main.go macro`
    - `go run main.go sector` (Verify sorted performance)
2.  **Fundamental Analysis**:
    - `go run main.go analyze AAPL` (Confirm scorecard matches expected fields)
3.  **Technicals**:
    - `go run main.go technical AAPL` (Confirm computed RSI-14 and SMA-50/200 positions)
4.  **Insiders**:
    - `go run main.go insider AAPL` (Confirm executive summary output)
5.  **Dashboard**:
    - `go run main.go dashboard` (Confirm unified visual output)
