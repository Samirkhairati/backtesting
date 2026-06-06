package simulator

import (
	"time"

	"github.com/samirkhairati/backtesting/engine/broker"
	"github.com/samirkhairati/backtesting/engine/data"
	"github.com/samirkhairati/backtesting/engine/portfolio"
	"github.com/samirkhairati/backtesting/engine/strategy"
)

type Engine struct {
	CurrentTime  time.Time
	Portfolio    *portfolio.Portfolio
	OrderManager *broker.OrderManager

	iterator data.BarIterator
	strategy strategy.Strategy
}

func NewEngine(iter data.BarIterator, strat strategy.Strategy) *Engine {
	// Start with $100,000 in cash by default
	pf := portfolio.New()
	pf.Deposit(100000)

	return &Engine{
		Portfolio:    pf,
		OrderManager: broker.NewOrderManager(pf),
		iterator:     iter,
		strategy:     strat,
	}
}

func (e *Engine) Run() error {
	for e.iterator.Next() {
		bar := e.iterator.Bar()
		e.CurrentTime = bar.Timestamp
		e.OrderManager.SetTime(bar.Timestamp)

		// 1. EXECUTE orders from yesterday using TODAY'S OPEN price
		e.OrderManager.ExecutePendingOrders(bar)

		// 2. Update portfolio with TODAY'S CLOSE price
		e.Portfolio.UpdateMarketPrice(bar.Symbol, bar.Close)

		// 3. Let strategy evaluate TODAY'S CLOSE and submit new orders for tomorrow
		e.strategy.OnBar(e.OrderManager, bar)

		// 4. Snapshot equity
		e.Portfolio.RecordEquity(bar.Timestamp)
	}

	return e.iterator.Err()
}
