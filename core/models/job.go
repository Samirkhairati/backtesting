package models

import "time"

// BacktestJob represents a backtest job stored in the database
type BacktestJob struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	StrategyName  string                 `json:"strategy_name"`
	Symbols       []string               `json:"symbols"`
	Status        string                 `json:"status"`
	CreatedAt     time.Time              `json:"created_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	ResultMetrics map[string]interface{} `json:"result_metrics,omitempty"`
}

// CreateBacktestRequest represents the incoming JSON payload to run a backtest
type CreateBacktestRequest struct {
	StrategyName string   `json:"strategy_name" binding:"required"`
	Symbols      []string `json:"symbols" binding:"required,min=1"`
}
