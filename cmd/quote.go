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

var quoteCmd = &cobra.Command{
	Use:   "quote [symbols...]",
	Short: "Get real-time quotes for one or more ticker symbols",
	Args:  cobra.MinimumNArgs(1),
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
		quotes, errs := client.FetchQuotesConcurrent(context.Background(), args)

		var validQuotes []finnhub.Quote
		var hasErrors bool

		for i, qErr := range errs {
			if qErr != nil {
				fmt.Fprintf(os.Stderr, "Error fetching %q: %v\n", strings.ToUpper(args[i]), qErr)
				hasErrors = true
			} else {
				validQuotes = append(validQuotes, quotes[i])
			}
		}

		if len(validQuotes) == 0 {
			return fmt.Errorf("no stock quotes retrieved successfully")
		}

		if JSONOutput {
			data, err := json.MarshalIndent(validQuotes, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to serialize quotes to JSON: %w", err)
			}
			fmt.Println(string(data))
		} else {
			printQuoteHeaders()
			for _, q := range validQuotes {
				fmt.Println(formatQuoteRow(q))
			}
		}

		if hasErrors {
			os.Exit(1)
		}
		return nil
	},
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func formatCell(val string, width int, colorFn func(string) string) string {
	padded := padRight(val, width)
	if colorFn != nil {
		return colorFn(padded)
	}
	return padded
}

func printQuoteHeaders() {
	headers := fmt.Sprintf("%s%s%s%s%s%s%s%s",
		formatCell("SYMBOL", 10, nil),
		formatCell("PRICE", 12, nil),
		formatCell("CHANGE", 10, nil),
		formatCell("% CHANGE", 10, nil),
		formatCell("OPEN", 10, nil),
		formatCell("HIGH", 10, nil),
		formatCell("LOW", 10, nil),
		formatCell("PREV CLOSE", 12, nil),
	)
	fmt.Println(bold(headers))
	fmt.Println(strings.Repeat("-", len(headers)))
}

func formatQuoteRow(q finnhub.Quote) string {
	sym := formatCell(q.Symbol, 10, nil)
	price := formatCell(fmt.Sprintf("%.2f", q.Current), 12, nil)

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

	changeCell := formatCell(changeStr, 10, colorFn)
	pctCell := formatCell(pctStr, 10, colorFn)

	open := formatCell(fmt.Sprintf("%.2f", q.Open), 10, nil)
	high := formatCell(fmt.Sprintf("%.2f", q.High), 10, nil)
	low := formatCell(fmt.Sprintf("%.2f", q.Low), 10, nil)
	prevClose := formatCell(fmt.Sprintf("%.2f", q.PrevClose), 12, nil)

	return fmt.Sprintf("%s%s%s%s%s%s%s%s", sym, price, changeCell, pctCell, open, high, low, prevClose)
}

func init() {
	RootCmd.AddCommand(quoteCmd)
}
