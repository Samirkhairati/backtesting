DROP TRIGGER IF EXISTS experiment_status_trigger ON experiments;
DROP FUNCTION IF EXISTS log_experiment_status();
DROP TABLE IF EXISTS experiment_status_logs;
