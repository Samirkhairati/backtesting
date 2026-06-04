package simulator

import (
	"time"

	"github.com/samirkhairati/backtesting/engine/data"
	"github.com/samirkhairati/backtesting/engine/strategy"
)

type Engine struct {
	CurrentTime time.Time

	iterator data.BarIterator
	strategy strategy.Strategy
}

func NewEngine(iter data.BarIterator, strat strategy.Strategy) *Engine {
	return &Engine{
		iterator: iter,
		strategy: strat,
	}
}

func (e *Engine) Run() error {
	for e.iterator.Next() {
		bar := e.iterator.Bar()

		e.CurrentTime = bar.Timestamp

		e.strategy.OnBar(bar)
	}

	return e.iterator.Err()
}
