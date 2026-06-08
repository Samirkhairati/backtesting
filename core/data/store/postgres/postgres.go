package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samirkhairati/backtesting/core/data"
	"github.com/samirkhairati/backtesting/core/data/store"
)

var _ store.Store = (*postgresStore)(nil)

type postgresStore struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) store.Store {
	return &postgresStore{db: db}
}

func (p *postgresStore) GetBars(ctx context.Context, symbol string) (data.BarIterator, error) {

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

func (p *postgresStore) GetMultiBars(ctx context.Context, symbols []string) (data.MultiBarIterator, error) {
	query := `
		SELECT symbol, timestamp, open, high, low, close, volume 
		FROM daily_bars 
		WHERE symbol = ANY($1) 
		ORDER BY timestamp ASC
	`
	rows, err := p.db.Query(ctx, query, symbols)
	if err != nil {
		return nil, err
	}

	iter := &postgresMultiIterator{
		rows: rows,
		bars: make(map[string]data.Bar),
	}

	// Read the very first row to initialize state
	if iter.rows.Next() {
		err := iter.rows.Scan(
			&iter.nextBar.Symbol,
			&iter.nextBar.Timestamp,
			&iter.nextBar.Open,
			&iter.nextBar.High,
			&iter.nextBar.Low,
			&iter.nextBar.Close,
			&iter.nextBar.Volume,
		)
		if err != nil {
			iter.err = err
		} else {
			iter.hasData = true
		}
	} else if err := iter.rows.Err(); err != nil {
		iter.err = err
	}

	return iter, nil
}

type postgresMultiIterator struct {
	rows      pgx.Rows
	err       error
	bars      map[string]data.Bar
	timestamp time.Time

	// lookahead state
	nextBar data.Bar
	hasData bool
}

func (i *postgresMultiIterator) Next() bool {
	if i.err != nil || !i.hasData {
		return false
	}

	// Clear previous bars
	i.bars = make(map[string]data.Bar)
	
	// The timestamp for this iteration is the timestamp of the lookahead bar
	i.timestamp = i.nextBar.Timestamp

	// Keep reading rows as long as they have the same timestamp
	for i.hasData && i.nextBar.Timestamp.Equal(i.timestamp) {
		i.bars[i.nextBar.Symbol] = i.nextBar

		// Fetch next row
		if i.rows.Next() {
			err := i.rows.Scan(
				&i.nextBar.Symbol,
				&i.nextBar.Timestamp,
				&i.nextBar.Open,
				&i.nextBar.High,
				&i.nextBar.Low,
				&i.nextBar.Close,
				&i.nextBar.Volume,
			)
			if err != nil {
				i.err = err
				i.hasData = false
			}
		} else {
			i.hasData = false
			if err := i.rows.Err(); err != nil {
				i.err = err
			}
		}
	}

	return len(i.bars) > 0
}

func (i *postgresMultiIterator) Bars() map[string]data.Bar {
	return i.bars
}

func (i *postgresMultiIterator) Timestamp() time.Time {
	return i.timestamp
}

func (i *postgresMultiIterator) Err() error {
	return i.err
}
