package strategy

import (
	"github.com/samirkhairati/backtesting/core/broker"
	"github.com/samirkhairati/backtesting/core/data"
)

type meanReversionState struct {
	rsi *RSI
}

type MeanReversionStrategy struct {
	states map[string]*meanReversionState
}

func NewMeanReversionStrategy() *MeanReversionStrategy {
	return &MeanReversionStrategy{
		states: make(map[string]*meanReversionState),
	}
}

func (s *MeanReversionStrategy) OnBars(b broker.Broker, bars map[string]data.Bar) {
	for symbol, bar := range bars {
		state, exists := s.states[symbol]
		if !exists {
			state = &meanReversionState{
				rsi: NewRSI(14),
			}
			s.states[symbol] = state
		}

		rsiVal := state.rsi.Update(bar.Close)

		if rsiVal == 0 {
			continue // not enough data
		}

		pos := b.GetPosition(symbol)
		isLong := pos != nil && pos.Quantity > 0

		// Buy when oversold
		if rsiVal < 30 {
			if !isLong {
				cash := b.GetCash()
				affordable := cash * 0.20 // 20% of cash allocation per signal
				shares := affordable / bar.Close
				if shares > 0 {
					b.SubmitOrder(symbol, shares)
				}
			}
		}

		// Sell when overbought
		if rsiVal > 70 {
			if isLong {
				b.ClosePosition(symbol)
			}
		}
	}
}
