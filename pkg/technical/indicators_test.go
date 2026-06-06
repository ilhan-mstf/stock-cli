package technical

import (
	"math"
	"testing"
)

func TestCalculateSMA(t *testing.T) {
	tests := []struct {
		name        string
		prices      []float64
		period      int
		expected    float64
		expectedErr bool
	}{
		{
			name:        "insufficient data",
			prices:      []float64{10.0, 12.0},
			period:      3,
			expectedErr: true,
		},
		{
			name:        "invalid period",
			prices:      []float64{10.0, 12.0},
			period:      0,
			expectedErr: true,
		},
		{
			name:        "valid SMA 3",
			prices:      []float64{10.0, 20.0, 30.0, 40.0, 50.0},
			period:      3,
			expected:    40.0, // (30+40+50)/3
			expectedErr: false,
		},
		{
			name:        "valid SMA 5",
			prices:      []float64{10.0, 20.0, 30.0, 40.0, 50.0},
			period:      5,
			expected:    30.0, // (10+20+30+40+50)/5
			expectedErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := CalculateSMA(tc.prices, tc.period)
			if tc.expectedErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if math.Abs(actual-tc.expected) > 1e-9 {
				t.Errorf("CalculateSMA() = %.2f; want %.2f", actual, tc.expected)
			}
		})
	}
}

func TestCalculateRSI(t *testing.T) {
	// Construct a rising series (should yield RSI close to 100)
	rising := make([]float64, 30)
	for i := 0; i < 30; i++ {
		rising[i] = 10.0 + float64(i)*1.5
	}

	// Construct a declining series (should yield RSI close to 0)
	declining := make([]float64, 30)
	for i := 0; i < 30; i++ {
		declining[i] = 100.0 - float64(i)*2.0
	}

	// Construct a flat series (should yield RSI of 50)
	flat := make([]float64, 30)
	for i := 0; i < 30; i++ {
		flat[i] = 25.0
	}

	tests := []struct {
		name        string
		prices      []float64
		period      int
		expectedMin float64
		expectedMax float64
		expectedErr bool
	}{
		{
			name:        "insufficient data",
			prices:      []float64{10.0, 11.0, 12.0},
			period:      14,
			expectedErr: true,
		},
		{
			name:        "invalid period",
			prices:      []float64{10.0, 11.0, 12.0},
			period:      -5,
			expectedErr: true,
		},
		{
			name:        "rising prices (overbought)",
			prices:      rising,
			period:      14,
			expectedMin: 95.0,
			expectedMax: 100.0,
			expectedErr: false,
		},
		{
			name:        "declining prices (oversold)",
			prices:      declining,
			period:      14,
			expectedMin: 0.0,
			expectedMax: 5.0,
			expectedErr: false,
		},
		{
			name:        "flat prices (neutral)",
			prices:      flat,
			period:      14,
			expectedMin: 49.9,
			expectedMax: 50.1,
			expectedErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := CalculateRSI(tc.prices, tc.period)
			if tc.expectedErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if actual < tc.expectedMin || actual > tc.expectedMax {
				t.Errorf("CalculateRSI() = %.2f; want between %.2f and %.2f", actual, tc.expectedMin, tc.expectedMax)
			}
		})
	}
}
