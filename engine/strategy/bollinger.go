package strategy

import (
	"log"

	"github.com/samirkhairati/backtesting/engine/broker"
	"github.com/samirkhairati/backtesting/engine/data"
)

type BollingerBandsStrategy struct {
	sma *SMA
}

func NewBollingerBandsStrategy() *BollingerBandsStrategy {
	return &BollingerBandsStrategy{
		sma: NewSMA(20),
	}
}

func (s *BollingerBandsStrategy) OnBar(b broker.Broker, bar data.Bar) {
	smaVal := s.sma.Update(bar.Close)

	if smaVal == 0 {
		return
	}

	stdDev := s.sma.StdDev()
	upperBand := smaVal + (2.0 * stdDev)
	lowerBand := smaVal - (2.0 * stdDev)

	pos := b.GetPosition(bar.Symbol)
	isLong := pos != nil && pos.Quantity > 0

	// Buy when price drops below the lower band
	if bar.Close < lowerBand {
		if !isLong {
			cash := b.GetCash()
			affordable := cash * 0.95
			shares := affordable / bar.Close
			if shares > 0 {
				log.Printf("[BollingerBands] %s Oversold below Lower Band. Buying %f shares.", bar.Symbol, shares)
				b.SubmitOrder(bar.Symbol, shares)
			}
		}
	}

	// Sell when price reverts to the mean (SMA) or upper band
	if bar.Close >= smaVal {
		if isLong {
			log.Printf("[BollingerBands] %s Reverted to mean. Closing position.", bar.Symbol)
			b.ClosePosition(bar.Symbol)
		}
	}
	
	_ = upperBand // To prevent unused variable error if we don't use it for shorting
}
