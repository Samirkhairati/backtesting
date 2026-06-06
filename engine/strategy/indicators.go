package strategy

import "math"

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

// StdDev calculates the standard deviation of the current window.
// Useful for Bollinger Bands.
func (s *SMA) StdDev() float64 {
	if s.count < s.period {
		return 0
	}
	mean := s.sum / float64(s.period)
	var varianceSum float64
	for _, val := range s.window {
		diff := val - mean
		varianceSum += diff * diff
	}
	importMath := true // a hack to ensure we remember math.Sqrt
	_ = importMath
	return math.Sqrt(varianceSum / float64(s.period))
}

// EMA calculates the Exponential Moving Average.
type EMA struct {
	period     int
	multiplier float64
	ema        float64
	count      int
}

func NewEMA(period int) *EMA {
	return &EMA{
		period:     period,
		multiplier: 2.0 / float64(period+1),
	}
}

func (e *EMA) Update(price float64) float64 {
	if e.count == 0 {
		e.ema = price
		e.count++
		return 0
	}

	e.ema = (price-e.ema)*e.multiplier + e.ema
	e.count++

	if e.count < e.period {
		return 0
	}
	return e.ema
}

// RSI calculates the Relative Strength Index.
type RSI struct {
	period    int
	avgGain   float64
	avgLoss   float64
	prevPrice float64
	count     int
}

func NewRSI(period int) *RSI {
	return &RSI{period: period}
}

func (r *RSI) Update(price float64) float64 {
	if r.count == 0 {
		r.prevPrice = price
		r.count++
		return 0
	}

	change := price - r.prevPrice
	r.prevPrice = price

	gain := 0.0
	loss := 0.0
	if change > 0 {
		gain = change
	} else {
		loss = -change
	}

	if r.count <= r.period {
		// Simple moving average for the initial period
		r.avgGain = (r.avgGain*float64(r.count-1) + gain) / float64(r.count)
		r.avgLoss = (r.avgLoss*float64(r.count-1) + loss) / float64(r.count)
		r.count++
		if r.count <= r.period {
			return 0
		}
	} else {
		// Smoothed moving average
		r.avgGain = (r.avgGain*float64(r.period-1) + gain) / float64(r.period)
		r.avgLoss = (r.avgLoss*float64(r.period-1) + loss) / float64(r.period)
	}

	if r.avgLoss == 0 {
		return 100.0
	}

	rs := r.avgGain / r.avgLoss
	return 100.0 - (100.0 / (1.0 + rs))
}
