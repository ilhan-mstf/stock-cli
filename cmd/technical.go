package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"stock/pkg/config"
	"stock/pkg/finnhub"
	"stock/pkg/technical"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var technicalCmd = &cobra.Command{
	Use:   "technical [symbol]",
	Short: "Scan technical indicators (SMA-50/200, RSI-14, Volume)",
	Long:  `Queries historical daily price candles for the last 300 days, calculating SMA-50, SMA-200, 14-day RSI, and current volume breakout.`,
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

		to := time.Now().Unix()
		from := time.Now().AddDate(0, 0, -300).Unix()

		client := finnhub.NewClient(apiKey, nil)
		candles, err := client.FetchCandles(context.Background(), symbol, "D", from, to)
		if err != nil {
			return fmt.Errorf("failed to fetch historical candle data: %w", err)
		}

		closes := candles.Close
		sma50, err := technical.CalculateSMA(closes, 50)
		if err != nil {
			return fmt.Errorf("failed to calculate SMA-50: %w", err)
		}

		sma200, err := technical.CalculateSMA(closes, 200)
		if err != nil {
			return fmt.Errorf("failed to calculate SMA-200: %w", err)
		}

		rsi14, err := technical.CalculateRSI(closes, 14)
		if err != nil {
			return fmt.Errorf("failed to calculate RSI-14: %w", err)
		}

		currentPrice := closes[len(closes)-1]

		// Volume metrics
		volLen := len(candles.Volume)
		currentVolume := float64(candles.Volume[volLen-1])
		var avgVol20 float64
		if volLen >= 20 {
			sumVol := 0.0
			for i := volLen - 20; i < volLen; i++ {
				sumVol += float64(candles.Volume[i])
			}
			avgVol20 = sumVol / 20.0
		} else {
			sumVol := 0.0
			for i := 0; i < volLen; i++ {
				sumVol += float64(candles.Volume[i])
			}
			avgVol20 = sumVol / float64(volLen)
		}

		breakoutFactor := currentVolume / avgVol20

		if JSONOutput {
			out := struct {
				Symbol       string  `json:"symbol"`
				CurrentPrice float64 `json:"current_price"`
				SMA50        float64 `json:"sma50"`
				SMA200       float64 `json:"sma200"`
				RSI14        float64 `json:"rsi14"`
				Volume       float64 `json:"volume"`
				AvgVolume20  float64 `json:"avg_volume_20"`
				VolumeFactor float64 `json:"volume_factor"`
			}{
				Symbol:       symbol,
				CurrentPrice: currentPrice,
				SMA50:        sma50,
				SMA200:       sma200,
				RSI14:        rsi14,
				Volume:       currentVolume,
				AvgVolume20:  avgVol20,
				VolumeFactor: breakoutFactor,
			}
			data, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		// Renders console scorecard
		fmt.Println(bold(fmt.Sprintf("\n=========================================\n       %s - TECHNICAL SCANNER          \n=========================================", symbol)))

		// 1. Moving averages
		sma50Diff := ((currentPrice - sma50) / sma50) * 100
		sma200Diff := ((currentPrice - sma200) / sma200) * 100

		var sma50Msg string
		if sma50Diff > 0 {
			sma50Msg = green(fmt.Sprintf("(Above SMA-50: +%.2f%% 🟩)", sma50Diff))
		} else {
			sma50Msg = red(fmt.Sprintf("(Below SMA-50: %.2f%% 🟥)", sma50Diff))
		}

		var sma200Msg string
		if sma200Diff > 0 {
			sma200Msg = green(fmt.Sprintf("(Above SMA-200: +%.2f%% 🟩)", sma200Diff))
		} else {
			sma200Msg = red(fmt.Sprintf("(Below SMA-200: %.2f%% 🟥)", sma200Diff))
		}

		fmt.Println(bold("\nTREND LINES (SIMPLE MOVING AVERAGES):"))
		fmt.Printf("  Current Price:             %.2f\n", currentPrice)
		fmt.Printf("  50-day Moving Average:     %.2f  %s\n", sma50, sma50Msg)
		fmt.Printf("  200-day Moving Average:    %.2f  %s\n", sma200, sma200Msg)

		// 2. RSI
		var rsiStatus string
		if rsi14 > 70 {
			rsiStatus = red(fmt.Sprintf("%.2f  (Overbought - Sell Alert 🟥)", rsi14))
		} else if rsi14 < 30 {
			rsiStatus = green(fmt.Sprintf("%.2f  (Oversold - Buy Opportunity 🟩)", rsi14))
		} else {
			rsiStatus = fmt.Sprintf("%.2f  (Neutral)", rsi14)
		}

		fmt.Println(bold("\nMOMENTUM OSCILLATOR:"))
		fmt.Printf("  14-day RSI:                %s\n", rsiStatus)

		// 3. Volume Breakdown
		var volBreakoutMsg string
		if breakoutFactor >= 1.5 {
			volBreakoutMsg = green(fmt.Sprintf("%.2fx  (High Volume Breakout 🟩)", breakoutFactor))
		} else if breakoutFactor <= 0.5 {
			volBreakoutMsg = red(fmt.Sprintf("%.2fx  (Low Volume)", breakoutFactor))
		} else {
			volBreakoutMsg = fmt.Sprintf("%.2fx  (Normal)", breakoutFactor)
		}

		fmt.Println(bold("\nMARKET LIQUIDITY & MECHANICS:"))
		fmt.Printf("  Current Volume:            %.0f shares\n", currentVolume)
		fmt.Printf("  20-day Average Volume:     %.0f shares\n", avgVol20)
		fmt.Printf("  Volume Breakout Factor:    %s\n", volBreakoutMsg)
		fmt.Println()

		return nil
	},
}

func init() {
	RootCmd.AddCommand(technicalCmd)
}
