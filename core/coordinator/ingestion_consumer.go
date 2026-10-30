package coordinator

import (
	"context"
	"time"

	"github.com/samirkhairati/backtesting/core/kafka"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/models"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type IngestionConsumer struct {
	Repo     *ExperimentRepository
	Producer kafka.Publisher
	consumer *kafka.Consumer
}

func NewIngestionConsumer(brokers []string, db DBQuerier, producer kafka.Publisher) (*IngestionConsumer, error) {
	consumer, err := kafka.NewConsumer(brokers, "coordinator-group", "ingestion-completed-jobs", "ingestion-job-events")
	if err != nil {
		return nil, err
	}
	repo := NewExperimentRepository(db)
	return &IngestionConsumer{
		Repo:     repo,
		Producer: producer,
		consumer: consumer,
	}, nil
}

func (ic *IngestionConsumer) Run(ctx context.Context) {
	ic.consumer.Run(ctx, ic.handle)
}

func (ic *IngestionConsumer) Close() {
	ic.consumer.Close()
}

func (ic *IngestionConsumer) handle(ctx context.Context, record *kgo.Record) error {
	if record.Topic == "ingestion-job-events" {
		var msg models.JobStartedMessage
		if err := ic.consumer.Unmarshal(record, &msg); err != nil {
			logger.Log.Error("Failed to unmarshal ingestion job started message", zap.Error(err))
			return nil
		}
		if err := ic.Repo.ClaimIngestionJob(ctx, msg.JobID, msg.WorkerID); err != nil {
			logger.Log.Error("Failed to claim ingestion job", zap.Error(err), zap.String("job_id", msg.JobID))
			return err
		}
		return nil
	}

	var msg models.IngestionCompletedMessage
	if err := ic.consumer.Unmarshal(record, &msg); err != nil {
		logger.Log.Error("Failed to unmarshal ingestion completed message", zap.Error(err))
		return nil // Do not retry malformed messages
	}

	if err := ic.Repo.UpdateIngestionJobStatus(ctx, msg.JobID, msg.Status); err != nil {
		return err
	}

	pending, total, err := ic.Repo.GetIngestionJobCounts(ctx, msg.ExperimentID)
	if err != nil {
		return err
	}

	completed := total - pending

	ic.Producer.Publish(ctx, "event-stream", msg.ExperimentID, models.ExperimentEvent{
		ExperimentID:  msg.ExperimentID,
		EventType:     models.EventTypeIngestionJobCompleted,
		Symbol:        msg.Symbol,
		TotalJobs:     total,
		CompletedJobs: completed,
	})

	if pending == 0 {
		service := NewExperimentService(ic.Repo, ic.Producer)
		if !service.Transition(ctx, msg.ExperimentID, models.ExperimentStatusPendingIngestion, models.ExperimentStatusCreatingBacktestJobs) {
			return nil // already moved on (or cancelled)
		}
		ic.Producer.Publish(ctx, "event-stream", msg.ExperimentID, models.ExperimentEvent{
			ExperimentID:  msg.ExperimentID,
			EventType:     models.EventTypeIngestionCompleted,
			TotalJobs:     total,
			CompletedJobs: completed,
		})

		// Creating thousands of jobs can take a while; do it off the consumer loop so
		// other ingestion messages keep flowing.
		go func(expID string) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			if err := service.GenerateBacktestJobs(bgCtx, expID); err != nil {
				logger.Log.Error("Failed to generate backtest jobs", zap.Error(err), zap.String("id", expID))
				service.Transition(bgCtx, expID, models.ExperimentStatusCreatingBacktestJobs, models.ExperimentStatusFailed)
			}
		}(msg.ExperimentID)
	}

	return nil
}
