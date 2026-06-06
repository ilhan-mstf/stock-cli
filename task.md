# Stock Market CLI Dashboard Tasks

- [x] Implement new structures and methods in `pkg/finnhub/client.go`
  - [x] Add Basic Financials structures & `/stock/metric` handler
  - [x] Add Candles structures & `/stock/candle` handler
  - [x] Add Insider Transactions structures & `/stock/insider-transactions` handler
- [x] Implement `pkg/technical/indicators.go` (SMA and RSI math calculations)
- [x] Implement Cobra commands
  - [x] `cmd/macro.go`
  - [x] `cmd/sector.go`
  - [x] `cmd/analyze.go`
  - [x] `cmd/technical.go`
  - [x] `cmd/insider.go`
  - [x] `cmd/dashboard.go`
- [x] Implement unit tests
  - [x] Test client additions in `pkg/finnhub/client_test.go`
  - [x] Test math calculations in `pkg/technical/indicators_test.go`
- [x] Verify implementation with automated tests
- [x] Perform manual verification steps
