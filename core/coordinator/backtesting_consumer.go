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
