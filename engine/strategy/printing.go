package strategy

import (
	"log"

	"github.com/samirkhairati/backtesting/engine/broker"
	"github.com/samirkhairati/backtesting/engine/data"
)

type PrintingStrategy struct {
	count int
}

func NewPrintingStrategy() *PrintingStrategy {
	return &PrintingStrategy{}
}

func (s *PrintingStrategy) OnBar(b broker.Broker, bar data.Bar) {
	s.count++

	if s.count <= 3 {
		log.Printf("[Strategy] Received Bar %d: %s [%s] O:%.2f H:%.2f L:%.2f C:%.2f V:%d",
			s.count, bar.Symbol, bar.Timestamp.Format("2006-01-02"),
			bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)
	}
}

func (s *PrintingStrategy) Count() int {
	return s.count
}
