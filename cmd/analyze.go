package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"stock-cli/pkg/config"
	"stock-cli/pkg/finnhub"
	"strings"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [symbol]",
	Short: "Perform deep fundamental analysis on a stock symbol",
	Long:  `Queries key valuation multiples, profitability margins, growth rates, capital returns (ROIC/ROE), and leverage ratios from Finnhub basic financials.`,
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
		financials, err := client.FetchBasicFinancials(context.Background(), symbol)
		if err != nil {
			return err
		}

		if JSONOutput {
			data, err := json.MarshalIndent(financials, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		m := financials.Metric

		fmt.Println(bold(fmt.Sprintf("\n=========================================\n       %s - FUNDAMENTAL SCORECARD      \n=========================================", symbol)))

		// 1. Valuation
		peNorm := formatMetricValue(m, "peNormalizedTTM", "")
		peExcl := formatMetricValue(m, "peBasicExclExtraTTM", "")
		psVal := formatMetricValue(m, "psTTM", "")
		pegVal := formatMetricWithColor(m, "pegTTM", "", func(val float64) func(string) string {
			if val > 0 && val < 1.5 {
				return green
			} else if val >= 3.0 {
				return red
			}
			return nil
		})

		fmt.Println(bold("\nVALUATION MULTIPLES:"))
		fmt.Printf("  P/E Ratio (Normalized):    %s\n", peNorm)
		fmt.Printf("  P/E Ratio (Basic Excl):    %s\n", peExcl)
		fmt.Printf("  Price / Sales (TTM):       %s\n", psVal)
		fmt.Printf("  PEG Ratio (TTM):           %s\n", pegVal)

		// 2. Returns & Margins
		roicVal := formatMetricWithColor(m, "roicTTM", "%", func(val float64) func(string) string {
			if val >= 15.0 {
				return green
			} else if val < 5.0 {
				return red
			}
			return nil
		})
		roeVal := formatMetricWithColor(m, "roeTTM", "%", func(val float64) func(string) string {
			if val >= 15.0 {
				return green
			} else if val < 5.0 {
				return red
			}
			return nil
		})
		grossMargin := formatMetricValue(m, "grossMarginTTM", "%")
		opMargin := formatMetricValue(m, "operatingMarginTTM", "%")

		fmt.Println(bold("\nQUALITY & EFFICIENCY METRICS:"))
		fmt.Printf("  Return on Capital (ROIC):  %s\n", roicVal)
		fmt.Printf("  Return on Equity (ROE):    %s\n", roeVal)
		fmt.Printf("  Gross Margin (TTM):        %s\n", grossMargin)
		fmt.Printf("  Operating Margin (TTM):    %s\n", opMargin)

		// 3. Growth YoY
		revGrowth := formatMetricWithColor(m, "revenueGrowthTTMYoy", "%", func(val float64) func(string) string {
			if val > 15.0 {
				return green
			} else if val < 0 {
				return red
			}
			return nil
		})
		epsGrowth := formatMetricWithColor(m, "epsGrowthTTMYoy", "%", func(val float64) func(string) string {
			if val > 15.0 {
				return green
			} else if val < 0 {
				return red
			}
			return nil
		})

		fmt.Println(bold("\nGROWTH RATES (YoY):"))
		fmt.Printf("  Revenue Growth (TTM):      %s\n", revGrowth)
		fmt.Printf("  EPS Growth (TTM):          %s\n", epsGrowth)

		// 4. Balance Sheet Health
		deVal := formatMetricWithColor(m, "totalDebt/totalEquityTTM", "", func(val float64) func(string) string {
			if val > 0 && val < 1.0 {
				return green
			} else if val >= 2.5 {
				return red
			}
			return nil
		})
		fcfYield := formatMetricWithColor(m, "freeCashFlowYieldTTM", "%", func(val float64) func(string) string {
			if val >= 8.0 {
				return green
			}
			return nil
		})

		fmt.Println(bold("\nFINANCIAL HEALTH & LIQUIDITY:"))
		fmt.Printf("  Debt / Equity Ratio:       %s\n", deVal)
		fmt.Printf("  Free Cash Flow Yield:      %s\n", fcfYield)
		fmt.Println()

		return nil
	},
}

func getFloatMetric(m map[string]interface{}, key string) (float64, bool) {
	val, exists := m[key]
	if !exists || val == nil {
		return 0, false
	}
	floatVal, ok := val.(float64)
	return floatVal, ok
}

func formatMetricValue(m map[string]interface{}, key string, suffix string) string {
	val, ok := getFloatMetric(m, key)
	if !ok {
		return "N/A"
	}
	return fmt.Sprintf("%.2f%s", val, suffix)
}

func formatMetricWithColor(m map[string]interface{}, key string, suffix string, colorLogic func(float64) func(string) string) string {
	val, ok := getFloatMetric(m, key)
	if !ok {
		return "N/A"
	}

	formatted := fmt.Sprintf("%.2f%s", val, suffix)
	if colorLogic != nil {
		colorFn := colorLogic(val)
		if colorFn != nil {
			return colorFn(formatted)
		}
	}
	return formatted
}

func init() {
	RootCmd.AddCommand(analyzeCmd)
}
