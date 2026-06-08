package strategy_test

import (
	"testing"

	"github.com/samirkhairati/backtesting/core/strategy"
)

func TestSMA(t *testing.T) {
	sma := strategy.NewSMA(3)

	// Feed 1: 10
	val := sma.Update(10)
	if val != 0 {
		t.Errorf("Expected 0 before buffer is full, got %f", val)
	}

	// Feed 2: 20
	val = sma.Update(20)
	if val != 0 {
		t.Errorf("Expected 0 before buffer is full, got %f", val)
	}

	// Feed 3: 30 (Buffer Full)
	// (10 + 20 + 30) / 3 = 20
	val = sma.Update(30)
	if val != 20 {
		t.Errorf("Expected 20, got %f", val)
	}

	// Feed 4: 40 (10 drops out)
	// (20 + 30 + 40) / 3 = 30
	val = sma.Update(40)
	if val != 30 {
		t.Errorf("Expected 30, got %f", val)
	}

	// Feed 5: 50 (20 drops out)
	// (30 + 40 + 50) / 3 = 40
	val = sma.Update(50)
	if val != 40 {
		t.Errorf("Expected 40, got %f", val)
	}

	// Test StdDev
	// Window is [30, 40, 50]. Mean is 40.
	// Variance = ((30-40)^2 + (40-40)^2 + (50-40)^2) / 3 = (100 + 0 + 100) / 3 = 200/3 = 66.666...
	// StdDev = sqrt(66.666...) ~ 8.1649658
	stdDev := sma.StdDev()
	if stdDev < 8.16 || stdDev > 8.17 {
		t.Errorf("Expected StdDev ~8.165, got %f", stdDev)
	}
}

func TestEMA(t *testing.T) {
	ema := strategy.NewEMA(3) // multiplier = 2 / 4 = 0.5
	
	val := ema.Update(10) // wait for more data
	if val != 0 {
		t.Errorf("Expected 0 initially, got %f", val)
	}
	
	val = ema.Update(20) // (20-10)*0.5 + 10 = 15
	if val != 0 {
		t.Errorf("Expected 0 until period met, got %f", val)
	}

	val = ema.Update(30) // (30-15)*0.5 + 15 = 22.5
	if val != 22.5 {
		t.Errorf("Expected 22.5, got %f", val)
	}
}

func TestRSI(t *testing.T) {
	rsi := strategy.NewRSI(2)

	// Price 10
	val := rsi.Update(10)
	if val != 0 { t.Errorf("Expected 0, got %f", val) }

	// Price 20 (Gain 10)
	val = rsi.Update(20)
	if val != 0 { t.Errorf("Expected 0, got %f", val) }

	// Price 15 (Loss 5)
	// Initial SMA: avgGain = (10+0)/2 = 5, avgLoss = (0+5)/2 = 2.5
	// RS = 5 / 2.5 = 2. RSI = 100 - (100/3) = 66.66...
	val = rsi.Update(15)
	if val < 66.66 || val > 66.67 {
		t.Errorf("Expected RSI ~66.66, got %f", val)
	}
}
