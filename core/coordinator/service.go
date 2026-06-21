package coordinator

import (
	"context"
	"encoding/json"
	"time"

	"github.com/samirkhairati/backtesting/core/kafka"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/models"
	"go.uber.org/zap"
)

type ExperimentService struct {
	Repo     *ExperimentRepository
	Producer kafka.Publisher
}

func NewExperimentService(repo *ExperimentRepository, producer kafka.Publisher) *ExperimentService {
	return &ExperimentService{
		Repo:     repo,
		Producer: producer,
	}
}

func (s *ExperimentService) CreateExperiment(ctx context.Context, userID, name, description, configJSON string) (string, string, error) {
	experimentID, err := s.Repo.InsertExperiment(ctx, userID, name, description, configJSON)
	if err != nil {
		logger.Log.Error("Failed to insert experiment", zap.Error(err))
		return "", "", err
	}

	var config struct {
		Symbols    []string `json:"symbols"`
		Timeframes []struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"timeframes"`
	}
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return experimentID, string(models.ExperimentStatusPending), err
	}

	if len(config.Symbols) == 0 || len(config.Timeframes) == 0 {
		return experimentID, string(models.ExperimentStatusReady), nil
	}

	startDateT, _ := time.Parse(time.RFC3339, config.Timeframes[0].Start)
	endDateT, _ := time.Parse(time.RFC3339, config.Timeframes[0].End)
	reqStart := startDateT.Format("2006-01-02")
	reqEnd := endDateT.Format("2006-01-02")

	var missingSymbols []string
	for _, sym := range config.Symbols {
		covStart, covEnd, err := s.Repo.GetMarketDataCoverage(ctx, sym)
		if err != nil || startDateT.Before(covStart) || endDateT.After(covEnd) {
			missingSymbols = append(missingSymbols, sym)
		}
	}

	var status models.ExperimentStatus
	if len(missingSymbols) > 0 {
		status = models.ExperimentStatusPendingIngestion
		err := s.Repo.UpdateExperimentStatus(ctx, experimentID, status)
		if err == nil && s.Producer != nil {
			for _, sym := range missingSymbols {
				jobID, err := s.Repo.InsertIngestionJob(ctx, experimentID, sym, reqStart, reqEnd)
				if err != nil {
					logger.Log.Error("Failed to insert ingestion job", zap.Error(err))
					continue
				}

				publishErr := s.Producer.Publish(ctx, "ingestion-jobs", experimentID, models.IngestionJobMessage{
					JobID:        jobID,
					ExperimentID: experimentID,
					Symbol:       sym,
					StartDate:    reqStart,
					EndDate:      reqEnd,
				})
				if publishErr != nil {
					logger.Log.Error("Failed to publish ingestion job", zap.Error(publishErr))
				}
			}
			publishErr := s.Producer.Publish(ctx, "event-stream", experimentID, models.ExperimentEvent{
				ExperimentID: experimentID,
				EventType:    models.EventTypeIngestionStarted,
				TotalJobs:    len(missingSymbols),
			})
			if publishErr != nil {
				logger.Log.Error("Failed to publish event stream message", zap.Error(publishErr))
			}
		}
	} else {
		status = models.ExperimentStatusReady
		s.Repo.UpdateExperimentStatus(ctx, experimentID, status)
	}

	logger.Log.Info("Experiment created", zap.String("id", experimentID), zap.String("status", string(status)))
	return experimentID, string(status), nil
}
