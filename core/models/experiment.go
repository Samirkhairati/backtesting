package models

import "time"

// CreateExperimentRequest represents the incoming JSON payload for an entire backtesting campaign
type CreateExperimentRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	Symbols     []string        `json:"symbols" binding:"required,min=1"`
	Strategies  []StrategySweep `json:"strategies" binding:"required,min=1"`
	Timeframes  []Timeframe     `json:"timeframes" binding:"required,min=1"`
}

type StrategySweep struct {
	Name       string                   `json:"name" binding:"required"`
	Parameters map[string][]interface{} `json:"parameters"`
}

type Timeframe struct {
	Start string `json:"start" binding:"required"`
	End   string `json:"end" binding:"required"`
}

// Experiment represents the parent experiment stored in the database
type Experiment struct {
	ID          string                  `json:"id"`
	UserID      string                  `json:"user_id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Config      CreateExperimentRequest `json:"config"`
	Status      string                  `json:"status"`
	CreatedAt   time.Time               `json:"created_at"`
}
