DROP TABLE IF EXISTS backtest_jobs;

CREATE TABLE backtest_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    strategy_name VARCHAR(255) NOT NULL,
    symbols TEXT[] NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    result_metrics JSONB
);

CREATE INDEX idx_backtest_jobs_user_id ON backtest_jobs(user_id);
CREATE INDEX idx_backtest_jobs_status ON backtest_jobs(status);
