package worker

import (
	"context"

	"time"

	"github.com/google/uuid"
	"github.com/samirkhairati/backtesting/core/data/store"
	"github.com/samirkhairati/backtesting/core/kafka"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/metrics"
	"github.com/samirkhairati/backtesting/core/models"
	"github.com/samirkhairati/backtesting/core/portfolio"
	"github.com/samirkhairati/backtesting/core/simulator"
	"github.com/samirkhairati/backtesting/core/strategy"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type Worker struct {
	ID       string
	Store    store.Store
	Producer kafka.Publisher
	consumer *kafka.Consumer
}

func NewWorker(brokers []string, dbStore store.Store, producer kafka.Publisher) (*Worker, error) {
	consumer, err := kafka.NewConsumer(brokers, "worker-group", "backtesting-jobs")
	if err != nil {
		return nil, err
	}
	return &Worker{
		ID:       uuid.New().String(),
		Store:    dbStore,
		Producer: producer,
		consumer: consumer,
	}, nil
}

func (w *Worker) Run(ctx context.Context) {
	go w.heartbeatLoop(ctx)
	w.consumer.Run(ctx, w.handle)
}

func (w *Worker) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hb := models.WorkerHeartbeat{
				WorkerID:  w.ID,
				Timestamp: time.Now(),
				Status:    "ACTIVE",
			}
			if err := w.Producer.Publish(ctx, "worker-heartbeats", w.ID, hb); err != nil {
				logger.Log.Error("Failed to publish heartbeat", zap.Error(err))
			}
		}
	}
}

func (w *Worker) Close() {
	w.consumer.Close()
}

func (w *Worker) handle(ctx context.Context, record *kgo.Record) error {
	var msg models.BacktestJobMessage
	if err := w.consumer.Unmarshal(record, &msg); err != nil {
		logger.Log.Error("Failed to unmarshal backtest job message", zap.Error(err))
		return nil // Ignore malformed
	}

	logger.Log.Info("Starting backtest job", zap.String("job_id", msg.JobID), zap.String("symbol", msg.Symbol), zap.String("worker_id", w.ID))

	// Publish job-started event
	startedMsg := models.JobStartedMessage{
		JobID:        msg.JobID,
		ExperimentID: msg.ExperimentID,
		WorkerID:     w.ID,
	}
	if err := w.Producer.Publish(ctx, "backtesting-job-events", msg.JobID, startedMsg); err != nil {
		logger.Log.Error("Failed to publish job-started event", zap.Error(err))
	}

	// Instantiate strategy
	strat, err := strategy.CreateStrategy(msg.StrategyName, msg.Parameters)
	if err != nil {
		logger.Log.Error("Failed to create strategy", zap.Error(err))
		w.publishCompleted(ctx, msg, "FAILED", nil)
		return nil
	}

	// Fetch data
	iter, err := w.Store.GetMultiBarsWithTimeframe(ctx, []string{msg.Symbol}, msg.StartDate, msg.EndDate)
	if err != nil {
		logger.Log.Error("Failed to get bars", zap.Error(err))
		w.publishCompleted(ctx, msg, "FAILED", nil)
		return nil
	}

	// Run simulation
	engine := simulator.NewEngine(iter, strat)
	if err := engine.Run(); err != nil {
		logger.Log.Error("Simulation failed", zap.Error(err))
		w.publishCompleted(ctx, msg, "FAILED", nil)
		return nil
	}

	// Generate report
	report := metrics.GenerateReport(engine.Portfolio)
	
	resultMap := map[string]interface{}{
		"TotalReturn":      report.TotalReturnPct,
		"AnnualizedReturn": report.CAGR,
		"MaxDrawdown":      report.MaxDrawdownPct,
		"SharpeRatio":      report.SharpeRatio,
		"WinRate":          report.TradeAnalytics.WinRatePct,
		"TotalTrades":      report.TradeAnalytics.TotalTrades,
		"EquityCurve":      downsampleEquityCurve(engine.Portfolio.EquityCurve, 100),
	}

	w.publishCompleted(ctx, msg, "COMPLETED", resultMap)
	logger.Log.Info("Completed backtest job", zap.String("job_id", msg.JobID))
	return nil
}

func (w *Worker) publishCompleted(ctx context.Context, msg models.BacktestJobMessage, status string, results map[string]interface{}) {
	completedMsg := models.BacktestCompletedMessage{
		JobID:         msg.JobID,
		ExperimentID:  msg.ExperimentID,
		Symbol:        msg.Symbol,
		Status:        status,
		WorkerID:      w.ID,
		ResultMetrics: results,
	}
	
	if err := w.Producer.Publish(ctx, "backtesting-completed-jobs", msg.ExperimentID, completedMsg); err != nil {
		logger.Log.Error("Failed to publish backtesting completed job", zap.Error(err))
	}
}

func downsampleEquityCurve(curve []portfolio.EquityPoint, maxPoints int) []portfolio.EquityPoint {
	if len(curve) <= maxPoints {
		return curve
	}
	
	result := make([]portfolio.EquityPoint, 0, maxPoints)
	step := float64(len(curve)-1) / float64(maxPoints-1)
	
	for i := 0; i < maxPoints-1; i++ {
		idx := int(float64(i) * step)
		result = append(result, curve[idx])
	}
	result = append(result, curve[len(curve)-1])
	
	return result
}
