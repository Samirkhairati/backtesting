ALTER TABLE experiments ADD COLUMN IF NOT EXISTS result_summary JSONB;
ALTER TABLE experiments ADD COLUMN IF NOT EXISTS equity_curves JSONB;
