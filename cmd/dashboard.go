package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"stock/pkg/config"
	"stock/pkg/finnhub"

	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Display the unified daily market dashboard",
	Long:  `Combines major market indices (Macro), sector performance leaders/lagged (Sectors), and the user's watchlist quotes in a single screen.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(CfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Compile list of unique symbols to fetch in one concurrent batch
		symbolMap := make(map[string]bool)

		macroSyms := []string{"SPY", "QQQ", "VIXY"}
		for _, s := range macroSyms {
			symbolMap[s] = true
		}

		sectorSyms := []string{"XLK", "XLF", "XLV", "XLY", "XLI", "XLE", "XLU", "XLP", "XLB", "XLRE", "XLC"}
		for _, s := range sectorSyms {
			symbolMap[s] = true
		}

		for _, s := range cfg.Watchlist {
			symbolMap[s] = true
		}

		var uniqueSymbols []string
		for s := range symbolMap {
			uniqueSymbols = append(uniqueSymbols, s)
		}

		apiKey, err := getAPIKey(os.Getenv("FINNHUB_API_KEY"), cfg.APIKey)
		if err != nil {
			return err
		}

		client := finnhub.NewClient(apiKey, nil)
		quotes, errs := client.FetchQuotesConcurrent(context.Background(), uniqueSymbols)

		// Map quotes for easy lookups
		quotesLookup := make(map[string]finnhub.Quote)
		var hasErrors bool

		for i, sym := range uniqueSymbols {
			if errs[i] != nil {
				fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", sym, errs[i])
				hasErrors = true
			} else {
				quotesLookup[sym] = quotes[i]
			}
		}

		if len(quotesLookup) == 0 {
			return fmt.Errorf("failed to load dashboard statistics")
		}

		if JSONOutput {
			// Structured JSON output for the whole dashboard
			out := struct {
				Macro     []finnhub.Quote `json:"macro"`
				Sectors   []SectorResult  `json:"sectors"`
				Watchlist []finnhub.Quote `json:"watchlist"`
			}{}

			for _, s := range macroSyms {
				if q, ok := quotesLookup[s]; ok {
					out.Macro = append(out.Macro, q)
				}
			}

			var sectors []SectorResult
			for _, s := range sectorSyms {
				if q, ok := quotesLookup[s]; ok {
					sectors = append(sectors, SectorResult{
						Symbol:        s,
						Name:          sectorNames[s],
						Price:         q.Current,
						PercentChange: q.PercentChange,
					})
				}
			}
			sort.Slice(sectors, func(i, j int) bool {
				return sectors[i].PercentChange > sectors[j].PercentChange
			})
			out.Sectors = sectors

			for _, s := range cfg.Watchlist {
				if q, ok := quotesLookup[s]; ok {
					out.Watchlist = append(out.Watchlist, q)
				}
			}

			data, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		// Renders beautiful terminal UI
		fmt.Println(bold("\n=========================================\n       DAILY MARKET DASHBOARD\n========================================="))

		// 1. Render Macro
		fmt.Println(bold("\n--- MARKET INDEX PROGRESS ---"))
		for _, s := range macroSyms {
			q, ok := quotesLookup[s]
			if !ok {
				continue
			}
			name := macroNames[s]
			fmt.Println(formatDashboardRow(s, name, q))
		}

		// 2. Render Sector performance (Top 3 and Bottom 1)
		var sectors []SectorResult
		for _, s := range sectorSyms {
			if q, ok := quotesLookup[s]; ok {
				sectors = append(sectors, SectorResult{
					Symbol:        s,
					Name:          sectorNames[s],
					Price:         q.Current,
					PercentChange: q.PercentChange,
				})
			}
		}

		if len(sectors) > 0 {
			sort.Slice(sectors, func(i, j int) bool {
				return sectors[i].PercentChange > sectors[j].PercentChange
			})

			fmt.Println(bold("\n--- SECTOR ROTATION LEADERBOARD ---"))
			// Top 3
			topCount := 3
			if len(sectors) < topCount {
				topCount = len(sectors)
			}
			for i := 0; i < topCount; i++ {
				r := sectors[i]
				rank := fmt.Sprintf("%d. %s", i+1, r.Symbol)
				fmt.Println(formatDashboardSectorRow(rank, r))
			}

			// Bottom 1 (if we have more than 3 sectors)
			if len(sectors) > 3 {
				fmt.Println("...")
				r := sectors[len(sectors)-1]
				rank := fmt.Sprintf("%d. %s", len(sectors), r.Symbol)
				fmt.Println(formatDashboardSectorRow(rank, r))
			}
		}

		// 3. Render Watchlist
		fmt.Println(bold("\n--- WATCHLIST STATUS ---"))
		var watchlistQuotes []finnhub.Quote
		for _, s := range cfg.Watchlist {
			if q, ok := quotesLookup[s]; ok {
				watchlistQuotes = append(watchlistQuotes, q)
			}
		}

		if len(watchlistQuotes) == 0 {
			fmt.Println("  No active watchlist quotes found. Add symbols via 'stock watch add <symbol>'.")
		} else {
			for _, q := range watchlistQuotes {
				fmt.Println(formatDashboardRow(q.Symbol, "", q))
			}
		}
		fmt.Println()

		if hasErrors {
			os.Exit(1)
		}
		return nil
	},
}

func formatDashboardRow(symbol string, name string, q finnhub.Quote) string {
	symCell := formatCell(symbol, 10, nil)

	// Trim name if too long or empty
	if name == "" {
		name = "Watchlist Stock"
	}
	if len(name) > 24 {
		name = name[:21] + "..."
	}
	nameCell := formatCell(name, 24, nil)
	priceCell := formatCell(fmt.Sprintf("%.2f", q.Current), 12, nil)

	changeVal := q.Change
	pctStr := fmt.Sprintf("%.2f%%", q.PercentChange)
	if changeVal > 0 {
		pctStr = "+" + pctStr
	}

	var colorFn func(string) string
	var icon string
	if changeVal > 0 {
		colorFn = green
		icon = "🟩"
	} else if changeVal < 0 {
		colorFn = red
		icon = "🟥"
	} else {
		icon = "⬜"
	}

	perfCell := formatCell(fmt.Sprintf("%s %s", pctStr, icon), 15, colorFn)
	return fmt.Sprintf("%s%s%s%s", symCell, nameCell, priceCell, perfCell)
}

func formatDashboardSectorRow(rankSym string, r SectorResult) string {
	rankCell := formatCell(rankSym, 10, nil)
	nameCell := formatCell(r.Name, 24, nil)
	priceCell := formatCell(fmt.Sprintf("%.2f", r.Price), 12, nil)

	perfStr := fmt.Sprintf("%.2f%%", r.PercentChange)
	if r.PercentChange > 0 {
		perfStr = "+" + perfStr
	}

	var colorFn func(string) string
	var icon string
	if r.PercentChange > 0 {
		colorFn = green
		icon = "🟩"
	} else if r.PercentChange < 0 {
		colorFn = red
		icon = "🟥"
	} else {
		icon = "⬜"
	}

	perfCell := formatCell(fmt.Sprintf("%s %s", perfStr, icon), 15, colorFn)
	return fmt.Sprintf("%s%s%s%s", rankCell, nameCell, priceCell, perfCell)
}

func init() {
	RootCmd.AddCommand(dashboardCmd)
}
