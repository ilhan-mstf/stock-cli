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

var insiderCmd = &cobra.Command{
	Use:   "insider [symbol]",
	Short: "Scan executive and director insider transactions",
	Long:  `Queries recent insider buying and selling transactions from Finnhub, summarizing net accumulation vs distribution.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol := strings.ToUpper(strings.TrimSpace(args[0]))

		cfg, err := config.Load(CfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		apiKey, err := getAPIKey(os.Getenv("FINNHUB_API_KEY"), cfg.APIKey)
		if err != nil {
			return err
		}

		client := finnhub.NewClient(apiKey, nil)
		transactions, err := client.FetchInsiderTransactions(context.Background(), symbol)
		if err != nil {
			return err
		}

		// Calculate summaries
		var totalBought int64
		var totalSold int64
		var netBalance int64

		for _, tx := range transactions {
			netBalance += tx.Change
			if tx.Change > 0 {
				totalBought += tx.Change
			} else {
				totalSold += tx.Change // negative value
			}
		}

		if JSONOutput {
			data, err := json.MarshalIndent(transactions, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(bold(fmt.Sprintf("\n=========================================\n       %s - INSIDER ACTIVITY           \n=========================================", symbol)))

		// Render summaries
		var balanceMsg string
		if netBalance > 0 {
			balanceMsg = green(fmt.Sprintf("+%s shares 🟩 (Net Accumulation)", formatInt(netBalance)))
		} else if netBalance < 0 {
			balanceMsg = red(fmt.Sprintf("%s shares 🟥 (Net Distribution)", formatInt(netBalance)))
		} else {
			balanceMsg = "0 shares (No net activity)"
		}

		fmt.Println(bold("\n6-MONTH INSIDER SUMMARY:"))
		fmt.Printf("  Total Buying:              +%s shares\n", formatInt(totalBought))
		fmt.Printf("  Total Selling:             %s shares\n", formatInt(totalSold))
		fmt.Printf("  Net Insider Balance:       %s\n", balanceMsg)

		// Renders recent log (up to 10 entries)
		fmt.Println(bold("\nRECENT TRANSACTION LOG:"))
		fmt.Println(bold(fmt.Sprintf("%s%s%s%s%s",
			formatCell("DATE", 12, nil),
			formatCell("INSIDER NAME", 22, nil),
			formatCell("CHANGE", 12, nil),
			formatCell("PRICE", 10, nil),
			formatCell("VALUE ($)", 16, nil),
		)))
		fmt.Println(strings.Repeat("-", 72))

		displayCount := 10
		if len(transactions) < displayCount {
			displayCount = len(transactions)
		}

		if displayCount == 0 {
			fmt.Println("  No recent insider transactions recorded.")
		} else {
			for i := 0; i < displayCount; i++ {
				tx := transactions[i]
				fmt.Println(formatInsiderRow(tx))
			}
		}
		fmt.Println()

		return nil
	},
}

func formatInt(n int64) string {
	// Simple comma formatting for large numbers
	in := fmt.Sprintf("%d", n)
	var out []rune
	// handle sign
	var sign string
	if strings.HasPrefix(in, "-") {
		sign = "-"
		in = in[1:]
	}
	runes := []rune(in)
	for i, r := range runes {
		if i > 0 && (len(runes)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, r)
	}
	return sign + string(out)
}

func formatInsiderRow(tx finnhub.InsiderTransaction) string {
	dateCell := formatCell(tx.TransactionDate, 12, nil)

	// Trim name if too long
	name := tx.Name
	if len(name) > 20 {
		name = name[:17] + "..."
	}
	nameCell := formatCell(name, 22, nil)

	changeStr := fmt.Sprintf("%d", tx.Change)
	if tx.Change > 0 {
		changeStr = "+" + changeStr
	}
	var colorFn func(string) string
	if tx.Change > 0 {
		colorFn = green
	} else if tx.Change < 0 {
		colorFn = red
	}
	changeCell := formatCell(changeStr, 12, colorFn)

	priceCell := formatCell(fmt.Sprintf("%.2f", tx.Price), 10, nil)

	txVal := float64(tx.Change) * tx.Price
	valStr := fmt.Sprintf("%.0f", txVal)
	if txVal > 0 {
		valStr = "+" + valStr
	}
	valCell := formatCell(formatStringComma(valStr), 16, colorFn)

	return fmt.Sprintf("%s%s%s%s%s", dateCell, nameCell, changeCell, priceCell, valCell)
}

func formatStringComma(s string) string {
	var sign string
	if strings.HasPrefix(s, "-") {
		sign = "-"
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		sign = "+"
		s = s[1:]
	}
	runes := []rune(s)
	var out []rune
	for i, r := range runes {
		if i > 0 && (len(runes)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, r)
	}
	return sign + string(out)
}

func init() {
	RootCmd.AddCommand(insiderCmd)
}
