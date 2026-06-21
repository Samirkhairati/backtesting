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
	DB       DBQuerier
	Producer kafka.Publisher
	consumer *kafka.Consumer
}

func NewIngestionConsumer(brokers []string, db DBQuerier, producer kafka.Publisher) (*IngestionConsumer, error) {
	consumer, err := kafka.NewConsumer(brokers, "coordinator-group", "ingestion-completed-jobs")
	if err != nil {
		return nil, err
	}
	return &IngestionConsumer{
		DB:       db,
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

	_, err := ic.DB.Exec(ctx, "UPDATE ingestion_jobs SET status=$1, completed_at=CURRENT_TIMESTAMP WHERE id=$2", msg.Status, msg.JobID)
	if err != nil {
		return err
	}

	var pending, total int
	err = ic.DB.QueryRow(ctx, "SELECT COUNT(*) FILTER (WHERE status='PENDING'), COUNT(*) FROM ingestion_jobs WHERE experiment_id=$1", msg.ExperimentID).Scan(&pending, &total)
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
		_, err = ic.DB.Exec(ctx, "UPDATE experiments SET status=$2 WHERE id=$1", msg.ExperimentID, models.ExperimentStatusIngested)
		if err != nil {
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
