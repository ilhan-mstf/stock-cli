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

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Manage and view your stock watchlist",
	Long: `Query real-time quotes for all symbols in your watchlist, 
or add, remove, and list symbols in the watchlist.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(CfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(cfg.Watchlist) == 0 {
			if JSONOutput {
				fmt.Println("[]")
				return nil
			}
			fmt.Println("Your watchlist is empty.")
			fmt.Println("Add symbols to it using:")
			fmt.Println("  stock watch add <symbol...>")
			return nil
		}

		apiKey, err := getAPIKey(os.Getenv("FINNHUB_API_KEY"), cfg.APIKey)
		if err != nil {
			return err
		}

		client := finnhub.NewClient(apiKey, nil)
		quotes, errs := client.FetchQuotesConcurrent(context.Background(), cfg.Watchlist)

		var validQuotes []finnhub.Quote
		var hasErrors bool

		for i, qErr := range errs {
			if qErr != nil {
				fmt.Fprintf(os.Stderr, "Error fetching %q: %v\n", cfg.Watchlist[i], qErr)
				hasErrors = true
			} else {
				validQuotes = append(validQuotes, quotes[i])
			}
		}

		if len(validQuotes) == 0 {
			return fmt.Errorf("no watchlist stock quotes retrieved successfully")
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

var watchAddCmd = &cobra.Command{
	Use:   "add [symbols...]",
	Short: "Add one or more symbols to the watchlist",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(CfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		var added []string
		for _, sym := range args {
			trimmed := strings.TrimSpace(sym)
			if trimmed == "" {
				if !JSONOutput {
					fmt.Println(red("Error: symbol cannot be empty or only whitespace"))
				}
				continue
			}
			symUpper := strings.ToUpper(trimmed)
			if cfg.AddSymbol(symUpper) {
				added = append(added, symUpper)
				if !JSONOutput {
					fmt.Printf("Symbol %s added to watchlist.\n", green(symUpper))
				}
			} else {
				if !JSONOutput {
					fmt.Printf("Symbol %s is already in the watchlist.\n", yellow(symUpper))
				}
			}
		}

		if len(added) > 0 {
			if err := config.Save(CfgFile, cfg); err != nil {
				return fmt.Errorf("failed to save watchlist: %w", err)
			}
		}

		if JSONOutput {
			out := struct {
				Status    string   `json:"status"`
				Added     []string `json:"added"`
				Watchlist []string `json:"watchlist"`
			}{
				Status:    "success",
				Added:     added,
				Watchlist: cfg.Watchlist,
			}
			data, _ := json.Marshal(out)
			fmt.Println(string(data))
		}

		return nil
	},
}

var watchRemoveCmd = &cobra.Command{
	Use:   "remove [symbols...]",
	Short: "Remove one or more symbols from the watchlist",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(CfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		var removed []string
		for _, sym := range args {
			trimmed := strings.TrimSpace(sym)
			if trimmed == "" {
				if !JSONOutput {
					fmt.Println(red("Error: symbol cannot be empty or only whitespace"))
				}
				continue
			}
			symUpper := strings.ToUpper(trimmed)
			if cfg.RemoveSymbol(symUpper) {
				removed = append(removed, symUpper)
				if !JSONOutput {
					fmt.Printf("Symbol %s removed from watchlist.\n", green(symUpper))
				}
			} else {
				if !JSONOutput {
					fmt.Printf("Symbol %s was not in the watchlist.\n", yellow(symUpper))
				}
			}
		}

		if len(removed) > 0 {
			if err := config.Save(CfgFile, cfg); err != nil {
				return fmt.Errorf("failed to save watchlist: %w", err)
			}
		}

		if JSONOutput {
			out := struct {
				Status    string   `json:"status"`
				Removed   []string `json:"removed"`
				Watchlist []string `json:"watchlist"`
			}{
				Status:    "success",
				Removed:   removed,
				Watchlist: cfg.Watchlist,
			}
			data, _ := json.Marshal(out)
			fmt.Println(string(data))
		}

		return nil
	},
}

var watchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all symbols in the watchlist",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(CfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if JSONOutput {
			data, _ := json.Marshal(cfg.Watchlist)
			fmt.Println(string(data))
			return nil
		}

		if len(cfg.Watchlist) == 0 {
			fmt.Println("Your watchlist is empty.")
			return nil
		}

		for _, sym := range cfg.Watchlist {
			fmt.Println(sym)
		}
		return nil
	},
}

func init() {
	watchCmd.AddCommand(watchAddCmd)
	watchCmd.AddCommand(watchRemoveCmd)
	watchCmd.AddCommand(watchListCmd)
	RootCmd.AddCommand(watchCmd)
}
