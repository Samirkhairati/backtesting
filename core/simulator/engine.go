package simulator

import (
	"time"

	"github.com/samirkhairati/backtesting/core/broker"
	"github.com/samirkhairati/backtesting/core/data"
	"github.com/samirkhairati/backtesting/core/portfolio"
	"github.com/samirkhairati/backtesting/core/strategy"
)

type Engine struct {
	CurrentTime  time.Time
	Portfolio    *portfolio.Portfolio
	OrderManager *broker.OrderManager

	iterator data.MultiBarIterator
	strategy strategy.Strategy
}

func NewEngine(iter data.MultiBarIterator, strat strategy.Strategy) *Engine {
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
		bars := e.iterator.Bars()
		ts := e.iterator.Timestamp()

		e.CurrentTime = ts
		e.OrderManager.SetTime(ts)

		// 1. EXECUTE orders from yesterday using TODAY'S OPEN price for ALL symbols
		for _, bar := range bars {
			e.OrderManager.ExecutePendingOrders(bar)
		}

		// 2. Update portfolio with TODAY'S CLOSE price for ALL symbols
		for symbol, bar := range bars {
			e.Portfolio.UpdateMarketPrice(symbol, bar.Close)
		}

		// 3. Let strategy evaluate TODAY'S CLOSE and submit new orders for tomorrow
		e.strategy.OnBars(e.OrderManager, bars)

		// 4. Snapshot equity (ONCE per timestamp)
		e.Portfolio.RecordEquity(ts)
	}

	return e.iterator.Err()
}
