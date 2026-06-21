package coordinator

import (
	"context"

	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samirkhairati/backtesting/core/kafka"
	"github.com/samirkhairati/backtesting/core/models"
	pb "github.com/samirkhairati/backtesting/core/proto/coordinator"
	"github.com/samirkhairati/backtesting/core/logger"
	"go.uber.org/zap"
)

// DBQuerier defines the interface for database queries
type DBQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type Server struct {
	pb.UnimplementedCoordinatorServiceServer
	DB       DBQuerier
	Producer kafka.Publisher
}

func (s *Server) CreateExperiment(ctx context.Context, req *pb.CreateExperimentRequest) (*pb.CreateExperimentResponse, error) {
	var experimentID string

	// Insert experiment as PENDING
	err := s.DB.QueryRow(
		ctx,
		"INSERT INTO experiments (user_id, name, description, config, status) VALUES ($1, $2, $3, $4, 'PENDING') RETURNING id",
		req.UserId,
		req.Name,
		req.Description,
		req.ConfigJson,
	).Scan(&experimentID)

	if err != nil {
		logger.Log.Error("Coordinator failed to insert experiment", zap.Error(err))
		return nil, err
	}

	// Parse config to get symbols and dates
	var config struct {
		Symbols    []string `json:"symbols"`
		Timeframes []struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"timeframes"`
	}
	if err := json.Unmarshal([]byte(req.ConfigJson), &config); err != nil {
		return nil, err
	}

	if len(config.Symbols) == 0 || len(config.Timeframes) == 0 {
		return &pb.CreateExperimentResponse{ExperimentId: experimentID, Status: "READY"}, nil
	}

	startDateT, _ := time.Parse(time.RFC3339, config.Timeframes[0].Start)
	endDateT, _ := time.Parse(time.RFC3339, config.Timeframes[0].End)
	reqStart := startDateT.Format("2006-01-02")
	reqEnd := endDateT.Format("2006-01-02")

	// Check coverage for all symbols
	var missingSymbols []string
	for _, sym := range config.Symbols {
		var covStart, covEnd time.Time
		err := s.DB.QueryRow(ctx, "SELECT start_date, end_date FROM market_data_coverage WHERE symbol = $1", sym).Scan(&covStart, &covEnd)
		if err != nil || startDateT.Before(covStart) || endDateT.After(covEnd) {
			missingSymbols = append(missingSymbols, sym)
		}
	}

	var status string
	if len(missingSymbols) > 0 {
		status = string(models.ExperimentStatusPendingIngestion)
		_, err := s.DB.Exec(ctx, "UPDATE experiments SET status = $1 WHERE id = $2", status, experimentID)
		if err == nil && s.Producer != nil {
			for _, sym := range missingSymbols {
				var jobID string
				err := s.DB.QueryRow(ctx, "INSERT INTO ingestion_jobs (experiment_id, symbol, start_date, end_date, status) VALUES ($1, $2, $3, $4, 'PENDING') RETURNING id",
					experimentID, sym, reqStart, reqEnd).Scan(&jobID)
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
		status = string(models.ExperimentStatusReady)
		s.DB.Exec(ctx, "UPDATE experiments SET status = $1 WHERE id = $2", status, experimentID)
	}

	logger.Log.Info("Coordinator successfully created experiment", zap.String("id", experimentID), zap.String("status", status))

	return &pb.CreateExperimentResponse{
		ExperimentId: experimentID,
		Status:       status,
	}, nil
}
