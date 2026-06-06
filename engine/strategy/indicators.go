package strategy

// SMA calculates the Simple Moving Average over the given period.
// It assumes the prices slice is ordered chronologically (oldest to newest)
// and takes the last `period` elements.
// Returns 0 if there are not enough data points.
func SMA(prices []float64, period int) float64 {
	if len(prices) < period || period <= 0 {
		return 0
	}

	sum := 0.0
	// Take the most recent `period` prices
	recentPrices := prices[len(prices)-period:]
	for _, p := range recentPrices {
		sum += p
	}

	return sum / float64(period)
}
