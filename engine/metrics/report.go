package metrics

import (
	"fmt"
	"math"
	"strings"

	"github.com/samirkhairati/backtesting/engine/portfolio"
)

type PerformanceReport struct {
	TotalReturnPct float64
	CAGR           float64
	MaxDrawdownPct float64
	SharpeRatio    float64
	InitialValue   float64
	FinalValue     float64
	TradeAnalytics *TradeAnalytics
}

func GenerateReport(pf *portfolio.Portfolio) *PerformanceReport {
	if len(pf.EquityCurve) == 0 {
		return &PerformanceReport{}
	}

	initialValue := pf.EquityCurve[0].Value
	finalValue := pf.EquityCurve[len(pf.EquityCurve)-1].Value
	totalReturn := ((finalValue - initialValue) / initialValue) * 100.0

	// CAGR
	startTime := pf.EquityCurve[0].Timestamp
	endTime := pf.EquityCurve[len(pf.EquityCurve)-1].Timestamp
	years := endTime.Sub(startTime).Hours() / (24.0 * 365.25)
	cagr := 0.0
	if years > 0 && finalValue > 0 {
		cagr = (math.Pow(finalValue/initialValue, 1.0/years) - 1.0) * 100.0
	}

	// Max Drawdown & Daily Returns (for Sharpe using Welford's algorithm)
	maxEquity := initialValue
	maxDrawdown := 0.0
	
	// Welford's online algorithm variables for Sharpe Ratio
	n := 0.0
	meanReturn := 0.0
	M2 := 0.0

	for i := 1; i < len(pf.EquityCurve); i++ {
		current := pf.EquityCurve[i].Value
		prev := pf.EquityCurve[i-1].Value
		
		if current > maxEquity {
			maxEquity = current
		}
		
		drawdown := ((maxEquity - current) / maxEquity) * 100.0
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}

		if prev > 0 {
			r := (current - prev) / prev
			n += 1.0
			delta := r - meanReturn
			meanReturn += delta / n
			delta2 := r - meanReturn
			M2 += delta * delta2
		}
	}

	// Sharpe Ratio
	sharpe := 0.0
	if n > 0 {
		variance := M2 / n
		stdDev := math.Sqrt(variance)

		if stdDev > 0 {
			// Annualized Sharpe Ratio (assuming 252 trading days)
			sharpe = (meanReturn / stdDev) * math.Sqrt(252)
		}
	}

	analytics := GenerateTradeAnalytics(pf.ClosedTrades)

	return &PerformanceReport{
		TotalReturnPct: totalReturn,
		CAGR:           cagr,
		MaxDrawdownPct: maxDrawdown,
		SharpeRatio:    sharpe,
		InitialValue:   initialValue,
		FinalValue:     finalValue,
		TradeAnalytics: analytics,
	}
}

func (r *PerformanceReport) String() string {
	var sb strings.Builder
	sb.WriteString("======================================\n")
	sb.WriteString("       FINAL SIMULATION RESULTS       \n")
	sb.WriteString("======================================\n")
	sb.WriteString(fmt.Sprintf("Initial Value    : $%.2f\n", r.InitialValue))
	sb.WriteString(fmt.Sprintf("Final Value      : $%.2f\n", r.FinalValue))
	sb.WriteString(fmt.Sprintf("Total Return     : %.2f%%\n", r.TotalReturnPct))
	sb.WriteString(fmt.Sprintf("CAGR             : %.2f%%\n", r.CAGR))
	sb.WriteString(fmt.Sprintf("Max Drawdown     : %.2f%%\n", r.MaxDrawdownPct))
	sb.WriteString(fmt.Sprintf("Sharpe Ratio     : %.2f\n", r.SharpeRatio))
	sb.WriteString("--------------------------------------\n")
	sb.WriteString("           TRADE ANALYTICS            \n")
	sb.WriteString("--------------------------------------\n")
	sb.WriteString(fmt.Sprintf("Total Trades     : %d\n", r.TradeAnalytics.TotalTrades))
	sb.WriteString(fmt.Sprintf("Win Rate         : %.2f%%\n", r.TradeAnalytics.WinRatePct))
	sb.WriteString(fmt.Sprintf("Profit Factor    : %.2f\n", r.TradeAnalytics.ProfitFactor))
	sb.WriteString(fmt.Sprintf("Average Winner   : $%.2f\n", r.TradeAnalytics.AvgWinner))
	sb.WriteString(fmt.Sprintf("Average Loser    : $%.2f\n", r.TradeAnalytics.AvgLoser))
	sb.WriteString(fmt.Sprintf("Largest Winner   : $%.2f\n", r.TradeAnalytics.LargestWinner))
	sb.WriteString(fmt.Sprintf("Largest Loser    : $%.2f\n", r.TradeAnalytics.LargestLoser))
	sb.WriteString("======================================")
	return sb.String()
}
