package strategy

import (
	"log"

	"github.com/samirkhairati/backtesting/engine/broker"
	"github.com/samirkhairati/backtesting/engine/data"
)

type bollingerState struct {
	sma *SMA
}

type BollingerBandsStrategy struct {
	states map[string]*bollingerState
}

func NewBollingerBandsStrategy() *BollingerBandsStrategy {
	return &BollingerBandsStrategy{
		states: make(map[string]*bollingerState),
	}
}

func (s *BollingerBandsStrategy) OnBars(b broker.Broker, bars map[string]data.Bar) {
	for symbol, bar := range bars {
		state, exists := s.states[symbol]
		if !exists {
			state = &bollingerState{
				sma: NewSMA(20),
			}
			s.states[symbol] = state
		}

		smaVal := state.sma.Update(bar.Close)

		if smaVal == 0 {
			continue
		}

		stdDev := state.sma.StdDev()
		upperBand := smaVal + (2.0 * stdDev)
		lowerBand := smaVal - (2.0 * stdDev)

		pos := b.GetPosition(symbol)
		isLong := pos != nil && pos.Quantity > 0

		// Buy when price drops below the lower band
		if bar.Close < lowerBand {
			if !isLong {
				cash := b.GetCash()
				affordable := cash * 0.20 // 20% of cash allocation per signal
				shares := affordable / bar.Close
				if shares > 0 {
					log.Printf("[BollingerBands] %s Oversold below Lower Band. Buying %f shares.", symbol, shares)
					b.SubmitOrder(symbol, shares)
				}
			}
		}

		// Sell when price reverts to the mean (SMA) or upper band
		if bar.Close >= smaVal {
			if isLong {
				log.Printf("[BollingerBands] %s Reverted to mean. Closing position.", symbol)
				b.ClosePosition(symbol)
			}
		}
		
		_ = upperBand // To prevent unused variable error if we don't use it for shorting
	}
}
