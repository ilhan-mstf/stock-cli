package cmd

import (
	"encoding/json"
	"fmt"
	"stock/pkg/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `Get or set CLI configuration parameters including the Finnhub API Key.`,
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Display current configuration parameters",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(CfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		maskedKey := config.MaskAPIKey(cfg.APIKey)

		if JSONOutput {
			out := struct {
				APIKey    string   `json:"api_key"`
				Watchlist []string `json:"watchlist"`
			}{
				APIKey:    maskedKey,
				Watchlist: cfg.Watchlist,
			}
			data, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("%s: %s\n", bold("api_key"), maskedKey)
		fmt.Printf("%s: %v\n", bold("watchlist"), cfg.Watchlist)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a configuration parameter",
}

var configSetAPIKeyCmd = &cobra.Command{
	Use:   "api-key [key]",
	Short: "Set the Finnhub API Key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(CfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		cfg.APIKey = args[0]
		if err := config.Save(CfgFile, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		if JSONOutput {
			out := map[string]string{"status": "success", "message": "API key updated successfully"}
			data, _ := json.Marshal(out)
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(green("API key saved successfully."))
		return nil
	},
}

func init() {
	configSetCmd.AddCommand(configSetAPIKeyCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	RootCmd.AddCommand(configCmd)
}
