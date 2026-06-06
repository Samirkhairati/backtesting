package broker_test

import (
	"testing"
	"time"

	"github.com/samirkhairati/backtesting/engine/broker"
	"github.com/samirkhairati/backtesting/engine/portfolio"
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
	pf.AddPosition("MSFT", 25, 200)

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
	pf.AddPosition("AMD", 10, 100)

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
