package broker

import (
	"time"

	"github.com/google/uuid"
)

type OrderType string
type OrderStatus string

const (
	OrderTypeMarket OrderType = "MARKET"

	OrderStatusPending  OrderStatus = "PENDING"
	OrderStatusFilled   OrderStatus = "FILLED"
	OrderStatusRejected OrderStatus = "REJECTED"
)

// Order represents a strategy's intent to trade.
type Order struct {
	ID        string
	Symbol    string
	Quantity  float64 // Positive = Buy, Negative = Sell
	Type      OrderType
	Status    OrderStatus
	CreatedAt time.Time
}

// NewMarketOrder creates a standard market order that gets queued for execution.
func NewMarketOrder(symbol string, quantity float64, timestamp time.Time) *Order {
	return &Order{
		ID:        uuid.New().String(),
		Symbol:    symbol,
		Quantity:  quantity,
		Type:      OrderTypeMarket,
		Status:    OrderStatusPending,
		CreatedAt: timestamp,
	}
}
