package strategy

import (
	"github.com/samirkhairati/backtesting/engine/data"
)

// Strategy represents a trading strategy that can receive market data.
// In future phases, this interface will expand to allow producing signals or orders.
type Strategy interface {
	// OnBar is called by the replay engine every time a new historical bar is revealed.
	OnBar(bar data.Bar)
}
