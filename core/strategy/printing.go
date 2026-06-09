package strategy

import (
	"github.com/samirkhairati/backtesting/core/logger"

	"github.com/samirkhairati/backtesting/core/broker"
	"github.com/samirkhairati/backtesting/core/data"
)

type PrintingStrategy struct {
	count int
}

func NewPrintingStrategy() *PrintingStrategy {
	return &PrintingStrategy{}
}

func (s *PrintingStrategy) OnBars(b broker.Broker, bars map[string]data.Bar) {
	s.count++

	if s.count <= 3 {
		for symbol, bar := range bars {
			logger.Log.Sugar().Infof("[Strategy] Received Bar %d: %s [%s] O:%.2f H:%.2f L:%.2f C:%.2f V:%d",
				s.count, symbol, bar.Timestamp.Format("2006-01-02"),
				bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)
		}
	}
}

func (s *PrintingStrategy) Count() int {
	return s.count
}
