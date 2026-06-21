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
	"strings"

	"github.com/samirkhairati/backtesting/core/kafka"
	pb "github.com/samirkhairati/backtesting/core/proto/coordinator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

	// Kafka Setup
	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokersStr == "" {
		kafkaBrokersStr = "kafka:9092"
	}
	kafkaBrokers := strings.Split(kafkaBrokersStr, ",")

	producer, err := kafka.NewProducer(kafkaBrokers)
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
	}
	defer producer.Close()

	// Register the coordinator service
	coordinatorServer := &coordinator.Server{
		DB:       dbPool,
		Producer: producer,
	}
	pb.RegisterCoordinatorServiceServer(grpcServer, coordinatorServer)
	
	// Register reflection service on gRPC server
	reflection.Register(grpcServer)

	// Graceful shutdown setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ingestionConsumer, err := coordinator.NewIngestionConsumer(kafkaBrokers, dbPool, producer)
	if err != nil {
		log.Fatalf("failed to create ingestion consumer: %v", err)
	}
	defer ingestionConsumer.Close()
	go ingestionConsumer.Run(ctx)

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
