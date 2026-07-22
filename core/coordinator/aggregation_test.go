package coordinator

import (
	"testing"
	"github.com/samirkhairati/backtesting/core/models"
)

func TestAggregateJobMetrics(t *testing.T) {
	tests := []struct {
		name          string
		jobs          []models.BacktestJob
		expectedSum   map[string]float64
		expectedCount int
	}{
		{
			name: "Zero Jobs",
			jobs: []models.BacktestJob{},
			expectedSum: map[string]float64{
				"TotalReturn":      0.0,
				"AnnualizedReturn": 0.0,
				"MaxDrawdown":      0.0,
				"SharpeRatio":      0.0,
				"WinRate":          0.0,
				"TotalTrades":      0.0,
			},
			expectedCount: 0,
		},
		{
			name: "Single Job",
			jobs: []models.BacktestJob{
				{
					Symbol:       "AAPL",
					StrategyName: "SMA",
					ResultMetrics: map[string]interface{}{
						"TotalReturn":      10.0,
						"AnnualizedReturn": 5.0,
						"MaxDrawdown":      2.0,
						"SharpeRatio":      1.5,
						"WinRate":          60.0,
						"TotalTrades":      100.0,
						"EquityCurve":      []interface{}{1.0, 1.1},
					},
				},
			},
			expectedSum: map[string]float64{
				"TotalReturn":      10.0,
				"AnnualizedReturn": 5.0,
				"MaxDrawdown":      2.0,
				"SharpeRatio":      1.5,
				"WinRate":          60.0,
				"TotalTrades":      100.0, // trades are summed
			},
			expectedCount: 1,
		},
		{
			name: "Multiple Jobs",
			jobs: []models.BacktestJob{
				{
					Symbol:       "AAPL",
					StrategyName: "SMA",
					ResultMetrics: map[string]interface{}{
						"TotalReturn":      10.0,
						"AnnualizedReturn": 5.0,
						"MaxDrawdown":      2.0,
						"SharpeRatio":      1.5,
						"WinRate":          60.0,
						"TotalTrades":      100.0,
					},
				},
				{
					Symbol:       "MSFT",
					StrategyName: "SMA",
					ResultMetrics: map[string]interface{}{
						"TotalReturn":      20.0,
						"AnnualizedReturn": 15.0,
						"MaxDrawdown":      4.0,
						"SharpeRatio":      2.5,
						"WinRate":          70.0,
						"TotalTrades":      200.0,
					},
				},
			},
			expectedSum: map[string]float64{
				"TotalReturn":      15.0,  // (10+20)/2
				"AnnualizedReturn": 10.0,  // (5+15)/2
				"MaxDrawdown":      3.0,   // (2+4)/2
				"SharpeRatio":      2.0,   // (1.5+2.5)/2
				"WinRate":          65.0,  // (60+70)/2
				"TotalTrades":      300.0, // 100+200
			},
			expectedCount: 2,
		},
		{
			name: "Missing Metrics",
			jobs: []models.BacktestJob{
				{
					Symbol:       "AAPL",
					StrategyName: "SMA",
					ResultMetrics: map[string]interface{}{
						"TotalReturn": 10.0,
					},
				},
				{
					Symbol:       "MSFT",
					StrategyName: "SMA",
					ResultMetrics: map[string]interface{}{
						"TotalReturn": 20.0,
						// missing other metrics, should default to 0 for them
					},
				},
			},
			expectedSum: map[string]float64{
				"TotalReturn":      15.0, // (10+20)/2
				"AnnualizedReturn": 0.0,  // (0+0)/2
				"MaxDrawdown":      0.0,  // (0+0)/2
				"SharpeRatio":      0.0,  // (0+0)/2
				"WinRate":          0.0,  // (0+0)/2
				"TotalTrades":      0.0,  // 0+0
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, equityCurves := AggregateJobMetrics(tt.jobs, nil)

			// Check averaged values and sum
			for key, expectedValue := range tt.expectedSum {
				if val, ok := summary[key].(float64); ok {
					if val != expectedValue {
						t.Errorf("expected %s to be %f, got %f", key, expectedValue, val)
					}
				} else {
					t.Errorf("missing key %s in result summary or not a float64", key)
				}
			}

			// Check JobMetrics size
			if jobMetrics, ok := summary["JobMetrics"].([]map[string]interface{}); ok {
				if len(jobMetrics) != tt.expectedCount {
					t.Errorf("expected JobMetrics length to be %d, got %d", tt.expectedCount, len(jobMetrics))
				}
			} else {
				t.Errorf("JobMetrics missing or not of expected type")
			}
			
			// Check EquityCurves for Single Job
			if tt.name == "Single Job" {
				if _, ok := equityCurves["AAPL - SMA"]; !ok {
					t.Errorf("expected equity curve for 'AAPL - SMA' to be present")
				}
			}
		})
	}
}
