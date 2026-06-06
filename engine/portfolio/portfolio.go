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
	ClosedTrades  []Trade
	OpenLots      map[string]*LotQueue
}

// New creates and initializes a new Portfolio.
func New() *Portfolio {
	return &Portfolio{
		Cash:          0.0,
		Positions:     make(map[string]*Position),
		CurrentPrices: make(map[string]float64),
		EquityCurve:   make([]EquityPoint, 0),
		ClosedTrades:  make([]Trade, 0),
		OpenLots:      make(map[string]*LotQueue),
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
// It matches sell orders against open tax lots using a FIFO queue to record Trades.
func (p *Portfolio) AddPosition(symbol string, quantity float64, price float64, timestamp time.Time) error {
	cost := quantity * price

	// If buying, check cash first
	if quantity > 0 && p.Cash < cost {
		return ErrInsufficientCash
	}

	p.Cash -= cost

	// Ensure LotQueue exists
	if p.OpenLots[symbol] == nil {
		p.OpenLots[symbol] = &LotQueue{}
	}

	queue := p.OpenLots[symbol]

	if quantity > 0 {
		// Long trade: Push a new Lot
		queue.Push(Lot{
			Symbol:    symbol,
			Quantity:  quantity,
			Price:     price,
			Timestamp: timestamp,
		})
	} else if quantity < 0 {
		// Short trade or closing a Long: Match against open Lots
		sellQuantity := -quantity

		for sellQuantity > 0 && !queue.IsEmpty() {
			lot := queue.PeekFront()
			if lot == nil {
				break
			}

			var filledQty float64
			if lot.Quantity <= sellQuantity {
				// We consume this entire lot
				filledQty = lot.Quantity
				sellQuantity -= filledQty
				queue.PopFront()
			} else {
				// We consume a portion of this lot
				filledQty = sellQuantity
				lot.Quantity -= filledQty
				sellQuantity = 0
			}

			// Record the closed trade
			pnl := (price - lot.Price) * filledQty
			returnPct := ((price - lot.Price) / lot.Price) * 100.0

			p.ClosedTrades = append(p.ClosedTrades, Trade{
				Symbol:     symbol,
				EntryTime:  lot.Timestamp,
				ExitTime:   timestamp,
				EntryPrice: lot.Price,
				ExitPrice:  price,
				Quantity:   filledQty,
				PnL:        pnl,
				ReturnPct:  returnPct,
			})
		}
		
		if sellQuantity > 0 {
			// We sold more than we had open (Shorting).
			// We push a short lot. For this engine, we mostly do Long only, 
			// but we handle it just in case.
			queue.Push(Lot{
				Symbol:    symbol,
				Quantity:  -sellQuantity, // negative quantity denotes a short lot
				Price:     price,
				Timestamp: timestamp,
			})
		}
	}

	// Update the aggregated Position view
	pos, exists := p.Positions[symbol]
	if !exists {
		p.Positions[symbol] = &Position{
			Symbol:       symbol,
			Quantity:     quantity,
			AveragePrice: price,
		}
		return nil
	}

	// Calculate new average price (simplified; a true average requires full lot traversal,
	// but this maintains our legacy average cost implementation for TotalValue).
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
