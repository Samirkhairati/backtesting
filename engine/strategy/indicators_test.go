package strategy_test

import (
	"testing"

	"github.com/samirkhairati/backtesting/engine/strategy"
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
}
