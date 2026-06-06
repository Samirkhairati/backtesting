package strategy

import (
	"log"

	"github.com/samirkhairati/backtesting/engine/broker"
	"github.com/samirkhairati/backtesting/engine/data"
)

type MACDStrategy struct {
	fastEMA     *EMA
	slowEMA     *EMA
	signalEMA   *EMA
	prevMACD    float64
	prevSignal  float64
	ready       bool
}

func NewMACDStrategy() *MACDStrategy {
	return &MACDStrategy{
		fastEMA:   NewEMA(12),
		slowEMA:   NewEMA(26),
		signalEMA: NewEMA(9),
	}
}

func (s *MACDStrategy) OnBar(b broker.Broker, bar data.Bar) {
	fast := s.fastEMA.Update(bar.Close)
	slow := s.slowEMA.Update(bar.Close)

	if slow == 0 {
		return
	}

	macdLine := fast - slow
	signalLine := s.signalEMA.Update(macdLine)

	if signalLine == 0 {
		return
	}

	if !s.ready {
		s.prevMACD = macdLine
		s.prevSignal = signalLine
		s.ready = true
		return
	}

	pos := b.GetPosition(bar.Symbol)
	isLong := pos != nil && pos.Quantity > 0

	// Bullish Cross
	if s.prevMACD <= s.prevSignal && macdLine > signalLine {
		if !isLong {
			cash := b.GetCash()
			affordable := cash * 0.95
			shares := affordable / bar.Close
			if shares > 0 {
				log.Printf("[MACD] %s Bullish Cross. Buying %f shares.", bar.Symbol, shares)
				b.SubmitOrder(bar.Symbol, shares)
			}
		}
	}

	// Bearish Cross
	if s.prevMACD >= s.prevSignal && macdLine < signalLine {
		if isLong {
			log.Printf("[MACD] %s Bearish Cross. Closing position.", bar.Symbol)
			b.ClosePosition(bar.Symbol)
		}
	}

	s.prevMACD = macdLine
	s.prevSignal = signalLine
}
