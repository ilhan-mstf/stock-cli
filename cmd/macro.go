package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"stock/pkg/config"
	"stock/pkg/finnhub"
	"strings"

	"github.com/spf13/cobra"
)

var macroSymbols = []string{"SPY", "QQQ", "DIA", "IWM", "VIXY", "TLT"}
var macroNames = map[string]string{
	"SPY":  "S&P 500 Index ETF",
	"QQQ":  "Nasdaq 100 Index ETF",
	"DIA":  "Dow Jones Industrial ETF",
	"IWM":  "Russell 2000 Index ETF",
	"VIXY": "VIX Short-Term Futures ETF",
	"TLT":  "20+ Year Treasury Bond ETF",
}

var macroCmd = &cobra.Command{
	Use:   "macro",
	Short: "Display macro economic and market-level index indicators",
	Long:  `Scan major market indexes (S&P 500, Nasdaq, Dow, Russell), volatility proxies, and long-term bond yields.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(CfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		apiKey, err := getAPIKey(os.Getenv("FINNHUB_API_KEY"), cfg.APIKey)
		if err != nil {
			return err
		}

		client := finnhub.NewClient(apiKey, nil)
		quotes, errs := client.FetchQuotesConcurrent(context.Background(), macroSymbols)

		var validQuotes []finnhub.Quote
		var hasErrors bool

		for i, qErr := range errs {
			if qErr != nil {
				fmt.Fprintf(os.Stderr, "Error fetching macro index %q: %v\n", macroSymbols[i], qErr)
				hasErrors = true
			} else {
				validQuotes = append(validQuotes, quotes[i])
			}
		}

		if len(validQuotes) == 0 {
			return fmt.Errorf("failed to retrieve macro index quotes")
		}

		if JSONOutput {
			data, err := json.MarshalIndent(validQuotes, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		} else {
			fmt.Println(bold(fmt.Sprintf("%s%s%s%s",
				formatCell("SYMBOL", 10, nil),
				formatCell("INDEX / ASSET NAME", 30, nil),
				formatCell("PRICE", 12, nil),
				formatCell("CHANGE (%)", 20, nil),
			)))
			fmt.Println(strings.Repeat("-", 72))
			for _, q := range validQuotes {
				name := macroNames[q.Symbol]
				fmt.Println(formatMacroRow(q.Symbol, name, q))
			}
		}

		if hasErrors {
			os.Exit(1)
		}
		return nil
	},
}

func formatMacroRow(symbol string, name string, q finnhub.Quote) string {
	symCell := formatCell(symbol, 10, nil)
	nameCell := formatCell(name, 30, nil)
	priceCell := formatCell(fmt.Sprintf("%.2f", q.Current), 12, nil)

	changeVal := q.Change
	changeStr := fmt.Sprintf("%.2f", changeVal)
	if changeVal > 0 {
		changeStr = "+" + changeStr
	}
	pctStr := fmt.Sprintf("%.2f%%", q.PercentChange)
	if changeVal > 0 {
		pctStr = "+" + pctStr
	}

	var colorFn func(string) string
	if changeVal > 0 {
		colorFn = green
	} else if changeVal < 0 {
		colorFn = red
	}

	changeCell := formatCell(fmt.Sprintf("%s (%s)", changeStr, pctStr), 20, colorFn)
	return fmt.Sprintf("%s%s%s%s", symCell, nameCell, priceCell, changeCell)
}

func init() {
	RootCmd.AddCommand(macroCmd)
}
