package data

import "time"

type MultiBarIterator interface {
	Next() bool

	Bars() map[string]Bar

	Timestamp() time.Time

	Err() error
}
