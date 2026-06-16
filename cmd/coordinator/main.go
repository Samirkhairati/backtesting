package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/samirkhairati/backtesting/core/coordinator"
	"github.com/samirkhairati/backtesting/core/db"
	"github.com/samirkhairati/backtesting/core/logger"
	pb "github.com/samirkhairati/backtesting/core/proto/coordinator"
	ingestionpb "github.com/samirkhairati/backtesting/core/proto/ingestion"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger.Init()

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(".env"); err != nil {
			logger.Log.Warn("No .env file found")
		}
	}

	ctx := context.Background()
	dbPool := db.Setup(ctx)
	defer dbPool.Close()

	port := "50051"

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	
	ingestionUrl := os.Getenv("INGESTION_SERVICE_URL")
	if ingestionUrl == "" {
		ingestionUrl = "ingestion-dev:50052"
	}
	
	conn, err := grpc.NewClient(ingestionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Warn("Could not connect to Ingestion service", zap.Error(err))
	} else {
		defer conn.Close()
	}

	var ingestionClient ingestionpb.IngestionServiceClient
	if conn != nil {
		ingestionClient = ingestionpb.NewIngestionServiceClient(conn)
	}

	// Register the coordinator service
	coordinatorServer := &coordinator.Server{
		DB: dbPool,
		IngestionClient: ingestionClient,
	}
	pb.RegisterCoordinatorServiceServer(grpcServer, coordinatorServer)
	
	// Register reflection service on gRPC server
	reflection.Register(grpcServer)

	// Graceful shutdown setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Log.Info("Coordinator gRPC server listening", zap.String("port", port))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Log.Fatal("failed to serve", zap.Error(err))
		}
	}()

	<-ctx.Done()

	logger.Log.Info("Shutting down gracefully, press Ctrl+C again to force")
	
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	grpcServer.GracefulStop()
	logger.Log.Info("Coordinator server exiting")
}
