package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/samirkhairati/backtesting/core/data/store/postgres"
	"github.com/samirkhairati/backtesting/core/data/store/cache"
	"github.com/samirkhairati/backtesting/core/db"
	"github.com/samirkhairati/backtesting/core/kafka"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/worker"
	"go.uber.org/zap"
)

func main() {
	logger.Init()

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(".env"); err != nil {
			logger.Log.Warn("No .env file found")
		}
	}

	log.Println("Starting Backtesting Engine Worker...")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool := db.Setup(ctx)
	defer pool.Close()

	redisClient := db.SetupRedis(ctx)
	defer redisClient.Close()

	pgStore := postgres.NewStore(pool)
	store := cache.NewRedisStore(redisClient, pgStore)

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

	w, err := worker.NewWorker(kafkaBrokers, store, producer, pool)
	if err != nil {
		log.Fatalf("failed to create worker: %v", err)
	}
	defer w.Close()

	logger.Log.Info("Worker is running and waiting for jobs", zap.Strings("brokers", kafkaBrokers))
	w.Run(ctx)
	
	<-ctx.Done()
	logger.Log.Info("Worker exiting gracefully")
}
