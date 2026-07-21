package coordinator

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/db"
)

func TestAggregation_Integration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	logger.Init()
	ctx := context.Background()

	dbPool := db.Setup(ctx)
	defer dbPool.Close()

	repo := NewExperimentRepository(dbPool)

	// 1. Setup a dummy user to satisfy foreign key constraints.
	// In some DB schemas, users table might not exist or might be managed externally.
	// We will attempt to insert a user, ignoring conflicts.
	userID := "b8c1d5e3-1b91-4501-a4b5-9c8a9e0f1a2b"
	_, _ = dbPool.Exec(ctx, "INSERT INTO users (id, username, password_hash) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", userID, "test_user", "hash")

	// 2. Create an experiment
	expID, err := repo.InsertExperiment(ctx, userID, "Agg Integration Test", "Desc", "{}")
	if err != nil {
		t.Fatalf("Failed to insert experiment: %v", err)
	}

	// 3. Create a couple of backtest jobs
	job1Metrics := map[string]interface{}{
		"TotalReturn":      10.0,
		"AnnualizedReturn": 5.0,
		"MaxDrawdown":      2.0,
		"SharpeRatio":      1.5,
		"WinRate":          60.0,
		"TotalTrades":      100.0,
		"EquityCurve":      []float64{1.0, 1.1},
	}
	job1MetricsJSON, _ := json.Marshal(job1Metrics)

	job1ID, err := repo.InsertBacktestJob(ctx, expID, "AAPL", "SMA", []byte(`{}`), "2020-01-01", "2020-12-31")
	if err != nil {
		t.Fatalf("Failed to insert backtest job 1: %v", err)
	}
	err = repo.UpdateBacktestJobStatus(ctx, job1ID, "COMPLETED", job1MetricsJSON)
	if err != nil {
		t.Fatalf("Failed to update backtest job 1 status: %v", err)
	}

	job2Metrics := map[string]interface{}{
		"TotalReturn":      20.0,
		"AnnualizedReturn": 15.0,
		"MaxDrawdown":      4.0,
		"SharpeRatio":      2.5,
		"WinRate":          70.0,
		"TotalTrades":      200.0,
		"EquityCurve":      []float64{1.0, 1.2},
	}
	job2MetricsJSON, _ := json.Marshal(job2Metrics)

	job2ID, err := repo.InsertBacktestJob(ctx, expID, "MSFT", "SMA", []byte(`{}`), "2020-01-01", "2020-12-31")
	if err != nil {
		t.Fatalf("Failed to insert backtest job 2: %v", err)
	}
	err = repo.UpdateBacktestJobStatus(ctx, job2ID, "COMPLETED", job2MetricsJSON)
	if err != nil {
		t.Fatalf("Failed to update backtest job 2 status: %v", err)
	}

	// 4. Fetch the jobs from DB and Aggregate
	jobs, err := repo.GetBacktestJobMetrics(ctx, expID)
	if err != nil {
		t.Fatalf("Failed to fetch backtest job metrics: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("Expected 2 jobs, got %d", len(jobs))
	}

	resultSummary, equityCurves := AggregateJobMetrics(jobs)

	// 5. Update experiment results
	err = repo.UpdateExperimentResults(ctx, expID, resultSummary, equityCurves)
	if err != nil {
		t.Fatalf("Failed to update experiment results: %v", err)
	}

	// 6. Verify that the experiment now has the correct results
	var resultSummaryJSON, equityCurvesJSON []byte
	err = dbPool.QueryRow(ctx, "SELECT result_summary, equity_curves FROM experiments WHERE id = $1", expID).Scan(&resultSummaryJSON, &equityCurvesJSON)
	if err != nil {
		t.Fatalf("Failed to fetch updated experiment: %v", err)
	}

	if len(resultSummaryJSON) == 0 || len(equityCurvesJSON) == 0 {
		t.Fatalf("Expected results and equity curves to be populated in experiment, but got empty arrays")
	}
	
	// Optional Cleanup
	_, _ = dbPool.Exec(ctx, "DELETE FROM backtest_jobs WHERE experiment_id = $1", expID)
	_, _ = dbPool.Exec(ctx, "DELETE FROM experiments WHERE id = $1", expID)
}
