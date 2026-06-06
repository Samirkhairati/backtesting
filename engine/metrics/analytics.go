package metrics

import "github.com/samirkhairati/backtesting/engine/portfolio"

type TradeAnalytics struct {
	TotalTrades   int
	WinningTrades int
	LosingTrades  int
	WinRatePct    float64
	GrossProfit   float64
	GrossLoss     float64
	ProfitFactor  float64
	AvgWinner     float64
	AvgLoser      float64
	LargestWinner float64
	LargestLoser  float64
}

func GenerateTradeAnalytics(trades []portfolio.Trade) *TradeAnalytics {
	if len(trades) == 0 {
		return &TradeAnalytics{}
	}

	a := &TradeAnalytics{
		TotalTrades: len(trades),
	}

	for _, t := range trades {
		if t.PnL > 0 {
			a.WinningTrades++
			a.GrossProfit += t.PnL
			if t.PnL > a.LargestWinner {
				a.LargestWinner = t.PnL
			}
		} else if t.PnL < 0 {
			a.LosingTrades++
			loss := -t.PnL
			a.GrossLoss += loss
			if loss > a.LargestLoser {
				a.LargestLoser = loss
			}
		}
	}

	if a.TotalTrades > 0 {
		a.WinRatePct = (float64(a.WinningTrades) / float64(a.TotalTrades)) * 100.0
	}

	if a.GrossLoss > 0 {
		a.ProfitFactor = a.GrossProfit / a.GrossLoss
	} else if a.GrossProfit > 0 {
		a.ProfitFactor = 999.99 // Represents infinite profit factor
	}

	if a.WinningTrades > 0 {
		a.AvgWinner = a.GrossProfit / float64(a.WinningTrades)
	}

	if a.LosingTrades > 0 {
		a.AvgLoser = a.GrossLoss / float64(a.LosingTrades)
	}

	return a
}
