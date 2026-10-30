package models

type IngestionJobMessage struct {
	JobID        string `json:"job_id"`
	ExperimentID string `json:"experiment_id"`
	Symbol       string `json:"symbol"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
}

type IngestionCompletedMessage struct {
	JobID        string             `json:"job_id"`
	ExperimentID string             `json:"experiment_id"`
	Symbol       string             `json:"symbol"`
	Status       IngestionJobStatus `json:"status"` // "COMPLETED" or "FAILED"
}

type ExperimentEvent struct {
	ExperimentID  string    `json:"experiment_id"`
	EventType     EventType `json:"event_type"` // INGESTION_STARTED, INGESTION_JOB_COMPLETED, INGESTION_COMPLETED
	Symbol        string    `json:"symbol,omitempty"`
	Timestamp     string `json:"timestamp,omitempty"`
	TotalJobs     int    `json:"total_jobs,omitempty"`
	CompletedJobs int    `json:"completed_jobs,omitempty"`
	// Status carries the new experiment status on STATUS_CHANGED events.
	Status string `json:"status,omitempty"`
}
