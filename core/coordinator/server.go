package coordinator

import (
	"context"

	"github.com/samirkhairati/backtesting/core/kafka"
	pb "github.com/samirkhairati/backtesting/core/proto/coordinator"
)

type Server struct {
	pb.UnimplementedCoordinatorServiceServer
	DB       DBQuerier
	Producer kafka.Publisher
	service  *ExperimentService
}

func NewServer(db DBQuerier, producer kafka.Publisher) *Server {
	repo := NewExperimentRepository(db)
	service := NewExperimentService(repo, producer)
	return &Server{
		DB:       db,
		Producer: producer,
		service:  service,
	}
}

func (s *Server) CreateExperiment(ctx context.Context, req *pb.CreateExperimentRequest) (*pb.CreateExperimentResponse, error) {
	experimentID, status, err := s.service.CreateExperiment(
		ctx,
		req.UserId,
		req.Name,
		req.Description,
		req.ConfigJson,
	)

	if err != nil {
		return nil, err
	}

	return &pb.CreateExperimentResponse{
		ExperimentId: experimentID,
		Status:       status,
	}, nil
}
