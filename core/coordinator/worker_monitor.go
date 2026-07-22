package coordinator

import (
	"context"
	"time"

	"github.com/samirkhairati/backtesting/core/kafka"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/models"
	"go.uber.org/zap"
)

type WorkerMonitor struct {
	Repo     *ExperimentRepository
	Producer kafka.Publisher
}

func NewWorkerMonitor(repo *ExperimentRepository, producer kafka.Publisher) *WorkerMonitor {
	return &WorkerMonitor{
		Repo:     repo,
		Producer: producer,
	}
}

func (m *WorkerMonitor) Run(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkDeadWorkers(ctx)
		}
	}
}

func (m *WorkerMonitor) checkDeadWorkers(ctx context.Context) {
	deadWorkers, err := m.Repo.GetDeadWorkers(ctx, 60)
	if err != nil {
		logger.Log.Error("Failed to get dead workers", zap.Error(err))
		return
	}

	for _, workerID := range deadWorkers {
		logger.Log.Warn("Worker appears dead, marking unhealthy and requeuing jobs", zap.String("worker_id", workerID))
		requeuedJobs, failedJobs, err := m.Repo.RequeueJobsForWorker(ctx, workerID)
		if err != nil {
			logger.Log.Error("Failed to requeue jobs for dead worker", zap.Error(err), zap.String("worker_id", workerID))
			continue
		}

		for _, job := range requeuedJobs {
			if err := m.Producer.Publish(ctx, "backtesting-jobs", job.JobID, job); err != nil {
				logger.Log.Error("Failed to republish requeued job", zap.Error(err), zap.String("job_id", job.JobID))
			}
		}

		// Publish completion events for jobs that permanently failed so the experiment can transition
		for _, job := range failedJobs {
			msg := models.BacktestCompletedMessage{
				JobID:        job.JobID,
				ExperimentID: job.ExperimentID,
				Symbol:       job.Symbol,
				Status:       "FAILED",
				WorkerID:     workerID,
			}
			if err := m.Producer.Publish(ctx, "backtesting-completed-jobs", job.ExperimentID, msg); err != nil {
				logger.Log.Error("Failed to publish FAILED status for permanently failed job", zap.Error(err), zap.String("job_id", job.JobID))
			}
		}

		// Requeue ingestion jobs for the dead worker
		requeuedIngestionJobs, failedIngestionJobs, err := m.Repo.RequeueIngestionJobsForWorker(ctx, workerID)
		if err != nil {
			logger.Log.Error("Failed to requeue ingestion jobs for dead worker", zap.Error(err), zap.String("worker_id", workerID))
			continue
		}

		for _, job := range requeuedIngestionJobs {
			if err := m.Producer.Publish(ctx, "ingestion-jobs", job.JobID, job); err != nil {
				logger.Log.Error("Failed to republish requeued ingestion job", zap.Error(err), zap.String("job_id", job.JobID))
			}
		}

		// Publish completion events for ingestion jobs that permanently failed
		for _, job := range failedIngestionJobs {
			msg := models.IngestionCompletedMessage{
				JobID:        job.JobID,
				ExperimentID: job.ExperimentID,
				Symbol:       job.Symbol,
				Status:       models.IngestionJobStatus("FAILED"),
			}
			if err := m.Producer.Publish(ctx, "ingestion-completed-jobs", job.ExperimentID, msg); err != nil {
				logger.Log.Error("Failed to publish FAILED status for permanently failed ingestion job", zap.Error(err), zap.String("job_id", job.JobID))
			}
		}
	}
}
