package main

import (
	"context"
	"os"

	"github.com/samirkhairati/backtesting/core/api/router"
	"github.com/samirkhairati/backtesting/core/db"
	"github.com/samirkhairati/backtesting/core/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "github.com/samirkhairati/backtesting/core/proto/coordinator"
)

func main() {
	// Initialize structured logger
	logger.Init()
	defer logger.Sync()

	port := "8080"

	coordinatorAddr := os.Getenv("COORDINATOR_ADDR")
	if coordinatorAddr == "" {
		logger.Log.Fatal("COORDINATOR_ADDR environment variable is not set")
	}

	logger.Log.Sugar().Infof("Starting API Server on port %s...", port)

	ctx := context.Background()
	dbPool := db.Setup(ctx)
	defer dbPool.Close()

	// Set up a connection to the Coordinator gRPC server.
	conn, err := grpc.NewClient(coordinatorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatal("did not connect to coordinator", zap.Error(err))
	}
	defer conn.Close()
	coordinatorClient := pb.NewCoordinatorServiceClient(conn)

	r := router.SetupRouter(dbPool, coordinatorClient)

	if err := r.Run(":" + port); err != nil {
		logger.Log.Fatal("Failed to start API server", zap.Error(err))
	}
}
