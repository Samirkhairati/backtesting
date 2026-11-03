package coordinator

import (
	"context"
	"testing"
	"os"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/samirkhairati/backtesting/core/db"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/models"
)

func TestWorkerHeartbeatAndRequeue(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	logger.Init()
	ctx := context.Background()
	testDB := db.Setup(ctx)
	defer testDB.Close()

	repo := NewExperimentRepository(testDB)

	// 1. Insert user and experiment
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, username, password_hash) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING", "00000000-0000-0000-0000-000000000000", "requeue_user", "hash")
	require.NoError(t, err)

	expID, err := repo.InsertExperiment(ctx, "00000000-0000-0000-0000-000000000000", "Test Requeue", "desc", "{}")
	require.NoError(t, err)

	// 2. Insert backtest job
	jobID, err := repo.InsertBacktestJob(ctx, expID, "AAPL", "SMA", []byte(`{}`), "2023-01-01", "2023-02-01")
	require.NoError(t, err)

	workerID := uuid.New().String()

	// 3. Upsert heartbeat (worker is alive)
	err = repo.UpsertWorkerHeartbeat(ctx, workerID, "ACTIVE", "backtesting")
	require.NoError(t, err)

	// 4. Worker claims job
	err = repo.ClaimJob(ctx, jobID, workerID)
	require.NoError(t, err)

	// 5. Update worker last_heartbeat to be in the past (to simulate death)
	_, err = testDB.Exec(ctx, "UPDATE workers SET last_heartbeat = CURRENT_TIMESTAMP - interval '30 seconds' WHERE id = $1", workerID)
	require.NoError(t, err)

	// 6. Check for dead workers
	deadWorkers, err := repo.GetDeadWorkers(ctx, 15)
	require.NoError(t, err)
	assert.Contains(t, deadWorkers, workerID)

	// 7. Requeue jobs for dead worker
	requeued, failed, err := repo.RequeueJobsForWorker(ctx, workerID)
	require.NoError(t, err)

	assert.Len(t, requeued, 1)
	assert.Len(t, failed, 0)
	assert.Equal(t, jobID, requeued[0].JobID)

	// 8. Verify job status is now PENDING and retry_count is 1
	var requeuedJobs []models.BacktestJob
	rows, _ := testDB.Query(ctx, "SELECT id, experiment_id, symbol, strategy_name, parameters, start_date, end_date FROM backtest_jobs WHERE status = 'PENDING' AND experiment_id = $1", expID)
	for rows.Next() {
		var job models.BacktestJob
		_ = rows.Scan(&job.ID, &job.ExperimentID, &job.Symbol, &job.StrategyName, &job.Parameters, &job.StartDate, &job.EndDate)
		requeuedJobs = append(requeuedJobs, job)
	}
	assert.Len(t, requeuedJobs, 1)

	var status string
	var retryCount int
	err = testDB.QueryRow(ctx, "SELECT status, retry_count FROM backtest_jobs WHERE id = $1", jobID).Scan(&status, &retryCount)
	require.NoError(t, err)

	assert.Equal(t, "PENDING", status)
	assert.Equal(t, 1, retryCount)
}

func TestIngestionWorkerHeartbeatAndRequeue(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	logger.Init()
	ctx := context.Background()
	testDB := db.Setup(ctx)
	defer testDB.Close()

	repo := NewExperimentRepository(testDB)

	// 1. Insert user and experiment
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, username, password_hash) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING", "00000000-0000-0000-0000-000000000000", "requeue_user_2", "hash")
	require.NoError(t, err)

	expID, err := repo.InsertExperiment(ctx, "00000000-0000-0000-0000-000000000000", "Test Ingestion Requeue", "desc", "{}")
	require.NoError(t, err)

	// 2. Insert ingestion job
	jobID, err := repo.InsertIngestionJob(ctx, expID, "GOOGL", "2023-01-01", "2023-02-01")
	require.NoError(t, err)

	workerID := uuid.New().String()

	// 3. Upsert heartbeat (worker is alive)
	err = repo.UpsertWorkerHeartbeat(ctx, workerID, "ACTIVE", "backtesting")
	require.NoError(t, err)

	// 4. Worker claims job
	err = repo.ClaimIngestionJob(ctx, jobID, workerID)
	require.NoError(t, err)

	// 5. Update worker last_heartbeat to be in the past (to simulate death)
	_, err = testDB.Exec(ctx, "UPDATE workers SET last_heartbeat = CURRENT_TIMESTAMP - interval '30 seconds' WHERE id = $1", workerID)
	require.NoError(t, err)

	// 6. Check for dead workers
	deadWorkers, err := repo.GetDeadWorkers(ctx, 15)
	require.NoError(t, err)
	assert.Contains(t, deadWorkers, workerID)

	// 7. Requeue jobs for dead worker
	requeued, failed, err := repo.RequeueIngestionJobsForWorker(ctx, workerID)
	require.NoError(t, err)

	assert.Len(t, requeued, 1)
	assert.Len(t, failed, 0)
	assert.Equal(t, jobID, requeued[0].JobID)

	// 8. Verify job status is now PENDING and retry_count is 1
	var status string
	var retryCount int
	err = testDB.QueryRow(ctx, "SELECT status, retry_count FROM ingestion_jobs WHERE id = $1", jobID).Scan(&status, &retryCount)
	require.NoError(t, err)

	assert.Equal(t, "PENDING", status)
	assert.Equal(t, 1, retryCount)
}
