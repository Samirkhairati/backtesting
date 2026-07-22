package models

// ExperimentStatus defines the possible states of an experiment
type ExperimentStatus string

const (
	ExperimentStatusPending          ExperimentStatus = "PENDING"
	ExperimentStatusPendingIngestion ExperimentStatus = "PENDING_INGESTION"
	ExperimentStatusReady            ExperimentStatus = "READY"
	ExperimentStatusIngested         ExperimentStatus = "INGESTED"
	ExperimentStatusRunning          ExperimentStatus = "RUNNING"
	ExperimentStatusAggregating      ExperimentStatus = "AGGREGATING"
	ExperimentStatusCompleted        ExperimentStatus = "COMPLETED"
	ExperimentStatusFailed           ExperimentStatus = "FAILED"
	ExperimentStatusCancelled        ExperimentStatus = "CANCELLED"
)

// IngestionJobStatus defines the possible states of an ingestion job
type IngestionJobStatus string

const (
	IngestionJobStatusPending   IngestionJobStatus = "PENDING"
	IngestionJobStatusCompleted IngestionJobStatus = "COMPLETED"
	IngestionJobStatusFailed    IngestionJobStatus = "FAILED"
)

// EventType defines the possible types of events published to the event-stream
type EventType string

const (
	EventTypeIngestionStarted      EventType = "INGESTION_STARTED"
	EventTypeIngestionJobCompleted EventType = "INGESTION_JOB_COMPLETED"
	EventTypeIngestionCompleted    EventType = "INGESTION_COMPLETED"
	EventTypeBacktestingStarted      EventType = "BACKTESTING_STARTED"
	EventTypeBacktestingJobCompleted EventType = "BACKTESTING_JOB_COMPLETED"
	EventTypeBacktestingCompleted    EventType = "BACKTESTING_COMPLETED"
	EventTypeAggregationStarted      EventType = "AGGREGATION_STARTED"
	EventTypeAggregationCompleted    EventType = "AGGREGATION_COMPLETED"
	EventTypeExperimentCancelled     EventType = "EXPERIMENT_CANCELLED"
)
