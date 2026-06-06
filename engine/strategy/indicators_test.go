package strategy_test

import (
	"testing"

	"github.com/samirkhairati/backtesting/engine/strategy"
)

func TestSMA(t *testing.T) {
	prices := []float64{10, 20, 30, 40, 50}

	// SMA period 3: (30 + 40 + 50) / 3 = 120 / 3 = 40
	sma := strategy.SMA(prices, 3)
	if sma != 40 {
		t.Errorf("Expected SMA 40, got %f", sma)
	}

	// SMA period 5: (10 + 20 + 30 + 40 + 50) / 5 = 150 / 5 = 30
	sma = strategy.SMA(prices, 5)
	if sma != 30 {
		t.Errorf("Expected SMA 30, got %f", sma)
	}

	// Not enough data
	sma = strategy.SMA(prices, 10)
	if sma != 0 {
		t.Errorf("Expected 0 for insufficient data, got %f", sma)
	}

	// Invalid period
	sma = strategy.SMA(prices, 0)
	if sma != 0 {
		t.Errorf("Expected 0 for period 0, got %f", sma)
	}
}
