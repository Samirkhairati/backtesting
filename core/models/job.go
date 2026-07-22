package models

import "time"

// BacktestJob represents a backtest job stored in the database
type BacktestJob struct {
	ID            string                 `json:"id"`
	ExperimentID  string                 `json:"experiment_id"`
	Symbol        string                 `json:"symbol"`
	StrategyName  string                 `json:"strategy_name"`
	Parameters    map[string]interface{} `json:"parameters"`
	StartDate     string                 `json:"start_date"`
	EndDate       string                 `json:"end_date"`
	Status        string                 `json:"status"`
	CreatedAt     time.Time              `json:"created_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	ResultMetrics map[string]interface{} `json:"result_metrics,omitempty"`
	WorkerID      *string                `json:"worker_id,omitempty"`
	RetryCount    int                    `json:"retry_count"`
}

// BacktestJobMessage is the payload sent to backtesting-jobs Kafka topic
type BacktestJobMessage struct {
	JobID        string                 `json:"job_id"`
	ExperimentID string                 `json:"experiment_id"`
	Symbol       string                 `json:"symbol"`
	StrategyName string                 `json:"strategy_name"`
	Parameters   map[string]interface{} `json:"parameters"`
	StartDate    string                 `json:"start_date"`
	EndDate      string                 `json:"end_date"`
}

// JobStartedMessage is the payload sent when a worker picks up a job
type JobStartedMessage struct {
	JobID        string `json:"job_id"`
	ExperimentID string `json:"experiment_id"`
	WorkerID     string `json:"worker_id"`
}

// BacktestCompletedMessage is the payload sent to backtesting-completed-jobs Kafka topic
type BacktestCompletedMessage struct {
	JobID         string                 `json:"job_id"`
	ExperimentID  string                 `json:"experiment_id"`
	Symbol        string                 `json:"symbol"`
	Status        string                 `json:"status"`
	WorkerID      string                 `json:"worker_id"`
	ResultMetrics map[string]interface{} `json:"result_metrics,omitempty"`
}

// CreateBacktestRequest represents the incoming JSON payload to run a backtest
type CreateBacktestRequest struct {
	StrategyName string   `json:"strategy_name" binding:"required"`
	Symbols      []string `json:"symbols" binding:"required,min=1"`
}

