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

// BacktestCompletedMessage is the payload sent to backtesting-completed-jobs Kafka topic
type BacktestCompletedMessage struct {
	JobID         string                 `json:"job_id"`
	ExperimentID  string                 `json:"experiment_id"`
	Symbol        string                 `json:"symbol"`
	Status        string                 `json:"status"`
	ResultMetrics map[string]interface{} `json:"result_metrics,omitempty"`
}

// CreateBacktestRequest represents the incoming JSON payload to run a backtest
type CreateBacktestRequest struct {
	StrategyName string   `json:"strategy_name" binding:"required"`
	Symbols      []string `json:"symbols" binding:"required,min=1"`
}

