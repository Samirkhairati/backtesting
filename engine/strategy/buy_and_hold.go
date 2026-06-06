package strategy

import (
	"log"

	"github.com/samirkhairati/backtesting/engine/broker"
	"github.com/samirkhairati/backtesting/engine/data"
)

type BuyAndHoldStrategy struct {
	hasBought bool
}

func NewBuyAndHoldStrategy() *BuyAndHoldStrategy {
	return &BuyAndHoldStrategy{
		hasBought: false,
	}
}

func (s *BuyAndHoldStrategy) OnBar(b broker.Broker, bar data.Bar) {
	if !s.hasBought {
		// Calculate max shares we can afford with our current cash
		cash := b.GetCash()
		// Since we execute at the next day's open, we'll leave a tiny buffer to avoid overdrawing
		// if the open price gaps up slightly. We'll use 95% of cash.
		affordable := cash * 0.95
		shares := affordable / bar.Close

		if shares > 0 {
			b.SubmitOrder(bar.Symbol, shares)
			log.Printf("[BuyAndHold] Submitting order for %f shares of %s based on close $%.2f", shares, bar.Symbol, bar.Close)
			s.hasBought = true
		}
	}
}
