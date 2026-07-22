package coordinator

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samirkhairati/backtesting/core/models"
)

// DBQuerier defines the interface for database queries
type DBQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type ExperimentRepository struct {
	DB DBQuerier
}

func NewExperimentRepository(db DBQuerier) *ExperimentRepository {
	return &ExperimentRepository{DB: db}
}

func (r *ExperimentRepository) InsertExperiment(ctx context.Context, userID, name, description, configJSON string) (string, error) {
	var experimentID string
	err := r.DB.QueryRow(
		ctx,
		"INSERT INTO experiments (user_id, name, description, config, status) VALUES ($1, $2, $3, $4, 'PENDING') RETURNING id",
		userID,
		name,
		description,
		configJSON,
	).Scan(&experimentID)
	return experimentID, err
}

func (r *ExperimentRepository) UpdateExperimentStatus(ctx context.Context, experimentID string, status models.ExperimentStatus) error {
	_, err := r.DB.Exec(ctx, "UPDATE experiments SET status = $1 WHERE id = $2", string(status), experimentID)
	return err
}

func (r *ExperimentRepository) TryTransitionExperimentStatus(ctx context.Context, experimentID string, fromStatus, toStatus models.ExperimentStatus) (bool, error) {
	tag, err := r.DB.Exec(ctx, "UPDATE experiments SET status = $1 WHERE id = $2 AND status = $3", string(toStatus), experimentID, string(fromStatus))
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *ExperimentRepository) GetMarketDataCoverage(ctx context.Context, symbol string) (string, string, error) {
	var covStart, covEnd string
	err := r.DB.QueryRow(ctx, "SELECT TO_CHAR(start_date, 'YYYY-MM-DD'), TO_CHAR(end_date, 'YYYY-MM-DD') FROM market_data_coverage WHERE symbol = $1", symbol).Scan(&covStart, &covEnd)
	return covStart, covEnd, err
}

func (r *ExperimentRepository) InsertIngestionJob(ctx context.Context, experimentID, symbol, startDate, endDate string) (string, error) {
	var jobID string
	err := r.DB.QueryRow(
		ctx,
		"INSERT INTO ingestion_jobs (experiment_id, symbol, start_date, end_date, status) VALUES ($1, $2, $3, $4, 'PENDING') RETURNING id",
		experimentID,
		symbol,
		startDate,
		endDate,
	).Scan(&jobID)
	return jobID, err
}

func (r *ExperimentRepository) UpdateIngestionJobStatus(ctx context.Context, jobID string, status models.IngestionJobStatus) error {
	_, err := r.DB.Exec(ctx, "UPDATE ingestion_jobs SET status=$1, completed_at=CURRENT_TIMESTAMP WHERE id=$2", string(status), jobID)
	return err
}

func (r *ExperimentRepository) GetIngestionJobCounts(ctx context.Context, experimentID string) (pending int, total int, err error) {
	err = r.DB.QueryRow(ctx, "SELECT COUNT(*) FILTER (WHERE status IN ('PENDING', 'RUNNING')), COUNT(*) FROM ingestion_jobs WHERE experiment_id=$1", experimentID).Scan(&pending, &total)
	return pending, total, err
}

func (r *ExperimentRepository) GetExperimentConfig(ctx context.Context, experimentID string) (string, error) {
	var config string
	err := r.DB.QueryRow(ctx, "SELECT config FROM experiments WHERE id=$1", experimentID).Scan(&config)
	return config, err
}

func (r *ExperimentRepository) InsertBacktestJob(ctx context.Context, experimentID, symbol, strategyName string, parameters []byte, startDate, endDate string) (string, error) {
	var jobID string
	// ON CONFLICT DO NOTHING makes job generation idempotent: if the coordinator
	// restarts and GenerateBacktestJobs is called again, duplicate jobs are silently skipped.
	err := r.DB.QueryRow(
		ctx,
		`INSERT INTO backtest_jobs (experiment_id, symbol, strategy_name, parameters, start_date, end_date, status) 
		 VALUES ($1, $2, $3, $4, $5, $6, 'PENDING') 
		 ON CONFLICT ON CONSTRAINT uq_backtest_job DO NOTHING
		 RETURNING id`,
		experimentID,
		symbol,
		strategyName,
		parameters,
		startDate,
		endDate,
	).Scan(&jobID)
	if err != nil && err.Error() == "no rows in result set" {
		// Conflict — job already exists, return empty ID (caller should skip publishing)
		return "", nil
	}
	return jobID, err
}

func (r *ExperimentRepository) UpdateBacktestJobStatus(ctx context.Context, jobID, status string, resultMetrics []byte) error {
	_, err := r.DB.Exec(ctx, "UPDATE backtest_jobs SET status=$1, completed_at=CURRENT_TIMESTAMP, result_metrics=$2 WHERE id=$3", status, resultMetrics, jobID)
	return err
}

func (r *ExperimentRepository) GetBacktestJobCounts(ctx context.Context, experimentID string) (pending int, total int, err error) {
	err = r.DB.QueryRow(ctx, "SELECT COUNT(*) FILTER (WHERE status IN ('PENDING', 'RUNNING')), COUNT(*) FROM backtest_jobs WHERE experiment_id=$1", experimentID).Scan(&pending, &total)
	return pending, total, err
}

func (r *ExperimentRepository) GetBacktestJobMetrics(ctx context.Context, experimentID string) ([]models.BacktestJob, error) {
	rows, err := r.DB.Query(ctx, "SELECT symbol, strategy_name, result_metrics FROM backtest_jobs WHERE experiment_id=$1 AND status='COMPLETED'", experimentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.BacktestJob
	for rows.Next() {
		var job models.BacktestJob
		if err := rows.Scan(&job.Symbol, &job.StrategyName, &job.ResultMetrics); err != nil {
			continue
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (r *ExperimentRepository) UpdateExperimentResults(ctx context.Context, experimentID string, resultSummary map[string]interface{}, equityCurves map[string]interface{}) error {
	_, err := r.DB.Exec(ctx, "UPDATE experiments SET result_summary=$1, equity_curves=$2 WHERE id=$3", resultSummary, equityCurves, experimentID)
	return err
}

func (r *ExperimentRepository) UpsertWorkerHeartbeat(ctx context.Context, workerID string, status string) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO workers (id, status, last_heartbeat) 
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE 
		SET last_heartbeat = CURRENT_TIMESTAMP, status = $2
	`, workerID, status)
	return err
}

func (r *ExperimentRepository) ClaimJob(ctx context.Context, jobID string, workerID string) error {
	// Upsert the worker first so the FK constraint on backtest_jobs.worker_id is satisfied
	// even if the heartbeat hasn't arrived yet.
	_, err := r.DB.Exec(ctx, `
		INSERT INTO workers (id, status, last_heartbeat)
		VALUES ($1, 'ACTIVE', CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE
		SET last_heartbeat = CURRENT_TIMESTAMP, status = 'ACTIVE'
	`, workerID)
	if err != nil {
		return fmt.Errorf("ClaimJob: failed to upsert worker: %w", err)
	}
	_, err = r.DB.Exec(ctx, "UPDATE backtest_jobs SET status='RUNNING', worker_id=$1 WHERE id=$2 AND status IN ('PENDING', 'RUNNING')", workerID, jobID)
	return err
}

func (r *ExperimentRepository) GetDeadWorkers(ctx context.Context, timeoutSeconds int) ([]string, error) {
	intervalStr := fmt.Sprintf("%d seconds", timeoutSeconds)
	rows, err := r.DB.Query(ctx, `
		SELECT id FROM workers 
		WHERE status = 'ACTIVE' 
		AND last_heartbeat < CURRENT_TIMESTAMP - $1::interval
	`, intervalStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workerIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			workerIDs = append(workerIDs, id)
		}
	}
	return workerIDs, nil
}

func (r *ExperimentRepository) RequeueJobsForWorker(ctx context.Context, workerID string) ([]models.BacktestJobMessage, []models.BacktestJobMessage, error) {
	// Atomically mark the worker UNHEALTHY and requeue its RUNNING jobs in one step.
	// The worker status and job requeue are done in separate statements intentionally:
	// the UPDATE ... RETURNING on backtest_jobs is atomic itself (locks rows), so
	// a late heartbeat arriving between the two statements cannot cause double-execution
	// because the job rows are already locked/updated.
	_, err := r.DB.Exec(ctx, "UPDATE workers SET status='UNHEALTHY' WHERE id=$1 AND status='ACTIVE'", workerID)
	if err != nil {
		return nil, nil, err
	}

	// Requeue RUNNING jobs with retry headroom; increment retry_count atomically.
	rows, err := r.DB.Query(ctx, `
		UPDATE backtest_jobs 
		SET status='PENDING', worker_id=NULL, retry_count=retry_count+1 
		WHERE worker_id=$1 AND status='RUNNING' AND retry_count < 3
		RETURNING id, experiment_id, symbol, strategy_name, parameters, start_date, end_date
	`, workerID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var requeuedJobs []models.BacktestJobMessage
	for rows.Next() {
		var job models.BacktestJobMessage
		var start, end time.Time
		if err := rows.Scan(&job.JobID, &job.ExperimentID, &job.Symbol, &job.StrategyName, &job.Parameters, &start, &end); err != nil {
			continue
		}
		job.StartDate = start.Format(time.RFC3339)
		job.EndDate = end.Format(time.RFC3339)
		requeuedJobs = append(requeuedJobs, job)
	}

	// Permanently fail jobs that exceeded the retry limit.
	var failedJobs []models.BacktestJobMessage
	rows2, err := r.DB.Query(ctx, "UPDATE backtest_jobs SET status='FAILED' WHERE worker_id=$1 AND status='RUNNING' RETURNING id, experiment_id, symbol", workerID)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var job models.BacktestJobMessage
			if err := rows2.Scan(&job.JobID, &job.ExperimentID, &job.Symbol); err == nil {
				failedJobs = append(failedJobs, job)
			}
		}
	}

	return requeuedJobs, failedJobs, nil
}

func (r *ExperimentRepository) ClaimIngestionJob(ctx context.Context, jobID string, workerID string) error {
	// Upsert the worker first so the FK constraint on ingestion_jobs.worker_id is satisfied
	// even if the heartbeat hasn't arrived yet.
	_, err := r.DB.Exec(ctx, `
		INSERT INTO workers (id, status, last_heartbeat)
		VALUES ($1, 'ACTIVE', CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE
		SET last_heartbeat = CURRENT_TIMESTAMP, status = 'ACTIVE'
	`, workerID)
	if err != nil {
		return fmt.Errorf("ClaimIngestionJob: failed to upsert worker: %w", err)
	}
	_, err = r.DB.Exec(ctx, "UPDATE ingestion_jobs SET status='RUNNING', worker_id=$1 WHERE id=$2 AND status IN ('PENDING', 'RUNNING')", workerID, jobID)
	return err
}

func (r *ExperimentRepository) RequeueIngestionJobsForWorker(ctx context.Context, workerID string) ([]models.IngestionJobMessage, []models.IngestionJobMessage, error) {
	// We don't mark worker UNHEALTHY again here since RequeueJobsForWorker already does it, 
	// or we can just ensure the caller handles worker state correctly.

	// Find running jobs for this worker with retry_count < 3
	rows, err := r.DB.Query(ctx, `
		UPDATE ingestion_jobs 
		SET status='PENDING', worker_id=NULL, retry_count=retry_count+1 
		WHERE worker_id=$1 AND status='RUNNING' AND retry_count < 3
		RETURNING id, experiment_id, symbol, start_date, end_date
	`, workerID)
	
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var requeuedJobs []models.IngestionJobMessage
	for rows.Next() {
		var job models.IngestionJobMessage
		var start, end time.Time
		if err := rows.Scan(&job.JobID, &job.ExperimentID, &job.Symbol, &start, &end); err != nil {
			continue
		}
		job.StartDate = start.Format(time.RFC3339)
		job.EndDate = end.Format(time.RFC3339)
		requeuedJobs = append(requeuedJobs, job)
	}
	
	// Fail the remaining jobs that exceeded retry count
	var failedJobs []models.IngestionJobMessage
	rows2, err := r.DB.Query(ctx, "UPDATE ingestion_jobs SET status='FAILED' WHERE worker_id=$1 AND status='RUNNING' RETURNING id, experiment_id, symbol", workerID)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var job models.IngestionJobMessage
			if err := rows2.Scan(&job.JobID, &job.ExperimentID, &job.Symbol); err == nil {
				failedJobs = append(failedJobs, job)
			}
		}
	}
	
	return requeuedJobs, failedJobs, nil
}

func (r *ExperimentRepository) CancelExperimentJobs(ctx context.Context, experimentID string) error {
	// Cancel all non-terminal ingestion jobs (PENDING and RUNNING).
	// Cancelling RUNNING jobs prevents workers from publishing completion events
	// that would incorrectly try to transition a CANCELLED experiment to AGGREGATING.
	if _, err := r.DB.Exec(ctx, "UPDATE ingestion_jobs SET status = 'CANCELLED' WHERE experiment_id = $1 AND status IN ('PENDING', 'RUNNING')", experimentID); err != nil {
		return err
	}

	// Cancel all non-terminal backtest jobs.
	if _, err := r.DB.Exec(ctx, "UPDATE backtest_jobs SET status = 'CANCELLED' WHERE experiment_id = $1 AND status IN ('PENDING', 'RUNNING')", experimentID); err != nil {
		return err
	}

	// Mark the experiment as CANCELLED (guard against overwriting COMPLETED).
	if _, err := r.DB.Exec(ctx, "UPDATE experiments SET status = 'CANCELLED' WHERE id = $1 AND status != 'COMPLETED'", experimentID); err != nil {
		return err
	}

	return nil
}
