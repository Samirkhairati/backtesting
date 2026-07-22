package strategy

import (
	"github.com/samirkhairati/backtesting/core/broker"
	"github.com/samirkhairati/backtesting/core/data"
)

type macdState struct {
	fastEMA    *EMA
	slowEMA    *EMA
	signalEMA  *EMA
	prevMACD   float64
	prevSignal float64
	ready      bool
}

type MACDStrategy struct {
	states map[string]*macdState
}

func NewMACDStrategy() *MACDStrategy {
	return &MACDStrategy{
		states: make(map[string]*macdState),
	}
}

func (s *MACDStrategy) OnBars(b broker.Broker, bars map[string]data.Bar) {
	for symbol, bar := range bars {
		state, exists := s.states[symbol]
		if !exists {
			state = &macdState{
				fastEMA:   NewEMA(12),
				slowEMA:   NewEMA(26),
				signalEMA: NewEMA(9),
			}
			s.states[symbol] = state
		}

		fast := state.fastEMA.Update(bar.Close)
		slow := state.slowEMA.Update(bar.Close)

		if slow == 0 {
			continue
		}

		macdLine := fast - slow
		signalLine := state.signalEMA.Update(macdLine)

		if signalLine == 0 {
			continue
		}

		if !state.ready {
			state.prevMACD = macdLine
			state.prevSignal = signalLine
			state.ready = true
			continue
		}

		pos := b.GetPosition(symbol)
		isLong := pos != nil && pos.Quantity > 0

		// Bullish Cross
		if state.prevMACD <= state.prevSignal && macdLine > signalLine {
			if !isLong {
				cash := b.GetCash()
				affordable := cash * 0.20 // 20% of cash as an arbitrary allocation
				shares := affordable / bar.Close
				if shares > 0 {
					b.SubmitOrder(symbol, shares)
				}
			}
		}

		// Bearish Cross
		if state.prevMACD >= state.prevSignal && macdLine < signalLine {
			if isLong {
				b.ClosePosition(symbol)
			}
		}

		state.prevMACD = macdLine
		state.prevSignal = signalLine
	}
}
