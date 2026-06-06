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

		// 1. Update portfolio with latest market data
		e.Portfolio.UpdateMarketPrice(bar.Symbol, bar.Close)

		// 2. Execute strategy, handing it the broker connection
		e.strategy.OnBar(e.OrderManager, bar)

		// 3. Record end-of-day equity
		e.Portfolio.RecordEquity(bar.Timestamp)
	}

	return e.iterator.Err()
}
