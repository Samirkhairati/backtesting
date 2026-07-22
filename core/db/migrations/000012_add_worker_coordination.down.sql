DROP INDEX IF EXISTS idx_backtest_jobs_worker_id;

ALTER TABLE backtest_jobs
DROP COLUMN IF EXISTS worker_id,
DROP COLUMN IF EXISTS retry_count;

DROP TABLE IF EXISTS workers;
