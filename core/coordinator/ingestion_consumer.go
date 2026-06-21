package coordinator

import (
	"context"

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
	consumer, err := kafka.NewConsumer(brokers, "coordinator-group", "ingestion-completed-jobs")
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
		if err := ic.Repo.UpdateExperimentStatus(ctx, msg.ExperimentID, models.ExperimentStatusIngested); err != nil {
			return err
		}
		ic.Producer.Publish(ctx, "event-stream", msg.ExperimentID, models.ExperimentEvent{
			ExperimentID:  msg.ExperimentID,
			EventType:     models.EventTypeIngestionCompleted,
			TotalJobs:     total,
			CompletedJobs: completed,
		})
	}

	return nil
}
