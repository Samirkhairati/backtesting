package strategy

import (
	"log"

	"github.com/samirkhairati/backtesting/engine/broker"
	"github.com/samirkhairati/backtesting/engine/data"
)

type MeanReversionStrategy struct {
	rsi *RSI
}

func NewMeanReversionStrategy() *MeanReversionStrategy {
	return &MeanReversionStrategy{
		rsi: NewRSI(14),
	}
}

func (s *MeanReversionStrategy) OnBar(b broker.Broker, bar data.Bar) {
	rsiVal := s.rsi.Update(bar.Close)

	if rsiVal == 0 {
		return // not enough data
	}

	pos := b.GetPosition(bar.Symbol)
	isLong := pos != nil && pos.Quantity > 0

	// Buy when oversold
	if rsiVal < 30 {
		if !isLong {
			cash := b.GetCash()
			affordable := cash * 0.95
			shares := affordable / bar.Close
			if shares > 0 {
				log.Printf("[MeanReversion] %s RSI Oversold (%.2f). Buying %f shares.", bar.Symbol, rsiVal, shares)
				b.SubmitOrder(bar.Symbol, shares)
			}
		}
	}

	// Sell when overbought
	if rsiVal > 70 {
		if isLong {
			log.Printf("[MeanReversion] %s RSI Overbought (%.2f). Closing position.", bar.Symbol, rsiVal)
			b.ClosePosition(bar.Symbol)
		}
	}
}
