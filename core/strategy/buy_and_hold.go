package strategy

import (
	"github.com/samirkhairati/backtesting/core/logger"

	"github.com/samirkhairati/backtesting/core/broker"
	"github.com/samirkhairati/backtesting/core/data"
)

type BuyAndHoldStrategy struct {
	hasBought bool
}

func NewBuyAndHoldStrategy() *BuyAndHoldStrategy {
	return &BuyAndHoldStrategy{
		hasBought: false,
	}
}

func (s *BuyAndHoldStrategy) OnBars(b broker.Broker, bars map[string]data.Bar) {
	if !s.hasBought && len(bars) > 0 {
		cash := b.GetCash()
		// 95% buffer to avoid overdrawing
		affordable := cash * 0.95

		// Split capital evenly among all available assets
		allocation := affordable / float64(len(bars))

		for symbol, bar := range bars {
			shares := allocation / bar.Close
			if shares > 0 {
				b.SubmitOrder(symbol, shares)
				logger.Log.Sugar().Infof("[BuyAndHold] Submitting order for %f shares of %s based on close $%.2f", shares, symbol, bar.Close)
			}
		}
		s.hasBought = true
	}
}
