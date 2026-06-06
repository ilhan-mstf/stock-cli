# Product Requirements Document (PRD) - Stock Market CLI (`stock`)

## 1. Product Overview
`stock` is a lightweight, high-performance command-line interface (CLI) tool written in Go that allows users to fetch real-time stock quotes, perform fundamental analysis, scan technical indicators, monitor insider activity, and manage a personalized watchlist directly from their terminal.

The application uses the official [Finnhub API](https://finnhub.io/) as its primary data provider. It is designed for developers, system administrators, and terminal-centric power users who need quick market updates and a top-down market scanning framework.

---

## 2. Target Audience & Use Cases
*   **The Keyboard-Driven Investor**: Wants to scan macro indicators, sector rotations, and individual company scorecards without leaving the terminal.
*   **The Technical Swing Trader**: Wants to quickly compute moving averages and RSI levels to identify entries and exits.
*   **The Automation Engineer**: Wants structured JSON outputs of technicals, fundamentals, and insider trades for custom scripting.

---

## 3. Functional Requirements

### 3.1. API Authentication & Precedence
*   Supports the Finnhub API Key via the environment variable `FINNHUB_API_KEY` (first priority) or local config (second priority).

### 3.2. Configuration & Watchlist Management
*   Persists watchlist symbols and API key configuration in `stock-cli/config.json`.
*   Includes validation to prevent blank symbols and duplicate symbols.

### 3.3. Extended Command Line Interface (CLI)

#### Subcommands

| Command | Args | Description | Example Output / Action |
| :--- | :--- | :--- | :--- |
| `stock quote <SYMBOL...>` | `>= 1` | Fetches real-time price quotes. | Terminal table of price, change, percent change, high, low, open, prev close. |
| `stock watch` | `0` | Fetches quotes for all watchlist symbols. | Displays quotes for all symbols in the watchlist. |
| `stock watch add/remove <SYMBOL...>` | `>= 1` | Manages watchlist symbols. | Adds or removes symbols with duplicates/missing checks. |
| `stock watch list` | `0` | Lists watchlist ticker symbols. | Prints watchlist symbols line-by-line. |
| `stock config set api-key <key>` | `1` | Saves the API key to configuration. | Writes key to config file. |
| `stock config get` | `0` | Displays current configuration. | Prints key (masked) and watchlist. |
| **`stock macro`** | `0` | Displays macro indices, volatility, and yields. | Prints quotes for S&P 500 (`SPY`), Nasdaq (`QQQ`), Dow Jones (`DIA`), Russell 2000 (`IWM`), Volatility (`VIXY`), and Treasury Yield. |
| **`stock sector`** | `0` | Displays sector performance ranking. | Renders the 11 Select Sector SPDR ETFs ranked by daily performance. |
| **`stock analyze <SYMBOL>`** | `1` | Performs deep fundamental analysis. | Prints valuation multiples (P/E, P/S, PEG), margins, growth rates, debt, and ROIC/ROE. |
| **`stock technical <SYMBOL>`** | `1` | Scans technical indicator setups. | Computes and prints current price vs. SMA-50/200, 14-day RSI status, and trading volume comparison. |
| **`stock insider <SYMBOL>`** | `1` | Scans executive insider activity. | Summarizes net buying/selling by C-suite executives and directors over the last 3-6 months. |
| **`stock dashboard`** | `0` | Displays the daily market dashboard. | Combines macro indices, top sector rotations, watchlist updates, and economic releases in one view. |

---

## 4. Non-Functional Requirements

### 4.1. Performance & Concurrency
*   Macro scans (6 indices) and sector scans (11 ETFs) must be fetched **concurrently** via goroutines to minimize network latency.
*   All outbound requests must enforce a 5-second network timeout.

### 4.2. Security
*   The API key must be masked in `config get`.
*   Config file saved permissions restricted to `0600`.

### 4.3. User Experience & Aesthetics
*   **Colors**: ANSI color-coding for daily changes (Green for positive, Red for negative, Yellow for flat) across quotes, sectors, indices, and RSI levels.
*   **Decimal Formatting**: Strict rounding of prices, percentages, multiples, and ratios to exactly 2 decimal places (`%.2f`).
*   **Indicator Alerts**:
    *   Highlight overbought RSI ($>70$) in Red (alert) and oversold RSI ($<30$) in Green (opportunity).
    *   Highlight strong fundamental targets (e.g. ROIC $> 15\%$) in Green.

---

## 5. Technical Stack
*   **Language**: Go 1.18+
*   **Libraries**: Cobra (`github.com/spf13/cobra`) for command routing. Standard library for HTTP requests, math calculations, and JSON parsing.
