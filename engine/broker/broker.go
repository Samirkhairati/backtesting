package broker

import (
	"time"

	"github.com/samirkhairati/backtesting/engine/portfolio"
)

// Broker is the restricted interface exposed to strategies to express trading intent.
type Broker interface {
	SubmitOrder(symbol string, quantity float64)
	ClosePosition(symbol string) error
	GetPosition(symbol string) *portfolio.Position
}

// OrderManager implements Broker and acts as the bridge between strategies and the execution engine.
type OrderManager struct {
	PendingOrders []*Order
	portfolio     *portfolio.Portfolio

	// internal clock injected by simulator to timestamp new orders
	currentTime time.Time 
}

// NewOrderManager initializes an empty OrderManager.
func NewOrderManager(pf *portfolio.Portfolio) *OrderManager {
	return &OrderManager{
		PendingOrders: make([]*Order, 0),
		portfolio:     pf,
	}
}

// SetTime is called by the simulator engine on every tick to keep the broker's clock synchronized.
func (m *OrderManager) SetTime(t time.Time) {
	m.currentTime = t
}

// SubmitOrder generates a new MARKET order and appends it to the pending queue.
func (m *OrderManager) SubmitOrder(symbol string, quantity float64) {
	if quantity == 0 {
		return
	}
	order := NewMarketOrder(symbol, quantity, m.currentTime)
	m.PendingOrders = append(m.PendingOrders, order)
}

// ClosePosition checks the current portfolio for the exact shares held and submits an inverse order to flatten it.
func (m *OrderManager) ClosePosition(symbol string) error {
	pos := m.GetPosition(symbol)
	if pos == nil || pos.Quantity == 0 {
		return nil // Nothing to close
	}
	// Submit the exact inverse quantity
	m.SubmitOrder(symbol, -pos.Quantity)
	return nil
}

// GetPosition safely provides a read-only copy of the requested position, or nil if none exists.
// We return a copy by value (converted back to pointer) to strictly prevent strategies from modifying state directly.
func (m *OrderManager) GetPosition(symbol string) *portfolio.Position {
	if pos, exists := m.portfolio.Positions[symbol]; exists {
		// Return a copy so the strategy can't mutate the real portfolio pointer
		copyPos := *pos 
		return &copyPos
	}
	return nil
}
