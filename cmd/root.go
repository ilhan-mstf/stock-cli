package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	CfgFile    string
	JSONOutput bool
	NoColor    bool
)

var RootCmd = &cobra.Command{
	Use:   "stock",
	Short: "Stock Market CLI - Query stock quotes and manage your watchlist",
	Long: `A fast and lightweight command-line interface to get real-time stock quotes 
and manage a personal watchlist using the Finnhub API.`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&CfgFile, "config", "", "config file path (default is $HOME/.config/stock-cli/config.json)")
	RootCmd.PersistentFlags().BoolVar(&JSONOutput, "json", false, "output in machine-readable JSON format")
	RootCmd.PersistentFlags().BoolVar(&NoColor, "no-color", false, "disable colored terminal output")
}

// Helpers used across commands

func coloredString(text string, colorCode string) string {
	if NoColor {
		return text
	}
	return fmt.Sprintf("\033[%sm%s\033[0m", colorCode, text)
}

func green(text string) string  { return coloredString(text, "32") }
func red(text string) string    { return coloredString(text, "31") }
func yellow(text string) string { return coloredString(text, "33") }
func bold(text string) string   { return coloredString(text, "1") }

func getAPIKey(envKey string, cfgKey string) (string, error) {
	if envKey != "" {
		return envKey, nil
	}
	if cfgKey != "" {
		return cfgKey, nil
	}
	return "", fmt.Errorf("Finnhub API key is not configured.\n" +
		"Please set the FINNHUB_API_KEY environment variable or run:\n" +
		"  stock config set api-key <key>")
}
