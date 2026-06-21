CREATE TABLE IF NOT EXISTS experiment_status_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_experiment_status_logs_experiment_id ON experiment_status_logs(experiment_id);

CREATE OR REPLACE FUNCTION log_experiment_status() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' OR NEW.status <> OLD.status THEN
        INSERT INTO experiment_status_logs (experiment_id, status) VALUES (NEW.id, NEW.status);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER experiment_status_trigger
AFTER INSERT OR UPDATE OF status ON experiments
FOR EACH ROW EXECUTE FUNCTION log_experiment_status();

-- Insert initial logs for existing experiments
INSERT INTO experiment_status_logs (experiment_id, status, created_at)
SELECT id, 'PENDING', created_at FROM experiments;

-- If an experiment is beyond PENDING, also log its current status
INSERT INTO experiment_status_logs (experiment_id, status, created_at)
SELECT id, status, created_at FROM experiments WHERE status <> 'PENDING';
