package technical

import (
	"fmt"
)

// CalculateSMA computes the Simple Moving Average (SMA) of the last 'period' prices.
func CalculateSMA(prices []float64, period int) (float64, error) {
	if len(prices) < period {
		return 0, fmt.Errorf("insufficient price data: got %d, want at least %d", len(prices), period)
	}
	if period <= 0 {
		return 0, fmt.Errorf("period must be greater than zero")
	}

	sum := 0.0
	startIndex := len(prices) - period
	for i := startIndex; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period), nil
}

// CalculateRSI computes the Wilder's Relative Strength Index (RSI) of the last 'period' prices.
func CalculateRSI(prices []float64, period int) (float64, error) {
	if len(prices) < period+1 {
		return 0, fmt.Errorf("insufficient price data for RSI: got %d, want at least %d", len(prices), period+1)
	}
	if period <= 0 {
		return 0, fmt.Errorf("period must be greater than zero")
	}

	// Calculate changes between successive prices
	changes := make([]float64, len(prices)-1)
	for i := 0; i < len(prices)-1; i++ {
		changes[i] = prices[i+1] - prices[i]
	}

	// Calculate initial average gain/loss over the first 'period' changes
	avgGain := 0.0
	avgLoss := 0.0
	for i := 0; i < period; i++ {
		change := changes[i]
		if change > 0 {
			avgGain += change
		} else {
			avgLoss -= change
		}
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Apply Wilder's smoothing technique for the remaining changes
	for i := period; i < len(changes); i++ {
		change := changes[i]
		currentGain := 0.0
		currentLoss := 0.0
		if change > 0 {
			currentGain = change
		} else {
			currentLoss = -change
		}

		avgGain = (avgGain*float64(period-1) + currentGain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + currentLoss) / float64(period)
	}

	if avgLoss == 0 {
		if avgGain > 0 {
			return 100.0, nil
		}
		return 50.0, nil // No movement
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))
	return rsi, nil
}
