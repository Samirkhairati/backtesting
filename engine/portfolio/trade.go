package portfolio

import "time"

// Trade represents a completed round-trip transaction (e.g. buy to open, sell to close).
type Trade struct {
	Symbol     string
	EntryTime  time.Time
	ExitTime   time.Time
	EntryPrice float64
	ExitPrice  float64
	Quantity   float64
	PnL        float64
	ReturnPct  float64
}

// Lot represents an open tax lot / individual buy or sell execution that has not yet been closed.
type Lot struct {
	Symbol    string
	Quantity  float64
	Price     float64
	Timestamp time.Time
}

// LotQueue is an optimized O(1) FIFO queue for open tax lots.
// We use a slice with a Head index to prevent O(N) shifts when closing lots.
type LotQueue struct {
	lots []Lot
	head int
}

// Push adds a new lot to the end of the queue.
func (q *LotQueue) Push(l Lot) {
	q.lots = append(q.lots, l)
}

// PopFront removes and returns the oldest lot in O(1) time.
func (q *LotQueue) PopFront() (Lot, bool) {
	if q.head >= len(q.lots) {
		return Lot{}, false
	}
	l := q.lots[q.head]
	q.head++
	return l, true
}

// PeekFront returns a pointer to the oldest lot without removing it.
func (q *LotQueue) PeekFront() *Lot {
	if q.head >= len(q.lots) {
		return nil
	}
	return &q.lots[q.head]
}

// IsEmpty returns true if there are no open lots.
func (q *LotQueue) IsEmpty() bool {
	return q.head >= len(q.lots)
}
