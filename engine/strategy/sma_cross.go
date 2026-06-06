package strategy

import (
	"log"

	"github.com/samirkhairati/backtesting/engine/broker"
	"github.com/samirkhairati/backtesting/engine/data"
)

type SMACrossStrategy struct {
	fastPeriod int
	slowPeriod int
	history    []float64
}

func NewSMACrossStrategy(fast, slow int) *SMACrossStrategy {
	return &SMACrossStrategy{
		fastPeriod: fast,
		slowPeriod: slow,
		history:    make([]float64, 0),
	}
}

func (s *SMACrossStrategy) OnBar(b broker.Broker, bar data.Bar) {
	s.history = append(s.history, bar.Close)

	// Wait until we have enough data for the slow SMA
	if len(s.history) < s.slowPeriod {
		return
	}

	fastSMA := SMA(s.history, s.fastPeriod)
	slowSMA := SMA(s.history, s.slowPeriod)

	// We also need yesterday's SMAs to detect a true "cross"
	prevFastSMA := SMA(s.history[:len(s.history)-1], s.fastPeriod)
	prevSlowSMA := SMA(s.history[:len(s.history)-1], s.slowPeriod)

	pos := b.GetPosition(bar.Symbol)
	isLong := pos != nil && pos.Quantity > 0

	// Golden Cross: Fast crosses above Slow
	if prevFastSMA <= prevSlowSMA && fastSMA > slowSMA {
		if !isLong {
			cash := b.GetCash()
			affordable := cash * 0.95 // 95% of cash to buffer gap ups
			shares := affordable / bar.Close
			if shares > 0 {
				log.Printf("[SMACross] %s Golden Cross detected (Fast: %.2f > Slow: %.2f). Buying %f shares.", bar.Symbol, fastSMA, slowSMA, shares)
				b.SubmitOrder(bar.Symbol, shares)
			}
		}
	}

	// Death Cross: Fast crosses below Slow
	if prevFastSMA >= prevSlowSMA && fastSMA < slowSMA {
		if isLong {
			log.Printf("[SMACross] %s Death Cross detected (Fast: %.2f < Slow: %.2f). Closing position.", bar.Symbol, fastSMA, slowSMA)
			b.ClosePosition(bar.Symbol)
		}
	}
}
