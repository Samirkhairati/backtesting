package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samirkhairati/backtesting/engine/data"
	"github.com/samirkhairati/backtesting/engine/data/store"
)

// ensure postgresStore implements store.Store
var _ store.Store = (*postgresStore)(nil)

type postgresStore struct {
	db *pgxpool.Pool
}

// NewStore creates a new PostgreSQL backed historical data store.
func NewStore(db *pgxpool.Pool) store.Store {
	return &postgresStore{db: db}
}

// GetBars retrieves bars for a given symbol, returning an iterator.
func (p *postgresStore) GetBars(ctx context.Context, symbol string) (data.BarIterator, error) {
	// Query rows ordered by timestamp ASC to ensure chronological replay
	query := `
		SELECT symbol, timestamp, open, high, low, close, volume 
		FROM daily_bars 
		WHERE symbol = $1 
		ORDER BY timestamp ASC
	`
	rows, err := p.db.Query(ctx, query, symbol)
	if err != nil {
		return nil, err
	}

	return &postgresIterator{rows: rows}, nil
}

// postgresIterator implements data.BarIterator
type postgresIterator struct {
	rows pgx.Rows
	err  error
	bar  data.Bar
}

func (i *postgresIterator) Next() bool {
	if i.err != nil {
		return false
	}
	
	if i.rows.Next() {
		err := i.rows.Scan(
			&i.bar.Symbol,
			&i.bar.Timestamp,
			&i.bar.Open,
			&i.bar.High,
			&i.bar.Low,
			&i.bar.Close,
			&i.bar.Volume,
		)
		if err != nil {
			i.err = err
			return false
		}
		return true
	}
	
	if err := i.rows.Err(); err != nil {
		i.err = err
	}
	return false
}

func (i *postgresIterator) Bar() data.Bar {
	return i.bar
}

func (i *postgresIterator) Err() error {
	return i.err
}
