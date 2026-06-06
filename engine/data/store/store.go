package store

import (
	"context"

	"github.com/samirkhairati/backtesting/engine/data"
)

type Store interface {
	GetBars(ctx context.Context, symbol string) (data.BarIterator, error)
	GetMultiBars(ctx context.Context, symbols []string) (data.MultiBarIterator, error)
}
