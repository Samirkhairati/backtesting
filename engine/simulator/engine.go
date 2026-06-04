package simulator

import (
	"time"

	"github.com/samirkhairati/backtesting/engine/data"
	"github.com/samirkhairati/backtesting/engine/strategy"
)

// Engine is the core loop of the backtesting system.
// It is responsible for simulating the passage of time and ensuring that
// strategies only ever see data chronologically (no lookahead bias).
type Engine struct {
	// CurrentTime is the simulated market clock. It advances with every new bar.
	CurrentTime time.Time

	iterator data.BarIterator
	strategy strategy.Strategy
}

// NewEngine initializes a new Replay Engine.
func NewEngine(iter data.BarIterator, strat strategy.Strategy) *Engine {
	return &Engine{
		iterator: iter,
		strategy: strat,
	}
}

// Run executes the simulation. It steps through the historical data one bar at a time,
// advancing the clock, and notifying the strategy.
func (e *Engine) Run() error {
	for e.iterator.Next() {
		bar := e.iterator.Bar()

		// Advance the simulated clock to the current bar's timestamp
		e.CurrentTime = bar.Timestamp

		// Reveal the data to the strategy
		e.strategy.OnBar(bar)
	}

	// Return any errors encountered during the data iteration
	return e.iterator.Err()
}
