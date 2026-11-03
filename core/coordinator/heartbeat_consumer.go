package coordinator

import (
	"context"

	"github.com/samirkhairati/backtesting/core/kafka"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/models"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type HeartbeatConsumer struct {
	Repo     *ExperimentRepository
	consumer *kafka.Consumer
}

func NewHeartbeatConsumer(brokers []string, db DBQuerier) (*HeartbeatConsumer, error) {
	consumer, err := kafka.NewConsumer(brokers, "coordinator-group-heartbeats", "worker-heartbeats")
	if err != nil {
		return nil, err
	}
	repo := NewExperimentRepository(db)
	return &HeartbeatConsumer{
		Repo:     repo,
		consumer: consumer,
	}, nil
}

func (hc *HeartbeatConsumer) Run(ctx context.Context) {
	hc.consumer.Run(ctx, hc.handle)
}

func (hc *HeartbeatConsumer) Close() {
	hc.consumer.Close()
}

func (hc *HeartbeatConsumer) handle(ctx context.Context, record *kgo.Record) error {
	var msg models.WorkerHeartbeat
	if err := hc.consumer.Unmarshal(record, &msg); err != nil {
		logger.Log.Error("Failed to unmarshal heartbeat message", zap.Error(err))
		return nil // Do not retry malformed messages
	}

	if err := hc.Repo.UpsertWorkerHeartbeat(ctx, msg.WorkerID, msg.Status, msg.Kind); err != nil {
		logger.Log.Error("Failed to upsert worker heartbeat", zap.Error(err), zap.String("worker_id", msg.WorkerID))
		return err
	}

	return nil
}
