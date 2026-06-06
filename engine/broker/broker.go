package broker

import (
	"log"
	"time"

	"github.com/samirkhairati/backtesting/engine/data"
	"github.com/samirkhairati/backtesting/engine/portfolio"
)

// Broker is the restricted interface exposed to strategies to express trading intent.
type Broker interface {
	SubmitOrder(symbol string, quantity float64)
	ClosePosition(symbol string) error
	GetPosition(symbol string) *portfolio.Position
}

// FillEvent tracks successful order executions
type FillEvent struct {
	OrderID   string
	Symbol    string
	Quantity  float64
	Price     float64
	Timestamp time.Time
}

// OrderManager implements Broker and acts as the bridge between strategies and the execution engine.
type OrderManager struct {
	PendingOrders []*Order
	Fills         []*FillEvent
	portfolio     *portfolio.Portfolio

	// internal clock injected by simulator to timestamp new orders
	currentTime time.Time 
}

// NewOrderManager initializes an empty OrderManager.
func NewOrderManager(pf *portfolio.Portfolio) *OrderManager {
	return &OrderManager{
		PendingOrders: make([]*Order, 0),
		Fills:         make([]*FillEvent, 0),
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

// ExecutePendingOrders processes all pending orders matching the current bar's symbol
// at the bar's Open price, effectively converting yesterday's signals into today's fills.
func (m *OrderManager) ExecutePendingOrders(bar data.Bar) {
	var remainingOrders []*Order

	for _, order := range m.PendingOrders {
		// Only process orders for the current symbol being updated
		if order.Symbol != bar.Symbol {
			remainingOrders = append(remainingOrders, order)
			continue
		}

		// Attempt to fill the order using today's OPEN price
		err := m.portfolio.AddPosition(order.Symbol, order.Quantity, bar.Open)
		
		if err != nil {
			// e.g. ErrInsufficientCash
			order.Status = OrderStatusRejected
			log.Printf("[Broker] Order REJECTED: %s %f shares @ $%.2f (Reason: %v)", order.Symbol, order.Quantity, bar.Open, err)
			// Order is dropped from queue (not appended to remainingOrders)
			continue
		}

		// Order successful
		order.Status = OrderStatusFilled
		fill := &FillEvent{
			OrderID:   order.ID,
			Symbol:    order.Symbol,
			Quantity:  order.Quantity,
			Price:     bar.Open,
			Timestamp: bar.Timestamp,
		}
		m.Fills = append(m.Fills, fill)
		
		log.Printf("[Broker] Order FILLED: %s %f shares @ $%.2f", order.Symbol, order.Quantity, bar.Open)
		// Order is dropped from pending queue
	}

	m.PendingOrders = remainingOrders
}
