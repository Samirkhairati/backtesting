package strategy

import (
	"github.com/samirkhairati/backtesting/engine/broker"
	"github.com/samirkhairati/backtesting/engine/data"
)

// Strategy defines the interface for trading algorithms
type Strategy interface {
	OnBar(broker broker.Broker, bar data.Bar)
}
