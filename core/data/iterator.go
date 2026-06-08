package data

type BarIterator interface {
	Next() bool

	Bar() Bar

	Err() error
}
