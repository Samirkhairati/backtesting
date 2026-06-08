package portfolio_test

import (
	"testing"
	"time"

	"github.com/samirkhairati/backtesting/core/portfolio"
)

func TestPortfolioCash(t *testing.T) {
	pf := portfolio.New()

	pf.Deposit(100000)
	if pf.Cash != 100000 {
		t.Errorf("Expected cash to be 100000, got %f", pf.Cash)
	}

	err := pf.Withdraw(10000)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if pf.Cash != 90000 {
		t.Errorf("Expected cash to be 90000, got %f", pf.Cash)
	}

	err = pf.Withdraw(100000)
	if err != portfolio.ErrInsufficientCash {
		t.Errorf("Expected ErrInsufficientCash, got %v", err)
	}
}

func TestInsufficientCash(t *testing.T) {
	pf := portfolio.New()
	pf.Deposit(20000)

	err := pf.AddPosition("AAPL", 100, 100, time.Now())
	if err != nil {
		t.Fatalf("Unexpected error adding position: %v", err)
	}

	if pf.Cash != 10000 {
		t.Errorf("Expected cash to be 10000 after purchase, got %f", pf.Cash)
	}

	// Attempt to buy more than cash allows
	err = pf.AddPosition("AAPL", 200, 100, time.Now())
	if err != portfolio.ErrInsufficientCash {
		t.Errorf("Expected ErrInsufficientCash, got %v", err)
	}
}

func TestPortfolioMarketValue(t *testing.T) {
	pf := portfolio.New()
	pf.Deposit(100000)

	// Buy 100 shares of AAPL at $100
	err := pf.AddPosition("AAPL", 100, 100, time.Now())
	if err != nil {
		t.Fatalf("Failed to add position: %v", err)
	}

	// Cash should now be 90000
	if pf.Cash != 90000 {
		t.Errorf("Expected cash 90000, got %f", pf.Cash)
	}

	// At this exact moment, if market price is unknown, it falls back to AveragePrice (100)
	// So TotalValue should still be 100000 (90000 cash + 10000 stock)
	if pf.TotalValue() != 100000 {
		t.Errorf("Expected total value 100000, got %f", pf.TotalValue())
	}

	// Now the market updates AAPL to $150
	pf.UpdateMarketPrice("AAPL", 150)

	// Total value should now be 90000 + (100 * 150) = 105000
	if pf.TotalValue() != 105000 {
		t.Errorf("Expected total value 105000, got %f", pf.TotalValue())
	}
}

func TestPortfolioEquityCurve(t *testing.T) {
	pf := portfolio.New()
	pf.Deposit(100000)

	// Day 1
	ts1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	pf.AddPosition("AAPL", 100, 150, ts1) // Cash = 85000
	pf.UpdateMarketPrice("AAPL", 150)
	pf.RecordEquity(ts1)

	// Day 2
	ts2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	pf.UpdateMarketPrice("AAPL", 160) // AAPL goes up
	pf.RecordEquity(ts2)

	if len(pf.EquityCurve) != 2 {
		t.Fatalf("Expected 2 equity points, got %d", len(pf.EquityCurve))
	}

	if pf.EquityCurve[0].Value != 100000 {
		t.Errorf("Day 1 equity expected 100000, got %f", pf.EquityCurve[0].Value)
	}
	if pf.EquityCurve[1].Value != 101000 { // 85000 + (100 * 160)
		t.Errorf("Day 2 equity expected 101000, got %f", pf.EquityCurve[1].Value)
	}
}

func TestAddPositionAveraging(t *testing.T) {
	pf := portfolio.New()
	pf.Deposit(100000)

	pf.AddPosition("AAPL", 100, 100, time.Now()) // 100 shares @ $100
	pf.AddPosition("AAPL", 100, 150, time.Now()) // 100 shares @ $150

	pos := pf.Positions["AAPL"]
	if pos.Quantity != 200 {
		t.Errorf("Expected 200 shares, got %f", pos.Quantity)
	}
	if pos.AveragePrice != 125 {
		t.Errorf("Expected average price 125, got %f", pos.AveragePrice)
	}

	// Sell all shares
	pf.AddPosition("AAPL", -200, 200, time.Now()) // Sell 200 shares @ $200

	if _, exists := pf.Positions["AAPL"]; exists {
		t.Errorf("Position should be removed after selling all shares")
	}

	// Initial 100000
	// Buy 1: -10000 -> 90000
	// Buy 2: -15000 -> 75000
	// Sell: +40000 -> 115000
	if pf.Cash != 115000 {
		t.Errorf("Expected cash to be 115000, got %f", pf.Cash)
	}
}

func TestShortSellingPnL(t *testing.T) {
	pf := portfolio.New()
	pf.Deposit(100000)

	// Short 100 shares at $150
	err := pf.AddPosition("AAPL", -100, 150, time.Now())
	if err != nil {
		t.Fatalf("Failed to open short: %v", err)
	}

	// Cash should be 100,000 + 15,000 = 115,000
	if pf.Cash != 115000 {
		t.Errorf("Expected cash 115000, got %f", pf.Cash)
	}

	// Market drops to $100
	pf.UpdateMarketPrice("AAPL", 100)
	// TotalValue = Cash (115000) + (-100 * 100) = 115000 - 10000 = 105000
	if pf.TotalValue() != 105000 {
		t.Errorf("Expected TotalValue 105000, got %f", pf.TotalValue())
	}

	// Buy to cover at $100
	err = pf.AddPosition("AAPL", 100, 100, time.Now())
	if err != nil {
		t.Fatalf("Failed to cover short: %v", err)
	}

	// Cash should be 115,000 - 10,000 = 105,000
	if pf.Cash != 105000 {
		t.Errorf("Expected cash 105000, got %f", pf.Cash)
	}

	// Check PnL on the closed trade
	if len(pf.ClosedTrades) != 1 {
		t.Fatalf("Expected 1 closed trade, got %d", len(pf.ClosedTrades))
	}
	
	tr := pf.ClosedTrades[0]
	if tr.PnL != 5000 { // (150 - 100) * 100
		t.Errorf("Expected PnL 5000, got %f", tr.PnL)
	}
}

func TestShortSellingMargin(t *testing.T) {
	pf := portfolio.New()
	pf.Deposit(1000) // Small account

	// Try to short $15,000 worth of AAPL, which exceeds TotalValue
	err := pf.AddPosition("AAPL", -100, 150, time.Now())
	if err != portfolio.ErrInsufficientMargin {
		t.Fatalf("Expected ErrInsufficientMargin, got %v", err)
	}
	
	// Try to short $500 worth of AAPL, which should succeed
	err = pf.AddPosition("AAPL", -5, 100, time.Now())
	if err != nil {
		t.Fatalf("Failed to open valid short: %v", err)
	}
}

func TestCloseShortPosition(t *testing.T) {
	pf := portfolio.New()
	pf.Deposit(100000)

	pf.AddPosition("AAPL", -100, 150, time.Now())
	pos, exists := pf.Positions["AAPL"]
	if !exists || pos.Quantity != -100 {
		t.Fatalf("Expected short position of -100")
	}

	// Close the short position by buying 100
	pf.AddPosition("AAPL", 100, 140, time.Now())
	if _, exists := pf.Positions["AAPL"]; exists {
		t.Fatalf("Position should be removed after covering")
	}
}

