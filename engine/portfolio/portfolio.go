package portfolio

import (
	"errors"
	"time"
)

var (
	ErrInsufficientCash = errors.New("insufficient cash")
)

// Position represents an open holding of a specific symbol.
type Position struct {
	Symbol       string
	Quantity     float64
	AveragePrice float64
}

// EquityPoint represents the total portfolio value at a specific point in time.
type EquityPoint struct {
	Timestamp time.Time
	Value     float64
}

// Portfolio holds the state of the backtest: cash, active positions, current market prices, and the equity curve.
type Portfolio struct {
	Cash          float64
	Positions     map[string]*Position
	CurrentPrices map[string]float64
	EquityCurve   []EquityPoint
}

// New creates and initializes a new Portfolio.
func New() *Portfolio {
	return &Portfolio{
		Cash:          0.0,
		Positions:     make(map[string]*Position),
		CurrentPrices: make(map[string]float64),
		EquityCurve:   make([]EquityPoint, 0),
	}
}

// Deposit adds cash to the portfolio.
func (p *Portfolio) Deposit(amount float64) {
	p.Cash += amount
}

// Withdraw removes cash from the portfolio. It returns an error if there is insufficient cash.
func (p *Portfolio) Withdraw(amount float64) error {
	if p.Cash < amount {
		return ErrInsufficientCash
	}
	p.Cash -= amount
	return nil
}

// UpdateMarketPrice is called by the simulation engine on every tick so the portfolio knows the current value of its holdings.
func (p *Portfolio) UpdateMarketPrice(symbol string, price float64) {
	p.CurrentPrices[symbol] = price
}

// TotalValue returns the total net liquidation value of the portfolio (Cash + Market Value of all Positions).
func (p *Portfolio) TotalValue() float64 {
	val := p.Cash
	for symbol, pos := range p.Positions {
		if currentPrice, ok := p.CurrentPrices[symbol]; ok {
			val += pos.Quantity * currentPrice
		} else {
			// Fallback to average price if market price is somehow unknown
			val += pos.Quantity * pos.AveragePrice
		}
	}
	return val
}

// RecordEquity snapshots the current total value of the portfolio and appends it to the EquityCurve.
func (p *Portfolio) RecordEquity(timestamp time.Time) {
	p.EquityCurve = append(p.EquityCurve, EquityPoint{
		Timestamp: timestamp,
		Value:     p.TotalValue(),
	})
}

// AddPosition is a low-level accounting hook used by the execution engine to adjust shares and cash.
// To avoid phantom profits, purchasing shares immediately deducts the exact required cash.
func (p *Portfolio) AddPosition(symbol string, quantity float64, price float64) error {
	cost := quantity * price

	// If buying (positive quantity), check cash first
	if quantity > 0 && p.Cash < cost {
		return ErrInsufficientCash
	}

	p.Cash -= cost

	pos, exists := p.Positions[symbol]
	if !exists {
		p.Positions[symbol] = &Position{
			Symbol:       symbol,
			Quantity:     quantity,
			AveragePrice: price,
		}
		return nil
	}

	// Update average price based on the new combined quantity
	totalCost := (pos.Quantity * pos.AveragePrice) + cost
	newQuantity := pos.Quantity + quantity

	if newQuantity == 0 {
		delete(p.Positions, symbol)
		return nil
	}

	pos.Quantity = newQuantity
	pos.AveragePrice = totalCost / newQuantity
	return nil
}
