package store

import (
	"context"

	"github.com/samirkhairati/backtesting/engine/data"
)

// Store defines how historical data is retrieved for the engine.
type Store interface {
	// GetBars returns an iterator for a specific symbol.
	GetBars(ctx context.Context, symbol string) (data.BarIterator, error)
}
