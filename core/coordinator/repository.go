package coordinator

import (
	"context"
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

func (r *ExperimentRepository) GetMarketDataCoverage(ctx context.Context, symbol string) (time.Time, time.Time, error) {
	var covStart, covEnd time.Time
	err := r.DB.QueryRow(ctx, "SELECT start_date, end_date FROM market_data_coverage WHERE symbol = $1", symbol).Scan(&covStart, &covEnd)
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
	err = r.DB.QueryRow(ctx, "SELECT COUNT(*) FILTER (WHERE status='PENDING'), COUNT(*) FROM ingestion_jobs WHERE experiment_id=$1", experimentID).Scan(&pending, &total)
	return pending, total, err
}

func (r *ExperimentRepository) GetExperimentConfig(ctx context.Context, experimentID string) (string, error) {
	var config string
	err := r.DB.QueryRow(ctx, "SELECT config FROM experiments WHERE id=$1", experimentID).Scan(&config)
	return config, err
}

func (r *ExperimentRepository) InsertBacktestJob(ctx context.Context, experimentID, symbol, strategyName string, parameters []byte, startDate, endDate string) (string, error) {
	var jobID string
	err := r.DB.QueryRow(
		ctx,
		"INSERT INTO backtest_jobs (experiment_id, symbol, strategy_name, parameters, start_date, end_date, status) VALUES ($1, $2, $3, $4, $5, $6, 'PENDING') RETURNING id",
		experimentID,
		symbol,
		strategyName,
		parameters,
		startDate,
		endDate,
	).Scan(&jobID)
	return jobID, err
}

func (r *ExperimentRepository) UpdateBacktestJobStatus(ctx context.Context, jobID, status string, resultMetrics []byte) error {
	_, err := r.DB.Exec(ctx, "UPDATE backtest_jobs SET status=$1, completed_at=CURRENT_TIMESTAMP, result_metrics=$2 WHERE id=$3", status, resultMetrics, jobID)
	return err
}

func (r *ExperimentRepository) GetBacktestJobCounts(ctx context.Context, experimentID string) (pending int, total int, err error) {
	err = r.DB.QueryRow(ctx, "SELECT COUNT(*) FILTER (WHERE status='PENDING'), COUNT(*) FROM backtest_jobs WHERE experiment_id=$1", experimentID).Scan(&pending, &total)
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
