package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"stock/pkg/config"
	"stock/pkg/finnhub"
	"strings"

	"github.com/spf13/cobra"
)

var sectorETFs = []string{"XLK", "XLF", "XLV", "XLY", "XLI", "XLE", "XLU", "XLP", "XLB", "XLRE", "XLC"}
var sectorNames = map[string]string{
	"XLK":  "Technology",
	"XLF":  "Financials",
	"XLV":  "Healthcare",
	"XLY":  "Consumer Discretionary",
	"XLI":  "Industrials",
	"XLE":  "Energy",
	"XLU":  "Utilities",
	"XLP":  "Consumer Staples",
	"XLB":  "Materials",
	"XLRE": "Real Estate",
	"XLC":  "Communication Services",
}

type SectorResult struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	PercentChange float64 `json:"percent_change"`
}

var sectorCmd = &cobra.Command{
	Use:   "sector",
	Short: "Display ranked sector rotation performance",
	Long:  `Queries 11 Select Sector SPDR ETFs representing core sectors of the economy and displays them ranked by daily performance.`,
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
		quotes, errs := client.FetchQuotesConcurrent(context.Background(), sectorETFs)

		var results []SectorResult
		var hasErrors bool

		for i, qErr := range errs {
			if qErr != nil {
				fmt.Fprintf(os.Stderr, "Error fetching sector %q: %v\n", sectorETFs[i], qErr)
				hasErrors = true
			} else {
				results = append(results, SectorResult{
					Symbol:        sectorETFs[i],
					Name:          sectorNames[sectorETFs[i]],
					Price:         quotes[i].Current,
					PercentChange: quotes[i].PercentChange,
				})
			}
		}

		if len(results) == 0 {
			return fmt.Errorf("failed to retrieve sector quotes")
		}

		// Sort results descending by percent change
		sort.Slice(results, func(i, j int) bool {
			return results[i].PercentChange > results[j].PercentChange
		})

		if JSONOutput {
			data, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		} else {
			fmt.Println(bold(fmt.Sprintf("%s%s%s%s",
				formatCell("RANK / SYM", 12, nil),
				formatCell("SECTOR GROUP", 26, nil),
				formatCell("PRICE", 10, nil),
				formatCell("PERF (%)", 15, nil),
			)))
			fmt.Println(strings.Repeat("-", 63))
			for i, r := range results {
				rankSym := fmt.Sprintf("%d. %s", i+1, r.Symbol)
				fmt.Println(formatSectorRow(rankSym, r))
			}
		}

		if hasErrors {
			os.Exit(1)
		}
		return nil
	},
}

func formatSectorRow(rankSym string, r SectorResult) string {
	rankCell := formatCell(rankSym, 12, nil)
	nameCell := formatCell(r.Name, 26, nil)
	priceCell := formatCell(fmt.Sprintf("%.2f", r.Price), 10, nil)

	perfStr := fmt.Sprintf("%.2f%%", r.PercentChange)
	if r.PercentChange > 0 {
		perfStr = "+" + perfStr
	}

	var colorFn func(string) string
	if r.PercentChange > 0 {
		colorFn = green
	} else if r.PercentChange < 0 {
		colorFn = red
	}

	perfCell := formatCell(perfStr, 15, colorFn)
	return fmt.Sprintf("%s%s%s%s", rankCell, nameCell, priceCell, perfCell)
}

func init() {
	RootCmd.AddCommand(sectorCmd)
}
