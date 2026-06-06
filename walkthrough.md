# Walkthrough - Stock Market CLI (`stock`)

This document summarizes the changes, test status, and manual validation steps for the newly built Go-based Stock Market CLI.

## Changes Made

1.  **Project Initialization**: Created Go module `stock` and fetched Cobra dependencies.
2.  **Configuration Helper (`pkg/config`)**:
    - Created configuration structures for API key storage and watchlists.
    - Implemented cross-platform standard config directory lookup via `os.UserConfigDir()`.
    - Added API key masking helper (`a*****************z`).
    - Added duplicate and missing symbol warnings for watchlists.
    - Fully unit tested in `pkg/config/config_test.go`.
3.  **Finnhub API Client (`pkg/finnhub`)**:
    - Built a robust concurrent client using Go goroutines and channels to retrieve quotes.
    - Added support for fetching basic metrics, daily candle history, and insider transactions.
    - Handled rate limit responses (`429`) and invalid/empty responses explicitly.
    - Configured a 5-second HTTP request timeout.
    - Fully unit tested with mocked HTTP transports in `pkg/finnhub/client_test.go`.
4.  **Technical Math Library (`pkg/technical`)**:
    - Written calculations for 50-day and 200-day Simple Moving Average (SMA).
    - Written calculations for 14-day Wilder's Relative Strength Index (RSI).
    - Fully unit tested in `pkg/technical/indicators_test.go`.
5.  **CLI Command Routing (`cmd/`)**:
    - Set up `root.go` with global flags (`--config`, `--json`, `--no-color`).
    - Created `config.go` with `config get` and `config set api-key`.
    - Created `quote.go` with custom padding table layout.
    - Created `watch.go` supporting `watch`, `watch add`, `watch remove`, and `watch list`.
    - **`macro.go`**: Queries macro index ETFs (`SPY`, `QQQ`, `DIA`, `IWM`, `VIXY`, `TLT`) concurrently.
    - **`sector.go`**: Queries 11 SPDR Sector ETFs concurrently, sorting them by daily change.
    - **`analyze.go`**: Queries and renders fundamental metrics (PE, PEG, ROIC, ROE, Gross Margin, Debt/Equity).
    - **`technical.go`**: Queries candle history to calculate and show current SMA-50/200, RSI-14, and volume breakout.
    - **`insider.go`**: Summarizes 6 months of executive insider buying and selling.
    - **`dashboard.go`**: Aggregates macro index metrics, top sector rotations, and watchlist performance in a single combined screen.
6.  **Entrypoint (`main.go`)**: Entrypoint calling `cmd.Execute()`.

---

## Verification Results

### Automated Tests
The whole test suite compiles and runs successfully with the Go race detector enabled:
```bash
$ go test -race -cover -v ./...
=== RUN   TestMaskAPIKey
--- PASS: TestMaskAPIKey (0.00s)
=== RUN   TestWatchlistOperations
--- PASS: TestWatchlistOperations (0.00s)
=== RUN   TestLoadSave
--- PASS: TestLoadSave (0.00s)
=== RUN   TestLoadCorruptedConfig
--- PASS: TestLoadCorruptedConfig (0.00s)
PASS
ok  	stock/pkg/config	1.558s	coverage: 69.2% of statements

=== RUN   TestFetchQuote
--- PASS: TestFetchQuote (0.01s)
=== RUN   TestFetchQuotesConcurrent
--- PASS: TestFetchQuotesConcurrent (0.00s)
=== RUN   TestFetchQuoteContextCancellation
--- PASS: TestFetchQuoteContextCancellation (0.00s)
=== RUN   TestFetchBasicFinancials
--- PASS: TestFetchBasicFinancials (0.00s)
=== RUN   TestFetchCandles
--- PASS: TestFetchCandles (0.00s)
=== RUN   TestFetchInsiderTransactions
--- PASS: TestFetchInsiderTransactions (0.00s)
PASS
ok  	stock/pkg/finnhub	1.996s	coverage: 77.9% of statements

=== RUN   TestCalculateSMA
--- PASS: TestCalculateSMA (0.00s)
=== RUN   TestCalculateRSI
--- PASS: TestCalculateRSI (0.00s)
PASS
ok  	stock/pkg/technical	1.742s	coverage: 100.0% of statements
```

---

## Manual Verification Steps

Set up the key in your local test config file:
```bash
go run ./cmd/stock --config test_config.json config set api-key YOUR_FINNHUB_TOKEN
```

### 1. Macro Scan
Query major indexes, bond proxies, and volatility:
```bash
go run ./cmd/stock --config test_config.json macro
```

### 2. Sector Scan
Query ranked sector ETF rotators:
```bash
go run ./cmd/stock --config test_config.json sector
```

### 3. Company Fundamentals
Scan core company scorecard:
```bash
go run ./cmd/stock --config test_config.json analyze AAPL
```

### 4. Technical Scan
Scan SMA support and RSI momentum levels:
```bash
go run ./cmd/stock --config test_config.json technical AAPL
```

### 5. Insider Trading
Check executive buying/selling balance:
```bash
go run ./cmd/stock --config test_config.json insider AAPL
```

### 6. Combined Market Dashboard
Show a single top-down market dashboard (Indices + Sector Leaderboard + Watchlist Status):
```bash
# Add some symbols to watchlist first
go run ./cmd/stock --config test_config.json watch add AAPL MSFT GOOG
# Show dashboard
go run ./cmd/stock --config test_config.json dashboard
```
