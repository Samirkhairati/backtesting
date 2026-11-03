package models

import "time"

// WorkerHeartbeat represents the payload sent to the worker-heartbeats Kafka topic
type WorkerHeartbeat struct {
	WorkerID  string    `json:"worker_id"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
	// Kind is "backtesting" or "ingestion"; empty (older workers) is treated as backtesting.
	Kind string `json:"kind,omitempty"`
}
