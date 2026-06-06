package strategy

import (
	"log"

	"github.com/samirkhairati/backtesting/engine/broker"
	"github.com/samirkhairati/backtesting/engine/data"
)

type SMACrossStrategy struct {
	fastSMA  *SMA
	slowSMA  *SMA
	prevFast float64
	prevSlow float64
	ready    bool
}

func NewSMACrossStrategy(fast, slow int) *SMACrossStrategy {
	return &SMACrossStrategy{
		fastSMA: NewSMA(fast),
		slowSMA: NewSMA(slow),
	}
}

func (s *SMACrossStrategy) OnBar(b broker.Broker, bar data.Bar) {
	fast := s.fastSMA.Update(bar.Close)
	slow := s.slowSMA.Update(bar.Close)

	// Wait until we have enough data for the slow SMA
	if slow == 0 {
		return
	}

	// We need 1 day of valid SMAs before we can detect a "cross"
	if !s.ready {
		s.prevFast = fast
		s.prevSlow = slow
		s.ready = true
		return
	}

	pos := b.GetPosition(bar.Symbol)
	isLong := pos != nil && pos.Quantity > 0

	// Golden Cross: Fast crosses above Slow
	if s.prevFast <= s.prevSlow && fast > slow {
		if !isLong {
			cash := b.GetCash()
			affordable := cash * 0.95 // 95% of cash to buffer gap ups
			shares := affordable / bar.Close
			if shares > 0 {
				log.Printf("[SMACross] %s Golden Cross detected (Fast: %.2f > Slow: %.2f). Buying %f shares.", bar.Symbol, fast, slow, shares)
				b.SubmitOrder(bar.Symbol, shares)
			}
		}
	}

	// Death Cross: Fast crosses below Slow
	if s.prevFast >= s.prevSlow && fast < slow {
		if isLong {
			log.Printf("[SMACross] %s Death Cross detected (Fast: %.2f < Slow: %.2f). Closing position.", bar.Symbol, fast, slow)
			b.ClosePosition(bar.Symbol)
		}
	}

	// Save today's values for tomorrow's comparison
	s.prevFast = fast
	s.prevSlow = slow
}
