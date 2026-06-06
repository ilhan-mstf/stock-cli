# Feature Extension Plan: Professional Market Dashboard

This document details how we can expand the lightweight `stock` CLI into a comprehensive, professional-grade market dashboard. It maps the requirements of a top-down investing framework (Macro $\rightarrow$ Sector $\rightarrow$ Company $\rightarrow$ Valuation $\rightarrow$ Technicals) to specific Go packages, CLI interfaces, and Finnhub API endpoints.

---

## 1. Top-Down Feature Roadmap & API Mapping

Here is how we can implement each level of the requested framework using the Finnhub API:

### A. The Macro Layer (`stock macro`)
*   **Indices & Volatility**: Track major index ETFs (`SPY` for S&P 500, `QQQ` for NASDAQ, `DIA` for Dow Jones, `IWM` for Russell 2000) and `VIX` using the `/quote` endpoint.
*   **Bond Yields**: Monitor the 10-year Treasury yield (`^TNX` or equivalent ETF proxy) to track changes in interest rate expectations.
*   **Economic Indicators**: Query Finnhub's Economic Indicator endpoint `/country/indicator?code=US` to retrieve key macro values:
    *   `CPI` / `PPI` (Inflation trends)
    *   `Unemployment Rate` (Jobs data)
    *   `Real GDP Growth` (Economic health)

### B. Sector Performance (`stock sector`)
*   **Sector Rotation Tracking**: Create a list of the 11 Select Sector SPDR ETFs representing major US market sectors:
    *   Technology (`XLK`), Financials (`XLF`), Healthcare (`XLV`), Consumer Discretionary (`XLY`), Industrials (`XLI`), Energy (`XLE`), Utilities (`XLU`), Consumer Staples (`XLP`), Materials (`XLB`), Real Estate (`XLRE`), Communication Services (`XLC`).
*   **Execution**: Query quotes for all 11 ETFs concurrently, format them by daily price percentage change, and print a ranked list showing which sectors are leading or lagging.

### C. Deep Company Analysis (`stock analyze <SYMBOL>`)
*   **Finnhub Endpoint**: Query `/stock/metric?symbol=<SYMBOL>&metric=all`. This returns comprehensive financial statement summaries, growth metrics, and valuation multiples.
*   **Financial Health & Returns**:
    *   *ROIC / ROE*: Retrieve `roicTTM` (Return on Invested Capital) and `roeTTM` (Return on Equity).
    *   *FCF Yield*: Calculate or retrieve FCF Yield (`freeCashFlowYieldTTM`).
    *   *Debt Levels*: Retrieve `totalDebt/totalEquityTTM` and interest coverage.
*   **Revenue & Earnings Growth**:
    *   *Growth Rates*: Retrieve `revenueGrowth3Y` and `epsGrowth3Y` (3-year compound annual growth rates).
*   **Valuation Multiples**:
    *   Extract current and historical averages for `peNormalizedTTM`, `peBasicExclExtraTTM`, `psTTM`, and `pegTTM`.

### D. Execution & Technical Timing (`stock technical <SYMBOL>`)
*   **Finnhub Endpoint**: Query `/stock/candle?symbol=<SYMBOL>&resolution=D&from=<START>&to=<END>`.
*   **Indicators**:
    *   *Moving Averages*: Calculate 50-day and 200-day Simple Moving Averages (SMA) from the closing price history.
    *   *RSI (Relative Strength Index)*: Calculate the 14-day RSI to detect overbought ($>70$) or oversold ($<30$) levels.
    *   *Volume Breakout*: Compare current volume against 20-day average trading volume.

### E. Insider Activity (`stock insider <SYMBOL>`)
*   **Finnhub Endpoint**: Query `/stock/insider-transactions?symbol=<SYMBOL>`.
*   **Analysis**: Filter transactions to aggregate shares bought vs. sold by C-suite executives and board directors over the last 3-6 months.

---

## 2. CLI Command Structure

We would introduce these new commands alongside our existing watchlist and quote commands:

```bash
# 1. Macro Dashboard: Prints economic indicators, VIX, and index performance
stock macro

# 2. Sector Rotation: Ranked list of sector performance (best to worst)
stock sector

# 3. Company Fundamentals: Renders the core fundamental scorecard
stock analyze AAPL

# 4. Technical Indicator Scanner: SMA-50/200 positions, RSI, and volume breakouts
stock technical AAPL

# 5. Insider Transactions: Summarizes executive transactions
stock insider AAPL

# 6. Combined 10-Minute Daily Dashboard: Aggregates Macro + Sectors + Watchlist updates
stock dashboard
```

---

## 3. Visual Dashboard Design (`stock dashboard`)

Running `stock dashboard` would produce a beautiful, color-coded multi-section layout:

```
=========================================
      DAILY FINANCIAL DASHBOARD
=========================================

--- MARKET INDEX PROGRESS ---
SPY (S&P 500)    5,350.25   +0.45%  [████░░░░░░]
QQQ (NASDAQ)    18,620.10   +0.82%  [████████░░]
VIX (Volatility)    13.40   -1.25%  (Risk-On)
10Y Treasury Yield   4.28%   +0.02%

--- SECTOR ROTATION LEADERS ---
1. XLK (Technology)      +1.45% 🟩
2. XLF (Financials)      +0.32% 🟩
3. XLE (Energy)          -0.89% 🟥

--- MACRO ECONOMIC WATCH ---
Inflation (CPI):    +3.4% YoY  (Stable)
Jobs Data (PMI):    51.2       (Expansion)
Fed Fund Rate:      5.25% - 5.50%

--- WATCHLIST STATUS ---
AAPL   190.25  +0.75%  (PE: 28.5)
MSFT   420.40  -0.12%  (PE: 35.2)  *Earnings in 3 Days
```

---

## 4. Suggested Implementation Stages

If you want to build this, we can divide the work into modular phases:

1.  **Phase 1: Company Scorecard (`stock analyze`)**
    *   Extend the Finnhub client to support `/stock/metric`.
    *   Extract ROIC, FCF, Debt, and Valuation.
    *   Build a structured command to print a table summarizing fundamentals.
2.  **Phase 2: Sector & Index Indicators (`stock sector`, `stock macro`)**
    *   Concurrently fetch indices (SPY, QQQ, VIX) and sector ETFs.
    *   Fetch macro economic variables.
3.  **Phase 3: Technical Scan Engine (`stock technical`)**
    *   Extend client to fetch daily candle data.
    *   Write arithmetic helpers in Go to calculate SMA-50, SMA-200, and 14-day RSI.
4.  **Phase 4: Dashboard Aggregation (`stock dashboard`)**
    *   Combine macro, sector, watchlist status, and news events into a single terminal screen.
