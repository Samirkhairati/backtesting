package data

// BarIterator provides sequential access to historical market data.
type BarIterator interface {
	// Next advances the iterator to the next bar and returns true if a bar is available.
	Next() bool
	// Bar returns the current bar.
	Bar() Bar
	// Err returns any error encountered during iteration.
	Err() error
}
