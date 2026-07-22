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
	consumer, err := kafka.NewConsumer(brokers, "coordinator-group-backtesting", "backtesting-completed-jobs", "backtesting-job-events")
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
	if record.Topic == "backtesting-job-events" {
		var msg models.JobStartedMessage
		if err := bc.consumer.Unmarshal(record, &msg); err != nil {
			logger.Log.Error("Failed to unmarshal job started message", zap.Error(err))
			return nil
		}
		
		if err := bc.Repo.ClaimJob(ctx, msg.JobID, msg.WorkerID); err != nil {
			logger.Log.Error("Failed to claim job", zap.Error(err), zap.String("job_id", msg.JobID))
			return err
		}
		return nil
	}

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
		updated, err := bc.Repo.TryTransitionExperimentStatus(ctx, msg.ExperimentID, models.ExperimentStatusRunning, models.ExperimentStatusAggregating)
		if err != nil {
			return err
		}
		if !updated {
			return nil // Another worker already started aggregation
		}
		bc.Producer.Publish(ctx, "event-stream", msg.ExperimentID, models.ExperimentEvent{
			ExperimentID:  msg.ExperimentID,
			EventType:     models.EventTypeAggregationStarted,
			TotalJobs:     total,
			CompletedJobs: completed,
		})

		jobs, err := bc.Repo.GetBacktestJobMetrics(ctx, msg.ExperimentID)
		if err != nil {
			logger.Log.Error("Failed to get backtest job metrics", zap.Error(err))
		} else {
			// Run aggregation asynchronously
			go func(expID string, jobs []models.BacktestJob, tot, comp int) {
				bgCtx := context.Background()
				if len(jobs) > 0 {
					resultSummary, equityCurves := AggregateJobMetrics(jobs)
					if err := bc.Repo.UpdateExperimentResults(bgCtx, expID, resultSummary, equityCurves); err != nil {
						logger.Log.Error("Failed to update experiment results", zap.Error(err))
					}
				}

				if err := bc.Repo.UpdateExperimentStatus(bgCtx, expID, models.ExperimentStatusCompleted); err != nil {
					logger.Log.Error("Failed to update experiment status to completed", zap.Error(err))
					return
				}
				
				bc.Producer.Publish(bgCtx, "event-stream", expID, models.ExperimentEvent{
					ExperimentID:  expID,
					EventType:     models.EventTypeAggregationCompleted,
					TotalJobs:     tot,
					CompletedJobs: comp,
				})

				bc.Producer.Publish(bgCtx, "event-stream", expID, models.ExperimentEvent{
					ExperimentID:  expID,
					EventType:     models.EventTypeBacktestingCompleted,
					TotalJobs:     tot,
					CompletedJobs: comp,
				})
			}(msg.ExperimentID, jobs, total, completed)
		}
	}

	return nil
}

// AggregateJobMetrics processes a slice of BacktestJob models and computes aggregate metrics 
// across all jobs. It returns a result summary map and an equity curves map.
func AggregateJobMetrics(jobs []models.BacktestJob) (map[string]interface{}, map[string]interface{}) {
	var totalReturn, annReturn, maxDrawdown, sharpe, winRate, trades float64
	equityCurves := make(map[string]interface{})
	jobMetrics := make([]map[string]interface{}, 0)

	for _, job := range jobs {
		label := job.Symbol + " - " + job.StrategyName

		var tr, md, wr float64
		if v, ok := job.ResultMetrics["TotalReturn"].(float64); ok {
			tr = v
			totalReturn += v
		}
		if ar, ok := job.ResultMetrics["AnnualizedReturn"].(float64); ok {
			annReturn += ar
		}
		if v, ok := job.ResultMetrics["MaxDrawdown"].(float64); ok {
			md = v
			maxDrawdown += v
		}
		if sr, ok := job.ResultMetrics["SharpeRatio"].(float64); ok {
			sharpe += sr
		}
		if v, ok := job.ResultMetrics["WinRate"].(float64); ok {
			wr = v
			winRate += v
		}
		if tt, ok := job.ResultMetrics["TotalTrades"].(float64); ok {
			trades += tt
		}

		if ec, ok := job.ResultMetrics["EquityCurve"]; ok {
			equityCurves[label] = ec
		}

		jobMetrics = append(jobMetrics, map[string]interface{}{
			"Symbol":      job.Symbol,
			"Strategy":    job.StrategyName,
			"TotalReturn": tr,
			"MaxDrawdown": md,
			"WinRate":     wr,
		})
	}

	count := float64(len(jobs))
	var resultSummary map[string]interface{}
	
	if count > 0 {
		resultSummary = map[string]interface{}{
			"TotalReturn":      totalReturn / count,
			"AnnualizedReturn": annReturn / count,
			"MaxDrawdown":      maxDrawdown / count,
			"SharpeRatio":      sharpe / count,
			"WinRate":          winRate / count,
			"TotalTrades":      trades, // Sum of trades
			"JobMetrics":       jobMetrics,
		}
	} else {
		resultSummary = map[string]interface{}{
			"TotalReturn":      0.0,
			"AnnualizedReturn": 0.0,
			"MaxDrawdown":      0.0,
			"SharpeRatio":      0.0,
			"WinRate":          0.0,
			"TotalTrades":      0.0,
			"JobMetrics":       jobMetrics,
		}
	}

	return resultSummary, equityCurves
}
