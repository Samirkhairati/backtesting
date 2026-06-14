package coordinator

import (
	"context"

	"github.com/jackc/pgx/v5"
	pb "github.com/samirkhairati/backtesting/core/proto/coordinator"
	"go.uber.org/zap"
	"github.com/samirkhairati/backtesting/core/logger"
)

// DBQuerier defines the interface for database queries
type DBQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Server struct {
	pb.UnimplementedCoordinatorServiceServer
	DB DBQuerier
}

func (s *Server) CreateExperiment(ctx context.Context, req *pb.CreateExperimentRequest) (*pb.CreateExperimentResponse, error) {
	var experimentID string

	// Parse out config string (since it's sent as JSON string, we can just insert it as JSONB)
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

	logger.Log.Info("Coordinator successfully created experiment", zap.String("id", experimentID))

	return &pb.CreateExperimentResponse{
		ExperimentId: experimentID,
		Status:       "PENDING",
	}, nil
}
