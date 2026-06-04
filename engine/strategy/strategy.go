package strategy

import (
	"github.com/samirkhairati/backtesting/engine/data"
)

type Strategy interface {
	OnBar(bar data.Bar)
}
