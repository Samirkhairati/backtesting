package data

import (
	"errors"
	"testing"
	"time"
)

func TestValidateBarDataset(t *testing.T) {
	t1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		bars    []Bar
		wantErr error
	}{
		{
			name: "valid dataset",
			bars: []Bar{
				{Symbol: "AAPL", Timestamp: t1, Open: 150, High: 155, Low: 149, Close: 154, Volume: 1000},
				{Symbol: "AAPL", Timestamp: t2, Open: 154, High: 158, Low: 153, Close: 157, Volume: 1200},
			},
			wantErr: nil,
		},
		{
			name: "negative price",
			bars: []Bar{
				{Symbol: "AAPL", Timestamp: t1, Open: 150, High: 155, Low: -1, Close: 154, Volume: 1000},
			},
			wantErr: ErrNegativePrice,
		},
		{
			name: "missing time",
			bars: []Bar{
				{Symbol: "AAPL", Timestamp: time.Time{}, Open: 150, High: 155, Low: 149, Close: 154, Volume: 1000},
			},
			wantErr: ErrMissingTime,
		},
		{
			name: "duplicate time",
			bars: []Bar{
				{Symbol: "AAPL", Timestamp: t1, Open: 150, High: 155, Low: 149, Close: 154, Volume: 1000},
				{Symbol: "AAPL", Timestamp: t1, Open: 154, High: 158, Low: 153, Close: 157, Volume: 1200},
			},
			wantErr: ErrDuplicateTime,
		},
		{
			name: "out of order time",
			bars: []Bar{
				{Symbol: "AAPL", Timestamp: t2, Open: 150, High: 155, Low: 149, Close: 154, Volume: 1000},
				{Symbol: "AAPL", Timestamp: t1, Open: 154, High: 158, Low: 153, Close: 157, Volume: 1200},
			},
			wantErr: ErrOutOfOrderTime,
		},
		{
			name: "negative volume",
			bars: []Bar{
				{Symbol: "AAPL", Timestamp: t3, Open: 150, High: 155, Low: 149, Close: 154, Volume: -1000},
			},
			wantErr: ErrZeroVolume,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBarDataset(tt.bars)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateBarDataset() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				if err == nil {
					t.Errorf("ValidateBarDataset() expected error, got nil")
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateBarDataset() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
