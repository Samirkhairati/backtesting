package data

import (
	"errors"
	"fmt"
)

var (
	ErrNegativePrice   = errors.New("dataset contains negative prices")
	ErrZeroVolume      = errors.New("dataset contains negative volume") // volume can be zero on non-trading days/halts sometimes, but negative is invalid. We will check negative.
	ErrMissingTime     = errors.New("dataset contains zero timestamp")
	ErrDuplicateTime   = errors.New("dataset contains duplicate timestamps")
	ErrOutOfOrderTime  = errors.New("dataset is out of order")
)

// ValidateBarDataset ensures a slice of bars conforms to all requirements.
func ValidateBarDataset(bars []Bar) error {
	if len(bars) == 0 {
		return nil
	}

	for i, bar := range bars {
		if bar.Timestamp.IsZero() {
			return fmt.Errorf("bar index %d: %w", i, ErrMissingTime)
		}
		if bar.Open < 0 || bar.High < 0 || bar.Low < 0 || bar.Close < 0 {
			return fmt.Errorf("bar index %d (symbol %s): %w", i, bar.Symbol, ErrNegativePrice)
		}
		if bar.Volume < 0 {
			return fmt.Errorf("bar index %d (symbol %s): %w", i, bar.Symbol, ErrZeroVolume)
		}

		if i > 0 {
			prev := bars[i-1]
			if bar.Timestamp.Equal(prev.Timestamp) {
				return fmt.Errorf("bar index %d: %w (%v)", i, ErrDuplicateTime, bar.Timestamp)
			}
			if bar.Timestamp.Before(prev.Timestamp) {
				return fmt.Errorf("bar index %d: %w (current %v is before previous %v)", i, ErrOutOfOrderTime, bar.Timestamp, prev.Timestamp)
			}
		}
	}

	return nil
}
