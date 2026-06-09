package portfolio

import (
	"errors"
	"math"
	"time"
)

var (
	ErrInsufficientCash   = errors.New("insufficient cash")
	ErrInsufficientMargin = errors.New("insufficient margin")
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
// It symmetrically matches buy/sell orders against open tax lots using a FIFO queue to record Trades.
func (p *Portfolio) AddPosition(symbol string, quantity float64, price float64, timestamp time.Time) error {
	cost := quantity * price

	// If buying, check cash
	if quantity > 0 && p.Cash < cost {
		return ErrInsufficientCash
	}

	// If shorting, check margin (short exposure shouldn't exceed TotalValue)
	if quantity < 0 && -cost > p.TotalValue() {
		return ErrInsufficientMargin
	}

	p.Cash -= cost

	// Ensure LotQueue exists
	if p.OpenLots[symbol] == nil {
		p.OpenLots[symbol] = &LotQueue{}
	}

	queue := p.OpenLots[symbol]
	remainingQty := quantity

	// Match against open lots of the opposite sign
	for remainingQty != 0 && !queue.IsEmpty() {
		lot := queue.PeekFront()
		if lot == nil {
			break
		}

		// If same sign, we don't match, we just append to position
		if (remainingQty > 0 && lot.Quantity > 0) || (remainingQty < 0 && lot.Quantity < 0) {
			break
		}

		// Opposite signs, we can match and close out the lot
		matchQty := math.Min(math.Abs(remainingQty), math.Abs(lot.Quantity))

		var pnl float64
		var returnPct float64

		if remainingQty < 0 {
			// Selling to close a Long
			pnl = (price - lot.Price) * matchQty
			returnPct = ((price - lot.Price) / lot.Price) * 100.0

			remainingQty += matchQty // selling removes positive quantity
			lot.Quantity -= matchQty
		} else {
			// Buying to close a Short
			pnl = (lot.Price - price) * matchQty
			returnPct = ((lot.Price - price) / lot.Price) * 100.0

			remainingQty -= matchQty // buying removes negative quantity
			lot.Quantity += matchQty
		}

		if lot.Quantity == 0 {
			queue.PopFront()
		}

		p.ClosedTrades = append(p.ClosedTrades, Trade{
			Symbol:     symbol,
			EntryTime:  lot.Timestamp,
			ExitTime:   timestamp,
			EntryPrice: lot.Price,
			ExitPrice:  price,
			Quantity:   matchQty,
			PnL:        pnl,
			ReturnPct:  returnPct,
		})
	}

	// If there's still quantity left, push it as a new lot
	if remainingQty != 0 {
		queue.Push(Lot{
			Symbol:    symbol,
			Quantity:  remainingQty,
			Price:     price,
			Timestamp: timestamp,
		})
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

	// Average cost calculation (only for increasing position size,
	// decreasing size doesn't change average cost basis of remaining shares)
	newQuantity := pos.Quantity + quantity
	if newQuantity == 0 {
		delete(p.Positions, symbol)
		return nil
	}

	// If the position is growing in its current direction, recalculate average cost
	if (pos.Quantity > 0 && quantity > 0) || (pos.Quantity < 0 && quantity < 0) {
		totalCost := (math.Abs(pos.Quantity) * pos.AveragePrice) + (math.Abs(quantity) * price)
		pos.AveragePrice = totalCost / math.Abs(newQuantity)
	}

	pos.Quantity = newQuantity
	return nil
}
