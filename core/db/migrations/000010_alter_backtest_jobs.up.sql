DROP TABLE IF EXISTS backtest_jobs;

CREATE TABLE backtest_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    symbol VARCHAR(50) NOT NULL,
    strategy_name VARCHAR(255) NOT NULL,
    parameters JSONB NOT NULL,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    result_metrics JSONB
);

CREATE INDEX idx_backtest_jobs_experiment_id ON backtest_jobs(experiment_id);
CREATE INDEX idx_backtest_jobs_status ON backtest_jobs(status);
