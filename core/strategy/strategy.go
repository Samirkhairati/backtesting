package strategy

import (
	"github.com/samirkhairati/backtesting/core/broker"
	"github.com/samirkhairati/backtesting/core/data"
)

// Strategy defines the interface for trading algorithms
type Strategy interface {
	OnBars(broker broker.Broker, bars map[string]data.Bar)
}
