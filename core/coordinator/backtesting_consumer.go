package coordinator

import (
	"context"

	"github.com/samirkhairati/backtesting/core/kafka"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/models"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"encoding/json"
)

type BacktestingConsumer struct {
	Repo     *ExperimentRepository
	Producer kafka.Publisher
	consumer *kafka.Consumer
}

func NewBacktestingConsumer(brokers []string, db DBQuerier, producer kafka.Publisher) (*BacktestingConsumer, error) {
	consumer, err := kafka.NewConsumer(brokers, "coordinator-group-backtesting", "backtesting-completed-jobs")
	if err != nil {
		return nil, err
	}
	repo := NewExperimentRepository(db)
	return &BacktestingConsumer{
		Repo:     repo,
		Producer: producer,
		consumer: consumer,
	}, nil
}

func (bc *BacktestingConsumer) Run(ctx context.Context) {
	bc.consumer.Run(ctx, bc.handle)
}

func (bc *BacktestingConsumer) Close() {
	bc.consumer.Close()
}

func (bc *BacktestingConsumer) handle(ctx context.Context, record *kgo.Record) error {
	var msg models.BacktestCompletedMessage
	if err := bc.consumer.Unmarshal(record, &msg); err != nil {
		logger.Log.Error("Failed to unmarshal backtesting completed message", zap.Error(err))
		return nil // Do not retry malformed messages
	}

	metricsJSON, _ := json.Marshal(msg.ResultMetrics)

	if err := bc.Repo.UpdateBacktestJobStatus(ctx, msg.JobID, msg.Status, metricsJSON); err != nil {
		return err
	}

	pending, total, err := bc.Repo.GetBacktestJobCounts(ctx, msg.ExperimentID)
	if err != nil {
		return err
	}

	completed := total - pending

	bc.Producer.Publish(ctx, "event-stream", msg.ExperimentID, models.ExperimentEvent{
		ExperimentID:  msg.ExperimentID,
		EventType:     models.EventTypeBacktestingJobCompleted,
		Symbol:        msg.Symbol,
		TotalJobs:     total,
		CompletedJobs: completed,
	})

	if pending == 0 {
		jobs, err := bc.Repo.GetBacktestJobMetrics(ctx, msg.ExperimentID)
		if err != nil {
			logger.Log.Error("Failed to get backtest job metrics", zap.Error(err))
		} else if len(jobs) > 0 {
			var totalReturn, annReturn, maxDrawdown, sharpe, winRate, trades float64
			equityCurves := make(map[string]interface{})
			jobMetrics := make([]map[string]interface{}, 0)
			
			for _, job := range jobs {
				label := job.Symbol + " - " + job.StrategyName
				
				var tr, md, wr float64
				if v, ok := job.ResultMetrics["TotalReturn"].(float64); ok { tr = v; totalReturn += v }
				if ar, ok := job.ResultMetrics["AnnualizedReturn"].(float64); ok { annReturn += ar }
				if v, ok := job.ResultMetrics["MaxDrawdown"].(float64); ok { md = v; maxDrawdown += v }
				if sr, ok := job.ResultMetrics["SharpeRatio"].(float64); ok { sharpe += sr }
				if v, ok := job.ResultMetrics["WinRate"].(float64); ok { wr = v; winRate += v }
				if tt, ok := job.ResultMetrics["TotalTrades"].(float64); ok { trades += tt }
				
				if ec, ok := job.ResultMetrics["EquityCurve"]; ok {
					equityCurves[label] = ec
				}
				
				jobMetrics = append(jobMetrics, map[string]interface{}{
					"Symbol": job.Symbol,
					"Strategy": job.StrategyName,
					"TotalReturn": tr,
					"MaxDrawdown": md,
					"WinRate": wr,
				})
			}

			count := float64(len(jobs))
			resultSummary := map[string]interface{}{
				"TotalReturn":      totalReturn / count,
				"AnnualizedReturn": annReturn / count,
				"MaxDrawdown":      maxDrawdown / count,
				"SharpeRatio":      sharpe / count,
				"WinRate":          winRate / count,
				"TotalTrades":      trades, // Sum of trades
				"JobMetrics":       jobMetrics,
			}

			if err := bc.Repo.UpdateExperimentResults(ctx, msg.ExperimentID, resultSummary, equityCurves); err != nil {
				logger.Log.Error("Failed to update experiment results", zap.Error(err))
			}
		}

		if err := bc.Repo.UpdateExperimentStatus(ctx, msg.ExperimentID, models.ExperimentStatusCompleted); err != nil {
			return err
		}
		bc.Producer.Publish(ctx, "event-stream", msg.ExperimentID, models.ExperimentEvent{
			ExperimentID:  msg.ExperimentID,
			EventType:     models.EventTypeBacktestingCompleted,
			TotalJobs:     total,
			CompletedJobs: completed,
		})
	}

	return nil
}
