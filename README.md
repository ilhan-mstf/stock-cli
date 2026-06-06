# Stock Market CLI (`stock`)

A fast, lightweight, and professional-grade command-line interface written in Go to monitor market indicators, sector rotations, company fundamentals, technical levels, and insider trading using the official **Finnhub API**.

Follows a top-down investing framework: **Macro $\rightarrow$ Sector $\rightarrow$ Company $\rightarrow$ Valuation $\rightarrow$ Technicals $\rightarrow$ Risk**.

---

## 🚀 Key Features

*   **Unified Dashboard**: Get S&P/Nasdaq indices, ranked sectors, and your watchlist in one screen.
*   **Macro Environment**: Check indices, volatility proxies (VIX), and long-term Treasury bond yields concurrently.
*   **Sector Rotations**: Scan SPDR Sector ETFs ranked by performance to see where institutional money is flowing.
*   **Company Fundamentals**: Inspect multiples (P/E, PEG, P/S), return margins (ROIC, ROE), growth rates, and leverage.
*   **Technical Scanning**: Calculate 50-day/200-day SMAs, 14-day RSI (Wilder's), and check for volume breakouts.
*   **Insider Summaries**: Aggregate C-suite buying and selling activities over the last 6 months.
*   **Watchlist Management**: Add, remove, list, and monitor ticker quotes with duplicates and blank-space protections.
*   **Colorized Outputs**: Clean tabular layouts with ANSI coloring (green for positive/oversold, red for negative/overbought) and structured `--json` outputs for scripting.

---

## 🛠 Installation

### Prerequisites
Make sure you have **Go 1.18 or higher** installed.

### From Source
1.  Clone the repository and navigate into the folder:
    ```bash
    git clone https://github.com/username/stock-cli.git
    cd stock-cli
    ```
2.  Build the binary:
    ```bash
    go build -o stock ./cmd/stock
    ```
3.  Install it to your `$GOPATH/bin`:
    ```bash
    go install ./cmd/stock
    ```

---

## 🔑 Authentication Setup

The CLI requires a free API token from [Finnhub](https://finnhub.io/). 

The CLI checks for the token in the following order of precedence:
1.  **Environment Variable**:
    ```bash
    export FINNHUB_API_KEY="your_api_token_here"
    ```
2.  **Local Configuration File**:
    ```bash
    stock config set api-key your_api_token_here
    ```
    *This saves the token to the cross-platform standard configuration directory (`~/.config/stock-cli/config.json` on Linux/macOS).*

---

## 📖 Command Reference & Examples

### 1. Market Index Scan (`stock macro`)
Displays major indices (S&P 500, Nasdaq 100, Dow Jones, Russell 2000), volatility index, and bonds:
```bash
stock macro
```

### 2. Sector Rotation Performance (`stock sector`)
Displays the 11 Select Sector SPDR ETFs ranked by daily performance:
```bash
stock sector
```

### 3. Company Scorecard (`stock analyze <SYMBOL>`)
Fetches valuation multiples, returns on capital (ROIC/ROE), revenue growth, and debt health:
```bash
stock analyze TSLA
```

### 4. Technical Scan (`stock technical <SYMBOL>`)
Calculates the current price position relative to moving averages, RSI-14 status, and volume breakout ratio:
```bash
stock technical AAPL
```

### 5. Insider Trading Summary (`stock insider <SYMBOL>`)
Analyzes transactions by executives and directors over the last 6 months, rendering summaries and a recent log:
```bash
stock insider MSFT
```

### 6. Watchlist Management (`stock watch`)
*   **Fetch Watchlist Quotes**: `stock watch`
*   **Add Symbols**: `stock watch add AAPL GOOG NVDA`
*   **Remove Symbols**: `stock watch remove AAPL`
*   **List Symbols**: `stock watch list`

### 7. Unified Daily Dashboard (`stock dashboard`)
Deduplicates network calls to fetch macro indices, the top sector rotators, and your watchlist in a single parallel batch request:
```bash
stock dashboard
```

---

## 🎛 Global Flags

Add these flags to any command to customize output behavior:

*   `--json`: Outputs data in structured JSON format instead of terminal tables. Useful for scripting (e.g., piping to `jq`).
    ```bash
    stock quote AAPL MSFT --json
    ```
*   `--no-color`: Disables color-coded outputs. Useful when redirects are active.
    ```bash
    stock sector --no-color
    ```
*   `--config <path>`: Overrides the default configuration file location.
    ```bash
    stock --config ./test_config.json watch list
    ```
