CREATE TABLE workers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    last_heartbeat TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE backtest_jobs
ADD COLUMN worker_id UUID REFERENCES workers(id) ON DELETE SET NULL,
ADD COLUMN retry_count INT NOT NULL DEFAULT 0;

CREATE INDEX idx_backtest_jobs_worker_id ON backtest_jobs(worker_id);
