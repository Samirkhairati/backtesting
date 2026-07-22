-- Add a unique constraint so GenerateBacktestJobs is idempotent.
-- On conflict (duplicate job), do nothing -- the existing job is preserved.
ALTER TABLE backtest_jobs
    ADD CONSTRAINT uq_backtest_job
    UNIQUE (experiment_id, symbol, strategy_name, parameters, start_date, end_date);
