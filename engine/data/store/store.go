package store

import (
	"context"

	"github.com/samirkhairati/backtesting/engine/data"
)

type Store interface {
	GetBars(ctx context.Context, symbol string) (data.BarIterator, error)
}
