package strategy

import (
	"log"

	"github.com/samirkhairati/backtesting/core/broker"
	"github.com/samirkhairati/backtesting/core/data"
)

type smaCrossState struct {
	fastSMA  *SMA
	slowSMA  *SMA
	prevFast float64
	prevSlow float64
	ready    bool
}

type SMACrossStrategy struct {
	fastPeriod int
	slowPeriod int
	states     map[string]*smaCrossState
}

func NewSMACrossStrategy(fast, slow int) *SMACrossStrategy {
	return &SMACrossStrategy{
		fastPeriod: fast,
		slowPeriod: slow,
		states:     make(map[string]*smaCrossState),
	}
}

func (s *SMACrossStrategy) OnBars(b broker.Broker, bars map[string]data.Bar) {
	for symbol, bar := range bars {
		state, exists := s.states[symbol]
		if !exists {
			state = &smaCrossState{
				fastSMA: NewSMA(s.fastPeriod),
				slowSMA: NewSMA(s.slowPeriod),
			}
			s.states[symbol] = state
		}

		fast := state.fastSMA.Update(bar.Close)
		slow := state.slowSMA.Update(bar.Close)

		// Wait until we have enough data for the slow SMA
		if slow == 0 {
			continue
		}

		// We need 1 day of valid SMAs before we can detect a "cross"
		if !state.ready {
			state.prevFast = fast
			state.prevSlow = slow
			state.ready = true
			continue
		}

		pos := b.GetPosition(symbol)
		isLong := pos != nil && pos.Quantity > 0

		// Golden Cross: Fast crosses above Slow
		if state.prevFast <= state.prevSlow && fast > slow {
			if !isLong {
				cash := b.GetCash()
				// For multi-asset, we shouldn't use all cash. Let's say we allocate a fraction
				// proportional to available cash to avoid starvation if multiple signals hit.
				// But for simplicity of this strategy, we'll just buy what we can with 20% of cash as an arbitrary allocation per signal.
				affordable := cash * 0.20
				shares := affordable / bar.Close
				if shares > 0 {
					log.Printf("[SMACross] %s Golden Cross detected (Fast: %.2f > Slow: %.2f). Buying %f shares.", symbol, fast, slow, shares)
					b.SubmitOrder(symbol, shares)
				}
			}
		}

		// Death Cross: Fast crosses below Slow
		if state.prevFast >= state.prevSlow && fast < slow {
			if isLong {
				log.Printf("[SMACross] %s Death Cross detected (Fast: %.2f < Slow: %.2f). Closing position.", symbol, fast, slow)
				b.ClosePosition(symbol)
			}
		}

		// Save today's values for tomorrow's comparison
		state.prevFast = fast
		state.prevSlow = slow
	}
}
