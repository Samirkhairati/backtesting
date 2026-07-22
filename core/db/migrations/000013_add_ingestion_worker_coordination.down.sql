ALTER TABLE ingestion_jobs
DROP COLUMN IF EXISTS worker_id,
DROP COLUMN IF EXISTS retry_count;
