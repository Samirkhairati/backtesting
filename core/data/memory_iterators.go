package data

import (
	"time"
)

// MemoryBarIterator implements BarIterator over a slice of bars in memory.
type MemoryBarIterator struct {
	bars  []Bar
	index int
}

func NewMemoryBarIterator(bars []Bar) *MemoryBarIterator {
	return &MemoryBarIterator{
		bars:  bars,
		index: -1,
	}
}

func (i *MemoryBarIterator) Next() bool {
	i.index++
	return i.index < len(i.bars)
}

func (i *MemoryBarIterator) Bar() Bar {
	if i.index >= 0 && i.index < len(i.bars) {
		return i.bars[i.index]
	}
	return Bar{}
}

func (i *MemoryBarIterator) Err() error {
	return nil
}

// MemoryMultiBarIterator implements MultiBarIterator by merging multiple slices of Bars in memory.
// It expects the individual slices to be already sorted by Timestamp ASC.
type MemoryMultiBarIterator struct {
	symbols       []string
	barsPerSymbol map[string][]Bar
	indices       map[string]int
	
	currentBars   map[string]Bar
	currentTS     time.Time
}

func NewMemoryMultiBarIterator(symbols []string, barsPerSymbol map[string][]Bar) *MemoryMultiBarIterator {
	indices := make(map[string]int)
	for _, sym := range symbols {
		indices[sym] = 0
	}
	return &MemoryMultiBarIterator{
		symbols:       symbols,
		barsPerSymbol: barsPerSymbol,
		indices:       indices,
	}
}

func (i *MemoryMultiBarIterator) Next() bool {
	// Find the minimum timestamp among all symbols at their current indices
	var minTS time.Time
	var hasData bool

	for _, sym := range i.symbols {
		idx := i.indices[sym]
		bars := i.barsPerSymbol[sym]
		if idx < len(bars) {
			ts := bars[idx].Timestamp
			if !hasData || ts.Before(minTS) {
				minTS = ts
				hasData = true
			}
		}
	}

	if !hasData {
		return false
	}

	i.currentTS = minTS
	i.currentBars = make(map[string]Bar)

	// Collect all bars for the minTS and advance their indices
	for _, sym := range i.symbols {
		idx := i.indices[sym]
		bars := i.barsPerSymbol[sym]
		if idx < len(bars) && bars[idx].Timestamp.Equal(minTS) {
			i.currentBars[sym] = bars[idx]
			i.indices[sym]++
		}
	}

	return true
}

func (i *MemoryMultiBarIterator) Bars() map[string]Bar {
	return i.currentBars
}

func (i *MemoryMultiBarIterator) Timestamp() time.Time {
	return i.currentTS
}

func (i *MemoryMultiBarIterator) Err() error {
	return nil
}
