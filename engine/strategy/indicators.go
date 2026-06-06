package strategy

// SMA provides an optimized O(1) calculation for Simple Moving Averages
// by maintaining a circular buffer and a rolling sum.
type SMA struct {
	period int
	window []float64
	idx    int
	sum    float64
	count  int
}

// NewSMA creates a new optimized moving average indicator.
func NewSMA(period int) *SMA {
	return &SMA{
		period: period,
		window: make([]float64, period),
	}
}

// Update feeds a new price into the rolling window and returns the current SMA.
// If the window is not yet full, it returns 0.
func (s *SMA) Update(price float64) float64 {
	// Subtract the oldest price from sum (if buffer wasn't full, this is 0)
	s.sum -= s.window[s.idx]

	// Add the new price
	s.window[s.idx] = price
	s.sum += price

	// Move circular index forward
	s.idx = (s.idx + 1) % s.period

	if s.count < s.period {
		s.count++
	}

	// Wait until buffer is full before returning valid SMA
	if s.count < s.period {
		return 0
	}

	return s.sum / float64(s.period)
}
