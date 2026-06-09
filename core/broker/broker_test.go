package broker_test

import (
	"testing"
	"time"

	"github.com/samirkhairati/backtesting/core/broker"
	"github.com/samirkhairati/backtesting/core/data"
	"github.com/samirkhairati/backtesting/core/portfolio"
)

func TestSubmitOrder(t *testing.T) {
	pf := portfolio.New()
	bm := broker.NewOrderManager(pf)

	now := time.Now()
	bm.SetTime(now)

	bm.SubmitOrder("AAPL", 100)
	bm.SubmitOrder("TSLA", -50)

	if len(bm.PendingOrders) != 2 {
		t.Fatalf("Expected 2 pending orders, got %d", len(bm.PendingOrders))
	}

	if bm.PendingOrders[0].Symbol != "AAPL" || bm.PendingOrders[0].Quantity != 100 {
		t.Errorf("Order 0 mismatch")
	}
	if bm.PendingOrders[1].Symbol != "TSLA" || bm.PendingOrders[1].Quantity != -50 {
		t.Errorf("Order 1 mismatch")
	}

	if bm.PendingOrders[0].CreatedAt != now {
		t.Errorf("Timestamp mismatch")
	}
}

func TestClosePosition(t *testing.T) {
	pf := portfolio.New()
	pf.Deposit(100000)
	// Force a raw position into the portfolio
	pf.AddPosition("MSFT", 25, 200, time.Now())

	bm := broker.NewOrderManager(pf)
	bm.SetTime(time.Now())

	// Trying to close a position that doesn't exist should silently succeed without creating an order
	err := bm.ClosePosition("FAKE")
	if err != nil {
		t.Errorf("Unexpected error closing fake position: %v", err)
	}
	if len(bm.PendingOrders) != 0 {
		t.Errorf("Expected 0 pending orders, got %d", len(bm.PendingOrders))
	}

	// Close real position
	err = bm.ClosePosition("MSFT")
	if err != nil {
		t.Errorf("Unexpected error closing real position: %v", err)
	}

	if len(bm.PendingOrders) != 1 {
		t.Fatalf("Expected 1 pending order, got %d", len(bm.PendingOrders))
	}

	sellOrder := bm.PendingOrders[0]
	if sellOrder.Symbol != "MSFT" || sellOrder.Quantity != -25 {
		t.Errorf("Expected to sell 25 MSFT, got %+v", sellOrder)
	}
}

func TestGetPositionImmutability(t *testing.T) {
	pf := portfolio.New()
	pf.Deposit(10000)
	pf.AddPosition("AMD", 10, 100, time.Now())

	bm := broker.NewOrderManager(pf)

	pos := bm.GetPosition("AMD")
	if pos == nil || pos.Quantity != 10 {
		t.Fatalf("Failed to retrieve position")
	}

	// Attempt to mutate the returned position directly
	pos.Quantity = 99999

	// Verify the true portfolio state was completely unaffected by the mutation attempt
	truePos := pf.Positions["AMD"]
	if truePos.Quantity != 10 {
		t.Errorf("SECURITY BREACH: Strategy successfully mutated the real portfolio state. Expected 10, got %f", truePos.Quantity)
	}
}

func TestExecutionEngine(t *testing.T) {
	pf := portfolio.New()
	pf.Deposit(1000)

	bm := broker.NewOrderManager(pf)
	bm.SetTime(time.Now())

	// Submit 2 orders: one valid, one that will be rejected due to insufficient cash
	bm.SubmitOrder("AAPL", 5)  // Should cost ~ $750 if Open is $150
	bm.SubmitOrder("TSLA", 10) // Should cost ~$2000, will fail

	if len(bm.PendingOrders) != 2 {
		t.Fatalf("Expected 2 pending orders")
	}

	bar1 := data.Bar{
		Symbol: "AAPL",
		Open:   150.0,
		Close:  155.0,
	}

	// Execution should process the AAPL order, and IGNORE the TSLA order (different symbol)
	bm.ExecutePendingOrders(bar1)

	if len(bm.PendingOrders) != 1 {
		t.Fatalf("Expected 1 pending order left (TSLA), got %d", len(bm.PendingOrders))
	}

	// Verify AAPL was filled
	if len(bm.Fills) != 1 {
		t.Fatalf("Expected 1 fill event, got %d", len(bm.Fills))
	}
	if bm.Fills[0].Symbol != "AAPL" || bm.Fills[0].Price != 150.0 {
		t.Errorf("Fill event incorrect: %+v", bm.Fills[0])
	}
	if pf.Cash != 250.0 { // 1000 - (5 * 150)
		t.Errorf("Expected cash to be 250, got %f", pf.Cash)
	}

	// Now send TSLA bar
	bar2 := data.Bar{
		Symbol: "TSLA",
		Open:   200.0,
	}

	// Execution should try to fill TSLA but REJECT it due to insufficient cash ($250 < $2000)
	bm.ExecutePendingOrders(bar2)

	if len(bm.PendingOrders) != 0 {
		t.Fatalf("Expected 0 pending orders (TSLA dropped), got %d", len(bm.PendingOrders))
	}
	if len(bm.Fills) != 1 { // Still 1, TSLA was rejected
		t.Fatalf("Expected 1 fill event, got %d", len(bm.Fills))
	}

	// Portfolio cash should remain exactly 250
	if pf.Cash != 250.0 {
		t.Errorf("Expected cash to remain 250, got %f", pf.Cash)
	}
}
